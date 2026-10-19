//go:build windows

package winservice

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const uninstallStopTimeout = 3 * time.Minute

var recoveryActions = []mgr.RecoveryAction{
	{Type: mgr.ServiceRestart, Delay: 10 * time.Second},
	{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
	{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
}

type InstallSpec struct {
	ExecutablePath string
	// Arguments are passed to the executable on every service start.
	Arguments []string
}

// Install registers the service, or updates an existing registration in place so the installer
// can rerun it on upgrade; the new settings apply from the next start. It needs administrator
// rights. The service runs as LocalSystem, starts automatically at boot and is restarted by the
// SCM when it crashes or stops with an error.
func Install(spec InstallSpec) error {
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to the service control manager: %w", err)
	}
	defer func() { _ = manager.Disconnect() }()

	service, err := createOrUpdateService(manager, spec)
	if err != nil {
		return err
	}
	defer func() { _ = service.Close() }()

	if err := service.SetRecoveryActions(recoveryActions, uint32(recoveryResetPeriod.Seconds())); err != nil {
		return fmt.Errorf("set the recovery actions: %w", err)
	}

	if err := service.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		return fmt.Errorf("apply the recovery actions to error exits: %w", err)
	}

	if err := setPreShutdownTimeout(service, preShutdownBudget); err != nil {
		return err
	}

	return registerEventSource()
}

// Uninstall stops the service if it runs, removes it and its event log source. Removing a
// service that is not installed succeeds.
func Uninstall() error {
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to the service control manager: %w", err)
	}
	defer func() { _ = manager.Disconnect() }()

	service, err := manager.OpenService(Name)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return removeEventSource()
	}

	if err != nil {
		return fmt.Errorf("open the %s service: %w", Name, err)
	}
	defer func() { _ = service.Close() }()

	if err := stopAndWait(service, uninstallStopTimeout); err != nil {
		return err
	}

	if err := service.Delete(); err != nil {
		return fmt.Errorf("delete the %s service: %w", Name, err)
	}

	return removeEventSource()
}

func createOrUpdateService(manager *mgr.Mgr, spec InstallSpec) (*mgr.Service, error) {
	config := mgr.Config{
		DisplayName:  displayName,
		Description:  description,
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
	}

	service, err := manager.OpenService(Name)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		service, err = manager.CreateService(Name, spec.ExecutablePath, config, spec.Arguments...)
		if err != nil {
			return nil, fmt.Errorf("create the %s service: %w", Name, err)
		}

		return service, nil
	}

	if err != nil {
		return nil, fmt.Errorf("open the %s service: %w", Name, err)
	}

	existingConfig, err := service.Config()
	if err != nil {
		_ = service.Close()

		return nil, fmt.Errorf("read the %s service configuration: %w", Name, err)
	}

	existingConfig.DisplayName = config.DisplayName
	existingConfig.Description = config.Description
	existingConfig.StartType = config.StartType
	existingConfig.ErrorControl = config.ErrorControl
	existingConfig.BinaryPathName = buildCommandLine(spec)

	if err := service.UpdateConfig(existingConfig); err != nil {
		_ = service.Close()

		return nil, fmt.Errorf("update the %s service: %w", Name, err)
	}

	return service, nil
}

func buildCommandLine(spec InstallSpec) string {
	quotedParts := []string{syscall.EscapeArg(spec.ExecutablePath)}

	for _, argument := range spec.Arguments {
		quotedParts = append(quotedParts, syscall.EscapeArg(argument))
	}

	return strings.Join(quotedParts, " ")
}

// The default pre-shutdown timeout is shorter than an orderly stop of the backend and PostgreSQL.
func setPreShutdownTimeout(service *mgr.Service, timeout time.Duration) error {
	info := struct{ PreshutdownTimeout uint32 }{PreshutdownTimeout: uint32(timeout.Milliseconds())}

	err := windows.ChangeServiceConfig2(
		service.Handle,
		windows.SERVICE_CONFIG_PRESHUTDOWN_INFO,
		(*byte)(unsafe.Pointer(&info)),
	)
	if err != nil {
		return fmt.Errorf("set the pre-shutdown timeout: %w", err)
	}

	return nil
}

func stopAndWait(service *mgr.Service, timeout time.Duration) error {
	status, err := service.Query()
	if err != nil {
		return fmt.Errorf("query the %s service: %w", Name, err)
	}

	if status.State == svc.Stopped {
		return nil
	}

	if status.State != svc.StopPending {
		if _, err := service.Control(svc.Stop); err != nil &&
			!errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
			return fmt.Errorf("stop the %s service: %w", Name, err)
		}
	}

	deadline := time.Now().UTC().Add(timeout)

	for time.Now().UTC().Before(deadline) {
		status, err = service.Query()
		if err != nil {
			return fmt.Errorf("query the %s service: %w", Name, err)
		}

		if status.State == svc.Stopped {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("the %s service did not stop within %s", Name, timeout)
}

func registerEventSource() error {
	if err := removeEventSource(); err != nil {
		return err
	}

	if err := eventlog.InstallAsEventCreate(Name, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		return fmt.Errorf("register the %s event log source: %w", Name, err)
	}

	return nil
}

func removeEventSource() error {
	err := eventlog.Remove(Name)
	if err != nil && !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return fmt.Errorf("remove the %s event log source: %w", Name, err)
	}

	return nil
}
