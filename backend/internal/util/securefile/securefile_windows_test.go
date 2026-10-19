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

func Test_WriteFileFor_WithServiceOnly_DaclLeavesOutTheWritingAccount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, WriteFileFor(path, []byte("v"), 0o600, ServiceOnly))

	assertOnlyTrustedPrincipals(t, path, true)

	user, err := currentUserSid()
	require.NoError(t, err)
	if user.IsWellKnown(windows.WinLocalSystemSid) {
		t.Skip("the test account is SYSTEM, which the ACL grants anyway")
	}

	for _, principal := range readDaclPrincipals(t, path) {
		require.False(t, principal.Equals(user), "the writing account must not be in the ACL")
	}
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

	user, err := currentUserSid()
	require.NoError(t, err)
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	require.NoError(t, err)
	admins, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	require.NoError(t, err)

	principals := readDaclPrincipals(t, path)
	require.NotEmpty(t, principals)

	for _, principal := range principals {
		require.True(t,
			principal.Equals(user) || principal.Equals(system) || principal.Equals(admins),
			"unexpected principal in ACL: %s", principal.String())
	}
}

func readDaclPrincipals(t *testing.T, path string) []*windows.SID {
	t.Helper()

	sd, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION,
	)
	require.NoError(t, err)

	dacl, _, err := sd.DACL()
	require.NoError(t, err)
	require.NotNil(t, dacl)

	principals := make([]*windows.SID, 0, dacl.AceCount)

	for i := uint16(0); i < dacl.AceCount; i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		require.NoError(t, windows.GetAce(dacl, uint32(i), &ace))

		principals = append(principals, (*windows.SID)(unsafe.Pointer(&ace.SidStart)))
	}

	return principals
}
