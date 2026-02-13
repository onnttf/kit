package concurrent

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	workChannelBufferMultiplier = 2
)

var (
	// ErrExecutorReused is returned when attempting to reuse an Executor.
	ErrExecutorReused = errors.New("executor already used; create a new one")
)

type execCounters struct {
	success   atomic.Int64
	failed    atomic.Int64
	retried   atomic.Int64
	cancelled atomic.Int64
}

type errorCounter struct {
	count atomic.Int64
}

// An Executor runs tasks concurrently with retry and error handling.
type Executor[T any] struct {
	config Config[T]

	counters execCounters

	abortOnce sync.Once
	abortInfo atomic.Pointer[AbortReason]

	errorCounts sync.Map
	sampleMu    sync.Mutex
	samples     []ErrorSample

	used atomic.Bool
}

// New returns a new Executor with the given config.
//
// Each Executor should be used only once for Run or RunStream.
// To process multiple batches, create a new Executor for each.
func New[T any](config Config[T]) (*Executor[T], error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	config.SetDefaults()

	return &Executor[T]{config: config}, nil
}

// Run processes items concurrently and returns the result.
func (e *Executor[T]) Run(ctx context.Context, items []T, handler Handler[T]) (*Result, error) {
	if !e.used.CompareAndSwap(false, true) {
		return nil, ErrExecutorReused
	}

	start := time.Now()

	result := &Result{
		Total:     len(items),
		StartTime: start,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if e.config.OnBegin != nil {
		e.config.OnBegin(ctx, len(items))
	}

	if len(items) == 0 {
		result.EndTime = time.Now()
		return result, nil
	}

	workCh := make(chan workItem[T], e.config.Concurrency*workChannelBufferMultiplier)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(workCh)
		for i, item := range items {
			select {
			case <-ctx.Done():
				return
			case workCh <- workItem[T]{id: i, data: item}:
			}
		}
	}()

	for i := 0; i < e.config.Concurrency; i++ {
		wg.Add(1)
		go e.worker(ctx, workCh, handler, cancel, &wg)
	}

	wg.Wait()

	e.populateResult(ctx, result)
	return result, nil
}

// RunStream processes items from a channel concurrently.
// The input channel should be closed by the caller when done.
//
// This is useful for streaming data where the total count is unknown
// or to avoid loading all items into memory.
func (e *Executor[T]) RunStream(
	ctx context.Context,
	in <-chan T,
	handler Handler[T],
) (*Result, error) {
	if !e.used.CompareAndSwap(false, true) {
		return nil, ErrExecutorReused
	}

	start := time.Now()

	result := &Result{
		StartTime: start,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if e.config.OnBegin != nil {
		e.config.OnBegin(ctx, 0)
	}

	workCh := make(chan workItem[T], e.config.Concurrency*workChannelBufferMultiplier)
	var wg sync.WaitGroup
	var count atomic.Int64

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(workCh)

		id := 0
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-in:
				if !ok {
					return
				}
				count.Add(1)
				select {
				case <-ctx.Done():
					return
				case workCh <- workItem[T]{id: id, data: item}:
					id++
				}
			}
		}
	}()

	for i := 0; i < e.config.Concurrency; i++ {
		wg.Add(1)
		go e.worker(ctx, workCh, handler, cancel, &wg)
	}

	wg.Wait()

	result.Total = int(count.Load())

	e.populateResult(ctx, result)
	return result, nil
}

func (e *Executor[T]) populateResult(ctx context.Context, result *Result) {
	result.Success = int(e.counters.success.Load())
	result.Failed = int(e.counters.failed.Load())
	result.Retried = int(e.counters.retried.Load())
	result.Cancelled = int(e.counters.cancelled.Load())

	if info := e.abortInfo.Load(); info != nil {
		result.Aborted = true
		result.AbortReason = info
	}

	result.ErrorSamples = e.samples
	result.ErrorCount = make(map[string]int)

	e.errorCounts.Range(func(key, value any) bool {
		result.ErrorCount[key.(string)] = int(value.(*errorCounter).count.Load())
		return true
	})

	result.EndTime = time.Now()

	if e.config.OnEnd != nil {
		e.config.OnEnd(ctx, result)
	}
}

func (e *Executor[T]) worker(
	ctx context.Context,
	workCh <-chan workItem[T],
	handler Handler[T],
	cancel context.CancelFunc,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for item := range workCh {
		e.runWithRetry(ctx, item, handler, cancel)
	}
}

func (e *Executor[T]) runWithRetry(
	ctx context.Context,
	item workItem[T],
	handler Handler[T],
	cancel context.CancelFunc,
) {
	for {
		select {
		case <-ctx.Done():
			e.counters.cancelled.Add(1)
			return
		default:
		}

		start := time.Now()

		if e.config.OnBefore != nil {
			e.config.OnBefore(ctx, item.data, item.attempt)
		}

		err := e.execute(ctx, item, handler, cancel)

		elapsed := time.Since(start)

		if e.config.OnAfter != nil {
			e.config.OnAfter(ctx, item.data, err, elapsed)
		}

		if err == nil {
			e.counters.success.Add(1)
			return
		}

		if e.config.OnError != nil {
			e.config.OnError(ctx, item.data, err, item.attempt)
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			e.counters.cancelled.Add(1)
			return
		}

		e.recordError(item, err)

		action := e.config.ErrorPolicy(err, item.data, item.attempt)

		switch action {
		case ActionRetry:
			if item.attempt >= e.config.MaxRetry {
				e.counters.failed.Add(1)
				return
			}
			e.counters.retried.Add(1)
			item.attempt++

			if e.config.Backoff != nil {
				timer := time.NewTimer(e.config.Backoff(item.attempt))
				select {
				case <-timer.C:
				case <-ctx.Done():
					timer.Stop()
					return
				}
			}

		case ActionAbort:
			e.counters.failed.Add(1)
			e.abort(item, err)
			cancel()
			return

		default:
			e.counters.failed.Add(1)
			return
		}
	}
}

func (e *Executor[T]) execute(
	ctx context.Context,
	item workItem[T],
	handler Handler[T],
	ctxCancel context.CancelFunc,
) (err error) {
	taskCtx := ctx
	var taskCancel context.CancelFunc

	if e.config.Timeout > 0 {
		taskCtx, taskCancel = context.WithTimeout(ctx, e.config.Timeout)
		defer func() {
			taskCancel()
			if errors.Is(err, context.DeadlineExceeded) {
				err = fmt.Errorf("task timeout after %v: %w", e.config.Timeout, err)
			}
		}()
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic: %v\n%s", p, debug.Stack())
			if e.config.PanicPolicy(p, item.data, item.attempt) == ActionAbort {
				e.abort(item, err)
				ctxCancel()
			}
		}
	}()

	return handler(taskCtx, item.data)
}

func (e *Executor[T]) abort(item workItem[T], err error) {
	e.abortOnce.Do(func() {
		e.abortInfo.Store(&AbortReason{
			TaskID:  item.id,
			Attempt: item.attempt,
			Error:   err,
			Time:    time.Now(),
		})
	})
}

func (e *Executor[T]) recordError(item workItem[T], err error) {
	if e.config.ErrorAggregation {
		key := err.Error()
		v, _ := e.errorCounts.LoadOrStore(key, &errorCounter{})
		v.(*errorCounter).count.Add(1)
	}

	if e.config.MaxErrorSamples > 0 {
		e.sampleMu.Lock()
		if len(e.samples) < e.config.MaxErrorSamples {
			e.samples = append(e.samples, ErrorSample{
				Error:     err,
				TaskID:    item.id,
				Attempt:   item.attempt,
				Timestamp: time.Now(),
			})
		}
		e.sampleMu.Unlock()
	}
}
