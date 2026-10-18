//go:build windows

package installcheck

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CheckMetadataClusterMajor_WhenDataFolderHasNoCluster_ReturnsNil(t *testing.T) {
	assert.NoError(t, CheckMetadataClusterMajor(t.TempDir(), "18"))
	assert.NoError(t, CheckMetadataClusterMajor(filepath.Join(t.TempDir(), "missing"), "18"))
}

func Test_CheckMetadataClusterMajor_WhenClusterHasBundledMajor_ReturnsNil(t *testing.T) {
	dataFolder := newDataFolderWithCluster(t, "18\n")

	assert.NoError(t, CheckMetadataClusterMajor(dataFolder, "18"))
}

func Test_CheckMetadataClusterMajor_WhenClusterHasAnotherMajor_ReturnsErrorAndKeepsCluster(t *testing.T) {
	dataFolder := newDataFolderWithCluster(t, "17\n")

	err := CheckMetadataClusterMajor(dataFolder, "18")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "created by PostgreSQL 17")
	assert.Contains(t, err.Error(), "bundles PostgreSQL 18")
	assert.Contains(t, err.Error(), "pg_upgrade")

	content, readErr := os.ReadFile(filepath.Join(dataFolder, "pgdata", "PG_VERSION"))
	require.NoError(t, readErr)
	assert.Equal(t, "17\n", string(content))
}

func newDataFolderWithCluster(t *testing.T, pgVersion string) string {
	t.Helper()

	dataFolder := t.TempDir()
	clusterFolder := filepath.Join(dataFolder, "pgdata")
	require.NoError(t, os.MkdirAll(clusterFolder, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(clusterFolder, "PG_VERSION"), []byte(pgVersion), 0o600))

	return dataFolder
}
