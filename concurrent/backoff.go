package concurrent

import (
	"math"
	"time"
)

// ConstantBackoff returns a BackoffFunc with constant delay.
//
// Example:
//
//	concurrent.ConstantBackoff(100 * time.Millisecond) // always returns 100ms
func ConstantBackoff(delay time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return delay
	}
}

// LinearBackoff returns a BackoffFunc with linear increase.
// delay = base * attempt
//
// Example:
//
//	concurrent.LinearBackoff(100 * time.Millisecond)(3) // returns 300ms
func LinearBackoff(base time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return base * time.Duration(attempt)
	}
}

// ExponentialBackoff returns a BackoffFunc with exponential increase.
// delay = base * (2 ^ (attempt - 1))
// The max parameter caps the maximum delay.
//
// Example:
//
//	concurrent.ExponentialBackoff(100*time.Millisecond, time.Second)(3) // returns 400ms
func ExponentialBackoff(base time.Duration, max time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		if attempt <= 0 {
			return 0
		}
		// Prevent overflow: cap at 62 to avoid float64 overflow (2^62 is safe)
		if attempt > 62 {
			attempt = 62
		}
		multiplier := math.Pow(2, float64(attempt-1))
		delay := time.Duration(float64(base) * multiplier)
		if max > 0 && delay > max {
			return max
		}
		return delay
	}
}

// FibonacciBackoff returns a BackoffFunc using Fibonacci sequence.
// delay = base * fibonacci(attempt)
// The max parameter caps the maximum delay.
//
// Example:
//
//	concurrent.FibonacciBackoff(100*time.Millisecond, time.Second)(5) // returns 500ms (fib(5)=5)
func FibonacciBackoff(base time.Duration, max time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		if attempt <= 0 {
			return 0
		}
		// Prevent overflow: Fibonacci grows quickly, cap at 92 (fib(92) < max int)
		if attempt > 92 {
			attempt = 92
		}
		fib := fibonacci(attempt)
		delay := base * time.Duration(fib)
		if max > 0 && delay > max {
			return max
		}
		return delay
	}
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}
