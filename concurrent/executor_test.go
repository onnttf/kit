package concurrent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)
		assert.NotNil(t, exec)
	})

	t.Run("invalid config", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 0})
		assert.Error(t, err)
		assert.Nil(t, exec)
	})
}

func TestExecutor_Run(t *testing.T) {
	t.Run("empty items", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)
		result, err := exec.Run(context.Background(), []int{}, func(context.Context, int) error {
			return nil
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
	})

	t.Run("all success", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 4})
		require.NoError(t, err)
		items := []int{1, 2, 3, 4, 5}
		result, err := exec.Run(context.Background(), items, func(context.Context, int) error {
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 5, result.Total)
		assert.Equal(t, 5, result.Success)
	})

	t.Run("with failures", func(t *testing.T) {
		exec, err := New(Config[int]{
			Concurrency: 2,
			ErrorPolicy: func(error, int, int) ErrorAction {
				return ActionContinue
			},
			PanicPolicy: func(any, int, int) ErrorAction {
				return ActionContinue
			},
		})
		require.NoError(t, err)
		items := []int{1, 2, 3}
		result, err := exec.Run(context.Background(), items, func(_ context.Context, item int) error {
			if item == 2 {
				return errors.New("item 2 error")
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, result.Success)
		assert.Greater(t, result.Failed, 0)
	})

	t.Run("abort on error", func(t *testing.T) {
		exec, err := New(Config[int]{
			Concurrency: 2,
			ErrorPolicy: AbortOnError[int](),
		})
		require.NoError(t, err)
		items := []int{1, 2, 3, 4, 5}
		result, err := exec.Run(context.Background(), items, func(_ context.Context, item int) error {
			if item == 3 {
				return errors.New("abort now")
			}
			return nil
		})
		require.NoError(t, err)
		assert.True(t, result.Aborted)
	})

	t.Run("reuse error", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)
		_, err = exec.Run(context.Background(), []int{1}, func(context.Context, int) error {
			return nil
		})
		require.NoError(t, err)
		_, err = exec.Run(context.Background(), []int{2}, func(context.Context, int) error {
			return nil
		})
		assert.ErrorIs(t, err, ErrExecutorReused)
	})

	t.Run("nil context", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)

		result, err := exec.Run(context.Background(), []int{1}, func(ctx context.Context, _ int) error {
			assert.NotNil(t, ctx)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, 1, result.Success)
	})
}

func TestExecutor_RunStream(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)
		in := make(chan int)
		close(in)
		result, err := exec.RunStream(context.Background(), in, func(context.Context, int) error {
			return nil
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("all success", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 2})
		require.NoError(t, err)
		in := make(chan int, 5)
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
		result, err := exec.RunStream(context.Background(), in, func(context.Context, int) error {
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 5, result.Total)
		assert.Equal(t, 5, result.Success)
	})

	t.Run("nil context", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)
		in := make(chan int, 1)
		in <- 1
		close(in)

		result, err := exec.RunStream(context.Background(), in, func(ctx context.Context, _ int) error {
			assert.NotNil(t, ctx)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, 1, result.Success)
	})

	t.Run("cancelled before enqueue does not count total", func(t *testing.T) {
		exec, err := New(Config[int]{Concurrency: 1})
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := exec.RunStream(ctx, make(chan int), func(context.Context, int) error {
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 0, result.Total)
		assert.Equal(t, 0, result.Success)
	})
}

func TestExecutor_Panic(t *testing.T) {
	t.Run("panic as continue", func(t *testing.T) {
		exec, err := New(Config[int]{
			Concurrency: 2,
			PanicPolicy: PanicAsContinue[int](),
		})
		require.NoError(t, err)
		items := []int{1, 2, 3, 4}
		result, err := exec.Run(context.Background(), items, func(_ context.Context, item int) error {
			if item == 2 {
				panic("test panic")
			}
			return nil
		})
		require.NoError(t, err)
		assert.Greater(t, result.Success, 0)
	})

	t.Run("panic as abort", func(t *testing.T) {
		exec, err := New(Config[int]{
			Concurrency: 2,
			PanicPolicy: PanicAsAbort[int](),
		})
		require.NoError(t, err)
		items := []int{1, 2, 3, 4}
		result, err := exec.Run(context.Background(), items, func(_ context.Context, item int) error {
			if item == 2 {
				panic("test panic")
			}
			return nil
		})
		require.NoError(t, err)
		assert.True(t, result.Aborted)
	})
}

func TestExecutor_ConcurrentRuns(t *testing.T) {
	var totalSuccess int
	for range 5 {
		exec, err := New(Config[int]{Concurrency: 2})
		require.NoError(t, err)
		r, err := exec.Run(context.Background(), []int{1, 2, 3}, func(context.Context, int) error {
			return nil
		})
		require.NoError(t, err)
		totalSuccess += r.Success
	}
	assert.Equal(t, 15, totalSuccess)
}

func BenchmarkExecutor_Run(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec, _ := New(Config[int]{Concurrency: 10})
		_, _ = exec.Run(context.Background(), items, func(context.Context, int) error {
			return nil
		})
	}
}

func TestExecutor_Run_RetrySuccess(t *testing.T) {
	var attempts atomic.Int64
	exec, err := New(Config[int]{
		Concurrency: 1,
		MaxRetry:    2,
		Backoff:     ConstantBackoff(0),
		ErrorPolicy: AlwaysRetry[int](),
	})
	require.NoError(t, err)

	result, err := exec.Run(context.Background(), []int{1}, func(context.Context, int) error {
		if attempts.Add(1) == 1 {
			return errors.New("temporary")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Success)
	assert.Equal(t, 1, result.Retried)
	assert.Equal(t, int64(2), attempts.Load())
}

func TestExecutor_Run_ContextCancellationDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	exec, err := New(Config[int]{
		Concurrency: 1,
		MaxRetry:    5,
		Backoff:     ConstantBackoff(time.Minute),
		ErrorPolicy: AlwaysRetry[int](),
		OnError: func(context.Context, int, error, int) {
			cancel()
		},
	})
	require.NoError(t, err)

	result, err := exec.Run(ctx, []int{1}, func(context.Context, int) error {
		return errors.New("temporary")
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Retried)
	assert.True(t, result.IsComplete())
}
