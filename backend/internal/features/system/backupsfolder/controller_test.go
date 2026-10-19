package system_backupsfolder

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	users_enums "databasus-backend/internal/features/users/enums"
	users_middleware "databasus-backend/internal/features/users/middleware"
	users_services "databasus-backend/internal/features/users/services"
	users_testing "databasus-backend/internal/features/users/testing"
	"databasus-backend/internal/util/backupsfolder"
	"databasus-backend/internal/util/backupsfolder/settingsfile"
	"databasus-backend/internal/util/locations"
	test_utils "databasus-backend/internal/util/testing"
)

const backupsFolderURL = "/api/v1/system/backups-folder"

func Test_GetBackupsFolder_WhenUserIsMember_ReturnsForbidden(t *testing.T) {
	router := createBackupsFolderTestRouter(t)
	member := users_testing.CreateTestUser(t.Context(), users_enums.UserRoleMember)
	enableBackupsFolderSetting(t)

	response := test_utils.MakeGetRequest(t, router, backupsFolderURL, "Bearer "+member.Token, http.StatusForbidden)

	assert.Contains(t, string(response.Body), "Insufficient permissions")
}

func Test_ChangeBackupsFolder_WhenUserIsMember_ReturnsForbiddenAndKeepsFolder(t *testing.T) {
	router := createBackupsFolderTestRouter(t)
	member := users_testing.CreateTestUser(t.Context(), users_enums.UserRoleMember)
	enableBackupsFolderSetting(t)
	folderBefore := backupsfolder.GetCurrentFolder()

	response := test_utils.MakePutRequest(
		t,
		router,
		backupsFolderURL,
		"Bearer "+member.Token,
		ChangeBackupsFolderRequest{Folder: filepath.Join(t.TempDir(), "backups")},
		http.StatusForbidden,
	)

	assert.Contains(t, string(response.Body), "Insufficient permissions")
	assert.Equal(t, folderBefore, backupsfolder.GetCurrentFolder())
}

func Test_ChangeBackupsFolder_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	router := createBackupsFolderTestRouter(t)

	test_utils.MakePutRequest(
		t,
		router,
		backupsFolderURL,
		"",
		ChangeBackupsFolderRequest{Folder: filepath.Join(t.TempDir(), "backups")},
		http.StatusUnauthorized,
	)
}

func Test_GetBackupsFolder_WhenLauncherNamedNoSettingsFile_IsNotChangeable(t *testing.T) {
	router := createBackupsFolderTestRouter(t)
	admin := users_testing.CreateTestUser(t.Context(), users_enums.UserRoleAdmin)
	t.Setenv(locations.BackupsFolderSettingsFileEnvVariable, "")

	var response BackupsFolderResponse
	test_utils.MakeGetRequestAndUnmarshal(
		t, router, backupsFolderURL, "Bearer "+admin.Token, http.StatusOK, &response,
	)

	assert.False(t, response.IsChangeable)
	assert.Empty(t, response.CurrentFolder)
}

func Test_ChangeBackupsFolder_WhenLauncherNamedNoSettingsFile_ReturnsBadRequest(t *testing.T) {
	router := createBackupsFolderTestRouter(t)
	admin := users_testing.CreateTestUser(t.Context(), users_enums.UserRoleAdmin)
	t.Setenv(locations.BackupsFolderSettingsFileEnvVariable, "")
	folderBefore := backupsfolder.GetCurrentFolder()

	response := test_utils.MakePutRequest(
		t,
		router,
		backupsFolderURL,
		"Bearer "+admin.Token,
		ChangeBackupsFolderRequest{Folder: filepath.Join(t.TempDir(), "backups")},
		http.StatusBadRequest,
	)

	assert.Contains(t, string(response.Body), ErrBackupsFolderNotChangeable.Error())
	assert.Equal(t, folderBefore, backupsfolder.GetCurrentFolder())
}

func createBackupsFolderTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/api/v1").Use(users_middleware.AuthMiddleware(users_services.GetUserService()))
	GetBackupsFolderController().RegisterRoutes(protected.(*gin.RouterGroup))

	auditLogRecorder := users_testing.InstallAuditLogRecorder()
	originalAuditLogWriter := backupsFolderService.auditLogWriter
	backupsFolderService.auditLogWriter = auditLogRecorder
	t.Cleanup(func() { backupsFolderService.auditLogWriter = originalAuditLogWriter })

	backupsfolder.RestoreFoldersAfterTest(t)

	return router
}

// enableBackupsFolderSetting makes the process look like one the Windows launcher started, with a
// settings file of its own for the test.
func enableBackupsFolderSetting(t *testing.T) string {
	t.Helper()

	settingsFilePath := settingsfile.GetSettingsFilePath(t.TempDir())
	t.Setenv(locations.BackupsFolderSettingsFileEnvVariable, settingsFilePath)

	return settingsFilePath
}
