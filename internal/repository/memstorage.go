package repository

import (
	"fmt"
	"sync"
	"context"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type MemStorage struct {
	Metrics map[string]*model.Metrics
	sync.RWMutex
}

func (m *MemStorage) GetStore(ctx context.Context) (map[string]*model.Metrics, error) {
	m.RLock()
	defer m.RUnlock()

	copyMemStorage := make(map[string]*model.Metrics, len(m.Metrics))
	for k, v := range m.Metrics {
		copyMemStorage[k] = v
	}
	return copyMemStorage, nil

}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Metrics: make(map[string]*model.Metrics),
	}
}

func (m *MemStorage) Close() error {
	return nil
}

func (m *MemStorage) AddMetric(ctx context.Context, agentID string, metricType string, metricName string, value float64, delta int64) error {
	key := agentID + "_" + metricName

	m.Lock()
	metric, ok := m.Metrics[key]
	if !ok {
		m.Metrics[key] = &model.Metrics{
			ID: metricName,
			MType: metricType,
		}
		metric = m.Metrics[key]
	}
	m.Unlock()
	// defer metric.Unlock()

	// metric.Lock()
	switch metricType {
	case model.Counter: {
		if metric.Delta == nil {
			metric.Delta = new(int64)
		}
		*metric.Delta += delta
		metric.Value = nil
	}
	
	case model.Gauge: {
		metric.Delta = nil
		metric.Value = &value
	}
	}
	
	return nil
}

func (m *MemStorage) GetMetric(ctx context.Context, agentID string, metricType string, metricName string) (value float64, delta int64, err error) {
	key := agentID + "_" + metricName

	m.RLock()
	agentMetrics, ok := m.Metrics[key]
	m.RUnlock()

	if !ok {
		return 0, 0, fmt.Errorf("metric not found, key: %q", key)
	}

	// agentMetrics.RLock()
	// defer agentMetrics.RUnlock()

	if agentMetrics.MType == model.Gauge && agentMetrics.Value != nil {
		return *agentMetrics.Value, 0, nil
	}
	if agentMetrics.MType == model.Counter && agentMetrics.Delta != nil {
		return 0, *agentMetrics.Delta, nil
	}

	return 0, 0, fmt.Errorf("metric %q has no value", key)
}

func (m *MemStorage) GetObjMetric(ctx context.Context, agentID string, metricType string, metricName string) (*model.Metrics, error) {
	key := agentID + "_" + metricName

	m.RLock()
	agentMetrics, ok := m.Metrics[key]
	m.RUnlock()

	if !ok {
		return nil, fmt.Errorf("metric not found, key: %q", key)
	}
	return  agentMetrics, nil
}

func (m *MemStorage) PingDB(ctx context.Context) error {
	return nil
}
