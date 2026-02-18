package repository

import (
	"context"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

func TestMemStorage_GetStore(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		metrics []model.Metrics
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemStorage()
			for _, metric := range tt.metrics {
				newMetric := metric
				m.Metrics[metric.ID] = &newMetric
			}
			storeSz := len(m.Metrics)
			got, gotErr := m.GetStore(context.Background())

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetStore() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Error("GetStore() succeeded unexpectedly")
			}

			for _, metric := range tt.metrics {
				delete(got, metric.ID)
			}
			if len(m.Metrics) != storeSz {
				t.Errorf("GetStore() unexpected store lenght, want %d, got %d", storeSz, len(m.Metrics))
			}
		})
	}
}

func TestMemStorage_AddMetric(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		agentID    string
		metricType string
		metricName string
		value      float64
		delta      int64
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemStorage()
			gotErr := m.AddMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName, tt.value, tt.delta)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddMetric() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddMetric() succeeded unexpectedly")
			}
		})
	}
}

func TestMemStorage_AddMetrics(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		agentID string
		metrics []model.Metrics
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemStorage()
			gotErr := m.AddMetrics(context.Background(), tt.agentID, tt.metrics)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddMetrics() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddMetrics() succeeded unexpectedly")
			}
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		agentID    string
		metricType string
		metricName string
		want       float64
		want2      int64
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemStorage()
			got, got2, gotErr := m.GetMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetMetric() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetMetric() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetMetric() = %v, want %v", got, tt.want)
			}
			if true {
				t.Errorf("GetMetric() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestMemStorage_GetObjMetric(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		agentID    string
		metricType string
		metricName string
		want       *model.Metrics
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemStorage()
			got, gotErr := m.GetObjMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetObjMetric() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetObjMetric() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetObjMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}
