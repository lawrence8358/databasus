//go:build windows

package system_backupsfolder

import (
	"fmt"
	"os"
	"path/filepath"

	"databasus-backend/internal/launcher/installcheck"
	"databasus-backend/internal/util/locations"
)

const isChangeSupported = true

// The backend runs as LocalSystem, so the folder is checked by the installer's own rules before
// anything is created, and the write probe runs again once the folder exists.
func prepareBackupsFolder(folder string) error {
	ownFolders := installcheck.OwnFolders{
		DataFolder:    locations.GetDataFolder(),
		InstallFolder: getExecutableFolder(),
	}

	if err := installcheck.CheckBackupsFolder(folder, ownFolders); err != nil {
		return err
	}

	if err := os.MkdirAll(folder, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", folder, err)
	}

	return installcheck.CheckBackupsFolder(folder, ownFolders)
}

func getExecutableFolder() string {
	executablePath, err := os.Executable()
	if err != nil {
		return ""
	}

	return filepath.Dir(executablePath)
}
