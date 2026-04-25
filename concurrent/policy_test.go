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
	for i := 0; i < 10; i++ {
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
	isTemporary := func(err error) bool {
		return err != nil && err.Error() == "temporary"
	}
	policy := RetryOnCondition[int](isTemporary)
	action := policy(errors.New("temporary"), 123, 0)
	assert.Equal(t, ActionRetry, action)
	action = policy(errors.New("permanent"), 123, 0)
	assert.Equal(t, ActionContinue, action)
}

func TestAbortOnCondition(t *testing.T) {
	isCritical := func(err error) bool {
		return err != nil && err.Error() == "critical"
	}
	policy := AbortOnCondition[int](isCritical)
	action := policy(errors.New("critical"), 123, 0)
	assert.Equal(t, ActionAbort, action)
}

func TestCombinePolicies(t *testing.T) {
	t.Run("stops on non-continue", func(t *testing.T) {
		policy := CombinePolicies[int](
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