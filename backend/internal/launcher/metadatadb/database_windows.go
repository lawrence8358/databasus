package metadatadb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var ErrServerExited = errors.New("the metadata database server exited")

type Database struct {
	settings Settings
	password string

	serverCommand *exec.Cmd
	serverExited  chan struct{}
	serverExitErr error

	unexpectedExits chan error
	isStopping      atomic.Bool
	stopOnce        sync.Once
	stopErr         error
}

// ctx bounds startup only; the server keeps running until Stop.
func Start(ctx context.Context, settings Settings) (*Database, error) {
	settings, err := settings.withDefaults()
	if err != nil {
		return nil, err
	}

	for _, folder := range []string{settings.DataFolder, settings.getLogFolder()} {
		if err := os.MkdirAll(folder, 0o700); err != nil {
			return nil, fmt.Errorf("create %s: %w", folder, err)
		}
	}

	password, err := ensureCluster(ctx, settings)
	if err != nil {
		return nil, err
	}

	if err := stopOrphanedServer(ctx, settings); err != nil {
		return nil, err
	}

	if err := ensurePortIsFree(settings.Port); err != nil {
		return nil, err
	}

	database := &Database{
		settings:        settings,
		password:        password,
		unexpectedExits: make(chan error, 1),
	}

	if err := database.startServerWithRecovery(ctx); err != nil {
		return nil, err
	}

	if err := ensureApplicationDatabase(ctx, database.getConnectionString(maintenanceDatabaseName)); err != nil {
		_ = database.Stop()

		return nil, err
	}

	go database.reportUnexpectedExit()

	settings.Logger.Info("metadata database is ready", "port", settings.Port)

	return database, nil
}

func (d *Database) GetDsn() string {
	return d.getConnectionString(applicationDatabaseName)
}

// GetUnexpectedExitChannel delivers one error if the server stops without Stop being called, so
// the launcher can bring the backend down instead of leaving it without its database.
func (d *Database) GetUnexpectedExitChannel() <-chan error {
	return d.unexpectedExits
}

// Stop asks for a fast shutdown, which ends sessions and checkpoints, and kills the server only if
// it has not exited within the stop timeout. A killed server is repaired by the next Start.
func (d *Database) Stop() error {
	d.stopOnce.Do(func() {
		d.isStopping.Store(true)
		d.stopErr = d.stopServer()
	})

	return d.stopErr
}

func (d *Database) stopServer() error {
	select {
	case <-d.serverExited:
		return nil
	default:
	}

	d.settings.Logger.Info("stopping the metadata database")

	stopErr := runPgCtlStop(d.settings)

	select {
	case <-d.serverExited:
		if stopErr != nil {
			d.settings.Logger.Warn("pg_ctl stop reported an error, but the server exited", "error", stopErr)
		}

		return nil
	case <-time.After(d.settings.StopTimeout):
	}

	d.settings.Logger.Error("the metadata database did not stop in time; killing it", "pg_ctl_error", stopErr)

	if err := d.serverCommand.Process.Kill(); err != nil {
		return fmt.Errorf("kill the metadata database server: %w", err)
	}

	<-d.serverExited

	return errors.New("the metadata database was killed after it did not stop in time")
}

func (d *Database) reportUnexpectedExit() {
	<-d.serverExited

	if d.isStopping.Load() {
		return
	}

	d.unexpectedExits <- fmt.Errorf("%w: %w", ErrServerExited, d.serverExitErr)
}

// A server that crashed normally recovers on its own by replaying WAL. When it cannot start at
// all, the WAL is reset as the Docker image does; this may lose the last transactions, which is
// preferred to a service that never comes back.
func (d *Database) startServerWithRecovery(ctx context.Context) error {
	err := d.startServer(ctx)
	if err == nil {
		return nil
	}

	d.settings.Logger.Error("the metadata database failed to start; resetting its WAL", "error", err)

	if resetErr := runPgResetwal(ctx, d.settings); resetErr != nil {
		return fmt.Errorf("%w; pg_resetwal also failed: %w", err, resetErr)
	}

	if retryErr := d.startServer(ctx); retryErr != nil {
		return fmt.Errorf("the metadata database did not start even after a WAL reset: %w", retryErr)
	}

	d.settings.Logger.Warn("the metadata database started after a WAL reset")

	return nil
}

func (d *Database) startServer(ctx context.Context) error {
	serverCommand, err := newRestrictedCommand(
		context.WithoutCancel(ctx),
		d.settings.getProgramPath("postgres"),
		getServerArguments(d.settings)...,
	)
	if err != nil {
		return err
	}

	serverCommand.Stdout = d.settings.ServerOutput
	serverCommand.Stderr = d.settings.ServerOutput

	if err := startRestricted(serverCommand); err != nil {
		return fmt.Errorf("start postgres: %w", err)
	}

	d.serverCommand = serverCommand
	d.serverExited = make(chan struct{})

	go func(exited chan struct{}) {
		d.serverExitErr = serverCommand.Wait()
		close(exited)
	}(d.serverExited)

	if err := d.waitUntilReady(ctx); err != nil {
		_ = serverCommand.Process.Kill()
		<-d.serverExited

		return err
	}

	return nil
}

func (d *Database) waitUntilReady(ctx context.Context) error {
	readinessCtx, cancel := context.WithTimeout(ctx, d.settings.StartTimeout)
	defer cancel()

	connectionString := d.getConnectionString(maintenanceDatabaseName)

	var lastErr error

	for {
		select {
		case <-d.serverExited:
			return fmt.Errorf("%w during startup (%w); see the output above and %s",
				ErrServerExited, d.serverExitErr, d.settings.getLogFolder())
		default:
		}

		lastErr = pingDatabase(readinessCtx, connectionString)
		if lastErr == nil {
			return nil
		}

		select {
		case <-readinessCtx.Done():
			return fmt.Errorf("the metadata database did not accept connections within %s: %w",
				d.settings.StartTimeout, lastErr)
		case <-d.serverExited:
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (d *Database) getConnectionString(databaseName string) string {
	return fmt.Sprintf(
		"host=127.0.0.1 port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.settings.Port,
		superuserName,
		quoteConnectionValue(d.password),
		databaseName,
	)
}

func getServerArguments(settings Settings) []string {
	logFolder := strings.ReplaceAll(settings.getLogFolder(), `\`, "/")

	return []string{
		"-D", settings.getClusterFolder(),
		"-c", "listen_addresses=127.0.0.1",
		"-c", fmt.Sprintf("port=%d", settings.Port),
		"-c", "unix_socket_directories=",
		"-c", "password_encryption=scram-sha-256",
		"-c", "logging_collector=on",
		"-c", "log_directory=" + logFolder,
		"-c", "log_filename=postgresql-%a.log",
		"-c", "log_truncate_on_rotation=on",
		"-c", "log_rotation_age=1d",
		"-c", "log_rotation_size=0",
	}
}

func quoteConnectionValue(value string) string {
	escapedValue := strings.ReplaceAll(value, `\`, `\\`)
	escapedValue = strings.ReplaceAll(escapedValue, `'`, `\'`)

	return "'" + escapedValue + "'"
}
