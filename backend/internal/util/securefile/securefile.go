// Package securefile writes files that hold credentials (secret key, .pgpass, client certificates,
// .my.cnf). On Linux the owner-only mode bits are enough. Windows ignores mode bits, so there the
// file gets a protected ACL that grants access only to the current account, SYSTEM and the
// Administrators group.
package securefile

import (
	"fmt"
	"os"
)

// WriteFile writes data to path so that only the owner can read it. The file is restricted before
// any secret content is written, so there is no window where it is readable by other users.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	if err := restrictToOwner(path, false); err != nil {
		_ = file.Close()

		return fmt.Errorf("failed to restrict access to %s: %w", path, err)
	}

	if _, err := file.Write(data); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

// RestrictDir restricts an existing directory to the owner. Files created inside it inherit the
// restriction on Windows. On Linux it sets the mode to perm.
func RestrictDir(path string, perm os.FileMode) error {
	if err := os.Chmod(path, perm); err != nil {
		return err
	}

	return restrictToOwner(path, true)
}
