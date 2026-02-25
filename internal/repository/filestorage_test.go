package repository

import (
	"context"
	"testing"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/utils"
)

func TestGetStore(t *testing.T) {

}
func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name            string
		fileStoragePath string
		storeInterval   time.Duration
		restore         bool
		wantErr         bool
	}{
		{
			name:            "NewFileStorage returns FileStorage successfully",
			fileStoragePath: "metrics_test.txt",
			storeInterval:   0,
			restore:         false,
			wantErr:         false,
		},
		{
			name:            "NewFileStorage restored FileStorage from the file",
			fileStoragePath: "metrics_test.txt",
			storeInterval:   0,
			restore:         true,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, gotErr := NewFileStorage(tt.fileStoragePath, tt.storeInterval, tt.restore)
			time.Sleep(1 * time.Second)

			if f.restore {
				if len(f.Metrics) == 0 {
					t.Errorf("restore from file, want > 0, got 0")
				}
			}

			f.Metrics["key1"] = &model.Metrics{
				ID:    "cpu",
				MType: model.Counter,
				Delta: utils.Int64Ptr(1),
			}

			f.syncChan <- &model.Metrics{
				ID:    "cpu",
				MType: model.Counter,
				Delta: utils.Int64Ptr(1),
			}
			time.Sleep(5 * time.Second)
			f.Close()
			if gotErr != nil != tt.wantErr {
				t.Errorf("NewFileStorage() failed: want error: %t, got: %v", tt.wantErr, gotErr)
			}
		})
	}
}

func TestFileStorage_AddMetric(t *testing.T) {
	tests := []struct {
		name            string
		fileStoragePath string
		storeInterval   time.Duration
		restore         bool
		agentID         string
		metricType      string
		metricName      string
		value           float64
		delta           int64
		wantErr         bool
	}{
		{
			name:            "add metric to the file",
			fileStoragePath: "metrics_test.txt",
			storeInterval:   time.Duration(0),
			restore:         false,
			agentID:         "127.0.0.1_cpu_test",
			metricType:      model.Counter,
			metricName:      "cpu",
			delta:           1,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFileStorage(tt.fileStoragePath, tt.storeInterval, tt.restore)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := f.AddMetric(context.Background(), tt.agentID, tt.metricType, tt.metricName, tt.value, tt.delta)
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
