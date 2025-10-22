package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)


type LocalMetrics struct {
	M         []model.Metrics
	PollCount *int64
	sync.RWMutex
}

func (lm *LocalMetrics) GetMetrics() {
	var memstat runtime.MemStats
	runtime.ReadMemStats(&memstat)
	lm.Lock()
	lm.M = append(lm.M, model.Metrics{ID: "Alloc", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.Alloc))})
	lm.M = append(lm.M, model.Metrics{ID: "BuckHashSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.BuckHashSys))})
	lm.M = append(lm.M, model.Metrics{ID: "Frees", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.Frees))})
	lm.M = append(lm.M, model.Metrics{ID: "GCCPUFraction", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.GCCPUFraction))})
	lm.M = append(lm.M, model.Metrics{ID: "GCSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.GCSys))})
	lm.M = append(lm.M, model.Metrics{ID: "HeapAlloc", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.HeapAlloc))})
	lm.M = append(lm.M, model.Metrics{ID: "HeapIdle", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.HeapIdle))})
	lm.M = append(lm.M, model.Metrics{ID: "HeapInuse", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.HeapInuse))})
	lm.M = append(lm.M, model.Metrics{ID: "HeapObjects", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.HeapObjects))})
	lm.M = append(lm.M, model.Metrics{ID: "HeapReleased", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.HeapReleased))})
	lm.M = append(lm.M, model.Metrics{ID: "HeapSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.HeapSys))})
	lm.M = append(lm.M, model.Metrics{ID: "LastGC", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.LastGC))})
	lm.M = append(lm.M, model.Metrics{ID: "Lookups", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.Lookups))})
	lm.M = append(lm.M, model.Metrics{ID: "MCacheInuse", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.MCacheInuse))})
	lm.M = append(lm.M, model.Metrics{ID: "MCacheSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.MCacheSys))})
	lm.M = append(lm.M, model.Metrics{ID: "MSpanInuse", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.MSpanInuse))})
	lm.M = append(lm.M, model.Metrics{ID: "MSpanSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.MSpanSys))})
	lm.M = append(lm.M, model.Metrics{ID: "Mallocs", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.Mallocs))})
	lm.M = append(lm.M, model.Metrics{ID: "NextGC", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.NextGC))})
	lm.M = append(lm.M, model.Metrics{ID: "NumForcedGC", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.NumForcedGC))})
	lm.M = append(lm.M, model.Metrics{ID: "NumGC", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.NumGC))})
	lm.M = append(lm.M, model.Metrics{ID: "OtherSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.OtherSys))})
	lm.M = append(lm.M, model.Metrics{ID: "PauseTotalNs", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.PauseTotalNs))})
	lm.M = append(lm.M, model.Metrics{ID: "StackInuse", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.StackInuse))})
	lm.M = append(lm.M, model.Metrics{ID: "StackSys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.StackSys))})
	lm.M = append(lm.M, model.Metrics{ID: "Sys", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.Sys))})
	lm.M = append(lm.M, model.Metrics{ID: "TotalAlloc", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(float64(memstat.TotalAlloc))})
	*lm.PollCount += 1
	lm.M = append(lm.M, model.Metrics{ID: "PollCount", MType: model.Counter, Delta: utils.Int64Ptr(*lm.PollCount), Value: nil})
	lm.M = append(lm.M, model.Metrics{ID: "RandomValue", MType: model.Gauge, Delta: nil, Value: utils.FloatPtr(rand.Float64())})
	lm.Unlock()
}

func (lm *LocalMetrics) SendMetrics(s string) {
	client := &http.Client{}
	var url string
	lm.RLock()
	copyMetrics := lm.M
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
	lm.M = nil
	lm.M = append(lm.M, nonsentMetrics...)
	lm.Unlock()
}

func (lm *LocalMetrics) ReadMetrics() {
	lm.RLock()
	for k, v := range lm.M {
		switch v.MType {
		case model.Counter:
			log.Printf("iter %d, ID: %s, Type: %s, Delta: %d\n", k, v.ID, v.MType, *v.Delta)
		case model.Gauge:
			log.Printf("iter %d, ID: %s, Type: %s, Values: %f\n", k, v.ID, v.MType, *v.Value)
		}
	}
	lm.RUnlock()
}