package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type simpleSender struct {
	client  *http.Client
	address string
}

func NewSimpleSender(client *http.Client, address string) *simpleSender {
	return &simpleSender{client: client, address: address}
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
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("read response body: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf(
			"sending metric (ID: %s, Type: %s) failed with status code: %d\nwith body: %s",
			metric.ID,
			metric.MType,
			response.StatusCode,
			string(body),
		)
		log.Println(msg)
		return fmt.Errorf("%s", msg)
	}
	log.Printf("sending metric to %v with the status: %v\nwith body: %s", url, response.Status, body)

	return nil
}

func (s *simpleSender) Close() error {
	return nil
}
