//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"

	"databasus-backend/internal/launcher/installcheck"
	"databasus-backend/internal/util/tools"
)

// checkInstall is run by the installer before it installs anything. It prints one line per
// problem, in words meant for the person running the installer, and exits 1 if there is any.
func checkInstall(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher check-install", flag.ContinueOnError)

	dataFolder := flags.String("data-dir", "", "program data folder the service will use")
	backupsFolder := flags.String("backups-dir", "", "backups folder the service will use")
	port := flags.Int("port", 0, "Web UI port; 0 skips the port check")

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	var problems []string

	for _, folder := range []struct{ name, path string }{
		{"Program data folder", *dataFolder},
		{"Backups folder", *backupsFolder},
	} {
		if folder.path == "" {
			continue
		}

		if err := installcheck.CheckServiceCanWriteFolder(folder.path); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", folder.name, err))
		}
	}

	if *dataFolder != "" {
		bundledMajor := string(tools.GetMetadataPostgresqlVersion())
		if err := installcheck.CheckMetadataClusterMajor(*dataFolder, bundledMajor); err != nil {
			problems = append(problems, fmt.Sprintf("Program data folder: %v", err))
		}
	}

	if *port != 0 {
		if err := installcheck.CheckPortIsFree(*port); err != nil {
			problems = append(problems, fmt.Sprintf("Web UI port: %v", err))
		}
	}

	for _, problem := range problems {
		_, _ = fmt.Fprintln(os.Stdout, problem)
	}

	if len(problems) > 0 {
		return 1
	}

	return 0
}
