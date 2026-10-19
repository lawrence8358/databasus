package tools

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var mariadbClientVersions = []MariadbClientVersion{
	MariadbClientLegacy,
	MariadbClientModern,
	MariadbClient13,
}

var mariadbRequired = []string{
	string(MariadbExecutableMariadbDump),
	string(MariadbExecutableMariadb),
}

// Both MariaDB --version shapes put the server line immediately before
// "-MariaDB": "mariadb-dump  Ver 10.19 Distrib 10.6.28-MariaDB, for ..." on
// 10.6, "mariadb-dump from 13.0.2-MariaDB, client 10.20 for ..." on newer
// clients. The other number in each line is the client protocol version.
var mariadbVersionPattern = regexp.MustCompile(`(\d[\d.]*)-MariaDB`)

type MariadbVersion string

const (
	MariadbVersion55   MariadbVersion = "5.5"
	MariadbVersion101  MariadbVersion = "10.1"
	MariadbVersion102  MariadbVersion = "10.2"
	MariadbVersion103  MariadbVersion = "10.3"
	MariadbVersion104  MariadbVersion = "10.4"
	MariadbVersion105  MariadbVersion = "10.5"
	MariadbVersion106  MariadbVersion = "10.6"
	MariadbVersion1011 MariadbVersion = "10.11"
	MariadbVersion114  MariadbVersion = "11.4"
	MariadbVersion118  MariadbVersion = "11.8"
	MariadbVersion120  MariadbVersion = "12.0"
	MariadbVersion130  MariadbVersion = "13.0"
)

// MariadbClientVersion is the client tool version installed in assets.
type MariadbClientVersion string

const (
	// MariadbClientLegacy is used for older MariaDB servers (5.5, 10.1) that
	// don't have the generation_expression column in information_schema.columns.
	MariadbClientLegacy MariadbClientVersion = "10.6"
	// MariadbClientModern is used for servers from 10.2 up to the 12 line.
	MariadbClientModern MariadbClientVersion = "12.3"
	// MariadbClient13 serves the 13 line, which the 12.3 client predates.
	MariadbClient13 MariadbClientVersion = "13.0"
)

type MariadbExecutable string

const (
	MariadbExecutableMariadbDump MariadbExecutable = "mariadb-dump"
	MariadbExecutableMariadb     MariadbExecutable = "mariadb"
)

// GetMariadbClientVersionForServer returns the client version designated to
// talk to the given server version. The modern client uses queries
// referencing generation_expression (added in MariaDB 10.2), so older servers
// (5.5, 10.1) need the 10.6 legacy client; the 13 line gets its own client.
func GetMariadbClientVersionForServer(serverVersion MariadbVersion) MariadbClientVersion {
	switch serverVersion {
	case MariadbVersion55, MariadbVersion101:
		return MariadbClientLegacy
	case MariadbVersion130:
		return MariadbClient13
	default:
		return MariadbClientModern
	}
}

// GetMariadbExecutable returns the absolute path to a MariaDB client binary
// appropriate for the given server version. It fails when this architecture
// ships no bundle for the designated client, so the caller reports a
// supported-version problem instead of a missing file.
func GetMariadbExecutable(
	serverVersion MariadbVersion,
	executable MariadbExecutable,
) (string, error) {
	if err := RequireMariadbBundle(serverVersion); err != nil {
		return "", err
	}

	clientVersion := GetMariadbClientVersionForServer(serverVersion)

	return getExecutablePath(getMariadbBinDir(clientVersion), string(executable)), nil
}

// RequireMariadbBundle reports whether this architecture ships the client
// designated for the given server version, naming the architecture when it
// does not.
func RequireMariadbBundle(serverVersion MariadbVersion) error {
	clientVersion := GetMariadbClientVersionForServer(serverVersion)

	if errs := checkBinDir(getMariadbBinDir(clientVersion), mariadbRequired); len(errs) > 0 {
		return fmt.Errorf(
			"MariaDB %s is not supported on %s: no client is shipped for this architecture",
			serverVersion, archAssetsKey(),
		)
	}

	return nil
}

func getMariadbBinDir(clientVersion MariadbClientVersion) string {
	return filepath.Join(
		AssetsToolsDir(),
		"mariadb",
		fmt.Sprintf("mariadb-%s", clientVersion),
		"bin",
	)
}

// checkMariadb verifies every MariaDB client bundle. Non-fatal — a missing
// bundle disables that client tier.
func checkMariadb() []ToolCheckResult {
	results := make([]ToolCheckResult, 0, len(mariadbClientVersions))

	for _, cv := range mariadbClientVersions {
		binDir := getMariadbBinDir(cv)

		results = append(results, ToolCheckResult{
			Db:      DbEngineMariadb,
			Version: string(cv),
			BinDir:  binDir,
			Errors:  runBinDirChecks(binDir, mariadbRequired, string(cv), parseMariadbClientVersion),
			IsFatal: false,
		})
	}

	return results
}

// parseMariadbClientVersion reads the server release the client was built for
// out of its --version output.
func parseMariadbClientVersion(output string) (string, error) {
	match := mariadbVersionPattern.FindStringSubmatch(output)
	if match == nil {
		return "", fmt.Errorf("could not read a MariaDB version out of %q", strings.TrimSpace(output))
	}

	return match[1], nil
}

// IsMariadbBackupVersionHigherThanRestoreVersion reports whether a backup
// produced on backupVersion would be downgrade-restoring onto restoreVersion.
func IsMariadbBackupVersionHigherThanRestoreVersion(
	backupVersion, restoreVersion MariadbVersion,
) (bool, error) {
	order, err := compareReleaseLines(string(backupVersion), string(restoreVersion))
	if err != nil {
		return false, fmt.Errorf("cannot order MariaDB versions: %w", err)
	}

	return order > 0, nil
}

// EscapeMariadbPassword escapes special characters for the MariaDB .my.cnf
// file format.
func EscapeMariadbPassword(password string) string {
	password = strings.ReplaceAll(password, "\\", "\\\\")
	password = strings.ReplaceAll(password, "\"", "\\\"")
	return password
}
