package launcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"databasus-backend/internal/util/tools"
)

const defaultBackendStopTimeout = 30 * time.Second

type Config struct {
	DataFolder     string
	BackupsFolder  string
	MetadataDbPort int
	// ExternalMetadataDsnFile, when set, names a file holding the connection string of an external
	// metadata database; the bundled PostgreSQL is then neither initialised nor started.
	ExternalMetadataDsnFile string
	BackendExecutable       string
	// WebListenHost and WebPort are where the backend serves the Web UI and API; an empty host
	// means every interface.
	WebListenHost string
	WebPort       int
	// AppVersion is what the backend reports as its version, as APP_VERSION does in the Docker image.
	AppVersion string
	// ChildOutput receives the backend's and the metadata server's console output; os.Stderr when nil.
	ChildOutput io.Writer
	Logger      *slog.Logger
}

// Run returns after both the metadata database and the backend have stopped: when ctx is
// cancelled, a requested stop that returns nil, or when either process exits on its own, which
// returns the reason.
func Run(ctx context.Context, config Config) error {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	childOutput := config.ChildOutput
	if childOutput == nil {
		childOutput = os.Stderr
	}

	if config.DataFolder == "" {
		return errors.New("the data folder is not set")
	}

	if _, err := os.Stat(config.BackendExecutable); err != nil {
		return fmt.Errorf("backend executable: %w", err)
	}

	if err := ensureWebPortIsFree(ctx, config.WebListenHost, config.WebPort); err != nil {
		return err
	}

	if err := writeFrontendRuntimeConfigIfUiPresent(config.BackendExecutable, logger); err != nil {
		return err
	}

	metadataDatabase, err := startMetadataDatabase(ctx, config, logger, childOutput)
	if err != nil {
		return err
	}

	backend, err := startBackend(ctx, backendStartSpec{
		executablePath: config.BackendExecutable,
		metadataDsn:    metadataDatabase.getDsn(),
		dataFolder:     config.DataFolder,
		backupsFolder:  config.BackupsFolder,
		webListenHost:  config.WebListenHost,
		webPort:        config.WebPort,
		appVersion:     config.AppVersion,
		output:         childOutput,
		stopTimeout:    defaultBackendStopTimeout,
		logger:         logger,
	})
	if err != nil {
		return errors.Join(err, metadataDatabase.stop())
	}

	logger.Info("Databasus is running")

	select {
	case <-ctx.Done():
		logger.Info("stop requested")

		return errors.Join(backend.stop(), metadataDatabase.stop())
	case <-backend.exited:
		return errors.Join(
			fmt.Errorf(
				"the backend exited: %w; its last output:\n%s",
				backend.exitErr,
				backend.outputTail.GetContents(),
			),
			metadataDatabase.stop(),
		)
	case databaseErr := <-metadataDatabase.exited:
		return errors.Join(databaseErr, backend.stop())
	}
}

// The backend only logs a failed listen and keeps running without its Web UI, so the launcher
// refuses to start instead, which surfaces the conflict in the service log and event log.
func ensureWebPortIsFree(ctx context.Context, host string, port int) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("the Web UI port %d is not available: %w", port, err)
	}

	return listener.Close()
}

func writeFrontendRuntimeConfigIfUiPresent(backendExecutable string, logger *slog.Logger) error {
	uiBuildFolder := filepath.Join(filepath.Dir(backendExecutable), "ui", "build")

	if info, err := os.Stat(uiBuildFolder); err != nil || !info.IsDir() {
		logger.Warn("no built Web UI beside the backend; serving the API only", "expected", uiBuildFolder)

		return nil
	}

	return writeFrontendRuntimeConfig(uiBuildFolder)
}

// The tools package panics when the bundle is missing, because the backend cannot run without its
// clients; the launcher reports the same condition as an ordinary startup error.
func findMetadataBinFolder() (binFolder string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("locate the bundled PostgreSQL server: %v", recovered)
		}
	}()

	return tools.GetMetadataPostgresqlBinDir(), nil
}
