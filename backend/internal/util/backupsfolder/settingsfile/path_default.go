//go:build !windows

package settingsfile

func isSamePath(firstPath, secondPath string) bool {
	return firstPath == secondPath
}
