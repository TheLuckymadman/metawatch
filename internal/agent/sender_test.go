package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
	"github.com/stretchr/testify/assert"
)

var metrics = []model.Metrics{
	{
		ID:    "cpu_count",
		MType: model.Counter,
		Delta: utils.Int64Ptr(1),
	},
	{
		ID:    "ram_gauge",
		MType: model.Gauge,
		Value: utils.FloatPtr(1.2),
	},
}

func SimpleSenderHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	pathItems := strings.Split(path, "/")
	if len(pathItems) != 5 {
		http.Error(
			w,
			fmt.Sprintf("wrong path parameters count in url path: %s, want: %d, got: %d", path, 5, len(pathItems)),
			http.StatusBadRequest,
		)
		return
	}
	metricType := pathItems[2]
	if metricType != model.Counter && metricType != model.Gauge {
		http.Error(w, fmt.Sprintf("wrong metric type: %s", metricType), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var reply = struct {
		Status string `json:"status"`
	}{Status: "ok"}
	body, err := json.Marshal(reply)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func JSONSenderHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	match, err := regexp.MatchString("^/update/$", path)
	if err != nil {
		msg := fmt.Sprintf("match path: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if !match {
		msg := fmt.Sprintf("wrong path, want: /update/, got: %s", path)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		msg := fmt.Sprintf("read body: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if strings.HasPrefix(string(body), "{") {
		fmt.Println("Metric has model.Metrics type")
		var metric model.Metrics
		err = json.Unmarshal(body, &metric)
		if err != nil {
			msg := fmt.Sprintf("umarshal body: %v", err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	} else if strings.HasPrefix(string(body), "[") {
		fmt.Println("Metric has []model.Metrics type")
		var metric []model.Metrics
		err = json.Unmarshal(body, &metric)
		if err != nil {
			msg := fmt.Sprintf("umarshal body: %v", err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	} else {
		msg := fmt.Sprintf("bad format body %s", string(body))
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var reply = struct {
		Status string `json:"status"`
	}{Status: "ok"}
	body, err = json.Marshal(reply)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func TestSendSimpleMetric(t *testing.T) {
	handler := http.HandlerFunc(SimpleSenderHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	sender := NewSimpleSender(client, server.URL)

	for _, m := range metrics {
		m := m
		t.Run(m.ID, func(t *testing.T) {
			err := sender.SendMetric(m)
			assert.NoError(t, err)
		})
	}
}

func TestSendJSONMetric(t *testing.T) {
	handler := http.HandlerFunc(JSONSenderHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()
	sender := NewJSONSender(client, server.URL, false, "")

	t.Log("test SendMetric")
	for _, m := range metrics {
		m := m
		t.Run(m.ID, func(t *testing.T) {
			err := sender.SendMetric(context.Background(), m)
			assert.NoError(t, err)
		})
	}

	t.Log("test SendMtrics")
	err := sender.SendMetrics(context.Background(), metrics)
	assert.NoError(t, err)
}
