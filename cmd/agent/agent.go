package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	_ "net/http/pprof"

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

	var wg sync.WaitGroup
	metricsQueue := make(chan []model.Metrics, cfg.RateLimit)
	failedMetrics := make(chan []model.Metrics, cfg.BatchSize)

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			lm.MetricsSender(id, ctx, metricsQueue, failedMetrics)
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		lm.StartBatching(ctx, cfg.ReportInterval, cfg.BatchSize, metricsQueue, failedMetrics)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lm.GetExtraMetrics(ctx, cfg.PollInterval)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lm.GetMetrics()
			}
		}
	}()

	if cfg.LogMetrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					lm.ReadMetrics()
				}
			}
		}()
	}

	// http.ListenAndServe("127.0.0.1:9090", nil)

	<-ctx.Done()
	wg.Wait()
	fmt.Println("Goroutines:", runtime.NumGoroutine())
	logger.Info("Agent is shutting down gracefully")
}
