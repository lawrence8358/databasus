//go:build windows

package winservice

import "time"

const (
	// Name is both the service name and the event log source.
	Name        = "Databasus"
	displayName = "Databasus"
	description = "Databasus backup service with its bundled metadata PostgreSQL."

	// Stopping waits up to 30 s for the backend and then for PostgreSQL's fast shutdown.
	stopWaitHint      = 2 * time.Minute
	preShutdownBudget = 2 * time.Minute

	recoveryResetPeriod = 24 * time.Hour

	eventIdStarted         = 1
	eventIdStopped         = 2
	eventIdStoppedOnItsOwn = 100
	eventIdStopFailed      = 101
)

// Exit codes the service reports as its service-specific exit code; any of them counts as a
// failure for the recovery actions.
const (
	exitCodeStoppedOnItsOwn = 1
	exitCodeStopFailed      = 2
)
