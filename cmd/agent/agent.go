package main

import (
	"flag"
	"log"
	"time"
	"strings"

	"github.com/TheLuckymadman/metawatch/internal/agent"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

var (
	serverURL     *string
	reportInterval *int
	pollInterval   *int
)

func main() {
	serverURL = flag.String("a", "http://127.0.0.1:8080", "server endpoint address in the format http://server:port")
	reportInterval = flag.Int("r", 2, "sending metrics frequency")
	pollInterval = flag.Int("p", 10, "fetching metrics frequency")
	flag.Parse()
	
	if !strings.HasPrefix(*serverURL, "http://") && !strings.HasPrefix(*serverURL, "https://") {
        *serverURL = "http://" + *serverURL
    }

	log.Printf("Start agent with the following params:\nserverUrl: %s, pollInterval: %d, reportInterval: %d", *serverURL, *pollInterval, *reportInterval)

	lm := agent.LocalMetrics{M: make([]model.Metrics, 0, 28), PollCount: utils.Int64Ptr(0)}

	go func() {
		for {
			lm.GetMetrics()
			time.Sleep(time.Duration(*pollInterval) * time.Second)
		}
	}()
	go func() {
		for {
			lm.ReadMetrics()
			time.Sleep(time.Duration(*reportInterval) * time.Second)
		}
	}()
	go func() {
		for {
			lm.SendMetrics(*serverURL)
			time.Sleep(time.Duration(*reportInterval) * time.Second)
		}
	}()

	select {}
}
