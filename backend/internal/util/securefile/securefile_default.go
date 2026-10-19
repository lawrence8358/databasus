//go:build !windows

package securefile

// POSIX mode bits passed to the open call already restrict the file; nothing more to do.
func restrict(_ string, _ bool, _ Readers) error {
	return nil
}
