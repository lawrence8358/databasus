//go:build windows

package winservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
)

type eventReporter interface {
	Info(eventId uint32, message string)
	Error(eventId uint32, message string)
}

type handler struct {
	run                func(ctx context.Context) error
	events             eventReporter
	stopWaitHint       time.Duration
	checkpointInterval time.Duration
}

// Execute reports Running as soon as the launcher starts, because first-time initialisation of
// the metadata database can outlast the 30 s the SCM allows a pending start.
func (h *handler) Execute(
	_ []string,
	requests <-chan svc.ChangeRequest,
	statuses chan<- svc.Status,
) (isServiceSpecificExitCode bool, exitCode uint32) {
	statuses <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runResult := make(chan error, 1)
	go func() { runResult <- h.run(ctx) }()

	statuses <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPreShutdown,
	}
	h.events.Info(eventIdStarted, "Databasus service started")

	for {
		select {
		case err := <-runResult:
			if err == nil {
				err = errors.New("the launcher returned without a stop request")
			}

			h.events.Error(eventIdStoppedOnItsOwn, fmt.Sprintf("Databasus stopped unexpectedly: %v", err))

			return true, exitCodeStoppedOnItsOwn
		case request := <-requests:
			switch request.Cmd {
			case svc.Interrogate:
				statuses <- request.CurrentStatus
			case svc.Stop, svc.Shutdown, svc.PreShutdown:
				cancel()

				if err := h.waitForStop(runResult, statuses); err != nil {
					h.events.Error(eventIdStopFailed, fmt.Sprintf("Databasus did not stop cleanly: %v", err))

					return true, exitCodeStopFailed
				}

				h.events.Info(eventIdStopped, "Databasus service stopped")

				return false, 0
			}
		}
	}
}

// The SCM treats a pending stop as hung once its wait hint passes without the checkpoint moving.
func (h *handler) waitForStop(runResult <-chan error, statuses chan<- svc.Status) error {
	checkpoint := uint32(1)
	statuses <- svc.Status{State: svc.StopPending, CheckPoint: checkpoint, WaitHint: uint32(h.stopWaitHint.Milliseconds())}

	ticker := time.NewTicker(h.checkpointInterval)
	defer ticker.Stop()

	for {
		select {
		case err := <-runResult:
			return err
		case <-ticker.C:
			checkpoint++
			statuses <- svc.Status{
				State:      svc.StopPending,
				CheckPoint: checkpoint,
				WaitHint:   uint32(h.stopWaitHint.Milliseconds()),
			}
		}
	}
}
