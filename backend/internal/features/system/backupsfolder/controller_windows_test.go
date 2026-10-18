//go:build windows

package system_backupsfolder

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"

	users_enums "databasus-backend/internal/features/users/enums"
	users_testing "databasus-backend/internal/features/users/testing"
	"databasus-backend/internal/util/backupsfolder"
	"databasus-backend/internal/util/backupsfolder/settingsfile"
	test_utils "databasus-backend/internal/util/testing"
)

func Test_ChangeBackupsFolder_WhenFolderIsValid_CreatesRecordsAndAuditsIt(t *testing.T) {
	router := createBackupsFolderTestRouter(t)
	admin := users_testing.CreateTestUser(t.Context(), users_enums.UserRoleAdmin)
	settingsFilePath := enableBackupsFolderSetting(t)
	folderBefore := backupsfolder.GetCurrentFolder()
	newFolder := filepath.Join(t.TempDir(), "moved", "backups")

	var response BackupsFolderResponse
	test_utils.MakePutRequestAndUnmarshal(
		t,
		router,
		backupsFolderURL,
		"Bearer "+admin.Token,
		ChangeBackupsFolderRequest{Folder: newFolder},
		http.StatusOK,
		&response,
	)

	assert.True(t, response.IsChangeable)
	assert.Equal(t, newFolder, response.CurrentFolder)
	assert.Equal(t, folderBefore, response.PreviousFolders[0])
	assert.DirExists(t, newFolder)
	assert.Equal(t, newFolder, backupsfolder.GetCurrentFolder())
	assert.Equal(t, filepath.Join(newFolder, "temp"), backupsfolder.GetTempFolder())

	recordedSettings, err := settingsfile.ReadSettings(settingsFilePath)
	require.NoError(t, err)
	assert.Equal(t, newFolder, recordedSettings.CurrentFolder)
	assert.Equal(t, folderBefore, recordedSettings.PreviousFolders[0])

	entries, err := os.ReadDir(newFolder)
	require.NoError(t, err)
	assert.Empty(t, entries, "the write probe must not leave files behind")

	assert.True(t, users_testing.GetAuditLogRecorder().HasEntryContaining(
		"Backups folder changed from", folderBefore, newFolder,
	))

	var readBack BackupsFolderResponse
	test_utils.MakeGetRequestAndUnmarshal(t, router, backupsFolderURL, "Bearer "+admin.Token, http.StatusOK, &readBack)
	assert.Equal(t, response, readBack)
}

func Test_ChangeBackupsFolder_WithRefusedFolders_ReturnsBadRequestAndKeepsFolder(t *testing.T) {
	router := createBackupsFolderTestRouter(t)
	admin := users_testing.CreateTestUser(t.Context(), users_enums.UserRoleAdmin)
	settingsFilePath := enableBackupsFolderSetting(t)
	folderBefore := backupsfolder.GetCurrentFolder()

	windowsFolder, err := windows.KnownFolderPath(windows.FOLDERID_Windows, windows.KF_FLAG_DEFAULT)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		folder          string
		expectedMessage string
	}{
		{"relative path", `backups`, "not an absolute path"},
		{"network path", `\\fileserver\share\backups`, "network path"},
		{"drive root", filepath.VolumeName(t.TempDir()) + `\`, "root of a drive"},
		{"Windows folder", filepath.Join(windowsFolder, "DatabasusBackups"), "system or program folder"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := test_utils.MakePutRequest(
				t,
				router,
				backupsFolderURL,
				"Bearer "+admin.Token,
				ChangeBackupsFolderRequest{Folder: testCase.folder},
				http.StatusBadRequest,
			)

			assert.Contains(t, string(response.Body), testCase.expectedMessage)
			assert.Equal(t, folderBefore, backupsfolder.GetCurrentFolder())
			assert.NoFileExists(t, settingsFilePath)
		})
	}

	assert.NoDirExists(t, filepath.Join(windowsFolder, "DatabasusBackups"))
}
