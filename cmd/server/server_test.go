package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/TheLuckymadman/metawatch/internal/handler"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/repository"
	"github.com/TheLuckymadman/metawatch/internal/service"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

func TestServer(t *testing.T) {
	type want struct {
		code int
		response string
	}

	tests := []struct {
		name string
		want want
		metric model.Metrics
	}{
		{
			name: "Send counter",
			want: want{
				code: 200,
				response: `{"status":"ok"}`, 
			},
			metric: model.Metrics{
				ID: "test_counter1",
				MType: model.Counter,
				Delta: utils.Int64Ptr(1),
			},
		},
		{
			name: "Send gauge",
			want: want{
				code: 200,
				response: `{"status":"ok"}`, 
			},
			metric: model.Metrics{
				ID: "test_gauge1",
				MType: model.Gauge,
				Value: utils.FloatPtr(20.001),
			},
		},
	}

	s := repository.NewMemStorage()
	srv := service.NewService(s)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T){
			r := chi.NewRouter()
			r.Post("/update/{type}/*", handler.MetricSetterHandler(srv))
			r.Get("/value/*", handler.MetricGetterHandler(srv))

			// set metrics
			var path string
			switch test.metric.MType {
			case model.Counter: path = fmt.Sprintf("/update/%v/%v/%d", test.metric.MType, test.metric.ID, *test.metric.Delta)
			case model.Gauge: path = fmt.Sprintf("/update/%v/%v/%f", test.metric.MType, test.metric.ID, *test.metric.Value)
			}
			req := httptest.NewRequest(http.MethodPost, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
	
			assert.Equal(t, test.want.code, result.StatusCode)
			body, _ := io.ReadAll(result.Body)
			assert.JSONEq(t, test.want.response, string(body))

			result.Body.Close()
			
			// get metrics
			path = fmt.Sprintf("/value/%v/%v", test.metric.MType, test.metric.ID)
			w = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(w, req)
			result = w.Result()

			body, _ = io.ReadAll(result.Body)
			t.Log(string(body))
			switch test.metric.MType {
			case model.Counter: assert.Equal(t, strconv.FormatInt(*test.metric.Delta, 10), string(body))
			case model.Gauge: assert.Equal(t, strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", *test.metric.Value), "0"), "."), string(body))
			}
			
			result.Body.Close()
		})
	}
}