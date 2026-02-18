package service

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

var (
	auditToFile *AuditToFile
	toFileOnce sync.Once
)

type AuditToFile struct {
	id       model.AuditType
	filePath string
	msg      model.AuditMsg
}

func GetAuditToFile(filePath string, msg model.AuditMsg) *AuditToFile {
	toFileOnce.Do(
		func() {
			auditToFile = &AuditToFile{
				id:       model.AuditToFile,
				filePath: filePath,
				msg:      msg,
			}
		})
	return auditToFile
}

func (a *AuditToFile) Update(ctx context.Context, metrics []model.Metrics, agentIP string) error {
	file, err := os.OpenFile(a.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("opening or creating file: %w", err)
	}
	data, err := a.msg.Marshal(metrics, agentIP)
	if err != nil {
		return fmt.Errorf("audit update: %w", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("write audit metrics to file: %w", err)
	}
	return nil
}

func (a *AuditToFile) GetID() string {
	return "AuditToFile"
}
