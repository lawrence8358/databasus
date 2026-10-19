package launcher

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/util/backupsfolder/settingsfile"
)

func Test_SelectBackupsFolder_WhenNothingRecorded_RecordsFlagFolder(t *testing.T) {
	dataFolder := t.TempDir()
	flagFolder := filepath.Join(t.TempDir(), "backups")

	selection, err := selectBackupsFolder(Config{DataFolder: dataFolder, BackupsFolder: flagFolder}, slog.Default())
	require.NoError(t, err)

	assert.Equal(t, flagFolder, selection.currentFolder)
	assert.Equal(t, settingsfile.GetSettingsFilePath(dataFolder), selection.settingsFilePath)

	settings, err := settingsfile.ReadSettings(selection.settingsFilePath)
	require.NoError(t, err)
	assert.Equal(t, flagFolder, settings.CurrentFolder)
}

func Test_SelectBackupsFolder_WithoutFlagOrRecord_UsesBackupsUnderDataFolder(t *testing.T) {
	dataFolder := t.TempDir()

	selection, err := selectBackupsFolder(Config{DataFolder: dataFolder}, slog.Default())
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dataFolder, "backups"), selection.currentFolder)
}

func Test_SelectBackupsFolder_WhenFolderRecorded_IgnoresFlag(t *testing.T) {
	dataFolder := t.TempDir()
	recordedFolder := filepath.Join(t.TempDir(), "recorded")
	require.NoError(t, settingsfile.WriteSettings(
		settingsfile.GetSettingsFilePath(dataFolder),
		settingsfile.Settings{CurrentFolder: recordedFolder},
	))

	selection, err := selectBackupsFolder(
		Config{DataFolder: dataFolder, BackupsFolder: filepath.Join(t.TempDir(), "from-command-line")},
		slog.Default(),
	)
	require.NoError(t, err)

	assert.Equal(t, recordedFolder, selection.currentFolder)
}
