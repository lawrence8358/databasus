package availability

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"databasus-backend/internal/util/tools"
)

var allEngines = []Engine{EnginePostgresql, EngineMysql, EngineMariadb, EngineMongodb}

func Test_BuildEngineDetection_WhenPlatformSupportsNoEngine_MarksEveryEngineUnavailableWithoutProbingTools(
	t *testing.T,
) {
	detected := buildEngineDetection(allEngines)

	assert.Len(t, detected.engines, len(allEngines))
	for _, engineAvailability := range detected.engines {
		assert.False(t, engineAvailability.IsAvailable, engineAvailability.Engine)
		assert.Contains(t, engineAvailability.UnavailableReason, "not supported on")
	}

	assert.Empty(t, detected.toolResultsByEngine)
}

func Test_BuildFeatureAvailabilities_WhenPhysicalBackupIsUnsupported_MarksPitrUnavailableToo(
	t *testing.T,
) {
	featureAvailabilities := buildFeatureAvailabilities([]Feature{FeaturePhysicalBackup})

	featuresByName := make(map[Feature]FeatureAvailability)
	for _, featureAvailability := range featureAvailabilities {
		featuresByName[featureAvailability.Feature] = featureAvailability
	}

	assert.False(t, featuresByName[FeaturePhysicalBackup].IsAvailable)
	assert.False(t, featuresByName[FeaturePitr].IsAvailable)
	assert.Contains(t, featuresByName[FeaturePitr].UnavailableReason, "requires physical backup")
	assert.True(t, featuresByName[FeatureVerificationAgent].IsAvailable)
}

func Test_IsClientToolsUsable_ForBundleOutcomes_ReportsWhetherTheEngineCanRun(t *testing.T) {
	brokenBundleErrors := []error{errors.New("client command not found")}

	testCases := []struct {
		name        string
		toolResults []tools.ToolCheckResult
		isUsable    bool
	}{
		{
			name: "non-fatal engine with one working bundle",
			toolResults: []tools.ToolCheckResult{
				{Version: "5.7", Errors: brokenBundleErrors},
				{Version: "8.0"},
			},
			isUsable: true,
		},
		{
			name: "non-fatal engine with every bundle broken",
			toolResults: []tools.ToolCheckResult{
				{Version: "5.7", Errors: brokenBundleErrors},
				{Version: "8.0", Errors: brokenBundleErrors},
			},
			isUsable: false,
		},
		{
			name: "fatal engine with one broken bundle",
			toolResults: []tools.ToolCheckResult{
				{Version: "17", IsFatal: true},
				{Version: "18", IsFatal: true, Errors: brokenBundleErrors},
			},
			isUsable: false,
		},
		{
			name:        "engine without bundles",
			toolResults: nil,
			isUsable:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.isUsable, isClientToolsUsable(testCase.toolResults))
		})
	}
}
