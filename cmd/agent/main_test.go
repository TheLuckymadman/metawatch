package main

import (
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

func TestGetMetrics(t *testing.T) {
	lm := localMetrics{m: make([]model.Metrics, 0, 28), pollCount: int64Ptr(0)}
	lm.getMetrics()
	if len(lm.m) == 0 {
		t.Errorf("No metrics to be collected")
	}
	if *lm.pollCount != 1 {
		t.Errorf("Expected pollCount to be incremented, got %d", *lm.pollCount)
	}
	expected := []string{"Alloc", "BuckHashSys", "Frees", "PollCount", "RandomValue"}
	for _, name := range expected {
		found := false
		for i := range lm.m {
			if lm.m[i].ID == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected metric %s not found", name)
		} 
	}
}

func TestSendMetrics(t *testing.T) {
	lm := localMetrics{
		m: []model.Metrics{
			{
			ID: "test_metric",
			MType: model.Counter,
			Delta: int64Ptr(1),
			Value: nil, 
			},
		},
		pollCount: int64Ptr(1),
	}

	lm.sendMetrics("localhost:8080")

	if len(lm.m) == 0 {
		t.Errorf("Expected the metrics be kept on send failure")
	}
}