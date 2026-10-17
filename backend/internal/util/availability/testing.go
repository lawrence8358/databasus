package availability

import (
	"slices"
	"testing"
)

// MarkUnavailableForTest overrides availability for the whole process until the test ends, so a
// test that calls it must not run in parallel with tests that create databases or use the marked
// features.
func MarkUnavailableForTest(t testing.TB, engines []Engine, features []Feature) {
	t.Helper()

	const reason = "disabled for this test"

	marked := GetAvailability()

	for i, engineAvailability := range marked.Engines {
		if slices.Contains(engines, engineAvailability.Engine) {
			marked.Engines[i] = EngineAvailability{
				Engine:            engineAvailability.Engine,
				UnavailableReason: reason,
			}
		}
	}

	for i, featureAvailability := range marked.Features {
		if slices.Contains(features, featureAvailability.Feature) {
			marked.Features[i] = FeatureAvailability{
				Feature:           featureAvailability.Feature,
				UnavailableReason: reason,
			}
		}
	}

	overriddenAvailability.Store(&marked)
	t.Cleanup(func() { overriddenAvailability.Store(nil) })
}
