package grpc

import (
	"context"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

// Service interface is used to call methods of the metric service
type Service interface {
	AddObjMetrics(ctx context.Context, metrics []model.Metrics, agentIP string) error
}
