// Package securefile writes files that hold credentials (secret key, .pgpass, client certificates,
// .my.cnf). On Linux the owner-only mode bits are enough. Windows ignores mode bits, so there the
// file gets a protected ACL that grants access only to SYSTEM, the Administrators group and, unless
// the caller asks for ServiceOnly, the current account.
package securefile

import (
	"fmt"
	"os"
)

// Readers says who besides SYSTEM and Administrators may open a restricted file on Windows.
type Readers int

const (
	// CurrentAccount also lets the account that writes the file read it back.
	CurrentAccount Readers = iota
	// ServiceOnly leaves out the writing account. The installer, running as an administrator,
	// writes files only the LocalSystem service needs; the installing user must not keep access.
	ServiceOnly
)

// WriteFile writes data to path so that only the owner can read it. The file is restricted before
// any secret content is written, so there is no window where it is readable by other users.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	return WriteFileFor(path, data, perm, CurrentAccount)
}

// WriteFileFor is WriteFile with an explicit choice of readers.
func WriteFileFor(path string, data []byte, perm os.FileMode, readers Readers) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	// The handle keeps the write access it was opened with after the DACL drops the writer.
	if err := restrict(path, false, readers); err != nil {
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
	return RestrictDirFor(path, perm, CurrentAccount)
}

// RestrictDirFor is RestrictDir with an explicit choice of readers.
func RestrictDirFor(path string, perm os.FileMode, readers Readers) error {
	if err := os.Chmod(path, perm); err != nil {
		return err
	}

	return restrict(path, true, readers)
}
