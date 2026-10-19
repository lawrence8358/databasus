//go:build windows

package installcheck

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func Test_CheckServiceCanWriteFolder_WhenFolderIsWritable_ReturnsNil(t *testing.T) {
	folder := t.TempDir()

	assert.NoError(t, CheckServiceCanWriteFolder(folder))

	entries, err := os.ReadDir(folder)
	require.NoError(t, err)
	assert.Empty(t, entries, "the write probe must not leave files behind")
}

func Test_CheckServiceCanWriteFolder_WhenFolderIsMissingUnderWritableParent_ReturnsNil(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "not", "yet", "created")

	assert.NoError(t, CheckServiceCanWriteFolder(folder))
	assert.NoDirExists(t, folder, "the check must not create the folder")
}

func Test_CheckServiceCanWriteFolder_WhenOnlyAdministratorsMayWrite_ReturnsNil(t *testing.T) {
	folder := t.TempDir()
	restrictFolderTo(t, folder, windows.WinBuiltinAdministratorsSid)

	err := checkAccountRightsOnly(folder, windows.WinLocalSystemSid)

	assert.NoError(t, err, "LocalSystem is a member of Administrators")
}

func Test_CheckServiceCanWriteFolder_WhenOnlyAnotherAccountMayWrite_ReturnsError(t *testing.T) {
	folder := t.TempDir()
	restrictFolderTo(t, folder, windows.WinBuiltinUsersSid)

	err := checkAccountRightsOnly(folder, windows.WinLocalSystemSid)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot write to")
}

func Test_CheckServiceCanWriteFolder_WhenParentDeniesCreatingFolders_ReturnsError(t *testing.T) {
	parent := t.TempDir()
	restrictFolderTo(t, parent, windows.WinBuiltinUsersSid)

	err := checkAccountRightsOnly(filepath.Join(parent, "backups"), windows.WinLocalSystemSid)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create")
}

func Test_CheckServiceCanWriteFolder_WhenPathIsNotLocalAndAbsolute_ReturnsError(t *testing.T) {
	for _, folder := range []string{`\\server\share\backups`, `//server/share`, `backups`, `\backups`, `C:backups`} {
		assert.Error(t, CheckServiceCanWriteFolder(folder), folder)
	}
}

func Test_CheckServiceCanWriteFolder_WhenPathIsAFile_ReturnsError(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0o600))

	err := CheckServiceCanWriteFolder(filepath.Join(filePath, "backups"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is a file")
}

func Test_CheckPortIsFree_WhenPortIsTaken_ReturnsError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = listener.Close() }()

	err = CheckPortIsFree(listener.Addr().(*net.TCPAddr).Port)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already in use")
}

func Test_CheckPortIsFree_WhenPortIsFree_ReturnsNil(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	assert.NoError(t, CheckPortIsFree(port))
}

// The test account usually cannot write a folder it restricted to someone else, so these cases
// check only the ACL evaluation for the given account, without the write probe.
func checkAccountRightsOnly(folder string, accountType windows.WELL_KNOWN_SID_TYPE) error {
	account, err := windows.CreateWellKnownSid(accountType)
	if err != nil {
		return err
	}

	_, err = checkAccountRights(filepath.Clean(folder), account, "the test account")

	return err
}

// restrictFolderTo replaces the folder's DACL with a protected one granting full control to
// grantee only; the test account keeps READ_CONTROL and WRITE_DAC as the owner, so cleanup can
// restore access.
func restrictFolderTo(t *testing.T, folder string, grantee windows.WELL_KNOWN_SID_TYPE) {
	t.Helper()

	granteeSid, err := windows.CreateWellKnownSid(grantee)
	require.NoError(t, err)

	setProtectedDacl(t, folder, granteeSid)

	t.Cleanup(func() {
		token := windows.GetCurrentProcessToken()

		user, err := token.GetTokenUser()
		if err != nil {
			return
		}

		setProtectedDacl(t, folder, user.User.Sid)
	})
}

func setProtectedDacl(t *testing.T, folder string, sid *windows.SID) {
	t.Helper()

	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_UNKNOWN,
			TrusteeValue: windows.TrusteeValueFromSID(sid),
		},
	}}, nil)
	require.NoError(t, err)

	err = windows.SetNamedSecurityInfo(
		folder,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
	require.NoError(t, err)
}
