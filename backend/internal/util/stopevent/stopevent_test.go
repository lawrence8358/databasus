//go:build windows

package stopevent

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const childModeEnvVariable = "STOPEVENT_TEST_CHILD"

func TestMain(m *testing.M) {
	if os.Getenv(childModeEnvVariable) == "1" {
		os.Exit(runChildThatWaitsForStop())
	}

	os.Exit(m.Run())
}

func Test_SetEvent_WhenChildInheritedIt_ChildStopsGracefully(t *testing.T) {
	event, err := New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = event.Close() })

	testExecutable, err := os.Executable()
	require.NoError(t, err)

	var childOutput strings.Builder

	child := exec.Command(testExecutable)
	child.Env = append(os.Environ(), childModeEnvVariable+"=1", event.GetEnvironmentEntry())
	child.Stdout = &childOutput
	child.SysProcAttr = &syscall.SysProcAttr{AdditionalInheritedHandles: []syscall.Handle{event.GetHandle()}}
	require.NoError(t, child.Start())

	childExited := make(chan error, 1)
	go func() { childExited <- child.Wait() }()

	time.Sleep(500 * time.Millisecond)
	require.NoError(t, event.Set())

	select {
	case err := <-childExited:
		require.NoError(t, err)
		assert.Contains(t, childOutput.String(), "interrupted")
	case <-time.After(10 * time.Second):
		_ = child.Process.Kill()
		t.Fatal("the child did not stop after the event was set")
	}
}

func Test_NotifyWhenParentRequestsStop_WithoutEvent_SendsNothing(t *testing.T) {
	t.Setenv(handleEnvVariable, "")

	channel := make(chan os.Signal, 1)
	require.NoError(t, NotifyWhenParentRequestsStop(channel))

	select {
	case <-channel:
		t.Fatal("got a stop request without a stop event")
	case <-time.After(200 * time.Millisecond):
	}
}

func Test_NotifyWhenParentRequestsStop_WithInvalidHandle_ReturnsError(t *testing.T) {
	t.Setenv(handleEnvVariable, "not-a-handle")

	assert.Error(t, NotifyWhenParentRequestsStop(make(chan os.Signal, 1)))
}

func runChildThatWaitsForStop() int {
	channel := make(chan os.Signal, 1)
	if err := NotifyWhenParentRequestsStop(channel); err != nil {
		return 2
	}

	select {
	case <-channel:
		_, _ = os.Stdout.WriteString("interrupted")

		return 0
	case <-time.After(30 * time.Second):
		return 3
	}
}
