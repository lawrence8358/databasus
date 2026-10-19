package system_backupsfolder

type BackupsFolderResponse struct {
	// IsChangeable is false outside an installation made by the Windows installer; the folders are
	// then left out.
	IsChangeable    bool     `json:"isChangeable"`
	CurrentFolder   string   `json:"currentFolder,omitempty"`
	PreviousFolders []string `json:"previousFolders"`
}

type ChangeBackupsFolderRequest struct {
	Folder string `json:"folder" binding:"required"`
}
