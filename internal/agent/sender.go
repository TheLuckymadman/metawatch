package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Sender interface {
	SendMetric(client *http.Client, s string, metric model.Metrics) error
}

func SendMetric(client *http.Client, s string, metric model.Metrics) error {
	var url string
	switch metric.MType {
	case model.Counter:
		url = fmt.Sprintf("%s/update/%s/%s/%d", s, metric.MType, metric.ID, *metric.Delta)
	case model.Gauge:
		url = fmt.Sprintf("%s/update/%s/%s/%f", s, metric.MType, metric.ID, *metric.Value)
	}
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("creating request failed: %v", err)
		return fmt.Errorf("creating request failed: %v", err)
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := client.Do(request)
	if err != nil {
		log.Printf("sending metric (ID: %s, Type: %s) failed with: %v", metric.ID, metric.MType, err)
		return fmt.Errorf("sending metric (ID: %s, Type: %s) failed with: %v", metric.ID, metric.MType, err)
	}
	log.Printf("sending metric on %v successfully: %v\n", url, response)
	response.Body.Close()
	return nil
}

func SendObjMetrics(client *http.Client, s string, metric model.Metrics) error {
	var url = fmt.Sprintf("%s/update/", s)
	metricJSON, err := json.Marshal(&metric)
	if err != nil {
		log.Printf("marshalling metric data failed with error: %v\n", err)
		return fmt.Errorf("marshalling metric data failed with error: %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metricJSON))
	if err != nil {
		log.Printf("creating request failed: %v", err)
		return fmt.Errorf("creating request failed: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Printf("sending metric (ID: %s, Type: %s) failed with: %v", metric.ID, metric.MType, err)
		return fmt.Errorf("sending metric (ID: %s, Type: %s) failed with: %v", metric.ID, metric.MType, err)
	}
	log.Printf("sending metric on %v successfully: %v\n", url, response)
	response.Body.Close()
	return nil
}