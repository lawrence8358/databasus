//go:build !windows

package system_backupsfolder

// Docker and Linux installations set the backups folder through their volume and environment, so
// the setting stays as it was before the Windows installer added it.
const isChangeSupported = false

func prepareBackupsFolder(_ string) error {
	return ErrBackupsFolderNotChangeable
}
