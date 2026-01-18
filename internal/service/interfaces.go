package service

import (
	"context"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Storage interface {
	AddMetric(ctx context.Context, agentID string, metricType string, metricName string, value float64, delta int64) error
	GetMetric(ctx context.Context, agentID string, metricType string, metricName string) (value float64, delta int64, err error)
	GetObjMetric(ctx context.Context, agentID string, metricType string, metricName string) (*model.Metrics, error)
	AddMetrics(ctx context.Context, agentID string, metrics []model.Metrics) error
	GetStore(ctx context.Context) (map[string]*model.Metrics, error)
	PingDB(ctx context.Context) error
	Close() error
}

type Observer interface {
	Update(ctx context.Context, metrics []model.Metrics, agentIP string) error
	GetID() string
}

type Audit interface {
	Update(ctx context.Context, metrics []model.Metrics, agentIP string) error
	GetID() string
}
