//go:build windows

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"databasus-backend/internal/launcher/webtls"
)

// The installer passes the PFX password in the environment, so it does not appear on a command
// line that other users of this computer can read.
const pfxPasswordEnvVariable = "DATABASUS_PFX_PASSWORD"

const (
	tlsSourceSelfSigned = "self-signed"
	tlsSourcePfx        = "pfx"
	tlsSourceExisting   = "existing"
)

// setUpTLS stores the HTTPS certificate the -https flag serves, or checks the stored one. It
// prints the certificate (subject, names, expiry, fingerprint) on success, and one line in words
// meant for the person running the installer on failure.
func setUpTLS(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher tls-setup", flag.ContinueOnError)

	dataFolder := flags.String("data-dir", getDefaultDataFolder(), "program data folder the service uses")
	source := flags.String("source", "", "self-signed, pfx or existing (check the stored certificate)")
	pfxFile := flags.String("pfx", "", "PFX file to import with -source pfx")
	isCheckOnly := flags.Bool(
		"check-only",
		false,
		"with -source pfx, check the file and password without storing anything",
	)

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	summary, err := runTLSSetup(*dataFolder, *source, *pfxFile, *isCheckOnly, time.Now())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, err)

		return 1
	}

	_, _ = fmt.Fprintln(os.Stdout, summary.String())

	return 0
}

func runTLSSetup(dataFolder, source, pfxFile string, isCheckOnly bool, now time.Time) (webtls.Summary, error) {
	switch source {
	case tlsSourceSelfSigned:
		return webtls.GenerateSelfSigned(dataFolder, webtls.DiscoverHostNames(), now)
	case tlsSourceExisting:
		return webtls.Check(dataFolder, now)
	case tlsSourcePfx:
		if pfxFile == "" {
			return webtls.Summary{}, errors.New("choose the PFX file to import")
		}

		pfxData, err := os.ReadFile(pfxFile)
		if err != nil {
			return webtls.Summary{}, fmt.Errorf("cannot read the PFX file: %w", err)
		}

		password := os.Getenv(pfxPasswordEnvVariable)

		if isCheckOnly {
			_, _, summary, err := webtls.ParsePfx(pfxData, password, now)

			return summary, err
		}

		return webtls.ImportPfx(dataFolder, pfxData, password, now)
	default:
		return webtls.Summary{}, fmt.Errorf(
			"-source must be %s, %s or %s",
			tlsSourceSelfSigned,
			tlsSourcePfx,
			tlsSourceExisting,
		)
	}
}
