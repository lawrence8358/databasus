package externaldb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"databasus-backend/internal/util/securefile"
)

const (
	// FileName is where the connection string is stored, inside the program data folder.
	FileName = "external-metadata-db.dsn"

	connectTimeout = 15 * time.Second
)

func GetFilePath(dataFolder string) string {
	return filepath.Join(dataFolder, FileName)
}

// Save stores the connection string so that only the service account and administrators can read it.
func Save(dataFolder, dsn string) error {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return errors.New("the connection string is empty")
	}

	if err := os.MkdirAll(dataFolder, 0o755); err != nil {
		return fmt.Errorf("create the data folder: %w", err)
	}

	return securefile.WriteFile(GetFilePath(dataFolder), []byte(dsn), 0o600)
}

// Load reads a stored connection string.
func Load(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read the external metadata database connection string: %w", err)
	}

	dsn := strings.TrimSpace(strings.TrimPrefix(string(content), "\xef\xbb\xbf"))
	if dsn == "" {
		return "", fmt.Errorf("the external metadata database connection string in %s is empty", path)
	}

	return dsn, nil
}

// TestConnection checks that the server accepts the credentials and can take writes. Errors never
// contain the password.
func TestConnection(ctx context.Context, dsn string) error {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("the connection string is not valid: %s", scrub(err.Error(), dsn, ""))
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("cannot connect to the external metadata database: %s",
			scrub(err.Error(), dsn, config.Password))
	}
	defer func() { _ = connection.Close(context.WithoutCancel(ctx)) }()

	var isStandby bool
	if err := connection.QueryRow(ctx, "SELECT pg_is_in_recovery()").Scan(&isStandby); err != nil {
		return fmt.Errorf("cannot query the external metadata database: %s",
			scrub(err.Error(), dsn, config.Password))
	}

	if isStandby {
		return errors.New("the external metadata database is a read-only standby; point to the primary server")
	}

	return nil
}

func scrub(message, dsn, password string) string {
	message = strings.ReplaceAll(message, dsn, "<connection string>")

	if password != "" {
		message = strings.ReplaceAll(message, password, "***")
	}

	return message
}
