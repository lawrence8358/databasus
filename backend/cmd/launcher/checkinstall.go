//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"

	"databasus-backend/internal/launcher/installcheck"
	"databasus-backend/internal/util/tools"
)

// The installer runs this before it installs anything and shows each printed line, so problems are
// worded for the person running Setup.
func checkInstall(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher check-install", flag.ContinueOnError)

	dataFolder := flags.String("data-dir", "", "program data folder the service will use")
	backupsFolder := flags.String("backups-dir", "", "backups folder the service will use")
	installFolder := flags.String("install-dir", "", "program folder; the backups folder must stay out of it")
	port := flags.Int("port", 0, "Web UI port; 0 skips the port check")
	isExternalMetadataDb := flags.Bool(
		"external-metadata-db",
		false,
		"the service will use an external metadata database, so the bundled cluster is not checked",
	)

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	var problems []string

	if *dataFolder != "" {
		if err := installcheck.CheckServiceCanWriteFolder(*dataFolder); err != nil {
			problems = append(problems, fmt.Sprintf("Program data folder: %v", err))
		}
	}

	if *backupsFolder != "" {
		err := installcheck.CheckBackupsFolder(*backupsFolder, installcheck.OwnFolders{
			DataFolder:    *dataFolder,
			InstallFolder: *installFolder,
		})
		if err != nil {
			problems = append(problems, fmt.Sprintf("Backups folder: %v", err))
		}
	}

	if *dataFolder != "" && !*isExternalMetadataDb {
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
