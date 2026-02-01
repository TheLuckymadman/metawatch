package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/handler/mocks"
	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

func TestHandlerStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func() http.HandlerFunc
		handlerPath    string
		reqURI         string
		reqMethod      string
		reqContentType string
		reqBody        string
		wantHTTPCode   int
		setupMock      func(m *mocks.MockService)
	}{
		{
			name: "status code OK for MetricSetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				mockSvc := mocks.NewMockService(ctrl)
				mockSvc.EXPECT().AddMetric(
					gomock.Any(),
					"cpu",
					"1",
					model.Counter,
					gomock.Any(),
				).Return(nil).Times(1)
				return MetricSetterHandler(mockSvc)
			},
			handlerPath:    "/update/{type}/*",
			reqURI:         "/update/counter/cpu/1",
			reqMethod:      http.MethodPost,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusOK,
		},
		{
			name: "status code Error for MetricSetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				mockSvc := mocks.NewMockService(ctrl)
				return MetricSetterHandler(mockSvc)
			},
			handlerPath:    "/update/{type}/*",
			reqURI:         "/update/counter/",
			reqMethod:      http.MethodPost,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusNotFound,
		},
		{
			name: "status code OK JSONSetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				mockSvc := mocks.NewMockService(ctrl)
				mockSvc.EXPECT().AddObjMetrics(
					gomock.Any(),
					[]model.Metrics{
						{
							ID:    "cpu",
							MType: model.Counter,
							Delta: utils.Int64Ptr(1),
						},
					},
					gomock.Any(),
				).Return(nil).Times(1)
				return JSONSetterHandler(mockSvc)
			},
			handlerPath:    "/update/",
			reqURI:         "/update/",
			reqMethod:      http.MethodPost,
			reqContentType: "application/json",
			reqBody: `{
				"id": "cpu",
				"type": "counter",
				"delta": 1
			}`,
			wantHTTPCode: http.StatusOK,
		},
		{
			name: "status code Error for JSONSetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				mockSvc := mocks.NewMockService(ctrl)
				return JSONSetterHandler(mockSvc)
			},
			handlerPath:    "/update/",
			reqURI:         "/update/",
			reqMethod:      http.MethodPost,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusMethodNotAllowed,
		},
		{
			name: "status code OK for MetricGetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				mockSvc := mocks.NewMockService(ctrl)
				mockSvc.EXPECT().GetObjMetric(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(
					&model.Metrics{
						ID:    "cpu",
						MType: model.Counter,
						Delta: utils.Int64Ptr(1),
					}, nil).Times(1)
				return MetricGetterHandler(mockSvc)
			},
			handlerPath:    "/value/*",
			reqURI:         "/value/counter/cpu",
			reqMethod:      http.MethodGet,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusOK,
		},
		{
			name: "wrong method Error for MetricGetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				mockSvc := mocks.NewMockService(ctrl)
				return MetricGetterHandler(mockSvc)
			},
			handlerPath:    "/value/{type}/*",
			reqURI:         "/value/counter/",
			reqMethod:      http.MethodPost,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusMethodNotAllowed,
		},
		{
			name: "status code OK JSONGetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				mockSvc := mocks.NewMockService(ctrl)
				mockSvc.EXPECT().GetObjMetric(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(
					&model.Metrics{
						ID:    "cpu",
						MType: model.Counter,
						Delta: utils.Int64Ptr(1),
					}, nil).Times(1)
				return JSONGetterHandler(mockSvc)
			},
			handlerPath:    "/value/",
			reqURI:         "/value/",
			reqMethod:      http.MethodPost,
			reqContentType: "application/json",
			reqBody: `{
				"id": "cpu",
				"type": "counter"
			}`,
			wantHTTPCode: http.StatusOK,
		},
		{
			name: "wrong method Error for JSONGetterHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				mockSvc := mocks.NewMockService(ctrl)
				return JSONGetterHandler(mockSvc)
			},
			handlerPath:    "/value/",
			reqURI:         "/value/",
			reqMethod:      http.MethodGet,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusMethodNotAllowed,
		},
		{
			name: "status code OK MetricsListHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				mockSvc := mocks.NewMockService(ctrl)
				mockSvc.EXPECT().ListMetric(
					gomock.Any(),
				).Return(
					make(map[string]*model.Metrics),
					[]string{},
					nil).Times(1)
				return MetricsListHandler(mockSvc)
			},
			handlerPath:    "/",
			reqURI:         "/",
			reqMethod:      http.MethodGet,
			reqContentType: "application/json",
			reqBody:        "",
			wantHTTPCode:   http.StatusOK,
		},
		{
			name: "wrong method Error for MetricsListHandler",
			setupHandler: func() http.HandlerFunc {
				ctrl := gomock.NewController(t)
				mockSvc := mocks.NewMockService(ctrl)
				return MetricsListHandler(mockSvc)
			},
			handlerPath:    "/",
			reqURI:         "/",
			reqMethod:      http.MethodPost,
			reqContentType: "",
			reqBody:        "",
			wantHTTPCode:   http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			handler := test.setupHandler()

			router := chi.NewRouter()
			switch test.reqMethod {
			case http.MethodPost:
				router.Post(test.handlerPath, handler)
			case http.MethodGet:
				router.Get(test.handlerPath, handler)
			}

			body := bytes.NewReader([]byte(test.reqBody))
			req := httptest.NewRequest(test.reqMethod, test.reqURI, body)
			if test.reqContentType != "" {
				req.Header.Set("Content-Type", test.reqContentType)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			t.Logf("response status: %d, body: %s", rec.Code, rec.Body)

			if rec.Code != test.wantHTTPCode {
				t.Errorf("want %d, got %d, body %s", test.wantHTTPCode, rec.Code, rec.Body)
			}
		})
	}
}

func TestMetricSetterHandler_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockService(ctrl)
	handler := MetricSetterHandler(mockSvc)

	mockSvc.EXPECT().AddMetric(
		gomock.Any(),
		"cpu",
		"1",
		model.Counter,
		"127.0.0.1",
	).Return(nil).Times(1)

	r := chi.NewRouter()
	r.Post("/update/{type}/*", handler)

	t.Run("test http status OK", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/update/counter/cpu/1", nil)
		req.RemoteAddr = "127.0.0.1"
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d\nbody: %s", http.StatusOK, rec.Code, rec.Body)
		}
	})

	t.Run("test http status error on the invalid path", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/update/counter/1", nil)
		req.RemoteAddr = "127.0.0.1"
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}
