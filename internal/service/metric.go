package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/TheLuckymadman/metawatch/internal/model"
	"github.com/TheLuckymadman/metawatch/internal/repository"
)

func AddMetric(metricName string, metricValue string, metricType string, agentIP string, s Storage) error {
	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse metric value:\n%v", err.Error())
		}
		err = s.AddMetric(agentIP, metricType, metricName, 0, value)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't parse metric value:\n%v", err.Error())
		}
		err = s.AddMetric(agentIP, metricType, metricName, value, 0)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	default:
		return fmt.Errorf("invalid metric type")
	}
	return nil
}

func AddObjMetric(metric model.Metrics, agentIP string, s Storage) error {
	switch metric.MType {
	case model.Counter: {
		err := s.AddMetric(agentIP, metric.MType, metric.ID, 0, *metric.Delta)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	}
	case model.Gauge: {
		err := s.AddMetric(agentIP, metric.MType, metric.ID, *metric.Value, 0)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:\n%v", err)
		}
	}
	default:
		return fmt.Errorf("invalid metric type") 
	}
	return nil
}

func GetMetric(metricName string, metricType string, agentIP string, s Storage) (result string, err error) {
	switch metricType {
	case model.Counter: {
		_, delta, err := s.GetMetric(agentIP, metricType, metricName)
		if err != nil {
			return "", fmt.Errorf("there was an error while getting metric:\n%v", err)
		}
		result = fmt.Sprintf("%d", delta)
	}
	case model.Gauge: {
		value, _, err := s.GetMetric(agentIP, metricType, metricName)
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

func GetObjMetric(metricReq model.Metrics, agentIP string, s Storage) (metricResp *model.Metrics, err error) {
	if metricReq.MType == model.Counter || metricReq.MType == model.Gauge {
		metricResp, err = s.GetObjMetric(agentIP, metricReq.MType, metricReq.ID)
		if err != nil {
			return nil, fmt.Errorf("there was an error while getting metric:\n%v", err)
		}
	} else {
		return nil, fmt.Errorf("invalid metric type")
	}
	return metricResp, err
}

func ListMetric(s Storage) (map[string]*model.Metrics, []string, error) {
	memStorage := s.(*repository.MemStorage)
	memStorage.RLock()
	copyMemStorage := s.(*repository.MemStorage).Metrics
	memStorage.RUnlock()

	sortedMetrics := make([]string, 0, len(copyMemStorage))
	for k := range copyMemStorage {
		sortedMetrics = append(sortedMetrics, k)
	}
	sort.Strings(sortedMetrics)

	return copyMemStorage, sortedMetrics, nil
}