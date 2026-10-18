package backupsfolder

import "testing"

// SwitchCurrentFolderForTest changes the backups folder the way an administrator does and puts
// the original folders back when the test ends.
func SwitchCurrentFolderForTest(t *testing.T, folder string) {
	t.Helper()

	originalFolders := getActiveFolders()
	t.Cleanup(func() { activeState.Store(originalFolders) })

	UseSettings(originalFolders.settings.WithCurrentFolder(folder))
}
