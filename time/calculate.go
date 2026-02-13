package time

import (
	"time"
)

// StartOfDay returns the start of day (00:00:00) for t.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
//	StartOfDay(t) // returns 2024-03-15 00:00:00
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of day (23:59:59) for t.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
//	EndOfDay(t) // returns 2024-03-15 23:59:59
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

// IsWeekend reports whether t falls on a Saturday or Sunday.
//
// Example:
//
//	IsWeekend(time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC)) // returns true (Saturday)
//	IsWeekend(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)) // returns false (Friday)
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday reports whether t falls on a weekday (Monday-Friday).
//
// Example:
//
//	IsWeekday(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)) // returns true (Friday)
//	IsWeekday(time.Date(2024, 3, 16, 0, 0, 0, 0, time.UTC)) // returns false (Saturday)
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

// AddBusinessDays adds n business days to t, skipping weekends.
// Negative n subtracts business days.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC) // Friday
//	AddBusinessDays(t, 1) // returns Monday 2024-03-18
func AddBusinessDays(t time.Time, days int) time.Time {
	current := t
	remain := days
	step := 1
	if days < 0 {
		step = -1
		remain = -days
	}
	for remain > 0 {
		current = current.AddDate(0, 0, step)
		if IsWeekday(current) {
			remain--
		}
	}
	return current
}

// BusinessDaysBetween returns the number of business days between start and end,
// excluding the start day.
//
// Example:
//
//	start := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC) // Friday
//	end := time.Date(2024, 3, 18, 0, 0, 0, 0, time.UTC)   // Monday
//	BusinessDaysBetween(start, end) // returns 1
func BusinessDaysBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	count := 0
	current := start
	for current.Before(end) {
		current = current.AddDate(0, 0, 1)
		if IsWeekday(current) {
			count++
		}
	}
	return count
}

// StartOfWeek returns the start of the week (Monday) for t.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC) // Friday
//	StartOfWeek(t) // returns Monday 2024-03-11 00:00:00
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
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC) // Friday
//	EndOfWeek(t) // returns Sunday 2024-03-17 23:59:59
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
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
//	StartOfMonth(t) // returns 2024-03-01 00:00:00
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month for t.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
//	EndOfMonth(t) // returns 2024-03-31 23:59:59
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(t.AddDate(0, 1, -1))
}

// StartOfYear returns the start of the year for t.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
//	StartOfYear(t) // returns 2024-01-01 00:00:00
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year for t.
//
// Example:
//
//	t := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
//	EndOfYear(t) // returns 2024-12-31 23:59:59
func EndOfYear(t time.Time) time.Time {
	return EndOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}

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
//	ParseInBeijing("2006-01-02 15:04", "2024-03-15 14:30") // returns 2024-03-15 14:30 +0800 CST
func ParseInBeijing(layout, value string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, value, location)
}
