package launcher

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"databasus-backend/internal/launcher/webtls"
)

const certificateExpiryWarning = 30 * 24 * time.Hour

// prepareWebTLS returns the certificate files the backend serves HTTPS with, or nil for plain
// HTTP. Plain HTTP is refused on anything but loopback, so passwords never cross the network in
// clear text.
func prepareWebTLS(config Config, logger *slog.Logger) (*webtls.Files, error) {
	if !config.IsHTTPS {
		if !isLoopbackHost(config.WebListenHost) {
			return nil, fmt.Errorf(
				"the Web UI may listen on %q only over HTTPS; pass -https after storing a certificate with the tls-setup command",
				config.WebListenHost,
			)
		}

		return nil, nil
	}

	now := time.Now().UTC()

	summary, err := webtls.Check(config.DataFolder, now)
	if err != nil {
		return nil, fmt.Errorf("HTTPS is on but %w; store one with the tls-setup command", err)
	}

	logger.Info("the Web UI is served over HTTPS only", "certificate", summary.String())

	if summary.NotAfter.Sub(now) < certificateExpiryWarning {
		logger.Warn("the HTTPS certificate expires soon; replace it by running Setup again",
			"expires", summary.NotAfter.Format(time.DateOnly))
	}

	files := webtls.GetFiles(config.DataFolder)

	return &files, nil
}

// An empty host means every interface.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}
