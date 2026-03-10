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

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	cfg := agentconfig.Load()
	//log.Printf("Start agent with the following params:\nserverUrl: %s, pollInterval: %d, reportInterval: %d", cfg.ServerURL, cfg.PollInterval, cfg.ReportInterval)
	logger.Info("Start agent with the following params",
		zap.String("serverUrl", cfg.ServerURL),
		zap.Duration("pollInterval", time.Duration(cfg.PollInterval)),
		zap.Duration("reportInterval", time.Duration(cfg.ReportInterval)),
		zap.String("Build version", buildVersion),
		zap.String("Build date", buildDate),
		zap.String("Build commit", buildCommit),
	)

	client := &http.Client{}
	sender := agent.NewJSONSender(client, cfg.ServerURL, cfg.Compress, cfg.Key, cfg.CryptoKey)
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
		defer func() {
			close(metricsQueue)
			wg.Done()
		}()
		lm.StartBatching(ctx, time.Duration(cfg.ReportInterval), cfg.BatchSize, metricsQueue, failedMetrics)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		lm.GetExtraMetrics(ctx, time.Duration(cfg.PollInterval))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(cfg.PollInterval))
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
			ticker := time.NewTicker(time.Duration(cfg.ReportInterval))
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
