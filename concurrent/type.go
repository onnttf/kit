package concurrent

import (
	"context"
	"time"
)

type Handler[T any] func(ctx context.Context, item T) error

type ErrorAction int

const (
	ActionContinue ErrorAction = iota

	ActionRetry

	ActionAbort
)

func (a ErrorAction) String() string {
	switch a {
	case ActionContinue:
		return "continue"
	case ActionRetry:
		return "retry"
	case ActionAbort:
		return "abort"
	default:
		return "unknown"
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
