package availability

import (
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"sync"

	"databasus-backend/internal/util/logger"
	"databasus-backend/internal/util/tools"
)

const clientToolsBrokenReason = "client tools are missing or broken, see the startup log"

var engineClientTools = []struct {
	engine      Engine
	toolsEngine tools.DbEngine
	displayName string
}{
	{EnginePostgresql, tools.DbEnginePostgresql, "PostgreSQL"},
	{EngineMysql, tools.DbEngineMysql, "MySQL"},
	{EngineMariadb, tools.DbEngineMariadb, "MariaDB"},
	{EngineMongodb, tools.DbEngineMongodb, "MongoDB"},
}

var featureDisplayNames = map[Feature]string{
	FeaturePhysicalBackup:    "physical backup",
	FeaturePitr:              "point-in-time recovery",
	FeatureVerificationAgent: "verification agent",
}

var orderedFeatures = []Feature{
	FeaturePhysicalBackup,
	FeaturePitr,
	FeatureVerificationAgent,
}

type engineDetection struct {
	engines             []EngineAvailability
	toolResultsByEngine map[Engine][]tools.ToolCheckResult
}

// The installation does not change while the process runs, and the healthcheck re-reads the
// PostgreSQL bundle on its own, so availability is settled once. Features are settled apart from
// engines so that asking about a feature never starts the client tools.
var (
	detectEngines = sync.OnceValue(func() engineDetection {
		return buildEngineDetection(getPlatformUnsupportedEngines())
	})
	detectFeatures = sync.OnceValue(func() []FeatureAvailability {
		return buildFeatureAvailabilities(getPlatformUnsupportedFeatures())
	})
)

func GetAvailability() Availability {
	return Availability{
		Engines:  getEngineAvailabilities(),
		Features: getFeatureAvailabilities(),
	}
}

func RequireEngine(engine Engine) error {
	for _, engineAvailability := range getEngineAvailabilities() {
		if engineAvailability.Engine != engine {
			continue
		}

		if engineAvailability.IsAvailable {
			return nil
		}

		return fmt.Errorf(
			"%w: %s is not available on this instance: %s",
			ErrEngineUnavailable,
			getEngineDisplayName(engine),
			engineAvailability.UnavailableReason,
		)
	}

	panic(fmt.Sprintf("unknown database engine %q", engine))
}

func RequireFeature(feature Feature) error {
	for _, featureAvailability := range getFeatureAvailabilities() {
		if featureAvailability.Feature != feature {
			continue
		}

		if featureAvailability.IsAvailable {
			return nil
		}

		return fmt.Errorf(
			"%w: %s is not available on this instance: %s",
			ErrFeatureUnavailable,
			featureDisplayNames[feature],
			featureAvailability.UnavailableReason,
		)
	}

	panic(fmt.Sprintf("unknown feature %q", feature))
}

// LogAvailabilityAndExitIfPostgresqlUnavailable runs at startup. Every other engine and feature
// only degrades: an instance that backs up no MySQL database has no reason to refuse to start over
// a missing mysqldump, while without a PostgreSQL client it cannot do anything at all.
func LogAvailabilityAndExitIfPostgresqlUnavailable(availabilityLogger *slog.Logger, isShowLogs bool) {
	detectedEngines := detectEngines()

	for _, engineAvailability := range detectedEngines.engines {
		engineLogger := availabilityLogger.With("engine", engineAvailability.Engine)

		logClientToolResults(
			engineLogger,
			detectedEngines.toolResultsByEngine[engineAvailability.Engine],
			isShowLogs,
		)

		if !engineAvailability.IsAvailable {
			engineLogger.Warn(
				"database engine unavailable",
				"reason", engineAvailability.UnavailableReason,
			)
		}
	}

	for _, featureAvailability := range detectFeatures() {
		if !featureAvailability.IsAvailable {
			availabilityLogger.Warn(
				"feature unavailable",
				"feature", featureAvailability.Feature,
				"reason", featureAvailability.UnavailableReason,
			)
		}
	}

	if !isEngineAvailableIn(detectedEngines.engines, EnginePostgresql) {
		availabilityLogger.Error("PostgreSQL client tools are required, exiting")
		logger.ExitAfterFlush(1)
	}
}

func getEngineAvailabilities() []EngineAvailability {
	engineAvailabilities := slices.Clone(detectEngines().engines)

	if override := testOverride.Load(); override != nil {
		for i, engineAvailability := range engineAvailabilities {
			if slices.Contains(override.unavailableEngines, engineAvailability.Engine) {
				engineAvailabilities[i] = EngineAvailability{
					Engine:            engineAvailability.Engine,
					UnavailableReason: testOverrideReason,
				}
			}
		}
	}

	return engineAvailabilities
}

func getFeatureAvailabilities() []FeatureAvailability {
	featureAvailabilities := slices.Clone(detectFeatures())

	if override := testOverride.Load(); override != nil {
		for i, featureAvailability := range featureAvailabilities {
			if slices.Contains(override.unavailableFeatures, featureAvailability.Feature) {
				featureAvailabilities[i] = FeatureAvailability{
					Feature:           featureAvailability.Feature,
					UnavailableReason: testOverrideReason,
				}
			}
		}
	}

	return featureAvailabilities
}

func getUnsupportedReason() string {
	return fmt.Sprintf("not supported on %s", runtime.GOOS)
}

func buildEngineDetection(unsupportedEngines []Engine) engineDetection {
	detected := engineDetection{
		toolResultsByEngine: make(map[Engine][]tools.ToolCheckResult),
	}

	for _, engineTools := range engineClientTools {
		if slices.Contains(unsupportedEngines, engineTools.engine) {
			detected.engines = append(detected.engines, EngineAvailability{
				Engine:            engineTools.engine,
				UnavailableReason: getUnsupportedReason(),
			})

			continue
		}

		toolResults := tools.CheckEngineClientTools(engineTools.toolsEngine)
		detected.toolResultsByEngine[engineTools.engine] = toolResults

		engineAvailability := EngineAvailability{Engine: engineTools.engine, IsAvailable: true}
		if !isClientToolsUsable(toolResults) {
			engineAvailability = EngineAvailability{
				Engine:            engineTools.engine,
				UnavailableReason: clientToolsBrokenReason,
			}
		}

		detected.engines = append(detected.engines, engineAvailability)
	}

	return detected
}

func buildFeatureAvailabilities(unsupportedFeatures []Feature) []FeatureAvailability {
	featureAvailabilities := make([]FeatureAvailability, 0, len(orderedFeatures))

	for _, feature := range orderedFeatures {
		featureAvailability := FeatureAvailability{Feature: feature, IsAvailable: true}

		switch {
		case slices.Contains(unsupportedFeatures, feature):
			featureAvailability = FeatureAvailability{
				Feature:           feature,
				UnavailableReason: getUnsupportedReason(),
			}

		case feature == FeaturePitr && slices.Contains(unsupportedFeatures, FeaturePhysicalBackup):
			featureAvailability = FeatureAvailability{
				Feature:           feature,
				UnavailableReason: "requires physical backup, which is not available",
			}
		}

		featureAvailabilities = append(featureAvailabilities, featureAvailability)
	}

	return featureAvailabilities
}

// A failing fatal-tier bundle makes the whole engine unusable, because the app picks the client by
// the managed database's version and must have every one. A non-fatal engine needs only one working
// bundle: a version without a client is refused when it is used.
func isClientToolsUsable(toolResults []tools.ToolCheckResult) bool {
	hasWorkingBundle := false

	for _, toolResult := range toolResults {
		if len(toolResult.Errors) == 0 {
			hasWorkingBundle = true

			continue
		}

		if toolResult.IsFatal {
			return false
		}
	}

	return hasWorkingBundle
}

func logClientToolResults(
	engineLogger *slog.Logger,
	toolResults []tools.ToolCheckResult,
	isShowLogs bool,
) {
	for _, toolResult := range toolResults {
		bundleLogger := engineLogger.With("version", toolResult.Version, "path", toolResult.BinDir)

		if len(toolResult.Errors) == 0 {
			if isShowLogs {
				bundleLogger.Info("client tools verified")
			}

			continue
		}

		for _, err := range toolResult.Errors {
			if toolResult.IsFatal {
				bundleLogger.Error("client tools check failed", "error", err)
			} else {
				bundleLogger.Warn("client tools check failed - support disabled", "error", err)
			}
		}
	}
}

func isEngineAvailableIn(engineAvailabilities []EngineAvailability, engine Engine) bool {
	return slices.ContainsFunc(
		engineAvailabilities,
		func(engineAvailability EngineAvailability) bool {
			return engineAvailability.Engine == engine && engineAvailability.IsAvailable
		},
	)
}

func getEngineDisplayName(engine Engine) string {
	for _, engineTools := range engineClientTools {
		if engineTools.engine == engine {
			return engineTools.displayName
		}
	}

	return string(engine)
}
