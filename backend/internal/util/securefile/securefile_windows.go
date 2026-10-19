//go:build windows

package securefile

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// restrict replaces the DACL of path with a protected one (inheritance from the parent is cut)
// that grants full control to SYSTEM, Administrators and, for CurrentAccount, the current account.
func restrict(path string, isDir bool, readers Readers) error {
	system, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return fmt.Errorf("failed to build SYSTEM SID: %w", err)
	}

	admins, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return fmt.Errorf("failed to build Administrators SID: %w", err)
	}

	sids := []*windows.SID{system, admins}

	if readers == CurrentAccount {
		user, err := currentUserSid()
		if err != nil {
			return err
		}

		sids = append(sids, user)
	}

	inheritance := uint32(windows.NO_INHERITANCE)
	if isDir {
		inheritance = windows.SUB_CONTAINERS_AND_OBJECTS_INHERIT
	}

	var entries []windows.EXPLICIT_ACCESS
	for _, sid := range sids {
		entries = append(entries, windows.EXPLICIT_ACCESS{
			AccessPermissions: windows.GENERIC_ALL,
			AccessMode:        windows.GRANT_ACCESS,
			Inheritance:       inheritance,
			Trustee: windows.TRUSTEE{
				TrusteeForm:  windows.TRUSTEE_IS_SID,
				TrusteeType:  windows.TRUSTEE_IS_UNKNOWN,
				TrusteeValue: windows.TrusteeValueFromSID(sid),
			},
		})
	}

	acl, err := windows.ACLFromEntries(entries, nil)
	if err != nil {
		return fmt.Errorf("failed to build ACL: %w", err)
	}

	return windows.SetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		acl,
		nil,
	)
}

func currentUserSid() (*windows.SID, error) {
	token := windows.GetCurrentProcessToken()

	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("failed to read current user: %w", err)
	}

	return tokenUser.User.Sid, nil
}
