package concurrent

import (
	"context"
	"fmt"
	"time"
)

// Config controls executor concurrency, retries, error handling, and callbacks.
type Config[T any] struct {
	Name string

	Concurrency int

	Timeout time.Duration

	MaxRetry int

	Backoff BackoffFunc

	ErrorPolicy ErrorPolicy[T]

	PanicPolicy PanicPolicy[T]

	MaxErrorSamples int

	ErrorAggregation bool

	OnBegin func(ctx context.Context, total int)

	OnBefore func(ctx context.Context, item T, attempt int)

	OnAfter func(ctx context.Context, item T, err error, elapsed time.Duration)

	OnError func(ctx context.Context, item T, err error, attempt int)

	OnEnd func(ctx context.Context, result *Result)
}

func (c *Config[T]) Validate() error {
	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be > 0, got %d", c.Concurrency)
	}
	if c.MaxRetry < 0 {
		return fmt.Errorf("max retry must be >= 0, got %d", c.MaxRetry)
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be >= 0, got %v", c.Timeout)
	}
	return nil
}

func (c *Config[T]) SetDefaults() {
	if c.Name == "" {
		c.Name = "executor"
	}
	if c.ErrorPolicy == nil {
		c.ErrorPolicy = AlwaysContinue[T]()
	}
	if c.PanicPolicy == nil {
		c.PanicPolicy = PanicAsAbort[T]()
	}
	if c.MaxErrorSamples == 0 {
		c.MaxErrorSamples = 100
	}
}
