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

// Result summarizes an executor run.
//
// Total, Success, Failed, Retried, Cancelled, and Aborted are populated by Run
// or RunStream. ErrorSamples holds up to Config.MaxErrorSamples. ErrorCount is
// always non-nil: an empty map means no aggregated counts (e.g. when
// Config.ErrorAggregation is false). Use HasErrors to check whether any item
// failed or the run was aborted.
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
	return r.EndTime.Sub(r.StartTime)
}

func (r *Result) HasErrors() bool {
	return r.Failed > 0 || r.Aborted
}

func (r *Result) SuccessRate() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Success) / float64(r.Total) * 100
}

func (r *Result) IsComplete() bool {
	return (r.Success + r.Failed + r.Cancelled) == r.Total
}
