package metadatadb

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"databasus-backend/internal/util/securefile"
)

// Only the IPv4 loopback is listed; a connection that matches no line is rejected, which covers
// every other address, replication connections and local sockets.
const clientAuthenticationConfig = `# Written by the Databasus launcher. Only password logins from this machine are accepted.
host all all 127.0.0.1/32 scram-sha-256
`

var serverVersionPattern = regexp.MustCompile(`\(PostgreSQL\)\s+(\d+)`)

func ensureCluster(ctx context.Context, settings Settings) (string, error) {
	clusterFolder := settings.getClusterFolder()

	if _, err := os.Stat(filepath.Join(clusterFolder, "PG_VERSION")); err == nil {
		if err := ensureClusterMatchesServerMajor(ctx, settings); err != nil {
			return "", err
		}

		return readPassword(settings)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if _, err := os.Stat(clusterFolder); err == nil {
		return "", fmt.Errorf(
			"%s exists but holds no PostgreSQL cluster; move it away to let a new one be created",
			clusterFolder,
		)
	}

	return createCluster(ctx, settings)
}

// The cluster is built in a side folder and renamed into place only once initdb has finished, so
// an interrupted first start leaves nothing that a later start would mistake for a usable cluster.
func createCluster(ctx context.Context, settings Settings) (string, error) {
	settings.Logger.Info("creating the metadata database cluster", "folder", settings.getClusterFolder())

	initializationFolder := settings.getClusterInitializationFolder()
	if err := os.RemoveAll(initializationFolder); err != nil {
		return "", fmt.Errorf("remove leftover %s: %w", initializationFolder, err)
	}

	if err := os.MkdirAll(initializationFolder, 0o700); err != nil {
		return "", err
	}

	if err := securefile.RestrictDir(initializationFolder, 0o700); err != nil {
		return "", fmt.Errorf("restrict access to %s: %w", initializationFolder, err)
	}

	password, err := generatePassword()
	if err != nil {
		return "", err
	}

	if err := writeMetadataPasswordFile(settings.getPasswordFilePath(), password); err != nil {
		return "", err
	}

	initdbCommand, err := newRestrictedCommand(
		ctx,
		settings.getProgramPath("initdb"),
		"--pgdata="+initializationFolder,
		"--username="+superuserName,
		"--pwfile="+settings.getPasswordFilePath(),
		"--auth=scram-sha-256",
		"--encoding=UTF8",
		"--locale-provider=builtin",
		"--builtin-locale=C.UTF-8",
		"--locale=C",
	)
	if err != nil {
		return "", err
	}

	var initdbOutput bytes.Buffer
	initdbCommand.Stdout = &initdbOutput
	initdbCommand.Stderr = &initdbOutput

	if err := runRestricted(initdbCommand); err != nil {
		return "", fmt.Errorf("initdb failed: %w\n%s", err, strings.TrimSpace(initdbOutput.String()))
	}

	hbaPath := filepath.Join(initializationFolder, "pg_hba.conf")
	if err := os.WriteFile(hbaPath, []byte(clientAuthenticationConfig), 0o600); err != nil {
		return "", err
	}

	if err := os.Rename(initializationFolder, settings.getClusterFolder()); err != nil {
		return "", fmt.Errorf("move the new cluster into place: %w", err)
	}

	return password, nil
}

// A new Databasus release may only swap the server binaries within one major. Starting a cluster
// of another major fails inside PostgreSQL with a terse message, so the mismatch is named here.
func ensureClusterMatchesServerMajor(ctx context.Context, settings Settings) error {
	clusterVersion, err := os.ReadFile(filepath.Join(settings.getClusterFolder(), "PG_VERSION"))
	if err != nil {
		return err
	}

	serverMajor, err := getServerMajor(ctx, settings)
	if err != nil {
		return err
	}

	clusterMajor := strings.TrimSpace(string(clusterVersion))
	if clusterMajor != serverMajor {
		return fmt.Errorf(
			"the metadata database in %s was created by PostgreSQL %s, but the bundled server is PostgreSQL %s; "+
				"a major version change needs pg_upgrade",
			settings.getClusterFolder(), clusterMajor, serverMajor,
		)
	}

	return nil
}

func getServerMajor(ctx context.Context, settings Settings) (string, error) {
	versionCommand, err := newRestrictedCommand(ctx, settings.getProgramPath("postgres"), "--version")
	if err != nil {
		return "", err
	}

	var versionOutput bytes.Buffer
	versionCommand.Stdout = &versionOutput
	versionCommand.Stderr = &versionOutput

	if err := runRestricted(versionCommand); err != nil {
		return "", fmt.Errorf("run postgres --version: %w\n%s", err, versionOutput.String())
	}

	match := serverVersionPattern.FindStringSubmatch(versionOutput.String())
	if match == nil {
		return "", fmt.Errorf("unrecognized postgres --version output: %q", versionOutput.String())
	}

	return match[1], nil
}

func generatePassword() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate the metadata database password: %w", err)
	}

	return hex.EncodeToString(randomBytes), nil
}

func writeMetadataPasswordFile(path, password string) error {
	if err := securefile.WriteFile(path, []byte(password+"\n"), 0o600); err != nil {
		return fmt.Errorf("write the metadata database password file: %w", err)
	}

	return nil
}

func readPassword(settings Settings) (string, error) {
	passwordFile, err := os.ReadFile(settings.getPasswordFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf(
				"the metadata database exists but its password file %s is missing",
				settings.getPasswordFilePath(),
			)
		}

		return "", err
	}

	password := strings.TrimSpace(string(passwordFile))
	if password == "" {
		return "", fmt.Errorf("the metadata database password file %s is empty", settings.getPasswordFilePath())
	}

	return password, nil
}
