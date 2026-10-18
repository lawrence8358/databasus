//go:build windows

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"databasus-backend/internal/launcher/externaldb"
)

// The installer hands the connection string over in this environment variable, because a command
// line is readable by other users and ends up in logs.
const installerDsnEnvVariable = "DATABASUS_INSTALLER_EXTERNAL_DB_DSN"

// testExternalDb checks the connection string from the environment, or from -file when the
// installer reuses the one stored by an earlier install. Prints the problem and exits 1 on failure.
func testExternalDb(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher test-external-db", flag.ContinueOnError)
	file := flags.String("file", "", "read the connection string from this file instead of the environment")

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	dsn, err := readInstallerDsn(*file)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)

		return 1
	}

	if err := externaldb.TestConnection(context.Background(), dsn); err != nil {
		fmt.Fprintln(os.Stdout, err)

		return 1
	}

	return 0
}

// saveExternalDb stores the connection string from the environment (or -file) in the data folder, readable
// only by the service account and administrators.
func saveExternalDb(arguments []string) int {
	flags := flag.NewFlagSet("databasus-launcher save-external-db", flag.ContinueOnError)
	dataFolder := flags.String("data-dir", "", "program data folder the service will use")
	file := flags.String("file", "", "read the connection string from this file instead of the environment")

	if err := flags.Parse(arguments); err != nil {
		return 2
	}

	if *dataFolder == "" {
		fmt.Fprintln(os.Stderr, "-data-dir is required")

		return 2
	}

	dsn, err := readInstallerDsn(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	if err := externaldb.Save(*dataFolder, dsn); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
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
