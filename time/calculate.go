package time

import (
	"fmt"
	"sync"
	"time"
)

var (
	beijingLocOnce sync.Once
	beijingLoc     *time.Location
)

// ParseInLocation parses a time string in the specified location.
//
// Example:
//
//	loc, _ := time.LoadLocation("America/New_York")
//	ParseInLocation("2006-01-02 15:04", "2024-03-15 14:30", loc)
func ParseInLocation(layout, value string, location *time.Location) (time.Time, error) {
	return time.ParseInLocation(layout, value, location)
}

// ParseInBeijing parses a time string in Beijing timezone (UTC+8).
//
// Example:
//
//	ParseInBeijing("2006-01-02 15:04", "2024-03-15 14:30")
func ParseInBeijing(layout, value string) (time.Time, error) {
	beijingLocOnce.Do(func() {
		beijingLoc, _ = time.LoadLocation("Asia/Shanghai")
	})
	if beijingLoc == nil {
		return time.Time{}, fmt.Errorf("failed to load Asia/Shanghai timezone")
	}
	return time.ParseInLocation(layout, value, beijingLoc)
}

// StartOfDay returns the start of day (00:00:00) for t.
//
// Example:
//
//	StartOfDay(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of day (23:59:59.999999999) for t.
//
// Example:
//
//	EndOfDay(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday) for t.
//
// Example:
//
//	StartOfWeek(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -(weekday - 1)))
}

// EndOfWeek returns the end of the week (Sunday) for t.
//
// Example:
//
//	EndOfWeek(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return EndOfDay(t)
	}
	return EndOfDay(t.AddDate(0, 0, 7-weekday))
}

// StartOfMonth returns the start of the month for t.
//
// Example:
//
//	StartOfMonth(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month for t.
//
// Example:
//
//	EndOfMonth(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// StartOfYear returns the start of the year for t.
//
// Example:
//
//	StartOfYear(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year for t.
//
// Example:
//
//	EndOfYear(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC))
func EndOfYear(t time.Time) time.Time {
	return EndOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}

// IsWeekend reports whether t falls on a Saturday or Sunday.
//
// Example:
//
//	IsWeekend(time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC))
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday reports whether t falls on a weekday (Monday-Friday).
//
// Example:
//
//	IsWeekday(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}
