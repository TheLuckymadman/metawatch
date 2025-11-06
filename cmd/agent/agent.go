package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/agent"
	"github.com/TheLuckymadman/metawatch/internal/config/agentconfig"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

var (
	serverURL     string
	pollInterval   int
	reportInterval int
)

func main() {
	cfg := agentconfig.Load()
	serverURL = cfg.ServerURL
	pollInterval = cfg.PollInterval
	reportInterval = cfg.ReportInterval
	
	log.Printf("Start agent with the following params:\nserverUrl: %s, pollInterval: %d, reportInterval: %d", serverURL, pollInterval, reportInterval)

	lm := agent.LocalMetrics{M: make([]model.Metrics, 0, 28), PollCount: utils.Int64Ptr(0)}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			default:
				lm.GetMetrics()
				time.Sleep(time.Duration(pollInterval) * time.Second)
			}
		}
	}()
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			default:
				lm.ReadMetrics()
				time.Sleep(time.Duration(reportInterval) * time.Second)
			}
		}
	}()
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			default:
				lm.SendMetrics(serverURL)
				time.Sleep(time.Duration(reportInterval) * time.Second)
			}
		}
	}()

	<- ctx.Done()
	log.Printf("Agent is shutting down gracefully")
}
