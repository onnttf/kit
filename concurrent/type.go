package concurrent

import (
	"context"
	"time"
)

// Handler processes one item. It should honor ctx cancellation when doing
// blocking work.
type Handler[T any] func(ctx context.Context, item T) error

type ErrorAction int

const (
	// ActionContinue records the error and continues with the next item.
	ActionContinue ErrorAction = iota

	// ActionRetry retries the item until MaxRetry is reached.
	ActionRetry

	// ActionAbort stops the executor after the current error.
	ActionAbort
)

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

type ErrorPolicy[T any] func(err error, item T, attempt int) ErrorAction

type PanicPolicy[T any] func(panicValue any, item T, attempt int) ErrorAction

type BackoffFunc func(attempt int) time.Duration

type workItem[T any] struct {
	id      int
	data    T
	attempt int
}
