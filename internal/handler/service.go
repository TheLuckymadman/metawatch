package handler

import (
	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Service interface {
	AddMetric(metricName string, metricValue string, metricType string, agentIP string) error
	AddObjMetric(metric model.Metrics, agentIP string) error
	GetMetric(metricName string, metricType string, agentIP string) (result string, err error)
	GetObjMetric(metricReq model.Metrics, agentIP string) (metricResp *model.Metrics, err error)
	ListMetric() (map[string]*model.Metrics, []string, error)
}