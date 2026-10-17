package tools

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var mysqlVersions = []MysqlVersion{
	MysqlVersion57,
	MysqlVersion80,
	MysqlVersion84,
	MysqlVersion9,
	MysqlVersion26,
}

var mysqlRequired = []string{
	string(MysqlExecutableMysqldump),
	string(MysqlExecutableMysql),
}

// MySQL 8.0 and later answer --version with "mysqldump  Ver 8.0.46 for ...",
// while 5.7 answers with its own tool version and puts the server line behind
// "Distrib": "mysqldump  Ver 10.13 Distrib 5.7.44, for ...".
var (
	mysqlDistribVersionPattern = regexp.MustCompile(`Distrib\s+(\d[\d.]*)`)
	mysqlVerVersionPattern     = regexp.MustCompile(`Ver\s+(\d[\d.]*)`)
)

type MysqlVersion string

const (
	MysqlVersion57 MysqlVersion = "5.7"
	MysqlVersion80 MysqlVersion = "8.0"
	MysqlVersion84 MysqlVersion = "8.4"
	MysqlVersion9  MysqlVersion = "9"
	MysqlVersion26 MysqlVersion = "26"
)

type MysqlExecutable string

const (
	MysqlExecutableMysqldump MysqlExecutable = "mysqldump"
	MysqlExecutableMysql     MysqlExecutable = "mysql"
)

// GetMysqlExecutable returns the absolute path to a MySQL client binary for
// the given version (mysqldump or mysql). It fails when this architecture
// ships no bundle for that version — upstream never built a 5.7 client for
// arm64 — so the caller reports a supported-version problem instead of a
// missing file.
func GetMysqlExecutable(version MysqlVersion, executable MysqlExecutable) (string, error) {
	if err := RequireMysqlBundle(version); err != nil {
		return "", err
	}

	return filepath.Join(getMysqlBinDir(version), string(executable)), nil
}

// RequireMysqlBundle reports whether this architecture ships a client for the
// given MySQL version, naming the architecture when it does not.
func RequireMysqlBundle(version MysqlVersion) error {
	if errs := checkBinDir(getMysqlBinDir(version), mysqlRequired); len(errs) > 0 {
		return fmt.Errorf(
			"MySQL %s is not supported on %s: no client is shipped for this architecture",
			version, archAssetsKey(),
		)
	}

	return nil
}

func getMysqlBinDir(version MysqlVersion) string {
	return filepath.Join(
		AssetsToolsDir(),
		"mysql",
		fmt.Sprintf("mysql-%s", version),
		"bin",
	)
}

// checkMysql verifies every supported MySQL version's bin directory. MySQL
// is non-fatal — a missing bundle disables that version's support.
func checkMysql() []ToolCheckResult {
	results := make([]ToolCheckResult, 0, len(mysqlVersions))

	for _, v := range mysqlVersions {
		binDir := getMysqlBinDir(v)

		results = append(results, ToolCheckResult{
			Db:      DbEngineMysql,
			Version: string(v),
			BinDir:  binDir,
			Errors:  runBinDirChecks(binDir, mysqlRequired, string(v), parseMysqlClientVersion),
			IsFatal: false,
		})
	}

	return results
}

// parseMysqlClientVersion reads the release the client reports out of its
// --version output.
func parseMysqlClientVersion(output string) (string, error) {
	if match := mysqlDistribVersionPattern.FindStringSubmatch(output); match != nil {
		return strings.TrimSuffix(match[1], "."), nil
	}

	if match := mysqlVerVersionPattern.FindStringSubmatch(output); match != nil {
		return strings.TrimSuffix(match[1], "."), nil
	}

	return "", fmt.Errorf("could not read a MySQL version out of %q", strings.TrimSpace(output))
}

// IsMysqlBackupVersionHigherThanRestoreVersion reports whether a backup
// produced on backupVersion would be downgrade-restoring onto restoreVersion.
func IsMysqlBackupVersionHigherThanRestoreVersion(
	backupVersion, restoreVersion MysqlVersion,
) (bool, error) {
	order, err := compareReleaseLines(string(backupVersion), string(restoreVersion))
	if err != nil {
		return false, fmt.Errorf("cannot order MySQL versions: %w", err)
	}

	return order > 0, nil
}

// EscapeMysqlPassword escapes special characters for the MySQL .my.cnf file
// format (passwords with special chars are double-quoted).
func EscapeMysqlPassword(password string) string {
	password = strings.ReplaceAll(password, "\\", "\\\\")
	password = strings.ReplaceAll(password, "\"", "\\\"")
	return password
}
