package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"databasus-backend/internal/config"
	"databasus-backend/internal/util/locations"
	"databasus-backend/internal/util/logger"
)

func runMigrations(log *slog.Logger) {
	migrationsFolder := locations.FindAppFolder("migrations")

	log.Info("running database migrations", "folder", migrationsFolder)

	appliedMigrations, err := applyMigrations(
		context.Background(),
		config.GetEnv().DatabaseDsn,
		migrationsFolder,
	)
	if err != nil {
		log.Error("failed to run migrations", "error", err)
		logger.ExitAfterFlush(1)
	}

	log.Info("database migrations completed successfully", "applied", appliedMigrations)
}

// The migrations run in-process rather than through the goose CLI, so a host only needs the
// binary and its migrations folder, not a goose executable on PATH.
func applyMigrations(ctx context.Context, databaseDsn, migrationsFolder string) ([]string, error) {
	if _, err := os.Stat(migrationsFolder); err != nil {
		return nil, fmt.Errorf("migrations folder %s is not readable: %w", migrationsFolder, err)
	}

	database, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open the metadata database: %w", err)
	}
	defer func() { _ = database.Close() }()

	migrationProvider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrationsFolder))
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations from %s: %w", migrationsFolder, err)
	}

	migrationResults, err := migrationProvider.Up(ctx)
	if err != nil {
		return nil, err
	}

	appliedMigrations := make([]string, 0, len(migrationResults))
	for _, migrationResult := range migrationResults {
		appliedMigrations = append(appliedMigrations, migrationResult.Source.Path)
	}

	return appliedMigrations, nil
}
