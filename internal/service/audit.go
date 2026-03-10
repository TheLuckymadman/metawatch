package service

import (
	"fmt"

	"github.com/TheLuckymadman/metawatch/internal/model"
)

func NewAudit(auditType model.AuditType, msg model.AuditMsg, filePath string, url string) (Audit, error) {
	switch auditType {
	case model.AuditToFile:
		return GetAuditToFile(filePath, msg), nil
	case model.AuditToServer:
		return GetAuditToServer(url, msg), nil
	default:
		return nil, fmt.Errorf("unknown audit type")
	}
}
