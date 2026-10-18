package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetServerAddress_WhenNothingIsSet_ListensOnEveryInterfaceAt4005(t *testing.T) {
	t.Setenv(listenHostEnvVariable, "")
	t.Setenv(portEnvVariable, "")

	address, err := getServerAddress()

	require.NoError(t, err)
	assert.Equal(t, ":4005", address.getListenAddress())
	assert.Equal(t, "http://localhost:4005", address.getLocalBaseURL())
}

func Test_GetServerAddress_WhenHostAndPortAreSet_ListensOnlyThere(t *testing.T) {
	t.Setenv(listenHostEnvVariable, "127.0.0.1")
	t.Setenv(portEnvVariable, "8123")

	address, err := getServerAddress()

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8123", address.getListenAddress())
	assert.Equal(t, "http://127.0.0.1:8123", address.getLocalBaseURL())
}

func Test_GetServerAddress_WhenHostIsUnspecified_HealthcheckUsesLocalhost(t *testing.T) {
	t.Setenv(listenHostEnvVariable, "::")
	t.Setenv(portEnvVariable, "")

	address, err := getServerAddress()

	require.NoError(t, err)
	assert.Equal(t, "[::]:4005", address.getListenAddress())
	assert.Equal(t, "http://localhost:4005", address.getLocalBaseURL())
}

func Test_GetServerAddress_WhenPortIsInvalid_ReturnsError(t *testing.T) {
	for _, invalidPort := range []string{"abc", "0", "65536", "-1"} {
		t.Setenv(portEnvVariable, invalidPort)

		_, err := getServerAddress()

		assert.Error(t, err, invalidPort)
	}
}
