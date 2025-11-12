package service

import (
	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Storage interface {
	AddMetric(agentID string, metricType string, metricName string, value float64, delta int64) error
	GetMetric(agentID string, metricType string, metricName string) (value float64, delta int64, err error)
	GetObjMetric(agentID string, metricType string, metricName string) (*model.Metrics, error)
	GetStore() map[string]*model.Metrics
	PingDB() error
}
