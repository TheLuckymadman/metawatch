package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/agent"
	"github.com/TheLuckymadman/metawatch/internal/config/agentconfig"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
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
	lm := agent.LocalMetrics{M: make([]model.Metrics, 0, 28), PollCount: utils.Int64Ptr(0)}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lm.ReadMetrics()
				time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lm.SendMetrics(cfg.ServerURL, cfg.BatchSize)
				time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
			}
		}
	}()

	<-ctx.Done()
	//log.Printf("Agent is shutting down gracefully")
	logger.Info("Agent is shutting down gracefully")
}
