//go:build windows

package main

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BuildServiceArguments_WhenLanAccessOverHTTPS_KeepsEmptyHostAndHTTPS(t *testing.T) {
	config, err := parseRunFlags([]string{"-data-dir", t.TempDir(), "-listen-host", "", "-https", "-port", "4443"})
	require.NoError(t, err)

	arguments, err := buildServiceArguments(config)
	require.NoError(t, err)

	assert.Contains(t, arguments, "-https")
	for i, argument := range arguments {
		if argument == "-listen-host" {
			require.Less(t, i+1, len(arguments))
			assert.Empty(t, arguments[i+1])
		}
	}
	// The service command line must keep the empty host as its own argument.
	assert.Equal(t, `""`, syscall.EscapeArg(""))
}

func Test_ParseRunFlags_ByDefault_ListensOnLoopbackOverHTTP(t *testing.T) {
	config, err := parseRunFlags(nil)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", config.WebListenHost)
	assert.False(t, config.IsHTTPS)
}
