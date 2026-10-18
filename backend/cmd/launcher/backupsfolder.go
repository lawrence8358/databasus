//go:build windows

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"databasus-backend/internal/launcher/installcheck"
	"databasus-backend/internal/util/backupsfolder/settingsfile"
)

// Setup runs this while the service is stopped. With -set it checks the folder by the same rules
// as check-install and the Web UI, then records it; the folder used before stays listed, so the
// backups written there can still be restored.
func setBackupsFolder(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher backups-folder", flag.ContinueOnError)

	dataFolder := flags.String("data-dir", getDefaultDataFolder(), "program data folder the service uses")
	newFolder := flags.String(
		"set",
		"",
		"backups folder to use from now on; without it the current settings are printed",
	)
	installFolder := flags.String(
		"install-dir",
		getLauncherFolder(),
		"program folder; the backups folder must stay out of it",
	)
	folderInUse := flags.String(
		"in-use",
		"",
		"backups folder the service has used so far, for a data folder that has not recorded one (an upgrade from a version without the settings file)",
	)

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	absoluteDataFolder, err := filepath.Abs(*dataFolder)
	if err != nil {
		return failWith(os.Stderr, fmt.Errorf("data folder: %w", err))
	}

	settingsFilePath := settingsfile.GetSettingsFilePath(absoluteDataFolder)

	settings, err := settingsfile.ReadSettings(settingsFilePath)
	isRecorded := err == nil

	switch {
	case isRecorded:
	case !errors.Is(err, settingsfile.ErrSettingsNotFound):
		return failWith(os.Stderr, err)
	case *folderInUse != "":
		settings = settingsfile.Settings{CurrentFolder: filepath.Clean(*folderInUse)}
	}

	if *newFolder == "" {
		if !isRecorded {
			return failWith(os.Stderr, fmt.Errorf(
				"%s does not exist yet; the service records the folder at its first start", settingsFilePath,
			))
		}

		printBackupsFolderSettings(settings)

		return 0
	}

	// The folder in use passed the rules of its time; an upgrade that keeps it must not fail on
	// rules added since.
	isUnchanged := settingsfile.IsSameFolder(settings.CurrentFolder, *newFolder)

	if isUnchanged && isRecorded {
		printBackupsFolderSettings(settings)

		return 0
	}

	if isUnchanged {
		return writeBackupsFolderSettings(absoluteDataFolder, settingsFilePath, settings)
	}

	err = installcheck.CheckBackupsFolder(*newFolder, installcheck.OwnFolders{
		DataFolder:    absoluteDataFolder,
		InstallFolder: *installFolder,
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, "Backups folder:", err)

		return 1
	}

	return writeBackupsFolderSettings(absoluteDataFolder, settingsFilePath, settings.WithCurrentFolder(*newFolder))
}

func writeBackupsFolderSettings(dataFolder, settingsFilePath string, settings settingsfile.Settings) int {
	if err := os.MkdirAll(dataFolder, 0o755); err != nil {
		return failWith(os.Stderr, fmt.Errorf("create the data folder: %w", err))
	}

	if err := settingsfile.WriteSettings(settingsFilePath, settings); err != nil {
		return failWith(os.Stderr, err)
	}

	printBackupsFolderSettings(settings)

	return 0
}

func printBackupsFolderSettings(settings settingsfile.Settings) {
	fmt.Printf("backups folder: %s\n", settings.CurrentFolder)

	for _, previousFolder := range settings.PreviousFolders {
		fmt.Printf("earlier backups folder: %s\n", previousFolder)
	}
}

func getLauncherFolder() string {
	executablePath, err := os.Executable()
	if err != nil {
		return ""
	}

	return filepath.Dir(executablePath)
}
