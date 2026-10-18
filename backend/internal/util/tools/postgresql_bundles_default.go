//go:build !windows

package tools

// Every supported server line is dumped by a client from its own line.
var postgresqlBundledVersions = []PostgresqlVersion{
	PostgresqlVersion12,
	PostgresqlVersion13,
	PostgresqlVersion14,
	PostgresqlVersion15,
	PostgresqlVersion16,
	PostgresqlVersion17,
	PostgresqlVersion18,
}

var postgresqlRequired = []string{
	string(PostgresqlExecutablePgDump),
	string(PostgresqlExecutablePsql),
}

// postgresqlRequiredV17Plus extends the base set with the physical-backup
// binaries. pg_basebackup --incremental and pg_combinebackup are PG 17+ only,
// so pre-17 versions stay on the smaller list.
var postgresqlRequiredV17Plus = []string{
	string(PostgresqlExecutablePgDump),
	string(PostgresqlExecutablePsql),
	string(PostgresqlExecutablePgBasebackup),
	string(PostgresqlExecutablePgReceivewal),
	string(PostgresqlExecutablePgCombinebackup),
}

func getPostgresqlBundleVersion(serverVersion PostgresqlVersion) PostgresqlVersion {
	return serverVersion
}

func getPostgresqlRequiredForVersion(version PostgresqlVersion) []string {
	if version == PostgresqlVersion17 || version == PostgresqlVersion18 {
		return postgresqlRequiredV17Plus
	}

	return postgresqlRequired
}
