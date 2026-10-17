package availability

import (
	"sync/atomic"
	"testing"
)

const testOverrideReason = "disabled for this test"

type unavailableForTest struct {
	unavailableEngines  []Engine
	unavailableFeatures []Feature
}

var testOverride atomic.Pointer[unavailableForTest]

// MarkUnavailableForTest overrides availability for the whole process until the test ends, so a
// test that calls it must not run in parallel with tests that create databases or use the marked
// features.
func MarkUnavailableForTest(t testing.TB, engines []Engine, features []Feature) {
	t.Helper()

	testOverride.Store(&unavailableForTest{unavailableEngines: engines, unavailableFeatures: features})
	t.Cleanup(func() { testOverride.Store(nil) })
}
