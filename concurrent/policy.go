package concurrent

import (
	"context"
	"errors"
	"sync"
)

func PanicAsAbort[T any]() PanicPolicy[T] {
	return func(_ any, _ T, _ int) ErrorAction {
		return ActionAbort
	}
}

func PanicAsContinue[T any]() PanicPolicy[T] {
	return func(_ any, _ T, _ int) ErrorAction {
		return ActionContinue
	}
}

func AlwaysContinue[T any]() ErrorPolicy[T] {
	return func(_ error, _ T, _ int) ErrorAction {
		return ActionContinue
	}
}

func AlwaysRetry[T any]() ErrorPolicy[T] {
	return func(_ error, _ T, _ int) ErrorAction {
		return ActionRetry
	}
}

// RetryOnTimeout retries errors wrapping context deadline or cancellation timeouts.
func RetryOnTimeout[T any]() ErrorPolicy[T] {
	return func(err error, _ T, _ int) ErrorAction {
		if errors.Is(err, context.DeadlineExceeded) {
			return ActionRetry
		}
		return ActionContinue
	}
}

func AbortOnError[T any]() ErrorPolicy[T] {
	return func(_ error, _ T, _ int) ErrorAction {
		return ActionAbort
	}
}

// AbortOnFirstError aborts only the first error seen by this policy instance.
func AbortOnFirstError[T any]() ErrorPolicy[T] {
	var once sync.Once
	return func(_ error, _ T, _ int) ErrorAction {
		var shouldAbort bool
		once.Do(func() {
			shouldAbort = true
		})
		if shouldAbort {
			return ActionAbort
		}
		return ActionContinue
	}
}

func RetryOnCondition[T any](shouldRetry func(error) bool) ErrorPolicy[T] {
	return func(err error, _ T, _ int) ErrorAction {
		if shouldRetry(err) {
			return ActionRetry
		}
		return ActionContinue
	}
}

func AbortOnCondition[T any](shouldAbort func(error) bool) ErrorPolicy[T] {
	return func(err error, _ T, _ int) ErrorAction {
		if shouldAbort(err) {
			return ActionAbort
		}
		return ActionContinue
	}
}

// CombinePolicies evaluates policies in order and returns the first non-continue action.
func CombinePolicies[T any](policies ...ErrorPolicy[T]) ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		for _, policy := range policies {
			if action := policy(err, item, attempt); action != ActionContinue {
				return action
			}
		}
		return ActionContinue
	}
}
