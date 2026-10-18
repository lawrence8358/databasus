package tools

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Every PostgreSQL client answers --version with "<tool> (PostgreSQL) 17.10".
var postgresqlVersionPattern = regexp.MustCompile(`\(PostgreSQL\)\s+(\d[\d.]*)`)

// GetPostgresqlExecutable returns the absolute path to the PostgreSQL client
// binary (e.g. pg_dump, pg_restore, psql) that serves a server of the given
// version.
func GetPostgresqlExecutable(
	serverVersion PostgresqlVersion,
	executable PostgresqlExecutable,
) string {
	return getExecutablePath(
		getPostgresqlBinDir(getPostgresqlBundleVersion(serverVersion)),
		string(executable),
	)
}

func getPostgresqlBinDir(version PostgresqlVersion) string {
	return filepath.Join(
		AssetsToolsDir(),
		"postgresql",
		fmt.Sprintf("postgresql-%s", version),
		"bin",
	)
}

// checkPostgresql verifies every bundled PG client's bin directory. PG is
// fatal-tier — the app reads the version from each managed database and must
// be able to invoke the client that serves it.
func checkPostgresql() []ToolCheckResult {
	results := make([]ToolCheckResult, 0, len(postgresqlBundledVersions))

	for _, v := range postgresqlBundledVersions {
		binDir := getPostgresqlBinDir(v)

		results = append(results, ToolCheckResult{
			Db:      DbEnginePostgresql,
			Version: string(v),
			BinDir:  binDir,
			Errors: runBinDirChecks(
				binDir,
				getPostgresqlRequiredForVersion(v),
				string(v),
				parsePostgresqlClientVersion,
			),
			IsFatal: true,
		})
	}

	return results
}

// parsePostgresqlClientVersion reads the release the client reports out of
// its --version output.
func parsePostgresqlClientVersion(output string) (string, error) {
	match := postgresqlVersionPattern.FindStringSubmatch(output)
	if match == nil {
		return "", fmt.Errorf(
			"could not read a PostgreSQL version out of %q",
			strings.TrimSpace(output),
		)
	}

	return match[1], nil
}

// EscapePgpassField escapes special characters for the .pgpass file format.
// PostgreSQL requires backslash → \\ and colon → \:; newlines and carriage
// returns are stripped to prevent format corruption.
func EscapePgpassField(field string) string {
	field = strings.ReplaceAll(field, "\r", "")
	field = strings.ReplaceAll(field, "\n", "")
	field = strings.ReplaceAll(field, "\\", "\\\\")
	field = strings.ReplaceAll(field, ":", "\\:")

	return field
}
