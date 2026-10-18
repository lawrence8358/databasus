//go:build windows

// Command launcher runs Databasus on Windows without Docker: it starts the bundled metadata
// PostgreSQL and the backend, and stops them in that reverse order on Ctrl+C.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"databasus-backend/internal/launcher"
	"databasus-backend/internal/launcher/metadatadb"
	"databasus-backend/internal/util/locations"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	config := launcher.Config{Logger: logger}

	flag.StringVar(&config.DataFolder, "data-dir", getDefaultDataFolder(),
		"program data folder: metadata database, secret key, logs")
	flag.StringVar(&config.BackupsFolder, "backups-dir", os.Getenv(locations.BackupsFolderEnvVariable),
		`local folder for backups (default: <data-dir>\backups)`)
	flag.IntVar(&config.MetadataDbPort, "metadata-db-port", metadatadb.DefaultPort,
		"port of the bundled metadata PostgreSQL on 127.0.0.1")
	flag.StringVar(&config.BackendExecutable, "backend", getDefaultBackendExecutable(),
		"path to databasus.exe")
	flag.Parse()

	if err := launcher.BindChildProcessesToLauncherLifetime(); err != nil {
		logger.Error("could not tie child processes to the launcher", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := launcher.Run(ctx, config); err != nil {
		logger.Error("Databasus stopped", "error", err)
		os.Exit(1)
	}

	logger.Info("Databasus stopped")
}

func getDefaultDataFolder() string {
	if dataFolder := os.Getenv(locations.DataFolderEnvVariable); dataFolder != "" {
		return dataFolder
	}

	return filepath.Join(os.Getenv("ProgramData"), "Databasus")
}

func getDefaultBackendExecutable() string {
	executablePath, err := os.Executable()
	if err != nil {
		return "databasus.exe"
	}

	return filepath.Join(filepath.Dir(executablePath), "databasus.exe")
}
