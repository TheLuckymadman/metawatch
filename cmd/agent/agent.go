package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/agent"
	"github.com/TheLuckymadman/metawatch/internal/config/agentconfig"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := agentconfig.Load()
	//log.Printf("Start agent with the following params:\nserverUrl: %s, pollInterval: %d, reportInterval: %d", cfg.ServerURL, cfg.PollInterval, cfg.ReportInterval)
	logger.Info("Start agent with the following params",
		zap.String("serverUrl", cfg.ServerURL),
		zap.Int("pollInterval", cfg.PollInterval),
		zap.Int("reportInterval", cfg.ReportInterval),
	)

	client := &http.Client{}
	sender := agent.NewJSONSender(client, cfg.ServerURL, cfg.Compress, cfg.Key)
	lm := agent.NewLocalMetrics(sender)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metricsQueue := make(chan []model.Metrics, cfg.RateLimit)
	failedMetrics := make(chan []model.Metrics, cfg.BatchSize)

	for i := 0; i < cfg.RateLimit; i++ {
		go lm.MetricsSender(i, ctx, metricsQueue, failedMetrics)
	}
	go lm.StartBatching(ctx, cfg.ReportInterval, cfg.BatchSize, metricsQueue, failedMetrics)
	go lm.GetExtraMetrics(ctx, cfg.PollInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lm.GetMetrics()
				time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			}
		}
	}()
	// go func() {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			lm.ReadMetrics()
	// 			time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
	// 		}
	// 	}
	// }()

	<-ctx.Done()
	close(failedMetrics)
	//log.Printf("Agent is shutting down gracefully")
	logger.Info("Agent is shutting down gracefully")
}
