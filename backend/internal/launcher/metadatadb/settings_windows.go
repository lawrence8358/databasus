package metadatadb

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultPort = 5437

	defaultStartTimeout = 2 * time.Minute
	defaultStopTimeout  = time.Minute

	superuserName           = "postgres"
	applicationDatabaseName = "databasus"
	maintenanceDatabaseName = "postgres"
)

type Settings struct {
	// DataFolder holds the cluster, its password file and the server log.
	DataFolder string

	// BinFolder is the bundled server's bin folder; lib and share sit beside it.
	BinFolder string

	Port int

	// ServerOutput receives what the server prints before its log collector takes over, which is
	// where startup failures such as an occupied port are reported.
	ServerOutput io.Writer

	Logger *slog.Logger

	StartTimeout time.Duration
	StopTimeout  time.Duration
}

func (s Settings) withDefaults() (Settings, error) {
	if s.DataFolder == "" {
		return s, errors.New("metadata database data folder is not set")
	}

	if s.BinFolder == "" {
		return s, errors.New("metadata database bin folder is not set")
	}

	absoluteDataFolder, err := filepath.Abs(s.DataFolder)
	if err != nil {
		return s, err
	}

	s.DataFolder = absoluteDataFolder

	if s.Port == 0 {
		s.Port = DefaultPort
	}

	if s.ServerOutput == nil {
		s.ServerOutput = os.Stderr
	}

	if s.Logger == nil {
		s.Logger = slog.Default()
	}

	if s.StartTimeout == 0 {
		s.StartTimeout = defaultStartTimeout
	}

	if s.StopTimeout == 0 {
		s.StopTimeout = defaultStopTimeout
	}

	return s, nil
}

func (s Settings) getClusterFolder() string {
	return filepath.Join(s.DataFolder, "pgdata")
}

func (s Settings) getClusterInitializationFolder() string {
	return filepath.Join(s.DataFolder, "pgdata.initializing")
}

func (s Settings) getPasswordFilePath() string {
	return filepath.Join(s.DataFolder, "metadata-db.password")
}

func (s Settings) getLogFolder() string {
	return filepath.Join(s.DataFolder, "logs")
}

func (s Settings) getProgramPath(programName string) string {
	return filepath.Join(s.BinFolder, programName+".exe")
}
