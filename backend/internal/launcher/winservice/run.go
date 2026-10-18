//go:build windows

package winservice

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

func IsRunningAsService() (bool, error) {
	return svc.IsWindowsService()
}

// Run blocks until the SCM stops the service. run gets a context cancelled on a stop, shutdown
// or pre-shutdown request and must return once everything it started has stopped.
func Run(logger *slog.Logger, run func(ctx context.Context) error) error {
	events := newEventReporter(logger)
	defer events.close()

	return svc.Run(Name, &handler{
		run:                run,
		events:             events,
		stopWaitHint:       stopWaitHint,
		checkpointInterval: 5 * time.Second,
	})
}

// The event log is a second copy of the service's own log lines that matter to an administrator
// looking at why the service is not running; failing to write it must not stop the service.
type loggingEventReporter struct {
	logger   *slog.Logger
	eventLog *eventlog.Log
}

func newEventReporter(logger *slog.Logger) *loggingEventReporter {
	eventLog, err := eventlog.Open(Name)
	if err != nil {
		logger.Warn("cannot write to the Windows event log", "error", err)
	}

	return &loggingEventReporter{logger: logger, eventLog: eventLog}
}

func (r *loggingEventReporter) Info(eventId uint32, message string) {
	r.logger.Info(message)

	if r.eventLog != nil {
		_ = r.eventLog.Info(eventId, message)
	}
}

func (r *loggingEventReporter) Error(eventId uint32, message string) {
	r.logger.Error(message)

	if r.eventLog != nil {
		_ = r.eventLog.Error(eventId, message)
	}
}

func (r *loggingEventReporter) close() {
	if r.eventLog != nil {
		_ = r.eventLog.Close()
	}
}
