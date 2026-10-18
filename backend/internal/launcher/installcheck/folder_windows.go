//go:build windows

package installcheck

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// The service creates the folder when it is missing, so a missing folder needs the right to add a
// subfolder to its nearest existing ancestor; an existing one needs the rights to add and list
// files in it.
const (
	fileListDirectory   = 0x0001
	fileAddFile         = 0x0002
	fileAddSubdirectory = 0x0004
	fileTraverse        = 0x0020

	requiredRightsOnExistingFolder = fileAddFile | fileAddSubdirectory | fileListDirectory | fileTraverse
	requiredRightsOnAncestor       = fileAddSubdirectory | fileTraverse

	serviceAccountDescription = "the Databasus service account (LocalSystem)"
)

func CheckServiceCanWriteFolder(folder string) error {
	serviceAccount, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return fmt.Errorf("build the LocalSystem SID: %w", err)
	}

	if err := checkIsLocalAbsolutePath(folder); err != nil {
		return err
	}

	cleanFolder := filepath.Clean(folder)

	if err := checkIsWritableFixedVolume(filepath.VolumeName(cleanFolder) + `\`); err != nil {
		return err
	}

	isFolderPresent, err := checkAccountRights(cleanFolder, serviceAccount, serviceAccountDescription)
	if err != nil {
		return err
	}

	if isFolderPresent {
		return probeWrite(cleanFolder)
	}

	return nil
}

func checkAccountRights(
	folder string,
	account *windows.SID,
	accountDescription string,
) (isFolderPresent bool, err error) {
	existingFolder, isFolderPresent, err := findNearestExistingFolder(folder)
	if err != nil {
		return false, err
	}

	requiredRights := uint32(requiredRightsOnAncestor)
	if isFolderPresent {
		requiredRights = requiredRightsOnExistingFolder
	}

	grantedRights, err := getGrantedRights(existingFolder, account)
	if err != nil {
		return false, fmt.Errorf("read the permissions of %s: %w", existingFolder, err)
	}

	if grantedRights&requiredRights == requiredRights {
		return isFolderPresent, nil
	}

	if isFolderPresent {
		return false, fmt.Errorf("%s cannot write to %s", accountDescription, existingFolder)
	}

	return false, fmt.Errorf("%s cannot create %s inside %s", accountDescription, folder, existingFolder)
}

func checkIsLocalAbsolutePath(folder string) error {
	if strings.HasPrefix(folder, `\\`) || strings.HasPrefix(folder, `//`) {
		return fmt.Errorf("%s is a network path; choose a folder on a local disk", folder)
	}

	volume := filepath.VolumeName(folder)
	if len(volume) != 2 || volume[1] != ':' || !filepath.IsAbs(folder) {
		return fmt.Errorf(`%q is not an absolute path such as D:\Databasus\backups`, folder)
	}

	return nil
}

func checkIsWritableFixedVolume(volumeRoot string) error {
	volumeRootPointer, err := windows.UTF16PtrFromString(volumeRoot)
	if err != nil {
		return err
	}

	switch windows.GetDriveType(volumeRootPointer) {
	case windows.DRIVE_FIXED:
	case windows.DRIVE_NO_ROOT_DIR, windows.DRIVE_UNKNOWN:
		return fmt.Errorf("drive %s does not exist", volumeRoot)
	default:
		return fmt.Errorf("drive %s is not a local fixed disk; choose a folder on a local disk", volumeRoot)
	}

	var fileSystemFlags uint32

	err = windows.GetVolumeInformation(volumeRootPointer, nil, 0, nil, nil, &fileSystemFlags, nil, 0)
	if err != nil {
		return fmt.Errorf("read volume information of %s: %w", volumeRoot, err)
	}

	if fileSystemFlags&windows.FILE_READ_ONLY_VOLUME != 0 {
		return fmt.Errorf("drive %s is read-only", volumeRoot)
	}

	return nil
}

func findNearestExistingFolder(folder string) (existingFolder string, isFolderPresent bool, err error) {
	current := folder

	for {
		info, statErr := os.Stat(current)

		switch {
		case statErr == nil && info.IsDir():
			return current, current == folder, nil
		case statErr == nil:
			return "", false, fmt.Errorf("%s is a file, not a folder", current)
		case !errors.Is(statErr, os.ErrNotExist):
			return "", false, fmt.Errorf("check %s: %w", current, statErr)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, fmt.Errorf("no part of %s exists", folder)
		}

		current = parent
	}
}

// The ACL check cannot see everything a real write would hit, such as a locked BitLocker volume,
// so an existing folder also gets a create-and-delete probe.
func probeWrite(folder string) error {
	probe, err := os.CreateTemp(folder, ".databasus-write-check-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w", folder, err)
	}

	probePath := probe.Name()

	return errors.Join(probe.Close(), os.Remove(probePath))
}
