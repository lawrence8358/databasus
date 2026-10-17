package availability

type Engine string

const (
	EnginePostgresql Engine = "POSTGRESQL"
	EngineMysql      Engine = "MYSQL"
	EngineMariadb    Engine = "MARIADB"
	EngineMongodb    Engine = "MONGODB"
)

type Feature string

const (
	FeaturePhysicalBackup    Feature = "PHYSICAL_BACKUP"
	FeaturePitr              Feature = "PITR"
	FeatureVerificationAgent Feature = "VERIFICATION_AGENT"
)
