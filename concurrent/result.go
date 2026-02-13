package concurrent

import "time"

// An ErrorSample records a single error occurrence.
type ErrorSample struct {
	Error     error     // the error
	TaskID    int       // ID of the failed task
	Attempt   int       // retry attempt when error occurred
	Timestamp time.Time // when the error occurred
}

// An AbortReason describes why execution was aborted.
type AbortReason struct {
	TaskID  int       // ID of the task that caused abort
	Attempt int       // retry attempt when abort occurred
	Error   error     // the error that triggered abort
	Time    time.Time // when abort was triggered
}

// A Result contains execution statistics.
type Result struct {
	Total     int // total items to process
	Success   int // successfully processed items
	Failed    int // failed items (after retries)
	Retried   int // total retry attempts
	Cancelled int // cancelled items

	Aborted     bool         // whether execution was aborted
	AbortReason *AbortReason // abort details

	StartTime time.Time // execution start time
	EndTime   time.Time // execution end time

	ErrorSamples []ErrorSample  // error samples
	ErrorCount   map[string]int // count per error message
}

// Duration returns the total execution duration.
func (r *Result) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// HasErrors reports whether there were any failures.
func (r *Result) HasErrors() bool {
	return r.Failed > 0 || r.Aborted
}

// SuccessRate returns the success rate as a percentage (0-100).
func (r *Result) SuccessRate() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Success) / float64(r.Total) * 100
}

// IsComplete reports whether all items were processed.
func (r *Result) IsComplete() bool {
	return (r.Success + r.Failed + r.Cancelled) == r.Total
}
