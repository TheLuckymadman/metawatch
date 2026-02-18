package service

import (
	"context"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/service/mocks"
	"github.com/TheLuckymadman/metawatch/internal/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAddMetrics(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue string
		metricType  string
		agentIP     string
		wantErr     bool
		mockSetup   func(m *mocks.MockStorage)
	}{
		{
			name:        "adding metrics of a counter type",
			metricName:  "cpu",
			metricValue: "1",
			metricType:  model.Counter,
			agentIP:     "127.0.0.1",
			wantErr:     false,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().AddMetric(
					gomock.Any(),
					gomock.Any(),
					model.Counter,
					"cpu",
					float64(0),
					int64(1),
				).Return(nil)
			},
		},
		{
			name:        "adding metrics of a wrong type",
			metricName:  "cpu",
			metricValue: "1",
			metricType:  "wrong",
			agentIP:     "127.0.0.1",
			wantErr:     true,
			mockSetup: func(m *mocks.MockStorage) {

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStorage := mocks.NewMockStorage(ctrl)
			test.mockSetup(mockStorage)
			svc := NewService(mockStorage)
			err := svc.AddMetric(
				context.Background(),
				test.metricName,
				test.metricValue,
				test.metricType,
				test.agentIP,
			)
			if err != nil != test.wantErr {
				t.Errorf("want error %v, got error %v", test.wantErr, err)
			}
		})
	}
}

func TestAddObjMetric(t *testing.T) {
	tests := []struct {
		name      string
		metrics   model.Metrics
		agentIP   string
		wantErr   bool
		mockSetup func(m *mocks.MockStorage)
	}{
		{
			name: "adding metrics of a valid type",
			metrics: model.Metrics{
				ID:    "ram",
				MType: model.Gauge,
				Value: utils.FloatPtr(100),
			},
			agentIP: "127.0.0.1",
			wantErr: false,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().AddMetrics(
					gomock.Any(),
					"127.0.0.1",
					gomock.AssignableToTypeOf([]model.Metrics{}),
				).Return(nil)
			},
		},
		{
			name: "adding metrics of a invalid type",
			metrics: model.Metrics{
				ID:    "ram",
				MType: "wrong",
				Value: utils.FloatPtr(100),
			},
			agentIP: "127.0.0.1",
			wantErr: true,
			mockSetup: func(m *mocks.MockStorage) {

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStorage := mocks.NewMockStorage(ctrl)
			test.mockSetup(mockStorage)
			svc := NewService(mockStorage)
			err := svc.AddObjMetric(
				context.Background(),
				test.metrics,
				test.agentIP,
			)
			if err != nil != test.wantErr {
				t.Errorf("want error %v, got error %v", test.wantErr, err)
			}
		})
	}
}

func TestAddObjMetrics(t *testing.T) {
	tests := []struct {
		name      string
		metrics   []model.Metrics
		agentIP   string
		wantErr   bool
		mockSetup func(m *mocks.MockStorage)
	}{
		{
			name: "adding metrics of a valid type",
			metrics: []model.Metrics{
				{
					ID:    "cpu",
					MType: model.Counter,
					Delta: utils.Int64Ptr(100),
				},
				{
					ID:    "ram",
					MType: model.Gauge,
					Value: utils.FloatPtr(100),
				},
			},
			agentIP: "127.0.0.1",
			wantErr: false,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().AddMetrics(
					gomock.Any(),
					"127.0.0.1",
					gomock.AssignableToTypeOf([]model.Metrics{}),
				).Return(nil)
			},
		},
		{
			name: "adding metrics of a invalid type",
			metrics: []model.Metrics{
				{
					ID:    "cpu",
					MType: model.Counter,
					Delta: utils.Int64Ptr(100),
				},
				{
					ID:    "ram",
					MType: "wrong",
					Value: utils.FloatPtr(100),
				},
			},
			agentIP: "127.0.0.1",
			wantErr: true,
			mockSetup: func(m *mocks.MockStorage) {

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStorage := mocks.NewMockStorage(ctrl)
			test.mockSetup(mockStorage)
			svc := NewService(mockStorage)
			err := svc.AddObjMetrics(
				context.Background(),
				test.metrics,
				test.agentIP,
			)
			if err != nil != test.wantErr {
				t.Errorf("want error %v, got error %v", test.wantErr, err)
			}
		})
	}
}

func TestGetObjMetric(t *testing.T) {
	tests := []struct {
		name      string
		metrics   model.Metrics
		agentIP   string
		wantErr   bool
		mockSetup func(m *mocks.MockStorage)
	}{
		{
			name: "get metrics of a valid type",
			metrics: model.Metrics{
				ID:    "cpu",
				MType: model.Counter,
				Delta: utils.Int64Ptr(100),
			},
			agentIP: "127.0.0.1",
			wantErr: false,
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().GetObjMetric(
					gomock.Any(),
					"127.0.0.1",
					gomock.Any(),
					gomock.Any(),
				).Return(
					&model.Metrics{
						ID:    "cpu",
						MType: model.Counter,
						Delta: utils.Int64Ptr(100),
					},
					nil,
				)
			},
		},
		{
			name: "get metrics of a invalid type",
			metrics: model.Metrics{
				ID:    "cpu",
				MType: "wrongType",
			},
			agentIP: "127.0.0.1",
			wantErr: true,
			mockSetup: func(m *mocks.MockStorage) {

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStorage := mocks.NewMockStorage(ctrl)
			test.mockSetup(mockStorage)
			svc := NewService(mockStorage)
			metricResp, err := svc.GetObjMetric(
				context.Background(),
				test.metrics,
				test.agentIP,
			)
			if err != nil {
				if !test.wantErr {
					t.Errorf("want error %v, got error %v", test.wantErr, err)
				}
				return
			}

			assert.Equal(t, test.metrics, *metricResp)
		})
	}
}

func TestListMetric(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		metrics   []model.Metrics
		agentIP   string
		mockSetup func(m *mocks.MockStorage)
	}{
		{
			name:    "add counter data",
			wantErr: false,
			metrics: []model.Metrics{
				{
					ID:    "cpu",
					MType: model.Counter,
					Delta: utils.Int64Ptr(1),
				},
				{
					ID:    "ram",
					MType: model.Gauge,
					Value: utils.FloatPtr(100000),
				},
			},
			agentIP: "127.0.0.1",
			mockSetup: func(m *mocks.MockStorage) {
				m.EXPECT().GetStore(gomock.Any()).Return(
					map[string]*model.Metrics{
						"127.0.0.1_cpu": {
							ID:    "cpu",
							MType: model.Counter,
							Delta: utils.Int64Ptr(1),
						},
						"127.0.0.1_ram": {
							ID:    "ram",
							MType: model.Gauge,
							Value: utils.FloatPtr(100000),
						},
					},
					nil,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStorage := mocks.NewMockStorage(ctrl)
			tt.mockSetup(mockStorage)

			svc := NewService(mockStorage)
			memStorage, _, err := svc.ListMetric(context.Background())

			assert.NoError(t, err)

			for _, metric := range tt.metrics {
				gotMetric := memStorage[tt.agentIP+"_"+metric.ID]
				assert.Equal(t, metric, *gotMetric)
			}
		})
	}
}
