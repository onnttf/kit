package concurrent

import (
	"context"
	"fmt"
	"time"
)

// A Config specifies the executor configuration.
type Config[T any] struct {
	// Name identifies this executor.
	Name string

	// Concurrency is the number of concurrent workers.
	// It must be greater than 0.
	Concurrency int

	// Timeout is the per-task timeout.
	// Zero means no timeout.
	Timeout time.Duration

	// MaxRetry is the maximum number of retry attempts.
	// Zero means no retries.
	MaxRetry int

	// Backoff returns the delay before each retry.
	Backoff BackoffFunc

	// ErrorPolicy determines how to handle errors.
	// If nil, defaults to AlwaysContinue.
	ErrorPolicy ErrorPolicy[T]

	// PanicPolicy determines how to handle panics.
	// If nil, defaults to PanicAsAbort.
	PanicPolicy PanicPolicy[T]

	// MaxErrorSamples limits error samples collected.
	// If zero, defaults to 100.
	MaxErrorSamples int

	// ErrorAggregation enables error grouping by message.
	ErrorAggregation bool

	// OnBegin is called when execution begins.
	OnBegin func(ctx context.Context, total int)

	// OnBefore is called before processing each item.
	OnBefore func(ctx context.Context, item T, attempt int)

	// OnAfter is called after processing each item.
	OnAfter func(ctx context.Context, item T, err error, elapsed time.Duration)

	// OnError is called when an error occurs.
	OnError func(ctx context.Context, item T, err error, attempt int)

	// OnEnd is called when execution completes.
	OnEnd func(ctx context.Context, result *Result)
}

// Validate checks whether the configuration is valid.
func (c *Config[T]) Validate() error {
	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be > 0, got %d", c.Concurrency)
	}
	if c.MaxRetry < 0 {
		return fmt.Errorf("maxRetry must be >= 0, got %d", c.MaxRetry)
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be >= 0, got %v", c.Timeout)
	}
	return nil
}

// SetDefaults sets default values for unset fields.
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
