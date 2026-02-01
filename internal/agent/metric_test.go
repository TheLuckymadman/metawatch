package agent

import (
	"context"

	"reflect"
	"testing"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockSender struct {
	mock.Mock
}

func (m *MockSender) SendMetrics(ctx context.Context, metric []model.Metrics) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func TestGetMetrics(t *testing.T) {
	mockSender := new(MockSender)
	mockSender.On("SendMetrics", mock.Anything, mock.Anything).Return(nil)

	localMetrics := NewLocalMetrics(mockSender)

	t.Run("type check", func(t *testing.T) {
		metricStoreType := reflect.TypeOf(localMetrics.M)
		t.Logf("metric store type: %s", metricStoreType.Kind())
		if metricStoreType.Kind().String() != "slice" {
			t.Errorf("wrong metric store type, want: 'slice', got: %s", metricStoreType.Kind().String())
		}
	})

	t.Run("GetMetrics populate the store", func(t *testing.T) {
		localMetrics.GetMetrics()
		if len(localMetrics.M) == 0 {
			t.Errorf("no metrics in the store after GetMetrics was executed")
		}
	})

	t.Run("SendMetrics removes sent metrics", func(t *testing.T) {
		localMetrics.SendMetrics(context.Background(), len(localMetrics.M))
		if len(localMetrics.M) != 0 {
			t.Errorf("metrics in the start after SendMetrics was called, want: 0, got: %d", len(localMetrics.M))
		}
	})

	t.Run("SendMetrics removes sent batch metrics equals of half store", func(t *testing.T) {
		localMetrics.GetMetrics()
		if len(localMetrics.M) <= 2 {
			t.Errorf("expected at least 2 metrics after GetMetrics, got %d", len(localMetrics.M))
		}
		metricsCount := len(localMetrics.M)
		half := len(localMetrics.M) / 2
		localMetrics.SendMetrics(context.Background(), half)
		t.Logf("send metrics %d from %d in the store", half, metricsCount)
		metricsLeftCount := metricsCount - half
		if len(localMetrics.M) != metricsLeftCount {
			t.Errorf("metrics sent and left in the store mismatch after SetMetrics was called, want: %d, got %d", metricsLeftCount, len(localMetrics.M))
		}
	})

}
