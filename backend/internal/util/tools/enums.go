package tools

import (
	"fmt"
	"strconv"
)

type PostgresqlExtension string

const (
	// needed for queries monitoring
	PostgresqlExtensionPgStatMonitor PostgresqlExtension = "pg_stat_statements"
)

type PostgresqlVersion string

const (
	PostgresqlVersion12 PostgresqlVersion = "12"
	PostgresqlVersion13 PostgresqlVersion = "13"
	PostgresqlVersion14 PostgresqlVersion = "14"
	PostgresqlVersion15 PostgresqlVersion = "15"
	PostgresqlVersion16 PostgresqlVersion = "16"
	PostgresqlVersion17 PostgresqlVersion = "17"
	PostgresqlVersion18 PostgresqlVersion = "18"
)

type PostgresqlExecutable string

const (
	PostgresqlExecutablePgDump          PostgresqlExecutable = "pg_dump"
	PostgresqlExecutablePgRestore       PostgresqlExecutable = "pg_restore"
	PostgresqlExecutablePsql            PostgresqlExecutable = "psql"
	PostgresqlExecutablePgBasebackup    PostgresqlExecutable = "pg_basebackup"
	PostgresqlExecutablePgReceivewal    PostgresqlExecutable = "pg_receivewal"
	PostgresqlExecutablePgCombinebackup PostgresqlExecutable = "pg_combinebackup"
)

func GetPostgresqlVersionEnum(version string) PostgresqlVersion {
	switch version {
	case "12":
		return PostgresqlVersion12
	case "13":
		return PostgresqlVersion13
	case "14":
		return PostgresqlVersion14
	case "15":
		return PostgresqlVersion15
	case "16":
		return PostgresqlVersion16
	case "17":
		return PostgresqlVersion17
	case "18":
		return PostgresqlVersion18
	default:
		panic(fmt.Sprintf("invalid postgresql version: %s", version))
	}
}

func IsBackupDbVersionHigherThanRestoreDbVersion(
	backupDbVersion, restoreDbVersion PostgresqlVersion,
) bool {
	backupDbVersionInt, err := strconv.Atoi(string(backupDbVersion))
	if err != nil {
		return false
	}

	restoreDbVersionInt, err := strconv.Atoi(string(restoreDbVersion))
	if err != nil {
		return false
	}

	return backupDbVersionInt > restoreDbVersionInt
}

// IsRestoreServerOlderThanDumpClient reports whether a server of the given
// version is older than the client that produced a dump. A dump from a newer
// client does not always restore onto an older server. An unknown client
// version (0) never counts as older.
func IsRestoreServerOlderThanDumpClient(
	dumpClientMajorVersion int,
	restoreServerVersion PostgresqlVersion,
) bool {
	if dumpClientMajorVersion <= 0 {
		return false
	}

	restoreServerMajorVersion, err := strconv.Atoi(string(restoreServerVersion))
	if err != nil {
		return false
	}

	return restoreServerMajorVersion < dumpClientMajorVersion
}
