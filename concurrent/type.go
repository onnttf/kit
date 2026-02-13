package concurrent

import (
	"context"
	"time"
)

// Handler processes a single item. It returns an error if processing fails.
type Handler[T any] func(ctx context.Context, item T) error

// ErrorAction specifies the action to take when an error occurs.
type ErrorAction int

const (
	// ActionContinue continues to the next item.
	ActionContinue ErrorAction = iota
	// ActionRetry retries the current item.
	ActionRetry
	// ActionAbort stops all execution.
	ActionAbort
)

// String returns the string representation of ErrorAction.
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

// ErrorPolicy determines how to handle errors. It receives the error,
// the failed item, and the attempt number (0-based), and returns an
// ErrorAction indicating how to proceed.
type ErrorPolicy[T any] func(err error, item T, attempt int) ErrorAction

// PanicPolicy determines how to handle panics. It receives the panic
// value, the item being processed, and the attempt number (0-based),
// and returns an ErrorAction indicating how to proceed.
type PanicPolicy[T any] func(panicValue any, item T, attempt int) ErrorAction

// BackoffFunc returns the delay before the next retry. The attempt
// parameter is 1-based (first retry is attempt 1).
type BackoffFunc func(attempt int) time.Duration

// workItem represents a single unit of work.
type workItem[T any] struct {
	id      int // unique identifier
	data    T   // item to process
	attempt int // retry attempt (0-based)
}
