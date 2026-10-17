package availability

import "errors"

var (
	ErrEngineUnavailable  = errors.New("engine unavailable")
	ErrFeatureUnavailable = errors.New("feature unavailable")
)
