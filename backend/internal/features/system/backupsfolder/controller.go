package system_backupsfolder

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	users_enums "databasus-backend/internal/features/users/enums"
	users_middleware "databasus-backend/internal/features/users/middleware"
)

type BackupsFolderController struct {
	backupsFolderService *BackupsFolderService
}

func (c *BackupsFolderController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET(
		"/system/backups-folder",
		users_middleware.RequireRole(users_enums.UserRoleAdmin),
		c.GetBackupsFolder,
	)
	router.PUT(
		"/system/backups-folder",
		users_middleware.RequireRole(users_enums.UserRoleAdmin),
		c.ChangeBackupsFolder,
	)
}

// GetBackupsFolder
// @Summary Get the local backups folder (admin only)
// @Description Returns the folder local storage writes new backups to and the folders used before it, which still hold older backups. Only an installation made by the Windows installer can change it; elsewhere isChangeable is false and no folder is returned
// @Tags system/backups-folder
// @Produce json
// @Security BearerAuth
// @Success 200 {object} BackupsFolderResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /system/backups-folder [get]
func (c *BackupsFolderController) GetBackupsFolder(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.backupsFolderService.GetBackupsFolder())
}

// ChangeBackupsFolder
// @Summary Change the local backups folder (admin only)
// @Description Makes new local backups go to another folder at once. The folder must be an absolute path on a local fixed disk, not a drive root, a Windows or program folder, or inside the program data or program folder; it is created when missing and must be writable by the service account. Existing backups are not moved and stay restorable from their folder
// @Tags system/backups-folder
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangeBackupsFolderRequest true "New backups folder"
// @Success 200 {object} BackupsFolderResponse
// @Failure 400 {object} map[string]string "Folder refused, or the installation cannot change it"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /system/backups-folder [put]
func (c *BackupsFolderController) ChangeBackupsFolder(ctx *gin.Context) {
	user, isOk := users_middleware.GetUserFromContext(ctx)
	if !isOk {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request ChangeBackupsFolderRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	response, err := c.backupsFolderService.ChangeBackupsFolder(ctx.Request.Context(), user, request)

	switch {
	case errors.Is(err, ErrBackupsFolderNotChangeable), errors.Is(err, ErrBackupsFolderRefused):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case err != nil:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change the backups folder"})
	default:
		ctx.JSON(http.StatusOK, response)
	}
}
