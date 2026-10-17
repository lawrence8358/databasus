package availability

import (
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"

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

type detection struct {
	availability        Availability
	toolResultsByEngine map[Engine][]tools.ToolCheckResult
}

// The installation does not change while the process runs, and the healthcheck re-reads the
// PostgreSQL bundle on its own, so availability is settled once.
var detectAvailability = sync.OnceValue(func() detection {
	return buildDetection(getPlatformUnsupportedEngines(), getPlatformUnsupportedFeatures())
})

var overriddenAvailability atomic.Pointer[Availability]

func GetAvailability() Availability {
	if overridden := overriddenAvailability.Load(); overridden != nil {
		return cloneAvailability(*overridden)
	}

	return cloneAvailability(detectAvailability().availability)
}

func RequireEngine(engine Engine) error {
	for _, engineAvailability := range GetAvailability().Engines {
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
	for _, featureAvailability := range GetAvailability().Features {
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
	detected := detectAvailability()

	for _, engineAvailability := range detected.availability.Engines {
		engineLogger := availabilityLogger.With("engine", engineAvailability.Engine)

		logClientToolResults(
			engineLogger,
			detected.toolResultsByEngine[engineAvailability.Engine],
			isShowLogs,
		)

		if !engineAvailability.IsAvailable {
			engineLogger.Warn(
				"database engine unavailable",
				"reason", engineAvailability.UnavailableReason,
			)
		}
	}

	for _, featureAvailability := range detected.availability.Features {
		if !featureAvailability.IsAvailable {
			availabilityLogger.Warn(
				"feature unavailable",
				"feature", featureAvailability.Feature,
				"reason", featureAvailability.UnavailableReason,
			)
		}
	}

	if !isEngineAvailableIn(detected.availability, EnginePostgresql) {
		availabilityLogger.Error("PostgreSQL client tools are required, exiting")
		logger.ExitAfterFlush(1)
	}
}

func buildDetection(unsupportedEngines []Engine, unsupportedFeatures []Feature) detection {
	unsupportedReason := fmt.Sprintf("not supported on %s", runtime.GOOS)

	detected := detection{
		toolResultsByEngine: make(map[Engine][]tools.ToolCheckResult),
	}

	for _, engineTools := range engineClientTools {
		if slices.Contains(unsupportedEngines, engineTools.engine) {
			detected.availability.Engines = append(detected.availability.Engines, EngineAvailability{
				Engine:            engineTools.engine,
				UnavailableReason: unsupportedReason,
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

		detected.availability.Engines = append(detected.availability.Engines, engineAvailability)
	}

	for _, feature := range orderedFeatures {
		featureAvailability := FeatureAvailability{Feature: feature, IsAvailable: true}

		switch {
		case slices.Contains(unsupportedFeatures, feature):
			featureAvailability = FeatureAvailability{
				Feature:           feature,
				UnavailableReason: unsupportedReason,
			}

		case feature == FeaturePitr && slices.Contains(unsupportedFeatures, FeaturePhysicalBackup):
			featureAvailability = FeatureAvailability{
				Feature:           feature,
				UnavailableReason: "requires physical backup, which is not available",
			}
		}

		detected.availability.Features = append(detected.availability.Features, featureAvailability)
	}

	return detected
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

func isEngineAvailableIn(availability Availability, engine Engine) bool {
	return slices.ContainsFunc(
		availability.Engines,
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

func cloneAvailability(availability Availability) Availability {
	return Availability{
		Engines:  slices.Clone(availability.Engines),
		Features: slices.Clone(availability.Features),
	}
}
