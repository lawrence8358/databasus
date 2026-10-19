package backupsfolder

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"databasus-backend/internal/config"
	"databasus-backend/internal/util/backupsfolder/settingsfile"
	"databasus-backend/internal/util/locations"
	"databasus-backend/internal/util/logger"
)

type activeFolders struct {
	settings   settingsfile.Settings
	tempFolder string
}

var (
	activeState     atomic.Pointer[activeFolders]
	initActiveState = sync.OnceFunc(func() {
		activeState.CompareAndSwap(nil, loadActiveFolders())
	})
)

// GetSettingsFilePathFromEnvironment returns the settings file the launcher named. Without one the
// backups folder cannot be changed while the process runs.
func GetSettingsFilePathFromEnvironment() (string, bool) {
	settingsFilePath := strings.TrimSpace(os.Getenv(locations.BackupsFolderSettingsFileEnvVariable))

	return settingsFilePath, settingsFilePath != ""
}

func GetSettings() settingsfile.Settings {
	settings := getActiveFolders().settings
	settings.PreviousFolders = slices.Clone(settings.PreviousFolders)

	return settings
}

func GetCurrentFolder() string {
	return getActiveFolders().settings.CurrentFolder
}

// GetTempFolder returns the folder for files still being written. It follows the backups folder
// to its disk, so finishing a local backup stays a rename.
func GetTempFolder() string {
	return getActiveFolders().tempFolder
}

// GetFoldersHoldingBackups lists where a stored file may be, the current folder first.
func GetFoldersHoldingBackups() []string {
	return getActiveFolders().settings.GetFoldersHoldingBackups()
}

// GetTempFoldersHoldingFiles lists the temp folders an interrupted write may have left a file in,
// the current one first.
func GetTempFoldersHoldingFiles() []string {
	folders := getActiveFolders()
	tempFolders := []string{folders.tempFolder}

	for _, previousFolder := range folders.settings.PreviousFolders {
		previousTempFolder := getTempFolderFor(previousFolder)
		if !slices.Contains(tempFolders, previousTempFolder) {
			tempFolders = append(tempFolders, previousTempFolder)
		}
	}

	return tempFolders
}

// UseSettings switches the running process to settings at once: new files go to the new current
// folder, and files written before are still found in the previous ones.
func UseSettings(settings settingsfile.Settings) {
	initActiveState()

	activeState.Store(&activeFolders{
		settings:   settings,
		tempFolder: getTempFolderFor(settings.CurrentFolder),
	})
}

func getActiveFolders() *activeFolders {
	initActiveState()

	return activeState.Load()
}

// The environment decides the current folder, as it does for the rest of the process; the
// settings file adds the folders used before it.
func loadActiveFolders() *activeFolders {
	environment := config.GetEnv()
	folders := &activeFolders{
		settings:   settingsfile.Settings{CurrentFolder: environment.DataFolder},
		tempFolder: environment.TempFolder,
	}

	settingsFilePath, isConfigured := GetSettingsFilePathFromEnvironment()
	if !isConfigured {
		return folders
	}

	settings, err := settingsfile.ReadSettings(settingsFilePath)
	if err != nil {
		logger.GetLogger().Error(
			"cannot read the backups folder settings; backups in earlier backups folders will not be found",
			"error", err,
		)

		return folders
	}

	folders.settings = settings.WithCurrentFolder(environment.DataFolder)

	return folders
}

func getTempFolderFor(backupsFolder string) string {
	if tempFolder := strings.TrimSpace(os.Getenv(locations.TempFolderEnvVariable)); tempFolder != "" {
		if absoluteTempFolder, err := filepath.Abs(tempFolder); err == nil {
			return absoluteTempFolder
		}

		return tempFolder
	}

	return filepath.Join(backupsFolder, "temp")
}
