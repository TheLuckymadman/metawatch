package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type localMetrics struct {
	m         []model.Metrics
	pollCount *int64
	sync.RWMutex
}

func (lm *localMetrics) getMetrics() {
	var memstat runtime.MemStats
	runtime.ReadMemStats(&memstat)
	lm.Lock()
	lm.m = append(lm.m, model.Metrics{ID: "Alloc", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Alloc))})
	lm.m = append(lm.m, model.Metrics{ID: "BuckHashSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.BuckHashSys))})
	lm.m = append(lm.m, model.Metrics{ID: "Frees", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "GCCPUFraction", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "GCSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "HeapAlloc", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "HeapIdle", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "HeapInuse", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "HeapObjects", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "HeapReleased", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "HeapSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "LastGC", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "Lookups", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "MCacheInuse", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "MCacheSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "MSpanInuse", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "MSpanSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "Mallocs", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "NextGC", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "NumForcedGC", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "NumGC", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "OtherSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "PauseTotalNs", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "StackInuse", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "StackSys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "Sys", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	lm.m = append(lm.m, model.Metrics{ID: "TotalAlloc", MType: model.Gauge, Delta: nil, Value: floatPtr(float64(memstat.Frees))})
	*lm.pollCount += 1
	lm.m = append(lm.m, model.Metrics{ID: "PollCount", MType: model.Counter, Delta: int64Ptr(*lm.pollCount), Value: nil})
	lm.m = append(lm.m, model.Metrics{ID: "RandomValue", MType: model.Gauge, Delta: nil, Value: floatPtr(rand.Float64())})
	lm.Unlock()
}

func (lm *localMetrics) sendMetrics(s string) {
	client := &http.Client{}
	var url string
	lm.RLock()
	copyMetrics := lm.m
	var nonsentMetrics []model.Metrics
	lm.RUnlock()
	for i := range copyMetrics {
		switch copyMetrics[i].MType {
		case model.Counter:
			url = fmt.Sprintf("%s/update/%s/%s/%d", s, copyMetrics[i].MType, copyMetrics[i].ID, *copyMetrics[i].Delta)
		case model.Gauge:
			url = fmt.Sprintf("%s/update/%s/%s/%f", s, copyMetrics[i].MType, copyMetrics[i].ID, *copyMetrics[i].Value)
		}
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Printf("Create request failed: %v", err)
			nonsentMetrics = append(nonsentMetrics, copyMetrics[i])
			continue
		}
		request.Header.Set("Content-Type", "text/plain")
		response, err := client.Do(request)
		if err != nil {
			log.Printf("Send metric (ID: %s, Type: %s) failed with: %v\n", copyMetrics[i].ID, copyMetrics[i].MType, err.Error())
			nonsentMetrics = append(nonsentMetrics, copyMetrics[i])
			continue

		}
		log.Printf("Send metric on %v successfully: %v\n", url, response)
		response.Body.Close()
	}
	lm.Lock()
	lm.m = nil
	lm.m = append(lm.m, nonsentMetrics...)
	lm.Unlock()
}

func (lm *localMetrics) readMetrics() {
	lm.RLock()
	for k, v := range lm.m {
		switch v.MType {
		case model.Counter:
			log.Printf("iter %d, ID: %s, Type: %s, Delta: %d\n", k, v.ID, v.MType, *v.Delta)
		case model.Gauge:
			log.Printf("iter %d, ID: %s, Type: %s, Values: %f\n", k, v.ID, v.MType, *v.Value)
		}
	}
	lm.RUnlock()
}

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func setIntParams(arg string, def int) int {
	n, err := strconv.Atoi(arg)
	if err != nil {
		log.Printf("Use non-integer argument %v, skip and use default %d sec value", arg, def)
	} else {
		return n
	}
	return def
}

func main() {
	serverURL := "http://127.0.0.1:8080"
	pollInterval := 2
	reportInterval := 10
	args := os.Args
	if len(args) == 2 {
		serverURL = args[1]
	}
	if len(args) == 3 {
		serverURL = args[1]
		pollInterval = setIntParams(args[2], pollInterval)
	}
	if len(args) == 4 {
		serverURL = args[1]
		pollInterval = setIntParams(args[2], pollInterval)
		reportInterval = setIntParams(args[3], reportInterval)
	}

	log.Printf("Start agent with the following params:\nserverUrl: %s, pollInterval: %d, reportInterval: %d", serverURL, pollInterval, reportInterval)

	lm := localMetrics{m: make([]model.Metrics, 0, 28), pollCount: int64Ptr(0)}

	go func() {
		for {
			lm.getMetrics()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()
	go func() {
		for {
			lm.readMetrics()
			time.Sleep(time.Duration(reportInterval) * time.Second)
		}
	}()
	go func() {
		for {
			lm.sendMetrics(serverURL)
			time.Sleep(time.Duration(reportInterval) * time.Second)
		}
	}()
	
	select {}
}
