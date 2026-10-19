//go:build windows

package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/util/backupsfolder/settingsfile"
)

func Test_SetBackupsFolder_WhenFolderChanges_KeepsEarlierFolderListed(t *testing.T) {
	dataFolder := t.TempDir()
	firstFolder := filepath.Join(t.TempDir(), "first")
	secondFolder := filepath.Join(t.TempDir(), "second")

	require.Equal(t, 0, setBackupsFolder([]string{"-data-dir", dataFolder, "-set", firstFolder}))
	require.Equal(t, 0, setBackupsFolder([]string{"-data-dir", dataFolder, "-set", secondFolder}))

	settings, err := settingsfile.ReadSettings(settingsfile.GetSettingsFilePath(dataFolder))
	require.NoError(t, err)
	assert.Equal(t, secondFolder, settings.CurrentFolder)
	assert.Equal(t, []string{firstFolder}, settings.PreviousFolders)
}

func Test_SetBackupsFolder_WhenFolderIsRefused_LeavesSettingsUnchanged(t *testing.T) {
	dataFolder := t.TempDir()
	firstFolder := filepath.Join(t.TempDir(), "first")

	require.Equal(t, 0, setBackupsFolder([]string{"-data-dir", dataFolder, "-set", firstFolder}))
	assert.Equal(t, 1, setBackupsFolder([]string{"-data-dir", dataFolder, "-set", filepath.Join(dataFolder, "pgdata")}))

	settings, err := settingsfile.ReadSettings(settingsfile.GetSettingsFilePath(dataFolder))
	require.NoError(t, err)
	assert.Equal(t, firstFolder, settings.CurrentFolder)
	assert.Empty(t, settings.PreviousFolders)
}

func Test_SetBackupsFolder_OnUpgradeFromVersionWithoutSettings_KeepsFolderInUseListed(t *testing.T) {
	dataFolder := t.TempDir()
	folderInUse := filepath.Join(t.TempDir(), "in-use")
	newFolder := filepath.Join(t.TempDir(), "new")

	require.Equal(t, 0, setBackupsFolder([]string{"-data-dir", dataFolder, "-in-use", folderInUse, "-set", newFolder}))

	settings, err := settingsfile.ReadSettings(settingsfile.GetSettingsFilePath(dataFolder))
	require.NoError(t, err)
	assert.Equal(t, newFolder, settings.CurrentFolder)
	assert.Equal(t, []string{folderInUse}, settings.PreviousFolders)
}

func Test_SetBackupsFolder_WhenKeepingFolderInUse_SkipsNewerRules(t *testing.T) {
	dataFolder := t.TempDir()
	driveRoot := filepath.VolumeName(dataFolder) + `\`

	require.Equal(t, 0, setBackupsFolder([]string{"-data-dir", dataFolder, "-in-use", driveRoot, "-set", driveRoot}))

	settings, err := settingsfile.ReadSettings(settingsfile.GetSettingsFilePath(dataFolder))
	require.NoError(t, err)
	assert.Equal(t, driveRoot, settings.CurrentFolder)
}
