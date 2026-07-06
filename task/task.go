package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrPanic = errors.New("task: panic")

type Func[T any, R any] func(context.Context, T) (R, error)

type Result[T any, R any] struct {
	Items []Item[T, R]
}

type Item[T any, R any] struct {
	Index    int
	Input    T
	Output   R
	Canceled bool
	Err      error
}

type Option func(*config)

type config struct {
	workers int
}

func WithWorkers(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.workers = n
		}
	}
}

func Run[T any, R any](ctx context.Context, items []T, fn Func[T, R], opts ...Option) Result[T, R] {
	cfg := config{workers: 1}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.workers > len(items) && len(items) > 0 {
		cfg.workers = len(items)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	result := Result[T, R]{Items: make([]Item[T, R], len(items))}
	for i, input := range items {
		result.Items[i] = Item[T, R]{Index: i, Input: input}
	}
	if fn == nil {
		err := errors.New("task: function is nil")
		for i := range result.Items {
			result.Items[i].Err = err
		}
		return result
	}
	if err := ctx.Err(); err != nil {
		for i := range result.Items {
			result.Items[i].Canceled = true
			result.Items[i].Err = err
		}
		return result
	}
	if len(items) == 0 {
		return result
	}

	workCh := make(chan int)
	started := make([]bool, len(items))
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(workCh)
		for i := range items {
			select {
			case <-ctx.Done():
				return
			case workCh <- i:
			}
		}
	}()

	for range cfg.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workCh {
				started[i] = true
				out, err := runOne(ctx, items[i], fn)
				result.Items[i].Output = out
				result.Items[i].Err = err
			}
		}()
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		for i := range result.Items {
			if started[i] {
				continue
			}
			result.Items[i].Canceled = true
			result.Items[i].Err = err
		}
	}
	return result
}

func (r Result[T, R]) HasErrors() bool {
	for _, item := range r.Items {
		if item.Err != nil {
			return true
		}
	}
	return false
}

func runOne[T any, R any](ctx context.Context, input T, fn Func[T, R]) (output R, err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%w: value=%v", ErrPanic, v)
		}
	}()
	return fn(ctx, input)
}
