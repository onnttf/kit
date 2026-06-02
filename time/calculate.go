package time

import "time"

// StartOfDay returns midnight at the start of t's day in t's location.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the final nanosecond of t's day in t's location.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the Monday of t's week in t's location.
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -(weekday - 1)))
}

// EndOfWeek returns the Sunday of t's week in t's location.
func EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return EndOfDay(t)
	}
	return EndOfDay(t.AddDate(0, 0, 7-weekday))
}

// StartOfMonth returns the first day of t's month at midnight in t's location.
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the final nanosecond of t's month in t's location.
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// StartOfYear returns January 1 of t's year at midnight in t's location.
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the final nanosecond of t's year in t's location.
func EndOfYear(t time.Time) time.Time {
	return EndOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}

func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}
