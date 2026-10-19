package system_backupsfolder

import (
	"sync"

	"databasus-backend/internal/features/audit_logs"
	"databasus-backend/internal/util/logger"
)

var backupsFolderService = &BackupsFolderService{
	audit_logs.GetAuditLogService(),
	logger.GetLogger(),
	sync.Mutex{},
}

var backupsFolderController = &BackupsFolderController{backupsFolderService}

func GetBackupsFolderService() *BackupsFolderService {
	return backupsFolderService
}

func GetBackupsFolderController() *BackupsFolderController {
	return backupsFolderController
}
