//go:build windows

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"databasus-backend/internal/launcher/webtls"
	"databasus-backend/internal/util/securefile"
)

// The installer passes the PFX password in the environment, so it does not appear on a command
// line that other users of this computer can read.
const pfxPasswordEnvVariable = "DATABASUS_PFX_PASSWORD"

const (
	tlsSourceSelfSigned = "self-signed"
	tlsSourcePfx        = "pfx"
	tlsSourceExisting   = "existing"
)

// The installer logs the certificate summary this prints, or shows its one-line failure to the
// person running Setup.
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

	summary, err := runTLSSetup(*dataFolder, *source, *pfxFile, *isCheckOnly, time.Now().UTC())
	if err != nil {
		return failWith(os.Stdout, err)
	}

	_, _ = fmt.Fprintln(os.Stdout, summary.String())

	return 0
}

func runTLSSetup(dataFolder, source, pfxFile string, isCheckOnly bool, now time.Time) (webtls.Summary, error) {
	switch source {
	case tlsSourceSelfSigned:
		return webtls.GenerateSelfSigned(dataFolder, webtls.DiscoverHostNames(), now, securefile.ServiceOnly)
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

		return webtls.ImportPfx(dataFolder, pfxData, password, now, securefile.ServiceOnly)
	default:
		return webtls.Summary{}, fmt.Errorf(
			"-source must be %s, %s or %s",
			tlsSourceSelfSigned,
			tlsSourcePfx,
			tlsSourceExisting,
		)
	}
}
