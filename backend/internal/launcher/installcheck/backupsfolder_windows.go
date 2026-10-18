//go:build windows

package installcheck

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// FILE_NAME_NORMALIZED | VOLUME_NAME_DOS, both zero, which x/sys/windows does not define.
const normalizedDosFinalPath = 0

// OwnFolders are the folders Databasus itself owns, which a backups folder must stay out of.
// An empty field is not checked.
type OwnFolders struct {
	DataFolder    string
	InstallFolder string
}

// CheckBackupsFolder is the one rule set for a backups folder, used by the installer and by the
// Web UI setting. The service writes there as LocalSystem, so besides being writable the folder
// must not be one where those writes would land among system or program files.
func CheckBackupsFolder(folder string, ownFolders OwnFolders) error {
	if strings.ContainsAny(folder, "\r\n\x00") {
		return fmt.Errorf("%q contains a line break", folder)
	}

	if err := checkIsLocalAbsolutePath(folder); err != nil {
		return err
	}

	cleanFolder := filepath.Clean(folder)

	if err := checkIsNotProtectedFolder(cleanFolder, ownFolders); err != nil {
		return err
	}

	// A junction or symbolic link inside the path could lead the writes somewhere the rules above
	// forbid, so the rules also apply to where the existing part of the path really points.
	resolvedFolder, err := resolveExistingPart(cleanFolder)
	if err != nil {
		return err
	}

	if !strings.EqualFold(resolvedFolder, cleanFolder) {
		if err := checkIsLocalAbsolutePath(resolvedFolder); err != nil {
			return fmt.Errorf("%s leads to %s: %w", cleanFolder, resolvedFolder, err)
		}

		if err := checkIsNotProtectedFolder(resolvedFolder, ownFolders); err != nil {
			return fmt.Errorf("%s leads to %s: %w", cleanFolder, resolvedFolder, err)
		}

		if err := checkIsWritableFixedVolume(filepath.VolumeName(resolvedFolder) + `\`); err != nil {
			return fmt.Errorf("%s leads to %s: %w", cleanFolder, resolvedFolder, err)
		}
	}

	return CheckServiceCanWriteFolder(cleanFolder)
}

func checkIsNotProtectedFolder(folder string, ownFolders OwnFolders) error {
	if filepath.Dir(folder) == folder {
		return fmt.Errorf(
			`%s is the root of a drive; choose a folder on it, such as %sDatabasusBackups`,
			folder,
			folder,
		)
	}

	for _, systemFolder := range getSystemFoldersAndTheirContents() {
		if isSameOrInside(folder, systemFolder) {
			return fmt.Errorf("%s is a Windows system or program folder; choose a folder of its own", folder)
		}
	}

	for _, sharedFolder := range getSharedRootFolders() {
		if strings.EqualFold(folder, sharedFolder) {
			return fmt.Errorf("%s is shared by many programs or users; choose a folder of its own inside it", folder)
		}
	}

	if ownFolders.InstallFolder != "" && isSameOrInside(folder, filepath.Clean(ownFolders.InstallFolder)) {
		return fmt.Errorf("%s is inside the Databasus program folder, which Setup replaces on upgrade", folder)
	}

	if ownFolders.DataFolder != "" {
		dataFolder := filepath.Clean(ownFolders.DataFolder)
		defaultBackupsFolder := filepath.Join(dataFolder, "backups")

		if isSameOrInside(folder, dataFolder) && !isSameOrInside(folder, defaultBackupsFolder) {
			return fmt.Errorf(
				"%s is inside the program data folder, which holds the metadata database; use %s or a folder outside %s",
				folder,
				defaultBackupsFolder,
				dataFolder,
			)
		}
	}

	return nil
}

func getSystemFoldersAndTheirContents() []string {
	return getKnownFolders(
		windows.FOLDERID_Windows,
		windows.FOLDERID_ProgramFiles,
		windows.FOLDERID_ProgramFilesX86,
	)
}

func getSharedRootFolders() []string {
	return getKnownFolders(
		windows.FOLDERID_ProgramData,
		windows.FOLDERID_UserProfiles,
		windows.FOLDERID_Public,
	)
}

func getKnownFolders(folderIDs ...*windows.KNOWNFOLDERID) []string {
	folders := []string{}

	for _, folderID := range folderIDs {
		folder, err := windows.KnownFolderPath(folderID, windows.KF_FLAG_DEFAULT)
		if err == nil && folder != "" {
			folders = append(folders, filepath.Clean(folder))
		}
	}

	return folders
}

func isSameOrInside(folder, parentFolder string) bool {
	if strings.EqualFold(folder, parentFolder) {
		return true
	}

	return strings.HasPrefix(strings.ToLower(folder), strings.ToLower(strings.TrimSuffix(parentFolder, `\`)+`\`))
}

func resolveExistingPart(folder string) (string, error) {
	existingFolder, _, err := findNearestExistingFolder(folder)
	if err != nil {
		return "", err
	}

	resolvedExistingFolder, err := getFinalPath(existingFolder)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", existingFolder, err)
	}

	missingPart, err := filepath.Rel(existingFolder, folder)
	if err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(resolvedExistingFolder, missingPart)), nil
}

// The final path follows junctions and symbolic links and expands 8.3 short names, which
// filepath.EvalSymlinks does not all do on Windows.
func getFinalPath(existingFolder string) (string, error) {
	folderPointer, err := windows.UTF16PtrFromString(existingFolder)
	if err != nil {
		return "", err
	}

	handle, err := windows.CreateFile(
		folderPointer,
		0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return "", err
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	buffer := make([]uint16, windows.MAX_LONG_PATH)

	length, err := windows.GetFinalPathNameByHandle(
		handle,
		&buffer[0],
		uint32(len(buffer)),
		normalizedDosFinalPath,
	)
	if err != nil {
		return "", err
	}

	finalPath := windows.UTF16ToString(buffer[:length])

	if uncPath, isUnc := strings.CutPrefix(finalPath, `\\?\UNC\`); isUnc {
		return `\\` + uncPath, nil
	}

	return strings.TrimPrefix(finalPath, `\\?\`), nil
}
