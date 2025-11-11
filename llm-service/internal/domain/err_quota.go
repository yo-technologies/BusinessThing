package domain

import "errors"

// QuotaExceededError wraps a base error to indicate quota rejection
type QuotaExceededError error

var ErrGenerationStopped = errors.New("generation stopped by user")
