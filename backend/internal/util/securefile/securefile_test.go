package securefile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_WriteFile_ContentRoundTrips_AndOverwriteReplacesIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")

	require.NoError(t, WriteFile(path, []byte("first-value"), 0o600))
	require.NoError(t, WriteFile(path, []byte("second"), 0o600))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "second", string(content))
}

func Test_WriteFile_OnPosix_ModeIsOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits only")
	}

	path := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, WriteFile(path, []byte("v"), 0o600))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func Test_RestrictDir_OnPosix_ModeIsApplied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits only")
	}

	dir := t.TempDir()
	require.NoError(t, RestrictDir(dir, 0o700))

	info, err := os.Stat(dir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}
