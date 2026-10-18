package launcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows"

	"databasus-backend/internal/util/locations"
)

const (
	databaseDsnEnvVariable = "DATABASE_DSN"
	envModeEnvVariable     = "ENV_MODE"
)

type backendProcess struct {
	command     *exec.Cmd
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
	stopTimeout    time.Duration
	logger         *slog.Logger
}

// The backend runs in its own process group so a Ctrl+C typed in the console reaches only the
// launcher, which then stops the backend before the database instead of both at once.
func startBackend(ctx context.Context, spec backendStartSpec) (*backendProcess, error) {
	command := exec.CommandContext(context.WithoutCancel(ctx), spec.executablePath)
	command.Dir = filepath.Dir(spec.executablePath)
	command.Env = buildBackendEnvironment(spec)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_PROCESS_GROUP}

	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("start the backend %s: %w", spec.executablePath, err)
	}

	backend := &backendProcess{
		command:     command,
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

func buildBackendEnvironment(spec backendStartSpec) []string {
	environment := os.Environ()

	if _, isSet := os.LookupEnv(envModeEnvVariable); !isSet {
		environment = append(environment, envModeEnvVariable+"=production")
	}

	environment = append(environment,
		databaseDsnEnvVariable+"="+spec.metadataDsn,
		locations.DataFolderEnvVariable+"="+spec.dataFolder,
	)

	if spec.backupsFolder != "" {
		environment = append(environment, locations.BackupsFolderEnvVariable+"="+spec.backupsFolder)
	}

	return environment
}

// The backend shuts down gracefully on os.Interrupt, which Go raises for CTRL_BREAK_EVENT. That
// event only reaches processes sharing the launcher's console; without one it cannot be sent, and
// the backend is killed instead.
func (b *backendProcess) stop() error {
	select {
	case <-b.exited:
		return nil
	default:
	}

	b.logger.Info("stopping the backend")

	err := windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(b.command.Process.Pid))
	if err != nil {
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
