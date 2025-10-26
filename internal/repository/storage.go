package repository

type Storage interface {
	SetMetric(agentID string, metricType string, metricName string, value float64, delta int64) error
	GetMetric(agentID string, metricType string, metricName string) (value float64, delta int64, err error)
}
