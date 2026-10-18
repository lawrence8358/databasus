//go:build !windows

package settingsfile

// Only the Windows launcher points the backend at a settings file; elsewhere the file's mode bits
// already say who could have written it.
func checkIsOwnedByTrustedAccount(_ string) error {
	return nil
}
