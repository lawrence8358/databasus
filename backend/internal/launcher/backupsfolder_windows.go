package launcher

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"databasus-backend/internal/util/backupsfolder/settingsfile"
)

type backupsFolderSelection struct {
	settingsFilePath string
	currentFolder    string
}

// The settings file in the data folder is the one place that says where backups go: the Web UI
// and Setup both change it there. -backups-dir only decides the folder of a data folder that has
// no settings file yet, so a service command line cannot undo a later change.
func selectBackupsFolder(config Config, logger *slog.Logger) (backupsFolderSelection, error) {
	settingsFilePath := settingsfile.GetSettingsFilePath(config.DataFolder)

	settings, err := settingsfile.ReadSettings(settingsFilePath)
	switch {
	case err == nil:
		if config.BackupsFolder != "" &&
			!settingsfile.IsSameFolder(getAbsoluteFolder(config.BackupsFolder), settings.CurrentFolder) {
			logger.Warn(
				"ignoring -backups-dir: the backups folder set in the Web UI or by Setup wins",
				"backups_folder", settings.CurrentFolder,
				"ignored_backups_folder", config.BackupsFolder,
				"settings_file", settingsFilePath,
			)
		}
	case errors.Is(err, settingsfile.ErrSettingsNotFound):
		settings, err = createBackupsFolderSettings(config, settingsFilePath)
		if err != nil {
			return backupsFolderSelection{}, err
		}

		logger.Info(
			"recorded the backups folder",
			"backups_folder",
			settings.CurrentFolder,
			"settings_file",
			settingsFilePath,
		)
	default:
		return backupsFolderSelection{}, fmt.Errorf("backups folder settings: %w", err)
	}

	return backupsFolderSelection{settingsFilePath: settingsFilePath, currentFolder: settings.CurrentFolder}, nil
}

func createBackupsFolderSettings(config Config, settingsFilePath string) (settingsfile.Settings, error) {
	backupsFolder := config.BackupsFolder
	if backupsFolder == "" {
		backupsFolder = filepath.Join(config.DataFolder, "backups")
	}

	settings := settingsfile.Settings{CurrentFolder: getAbsoluteFolder(backupsFolder)}

	if err := os.MkdirAll(config.DataFolder, 0o755); err != nil {
		return settingsfile.Settings{}, fmt.Errorf("create the data folder: %w", err)
	}

	if err := settingsfile.WriteSettings(settingsFilePath, settings); err != nil {
		return settingsfile.Settings{}, fmt.Errorf("record the backups folder: %w", err)
	}

	return settings, nil
}

func getAbsoluteFolder(folder string) string {
	absoluteFolder, err := filepath.Abs(folder)
	if err != nil {
		return filepath.Clean(folder)
	}

	return absoluteFolder
}
