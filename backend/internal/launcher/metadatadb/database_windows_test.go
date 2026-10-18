package metadatadb

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/util/tools"
)

func Test_Start_WithEmptyDataFolder_CreatesClusterReachableOnlyWithItsPassword(t *testing.T) {
	settings := newTestSettings(t)

	database := startTestDatabase(t, settings)

	connection := connectTestDatabase(t, database.GetDsn())

	var listenAddresses string
	require.NoError(t, connection.QueryRow(t.Context(), "SHOW listen_addresses").Scan(&listenAddresses))
	assert.Equal(t, "127.0.0.1", listenAddresses)

	var currentDatabase string
	require.NoError(t, connection.QueryRow(t.Context(), "SELECT current_database()").Scan(&currentDatabase))
	assert.Equal(t, applicationDatabaseName, currentDatabase)

	wrongPasswordDsn := "host=127.0.0.1 port=" + strconv.Itoa(settings.Port) +
		" user=postgres password=wrong dbname=databasus sslmode=disable"
	_, err := pgx.Connect(t.Context(), wrongPasswordDsn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password authentication failed")

	passwordFile, err := os.ReadFile(filepath.Join(settings.DataFolder, "metadata-db.password"))
	require.NoError(t, err)
	assert.Len(t, string(passwordFile), 65)

	if lanAddress, isFound := findNonLoopbackAddress(); isFound {
		_, err := net.DialTimeout("tcp", net.JoinHostPort(lanAddress, strconv.Itoa(settings.Port)), 2*time.Second)
		assert.Error(t, err, "the metadata database must not be reachable on %s", lanAddress)
	}
}

func Test_Start_AfterStop_KeepsExistingData(t *testing.T) {
	settings := newTestSettings(t)

	firstDatabase := startTestDatabase(t, settings)
	firstDsn := firstDatabase.GetDsn()
	createMarkerRow(t, firstDsn, "kept across restarts")
	require.NoError(t, firstDatabase.Stop())

	secondDatabase := startTestDatabase(t, settings)

	assert.Equal(t, firstDsn, secondDatabase.GetDsn())
	assert.Equal(t, "kept across restarts", readMarkerRow(t, secondDatabase.GetDsn()))
}

func Test_Stop_ShutsServerDownCleanly(t *testing.T) {
	settings := newTestSettings(t)

	database := startTestDatabase(t, settings)
	require.NoError(t, database.Stop())

	assert.NoFileExists(t, filepath.Join(settings.DataFolder, "pgdata", "postmaster.pid"))
	assert.NoError(t, ensurePortIsFree(settings.Port))

	select {
	case exitErr := <-database.GetUnexpectedExitChannel():
		t.Fatalf("a requested stop was reported as an unexpected exit: %v", exitErr)
	case <-time.After(500 * time.Millisecond):
	}
}

func Test_Start_AfterServerWasKilled_RecoversWithCommittedData(t *testing.T) {
	settings := newTestSettings(t)

	killedDatabase := startTestDatabase(t, settings)
	createMarkerRow(t, killedDatabase.GetDsn(), "survives a killed server")

	require.NoError(t, killedDatabase.serverCommand.Process.Kill())

	select {
	case exitErr := <-killedDatabase.GetUnexpectedExitChannel():
		assert.ErrorIs(t, exitErr, ErrServerExited)
	case <-time.After(30 * time.Second):
		t.Fatal("the killed server was not reported as an unexpected exit")
	}

	assert.FileExists(t, filepath.Join(settings.DataFolder, "pgdata", "postmaster.pid"))

	recoveredDatabase := startTestDatabase(t, settings)

	assert.Equal(t, "survives a killed server", readMarkerRow(t, recoveredDatabase.GetDsn()))
}

func newTestSettings(t *testing.T) Settings {
	t.Helper()

	return Settings{
		DataFolder:   filepath.Join(t.TempDir(), "data"),
		BinFolder:    tools.GetMetadataPostgresqlBinDir(),
		Port:         findFreePort(t),
		ServerOutput: io.Discard,
		Logger:       slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

func startTestDatabase(t *testing.T, settings Settings) *Database {
	t.Helper()

	startCtx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()

	database, err := Start(startCtx, settings)
	require.NoError(t, err)

	t.Cleanup(func() { _ = database.Stop() })

	return database
}

func connectTestDatabase(t *testing.T, dsn string) *pgx.Conn {
	t.Helper()

	connection, err := pgx.Connect(t.Context(), dsn)
	require.NoError(t, err)

	t.Cleanup(func() { _ = connection.Close(context.Background()) })

	return connection
}

func createMarkerRow(t *testing.T, dsn, marker string) {
	t.Helper()

	connection, err := pgx.Connect(t.Context(), dsn)
	require.NoError(t, err)
	defer func() { _ = connection.Close(context.Background()) }()

	_, err = connection.Exec(t.Context(), "CREATE TABLE launcher_marker (marker text)")
	require.NoError(t, err)

	_, err = connection.Exec(t.Context(), "INSERT INTO launcher_marker VALUES ($1)", marker)
	require.NoError(t, err)
}

func readMarkerRow(t *testing.T, dsn string) string {
	t.Helper()

	connection, err := pgx.Connect(t.Context(), dsn)
	require.NoError(t, err)
	defer func() { _ = connection.Close(context.Background()) }()

	var marker string
	require.NoError(t, connection.QueryRow(t.Context(), "SELECT marker FROM launcher_marker").Scan(&marker))

	return marker
}

func findFreePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()

	return listener.Addr().(*net.TCPAddr).Port
}

func findNonLoopbackAddress() (string, bool) {
	interfaceAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", false
	}

	for _, interfaceAddress := range interfaceAddresses {
		ipNetwork, isIPNetwork := interfaceAddress.(*net.IPNet)
		if !isIPNetwork || ipNetwork.IP.IsLoopback() || ipNetwork.IP.To4() == nil {
			continue
		}

		return ipNetwork.IP.String(), true
	}

	return "", false
}
