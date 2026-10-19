//go:build windows

package settingsfile

import "strings"

func isSamePath(firstPath, secondPath string) bool {
	return strings.EqualFold(firstPath, secondPath)
}
