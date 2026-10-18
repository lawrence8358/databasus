package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"databasus-backend/internal/util/stopevent"
)

// Under the Windows launcher the backend has no console to receive Ctrl+Break from, so the
// launcher asks it to stop through an inherited event instead.
func notifyOnShutdownRequest(log *slog.Logger, quit chan<- os.Signal) {
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	if err := stopevent.NotifyWhenParentRequestsStop(quit); err != nil {
		log.Warn("the launcher's stop request cannot be received; the backend will be killed on stop", "error", err)
	}
}
