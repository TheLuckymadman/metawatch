package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

type Service struct {
	storage  Storage
	register map[string]Observer
}

func NewService(s Storage) *Service {
	return &Service{s, make(map[string]Observer)}
}

func (s *Service) Register(o Observer) {
	srvID := o.GetID()
	fmt.Printf("register %s as a subscriber in Service\n", srvID)
	s.register[srvID] = o
}

func (s *Service) Deregister(o Observer) {
	delete(s.register, o.GetID())
}

func (s *Service) Notify(ctx context.Context, metrics []model.Metrics, agentIP string) {
	for _, o := range s.register {
		o.Update(ctx, metrics, agentIP)
	}
}

func (s *Service) AddMetric(ctx context.Context, metricName string, metricValue string, metricType string, agentIP string) error {
	switch metricType {
	case model.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse metric value:%w", err)
		}
		err = s.storage.AddMetric(ctx, agentIP, metricType, metricName, 0, value)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:%w", err)
		}
		m := model.Metrics{ID: metricName, MType: model.Counter, Delta: &value}
		s.Notify(ctx, []model.Metrics{m}, agentIP)
	case model.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't parse metric value:%w", err)
		}
		err = s.storage.AddMetric(ctx, agentIP, metricType, metricName, value, 0)
		if err != nil {
			return fmt.Errorf("there was an error while adding metrics to the database:%w", err)
		}
		m := model.Metrics{ID: metricName, MType: model.Gauge, Value: &value}
		s.Notify(ctx, []model.Metrics{m}, agentIP)
	default:
		return fmt.Errorf("invalid metric type")
	}
	return nil
}

func (s *Service) AddObjMetric(ctx context.Context, metric model.Metrics, agentIP string) error {
	metrics := []model.Metrics{metric}
	switch metric.MType {
	case model.Counter:
		{
			if metric.Delta == nil {
				return fmt.Errorf("no delta value error")
			}
		}
	case model.Gauge:
		{
			if metric.Value == nil {
				return fmt.Errorf("no value error")
			}
		}
	default:
		return fmt.Errorf("invalid metric type")
	}

	err := s.storage.AddMetrics(ctx, agentIP, metrics)
	if err != nil {
		return fmt.Errorf("there was an error while adding metrics to the database:%w", err)
	}
	s.Notify(ctx, metrics, agentIP)
	return nil
}

func (s *Service) AddObjMetrics(ctx context.Context, metrics []model.Metrics, agentIP string) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics provided")
	}
	for _, metric := range metrics {
		switch metric.MType {
		case model.Counter:
			{
				if metric.Delta == nil || metric.ID == "" {
					return fmt.Errorf("missed metric attributes %v", metric)
				}
			}
		case model.Gauge:
			{
				if metric.Value == nil || metric.ID == "" {
					return fmt.Errorf("missed metric attributes %v", metric)
				}
			}
		default:
			return fmt.Errorf("wrong metric  type")
		}
	}
	err := s.storage.AddMetrics(ctx, agentIP, metrics)
	if err != nil {
		return fmt.Errorf("there was an error while adding metrics to the database:%w", err)
	}
	s.Notify(ctx, metrics, agentIP)
	return nil
}

func (s *Service) GetObjMetric(ctx context.Context, metricReq model.Metrics, agentIP string) (metricResp *model.Metrics, err error) {
	if metricReq.MType == model.Counter || metricReq.MType == model.Gauge {
		metricResp, err = s.storage.GetObjMetric(ctx, agentIP, metricReq.MType, metricReq.ID)
		if err != nil {
			return nil, fmt.Errorf("there was an error while getting metric:%w", err)
		}
	} else {
		return nil, fmt.Errorf("invalid metric type")
	}
	return metricResp, err
}

func (s *Service) ListMetric(ctx context.Context) (map[string]*model.Metrics, []string, error) {
	memStorage, err := s.storage.GetStore(ctx)
	if err != nil {
		return nil, nil, err
	}

	sortedMetrics := make([]string, 0, len(memStorage))
	for k := range memStorage {
		sortedMetrics = append(sortedMetrics, k)
	}
	sort.Strings(sortedMetrics)

	return memStorage, sortedMetrics, nil
}

func (s *Service) PingDB(ctx context.Context) error {
	return s.storage.PingDB(ctx)
}
