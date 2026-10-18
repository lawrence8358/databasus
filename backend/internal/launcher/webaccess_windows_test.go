package launcher

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"databasus-backend/internal/launcher/webtls"
	"databasus-backend/internal/util/securefile"
)

func Test_PrepareWebTLS_WhenHTTPIsOnLoopback_ServesPlainHTTP(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "localhost"} {
		files, err := prepareWebTLS(Config{DataFolder: t.TempDir(), WebListenHost: host}, slog.Default())

		require.NoError(t, err, host)
		assert.Nil(t, files, host)
	}
}

func Test_PrepareWebTLS_WhenHTTPWouldReachTheNetwork_Refuses(t *testing.T) {
	for _, host := range []string{"", "0.0.0.0", "::", "192.168.1.20"} {
		_, err := prepareWebTLS(Config{DataFolder: t.TempDir(), WebListenHost: host}, slog.Default())

		assert.ErrorContains(t, err, "only over HTTPS", host)
	}
}

func Test_PrepareWebTLS_WhenHTTPSHasNoCertificate_Refuses(t *testing.T) {
	_, err := prepareWebTLS(Config{DataFolder: t.TempDir(), WebListenHost: "", IsHTTPS: true}, slog.Default())

	assert.ErrorContains(t, err, "no HTTPS certificate")
}

func Test_PrepareWebTLS_WhenHTTPSHasCertificate_ReturnsItsFiles(t *testing.T) {
	dataFolder := t.TempDir()
	_, err := webtls.GenerateSelfSigned(dataFolder, webtls.HostNames{DNSNames: []string{"localhost"}}, time.Now().UTC(), securefile.CurrentAccount)
	require.NoError(t, err)

	files, err := prepareWebTLS(Config{DataFolder: dataFolder, WebListenHost: "", IsHTTPS: true}, slog.Default())

	require.NoError(t, err)
	assert.Equal(t, webtls.GetFiles(dataFolder), *files)
}
