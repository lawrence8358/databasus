//go:build !windows

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func notifyOnShutdownRequest(_ *slog.Logger, quit chan<- os.Signal) {
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
}
