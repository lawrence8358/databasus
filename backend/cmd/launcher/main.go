//go:build windows

// Command launcher runs Databasus on Windows without Docker: it starts the bundled metadata
// PostgreSQL and the backend, and stops them in that reverse order. It runs in a console (stopped
// with Ctrl+C) or as the Databasus Windows service, which its install and uninstall commands
// register and remove.
package main

import (
	"fmt"
	"os"

	"databasus-backend/internal/launcher/winservice"
)

// Version is set at build time with -ldflags "-X main.Version=<version>".
var Version = ""

const usage = `Usage:
  databasus-launcher.exe [flags]            run in this console until Ctrl+C
  databasus-launcher.exe install [flags]    register the Databasus service with these flags (administrator)
  databasus-launcher.exe uninstall          stop and remove the Databasus service (administrator)
  databasus-launcher.exe check-install -data-dir <dir> -backups-dir <dir> -port <port>
                                            check the installer's choices; prints one line per problem
  databasus-launcher.exe tls-setup -data-dir <dir> -source self-signed|pfx|existing [-pfx <file>] [-check-only]
                                            store or check the HTTPS certificate for -https; the PFX
                                            password is read from the DATABASUS_PFX_PASSWORD variable

Flags:
`

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			return installService(os.Args[2:])
		case "uninstall":
			return uninstallService()
		case "check-install":
			return checkInstall(os.Args[2:])
		case "tls-setup":
			return setUpTLS(os.Args[2:])
		}
	}

	isService, err := winservice.IsRunningAsService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot tell whether this runs as a service: %v\n", err)

		return 1
	}

	if isService {
		return runAsService(os.Args[1:])
	}

	return runInConsole(os.Args[1:])
}
