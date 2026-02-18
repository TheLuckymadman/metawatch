package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/go-chi/chi/v5"
)

type fakeService struct{}

func (f *fakeService) AddMetric(
	ctx context.Context,
	name string,
	value string,
	metricType string,
	agentIP string,
) error {
	return nil
}

func (f *fakeService) AddObjMetric(ctx context.Context, metric model.Metrics, agentIP string) error {
	return nil
}
func (f *fakeService) AddObjMetrics(ctx context.Context, metrics []model.Metrics, agentIP string) error {
	return nil
}
func (f *fakeService) GetObjMetric(ctx context.Context, metricReq model.Metrics, agentIP string) (metricResp *model.Metrics, err error) {
	return nil, nil
}
func (f *fakeService) ListMetric(ctx context.Context) (map[string]*model.Metrics, []string, error) {
	return nil, nil, nil
}
func (f *fakeService) PingDB(ctx context.Context) error {
	return nil
}

func ExampleMetricSetterHandler() {
	svc := &fakeService{}

	handler := MetricSetterHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/update/gauge/temperature/42",
		nil,
	)
	req.RemoteAddr = "127.0.0.1:12345"

	r := chi.NewRouter()
	r.Post("/update/{type}/*", handler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))

	// Output:
	// 200
	// {"status":"ok"}
}

func ExampleJSONSetterHandler() {
	svc := &fakeService{}

	handler := JSONSetterHandler(svc)

	reqBody := []byte(`{
		"id": "cpu",
		"type": "counter",
		"delta": 1
	}`)
	body := bytes.NewReader(reqBody)
	req := httptest.NewRequest(
		http.MethodPost,
		"/update/",
		body,
	)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Content-Type", "application/json")

	r := chi.NewRouter()
	r.Post("/update/", handler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))

	// Output:
	// 200
	// {"status":"ok"}
}
