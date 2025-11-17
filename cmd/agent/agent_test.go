package main

import (
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/agent"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

func TestGetMetrics(t *testing.T) {
	lm := agent.LocalMetrics{M: make([]model.Metrics, 0, 28), PollCount: utils.Int64Ptr(0)}
	lm.GetMetrics()
	if len(lm.M) == 0 {
		t.Errorf("No metrics to be collected")
	}
	if *lm.PollCount != 1 {
		t.Errorf("Expected pollCount to be incremented, got %d", *lm.PollCount)
	}
	expected := []string{"Alloc", "BuckHashSys", "Frees", "PollCount", "RandomValue"}
	for _, name := range expected {
		found := false
		for i := range lm.M {
			if lm.M[i].ID == name {
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
	lm := agent.LocalMetrics{
		M: []model.Metrics{
			{
			ID: "test_metric",
			MType: model.Counter,
			Delta: utils.Int64Ptr(1),
			Value: nil, 
			},
		},
		PollCount: utils.Int64Ptr(1),
	}

	lm.SendMetrics("localhost:8080", 100)

	if len(lm.M) == 0 {
		t.Errorf("Expected the metrics be kept on send failure")
	}
}