package concurrent

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicAsAbort(t *testing.T) {
	policy := PanicAsAbort[int]()
	action := policy("panic value", 123, 0)
	assert.Equal(t, ActionAbort, action)
}

func TestPanicAsContinue(t *testing.T) {
	policy := PanicAsContinue[int]()
	action := policy("panic value", 123, 0)
	assert.Equal(t, ActionContinue, action)
}

func TestAlwaysContinue(t *testing.T) {
	policy := AlwaysContinue[string]()
	action := policy(nil, "test", 0)
	assert.Equal(t, ActionContinue, action)
}

func TestAlwaysRetry(t *testing.T) {
	policy := AlwaysRetry[int]()
	action := policy(nil, 123, 0)
	assert.Equal(t, ActionRetry, action)
}

func TestRetryOnTimeout(t *testing.T) {
	policy := RetryOnTimeout[int]()
	action := policy(context.DeadlineExceeded, 123, 0)
	assert.Equal(t, ActionRetry, action)
	action = policy(errors.New("other"), 123, 0)
	assert.Equal(t, ActionContinue, action)
}

func TestAbortOnError(t *testing.T) {
	policy := AbortOnError[int]()
	action := policy(errors.New("error"), 123, 0)
	assert.Equal(t, ActionAbort, action)
}

func TestAbortOnFirstError(t *testing.T) {
	policy := AbortOnFirstError[int]()
	action := policy(errors.New("first"), 123, 0)
	assert.Equal(t, ActionAbort, action)
	action = policy(errors.New("second"), 456, 0)
	assert.Equal(t, ActionContinue, action)
}

func TestAbortOnFirstError_Concurrent(t *testing.T) {
	policy := AbortOnFirstError[int]()
	var wg sync.WaitGroup
	results := make([]ErrorAction, 10)
	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = policy(errors.New("error"), idx, 0)
		}(i)
	}
	wg.Wait()
	abortCount := 0
	for _, r := range results {
		if r == ActionAbort {
			abortCount++
		}
	}
	assert.Equal(t, 1, abortCount)
}

func TestRetryOnCondition(t *testing.T) {
	errTemporary := errors.New("temporary")
	errPermanent := errors.New("permanent")
	isTemporary := func(err error) bool {
		return errors.Is(err, errTemporary)
	}
	policy := RetryOnCondition[int](isTemporary)
	action := policy(errTemporary, 123, 0)
	assert.Equal(t, ActionRetry, action)
	action = policy(errPermanent, 123, 0)
	assert.Equal(t, ActionContinue, action)
}

func TestAbortOnCondition(t *testing.T) {
	errCritical := errors.New("critical")
	isCritical := func(err error) bool {
		return errors.Is(err, errCritical)
	}
	policy := AbortOnCondition[int](isCritical)
	action := policy(errCritical, 123, 0)
	assert.Equal(t, ActionAbort, action)
}

func TestCombinePolicies(t *testing.T) {
	t.Run("stops on non-continue", func(t *testing.T) {
		policy := CombinePolicies(
			AlwaysContinue[int](),
			AbortOnError[int](),
			AlwaysRetry[int](),
		)
		action := policy(errors.New("error"), 123, 0)
		assert.Equal(t, ActionAbort, action)
	})

	t.Run("empty", func(t *testing.T) {
		policy := CombinePolicies[int]()
		action := policy(errors.New("error"), 123, 0)
		assert.Equal(t, ActionContinue, action)
	})
}
