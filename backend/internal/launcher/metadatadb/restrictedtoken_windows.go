package metadatadb

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const disableMaxPrivilege = 0x1

var procCreateRestrictedToken = windows.NewLazySystemDLL("advapi32.dll").NewProc("CreateRestrictedToken")

// PostgreSQL refuses to run with administrative rights, and a service account such as LocalSystem
// has them. pg_ctl solves this by starting the server under a restricted copy of its own token,
// with the Administrators and Power Users groups set to deny-only and every privilege dropped; the
// same is done here so the server can stay a direct child of the launcher.
func newRestrictedCommand(ctx context.Context, programPath string, arguments ...string) (*exec.Cmd, error) {
	restrictedToken, err := createRestrictedToken()
	if err != nil {
		return nil, err
	}

	command := exec.CommandContext(ctx, programPath, arguments...)
	// initdb and pg_ctl restrict themselves again unless told the token is restricted already.
	command.Env = append(os.Environ(), "PG_RESTRICT_EXEC=1")
	command.SysProcAttr = &syscall.SysProcAttr{
		Token:         syscall.Token(restrictedToken),
		HideWindow:    true,
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP,
	}

	return command, nil
}

func createRestrictedToken() (windows.Token, error) {
	var processToken windows.Token

	err := windows.OpenProcessToken(
		windows.CurrentProcess(),
		windows.TOKEN_ALL_ACCESS,
		&processToken,
	)
	if err != nil {
		return 0, fmt.Errorf("open process token: %w", err)
	}
	defer func() { _ = processToken.Close() }()

	administratorsSid, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return 0, fmt.Errorf("create Administrators SID: %w", err)
	}

	powerUsersSid, err := windows.CreateWellKnownSid(windows.WinBuiltinPowerUsersSid)
	if err != nil {
		return 0, fmt.Errorf("create Power Users SID: %w", err)
	}

	sidsToDisable := []windows.SIDAndAttributes{
		{Sid: administratorsSid},
		{Sid: powerUsersSid},
	}

	var restrictedToken windows.Token

	result, _, callErr := procCreateRestrictedToken.Call(
		uintptr(processToken),
		disableMaxPrivilege,
		uintptr(len(sidsToDisable)),
		uintptr(unsafe.Pointer(&sidsToDisable[0])),
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&restrictedToken)),
	)
	if result == 0 {
		return 0, fmt.Errorf("create restricted token: %w", callErr)
	}

	if err := grantTokenUserInDefaultDacl(restrictedToken); err != nil {
		_ = restrictedToken.Close()

		return 0, err
	}

	return restrictedToken, nil
}

// tokenDefaultDacl mirrors TOKEN_DEFAULT_DACL.
type tokenDefaultDacl struct {
	DefaultDacl *windows.ACL
}

// An elevated administrator's token gives its default DACL, and so every object the restricted
// process creates (its own token, child processes such as initdb's postgres), only to
// Administrators and SYSTEM. With Administrators set to deny-only the process could then not
// open its own token or its children. pg_ctl adds the token user for the same reason
// (AddUserToTokenDacl). Under LocalSystem the user is SYSTEM, so this changes nothing there.
func grantTokenUserInDefaultDacl(token windows.Token) error {
	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return fmt.Errorf("read the restricted token user: %w", err)
	}

	var size uint32
	_ = windows.GetTokenInformation(token, windows.TokenDefaultDacl, nil, 0, &size)
	if size == 0 {
		return fmt.Errorf("read the restricted token default DACL size")
	}

	buffer := make([]byte, size)
	if err := windows.GetTokenInformation(token, windows.TokenDefaultDacl, &buffer[0], size, &size); err != nil {
		return fmt.Errorf("read the restricted token default DACL: %w", err)
	}

	currentDacl := (*tokenDefaultDacl)(unsafe.Pointer(&buffer[0])).DefaultDacl

	newDacl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.GRANT_ACCESS,
		Inheritance:       windows.OBJECT_INHERIT_ACE,
		Trustee: windows.TRUSTEE{
			TrusteeForm:  windows.TRUSTEE_IS_SID,
			TrusteeType:  windows.TRUSTEE_IS_USER,
			TrusteeValue: windows.TrusteeValueFromSID(tokenUser.User.Sid),
		},
	}}, currentDacl)
	if err != nil {
		return fmt.Errorf("build the restricted token default DACL: %w", err)
	}

	newDefault := tokenDefaultDacl{DefaultDacl: newDacl}

	err = windows.SetTokenInformation(
		token,
		windows.TokenDefaultDacl,
		(*byte)(unsafe.Pointer(&newDefault)),
		uint32(unsafe.Sizeof(newDefault)),
	)
	if err != nil {
		return fmt.Errorf("set the restricted token default DACL: %w", err)
	}

	return nil
}

func startRestricted(command *exec.Cmd) error {
	err := command.Start()
	_ = windows.Token(command.SysProcAttr.Token).Close()

	return err
}

func runRestricted(command *exec.Cmd) error {
	if err := startRestricted(command); err != nil {
		return err
	}

	return command.Wait()
}
