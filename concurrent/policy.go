package concurrent

import (
	"context"
	"errors"
	"sync"
)

// PanicAsAbort returns a PanicPolicy that aborts on any panic.
func PanicAsAbort[T any]() PanicPolicy[T] {
	return func(panicValue any, item T, attempt int) ErrorAction {
		return ActionAbort
	}
}

// PanicAsContinue returns a PanicPolicy that continues after panics.
func PanicAsContinue[T any]() PanicPolicy[T] {
	return func(panicValue any, item T, attempt int) ErrorAction {
		return ActionContinue
	}
}

// AlwaysContinue returns an ErrorPolicy that always continues.
// This is the default policy.
func AlwaysContinue[T any]() ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		return ActionContinue
	}
}

// AlwaysRetry returns an ErrorPolicy that always retries.
func AlwaysRetry[T any]() ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		return ActionRetry
	}
}

// RetryOnTimeout returns an ErrorPolicy that retries on timeout.
func RetryOnTimeout[T any]() ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		if errors.Is(err, context.DeadlineExceeded) {
			return ActionRetry
		}
		return ActionContinue
	}
}

// AbortOnError returns an ErrorPolicy that aborts on any error.
func AbortOnError[T any]() ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		return ActionAbort
	}
}

// AbortOnFirstError returns an ErrorPolicy that aborts on the first error.
// Due to concurrent execution, multiple errors may occur before abort takes effect.
func AbortOnFirstError[T any]() ErrorPolicy[T] {
	var once sync.Once
	return func(err error, item T, attempt int) ErrorAction {
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

// RetryOnCondition returns an ErrorPolicy that retries based on a condition.
func RetryOnCondition[T any](shouldRetry func(error) bool) ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		if shouldRetry(err) {
			return ActionRetry
		}
		return ActionContinue
	}
}

// AbortOnCondition returns an ErrorPolicy that aborts based on a condition.
func AbortOnCondition[T any](shouldAbort func(error) bool) ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		if shouldAbort(err) {
			return ActionAbort
		}
		return ActionContinue
	}
}

// CombineErrorPolicies returns an ErrorPolicy that combines multiple policies.
// It evaluates policies in order and returns the first non-Continue action.
func CombineErrorPolicies[T any](policies ...ErrorPolicy[T]) ErrorPolicy[T] {
	return func(err error, item T, attempt int) ErrorAction {
		for _, policy := range policies {
			if action := policy(err, item, attempt); action != ActionContinue {
				return action
			}
		}
		return ActionContinue
	}
}
