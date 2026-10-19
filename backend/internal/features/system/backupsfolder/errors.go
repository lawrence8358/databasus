package system_backupsfolder

import "errors"

var (
	ErrBackupsFolderNotChangeable = errors.New(
		"the backups folder can be changed only on an installation made by the Windows installer",
	)
	ErrBackupsFolderRefused = errors.New("this folder cannot hold backups")
)
