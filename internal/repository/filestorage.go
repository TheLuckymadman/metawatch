package repository

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type FileStorage struct {
	MemStorage
	fileStoragePath string
	storeInterval int
	restore bool
	stopChan chan struct{}
	syncChan chan *model.Metrics
}

func NewFileStorage(fileStoragePath string, storeInterval int, restore bool) *FileStorage {
	f := FileStorage{
		MemStorage: MemStorage{Metrics: make(map[string]*model.Metrics)},
		fileStoragePath: fileStoragePath, 
		storeInterval: storeInterval, 
		restore: restore,
		stopChan: make(chan struct{}),
		syncChan: make(chan *model.Metrics, 100),
	}
	if f.restore {
		f.loadFromFile()
	}
	go f.fileSyncRunner()
	
	return &f
}

func (f *FileStorage) loadFromFile() {
	file, err := os.Open(f.fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File %s not found, skipping restore", f.fileStoragePath)
			return
		}
		log.Printf("Opening or creating file error: %v", err)
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(file)
	if err != nil {
		log.Printf("Reading file error: %v", err)
	}
	err = json.Unmarshal(buf.Bytes(), &f.MemStorage.Metrics)
	if err != nil {
		log.Printf("Error unmarshalling data from the file: %v", err)
		return
	}
}

func (f *FileStorage) fileSyncRunner() {
	if f.storeInterval == 0 {
		for {
			select {
			case <- f.syncChan:
				f.saveToFile()
			case <- f.stopChan:
				log.Println("Stop file sync runner")
				return
			}
		}
	} else {
		ticker := time.NewTicker(time.Duration(f.storeInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <- ticker.C:
				f.saveToFile()
			case <- f.stopChan:
				log.Println("Stop file sync runner")
				return
			}
		}
	}
}

func (f *FileStorage) saveToFile() {
	var m []byte
	var err error

	file, err := os.OpenFile(f.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("Opening or creating file error: %v", err)
	}
	defer file.Close()

	f.MemStorage.RLock()
	memStorage := f.MemStorage.Metrics
	f.MemStorage.RUnlock()

	m, err = json.MarshalIndent(memStorage, "", "	")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	if _, err := file.Write(m); err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
}

func (f *FileStorage) Stop() {
	close(f.stopChan)
}

func (f *FileStorage) AddMetric(agentID string, metricType string, metricName string, value float64, delta int64) error {
	key := agentID + "_" + metricName

	f.Lock()
	metric, ok := f.Metrics[key]
	if !ok {
		f.Metrics[key] = &model.Metrics{
			ID: metricName,
			MType: metricType,
		}
		metric = f.Metrics[key]
	}
	f.Unlock()

	if f.storeInterval == 0 {
		select {
		case f.syncChan <- metric:
		default:
			log.Println("syncChan is full, skipping writing")
		}
	}
	// defer metric.Unlock()

	// metric.Lock()
	switch metricType {
	case model.Counter: {
		if metric.Delta == nil {
			metric.Delta = new(int64)
		}
		*metric.Delta += delta
		metric.Value = nil
	}
	
	case model.Gauge: {
		metric.Delta = nil
		metric.Value = &value
	}
	}
	
	return nil
}
