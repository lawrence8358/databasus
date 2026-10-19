package locations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Folders_WhenNothingConfigured_KeepTheOriginalLayout(t *testing.T) {
	clearFolderEnvVariables(t)

	defaultDataFolder := filepath.Join(filepath.Dir(GetBackendRoot()), "databasus-data")

	assert.Equal(t, defaultDataFolder, GetDataFolder())
	assert.Equal(t, filepath.Join(defaultDataFolder, "backups"), GetBackupsFolder())
	assert.Equal(t, filepath.Join(defaultDataFolder, "temp"), GetTempFolder())
}

func Test_Folders_WithOnlyDataFolderConfigured_PlaceEverythingUnderIt(t *testing.T) {
	clearFolderEnvVariables(t)
	dataFolder := t.TempDir()
	t.Setenv(DataFolderEnvVariable, dataFolder)

	assert.Equal(t, dataFolder, GetDataFolder())
	assert.Equal(t, filepath.Join(dataFolder, "backups"), GetBackupsFolder())
	assert.Equal(t, filepath.Join(dataFolder, "temp"), GetTempFolder())
}

func Test_Folders_WithBackupsFolderConfigured_PutTempUnderBackups(t *testing.T) {
	clearFolderEnvVariables(t)
	dataFolder := t.TempDir()
	backupsFolder := t.TempDir()
	t.Setenv(DataFolderEnvVariable, dataFolder)
	t.Setenv(BackupsFolderEnvVariable, backupsFolder)

	assert.Equal(t, dataFolder, GetDataFolder())
	assert.Equal(t, backupsFolder, GetBackupsFolder())
	assert.Equal(t, filepath.Join(backupsFolder, "temp"), GetTempFolder())
}

func Test_GetTempFolder_WithTempFolderConfigured_UsesIt(t *testing.T) {
	clearFolderEnvVariables(t)
	tempFolder := t.TempDir()
	t.Setenv(BackupsFolderEnvVariable, t.TempDir())
	t.Setenv(TempFolderEnvVariable, tempFolder)

	assert.Equal(t, tempFolder, GetTempFolder())
}

func Test_GetDataFolder_WithRelativePath_ResolvesAgainstWorkingDirectory(t *testing.T) {
	clearFolderEnvVariables(t)
	t.Setenv(DataFolderEnvVariable, "relative-data")

	workingDirectory, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(workingDirectory, "relative-data"), GetDataFolder())
}

func Test_FindAppFolder_FromForeignWorkingDirectory_FindsFolderBesideExecutable(t *testing.T) {
	executablePath, err := os.Executable()
	require.NoError(t, err)

	executableFolder := filepath.Dir(executablePath)
	if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
		executableFolder = filepath.Dir(resolvedPath)
	}

	shippedFolder, err := os.MkdirTemp(executableFolder, "shipped-")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(shippedFolder) })

	t.Chdir(t.TempDir())

	assert.Equal(t, shippedFolder, FindAppFolder(filepath.Base(shippedFolder)))
}

func Test_FindAppFolder_WhenAbsentBesideExecutable_FallsBackToBackendRoot(t *testing.T) {
	assert.Equal(t, filepath.Join(GetBackendRoot(), "migrations"), FindAppFolder("migrations"))
}

func clearFolderEnvVariables(t *testing.T) {
	t.Helper()

	for _, envVariable := range []string{
		DataFolderEnvVariable,
		BackupsFolderEnvVariable,
		TempFolderEnvVariable,
	} {
		t.Setenv(envVariable, "")
	}
}
