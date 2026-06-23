package concurrent

import "time"

type ErrorSample struct {
	Error     error
	TaskID    int
	Attempt   int
	Timestamp time.Time
}

type AbortReason struct {
	TaskID  int
	Attempt int
	Error   error
	Time    time.Time
}

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

func (r *Result) Duration() time.Duration {
	if r == nil || r.StartTime.IsZero() || r.EndTime.IsZero() {
		return 0
	}
	return r.EndTime.Sub(r.StartTime)
}

func (r *Result) HasErrors() bool {
	return r != nil && (r.Failed > 0 || r.Aborted)
}

func (r *Result) SuccessRate() float64 {
	if r == nil || r.Total == 0 {
		return 0
	}
	return float64(r.Success) / float64(r.Total) * 100
}

func (r *Result) IsComplete() bool {
	if r == nil {
		return false
	}
	return (r.Success + r.Failed + r.Cancelled) == r.Total
}
