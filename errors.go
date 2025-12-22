package gsignal

import "errors"

// Exported sentinel errors.
// Can be compared with errors.Is.
var (
	ErrHubClosed        = errors.New("gsignal: hub is closed")
	ErrInvalidEvent     = errors.New("gsignal: invalid event")
	ErrInvalidSubscribe = errors.New("gsignal: invalid subscription")
)
