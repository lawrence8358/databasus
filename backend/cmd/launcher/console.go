//go:build windows

package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"databasus-backend/internal/launcher"
)

func runInConsole(arguments []string) int {
	config, err := parseRunFlags(arguments)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		return 2
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config.Logger = logger

	if err := launcher.BindChildProcessesToLauncherLifetime(); err != nil {
		logger.Error("could not tie child processes to the launcher", "error", err)

		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := launcher.Run(ctx, config); err != nil {
		logger.Error("Databasus stopped", "error", err)

		return 1
	}

	logger.Info("Databasus stopped")

	return 0
}
