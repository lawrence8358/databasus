package system_availability

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"databasus-backend/internal/util/availability"
)

// RequireFeature refuses a request before its handler runs, so work that needs an unavailable
// feature never starts and then fails halfway.
func RequireFeature(feature availability.Feature) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := availability.RequireFeature(feature); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.Next()
	}
}
