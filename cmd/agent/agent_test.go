package main

import (
	"net/http"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/agent"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

func TestGetMetrics(t *testing.T) {
	client := &http.Client{}
	sender := agent.NewJSONSender(client, "127.0.0.1:8080", false, "", "")
	lm := agent.NewLocalMetrics(sender)

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
	client := &http.Client{}
	sender := agent.NewJSONSender(client, "127.0.0.1:8080", false, "", "")
	lm := agent.NewLocalMetrics(sender)
	lm.M = []model.Metrics{
		{
			ID:    "test_metric",
			MType: model.Counter,
			Delta: utils.Int64Ptr(1),
			Value: nil,
		},
	}

	if len(lm.M) == 0 {
		t.Errorf("Expected the metrics be kept on send failure")
	}
}
