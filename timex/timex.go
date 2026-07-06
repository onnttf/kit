package timex

import "time"

type Range struct {
	Start time.Time
	End   time.Time
}

func Day(t time.Time) Range {
	start := StartOfDay(t)
	return Range{Start: start, End: start.AddDate(0, 0, 1)}
}

func Week(t time.Time, start time.Weekday) Range {
	s := StartOfWeek(t, start)
	return Range{Start: s, End: s.AddDate(0, 0, 7)}
}

func Month(t time.Time) Range {
	start := StartOfMonth(t)
	return Range{Start: start, End: start.AddDate(0, 1, 0)}
}

func Year(t time.Time) Range {
	start := StartOfYear(t)
	return Range{Start: start, End: start.AddDate(1, 0, 0)}
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return Day(t).End.Add(-time.Nanosecond)
}

func StartOfWeek(t time.Time, start time.Weekday) time.Time {
	days := (int(t.Weekday()) - int(start) + 7) % 7
	return StartOfDay(t.AddDate(0, 0, -days))
}

func EndOfWeek(t time.Time, start time.Weekday) time.Time {
	return Week(t, start).End.Add(-time.Nanosecond)
}

func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
	return Month(t).End.Add(-time.Nanosecond)
}

func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func EndOfYear(t time.Time) time.Time {
	return Year(t).End.Add(-time.Nanosecond)
}
