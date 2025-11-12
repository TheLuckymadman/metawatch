package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type PGStorage struct {
	DB *sql.DB
}

func NewPGDB(dsn string) (*PGStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return &PGStorage{db}, nil
}

func (p *PGStorage) Close() error {
	return p.DB.Close()
}

func (p *PGStorage) PingDB() error {
	if err := p.DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}
	return nil
}

func (p *PGStorage) AddMetric(agentID string, metricType string, metricName string, value float64, delta int64) error {
	return nil
}
	
func (p *PGStorage) GetMetric(agentID string, metricType string, metricName string) (value float64, delta int64, err error) {
	return 0, 0, nil
}
	
func (p *PGStorage) GetObjMetric(agentID string, metricType string, metricName string) (*model.Metrics, error) {
	return nil, nil
}
	
func (p *PGStorage) GetStore() map[string]*model.Metrics {
	return nil
}