//go:build windows

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"databasus-backend/internal/launcher/externaldb"
	"databasus-backend/internal/util/securefile"
)

// The installer hands the connection string over in this environment variable, because a command
// line is readable by other users and ends up in logs.
const installerDsnEnvVariable = "DATABASUS_INSTALLER_EXTERNAL_DB_DSN"

// The -file form lets the installer reuse the string stored by an earlier install.
func verifyExternalDb(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher test-external-db", flag.ContinueOnError)
	file := flags.String("file", "", "read the connection string from this file instead of the environment")

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	dsn, err := readInstallerDsn(*file)
	if err != nil {
		return failWith(os.Stdout, err)
	}

	if err := externaldb.VerifyConnection(context.Background(), dsn); err != nil {
		return failWith(os.Stdout, err)
	}

	return 0
}

// The installer runs this as an administrator for the LocalSystem service, so the stored file leaves
// the installing account out of its ACL.
func saveExternalDb(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher save-external-db", flag.ContinueOnError)
	dataFolder := flags.String("data-dir", "", "program data folder the service will use")
	file := flags.String("file", "", "read the connection string from this file instead of the environment")

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	if *dataFolder == "" {
		_, _ = fmt.Fprintln(os.Stderr, "-data-dir is required")

		return 2
	}

	dsn, err := readInstallerDsn(*file)
	if err != nil {
		return failWith(os.Stderr, err)
	}

	if err := externaldb.Save(*dataFolder, dsn, securefile.ServiceOnly); err != nil {
		return failWith(os.Stderr, err)
	}

	fmt.Printf("the connection string is stored in %s\n", externaldb.GetFilePath(*dataFolder))

	return 0
}

func readInstallerDsn(file string) (string, error) {
	if file != "" {
		return externaldb.Load(file)
	}

	dsn := strings.TrimSpace(os.Getenv(installerDsnEnvVariable))
	if dsn == "" {
		return "", fmt.Errorf("the connection string is empty")
	}

	return dsn, nil
}
