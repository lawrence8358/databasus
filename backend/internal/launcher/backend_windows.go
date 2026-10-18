package launcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sys/windows"

	"databasus-backend/internal/util/locations"
	"databasus-backend/internal/util/stopevent"
)

const (
	databaseDsnEnvVariable   = "DATABASE_DSN"
	envModeEnvVariable       = "ENV_MODE"
	webListenHostEnvVariable = "DATABASUS_LISTEN_HOST"
	webPortEnvVariable       = "DATABASUS_PORT"
	appVersionEnvVariable    = "APP_VERSION"

	backendOutputTailBytes = 2048
)

type backendProcess struct {
	command     *exec.Cmd
	stopEvent   *stopevent.Event
	outputTail  *outputTail
	exited      chan struct{}
	exitErr     error
	logger      *slog.Logger
	stopTimeout time.Duration
}

type backendStartSpec struct {
	executablePath string
	metadataDsn    string
	dataFolder     string
	backupsFolder  string
	webListenHost  string
	webPort        int
	appVersion     string
	output         io.Writer
	stopTimeout    time.Duration
	logger         *slog.Logger
}

func startBackend(ctx context.Context, spec backendStartSpec) (*backendProcess, error) {
	stopEvent, err := stopevent.New()
	if err != nil {
		return nil, err
	}

	command := exec.CommandContext(context.WithoutCancel(ctx), spec.executablePath)
	command.Dir = filepath.Dir(spec.executablePath)
	command.Env = append(buildBackendEnvironment(spec), stopEvent.GetEnvironmentEntry())
	outputTail := newOutputTail(backendOutputTailBytes)
	output := io.MultiWriter(spec.output, outputTail)

	command.Stdout = output
	command.Stderr = output
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags:              getBackendCreationFlags(),
		AdditionalInheritedHandles: []syscall.Handle{stopEvent.GetHandle()},
	}

	if err := command.Start(); err != nil {
		_ = stopEvent.Close()

		return nil, fmt.Errorf("start the backend %s: %w", spec.executablePath, err)
	}

	backend := &backendProcess{
		command:     command,
		stopEvent:   stopEvent,
		outputTail:  outputTail,
		exited:      make(chan struct{}),
		logger:      spec.logger,
		stopTimeout: spec.stopTimeout,
	}

	go func() {
		backend.exitErr = command.Wait()
		close(backend.exited)
	}()

	return backend, nil
}

// With a console, the backend gets its own process group so a Ctrl+C typed there reaches only
// the launcher, which then stops the backend before the database. Without one (a service), the
// backend must not get a console of its own: a console in the services session receives
// CTRL_LOGOFF_EVENT whenever any user signs out, and Go turns that into SIGTERM.
func getBackendCreationFlags() uint32 {
	if _, err := windows.GetConsoleCP(); err != nil {
		return windows.DETACHED_PROCESS
	}

	return windows.CREATE_NEW_PROCESS_GROUP
}

func buildBackendEnvironment(spec backendStartSpec) []string {
	environment := os.Environ()

	if _, isSet := os.LookupEnv(envModeEnvVariable); !isSet {
		environment = append(environment, envModeEnvVariable+"=production")
	}

	if _, isSet := os.LookupEnv(appVersionEnvVariable); !isSet && spec.appVersion != "" {
		environment = append(environment, appVersionEnvVariable+"="+spec.appVersion)
	}

	environment = append(environment,
		databaseDsnEnvVariable+"="+spec.metadataDsn,
		locations.DataFolderEnvVariable+"="+spec.dataFolder,
		webListenHostEnvVariable+"="+spec.webListenHost,
		webPortEnvVariable+"="+strconv.Itoa(spec.webPort),
	)

	if spec.backupsFolder != "" {
		environment = append(environment, locations.BackupsFolderEnvVariable+"="+spec.backupsFolder)
	}

	return environment
}

func (b *backendProcess) stop() error {
	defer func() { _ = b.stopEvent.Close() }()

	select {
	case <-b.exited:
		return nil
	default:
	}

	b.logger.Info("stopping the backend")

	if err := b.stopEvent.Set(); err != nil {
		b.logger.Warn("could not ask the backend to stop gracefully", "error", err)
	} else {
		select {
		case <-b.exited:
			return nil
		case <-time.After(b.stopTimeout):
			b.logger.Error("the backend did not stop in time; killing it")
		}
	}

	if err := b.command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("kill the backend: %w", err)
	}

	<-b.exited

	return nil
}
