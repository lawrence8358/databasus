//go:build windows

package settingsfile

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// A standard user may create files in a folder under ProgramData before Databasus does, so the
// settings file counts only when an account that could have written it anyway owns it.
func checkIsOwnedByTrustedAccount(path string) error {
	securityDescriptor, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	if err != nil {
		return fmt.Errorf("read the owner of %s: %w", path, err)
	}

	owner, _, err := securityDescriptor.Owner()
	if err != nil {
		return fmt.Errorf("read the owner of %s: %w", path, err)
	}

	if owner.IsWellKnown(windows.WinLocalSystemSid) || owner.IsWellKnown(windows.WinBuiltinAdministratorsSid) {
		return nil
	}

	tokenUser, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("read the current account: %w", err)
	}

	if windows.EqualSid(owner, tokenUser.User.Sid) {
		return nil
	}

	return fmt.Errorf(
		"refusing %s: it is owned by %s, not by SYSTEM, Administrators or the account Databasus "+
			"runs as; delete it and set the backups folder again",
		path,
		owner.String(),
	)
}
