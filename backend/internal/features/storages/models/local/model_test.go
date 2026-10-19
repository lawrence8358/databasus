package local_storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/config"
	"databasus-backend/internal/util/backupsfolder"
	"databasus-backend/internal/util/encryption"
	"databasus-backend/internal/util/logger"
)

func Test_DeleteFile_WhenWriteWasInterrupted_RemovesTheStagingFile(t *testing.T) {
	storage := &LocalStorage{}
	fileName := "interrupted-" + t.Name()
	stagingPath := filepath.Join(config.GetEnv().TempFolder, fileName)

	require.NoError(t, os.MkdirAll(config.GetEnv().TempFolder, 0o755))
	require.NoError(t, os.WriteFile(stagingPath, []byte("partial"), 0o600))

	err := storage.DeleteFile(t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName)

	require.NoError(t, err)
	assert.NoFileExists(t, stagingPath)
}

func Test_DeleteFile_WhenFileWasPublished_RemovesBothPaths(t *testing.T) {
	storage := &LocalStorage{}
	fileName := "published-" + t.Name()

	require.NoError(t, storage.SaveFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName, strings.NewReader("payload"),
	))

	stagingPath := filepath.Join(config.GetEnv().TempFolder, fileName)
	require.NoError(t, os.WriteFile(stagingPath, []byte("partial"), 0o600))

	err := storage.DeleteFile(t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName)

	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(config.GetEnv().DataFolder, fileName))
	assert.NoFileExists(t, stagingPath)
}

func Test_DeleteFile_WhenNothingExists_Succeeds(t *testing.T) {
	storage := &LocalStorage{}

	err := storage.DeleteFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), "never-written-"+t.Name(),
	)

	assert.NoError(t, err)
}

func Test_DeleteFile_WhenCalledTwice_Succeeds(t *testing.T) {
	storage := &LocalStorage{}
	fileName := "twice-" + t.Name()

	require.NoError(t, storage.SaveFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName, strings.NewReader("payload"),
	))

	require.NoError(t, storage.DeleteFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName,
	))

	assert.NoError(t, storage.DeleteFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName,
	))
}

func Test_DeleteFile_WhenRemovalIsRefused_ReturnsTheFilesystemError(t *testing.T) {
	storage := &LocalStorage{}
	fileName := filepath.Join("locked-"+t.Name(), "artifact")
	directory := filepath.Join(config.GetEnv().DataFolder, filepath.Dir(fileName))

	require.NoError(t, os.MkdirAll(directory, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(config.GetEnv().DataFolder, fileName), []byte("x"), 0o600))
	require.NoError(t, os.Chmod(directory, 0o500))

	t.Cleanup(func() {
		_ = os.Chmod(directory, 0o755)
		_ = os.RemoveAll(directory)
	})

	err := storage.DeleteFile(t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName)

	assert.Error(t, err, "an unconfirmed absence must not be reported as a successful deletion")
}

func Test_GetFile_AfterBackupsFolderChanged_ReadsFileFromEarlierFolder(t *testing.T) {
	storage := &LocalStorage{}
	fileName := "earlier-folder-" + t.Name()
	earlierPath := filepath.Join(backupsfolder.GetCurrentFolder(), fileName)

	require.NoError(t, storage.SaveFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName, strings.NewReader("old backup"),
	))
	t.Cleanup(func() { _ = os.Remove(earlierPath) })

	backupsfolder.SwitchCurrentFolderForTest(t, t.TempDir())

	file, err := storage.GetFile(t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName)
	require.NoError(t, err)

	content, err := io.ReadAll(file)
	require.NoError(t, file.Close())
	require.NoError(t, err)
	assert.Equal(t, "old backup", string(content))
}

func Test_SaveFile_AfterBackupsFolderChanged_WritesToNewFolder(t *testing.T) {
	storage := &LocalStorage{}
	fileName := "new-folder-" + t.Name()
	earlierFolder := backupsfolder.GetCurrentFolder()
	newFolder := t.TempDir()

	backupsfolder.SwitchCurrentFolderForTest(t, newFolder)

	require.NoError(t, storage.SaveFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName, strings.NewReader("new backup"),
	))

	assert.FileExists(t, filepath.Join(newFolder, fileName))
	assert.NoFileExists(t, filepath.Join(earlierFolder, fileName))
	assert.NoFileExists(t, filepath.Join(newFolder, "temp", fileName))
}

func Test_DeleteFile_AfterBackupsFolderChanged_RemovesFileFromEarlierFolder(t *testing.T) {
	storage := &LocalStorage{}
	fileName := "delete-earlier-" + t.Name()
	earlierPath := filepath.Join(backupsfolder.GetCurrentFolder(), fileName)

	require.NoError(t, storage.SaveFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName, strings.NewReader("old backup"),
	))

	backupsfolder.SwitchCurrentFolderForTest(t, t.TempDir())

	require.NoError(t, storage.DeleteFile(
		t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName,
	))

	assert.NoFileExists(t, earlierPath)

	_, err := storage.GetFile(t.Context(), encryption.GetFieldEncryptor(), logger.GetLogger(), fileName)
	assert.Error(t, err)
}
