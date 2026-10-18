package externaldb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"databasus-backend/internal/util/securefile"
)

func Test_SaveThenLoad_ReturnsTrimmedConnectionString(t *testing.T) {
	folder := t.TempDir()

	require.NoError(t, Save(folder, "  host=db user=u password=secret dbname=d\r\n", securefile.CurrentAccount))

	dsn, err := Load(GetFilePath(folder))
	require.NoError(t, err)
	require.Equal(t, "host=db user=u password=secret dbname=d", dsn)
}

func Test_Save_WithEmptyString_ReturnsError(t *testing.T) {
	require.Error(t, Save(t.TempDir(), "  ", securefile.CurrentAccount))
}

func Test_Load_WithBomAndEmptyContent_HandlesBoth(t *testing.T) {
	path := GetFilePath(t.TempDir())
	require.NoError(t, os.WriteFile(path, []byte("\xef\xbb\xbfhost=db"), 0o600))

	dsn, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "host=db", dsn)

	require.NoError(t, os.WriteFile(path, []byte(" \n"), 0o600))

	_, err = Load(path)
	require.Error(t, err)
}

func Test_VerifyConnection_WithUnreachableServer_ErrorDoesNotLeakPassword(t *testing.T) {
	dsn := "postgres://user:hunter2secret@127.0.0.1:1/db?sslmode=disable&connect_timeout=2"

	err := VerifyConnection(t.Context(), dsn)

	require.Error(t, err)
	require.NotContains(t, err.Error(), "hunter2secret")
}

func Test_VerifyConnection_WithMalformedString_ErrorDoesNotLeakPassword(t *testing.T) {
	err := VerifyConnection(t.Context(), "postgres://user:hunter2secret@host:notaport/db")

	require.Error(t, err)
	require.NotContains(t, err.Error(), "hunter2secret")
}
