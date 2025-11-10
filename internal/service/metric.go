package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Service struct {
	storage Storage
}

func NewService(s Storage) *Service {
	return &Service{s}
}

func (s *Service) AddMetric(metricName string, metricValue string, metricType string, agentIP string) error {
	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse metric value:\n%v", err.Error())
		}
		err = s.storage.AddMetric(agentIP, metricType, metricName, 0, value)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't parse metric value:\n%v", err.Error())
		}
		err = s.storage.AddMetric(agentIP, metricType, metricName, value, 0)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	default:
		return fmt.Errorf("invalid metric type")
	}
	return nil
}

func (s *Service) AddObjMetric(metric model.Metrics, agentIP string) error {
	switch metric.MType {
	case model.Counter: {
		err := s.storage.AddMetric(agentIP, metric.MType, metric.ID, 0, *metric.Delta)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	}
	case model.Gauge: {
		err := s.storage.AddMetric(agentIP, metric.MType, metric.ID, *metric.Value, 0)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	}
	default:
		return fmt.Errorf("invalid metric type") 
	}
	return nil
}

func (s *Service) GetMetric(metricName string, metricType string, agentIP string) (result string, err error) {
	switch metricType {
	case model.Counter: {
		_, delta, err := s.storage.GetMetric(agentIP, metricType, metricName)
		if err != nil {
			return "", fmt.Errorf("there was an error while getting metric:\n%v", err)
		}
		result = fmt.Sprintf("%d", delta)
	}
	case model.Gauge: {
		value, _, err := s.storage.GetMetric(agentIP, metricType, metricName)
		if err != nil {
			return "", fmt.Errorf("there was an error while getting metric:\n%v", err)
		}
		result = strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", value), "0"), ".")
	}
	default:
		return "", fmt.Errorf("invalid metric type")
	}
	return result, nil
}

func (s *Service) GetObjMetric(metricReq model.Metrics, agentIP string) (metricResp *model.Metrics, err error) {
	if metricReq.MType == model.Counter || metricReq.MType == model.Gauge {
		metricResp, err = s.storage.GetObjMetric(agentIP, metricReq.MType, metricReq.ID)
		if err != nil {
			return nil, fmt.Errorf("there was an error while getting metric:\n%v", err)
		}
	} else {
		return nil, fmt.Errorf("invalid metric type")
	}
	return metricResp, err
}

func (s *Service) ListMetric() (map[string]*model.Metrics, []string, error) {
	memStorage := s.storage.GetStore()

	sortedMetrics := make([]string, 0, len(memStorage))
	for k := range memStorage {
		sortedMetrics = append(sortedMetrics, k)
	}
	sort.Strings(sortedMetrics)

	return memStorage, sortedMetrics, nil
}