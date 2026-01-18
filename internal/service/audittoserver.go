package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

var (
	auditToServer *AuditToServer
	toServerOnce  sync.Once
)

type AuditToServer struct {
	id         model.AuditType
	url        string
	msg        model.AuditMsg
	httpClient http.Client
}

func GetAuditToServer(url string, msg model.AuditMsg) *AuditToServer {
	toServerOnce.Do(
		func() {
			auditToServer = &AuditToServer{
				id:  model.AuditToServer,
				url: url,
				msg: msg,
				httpClient: http.Client{
					Timeout: 5 * time.Second,
				},
			}
		})
	return auditToServer
}

func (a *AuditToServer) Update(ctx context.Context, metrics []model.Metrics, agentIP string) error {
	data, err := a.msg.Marshal(metrics, agentIP)
	if err != nil {
		return fmt.Errorf("audit update: %w", err)
	}
	dataReader := bytes.NewReader(data)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, dataReader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("audit update: %w", err)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("audit update: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"audit update error: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}
	fmt.Printf("the audit data was sent successfully.\nAudit server response status: %s\n", resp.Status)
	return nil
}

func (a *AuditToServer) GetID() string {
	return "AuditToServer"
}
