//go:build windows

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc/eventlog"
	"gopkg.in/natefinch/lumberjack.v2"

	"databasus-backend/internal/launcher"
	"databasus-backend/internal/launcher/winservice"
)

const (
	serviceLogFileName     = "launcher.log"
	serviceLogMaxSizeMB    = 10
	serviceLogMaxBackups   = 5
	serviceStartupFailedId = 102
)

func installService(arguments []string) int {
	config, err := parseRunFlags(arguments)
	if err != nil {
		return 2
	}

	serviceArguments, err := buildServiceArguments(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot locate the launcher executable: %v\n", err)

		return 1
	}

	err = winservice.Install(winservice.InstallSpec{ExecutablePath: executablePath, Arguments: serviceArguments})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	fmt.Printf("the %s service is installed; start it with: sc.exe start %s\n", winservice.Name, winservice.Name)

	return 0
}

func uninstallService() int {
	if err := winservice.Uninstall(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	fmt.Printf("the %s service is removed\n", winservice.Name)

	return 0
}

// A service has no console, so the launcher's own lines and the console output of the backend
// and the metadata server go to one rotated file under the data folder's logs folder.
func runAsService(arguments []string) int {
	config, err := parseRunFlags(arguments)
	if err != nil {
		reportStartupFailure(fmt.Sprintf("Databasus service has invalid arguments: %v", err))

		return 2
	}

	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(config.DataFolder, "logs", serviceLogFileName),
		MaxSize:    serviceLogMaxSizeMB,
		MaxBackups: serviceLogMaxBackups,
	}
	defer func() { _ = logFile.Close() }()

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	config.Logger = logger
	config.ChildOutput = logFile

	if err := launcher.BindChildProcessesToLauncherLifetime(); err != nil {
		logger.Error("could not tie child processes to the launcher", "error", err)
		reportStartupFailure(fmt.Sprintf("Databasus service could not start: %v", err))

		return 1
	}

	err = winservice.Run(logger, func(ctx context.Context) error {
		return launcher.Run(ctx, config)
	})
	if err != nil {
		logger.Error("the service control manager connection failed", "error", err)

		return 1
	}

	return 0
}

func reportStartupFailure(message string) {
	eventLog, err := eventlog.Open(winservice.Name)
	if err != nil {
		return
	}
	defer func() { _ = eventLog.Close() }()

	_ = eventLog.Error(serviceStartupFailedId, message)
}
