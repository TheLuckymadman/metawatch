package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type FileStorage struct {
	MemStorage
	fileStoragePath string
	storeInterval   time.Duration
	restore         bool
	stopChan        chan struct{}
	syncChan        chan *model.Metrics
	wg              *sync.WaitGroup
}

func NewFileStorage(fileStoragePath string, storeInterval time.Duration, restore bool) (*FileStorage, error) {
	f := FileStorage{
		MemStorage:      MemStorage{Metrics: make(map[string]*model.Metrics)},
		fileStoragePath: fileStoragePath,
		storeInterval:   storeInterval,
		restore:         restore,
		stopChan:        make(chan struct{}),
		syncChan:        make(chan *model.Metrics, 100),
		wg:              &sync.WaitGroup{},
	}
	if f.restore {
		if err := f.loadFromFile(); err != nil {
			return nil, err
		}
	}
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		if err := f.fileSyncRunner(); err != nil {
			log.Printf("file sync runner error: %v", err)
		}
	}()

	return &f, nil
}

func (f *FileStorage) loadFromFile() error {
	file, err := os.Open(f.fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File %s not found, skipping restore", f.fileStoragePath)
			return nil
		}
		log.Printf("Opening file error: %v", err)
		return fmt.Errorf("opening file error: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(file)
	if err != nil {
		log.Printf("Reading file error: %v", err)
		return fmt.Errorf("reading file error: %w", err)
	}
	err = json.Unmarshal(buf.Bytes(), &f.MemStorage.Metrics)
	if err != nil {
		log.Printf("Error unmarshalling data from the file: %v", err)
		return fmt.Errorf("error unmarshalling data from the file: %w", err)
	}
	return nil
}

func (f *FileStorage) fileSyncRunner() error {
	if f.storeInterval == 0 {
		for {
			select {
			case <-f.syncChan:
				err := f.saveToFile()
				if err != nil {
					return err
				}
			case <-f.stopChan:
				log.Println("Stop file sync runner")
				return nil
			}
		}
	} else {
		ticker := time.NewTicker(f.storeInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := f.saveToFile()
				if err != nil {
					return err
				}
			case <-f.stopChan:
				log.Println("Stop file sync runner")
				return nil
			}
		}
	}
}

func (f *FileStorage) saveToFile() error {
	var m []byte
	var err error

	file, err := os.OpenFile(f.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("Opening or creating file error: %v", err)
		return err
	}
	defer file.Close()

	f.MemStorage.RLock()
	memStorage := f.MemStorage.Metrics
	f.MemStorage.RUnlock()

	m, err = json.MarshalIndent(memStorage, "", "	")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return err
	}

	if _, err := file.Write(m); err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return err
	}
	return nil
}

func (f *FileStorage) Close() error {
	close(f.stopChan)
	f.wg.Wait()
	f.saveToFile()
	return nil
}

func (f *FileStorage) AddMetric(ctx context.Context, agentID string, metricType string, metricName string, value float64, delta int64) error {
	key := agentID + "_" + metricName

	f.Lock()
	metric, ok := f.Metrics[key]
	if !ok {
		f.Metrics[key] = &model.Metrics{
			ID:    metricName,
			MType: metricType,
		}
		metric = f.Metrics[key]
	}
	f.Unlock()

	if f.storeInterval == 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-f.stopChan:
			return fmt.Errorf("drop data because of a stop signal happens")
		default:
		}

		select {
		case f.syncChan <- metric:
		default:
			log.Println("syncChan is full, skipping writing")
		}
	}
	// defer metric.Unlock()

	// metric.Lock()
	switch metricType {
	case model.Counter:
		{
			if metric.Delta == nil {
				metric.Delta = new(int64)
			}
			*metric.Delta += delta
			metric.Value = nil
		}

	case model.Gauge:
		{
			metric.Delta = nil
			metric.Value = &value
		}
	}

	return nil
}
