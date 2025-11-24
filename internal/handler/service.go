package handler

import (
	"context"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Service interface {
	AddMetric(ctx context.Context, metricName string, metricValue string, metricType string, agentIP string) error
	AddObjMetric(ctx context.Context, metric model.Metrics, agentIP string) error
	AddObjMetrics(ctx context.Context, metrics []model.Metrics, agentIP string) error
	GetObjMetric(ctx context.Context, metricReq model.Metrics, agentIP string) (metricResp *model.Metrics, err error)
	ListMetric(ctx context.Context) (map[string]*model.Metrics, []string, error)
	PingDB(ctx context.Context) error
}