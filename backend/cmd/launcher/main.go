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

const usage = `Usage:
  databasus-launcher.exe [flags]            run in this console until Ctrl+C
  databasus-launcher.exe install [flags]    register the Databasus service with these flags (administrator)
  databasus-launcher.exe uninstall          stop and remove the Databasus service (administrator)
  databasus-launcher.exe check-install -data-dir <dir> -backups-dir <dir> -port <port>
                                            check the installer's choices; prints one line per problem

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
