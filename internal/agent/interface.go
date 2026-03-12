package agent

import (
	"context"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Sender interface {
	SendMetrics(ctx context.Context, metric []model.Metrics, localIP string) error
	Close() error
}
