package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type simpleSender struct {
	client  *http.Client
	address string
}

type jsonSender struct {
	client   *http.Client
	address  string
	compress bool
	key      string
}

func NewJSONSender(client *http.Client, address string, compress bool, key string) Sender {
	return &jsonSender{client, address, compress, key}
}

func (s *simpleSender) SendMetric(metric model.Metrics) error {
	var url string
	switch metric.MType {
	case model.Counter:
		url = fmt.Sprintf("%s/update/%s/%s/%d", s.address, metric.MType, metric.ID, *metric.Delta)
	case model.Gauge:
		url = fmt.Sprintf("%s/update/%s/%s/%f", s.address, metric.MType, metric.ID, *metric.Value)
	}
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("creating request failed: %v", err)
		return fmt.Errorf("creating request failed: %w", err)
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := s.client.Do(request)
	if err != nil {
		log.Printf("sending metric (ID: %s, Type: %s) failed with: %v", metric.ID, metric.MType, err)
		return fmt.Errorf("sending metric (ID: %s, Type: %s) failed with: %w", metric.ID, metric.MType, err)
	}
	log.Printf("sending metric on %v successfully: %v\n", url, response)
	response.Body.Close()
	return nil
}

func (j *jsonSender) SendMetric(metric model.Metrics) error {
	var url = fmt.Sprintf("%s/update/", j.address)
	metricJSON, err := json.Marshal(&metric)
	if err != nil {
		log.Printf("marshalling metric data failed with error: %v", err)
		return fmt.Errorf("marshalling metric data failed with error: %w", err)
	}
	var body bytes.Buffer
	if j.compress {

		gzipBody := gzip.NewWriter(&body)
		_, err = gzipBody.Write(metricJSON)

		if err != nil {
			return fmt.Errorf("gzip write failed: %w", err)
		}

		if err := gzipBody.Close(); err != nil {
			return fmt.Errorf("gzip close failed: %w", err)
		}
	} else {
		body.Write(metricJSON)
	}

	request, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		log.Printf("creating request failed: %v", err)
		return fmt.Errorf("creating request failed: %w", err)
	}
	if j.compress {
		request.Header.Set("Content-Encoding", "gzip")
	}
	request.Header.Set("Content-Type", "application/json")
	if j.key != "" {
		request.Header.Set("HashSHA256", generateHMAC(j.key, metricJSON))

	}
	response, err := j.client.Do(request)
	if err != nil {
		log.Printf("sending metric (ID: %s, Type: %s) failed with: %v", metric.ID, metric.MType, err)
		return fmt.Errorf("sending metric (ID: %s, Type: %s) failed with: %w", metric.ID, metric.MType, err)
	}
	defer response.Body.Close()
	log.Printf("sending metric on %v with status: %v\n", url, response.Status)

	return nil
}

func (j *jsonSender) SendMetrics(metrics []model.Metrics) error {
	var url = fmt.Sprintf("%s/update/", j.address)
	metricsJSON, err := json.Marshal(&metrics)
	if err != nil {
		log.Printf("marshalling metrics data failed with error: %v", err)
		return fmt.Errorf("marshalling metrics data failed with error: %w", err)
	}
	var body bytes.Buffer
	if j.compress {

		gzipBody := gzip.NewWriter(&body)
		_, err = gzipBody.Write(metricsJSON)

		if err != nil {
			return fmt.Errorf("gzip write failed: %w", err)
		}

		if err := gzipBody.Close(); err != nil {
			return fmt.Errorf("gzip close failed: %w", err)
		}
	} else {
		body.Write(metricsJSON)
	}

	request, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		log.Printf("creating request failed: %v", err)
		return fmt.Errorf("creating request failed: %w", err)
	}
	if j.compress {
		request.Header.Set("Content-Encoding", "gzip")
	}
	request.Header.Set("Content-Type", "application/json")
	if j.key != "" {
		request.Header.Set("HashSHA256", generateHMAC(j.key, metricsJSON))

	}
	response, err := j.client.Do(request)
	if err != nil {
		log.Printf("sending the metrics batch failed with: %v", err)
		return fmt.Errorf("sending the metrics batch failed with: %w", err)
	}
	defer response.Body.Close()
	log.Printf("sending metrics batch on %v with status: %v", url, response.Status)

	return nil
}

func generateHMAC(key string, body []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err != nil {
		log.Printf("HMAC generation failed: %v", err)
	}
	result := h.Sum(nil)
	return hex.EncodeToString(result)
}
