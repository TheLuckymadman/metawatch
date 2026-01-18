package service

import (
	"github.com/TheLuckymadman/metawatch/internal/model"
)

func NewAudit(auditType model.AuditType, msg model.AuditMsg, filePath string, url string) Audit {
	switch auditType {
	case model.AuditToFile:
		return GetAuditToFile(filePath, msg)
	case model.AuditToServer:
		return GetAuditToServer(url, msg)
	default:
		panic("unknown audit type")
	}
}
