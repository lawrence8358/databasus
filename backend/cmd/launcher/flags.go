//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"databasus-backend/internal/launcher"
	"databasus-backend/internal/launcher/metadatadb"
	"databasus-backend/internal/util/locations"
)

const (
	defaultWebListenHost = "127.0.0.1"
	defaultWebPort       = 4005
)

func parseRunFlags(arguments []string) (launcher.Config, error) {
	config := launcher.Config{AppVersion: Version}

	flags := flag.NewFlagSet("databasus-launcher", flag.ContinueOnError)
	flags.Usage = func() {
		_, _ = fmt.Fprint(flags.Output(), usage)
		flags.PrintDefaults()
	}

	flags.StringVar(&config.DataFolder, "data-dir", getDefaultDataFolder(),
		"program data folder: metadata database, secret key, logs")
	flags.StringVar(&config.BackupsFolder, "backups-dir", os.Getenv(locations.BackupsFolderEnvVariable),
		`local folder for backups (default: <data-dir>\backups)`)
	flags.IntVar(&config.MetadataDbPort, "metadata-db-port", metadatadb.DefaultPort,
		"port of the bundled metadata PostgreSQL on 127.0.0.1")
	flags.StringVar(&config.BackendExecutable, "backend", getDefaultBackendExecutable(),
		"path to databasus.exe")
	flags.StringVar(&config.WebListenHost, "listen-host", defaultWebListenHost,
		`address the Web UI listens on; "" for every interface`)
	flags.IntVar(&config.WebPort, "port", defaultWebPort, "port of the Web UI")

	if err := flags.Parse(arguments); err != nil {
		return launcher.Config{}, err
	}

	if flags.NArg() > 0 {
		return launcher.Config{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}

	if config.WebPort < 1 || config.WebPort > 65535 {
		return launcher.Config{}, fmt.Errorf("-port must be from 1 to 65535, got %d", config.WebPort)
	}

	return config, nil
}

// A service starts in System32 with its own environment, so every path it gets is made absolute
// at install time, and the environment-derived defaults are frozen into its command line.
func buildServiceArguments(config launcher.Config) ([]string, error) {
	dataFolder, err := filepath.Abs(config.DataFolder)
	if err != nil {
		return nil, fmt.Errorf("data folder: %w", err)
	}

	backendExecutable, err := filepath.Abs(config.BackendExecutable)
	if err != nil {
		return nil, fmt.Errorf("backend executable: %w", err)
	}

	arguments := []string{
		"-data-dir", dataFolder,
		"-metadata-db-port", strconv.Itoa(config.MetadataDbPort),
		"-backend", backendExecutable,
		"-listen-host", config.WebListenHost,
		"-port", strconv.Itoa(config.WebPort),
	}

	if config.BackupsFolder != "" {
		backupsFolder, err := filepath.Abs(config.BackupsFolder)
		if err != nil {
			return nil, fmt.Errorf("backups folder: %w", err)
		}

		arguments = append(arguments, "-backups-dir", backupsFolder)
	}

	return arguments, nil
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
