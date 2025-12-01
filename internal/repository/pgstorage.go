package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

type DBInitMode int

const (
	ManagedExternally DBInitMode = iota
	ManagedInternally
	ManagedInternallyForced
)

func (m *DBInitMode) UnmarshalText(text []byte) error {
	s := strings.ToLower(strings.TrimSpace(string(text)))
	switch s {
	case "external", "0":
		*m = ManagedExternally
	case "internal", "1":
		*m = ManagedInternally
	case "reset", "forced", "2":
		*m = ManagedInternallyForced
	}
	return nil
}

type PGStorage struct {
	DB *sql.DB
}

type MigrationCMD int

const (
	UP MigrationCMD = iota
	DOWN
)

func NewPGDB(dsn string, mode DBInitMode) (*PGStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}
	storage := PGStorage{db}
	switch mode {
	case ManagedExternally:
		log.Println("External managed DB is chosen")
	case ManagedInternallyForced:
		log.Println("Internal managed DB is chosen with recreating DB entities")
		err := storage.RunMigration("migrations", DOWN)
		if err != nil {
			return nil, err
		}
		err = storage.RunMigration("migrations", UP)
		if err != nil {
			return nil, err
		}
	case ManagedInternally:
		log.Println("Internal managed DB is chosen, with creating DB entities if they don't exist")
		err := storage.RunMigration("migrations", UP)
		if err != nil {
			return nil, err
		}
	default:
		log.Println("Wrong mode, use external managed DB as default")
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return &storage, nil
}

func (p *PGStorage) RunMigration(migrationsDir string, cmd MigrationCMD) error {
	driver, err := postgres.WithInstance(p.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: open driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: new instance: %w", err)
	}
	switch cmd {
	case UP:
		err = m.Up()
		if err != nil {
			log.Printf("migrate up: %v", err)
		}
	case DOWN:
		err = m.Down()
		if err != nil {
			return fmt.Errorf("migrate down: %w", err)
		}
	default:
		return fmt.Errorf("wrong migrate cmd %d", cmd)
	}
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (p *PGStorage) Close() error {
	return p.DB.Close()
}

func (p *PGStorage) PingDB(ctx context.Context) error {
	type result struct{}
	f := func() (result, error) {
		return result{}, p.DB.PingContext(ctx)
	}
	_, err := utils.WithRetry(ctx, f)
	return err
}

func (p *PGStorage) AddMetric(ctx context.Context, agentID string, metricType string, metricName string, value float64, delta int64) (err error) {
	log.Println("AddMetric is starting")
	type result struct{}
	f := func() (result result, err error) {
		tx, err := p.DB.BeginTx(ctx, nil)
		if err != nil {
			return result, err
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		_, err = tx.ExecContext(ctx, `INSERT INTO metrics (agent_id, id, mtype, delta, value, updated_at) 
			VALUES ($1,$2,$3,
				CASE WHEN $3='counter' THEN $4::bigint ELSE NULL END,
				CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
				now())
			ON CONFLICT (agent_id, id, mtype) DO UPDATE
				SET delta = CASE WHEN $3='counter' THEN COALESCE(metrics.delta,0)::bigint + $4::bigint ELSE NULL END,
				value = CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
				updated_at = now()
			`, agentID, metricName, metricType, delta, value)
		if err != nil {
			return result, err
		}
		return result, nil
	}
	_, err = utils.WithRetry(ctx, f)
	return err
}

func (p *PGStorage) AddMetrics(ctx context.Context, agentID string, metrics []model.Metrics) (err error) {
	log.Println("AddMetrics is starting")
	type result struct{}
	f := func() (result result, err error) {
		tx, err := p.DB.BeginTx(ctx, nil)
		if err != nil {
			return result, err
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		for _, m := range metrics {
			var d any
			if m.Delta != nil {
				d = *m.Delta
			} else {
				d = nil
			}
			var v any
			if m.Value != nil {
				v = *m.Value
			} else {
				v = nil
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO metrics (agent_id, id, mtype, delta, value, updated_at) 
			VALUES ($1,$2,$3,
				CASE WHEN $3='counter' THEN $4::bigint ELSE NULL END,
				CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
				now())
			ON CONFLICT (agent_id, id, mtype) DO UPDATE
				SET delta = CASE WHEN $3='counter' THEN COALESCE(metrics.delta,0)::bigint + $4::bigint ELSE NULL END,
				value = CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
				updated_at = now()
				`, agentID, m.ID, m.MType, d, v)
			if err != nil {
				return result, err
			}
		}
		return result, nil
	}
	_, err = utils.WithRetry(ctx, f)
	return err
}

func (p *PGStorage) GetMetric(ctx context.Context, agentID string, metricType string, metricName string) (value float64, delta int64, err error) {
	type MetricResult struct {
		value float64
		delta int64
	}
	f := func() (result MetricResult, err error) {
		var v sql.NullFloat64
		var d sql.NullInt64

		row := p.DB.QueryRowContext(ctx, `
			SELECT value, delta
			FROM metrics
			WHERE agent_id = $1 AND id = $2 AND mtype = $3
			`, agentID, metricName, metricType)
		if err = row.Scan(&v, &d); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return result, fmt.Errorf("metric not found: agent_id=%s id=%s type=%s", agentID, metricName, metricType)
			}
			return result, err
		}

		switch metricType {
		case model.Counter:
			if d.Valid {
				result.delta = d.Int64
				return result, nil
			} else {
				return result, fmt.Errorf("counter metric has no delta: agent_id=%s id=%s", agentID, metricName)
			}

		case model.Gauge:
			if v.Valid {
				result.value = v.Float64
				return result, nil
			}
			return result, fmt.Errorf("gauge metric has no value: agent_id=%s id=%s", agentID, metricName)

		default:
			return result, fmt.Errorf("unknown metric type: %s", metricType)
		}
	}
	result, err := utils.WithRetry(ctx, f)
	return result.value, result.delta, err
}

func (p *PGStorage) GetObjMetric(ctx context.Context, agentID string, metricType string, metricName string) (m *model.Metrics, err error) {
	f := func() (m *model.Metrics, err error) {
		var id, mtype string
		var v sql.NullFloat64
		var d sql.NullInt64
		row := p.DB.QueryRowContext(ctx, `
			SELECT id, mtype, delta, value
			FROM metrics
			WHERE agent_id = $1 AND id = $2 AND mtype = $3
			`, agentID, metricName, metricType)
		if err = row.Scan(&id, &mtype, &d, &v); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("metric not found: agent_id=%s id=%s type=%s", agentID, metricName, metricType)
			}
			return nil, err
		}
		m = &model.Metrics{
			ID:    id,
			MType: mtype,
		}
		if d.Valid {
			delta := d.Int64
			m.Delta = &delta
		}
		if v.Valid {
			value := v.Float64
			m.Value = &value
		}

		return m, err
	}
	m, err = utils.WithRetry(ctx, f)
	return m, err
}

func (p *PGStorage) GetStore(ctx context.Context) (store map[string]*model.Metrics, err error) {
	f := func() (store map[string]*model.Metrics, err error) {
		rows, err := p.DB.QueryContext(ctx, `
			SELECT agent_id, id, mtype, delta, value
			FROM metrics
			ORDER BY agent_id, id`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		store = make(map[string]*model.Metrics)
		for rows.Next() {
			var agentID, id, mtype string
			var v sql.NullFloat64
			var d sql.NullInt64

			if err = rows.Scan(&agentID, &id, &mtype, &d, &v); err != nil {
				return nil, err
			}
			metric := model.Metrics{
				ID:    id,
				MType: mtype,
			}
			if d.Valid {
				delta := d.Int64
				metric.Delta = &(delta)
			}
			if v.Valid {
				value := v.Float64
				metric.Value = &(value)
			}
			store[agentID+"_"+id] = &metric
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
		return store, nil
	}
	store, err = utils.WithRetry(ctx, f)
	return store, err
}
