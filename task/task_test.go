package task

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestRunOrderedResults(t *testing.T) {
	result := Run(context.Background(), []int{1, 2, 3}, func(_ context.Context, v int) (int, error) {
		return v * 2, nil
	}, WithWorkers(3))

	var got []int
	for _, item := range result.Items {
		got = append(got, item.Output)
		if item.Err != nil {
			t.Fatalf("item = %+v", item)
		}
	}
	if !reflect.DeepEqual(got, []int{2, 4, 6}) {
		t.Fatalf("outputs = %v", got)
	}
}

func TestRunCollectsErrorsAndPanics(t *testing.T) {
	result := Run(context.Background(), []int{1, 2, 3}, func(_ context.Context, v int) (int, error) {
		if v == 2 {
			return 0, errors.New("boom")
		}
		if v == 3 {
			panic("bad")
		}
		return v, nil
	}, WithWorkers(2))

	if !result.HasErrors() {
		t.Fatalf("HasErrors() = false")
	}
	if result.Items[1].Err == nil {
		t.Fatalf("expected item error")
	}
	if !errors.Is(result.Items[2].Err, ErrPanic) {
		t.Fatalf("panic error = %v", result.Items[2].Err)
	}
}

func TestRunMarksNotStartedItemsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := Run(ctx, []int{1, 2, 3}, func(_ context.Context, v int) (int, error) {
		return v, nil
	}, WithWorkers(2))

	for _, item := range result.Items {
		if !item.Canceled || !errors.Is(item.Err, context.Canceled) {
			t.Fatalf("item = %+v", item)
		}
	}
}

func TestRunEmptyNilFunctionAndNilOptions(t *testing.T) {
	result := Run[int, int](
		context.Background(),
		nil,
		func(_ context.Context, v int) (int, error) { return v, nil },
		nil,
	)
	if len(result.Items) != 0 || result.HasErrors() {
		t.Fatalf("Run(empty) = %+v", result)
	}

	result = Run[int, int](context.Background(), []int{1, 2}, nil)
	if !result.HasErrors() {
		t.Fatalf("Run(nil fn).HasErrors() = false")
	}
	for _, item := range result.Items {
		if item.Err == nil || item.Canceled {
			t.Fatalf("nil fn item = %+v", item)
		}
	}
}

func TestRunWorkerBoundsAndInputMetadata(t *testing.T) {
	var concurrent atomic.Int64
	var maxConcurrent atomic.Int64

	items := []int{10, 20}
	result := Run(context.Background(), items, func(_ context.Context, v int) (int, error) {
		cur := concurrent.Add(1)
		for {
			max := maxConcurrent.Load()
			if cur <= max || maxConcurrent.CompareAndSwap(max, cur) {
				break
			}
		}
		concurrent.Add(-1)
		return v + 1, nil
	}, WithWorkers(10))

	if result.HasErrors() {
		t.Fatalf("Run() errors = %+v", result.Items)
	}
	if maxConcurrent.Load() > int64(len(items)) {
		t.Fatalf("workers not capped, max concurrent = %d", maxConcurrent.Load())
	}
	for i, item := range result.Items {
		if item.Index != i || item.Input != items[i] || item.Output != items[i]+1 {
			t.Fatalf("item[%d] = %+v", i, item)
		}
	}
}

func TestRunMarksItemsCanceledAfterPartialWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls atomic.Int64
	result := Run(ctx, []int{1, 2, 3, 4, 5}, func(_ context.Context, v int) (int, error) {
		if calls.Add(1) == 1 {
			cancel()
		}
		return v, nil
	}, WithWorkers(1))

	if !result.HasErrors() {
		t.Fatalf("expected cancellation errors: %+v", result.Items)
	}
	var canceled int
	for _, item := range result.Items {
		if item.Canceled {
			canceled++
			if !errors.Is(item.Err, context.Canceled) {
				t.Fatalf("canceled item error = %v", item.Err)
			}
		}
	}
	if canceled == 0 {
		t.Fatalf("no items marked canceled: %+v", result.Items)
	}
}
