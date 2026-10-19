//go:build windows

package installcheck

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// An upgrade only swaps the server binaries, so a cluster of another major would make the service
// fail at start; the installer refuses before it stops the service or touches any file.
func CheckMetadataClusterMajor(dataFolder, bundledMajor string) error {
	versionFile := filepath.Join(dataFolder, "pgdata", "PG_VERSION")

	content, err := os.ReadFile(versionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("cannot read the version of the existing metadata database: %w", err)
	}

	clusterMajor := strings.TrimSpace(string(content))
	if clusterMajor == bundledMajor {
		return nil
	}

	return fmt.Errorf(
		"the metadata database in %s was created by PostgreSQL %s, but this version of Databasus "+
			"bundles PostgreSQL %s. Moving the data to another PostgreSQL major version needs pg_upgrade, "+
			"which Setup does not do. Nothing was changed",
		filepath.Join(dataFolder, "pgdata"), clusterMajor, bundledMajor,
	)
}
