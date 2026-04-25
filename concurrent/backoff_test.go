package concurrent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstantBackoff(t *testing.T) {
	backoff := ConstantBackoff(100 * time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, backoff(1))
	assert.Equal(t, 100*time.Millisecond, backoff(5))
}

func TestLinearBackoff(t *testing.T) {
	backoff := LinearBackoff(50 * time.Millisecond)
	assert.Equal(t, 50*time.Millisecond, backoff(1))
	assert.Equal(t, 100*time.Millisecond, backoff(2))
	assert.Equal(t, 150*time.Millisecond, backoff(3))
}

func TestExponentialBackoff(t *testing.T) {
	backoff := ExponentialBackoff(time.Millisecond, 0)
	assert.Equal(t, time.Millisecond, backoff(1))
	assert.Equal(t, 2*time.Millisecond, backoff(2))
	assert.Equal(t, 4*time.Millisecond, backoff(3))
}

func TestExponentialBackoff_WithMax(t *testing.T) {
	backoff := ExponentialBackoff(time.Millisecond, 10*time.Millisecond)
	assert.Equal(t, time.Millisecond, backoff(1))
	assert.Equal(t, 2*time.Millisecond, backoff(2))
	assert.Equal(t, 8*time.Millisecond, backoff(4))
	assert.Equal(t, 10*time.Millisecond, backoff(5))
}

func TestExponentialBackoff_ZeroOrNegativeAttempt(t *testing.T) {
	backoff := ExponentialBackoff(time.Millisecond, 0)
	assert.Equal(t, time.Duration(0), backoff(0))
	assert.Equal(t, time.Duration(0), backoff(-1))
}

func TestFibonacciBackoff(t *testing.T) {
	backoff := FibonacciBackoff(time.Millisecond, 0)
	assert.Equal(t, 1*time.Millisecond, backoff(1))
	assert.Equal(t, 1*time.Millisecond, backoff(2))
	assert.Equal(t, 2*time.Millisecond, backoff(3))
	assert.Equal(t, 3*time.Millisecond, backoff(4))
}

func TestFibonacciBackoff_WithMax(t *testing.T) {
	backoff := FibonacciBackoff(time.Millisecond, 5*time.Millisecond)
	assert.Equal(t, 1*time.Millisecond, backoff(1))
	assert.Equal(t, 2*time.Millisecond, backoff(3))
	assert.Equal(t, 5*time.Millisecond, backoff(6))
	assert.Equal(t, 5*time.Millisecond, backoff(100))
}

func TestFibonacci(t *testing.T) {
	tests := []struct {
		n      int
		expect int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
		{6, 8},
		{7, 13},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, fibonacci(tt.n))
	}
}