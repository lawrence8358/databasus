package system_availability

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"databasus-backend/internal/util/availability"
)

type AvailabilityController struct{}

func (c *AvailabilityController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/system/availability", c.GetAvailability)
}

// GetAvailability
// @Summary Get engine and feature availability
// @Description Returns which database engines and which features this instance can run. An engine is unavailable when the platform does not support it or its client tools are missing or broken; a feature is unavailable when the platform does not support it.
// @Tags system/availability
// @Produce json
// @Security BearerAuth
// @Success 200 {object} availability.Availability
// @Failure 401 {object} map[string]string
// @Router /system/availability [get]
func (c *AvailabilityController) GetAvailability(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, availability.GetAvailability())
}
