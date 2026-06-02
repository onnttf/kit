package concurrent

import "time"

// ConstantBackoff returns the same delay for every retry attempt.
func ConstantBackoff(delay time.Duration) BackoffFunc {
	return func(_ int) time.Duration {
		return delay
	}
}

// LinearBackoff returns base multiplied by the retry attempt number.
func LinearBackoff(base time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return cappedDelay(base, int64(attempt), 0)
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
		return cappedDelay(base, 1<<uint(attempt-1), maxDelay)
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
		return cappedDelay(base, int64(fibonacci(attempt)), maxDelay)
	}
}

func cappedDelay(base time.Duration, multiplier int64, maxDelay time.Duration) time.Duration {
	if base <= 0 || multiplier <= 0 {
		return 0
	}
	if maxDelay > 0 && base > maxDelay/time.Duration(multiplier) {
		return maxDelay
	}
	if base > time.Duration(1<<63-1)/time.Duration(multiplier) {
		if maxDelay > 0 {
			return maxDelay
		}
		return time.Duration(1<<63 - 1)
	}
	delay := base * time.Duration(multiplier)
	if maxDelay > 0 && delay > maxDelay {
		return maxDelay
	}
	return delay
}

func fibonacci(n int) int64 {
	if n <= 1 {
		return int64(n)
	}
	var a, b int64 = 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}
