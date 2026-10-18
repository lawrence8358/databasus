package system_backupsfolder

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	audit_logs_models "databasus-backend/internal/features/audit_logs/models"
	users_interfaces "databasus-backend/internal/features/users/interfaces"
	users_models "databasus-backend/internal/features/users/models"
	"databasus-backend/internal/util/backupsfolder"
	"databasus-backend/internal/util/backupsfolder/settingsfile"
)

type BackupsFolderService struct {
	auditLogWriter users_interfaces.AuditLogWriter
	logger         *slog.Logger
	changeMutex    sync.Mutex
}

func (s *BackupsFolderService) GetBackupsFolder() *BackupsFolderResponse {
	if _, isChangeable := getSettingsFilePath(); !isChangeable {
		return &BackupsFolderResponse{IsChangeable: false, PreviousFolders: []string{}}
	}

	return buildResponse(backupsfolder.GetSettings())
}

// ChangeBackupsFolder switches where new backups go at once. Nothing is moved: backups written
// before stay in their folder, which is kept in the settings so they can still be restored and
// deleted by retention.
func (s *BackupsFolderService) ChangeBackupsFolder(
	ctx context.Context,
	changedBy *users_models.User,
	request ChangeBackupsFolderRequest,
) (*BackupsFolderResponse, error) {
	settingsFilePath, isChangeable := getSettingsFilePath()
	if !isChangeable {
		return nil, ErrBackupsFolderNotChangeable
	}

	s.changeMutex.Lock()
	defer s.changeMutex.Unlock()

	currentSettings := backupsfolder.GetSettings()
	newFolder := strings.TrimSpace(request.Folder)

	if settingsfile.IsSameFolder(currentSettings.CurrentFolder, newFolder) {
		return buildResponse(currentSettings), nil
	}

	if err := prepareBackupsFolder(newFolder); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBackupsFolderRefused, err)
	}

	changedSettings := currentSettings.WithCurrentFolder(newFolder)

	if err := settingsfile.WriteSettings(settingsFilePath, changedSettings); err != nil {
		return nil, fmt.Errorf("failed to record the backups folder: %w", err)
	}

	backupsfolder.UseSettings(changedSettings)

	s.logger.InfoContext(
		ctx,
		"backups folder changed",
		"user_id", changedBy.ID,
		"previous_backups_folder", currentSettings.CurrentFolder,
		"backups_folder", changedSettings.CurrentFolder,
	)

	s.auditLogWriter.WriteAuditLog(ctx, audit_logs_models.AuditEntry{
		Message: fmt.Sprintf(
			"Backups folder changed from %s to %s",
			currentSettings.CurrentFolder,
			changedSettings.CurrentFolder,
		),
		UserID:      &changedBy.ID,
		WorkspaceID: nil,
	})

	return buildResponse(changedSettings), nil
}

func getSettingsFilePath() (string, bool) {
	if !isChangeSupported {
		return "", false
	}

	return backupsfolder.GetSettingsFilePathFromEnvironment()
}

func buildResponse(settings settingsfile.Settings) *BackupsFolderResponse {
	return &BackupsFolderResponse{
		IsChangeable:    true,
		CurrentFolder:   settings.CurrentFolder,
		PreviousFolders: settings.PreviousFolders,
	}
}
