package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)


func TestPGStorage_AddMetric(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("load envs: %v", err)
	}

	DSN := os.Getenv("DATABASE_DSN")
	if DSN == "" {
		t.Logf("skip test because of no DSN")
		return
	}

	tests := []struct {
		name       string
		dsn        string
		mode       DBInitMode
		agentID    string
		metricType string
		metricName string
		value      float64
		delta      int64
		wantErr    bool
	}{
		{
			name:       "add counter to DB",
			dsn:        DSN,
			mode:       ManagedExternally,
			agentID:    "127.0.0.1_cpu_test",
			metricType: model.Counter,
			metricName: "cpu",
			delta:      1,
			wantErr:    false,
		},
		{
			name:       "add gauge to DB",
			dsn:        DSN,
			mode:       ManagedExternally,
			agentID:    "127.0.0.1_ram_test",
			metricType: model.Gauge,
			metricName: "ram",
			value:      1000,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPGDB(tt.dsn, tt.mode)
			if err != nil {
				t.Fatalf("AddMetric could not construct DB object: %v", err)
			}

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)
			if err != nil {
				t.Fatalf("AddMetric() delete metric in DB: %v", err)
			}

			gotErr := p.AddMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName, tt.value, tt.delta)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddMetric() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddMetric() succeeded unexpectedly")
			}

			row := p.DB.QueryRow(`
			SELECT agent_id, id, mtype, delta, value, hash
			FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)

			var fetchedMetric model.Metrics
			var agentID string
			var v sql.NullFloat64
			var d sql.NullInt64
			var h sql.NullString
			err = row.Scan(&agentID, &fetchedMetric.ID, &fetchedMetric.MType, &d, &v, &h)
			if err != nil {
				t.Errorf("AddMetric() row execution: %v", err)
			}

			switch tt.metricType {
			case model.Counter:
				if d.Valid {
					fetchedMetric.Delta = utils.Int64Ptr(d.Int64)
					fetchedMetric.Value = utils.FloatPtr(0)
				}
			case model.Gauge:
				if v.Valid {
					fetchedMetric.Value = utils.FloatPtr(v.Float64)
					fetchedMetric.Delta = utils.Int64Ptr(0)
				}
			default:
				t.Errorf("unknown metric type: %s", tt.metricType)
			}

			passedMetric := model.Metrics{
				ID:    tt.metricName,
				MType: tt.metricType,
				Value: utils.FloatPtr(tt.value),
				Delta: utils.Int64Ptr(tt.delta),
			}

			assert.Equal(t, fetchedMetric, passedMetric)

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)
			if err != nil {
				t.Fatalf("AddMetric() delete metric in DB: %v", err)
			}
		})
	}
}

func TestPGStorage_AddMetrics(t *testing.T) {

	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("load envs: %v", err)
	}

	DSN := os.Getenv("DATABASE_DSN")
	if DSN == "" {
		t.Logf("skip test because of no DSN")
		return
	}

	tests := []struct {
		name    string
		dsn     string
		mode    DBInitMode
		agentID string
		metrics []model.Metrics
		wantErr bool
	}{
		{
			name:    "check the inceremt of a counter metric in DB",
			dsn:     DSN,
			mode:    ManagedExternally,
			agentID: "127.0.0.1_cpu_test",
			metrics: []model.Metrics{
				{
					ID:    "cpu",
					MType: model.Counter,
					Delta: utils.Int64Ptr(1),
				},
				{
					ID:    "cpu",
					MType: model.Counter,
					Delta: utils.Int64Ptr(1),
				},
			},
			wantErr: false,
		},
		{
			name:    "add metrics with an invalid type to DB",
			dsn:     DSN,
			mode:    ManagedExternally,
			agentID: "127.0.0.1_ram_test",
			metrics: []model.Metrics{
				{
					ID:    "cpu",
					MType: "wrong type",
					Delta: utils.Int64Ptr(1),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPGDB(tt.dsn, tt.mode)
			if err != nil {
				t.Fatalf("AddMetrics could not construct DB object: %v", err)
			}

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metrics[0].ID)
			if err != nil {
				t.Fatalf("AddMetrics() delete metric in DB: %v", err)
			}

			gotErr := p.AddMetrics(context.Background(), tt.agentID, tt.metrics)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddMetrics() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddMetrics() succeeded unexpectedly")
			}

			row := p.DB.QueryRow(`
			SELECT delta
			FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metrics[0].ID)

			var d sql.NullInt64
			err = row.Scan(&d)
			if err != nil {
				t.Errorf("AddMetric() row execution: %v", err)
			}

			if !d.Valid {
				t.Errorf("AddMetric() return metric data is invalid: %v", d)
			}
			var wantDelta int64
			for _, m := range tt.metrics {
				wantDelta += *m.Delta
			}
			assert.Equal(t, d.Int64, wantDelta)

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metrics[0].ID)
			if err != nil {
				t.Fatalf("AddMetric() delete metric in DB: %v", err)
			}
		})
	}
}

func TestPGStorage_GetObjMetric(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("GetObjMetric() load envs: %v", err)
	}

	DSN := os.Getenv("DATABASE_DSN")
	if DSN == "" {
		t.Logf("skip test because of no DSN")
		return
	}

	tests := []struct {
		name       string
		dsn        string
		mode       DBInitMode
		agentID    string
		metricType string
		metricName string
		value      float64
		delta      int64
		wantErr    bool
	}{
		{
			name:       "check counter to DB",
			dsn:        DSN,
			mode:       ManagedExternally,
			agentID:    "127.0.0.1_cpu_test",
			metricType: model.Counter,
			metricName: "cpu",
			delta:      1,
			wantErr:    false,
		},
		{
			name:       "check gauge to DB",
			dsn:        DSN,
			mode:       ManagedExternally,
			agentID:    "127.0.0.1_ram_test",
			metricType: model.Gauge,
			metricName: "ram",
			value:      1000,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPGDB(tt.dsn, tt.mode)
			if err != nil {
				t.Fatalf("GetObjMetric() could not construct DB object: %v", err)
			}

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)
			if err != nil {
				t.Fatalf("GetObjMetric() delete metric in DB: %v", err)
			}

			_, err = p.DB.Exec(`
			INSERT INTO metrics (agent_id, id, mtype, delta, value, updated_at) 
			VALUES ($1, $2, $3,
				CASE WHEN $3='counter' THEN $4::bigint ELSE NULL END,
				CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
				now()) 
			`, tt.agentID, tt.metricName, tt.metricType, tt.delta, tt.value)
			if err != nil {
				t.Fatalf("GetObjMetric() insert metric: %v", err)
			}

			value, delta, err := p.GetMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetObjMetric() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetObjMetric() succeeded unexpectedly")
			}

			switch tt.metricType {
			case model.Counter:
				assert.Equal(t, delta, tt.delta)
			case model.Gauge:
				assert.Equal(t, value, tt.value)
			}

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)
			if err != nil {
				t.Fatalf("GetObjMetric() delete metric in DB: %v", err)
			}
		})
	}
}

func TestPGStorage_GetObjMetrics(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("GetObjMetric() load envs: %v", err)
	}

	DSN := os.Getenv("DATABASE_DSN")
	if DSN == "" {
		t.Logf("skip test because of no DSN")
		return
	}

	tests := []struct {
		name       string
		dsn        string
		mode       DBInitMode
		agentID    string
		metricType string
		metricName string
		value      *float64
		delta      *int64
		wantErr    bool
	}{
		{
			name:       "check counter to DB",
			dsn:        DSN,
			mode:       ManagedExternally,
			agentID:    "127.0.0.1_cpu_test",
			metricType: model.Counter,
			metricName: "cpu",
			delta:      utils.Int64Ptr(1),
			wantErr:    false,
		},
		{
			name:       "check gauge to DB",
			dsn:        DSN,
			mode:       ManagedExternally,
			agentID:    "127.0.0.1_ram_test",
			metricType: model.Gauge,
			metricName: "ram",
			value:      utils.FloatPtr(1000),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPGDB(tt.dsn, tt.mode)
			if err != nil {
				t.Fatalf("GetObjMetrics() could not construct DB object: %v", err)
			}

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)
			if err != nil {
				t.Fatalf("GetObjMetrics() delete metric in DB: %v", err)
			}

			_, err = p.DB.Exec(`
			INSERT INTO metrics (agent_id, id, mtype, delta, value, updated_at) 
			VALUES ($1, $2, $3,
				CASE WHEN $3='counter' THEN $4::bigint ELSE NULL END,
				CASE WHEN $3='gauge' THEN $5::double precision ELSE NULL END,
				now()) 
			`, tt.agentID, tt.metricName, tt.metricType, tt.delta, tt.value)
			if err != nil {
				t.Fatalf("GetObjMetric() insert metric: %v", err)
			}

			gotMetric, err := p.GetObjMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("GetObjMetrics() failed: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetObjMetrics() succeeded unexpectedly")
			}

			wantMetric := model.Metrics{
				ID:    tt.metricName,
				MType: tt.metricType,
				Delta: tt.delta,
				Value: tt.value,
			}

			assert.Equal(t, wantMetric, *gotMetric)

			_, err = p.DB.Exec(`
			DELETE FROM metrics
			WHERE agent_id = $1 AND id = $2
			`, tt.agentID, tt.metricName)
			if err != nil {
				t.Fatalf("GetObjMetrics() delete metric in DB: %v", err)
			}
		})
	}
}
