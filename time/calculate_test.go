package time

import (
	"testing"
	"time"
)

// Test Suite for StartOfDay Function

func TestStartOfDay_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantHour int
		wantMin  int
		wantSec  int
		wantNano int
	}{
		{
			name:     "morning time",
			input:    time.Date(2024, 3, 15, 10, 30, 45, 123456789, time.UTC),
			wantHour: 0,
			wantMin:  0,
			wantSec:  0,
			wantNano: 0,
		},
		{
			name:     "midnight",
			input:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			wantHour: 0,
			wantMin:  0,
			wantSec:  0,
			wantNano: 0,
		},
		{
			name:     "end of day",
			input:    time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
			wantHour: 0,
			wantMin:  0,
			wantSec:  0,
			wantNano: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StartOfDay(tt.input)

			if got.Year() != tt.input.Year() {
				t.Errorf("Year: got %d, want %d", got.Year(), tt.input.Year())
			}
			if got.Month() != tt.input.Month() {
				t.Errorf("Month: got %d, want %d", got.Month(), tt.input.Month())
			}
			if got.Day() != tt.input.Day() {
				t.Errorf("Day: got %d, want %d", got.Day(), tt.input.Day())
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("Hour: got %d, want %d", got.Hour(), tt.wantHour)
			}
			if got.Minute() != tt.wantMin {
				t.Errorf("Minute: got %d, want %d", got.Minute(), tt.wantMin)
			}
			if got.Second() != tt.wantSec {
				t.Errorf("Second: got %d, want %d", got.Second(), tt.wantSec)
			}
			if got.Nanosecond() != tt.wantNano {
				t.Errorf("Nanosecond: got %d, want %d", got.Nanosecond(), tt.wantNano)
			}
		})
	}
}

func TestStartOfDay_Location(t *testing.T) {
	// Test with different timezones
	locations := []*time.Location{
		time.UTC,
		time.FixedZone("EST", -5*3600),
		time.FixedZone("JST", 9*3600),
	}

	for _, loc := range locations {
		t.Run(loc.String(), func(t *testing.T) {
			input := time.Date(2024, 3, 15, 14, 30, 45, 0, loc)
			got := StartOfDay(input)

			// Verify location is preserved
			if got.Location() != loc {
				t.Errorf("Location: got %v, want %v", got.Location(), loc)
			}

			// Verify it's start of day in that timezone
			if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
				t.Errorf("Expected 00:00:00, got %02d:%02d:%02d",
					got.Hour(), got.Minute(), got.Second())
			}
		})
	}
}

func TestStartOfDay_SpecialDates(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "leap year Feb 29",
			input: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
		},
		{
			name:  "year boundary",
			input: time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "month boundary",
			input: time.Date(2024, 3, 31, 15, 30, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StartOfDay(tt.input)

			if got.Year() != tt.input.Year() ||
				got.Month() != tt.input.Month() ||
				got.Day() != tt.input.Day() {
				t.Errorf("Date changed: got %v, want %v", got, tt.input)
			}

			if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
				t.Errorf("Not start of day: %v", got)
			}
		})
	}
}

// Test Suite for EndOfDay Function

func TestEndOfDay_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantHour int
		wantMin  int
		wantSec  int
		wantNano int
	}{
		{
			name:     "morning time",
			input:    time.Date(2024, 3, 15, 10, 30, 45, 123456789, time.UTC),
			wantHour: 23,
			wantMin:  59,
			wantSec:  59,
			wantNano: 999999999,
		},
		{
			name:     "midnight",
			input:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			wantHour: 23,
			wantMin:  59,
			wantSec:  59,
			wantNano: 999999999,
		},
		{
			name:     "already end of day",
			input:    time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
			wantHour: 23,
			wantMin:  59,
			wantSec:  59,
			wantNano: 999999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EndOfDay(tt.input)

			if got.Year() != tt.input.Year() {
				t.Errorf("Year: got %d, want %d", got.Year(), tt.input.Year())
			}
			if got.Month() != tt.input.Month() {
				t.Errorf("Month: got %d, want %d", got.Month(), tt.input.Month())
			}
			if got.Day() != tt.input.Day() {
				t.Errorf("Day: got %d, want %d", got.Day(), tt.input.Day())
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("Hour: got %d, want %d", got.Hour(), tt.wantHour)
			}
			if got.Minute() != tt.wantMin {
				t.Errorf("Minute: got %d, want %d", got.Minute(), tt.wantMin)
			}
			if got.Second() != tt.wantSec {
				t.Errorf("Second: got %d, want %d", got.Second(), tt.wantSec)
			}
			if got.Nanosecond() != tt.wantNano {
				t.Errorf("Nanosecond: got %d, want %d", got.Nanosecond(), tt.wantNano)
			}
		})
	}
}

func TestEndOfDay_Location(t *testing.T) {
	locations := []*time.Location{
		time.UTC,
		time.FixedZone("EST", -5*3600),
		time.FixedZone("JST", 9*3600),
	}

	for _, loc := range locations {
		t.Run(loc.String(), func(t *testing.T) {
			input := time.Date(2024, 3, 15, 14, 30, 45, 0, loc)
			got := EndOfDay(input)

			// Verify location is preserved
			if got.Location() != loc {
				t.Errorf("Location: got %v, want %v", got.Location(), loc)
			}

			// Verify it's end of day
			if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 || got.Nanosecond() != 999999999 {
				t.Errorf("Expected 23:59:59.999999999, got %02d:%02d:%02d.%d",
					got.Hour(), got.Minute(), got.Second(), got.Nanosecond())
			}
		})
	}
}

func TestEndOfDay_SpecialDates(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "leap year Feb 29",
			input: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
		},
		{
			name:  "year boundary",
			input: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "month boundary",
			input: time.Date(2024, 3, 31, 15, 30, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EndOfDay(tt.input)

			if got.Year() != tt.input.Year() ||
				got.Month() != tt.input.Month() ||
				got.Day() != tt.input.Day() {
				t.Errorf("Date changed: got %v, want %v", got, tt.input)
			}

			if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
				t.Errorf("Not end of day: %v", got)
			}
		})
	}
}

// Integration Tests

func TestStartAndEndOfDay_Range(t *testing.T) {
	input := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)

	start := StartOfDay(input)
	end := EndOfDay(input)

	// Verify start is before end
	if !start.Before(end) {
		t.Error("StartOfDay should be before EndOfDay")
	}

	// Verify both are on the same date
	if start.Year() != end.Year() || start.Month() != end.Month() || start.Day() != end.Day() {
		t.Error("StartOfDay and EndOfDay should be on the same date")
	}

	// Verify the duration
	duration := end.Sub(start)
	expectedDuration := 24*time.Hour - time.Nanosecond
	if duration != expectedDuration {
		t.Errorf("Duration between start and end: got %v, want %v", duration, expectedDuration)
	}
}

func TestStartOfDay_Idempotent(t *testing.T) {
	input := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	first := StartOfDay(input)
	second := StartOfDay(first)

	if !first.Equal(second) {
		t.Errorf("StartOfDay should be idempotent: first=%v, second=%v", first, second)
	}
}

func TestEndOfDay_Idempotent(t *testing.T) {
	input := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	first := EndOfDay(input)
	second := EndOfDay(first)

	if !first.Equal(second) {
		t.Errorf("EndOfDay should be idempotent: first=%v, second=%v", first, second)
	}
}

// Benchmark Tests

func BenchmarkStartOfDay(b *testing.B) {
	t := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StartOfDay(t)
	}
}

func BenchmarkEndOfDay(b *testing.B) {
	t := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EndOfDay(t)
	}
}

func BenchmarkStartAndEndOfDay(b *testing.B) {
	t := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StartOfDay(t)
		_ = EndOfDay(t)
	}
}
