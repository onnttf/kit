package concurrent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResult_Duration(t *testing.T) {
	start := time.Now()
	end := start.Add(5 * time.Second)
	result := &Result{
		StartTime: start,
		EndTime:   end,
	}
	assert.Equal(t, 5*time.Second, result.Duration())
}

func TestResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"no errors", Result{Failed: 0, Aborted: false}, false},
		{"has failed", Result{Failed: 1, Aborted: false}, true},
		{"aborted", Result{Failed: 0, Aborted: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.HasErrors())
		})
	}
}

func TestResult_SuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected float64
	}{
		{"zero total", Result{Total: 0}, 0},
		{"all success", Result{Total: 100, Success: 100}, 100},
		{"all failed", Result{Total: 100, Success: 0, Failed: 100}, 0},
		{"partial success", Result{Total: 100, Success: 75}, 75},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, tt.result.SuccessRate(), 0.01)
		})
	}
}

func TestResult_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"complete", Result{Total: 10, Success: 10, Failed: 0, Cancelled: 0}, true},
		{"incomplete", Result{Total: 10, Success: 5, Failed: 0, Cancelled: 0}, false},
		{"empty", Result{Total: 0, Success: 0, Failed: 0, Cancelled: 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsComplete())
		})
	}
}

func TestExecutor_Run_EmptyItemsInitializesErrorCount(t *testing.T) {
	exec, err := New(Config[int]{Concurrency: 1})
	require.NoError(t, err)

	result, err := exec.Run(context.Background(), nil, func(context.Context, int) error {
		return nil
	})

	require.NoError(t, err)
	assert.NotNil(t, result.ErrorCount)
	assert.Empty(t, result.ErrorCount)
}
