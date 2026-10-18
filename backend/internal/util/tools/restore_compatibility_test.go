package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsRestoreServerOlderThanDumpClient_ByVersions_ReturnsExpected(t *testing.T) {
	versionCases := []struct {
		name          string
		clientMajor   int
		serverTarget  PostgresqlVersion
		isServerOlder bool
	}{
		{"server older than client", 18, PostgresqlVersion12, true},
		{"server one major older", 18, PostgresqlVersion17, true},
		{"server same as client", 18, PostgresqlVersion18, false},
		{"server newer than client", 16, PostgresqlVersion18, false},
		{"client version unknown", 0, PostgresqlVersion12, false},
		{"server version unparseable", 18, PostgresqlVersion("abc"), false},
	}

	for _, versionCase := range versionCases {
		t.Run(versionCase.name, func(t *testing.T) {
			assert.Equal(
				t,
				versionCase.isServerOlder,
				IsRestoreServerOlderThanDumpClient(versionCase.clientMajor, versionCase.serverTarget),
			)
		})
	}
}
