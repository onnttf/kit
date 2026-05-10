package concurrent

import (
	"context"
	"errors"
	"sync"
)

// PanicAsAbort returns a panic policy that aborts the executor.
func PanicAsAbort[T any]() PanicPolicy[T] {
	return func(_ any, _ T, _ int) ErrorAction {
		return ActionAbort
	}
}

// PanicAsContinue returns a panic policy that records the panic as an item error.
func PanicAsContinue[T any]() PanicPolicy[T] {
	return func(_ any, _ T, _ int) ErrorAction {
		return ActionContinue
	}
}

// AlwaysContinue returns an error policy that records errors and continues.
func AlwaysContinue[T any]() ErrorPolicy[T] {
	return func(_ error, _ T, _ int) ErrorAction {
		return ActionContinue
	}
}

// AlwaysRetry returns an error policy that retries every error.
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

// AbortOnError aborts the executor on every handler error.
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

// RetryOnCondition retries when shouldRetry returns true.
func RetryOnCondition[T any](shouldRetry func(error) bool) ErrorPolicy[T] {
	return func(err error, _ T, _ int) ErrorAction {
		if shouldRetry(err) {
			return ActionRetry
		}
		return ActionContinue
	}
}

// AbortOnCondition aborts when shouldAbort returns true.
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
