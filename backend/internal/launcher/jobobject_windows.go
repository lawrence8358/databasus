package launcher

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// BindChildProcessesToLauncherLifetime puts the launcher in a job object that kills every process
// in it when the launcher's last handle closes, which happens however the launcher ends. Children
// inherit the job, so a killed launcher cannot leave the backend or the metadata database running
// with the port and data directory still held. The handle is deliberately never closed.
func BindChildProcessesToLauncherLifetime() error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return fmt.Errorf("create job object: %w", err)
	}

	limitInformation := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}

	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limitInformation)),
		uint32(unsafe.Sizeof(limitInformation)),
	)
	if err != nil {
		_ = windows.CloseHandle(job)

		return fmt.Errorf("configure job object: %w", err)
	}

	if err := windows.AssignProcessToJobObject(job, windows.CurrentProcess()); err != nil {
		_ = windows.CloseHandle(job)

		return fmt.Errorf("assign the launcher to its job object: %w", err)
	}

	return nil
}
