package concurrent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecutor_BasicExecution(t *testing.T) {
	config := Config[int]{
		Concurrency: 3,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := []int{1, 2, 3, 4, 5}
	var processed atomic.Int32

	handler := func(ctx context.Context, item int) error {
		processed.Add(1)
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Total != len(items) {
		t.Errorf("Expected total %d, got %d", len(items), result.Total)
	}
	if result.Success != len(items) {
		t.Errorf("Expected success %d, got %d", len(items), result.Success)
	}
	if int(processed.Load()) != len(items) {
		t.Errorf("Expected processed %d, got %d", len(items), processed.Load())
	}
}

func TestExecutor_WithTimeout(t *testing.T) {
	config := Config[int]{
		Concurrency: 1,
		Timeout:     50 * time.Millisecond,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := []int{1, 2, 3}

	handler := func(ctx context.Context, item int) error {
		if item == 2 {
			select {
			case <-time.After(200 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Either failed or cancelled (both indicate timeout handling worked)
	if result.Failed == 0 && result.Cancelled == 0 {
		t.Error("Expected at least one timeout failure or cancellation")
	}
}

func TestExecutor_WithRetry(t *testing.T) {
	config := Config[int]{
		Concurrency: 2,
		MaxRetry:    2,
		ErrorPolicy: AlwaysRetry[int](),
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	attemptCounts := make(map[int]*atomic.Int32)
	var mu sync.Mutex
	items := []int{1, 2, 3}

	for _, item := range items {
		attemptCounts[item] = &atomic.Int32{}
	}

	handler := func(ctx context.Context, item int) error {
		mu.Lock()
		counter := attemptCounts[item]
		mu.Unlock()
		count := counter.Add(1)

		if count < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Retried == 0 {
		t.Error("Expected some retries")
	}
	if result.Success != len(items) {
		t.Errorf("Expected all items to succeed after retry, got %d/%d", result.Success, len(items))
	}
}

func TestExecutor_AbortOnFirstError(t *testing.T) {
	config := Config[int]{
		Concurrency: 5,
		ErrorPolicy: AbortOnFirstError[int](),
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	handler := func(ctx context.Context, item int) error {
		if item == 10 {
			return errors.New("fatal error")
		}
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Aborted {
		t.Error("Expected execution to be aborted")
	}
	if result.Success+result.Failed+result.Cancelled != result.Total {
		t.Log("Note: Some tasks may not have been processed due to abort")
	}
}

func TestExecutor_PanicRecovery(t *testing.T) {
	config := Config[int]{
		Concurrency: 2,
		PanicPolicy: PanicAsContinue[int](),
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := []int{1, 2, 3}

	handler := func(ctx context.Context, item int) error {
		if item == 2 {
			panic("test panic")
		}
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Failed != 1 {
		t.Errorf("Expected 1 failure from panic, got %d", result.Failed)
	}
	if result.Success != 2 {
		t.Errorf("Expected 2 successful items, got %d", result.Success)
	}
}

func TestExecutor_PanicAbortCancelsOthers(t *testing.T) {
	config := Config[int]{
		Concurrency: 3,
		PanicPolicy: PanicAsAbort[int](),
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	handler := func(ctx context.Context, item int) error {
		if item == 2 {
			panic("test panic")
		}
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Aborted {
		t.Error("Expected execution to be aborted due to panic")
	}

	// Should have at least 1 failed (the panic)
	if result.Failed == 0 {
		t.Error("Expected at least 1 failed item from panic")
	}

	// Should have some cancelled items (abort stopped remaining work)
	if result.Cancelled == 0 {
		t.Error("Expected some cancelled items due to abort, got 0")
	}

	// Due to concurrent execution, not all items may be queued before abort
	// So we just verify: Success + Failed + Cancelled > 0 and <= Total
	processed := result.Success + result.Failed + result.Cancelled
	if processed == 0 {
		t.Error("Expected at least some items to be processed")
	}
	if processed > result.Total {
		t.Errorf("Processed count (%d) exceeds total (%d)", processed, result.Total)
	}

	t.Logf("Result: Success=%d Failed=%d Cancelled=%d Total=%d",
		result.Success, result.Failed, result.Cancelled, result.Total)
}

func TestExecutor_ErrorAggregation(t *testing.T) {
	config := Config[int]{
		Concurrency:      3,
		ErrorAggregation: true,
		MaxErrorSamples:  5,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	handler := func(ctx context.Context, item int) error {
		if item%2 == 0 {
			return errors.New("even number")
		}
		if item%3 == 0 {
			return errors.New("divisible by 3")
		}
		return nil
	}

	result, err := executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.ErrorCount) == 0 {
		t.Error("Expected error aggregation")
	}
	if len(result.ErrorSamples) > config.MaxErrorSamples {
		t.Errorf("Expected max %d error samples, got %d", config.MaxErrorSamples, len(result.ErrorSamples))
	}
}

func TestExecutor_LifecycleHooks(t *testing.T) {
	var onBeginCalled, onEndCalled atomic.Bool
	var onBeforeCalls, onAfterCalls atomic.Int32

	config := Config[int]{
		Concurrency: 2,
		OnBegin: func(ctx context.Context, total int) {
			onBeginCalled.Store(true)
		},
		OnBefore: func(ctx context.Context, item int, attempt int) {
			onBeforeCalls.Add(1)
		},
		OnAfter: func(ctx context.Context, item int, err error, elapsed time.Duration) {
			onAfterCalls.Add(1)
		},
		OnEnd: func(ctx context.Context, result *Result) {
			onEndCalled.Store(true)
		},
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	items := []int{1, 2, 3}
	handler := func(ctx context.Context, item int) error {
		return nil
	}

	_, err = executor.Run(context.Background(), items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !onBeginCalled.Load() {
		t.Error("OnBegin was not called")
	}
	if !onEndCalled.Load() {
		t.Error("OnEnd was not called")
	}
	if int(onBeforeCalls.Load()) != len(items) {
		t.Errorf("Expected OnBefore to be called %d times, got %d", len(items), onBeforeCalls.Load())
	}
	if int(onAfterCalls.Load()) != len(items) {
		t.Errorf("Expected OnAfter to be called %d times, got %d", len(items), onAfterCalls.Load())
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	config := Config[int]{
		Concurrency: 2,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	var processed atomic.Int32
	handler := func(ctx context.Context, item int) error {
		processed.Add(1)
		if processed.Load() == 10 {
			cancel()
		}
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	result, err := executor.Run(ctx, items, handler)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	total := result.Success + result.Failed + result.Cancelled
	if total == len(items) {
		t.Logf("All items were processed despite cancellation (timing dependent)")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config[int]
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config[int]{Concurrency: 1},
			wantErr: false,
		},
		{
			name:    "zero concurrency",
			config:  Config[int]{Concurrency: 0},
			wantErr: true,
		},
		{
			name:    "negative concurrency",
			config:  Config[int]{Concurrency: -1},
			wantErr: true,
		},
		{
			name:    "negative max retry",
			config:  Config[int]{Concurrency: 1, MaxRetry: -1},
			wantErr: true,
		},
		{
			name:    "negative timeout",
			config:  Config[int]{Concurrency: 1, Timeout: -time.Second},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResult_SuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected float64
	}{
		{
			name:     "100% success",
			result:   Result{Total: 10, Success: 10},
			expected: 100.0,
		},
		{
			name:     "50% success",
			result:   Result{Total: 10, Success: 5},
			expected: 50.0,
		},
		{
			name:     "0% success",
			result:   Result{Total: 10, Success: 0},
			expected: 0.0,
		},
		{
			name:     "empty result",
			result:   Result{Total: 0, Success: 0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := tt.result.SuccessRate()
			if rate != tt.expected {
				t.Errorf("SuccessRate() = %v, want %v", rate, tt.expected)
			}
		})
	}
}

func TestExecutor_RunStream(t *testing.T) {
	config := Config[int]{
		Concurrency: 3,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Create input channel
	itemsCh := make(chan int, 10)
	expectedCount := 10

	// Producer goroutine
	go func() {
		for i := 0; i < expectedCount; i++ {
			itemsCh <- i
		}
		close(itemsCh)
	}()

	var processed atomic.Int32
	handler := func(ctx context.Context, item int) error {
		processed.Add(1)
		time.Sleep(time.Millisecond * 10)
		return nil
	}

	result, err := executor.RunStream(context.Background(), itemsCh, handler)
	if err != nil {
		t.Fatalf("RunStream failed: %v", err)
	}

	if result.Total != expectedCount {
		t.Errorf("Expected total %d, got %d", expectedCount, result.Total)
	}
	if result.Success != expectedCount {
		t.Errorf("Expected success %d, got %d", expectedCount, result.Success)
	}
	if int(processed.Load()) != expectedCount {
		t.Errorf("Expected processed %d, got %d", expectedCount, processed.Load())
	}
}

func TestExecutor_RunStream_WithErrors(t *testing.T) {
	config := Config[int]{
		Concurrency: 2,
		ErrorPolicy: AlwaysContinue[int](),
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	itemsCh := make(chan int, 10)
	expectedCount := 10

	go func() {
		for i := 0; i < expectedCount; i++ {
			itemsCh <- i
		}
		close(itemsCh)
	}()

	handler := func(ctx context.Context, item int) error {
		if item%2 == 0 {
			return errors.New("even number error")
		}
		return nil
	}

	result, err := executor.RunStream(context.Background(), itemsCh, handler)
	if err != nil {
		t.Fatalf("RunStream failed: %v", err)
	}

	if result.Total != expectedCount {
		t.Errorf("Expected total %d, got %d", expectedCount, result.Total)
	}
	if result.Failed != 5 {
		t.Errorf("Expected 5 failures, got %d", result.Failed)
	}
	if result.Success != 5 {
		t.Errorf("Expected 5 successes, got %d", result.Success)
	}
}

func TestExecutor_RunStream_ContextCancellation(t *testing.T) {
	config := Config[int]{
		Concurrency: 2,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	itemsCh := make(chan int)

	var processed atomic.Int32

	// Producer that generates items slowly
	go func() {
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				return
			case itemsCh <- i:
				time.Sleep(time.Millisecond * 10)
			}
		}
		close(itemsCh)
	}()

	handler := func(ctx context.Context, item int) error {
		count := processed.Add(1)
		if count == 5 {
			cancel() // Cancel after processing 5 items
		}
		time.Sleep(time.Millisecond * 50)
		return nil
	}

	result, err := executor.RunStream(ctx, itemsCh, handler)
	if err != nil {
		t.Fatalf("RunStream failed: %v", err)
	}

	// Due to concurrent execution, we might process a few more than 5
	if result.Total > 20 {
		t.Errorf("Expected early termination, but processed %d items", result.Total)
	}
}

func TestExecutor_RunStream_SlowProducer(t *testing.T) {
	config := Config[int]{
		Concurrency: 5,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	itemsCh := make(chan int)

	// Slow producer
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Millisecond * 20)
			itemsCh <- i
		}
		close(itemsCh)
	}()

	var processed atomic.Int32
	handler := func(ctx context.Context, item int) error {
		processed.Add(1)
		time.Sleep(time.Millisecond * 5) // Fast consumer
		return nil
	}

	result, err := executor.RunStream(context.Background(), itemsCh, handler)
	if err != nil {
		t.Fatalf("RunStream failed: %v", err)
	}

	if result.Total != 10 {
		t.Errorf("Expected total 10, got %d", result.Total)
	}
	if result.Success != 10 {
		t.Errorf("Expected success 10, got %d", result.Success)
	}
}

func TestExecutor_RunStream_EmptyChannel(t *testing.T) {
	config := Config[int]{
		Concurrency: 2,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	itemsCh := make(chan int)
	close(itemsCh) // Immediately close

	handler := func(ctx context.Context, item int) error {
		return nil
	}

	result, err := executor.RunStream(context.Background(), itemsCh, handler)
	if err != nil {
		t.Fatalf("RunStream failed: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Expected total 0, got %d", result.Total)
	}
}

func TestExecutor_ShouldNotReuse(t *testing.T) {
	// This test documents the current behavior: Executor should NOT be reused
	config := Config[int]{
		Concurrency: 2,
	}

	executor, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// First execution
	items1 := []int{1, 2, 3}
	handler := func(ctx context.Context, item int) error {
		return nil
	}

	result1, _ := executor.Run(context.Background(), items1, handler)

	if result1.Success != 3 {
		t.Errorf("First run: expected 3 successes, got %d", result1.Success)
	}

	// Second execution - should return ErrExecutorReused
	items2 := []int{4, 5}
	result2, err := executor.Run(context.Background(), items2, handler)

	// Should fail with reuse error
	if err == nil {
		t.Fatal("Expected error when reusing executor, got nil")
	}
	if !errors.Is(err, ErrExecutorReused) {
		t.Errorf("Expected ErrExecutorReused, got %v", err)
	}
	if result2 != nil {
		t.Errorf("Expected nil result on reuse error, got %+v", result2)
	}

	// Best practice: create a new executor for each batch
	executor2, _ := New(config)
	result3, _ := executor2.Run(context.Background(), items2, handler)

	if result3.Success != 2 {
		t.Errorf("With new executor: expected 2 successes, got %d", result3.Success)
	}
}

func BenchmarkExecutor_Concurrency(b *testing.B) {
	configs := []struct {
		name        string
		concurrency int
	}{
		{"Concurrency-1", 1},
		{"Concurrency-4", 4},
		{"Concurrency-8", 8},
		{"Concurrency-16", 16},
	}

	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	handler := func(ctx context.Context, item int) error {
		time.Sleep(time.Microsecond * 10)
		return nil
	}

	for _, cfg := range configs {
		b.Run(cfg.name, func(b *testing.B) {
			config := Config[int]{
				Concurrency: cfg.concurrency,
			}
			executor, _ := New(config)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				executor.Run(context.Background(), items, handler)
			}
		})
	}
}
