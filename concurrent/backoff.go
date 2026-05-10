package concurrent

import (
	"math"
	"time"
)

// ConstantBackoff returns the same delay for every retry attempt.
func ConstantBackoff(delay time.Duration) BackoffFunc {
	return func(_ int) time.Duration {
		return delay
	}
}

// LinearBackoff returns base multiplied by the retry attempt number.
func LinearBackoff(base time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return base * time.Duration(attempt)
	}
}

// ExponentialBackoff returns an exponential delay capped by maxDelay when maxDelay is positive.
func ExponentialBackoff(base time.Duration, maxDelay time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		if attempt <= 0 {
			return 0
		}

		if attempt > 62 {
			attempt = 62
		}
		multiplier := math.Pow(2, float64(attempt-1))
		delay := time.Duration(float64(base) * multiplier)
		if maxDelay > 0 && delay > maxDelay {
			return maxDelay
		}
		return delay
	}
}

// FibonacciBackoff returns a Fibonacci delay capped by maxDelay when maxDelay is positive.
func FibonacciBackoff(base time.Duration, maxDelay time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		if attempt <= 0 {
			return 0
		}

		if attempt > 92 {
			attempt = 92
		}
		fib := fibonacci(attempt)
		delay := base * time.Duration(fib)
		if maxDelay > 0 && delay > maxDelay {
			return maxDelay
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
