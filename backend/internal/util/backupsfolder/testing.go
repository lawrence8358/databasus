package backupsfolder

import "testing"

// RestoreFoldersAfterTest puts the folders of the running process back when the test ends, for a
// test that changes them through the API.
func RestoreFoldersAfterTest(t *testing.T) {
	t.Helper()

	originalFolders := getActiveFolders()
	t.Cleanup(func() { activeState.Store(originalFolders) })
}

// SwitchCurrentFolderForTest changes the backups folder the way an administrator does and puts
// the original folders back when the test ends.
func SwitchCurrentFolderForTest(t *testing.T, folder string) {
	t.Helper()

	RestoreFoldersAfterTest(t)
	UseSettings(GetSettings().WithCurrentFolder(folder))
}
