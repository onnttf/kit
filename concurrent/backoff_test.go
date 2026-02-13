package concurrent

import (
	"testing"
	"time"
)

func TestExponentialBackoff_LargeAttempt(t *testing.T) {
	backoff := ExponentialBackoff(time.Second, time.Minute)

	// Test with very large attempt (should not overflow)
	delay := backoff(100)
	if delay != time.Minute {
		t.Errorf("Expected max delay of %v, got %v", time.Minute, delay)
	}

	// Test with extremely large attempt
	delay = backoff(1000)
	if delay != time.Minute {
		t.Errorf("Expected max delay of %v, got %v", time.Minute, delay)
	}
}

func TestFibonacciBackoff_LargeAttempt(t *testing.T) {
	backoff := FibonacciBackoff(time.Millisecond, time.Minute)

	// Test with very large attempt (should not overflow)
	delay := backoff(100)
	if delay != time.Minute {
		t.Errorf("Expected max delay of %v, got %v", time.Minute, delay)
	}

	// Test with extremely large attempt
	delay = backoff(1000)
	if delay != time.Minute {
		t.Errorf("Expected max delay of %v, got %v", time.Minute, delay)
	}
}

func TestConstantBackoff(t *testing.T) {
	delay := time.Second
	backoff := ConstantBackoff(delay)

	for i := 1; i <= 10; i++ {
		if d := backoff(i); d != delay {
			t.Errorf("Expected constant delay %v, got %v", delay, d)
		}
	}
}

func TestLinearBackoff(t *testing.T) {
	base := time.Second
	backoff := LinearBackoff(base)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{5, 5 * time.Second},
		{10, 10 * time.Second},
	}

	for _, tt := range tests {
		if d := backoff(tt.attempt); d != tt.expected {
			t.Errorf("Attempt %d: expected %v, got %v", tt.attempt, tt.expected, d)
		}
	}
}

func TestExponentialBackoff_WithMax(t *testing.T) {
	base := 100 * time.Millisecond
	max := time.Second
	backoff := ExponentialBackoff(base, max)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 100 * time.Millisecond},  // 100ms * 2^0
		{2, 200 * time.Millisecond},  // 100ms * 2^1
		{3, 400 * time.Millisecond},  // 100ms * 2^2
		{4, 800 * time.Millisecond},  // 100ms * 2^3
		{5, time.Second},             // 100ms * 2^4 = 1.6s, capped at 1s
		{10, time.Second},            // Capped
	}

	for _, tt := range tests {
		d := backoff(tt.attempt)
		if d != tt.expected {
			t.Errorf("Attempt %d: expected %v, got %v", tt.attempt, tt.expected, d)
		}
	}
}

func TestFibonacciBackoff_WithMax(t *testing.T) {
	base := 10 * time.Millisecond
	max := 200 * time.Millisecond
	backoff := FibonacciBackoff(base, max)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 10 * time.Millisecond},   // 10ms * fib(1) = 10ms * 1
		{2, 10 * time.Millisecond},   // 10ms * fib(2) = 10ms * 1
		{3, 20 * time.Millisecond},   // 10ms * fib(3) = 10ms * 2
		{4, 30 * time.Millisecond},   // 10ms * fib(4) = 10ms * 3
		{5, 50 * time.Millisecond},   // 10ms * fib(5) = 10ms * 5
		{6, 80 * time.Millisecond},   // 10ms * fib(6) = 10ms * 8
		{7, 130 * time.Millisecond},  // 10ms * fib(7) = 10ms * 13
		{8, 200 * time.Millisecond},  // 10ms * fib(8) = 10ms * 21 = 210ms, capped
		{10, 200 * time.Millisecond}, // Capped
	}

	for _, tt := range tests {
		d := backoff(tt.attempt)
		if d != tt.expected {
			t.Errorf("Attempt %d: expected %v, got %v", tt.attempt, tt.expected, d)
		}
	}
}
