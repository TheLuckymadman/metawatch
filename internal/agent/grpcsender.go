package agent

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/TheLuckymadman/metawatch/internal/model"
	pb "github.com/TheLuckymadman/metawatch/internal/proto"
)

type grpcSender struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn
}

func NewGRPCSender(address string) (*grpcSender, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("new grpc sender error: %w", err)
	}

	c := pb.NewMetricsClient(conn)

	return &grpcSender{client: c, conn: conn}, nil
}

func (s *grpcSender) SendMetrics(ctx context.Context, metric []model.Metrics, localIP string) error {
	metrics := make([]*pb.Metric, len(metric))

	for i, m := range metric {
		switch m.MType {
		case model.Counter:
			{
				protMetric := pb.Metric{}
				protMetric.Id = m.ID
				protMetric.Type = pb.Metric_COUNTER
				protMetric.Delta = *m.Delta
				metrics[i] = &protMetric
			}
		case model.Gauge:
			{
				protMetric := pb.Metric{}
				protMetric.Id = m.ID
				protMetric.Type = pb.Metric_GAUGE
				protMetric.Value = *m.Value
				metrics[i] = &protMetric
			}
		default:
			return fmt.Errorf("grpcSender, unknown metric type %s", m.MType)
		}
	}

	req := pb.UpdateMetricsRequest{}
	req.Metrics = metrics
	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", localIP)
	_, err := s.client.UpdateMetrics(ctx, &req)
	if err != nil {
		return fmt.Errorf("grpcSender, grpc response error: %w", err)
	}
	log.Printf("grpcSender, sent %d metrics successfully\n", len(metrics))

	return nil
}

func (s *grpcSender) Close() error {
	err := s.conn.Close()
	if err != nil {
		return fmt.Errorf("error on grpc connection close: %w", err)
	}
	return nil
}
