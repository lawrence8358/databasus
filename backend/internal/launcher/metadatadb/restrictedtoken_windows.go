package metadatadb

import (
	"context"
	"fmt"
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

	return restrictedToken, nil
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
