package availability

// The Windows build ships only the PostgreSQL client tools and runs logical backups only; the
// physical backup path and the verification agent depend on Linux-only tooling.
func getPlatformUnsupportedEngines() []Engine {
	return []Engine{EngineMysql, EngineMariadb, EngineMongodb}
}

func getPlatformUnsupportedFeatures() []Feature {
	return []Feature{FeaturePhysicalBackup, FeaturePitr, FeatureVerificationAgent}
}
