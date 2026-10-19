package settingsfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_WriteSettings_ThenReadSettings_ReturnsSameFolders(t *testing.T) {
	settingsFilePath := GetSettingsFilePath(t.TempDir())
	writtenSettings := Settings{
		CurrentFolder:   filepath.Join(t.TempDir(), "current"),
		PreviousFolders: []string{filepath.Join(t.TempDir(), "first"), filepath.Join(t.TempDir(), "second")},
	}

	require.NoError(t, WriteSettings(settingsFilePath, writtenSettings))

	readSettings, err := ReadSettings(settingsFilePath)
	require.NoError(t, err)
	assert.Equal(t, writtenSettings, readSettings)
	assert.NoFileExists(t, settingsFilePath+".pending")
}

func Test_ReadSettings_WhenFileIsMissing_ReturnsNotFound(t *testing.T) {
	_, err := ReadSettings(GetSettingsFilePath(t.TempDir()))

	assert.ErrorIs(t, err, ErrSettingsNotFound)
}

func Test_ReadSettings_WhenFileHasNoCurrentFolder_ReturnsError(t *testing.T) {
	settingsFilePath := GetSettingsFilePath(t.TempDir())
	require.NoError(t, os.WriteFile(settingsFilePath, []byte("previous="+t.TempDir()+"\n"), 0o600))

	_, err := ReadSettings(settingsFilePath)

	assert.ErrorContains(t, err, "no current folder")
}

func Test_ReadSettings_WhenFolderIsRelative_ReturnsError(t *testing.T) {
	settingsFilePath := GetSettingsFilePath(t.TempDir())
	require.NoError(t, os.WriteFile(settingsFilePath, []byte("current=backups\n"), 0o600))

	_, err := ReadSettings(settingsFilePath)

	assert.ErrorContains(t, err, "not an absolute path")
}

func Test_WriteSettings_WhenFolderHasLineBreak_RefusesIt(t *testing.T) {
	settingsFilePath := GetSettingsFilePath(t.TempDir())
	folderWithLineBreak := filepath.Join(t.TempDir(), "first") + "\nprevious=" + t.TempDir()

	err := WriteSettings(settingsFilePath, Settings{CurrentFolder: folderWithLineBreak})

	assert.ErrorContains(t, err, "line break")
	assert.NoFileExists(t, settingsFilePath)
}

func Test_WithCurrentFolder_KeepsReplacedFolderFirstAmongPrevious(t *testing.T) {
	firstFolder := filepath.Join(t.TempDir(), "first")
	secondFolder := filepath.Join(t.TempDir(), "second")
	thirdFolder := filepath.Join(t.TempDir(), "third")

	settings := Settings{CurrentFolder: firstFolder}.
		WithCurrentFolder(secondFolder).
		WithCurrentFolder(thirdFolder)

	assert.Equal(t, thirdFolder, settings.CurrentFolder)
	assert.Equal(t, []string{secondFolder, firstFolder}, settings.PreviousFolders)
}

func Test_WithCurrentFolder_WhenGoingBackToEarlierFolder_ListsEachFolderOnce(t *testing.T) {
	firstFolder := filepath.Join(t.TempDir(), "first")
	secondFolder := filepath.Join(t.TempDir(), "second")

	settings := Settings{CurrentFolder: firstFolder}.
		WithCurrentFolder(secondFolder).
		WithCurrentFolder(firstFolder)

	assert.Equal(t, firstFolder, settings.CurrentFolder)
	assert.Equal(t, []string{secondFolder}, settings.PreviousFolders)
	assert.Equal(t, []string{firstFolder, secondFolder}, settings.GetFoldersHoldingBackups())
}

func Test_WithCurrentFolder_WhenFolderIsUnchanged_ReturnsSameSettings(t *testing.T) {
	settings := Settings{
		CurrentFolder:   filepath.Join(t.TempDir(), "current"),
		PreviousFolders: []string{filepath.Join(t.TempDir(), "earlier")},
	}

	assert.Equal(t, settings, settings.WithCurrentFolder(settings.CurrentFolder+string(filepath.Separator)))
}
