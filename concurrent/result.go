package concurrent

import "time"

// ErrorSample records one sampled item error.
type ErrorSample struct {
	Error     error
	TaskID    int
	Attempt   int
	Timestamp time.Time
}

// AbortReason records the item that caused executor abortion.
type AbortReason struct {
	TaskID  int
	Attempt int
	Error   error
	Time    time.Time
}

// Result summarizes an executor run.
type Result struct {
	Total     int
	Success   int
	Failed    int
	Retried   int
	Cancelled int

	Aborted     bool
	AbortReason *AbortReason

	StartTime time.Time
	EndTime   time.Time

	ErrorSamples []ErrorSample
	ErrorCount   map[string]int
}

// Duration returns the elapsed time between StartTime and EndTime.
func (r *Result) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// HasErrors reports whether the run failed any item or aborted.
func (r *Result) HasErrors() bool {
	return r.Failed > 0 || r.Aborted
}

// SuccessRate returns the percentage of successful items.
func (r *Result) SuccessRate() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Success) / float64(r.Total) * 100
}

// IsComplete reports whether all queued items reached a terminal state.
func (r *Result) IsComplete() bool {
	return (r.Success + r.Failed + r.Cancelled) == r.Total
}
