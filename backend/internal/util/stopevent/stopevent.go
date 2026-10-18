//go:build windows

package stopevent

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const handleEnvVariable = "DATABASUS_STOP_EVENT_HANDLE"

type Event struct {
	handle windows.Handle
}

func New() (*Event, error) {
	securityAttributes := windows.SecurityAttributes{InheritHandle: 1}
	securityAttributes.Length = uint32(unsafe.Sizeof(securityAttributes))

	handle, err := windows.CreateEvent(&securityAttributes, 1, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("create the stop event: %w", err)
	}

	return &Event{handle: handle}, nil
}

// GetHandle must go into the child's SysProcAttr.AdditionalInheritedHandles, otherwise the value
// in GetEnvironmentEntry names nothing in the child.
func (e *Event) GetHandle() syscall.Handle {
	return syscall.Handle(e.handle)
}

func (e *Event) GetEnvironmentEntry() string {
	return handleEnvVariable + "=" + strconv.FormatUint(uint64(e.handle), 10)
}

func (e *Event) Set() error {
	return windows.SetEvent(e.handle)
}

func (e *Event) Close() error {
	return windows.CloseHandle(e.handle)
}

// NotifyWhenParentRequestsStop sends os.Interrupt on channel once the parent sets the event it
// passed down. A process started without one is left alone.
func NotifyWhenParentRequestsStop(channel chan<- os.Signal) error {
	rawHandle := os.Getenv(handleEnvVariable)
	if rawHandle == "" {
		return nil
	}

	handleValue, err := strconv.ParseUint(rawHandle, 10, 64)
	if err != nil {
		return fmt.Errorf("%s is not a handle: %w", handleEnvVariable, err)
	}

	handle := windows.Handle(handleValue)

	if _, err := windows.WaitForSingleObject(handle, 0); err != nil {
		return fmt.Errorf("%s does not name a usable event: %w", handleEnvVariable, err)
	}

	go func() {
		if _, err := windows.WaitForSingleObject(handle, windows.INFINITE); err != nil {
			return
		}

		select {
		case channel <- os.Interrupt:
		default:
		}
	}()

	return nil
}
