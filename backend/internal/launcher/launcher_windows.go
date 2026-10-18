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

	"databasus-backend/internal/launcher/metadatadb"
	"databasus-backend/internal/util/tools"
)

const defaultBackendStopTimeout = 30 * time.Second

type Config struct {
	DataFolder        string
	BackupsFolder     string
	MetadataDbPort    int
	BackendExecutable string
	// WebListenHost and WebPort are where the backend serves the Web UI and API; an empty host
	// means every interface.
	WebListenHost string
	WebPort       int
	// IsHTTPS serves the Web UI over HTTPS only, with the certificate stored in the data folder by
	// the tls-setup command. A listen host other than loopback requires it.
	IsHTTPS bool
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

	webTLSFiles, err := prepareWebTLS(config, logger)
	if err != nil {
		return err
	}

	if err := writeFrontendRuntimeConfigIfUiPresent(config.BackendExecutable, logger); err != nil {
		return err
	}

	metadataBinFolder, err := findMetadataBinFolder()
	if err != nil {
		return err
	}

	metadataDatabase, err := metadatadb.Start(ctx, metadatadb.Settings{
		DataFolder:   config.DataFolder,
		BinFolder:    metadataBinFolder,
		Port:         config.MetadataDbPort,
		Logger:       logger,
		ServerOutput: childOutput,
	})
	if err != nil {
		return fmt.Errorf("start the metadata database: %w", err)
	}

	backend, err := startBackend(ctx, backendStartSpec{
		executablePath: config.BackendExecutable,
		metadataDsn:    metadataDatabase.GetDsn(),
		dataFolder:     config.DataFolder,
		backupsFolder:  config.BackupsFolder,
		webListenHost:  config.WebListenHost,
		webPort:        config.WebPort,
		webTLSFiles:    webTLSFiles,
		appVersion:     config.AppVersion,
		output:         childOutput,
		stopTimeout:    defaultBackendStopTimeout,
		logger:         logger,
	})
	if err != nil {
		return errors.Join(err, metadataDatabase.Stop())
	}

	logger.Info("Databasus is running")

	select {
	case <-ctx.Done():
		logger.Info("stop requested")

		return errors.Join(backend.stop(), metadataDatabase.Stop())
	case <-backend.exited:
		return errors.Join(
			fmt.Errorf(
				"the backend exited: %w; its last output:\n%s",
				backend.exitErr,
				backend.outputTail.GetContents(),
			),
			metadataDatabase.Stop(),
		)
	case databaseErr := <-metadataDatabase.GetUnexpectedExitChannel():
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
