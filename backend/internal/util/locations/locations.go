// Package locations resolves where the application keeps its data and where it finds the files it
// ships with. It depends on nothing else in the module, so the logger and the configuration can
// both use it before either is initialized.
package locations

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	DataFolderEnvVariable    = "DATABASUS_DATA_DIR"
	BackupsFolderEnvVariable = "DATABASUS_BACKUPS_DIR"
	TempFolderEnvVariable    = "DATABASUS_TEMP_DIR"
	// BackupsFolderSettingsFileEnvVariable names the file that records the backups folder an
	// administrator chose after installation. The Windows launcher sets it; without it the backups
	// folder is fixed for the life of the process.
	BackupsFolderSettingsFileEnvVariable = "DATABASUS_BACKUPS_FOLDER_FILE"
)

// GetBackendRoot returns the nearest directory above the working directory that holds go.mod. A
// release image carries no go.mod, so there the walk ends at the filesystem root, which is what
// places the default data folder at /databasus-data.
func GetBackendRoot() string {
	backendRoot, err := os.Getwd()
	if err != nil {
		return "."
	}

	for {
		if _, err := os.Stat(filepath.Join(backendRoot, "go.mod")); err == nil {
			return backendRoot
		}

		parent := filepath.Dir(backendRoot)
		if parent == backendRoot {
			return backendRoot
		}

		backendRoot = parent
	}
}

func GetDataFolder() string {
	if dataFolder, isSet := lookupFolder(DataFolderEnvVariable); isSet {
		return dataFolder
	}

	return filepath.Join(filepath.Dir(GetBackendRoot()), "databasus-data")
}

func GetBackupsFolder() string {
	if backupsFolder, isSet := lookupFolder(BackupsFolderEnvVariable); isSet {
		return backupsFolder
	}

	return filepath.Join(GetDataFolder(), "backups")
}

// GetTempFolder returns the folder for files still being written. A configured backups folder may
// sit on another disk than the data folder, so the temp folder follows it there: finishing a local
// backup is then a rename on one disk instead of a copy across two. Without a configured backups
// folder the original layout stays, where both folders share the data volume.
func GetTempFolder() string {
	if tempFolder, isSet := lookupFolder(TempFolderEnvVariable); isSet {
		return tempFolder
	}

	if backupsFolder, isSet := lookupFolder(BackupsFolderEnvVariable); isSet {
		return filepath.Join(backupsFolder, "temp")
	}

	return filepath.Join(GetDataFolder(), "temp")
}

// FindAppFolder locates a folder shipped next to the executable, such as migrations or the built
// Web UI. A service is started from an arbitrary working directory, so the executable's own folder
// is checked first. `go run` builds into a temporary folder, so a development checkout falls back
// to the backend root. When neither holds the folder, the path next to the executable is returned
// so the caller's error names where the folder was expected.
func FindAppFolder(relativePath string) string {
	candidateRoots := make([]string, 0, 2)

	executableFolder, isFound := getExecutableFolder()
	if isFound {
		candidateRoots = append(candidateRoots, executableFolder)
	}

	candidateRoots = append(candidateRoots, GetBackendRoot())

	for _, candidateRoot := range candidateRoots {
		candidatePath := filepath.Join(candidateRoot, relativePath)
		if info, err := os.Stat(candidatePath); err == nil && info.IsDir() {
			return candidatePath
		}
	}

	return filepath.Join(candidateRoots[0], relativePath)
}

func getExecutableFolder() (string, bool) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", false
	}

	// The image exposes the binary through a symlink on PATH; its target sits beside the shipped
	// folders, the link does not.
	if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
		executablePath = resolvedPath
	}

	return filepath.Dir(executablePath), true
}

func lookupFolder(envVariable string) (string, bool) {
	folder := strings.TrimSpace(os.Getenv(envVariable))
	if folder == "" {
		return "", false
	}

	if absoluteFolder, err := filepath.Abs(folder); err == nil {
		return absoluteFolder, true
	}

	return folder, true
}
