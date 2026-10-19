//go:build windows

// Package installcheck answers the installer's questions before anything is installed: can the
// Databasus service (LocalSystem) write the chosen folders, are they on a local disk, and is the
// chosen Web UI port free. The installer runs these checks as an administrator, so the folder
// check evaluates the folder's ACL for the service account instead of trying a write as itself.
package installcheck
