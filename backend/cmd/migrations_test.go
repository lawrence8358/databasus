package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/config"
	"databasus-backend/internal/util/locations"
)

func Test_ApplyMigrations_WithoutGooseOnPath_MigratesEmptyDatabaseOnce(t *testing.T) {
	t.Setenv("PATH", "")

	migrationsFolder := locations.FindAppFolder("migrations")
	databaseDsn := createEmptyDatabase(t)

	migrationFiles, err := filepath.Glob(filepath.Join(migrationsFolder, "*.sql"))
	require.NoError(t, err)

	firstRunMigrations, err := applyMigrations(t.Context(), databaseDsn, migrationsFolder)
	require.NoError(t, err)
	assert.Len(t, firstRunMigrations, len(migrationFiles))

	secondRunMigrations, err := applyMigrations(t.Context(), databaseDsn, migrationsFolder)
	require.NoError(t, err)
	assert.Empty(t, secondRunMigrations)
}

func Test_ApplyMigrations_WithMissingFolder_ReturnsErrorNamingIt(t *testing.T) {
	missingFolder := filepath.Join(t.TempDir(), "migrations")

	_, err := applyMigrations(t.Context(), config.GetEnv().DatabaseDsn, missingFolder)

	require.Error(t, err)
	assert.Contains(t, err.Error(), missingFolder)
}

func createEmptyDatabase(t *testing.T) string {
	t.Helper()

	databaseName := "migrations_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	_, systemDsn, err := config.RewriteDbName(config.GetEnv().DatabaseDsn, "postgres")
	require.NoError(t, err)

	_, databaseDsn, err := config.RewriteDbName(config.GetEnv().DatabaseDsn, databaseName)
	require.NoError(t, err)

	systemDatabase, err := sql.Open("pgx", systemDsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = systemDatabase.Close() })

	_, err = systemDatabase.ExecContext(t.Context(), fmt.Sprintf("CREATE DATABASE %s", databaseName))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, dropErr := systemDatabase.ExecContext(
			context.Background(),
			fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", databaseName),
		)
		if dropErr != nil {
			fmt.Fprintf(os.Stderr, "failed to drop %s: %v\n", databaseName, dropErr)
		}
	})

	return databaseDsn
}
