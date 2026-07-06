package timex

import (
	"testing"
	"time"
)

func TestRangeBoundaries(t *testing.T) {
	loc := time.FixedZone("test", 8*60*60)
	now := time.Date(2026, 7, 6, 12, 30, 0, 0, loc)

	day := Day(now)
	if !day.Start.Equal(time.Date(2026, 7, 6, 0, 0, 0, 0, loc)) {
		t.Fatalf("Day().Start = %v", day.Start)
	}
	if !day.End.Equal(time.Date(2026, 7, 7, 0, 0, 0, 0, loc)) {
		t.Fatalf("Day().End = %v", day.End)
	}
	if !EndOfDay(now).Equal(day.End.Add(-time.Nanosecond)) {
		t.Fatalf("EndOfDay() = %v", EndOfDay(now))
	}
}

func TestWeekUsesExplicitStart(t *testing.T) {
	tm := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	if got := StartOfWeek(tm, time.Monday); !got.Equal(time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("StartOfWeek(monday) = %v", got)
	}
	if got := StartOfWeek(tm, time.Sunday); !got.Equal(time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("StartOfWeek(sunday) = %v", got)
	}
}

func TestMonthAndYearBoundaries(t *testing.T) {
	loc := time.FixedZone("test", 8*60*60)
	tm := time.Date(2024, time.February, 29, 23, 59, 59, 123, loc)

	month := Month(tm)
	if !month.Start.Equal(time.Date(2024, time.February, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("Month().Start = %v", month.Start)
	}
	if !month.End.Equal(time.Date(2024, time.March, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("Month().End = %v", month.End)
	}
	monthEnd := time.Date(
		2024,
		time.February,
		29,
		23,
		59,
		59,
		int(time.Second-time.Nanosecond),
		loc,
	)
	if !EndOfMonth(tm).Equal(monthEnd) {
		t.Fatalf("EndOfMonth() = %v", EndOfMonth(tm))
	}

	year := Year(tm)
	if !year.Start.Equal(time.Date(2024, time.January, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("Year().Start = %v", year.Start)
	}
	if !year.End.Equal(time.Date(2025, time.January, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("Year().End = %v", year.End)
	}
	yearEnd := time.Date(
		2024,
		time.December,
		31,
		23,
		59,
		59,
		int(time.Second-time.Nanosecond),
		loc,
	)
	if !EndOfYear(tm).Equal(yearEnd) {
		t.Fatalf("EndOfYear() = %v", EndOfYear(tm))
	}
}

func TestWeekRangeAndEnd(t *testing.T) {
	tm := time.Date(2026, time.July, 8, 12, 0, 0, 0, time.UTC)
	week := Week(tm, time.Monday)
	if !week.Start.Equal(time.Date(2026, time.July, 6, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Week().Start = %v", week.Start)
	}
	if !week.End.Equal(time.Date(2026, time.July, 13, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("Week().End = %v", week.End)
	}
	if !EndOfWeek(tm, time.Monday).Equal(week.End.Add(-time.Nanosecond)) {
		t.Fatalf("EndOfWeek() = %v", EndOfWeek(tm, time.Monday))
	}
}
