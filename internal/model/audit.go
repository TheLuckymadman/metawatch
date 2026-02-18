package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type AuditType int

const (
	AuditToFile AuditType = iota
	AuditToServer
)

type AuditMsg struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

func (m *AuditMsg) Marshal(metrics []Metrics, agentIP string) ([]byte, error) {
	var metricNames []string
	for _, m := range metrics {
		metricNames = append(metricNames, m.ID)
	}
	m.TS = time.Now().Unix()
	m.Metrics = metricNames
	m.IPAddress = agentIP
	data, err := json.MarshalIndent(m, "", "	")
	if err != nil {
		return nil, fmt.Errorf("marshal audit metrics: %w", err)
	}
	return data, nil
}
