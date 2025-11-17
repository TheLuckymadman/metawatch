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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/TheLuckymadman/metawatch/internal/model"
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
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres", driver)
	if err != nil {
		return err
	}
	switch cmd {
	case UP:
		err = m.Up()
	case DOWN:
		err = m.Down()
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
	if err := p.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}
	return nil
}

func (p *PGStorage) AddMetric(ctx context.Context, agentID string, metricType string, metricName string, value float64, delta int64) (err error) {
	log.Println("AddMetric is starting")
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	res, err := tx.ExecContext(ctx, `UPDATE metrics 
		SET delta = CASE WHEN $3='counter' THEN COALESCE(delta,0)::bigint + $4 ELSE NULL END,
			value = CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
			updated_at = now()
		WHERE agent_id=$1 AND id=$2 AND mtype=$3
		`, agentID, metricName, metricType, delta, value)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO metrics (agent_id, id, mtype, delta, value, updated_at) 
		VALUES ($1,$2,$3,
			CASE WHEN $3='counter' THEN $4::bigint ELSE NULL END,
			CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
			now())
			`, agentID, metricName, metricType, delta, value)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				_, err = tx.ExecContext(ctx, `
					UPDATE metrics
					SET delta      = CASE WHEN $3='counter' THEN COALESCE(delta,0)::bigint + $4 ELSE NULL END,
						value      = CASE WHEN $3='gauge'   THEN $5::double precision               ELSE NULL END,
						updated_at = now()
					WHERE agent_id=$1 AND id=$2 AND mtype=$3
				`, agentID, metricName, metricType, delta, value)
				if err != nil {
					return err
				}
				return nil
			}
			return err
		}
		return nil
	}
	return nil
}

func (p *PGStorage) AddMetrics(ctx context.Context, agentID string, metrics []model.Metrics) (err error) {
	log.Println("AddMetrics is starting")
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
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
		res, err := tx.ExecContext(ctx, `UPDATE metrics 
		SET delta = CASE WHEN $3='counter' THEN COALESCE(delta,0)::bigint + $4 ELSE NULL END,
			value = CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
			updated_at = now()
		WHERE agent_id=$1 AND id=$2 AND mtype=$3
		`, agentID, m.ID, m.MType, d, v)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			_, err = tx.ExecContext(ctx, `
		INSERT INTO metrics (agent_id, id, mtype, delta, value, updated_at) 
		VALUES ($1,$2,$3,
			CASE WHEN $3='counter' THEN $4::bigint ELSE NULL END,
			CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
			now())
			`, agentID, m.ID, m.MType, d, v)
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
					_, err = tx.ExecContext(ctx, `
					UPDATE metrics
					SET delta      = CASE WHEN $3='counter' THEN COALESCE(delta,0)::bigint + $4 ELSE NULL END,
						value      = CASE WHEN $3='gauge'   THEN $5::double precision               ELSE NULL END,
						updated_at = now()
					WHERE agent_id=$1 AND id=$2 AND mtype=$3
				`, agentID, m.ID, m.MType, d, v)
					if err != nil {
						return err
					}
					continue
				}
				return err
			}
		}
	}
	return nil
}

func (p *PGStorage) GetMetric(ctx context.Context, agentID string, metricType string, metricName string) (value float64, delta int64, err error) {
	var v sql.NullFloat64
	var d sql.NullInt64

	row := p.DB.QueryRowContext(ctx, `
		SELECT value, delta
		FROM metrics
		WHERE agent_id = $1 AND id = $2 AND mtype = $3
		`, agentID, metricName, metricType)
	if err = row.Scan(&v, &d); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("metric not found: agent_id=%s id=%s type=%s", agentID, metricName, metricType)
		}
		return 0, 0, err
	}

	switch metricType {
	case model.Counter:
		if d.Valid {
			return 0, d.Int64, nil
		} else {
			return 0, 0, fmt.Errorf("counter metric has no delta: agent_id=%s id=%s", agentID, metricName)
		}

	case model.Gauge:
		if v.Valid {
			return v.Float64, 0, nil
		}
		return 0, 0, fmt.Errorf("gauge metric has no value: agent_id=%s id=%s", agentID, metricName)

	default:
		return 0, 0, fmt.Errorf("unknown metric type: %s", metricType)
	}
}

func (p *PGStorage) GetObjMetric(ctx context.Context, agentID string, metricType string, metricName string) (m *model.Metrics, err error) {
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
		m.Delta = &(d.Int64)
	}
	if v.Valid {
		m.Value = &(v.Float64)
	}

	return m, err
}

func (p *PGStorage) GetStore(ctx context.Context) (store map[string]*model.Metrics, err error) {
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
			metric.Delta = &(d.Int64)
		}
		if v.Valid {
			metric.Value = &(v.Float64)
		}
		store[agentID+"_"+id] = &metric
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return store, nil
}
