//go:build windows

package winservice

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc"
)

func Test_StopRequest_WhileRunning_CancelsLauncherAndStopsCleanly(t *testing.T) {
	for _, command := range []svc.Cmd{svc.Stop, svc.Shutdown, svc.PreShutdown} {
		runner := newBlockingRunner(nil)
		session := startHandler(t, runner.run)

		session.waitForState(t, svc.Running)
		session.requests <- svc.ChangeRequest{Cmd: command}

		result := session.waitForResult(t)

		assert.False(t, result.isServiceSpecificExitCode)
		assert.Equal(t, uint32(0), result.exitCode)
		assert.True(t, runner.wasCancelled())
		assert.Contains(t, session.events.getInfoMessages(), "Databasus service stopped")
	}
}

func Test_ServiceRunning_AcceptsStopShutdownAndPreShutdown(t *testing.T) {
	session := startHandler(t, newBlockingRunner(nil).run)

	running := session.waitForState(t, svc.Running)

	assert.Equal(t, svc.AcceptStop|svc.AcceptShutdown|svc.AcceptPreShutdown, running.Accepts)

	session.requests <- svc.ChangeRequest{Cmd: svc.Stop}
	session.waitForResult(t)
}

func Test_LauncherFails_WithoutStopRequest_ReportsErrorExitAndEvent(t *testing.T) {
	session := startHandler(t, func(context.Context) error {
		return errors.New("the PostgreSQL client tools are missing")
	})

	result := session.waitForResult(t)

	assert.True(t, result.isServiceSpecificExitCode)
	assert.Equal(t, uint32(exitCodeStoppedOnItsOwn), result.exitCode)
	require.Len(t, session.events.getErrorMessages(), 1)
	assert.Contains(t, session.events.getErrorMessages()[0], "the PostgreSQL client tools are missing")
}

func Test_LauncherReturnsNil_WithoutStopRequest_ReportsErrorExit(t *testing.T) {
	session := startHandler(t, func(context.Context) error { return nil })

	result := session.waitForResult(t)

	assert.True(t, result.isServiceSpecificExitCode)
	assert.Equal(t, uint32(exitCodeStoppedOnItsOwn), result.exitCode)
}

func Test_StopRequest_WhenLauncherStopsWithError_ReportsErrorExitAndEvent(t *testing.T) {
	session := startHandler(t, newBlockingRunner(errors.New("pg_ctl stop failed")).run)

	session.waitForState(t, svc.Running)
	session.requests <- svc.ChangeRequest{Cmd: svc.Stop}

	result := session.waitForResult(t)

	assert.True(t, result.isServiceSpecificExitCode)
	assert.Equal(t, uint32(exitCodeStopFailed), result.exitCode)
	require.Len(t, session.events.getErrorMessages(), 1)
	assert.Contains(t, session.events.getErrorMessages()[0], "pg_ctl stop failed")
}

func Test_SlowStop_WhileWaiting_AdvancesStopPendingCheckpoint(t *testing.T) {
	releaseStop := make(chan struct{})
	session := startHandler(t, func(ctx context.Context) error {
		<-ctx.Done()
		<-releaseStop

		return nil
	})

	session.waitForState(t, svc.Running)
	session.requests <- svc.ChangeRequest{Cmd: svc.Stop}

	first := session.waitForState(t, svc.StopPending)
	second := session.waitForState(t, svc.StopPending)
	close(releaseStop)

	assert.Greater(t, second.CheckPoint, first.CheckPoint)
	assert.NotZero(t, first.WaitHint)
	session.waitForResult(t)
}

func Test_InterrogateRequest_ReturnsCurrentStatus(t *testing.T) {
	session := startHandler(t, newBlockingRunner(nil).run)

	running := session.waitForState(t, svc.Running)
	session.requests <- svc.ChangeRequest{Cmd: svc.Interrogate, CurrentStatus: running}

	assert.Equal(t, running, session.waitForState(t, svc.Running))

	session.requests <- svc.ChangeRequest{Cmd: svc.Stop}
	session.waitForResult(t)
}

type handlerResult struct {
	isServiceSpecificExitCode bool
	exitCode                  uint32
}

type handlerSession struct {
	requests chan svc.ChangeRequest
	statuses chan svc.Status
	results  chan handlerResult
	events   *recordingEventReporter
}

func startHandler(t *testing.T, run func(ctx context.Context) error) *handlerSession {
	t.Helper()

	session := &handlerSession{
		requests: make(chan svc.ChangeRequest),
		statuses: make(chan svc.Status, 100),
		results:  make(chan handlerResult, 1),
		events:   &recordingEventReporter{},
	}

	serviceHandler := &handler{
		run:                run,
		events:             session.events,
		stopWaitHint:       time.Minute,
		checkpointInterval: 50 * time.Millisecond,
	}

	go func() {
		isServiceSpecific, exitCode := serviceHandler.Execute(nil, session.requests, session.statuses)
		session.results <- handlerResult{isServiceSpecificExitCode: isServiceSpecific, exitCode: exitCode}
	}()

	return session
}

func (s *handlerSession) waitForState(t *testing.T, state svc.State) svc.Status {
	t.Helper()

	timeout := time.After(5 * time.Second)

	for {
		select {
		case status := <-s.statuses:
			if status.State == state {
				return status
			}
		case <-timeout:
			t.Fatalf("the handler never reported state %d", state)
		}
	}
}

func (s *handlerSession) waitForResult(t *testing.T) handlerResult {
	t.Helper()

	select {
	case result := <-s.results:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("the handler did not return")

		return handlerResult{}
	}
}

type blockingRunner struct {
	stopErr     error
	mutex       sync.Mutex
	isCancelled bool
}

func newBlockingRunner(stopErr error) *blockingRunner {
	return &blockingRunner{stopErr: stopErr}
}

func (r *blockingRunner) run(ctx context.Context) error {
	<-ctx.Done()

	r.mutex.Lock()
	r.isCancelled = true
	r.mutex.Unlock()

	return r.stopErr
}

func (r *blockingRunner) wasCancelled() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.isCancelled
}

type recordingEventReporter struct {
	mutex         sync.Mutex
	infoMessages  []string
	errorMessages []string
}

func (r *recordingEventReporter) Info(_ uint32, message string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.infoMessages = append(r.infoMessages, message)
}

func (r *recordingEventReporter) Error(_ uint32, message string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.errorMessages = append(r.errorMessages, message)
}

func (r *recordingEventReporter) getInfoMessages() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return append([]string(nil), r.infoMessages...)
}

func (r *recordingEventReporter) getErrorMessages() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return append([]string(nil), r.errorMessages...)
}
