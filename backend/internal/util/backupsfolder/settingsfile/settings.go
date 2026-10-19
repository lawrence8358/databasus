// Package settingsfile reads and writes the file in the program data folder that records the
// backups folder an administrator chose and the folders used before it. The Windows launcher, the
// installer (through the launcher) and the backend share it, so it is the one place that decides
// where backups go.
package settingsfile

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"databasus-backend/internal/util/securefile"
)

const (
	SettingsFileName = "backups-folder.txt"

	currentFolderKey  = "current"
	previousFolderKey = "previous"
	settingsHeader    = "# Written by Databasus. Change the backups folder in the Web UI " +
		"(Databasus settings) or by running Setup again.\n"
)

var ErrSettingsNotFound = errors.New("the backups folder settings file does not exist")

type Settings struct {
	CurrentFolder string
	// PreviousFolders hold backups written before the folder was changed, newest first.
	PreviousFolders []string
}

func GetSettingsFilePath(dataFolder string) string {
	return filepath.Join(dataFolder, SettingsFileName)
}

// WithCurrentFolder makes folder the current one and keeps the replaced folder as the first
// previous one, so the backups written there stay reachable.
func (s Settings) WithCurrentFolder(folder string) Settings {
	folder = filepath.Clean(folder)

	if IsSameFolder(s.CurrentFolder, folder) {
		return s
	}

	previousFolders := []string{}
	if s.CurrentFolder != "" {
		previousFolders = append(previousFolders, s.CurrentFolder)
	}

	for _, previousFolder := range s.PreviousFolders {
		isListed := slices.ContainsFunc(previousFolders, func(listedFolder string) bool {
			return IsSameFolder(listedFolder, previousFolder)
		})

		if !IsSameFolder(previousFolder, folder) && !isListed {
			previousFolders = append(previousFolders, previousFolder)
		}
	}

	return Settings{CurrentFolder: folder, PreviousFolders: previousFolders}
}

// GetFoldersHoldingBackups lists where a backup file may be, the current folder first.
func (s Settings) GetFoldersHoldingBackups() []string {
	return append([]string{s.CurrentFolder}, s.PreviousFolders...)
}

// ReadSettings returns ErrSettingsNotFound when the file does not exist. A file owned by anyone
// other than SYSTEM, the Administrators group or this process's own account is refused, because
// Databasus writes backups wherever it points, as LocalSystem on Windows.
func ReadSettings(path string) (Settings, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Settings{}, ErrSettingsNotFound
	}

	if err != nil {
		return Settings{}, fmt.Errorf("read %s: %w", path, err)
	}

	if err := checkIsOwnedByTrustedAccount(path); err != nil {
		return Settings{}, err
	}

	settings, err := parseSettings(string(content))
	if err != nil {
		return Settings{}, fmt.Errorf("%s: %w", path, err)
	}

	return settings, nil
}

// WriteSettings replaces the file with one rename, so a crash leaves either the old or the new
// settings, never half of them.
func WriteSettings(path string, settings Settings) error {
	for _, folder := range settings.GetFoldersHoldingBackups() {
		if err := checkIsUsableFolderPath(folder); err != nil {
			return err
		}
	}

	pendingPath := path + ".pending"

	if err := securefile.WriteFile(pendingPath, []byte(formatSettings(settings)), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", pendingPath, err)
	}

	if err := os.Rename(pendingPath, path); err != nil {
		_ = os.Remove(pendingPath)

		return fmt.Errorf("replace %s: %w", path, err)
	}

	return nil
}

// IsSameFolder compares two folder paths the way the file system does: case-insensitively on
// Windows, ignoring a trailing separator.
func IsSameFolder(firstFolder, secondFolder string) bool {
	if firstFolder == "" || secondFolder == "" {
		return firstFolder == secondFolder
	}

	return isSamePath(filepath.Clean(firstFolder), filepath.Clean(secondFolder))
}

func parseSettings(content string) (Settings, error) {
	settings := Settings{}
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, folder, isPair := strings.Cut(line, "=")
		if !isPair {
			return Settings{}, fmt.Errorf("unexpected line %q", line)
		}

		folder = strings.TrimSpace(folder)
		if err := checkIsUsableFolderPath(folder); err != nil {
			return Settings{}, err
		}

		switch strings.TrimSpace(key) {
		case currentFolderKey:
			if settings.CurrentFolder != "" {
				return Settings{}, errors.New("more than one current folder")
			}

			settings.CurrentFolder = filepath.Clean(folder)
		case previousFolderKey:
			settings.PreviousFolders = append(settings.PreviousFolders, filepath.Clean(folder))
		default:
			return Settings{}, fmt.Errorf("unknown key %q", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return Settings{}, err
	}

	if settings.CurrentFolder == "" {
		return Settings{}, errors.New("no current folder")
	}

	return settings, nil
}

func formatSettings(settings Settings) string {
	var builder strings.Builder

	builder.WriteString(settingsHeader)
	builder.WriteString(currentFolderKey + "=" + settings.CurrentFolder + "\n")

	for _, previousFolder := range settings.PreviousFolders {
		builder.WriteString(previousFolderKey + "=" + previousFolder + "\n")
	}

	return builder.String()
}

// The file holds one folder per line, which a line break inside a folder would split.
func checkIsUsableFolderPath(folder string) error {
	if folder == "" {
		return errors.New("the folder is empty")
	}

	if strings.ContainsAny(folder, "\r\n\x00") {
		return fmt.Errorf("the folder %q contains a line break or a NUL character", folder)
	}

	if !filepath.IsAbs(folder) {
		return fmt.Errorf("the folder %q is not an absolute path", folder)
	}

	return nil
}
