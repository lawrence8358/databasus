package system_availability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/util/availability"
)

func buildAvailabilityRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	v1 := router.Group("/api/v1")
	GetAvailabilityController().RegisterRoutes(v1)
	v1.GET(
		"/test/requires-verification-agent",
		RequireFeature(availability.FeatureVerificationAgent),
		func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) },
	)

	return router
}

func getAvailabilityViaAPI(t *testing.T, router *gin.Engine) availability.Availability {
	t.Helper()

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/system/availability", nil))
	require.Equal(t, http.StatusOK, response.Code)

	var reportedAvailability availability.Availability
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &reportedAvailability))

	return reportedAvailability
}

func Test_GetAvailability_OnDockerImage_ReportsEveryEngineAndFeatureAvailable(t *testing.T) {
	reportedAvailability := getAvailabilityViaAPI(t, buildAvailabilityRouter())

	reportedEngines := make([]availability.Engine, 0, len(reportedAvailability.Engines))
	for _, engineAvailability := range reportedAvailability.Engines {
		reportedEngines = append(reportedEngines, engineAvailability.Engine)
		assert.True(t, engineAvailability.IsAvailable, engineAvailability.Engine)
		assert.Empty(t, engineAvailability.UnavailableReason, engineAvailability.Engine)
	}

	reportedFeatures := make([]availability.Feature, 0, len(reportedAvailability.Features))
	for _, featureAvailability := range reportedAvailability.Features {
		reportedFeatures = append(reportedFeatures, featureAvailability.Feature)
		assert.True(t, featureAvailability.IsAvailable, featureAvailability.Feature)
	}

	assert.ElementsMatch(t, []availability.Engine{
		availability.EnginePostgresql,
		availability.EngineMysql,
		availability.EngineMariadb,
		availability.EngineMongodb,
	}, reportedEngines)
	assert.ElementsMatch(t, []availability.Feature{
		availability.FeaturePhysicalBackup,
		availability.FeaturePitr,
		availability.FeatureVerificationAgent,
	}, reportedFeatures)
}

func Test_GetAvailability_WhenEngineAndFeatureAreUnavailable_ReportsThemWithReason(t *testing.T) {
	availability.MarkUnavailableForTest(
		t,
		[]availability.Engine{availability.EngineMysql},
		[]availability.Feature{availability.FeatureVerificationAgent},
	)

	reportedAvailability := getAvailabilityViaAPI(t, buildAvailabilityRouter())

	for _, engineAvailability := range reportedAvailability.Engines {
		isMysql := engineAvailability.Engine == availability.EngineMysql
		assert.Equal(t, !isMysql, engineAvailability.IsAvailable, engineAvailability.Engine)

		if isMysql {
			assert.NotEmpty(t, engineAvailability.UnavailableReason)
		}
	}

	for _, featureAvailability := range reportedAvailability.Features {
		isVerificationAgent := featureAvailability.Feature == availability.FeatureVerificationAgent
		assert.Equal(t, !isVerificationAgent, featureAvailability.IsAvailable, featureAvailability.Feature)
	}
}

func Test_RequireFeature_WhenFeatureIsUnavailable_RefusesBeforeTheHandlerRuns(t *testing.T) {
	availability.MarkUnavailableForTest(
		t, nil, []availability.Feature{availability.FeatureVerificationAgent},
	)

	response := httptest.NewRecorder()
	buildAvailabilityRouter().ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/test/requires-verification-agent", nil),
	)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), "verification agent is not available on this instance")
}

func Test_RequireFeature_WhenFeatureIsAvailable_RunsTheHandler(t *testing.T) {
	response := httptest.NewRecorder()
	buildAvailabilityRouter().ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/v1/test/requires-verification-agent", nil),
	)

	assert.Equal(t, http.StatusNoContent, response.Code)
}
