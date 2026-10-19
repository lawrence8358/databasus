package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// clientVersionTimeout bounds one --version call. A client that has not
// answered by then is as unusable as one that is missing.
const clientVersionTimeout = 10 * time.Second

// binDirVersionParser reads the release a client reports out of its
// --version output. The shape differs per engine, so each engine brings its
// own parser rather than the check guessing at the first number it finds.
type binDirVersionParser func(output string) (string, error)

// executionVerdicts caches, per bin directory, the outcome of starting its
// clients. The container probes the health endpoint every 30 seconds, and
// starting every shipped binary on each probe is how a loaded host reports a
// false unhealthy, so the binaries run once per process and later checks
// reuse the verdict.
var executionVerdicts sync.Map

// ToolCheckResult is one DB-version bundle's verification outcome. A bundle
// is considered fatal when missing/broken should crash the app at startup
// (Postgres) and non-fatal otherwise (MySQL, MariaDB, MongoDB).
type ToolCheckResult struct {
	Db      DbEngine
	Version string
	BinDir  string
	Errors  []error
	IsFatal bool
}

// checkBinDir verifies binDir exists and that every command in
// requiredCommands is present and, where the platform has one, has the
// executable bit set. Returns the collected errors. If the directory does not
// exist, returns a single not-found error so callers can treat it differently
// for fatal vs non-fatal bundles.
func checkBinDir(binDir string, requiredCommands []string) []error {
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return []error{fmt.Errorf("client tools bin directory not found: %s", binDir)}
	}

	var errs []error
	for _, cmd := range requiredCommands {
		cmdPath := getExecutablePath(binDir, cmd)

		info, err := os.Stat(cmdPath)
		if os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("client command not found: %s", cmdPath))
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("cannot stat %s: %w", cmdPath, err))
			continue
		}

		if isExecutableBitRequired && info.Mode().Perm()&0o111 == 0 {
			errs = append(errs, fmt.Errorf("client command not executable: %s", cmdPath))
		}
	}

	return errs
}

// runBinDirChecks verifies the bundle's files are in place and, once per
// process, that they start and report a version on the release line the
// bundle is registered under. Pass an empty releaseLine for a bundle whose
// directory is not named after a version.
func runBinDirChecks(
	binDir string,
	requiredCommands []string,
	releaseLine string,
	parseVersion binDirVersionParser,
) []error {
	if errs := checkBinDir(binDir, requiredCommands); len(errs) > 0 {
		return errs
	}

	if cached, ok := executionVerdicts.Load(binDir); ok {
		if verdict, isVerdict := cached.([]error); isVerdict {
			return verdict
		}
	}

	errs := runClients(binDir, requiredCommands, releaseLine, parseVersion)
	executionVerdicts.Store(binDir, errs)

	return errs
}

// runClients starts every required command with --version and checks what it
// answers. Presence and the executable bit say nothing about a missing shared
// library, which is the failure a new bundle is most likely to introduce.
func runClients(
	binDir string,
	requiredCommands []string,
	releaseLine string,
	parseVersion binDirVersionParser,
) []error {
	var errs []error

	for _, cmd := range requiredCommands {
		cmdPath := getExecutablePath(binDir, cmd)

		ctx, cancel := context.WithTimeout(context.Background(), clientVersionTimeout)
		output, err := exec.CommandContext(ctx, cmdPath, "--version").CombinedOutput()
		cancel()

		if err != nil {
			errs = append(errs, fmt.Errorf(
				"client command cannot run: %s: %w: %s",
				cmdPath, err, strings.TrimSpace(string(output)),
			))

			continue
		}

		if releaseLine == "" {
			continue
		}

		reported, err := parseVersion(string(output))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", cmdPath, err))

			continue
		}

		if !isVersionOnReleaseLine(reported, releaseLine) {
			errs = append(errs, fmt.Errorf(
				"client command %s reports version %s, which is not on the %s release line its bundle is registered under",
				cmdPath,
				reported,
				releaseLine,
			))
		}
	}

	return errs
}

// getExecutablePath names a client inside its bin directory. Commands are
// registered by their POSIX names; Windows needs the .exe suffix on the file.
func getExecutablePath(binDir, command string) string {
	return filepath.Join(binDir, command+executableSuffix)
}

type DbEngine string

const (
	DbEnginePostgresql DbEngine = "postgresql"
	DbEngineMysql      DbEngine = "mysql"
	DbEngineMariadb    DbEngine = "mariadb"
	DbEngineMongodb    DbEngine = "mongodb"
)

var clientToolsChecksByEngine = map[DbEngine]func() []ToolCheckResult{
	DbEnginePostgresql: checkPostgresql,
	DbEngineMysql:      checkMysql,
	DbEngineMariadb:    checkMariadb,
	DbEngineMongodb:    checkMongodb,
}

// CheckAllClientTools verifies every expected client binary for the current
// arch is present, executable, starts, and reports a version on the release
// line its bundle is registered under. Pure: never logs, never exits.
func CheckAllClientTools() []ToolCheckResult {
	results := checkPostgresql()
	results = append(results, checkMysql()...)
	results = append(results, checkMariadb()...)
	results = append(results, checkMongodb()...)

	return results
}

// CheckEngineClientTools runs the same checks as CheckAllClientTools for one
// engine's bundles only, so an engine the platform does not support is never
// probed.
func CheckEngineClientTools(engine DbEngine) []ToolCheckResult {
	check, isKnown := clientToolsChecksByEngine[engine]
	if !isKnown {
		panic(fmt.Sprintf("no client tools check registered for engine %q", engine))
	}

	return check()
}

// ClientToolsHealthError returns an aggregated error if any fatal-tier bundle
// has check failures, else nil. Used by the healthcheck endpoint, which
// re-reads the bundle files on every probe but reuses the startup verdict on
// whether the binaries start.
func ClientToolsHealthError() error {
	results := CheckAllClientTools()

	var msgs []string
	for _, r := range results {
		if !r.IsFatal {
			continue
		}

		for _, err := range r.Errors {
			msgs = append(msgs, fmt.Sprintf("%s %s: %s", r.Db, r.Version, err.Error()))
		}
	}

	if len(msgs) == 0 {
		return nil
	}

	return errors.New("client tools broken: " + strings.Join(msgs, "; "))
}
