package grpc

import (
	"context"
	"fmt"
	"log"

	"github.com/TheLuckymadman/metawatch/internal/model"
	pb "github.com/TheLuckymadman/metawatch/internal/proto"
	"github.com/TheLuckymadman/metawatch/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricServer struct {
	pb.UnimplementedMetricsServer
	s Service
}

func NewMetricServer(svc Service) *MetricServer {
	return &MetricServer{s: svc}
}

func (ms *MetricServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	agentIP, err := getIP(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("grpc request from %s\n", agentIP)

	protoMetrics := in.GetMetrics()
	metrics := make([]model.Metrics, len(protoMetrics))

	for i, m := range protoMetrics {
		switch m.GetType() {
		case pb.Metric_COUNTER:
			{
				metrics[i] = model.Metrics{
					ID:    m.GetId(),
					MType: model.Counter,
					Delta: utils.Int64Ptr(m.GetDelta()),
					Value: utils.FloatPtr(m.GetValue()),
				}
			}
		case pb.Metric_GAUGE:
			{
				metrics[i] = model.Metrics{
					ID:    m.GetId(),
					MType: model.Gauge,
					Delta: utils.Int64Ptr(m.GetDelta()),
					Value: utils.FloatPtr(m.GetValue()),
				}
			}
		default:
			return nil, status.Errorf(codes.FailedPrecondition, "grpc MetricServer, unknown metric type %v", m.GetType())
		}
	}
	if err = ms.s.AddObjMetrics(ctx, metrics, agentIP); err != nil {
		log.Printf("grpc MetricServer, add metrics failed: %v", err)
		return nil, status.Errorf(codes.Internal, "add metrics failed: %v", err)
	}
	return &pb.UpdateMetricsResponse{}, nil
}
