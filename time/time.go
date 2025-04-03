package time

import "time"

// NowInUTC returns the current time in UTC.
func NowInUTC() time.Time {
	return time.Now().UTC()
}
