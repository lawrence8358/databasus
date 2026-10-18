package local_storage

import (
	"errors"
	"syscall"
)

// errNotSameDevice is ERROR_NOT_SAME_DEVICE, which MoveFileEx returns when
// a rename crosses drives. Go maps syscall.EXDEV on Windows to an invented
// value no API call ever returns, so the Win32 code is checked instead.
const errNotSameDevice syscall.Errno = 17

func isCrossDeviceError(err error) bool {
	return errors.Is(err, errNotSameDevice)
}
