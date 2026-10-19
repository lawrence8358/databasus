// Package backupsfolder tells local storage where to write and where to look for backup files
// while the process runs. Normally that is the one backups folder from the environment. When the
// Windows launcher names a settings file (see settingsfile), an administrator can switch the
// folder at runtime; backups already written stay where they are, and files are looked up in the
// current folder first, then in the folders used before it.
package backupsfolder
