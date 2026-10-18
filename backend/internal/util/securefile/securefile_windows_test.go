//go:build windows

package securefile

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func Test_WriteFile_OnWindows_DaclIsProtectedAndOnlyOwnerSystemAdmins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, WriteFile(path, []byte("v"), 0o600))

	assertOnlyTrustedPrincipals(t, path, true)
}

func Test_RestrictDir_OnWindows_FilesCreatedInsideInheritRestriction(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, RestrictDir(dir, 0o700))

	inner := filepath.Join(dir, "plain.txt")
	require.NoError(t, os.WriteFile(inner, []byte("v"), 0o600))

	assertOnlyTrustedPrincipals(t, dir, true)
	assertOnlyTrustedPrincipals(t, inner, false)
}

func assertOnlyTrustedPrincipals(t *testing.T, path string, wantProtected bool) {
	t.Helper()

	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	require.NoError(t, err)

	control, _, err := sd.Control()
	require.NoError(t, err)
	if wantProtected {
		require.NotZero(t, control&windows.SE_DACL_PROTECTED, "DACL must not inherit from the parent")
	}

	dacl, _, err := sd.DACL()
	require.NoError(t, err)
	require.NotNil(t, dacl)

	user, err := currentUserSid()
	require.NoError(t, err)
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	require.NoError(t, err)
	admins, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	require.NoError(t, err)

	require.NotZero(t, dacl.AceCount)

	for i := uint16(0); i < dacl.AceCount; i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		require.NoError(t, windows.GetAce(dacl, uint32(i), &ace))

		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		require.True(t,
			sid.Equals(user) || sid.Equals(system) || sid.Equals(admins),
			"unexpected principal in ACL: %s", sid.String())
	}
}
