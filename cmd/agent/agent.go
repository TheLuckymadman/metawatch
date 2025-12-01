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
				lm.SendMetrics(cfg.BatchSize)
				time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
			}
		}
	}()

	<-ctx.Done()
	//log.Printf("Agent is shutting down gracefully")
	logger.Info("Agent is shutting down gracefully")
}
