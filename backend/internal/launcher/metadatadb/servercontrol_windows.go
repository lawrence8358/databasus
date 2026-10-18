package metadatadb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const pingTimeout = 3 * time.Second

// A launcher that was killed can leave its server running, since nothing told it to stop. That
// server holds the port and the data directory, so it is shut down before a fresh one starts.
func stopOrphanedServer(ctx context.Context, settings Settings) error {
	statusCommand := exec.CommandContext(
		ctx,
		settings.getProgramPath("pg_ctl"),
		"status",
		"--pgdata="+settings.getClusterFolder(),
	)

	err := statusCommand.Run()
	if err == nil {
		settings.Logger.Warn("a metadata database server from an earlier run is still running; stopping it")

		return runPgCtlStop(settings)
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil
	}

	return fmt.Errorf("run pg_ctl status: %w", err)
}

func runPgCtlStop(settings Settings) error {
	stopCommand := exec.Command(
		settings.getProgramPath("pg_ctl"),
		"stop",
		"--pgdata="+settings.getClusterFolder(),
		"--mode=fast",
		"--wait",
		"--timeout="+strconv.Itoa(int(settings.StopTimeout.Seconds())),
	)

	var stopOutput bytes.Buffer
	stopCommand.Stdout = &stopOutput
	stopCommand.Stderr = &stopOutput

	if err := stopCommand.Run(); err != nil {
		return fmt.Errorf("pg_ctl stop: %w: %s", err, strings.TrimSpace(stopOutput.String()))
	}

	return nil
}

func runPgResetwal(ctx context.Context, settings Settings) error {
	resetCommand, err := newRestrictedCommand(
		ctx,
		settings.getProgramPath("pg_resetwal"),
		"--force",
		"--pgdata="+settings.getClusterFolder(),
	)
	if err != nil {
		return err
	}

	var resetOutput bytes.Buffer
	resetCommand.Stdout = &resetOutput
	resetCommand.Stderr = &resetOutput

	if err := runRestricted(resetCommand); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(resetOutput.String()))
	}

	settings.Logger.Warn("pg_resetwal finished", "output", strings.TrimSpace(resetOutput.String()))

	return nil
}

// Another program on the port would make the server fail to start, which the recovery path would
// then misread as a damaged cluster and answer with a WAL reset. Checking first keeps the two
// apart and names the real problem.
func ensurePortIsFree(port int) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("port %d on 127.0.0.1, needed by the metadata database, is in use: %w", port, err)
	}

	return listener.Close()
}

func pingDatabase(ctx context.Context, connectionString string) error {
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	connection, err := pgx.Connect(pingCtx, connectionString)
	if err != nil {
		return err
	}

	return connection.Close(context.Background())
}

func ensureApplicationDatabase(ctx context.Context, maintenanceConnectionString string) error {
	connection, err := pgx.Connect(ctx, maintenanceConnectionString)
	if err != nil {
		return fmt.Errorf("connect to the metadata database server: %w", err)
	}
	defer func() { _ = connection.Close(context.Background()) }()

	var isDatabasePresent bool

	err = connection.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)",
		applicationDatabaseName,
	).Scan(&isDatabasePresent)
	if err != nil {
		return fmt.Errorf("look up the %s database: %w", applicationDatabaseName, err)
	}

	if isDatabasePresent {
		return nil
	}

	createStatement := "CREATE DATABASE " + pgx.Identifier{applicationDatabaseName}.Sanitize() +
		" OWNER " + pgx.Identifier{superuserName}.Sanitize()
	if _, err := connection.Exec(ctx, createStatement); err != nil {
		return fmt.Errorf("create the %s database: %w", applicationDatabaseName, err)
	}

	return nil
}
