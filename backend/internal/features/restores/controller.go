package restores

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	restores_core "databasus-backend/internal/features/restores/core"
	users_middleware "databasus-backend/internal/features/users/middleware"
)

type RestoreController struct {
	restoreService *RestoreService
}

func (c *RestoreController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/restores/:backupId", c.GetRestores)
	router.POST("/restores/:backupId/restore", c.RestoreBackup)
	router.POST("/restores/:backupId/compatibility", c.CheckRestoreCompatibility)
	router.POST("/restores/cancel/:restoreId", c.CancelRestore)
}

// GetRestores
// @Summary Get restores for a backup
// @Description Get all restores for a specific backup
// @Tags restores
// @Produce json
// @Param backupId path string true "Backup ID"
// @Success 200 {array} restores_core.Restore
// @Failure 400
// @Failure 401
// @Router /restores/{backupId} [get]
func (c *RestoreController) GetRestores(ctx *gin.Context) {
	user, ok := users_middleware.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	backupID, err := uuid.Parse(ctx.Param("backupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup ID"})
		return
	}

	restores, err := c.restoreService.GetRestores(ctx.Request.Context(), user, backupID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, restores)
}

// RestoreBackup
// @Summary Restore a backup
// @Description Start a restore process for a specific backup
// @Tags restores
// @Param backupId path string true "Backup ID"
// @Success 200 {object} map[string]string
// @Failure 400
// @Failure 401
// @Router /restores/{backupId}/restore [post]
func (c *RestoreController) RestoreBackup(ctx *gin.Context) {
	user, ok := users_middleware.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	backupID, err := uuid.Parse(ctx.Param("backupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup ID"})
		return
	}

	var requestDTO restores_core.RestoreBackupRequest
	if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.restoreService.RestoreBackupWithAuth(ctx.Request.Context(), user, backupID, requestDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "restore started successfully"})
}

// CheckRestoreCompatibility
// @Summary Check whether a restore target is older than the dump client
// @Description Detects the target server version and reports whether it is older than the PostgreSQL client that produced the dump. The restore stays allowed; this only drives a warning.
// @Tags restores
// @Accept json
// @Produce json
// @Param backupId path string true "Backup ID"
// @Param request body restores_core.RestoreBackupRequest true "Restore target"
// @Success 200 {object} restores_core.RestoreCompatibilityResponse
// @Failure 400
// @Failure 401
// @Router /restores/{backupId}/compatibility [post]
func (c *RestoreController) CheckRestoreCompatibility(ctx *gin.Context) {
	user, ok := users_middleware.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	backupID, err := uuid.Parse(ctx.Param("backupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup ID"})
		return
	}

	var requestDTO restores_core.RestoreBackupRequest
	if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.restoreService.CheckRestoreCompatibility(
		ctx.Request.Context(),
		user,
		backupID,
		requestDTO,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// CancelRestore
// @Summary Cancel an in-progress restore
// @Description Cancel a restore that is currently in progress
// @Tags restores
// @Param restoreId path string true "Restore ID"
// @Success 204
// @Failure 400
// @Failure 401
// @Router /restores/cancel/{restoreId} [post]
func (c *RestoreController) CancelRestore(ctx *gin.Context) {
	user, ok := users_middleware.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	restoreID, err := uuid.Parse(ctx.Param("restoreId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid restore ID"})
		return
	}

	if err := c.restoreService.CancelRestore(ctx.Request.Context(), user, restoreID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
