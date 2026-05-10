package concurrent

import (
	"context"
	"time"
)

// Handler processes one item. It should honor ctx cancellation when doing
// blocking work.
type Handler[T any] func(ctx context.Context, item T) error

// ErrorAction controls what the executor does after a handler error.
type ErrorAction int

const (
	// ActionContinue records the error and continues with the next item.
	ActionContinue ErrorAction = iota

	// ActionRetry retries the item until MaxRetry is reached.
	ActionRetry

	// ActionAbort stops the executor after the current error.
	ActionAbort
)

// String returns a stable label for a retry action.
func (a ErrorAction) String() string {
	switch a {
	case ActionContinue:
		return "Continue"
	case ActionRetry:
		return "Retry"
	case ActionAbort:
		return "Abort"
	default:
		return "Unknown"
	}
}

// ErrorPolicy maps a handler error to the next executor action.
type ErrorPolicy[T any] func(err error, item T, attempt int) ErrorAction

// PanicPolicy maps a recovered panic to the next executor action.
type PanicPolicy[T any] func(panicValue any, item T, attempt int) ErrorAction

// BackoffFunc returns the delay before a retry attempt.
type BackoffFunc func(attempt int) time.Duration

type workItem[T any] struct {
	id      int
	data    T
	attempt int
}
