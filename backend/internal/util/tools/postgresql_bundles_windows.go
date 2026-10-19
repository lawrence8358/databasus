package tools

// The Windows build ships only the newest client and runs logical backups
// only. A newer pg_dump reads every older supported server, so one client
// serves them all; restoring its dump onto an older server is the known risk
// the restore screen warns about.
const postgresqlWindowsBundleVersion = PostgresqlVersion18

var postgresqlBundledVersions = []PostgresqlVersion{postgresqlWindowsBundleVersion}

var postgresqlRequired = []string{
	string(PostgresqlExecutablePgDump),
	string(PostgresqlExecutablePgRestore),
	string(PostgresqlExecutablePsql),
}

func getPostgresqlBundleVersion(_ PostgresqlVersion) PostgresqlVersion {
	return postgresqlWindowsBundleVersion
}

func getPostgresqlRequiredForVersion(_ PostgresqlVersion) []string {
	return postgresqlRequired
}

// The metadata database is pinned to one major so an upgrade of Databasus only swaps binaries. It
// shares the client bundle while both are on the same major; moving the clients to a newer major
// leaves this constant behind until a planned pg_upgrade of the metadata data directory.
const metadataPostgresqlVersion = PostgresqlVersion18

func GetMetadataPostgresqlVersion() PostgresqlVersion {
	return metadataPostgresqlVersion
}

func GetMetadataPostgresqlBinDir() string {
	return getPostgresqlBinDir(metadataPostgresqlVersion)
}
