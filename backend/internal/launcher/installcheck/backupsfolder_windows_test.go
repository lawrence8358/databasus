//go:build windows

package installcheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func Test_CheckBackupsFolder_WhenFolderIsOnLocalDisk_ReturnsNil(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "backups")

	assert.NoError(t, CheckBackupsFolder(folder, OwnFolders{}))
}

func Test_CheckBackupsFolder_WithRefusedFolders_ExplainsWhy(t *testing.T) {
	windowsFolder := getKnownFolderForTest(t, windows.FOLDERID_Windows)
	programFilesFolder := getKnownFolderForTest(t, windows.FOLDERID_ProgramFiles)
	programDataFolder := getKnownFolderForTest(t, windows.FOLDERID_ProgramData)
	dataFolder := t.TempDir()
	installFolder := t.TempDir()

	testCases := []struct {
		name            string
		folder          string
		expectedMessage string
	}{
		{"relative path", `backups`, "not an absolute path"},
		{"network path", `\\server\share\backups`, "network path"},
		{"drive root", filepath.VolumeName(dataFolder) + `\`, "root of a drive"},
		{"Windows folder", filepath.Join(windowsFolder, "Temp", "backups"), "system or program folder"},
		{"Program Files", filepath.Join(programFilesFolder, "Databasus backups"), "system or program folder"},
		{"ProgramData itself", programDataFolder, "shared by many programs"},
		{"line break", filepath.Join(dataFolder, "a") + "\nprevious=x", "line break"},
		{"inside install folder", filepath.Join(installFolder, "backups"), "program folder"},
		{"metadata database folder", filepath.Join(dataFolder, "pgdata"), "program data folder"},
		{"program data folder itself", dataFolder, "program data folder"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := CheckBackupsFolder(testCase.folder, OwnFolders{
				DataFolder:    dataFolder,
				InstallFolder: installFolder,
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedMessage)
		})
	}
}

func Test_CheckBackupsFolder_WhenFolderIsDefaultUnderDataFolder_ReturnsNil(t *testing.T) {
	dataFolder := t.TempDir()

	assert.NoError(t, CheckBackupsFolder(filepath.Join(dataFolder, "backups"), OwnFolders{DataFolder: dataFolder}))
}

func Test_CheckBackupsFolder_WhenJunctionLeadsIntoProtectedFolder_ReturnsError(t *testing.T) {
	dataFolder := t.TempDir()
	junction := filepath.Join(t.TempDir(), "looks-harmless")

	output, err := exec.Command("cmd", "/c", "mklink", "/J", junction, dataFolder).CombinedOutput()
	require.NoError(t, err, string(output))
	t.Cleanup(func() { _ = os.Remove(junction) })

	err = CheckBackupsFolder(filepath.Join(junction, "pgdata"), OwnFolders{DataFolder: dataFolder})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "program data folder")
}

func getKnownFolderForTest(t *testing.T, folderID *windows.KNOWNFOLDERID) string {
	t.Helper()

	folder, err := windows.KnownFolderPath(folderID, windows.KF_FLAG_DEFAULT)
	require.NoError(t, err)

	return folder
}
