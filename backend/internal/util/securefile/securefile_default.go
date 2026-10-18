//go:build !windows

package securefile

// POSIX mode bits passed to the open call already restrict the file; nothing more to do.
func restrictToOwner(_ string, _ bool) error {
	return nil
}
