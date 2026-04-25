package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInLocation(t *testing.T) {
	loc := time.UTC
	result, err := ParseInLocation("2006-01-02", "2024-03-15", loc)
	require.NoError(t, err)
	assert.Equal(t, 2024, result.Year())
	assert.Equal(t, time.March, result.Month())
	assert.Equal(t, 15, result.Day())
}

func TestStartOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "normal case",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "already start",
			input:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "end of day",
			input:    time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
			expected: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfDay(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "normal case",
			input:    time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC),
			expected: time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "already end",
			input:    time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
			expected: time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "start of day",
			input:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 3, 15, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfDay(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartOfWeek(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "monday",
			input:    time.Date(2024, 3, 11, 14, 0, 0, 0, loc), // Monday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
		{
			name:     "tuesday",
			input:    time.Date(2024, 3, 12, 14, 0, 0, 0, loc), // Tuesday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
		{
			name:     "wednesday",
			input:    time.Date(2024, 3, 13, 14, 0, 0, 0, loc), // Wednesday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
		{
			name:     "thursday",
			input:    time.Date(2024, 3, 14, 14, 0, 0, 0, loc), // Thursday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
		{
			name:     "friday",
			input:    time.Date(2024, 3, 15, 14, 0, 0, 0, loc), // Friday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
		{
			name:     "saturday",
			input:    time.Date(2024, 3, 16, 14, 0, 0, 0, loc), // Saturday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
		{
			name:     "sunday",
			input:    time.Date(2024, 3, 17, 14, 0, 0, 0, loc), // Sunday
			expected: time.Date(2024, 3, 11, 0, 0, 0, 0, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeek(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndOfWeek(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "monday",
			input:    time.Date(2024, 3, 11, 14, 0, 0, 0, loc), // Monday
			expected: time.Date(2024, 3, 17, 23, 59, 59, 999999999, loc),
		},
		{
			name:     "sunday",
			input:    time.Date(2024, 3, 17, 14, 0, 0, 0, loc), // Sunday
			expected: time.Date(2024, 3, 17, 23, 59, 59, 999999999, loc),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfWeek(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartOfMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "middle of month",
			input:    time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "first of month",
			input:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "last of month",
			input:    time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfMonth(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndOfMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "march 31 days",
			input:    time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 3, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "april 30 days",
			input:    time.Date(2024, 4, 15, 14, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 4, 30, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "february leap year",
			input:    time.Date(2024, 2, 15, 14, 0, 0, 0, time.UTC), // 2024 is leap year
			expected: time.Date(2024, 2, 29, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "february non-leap year",
			input:    time.Date(2023, 2, 15, 14, 0, 0, 0, time.UTC), // 2023 is not leap year
			expected: time.Date(2023, 2, 28, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfMonth(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartOfYear(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "middle of year",
			input:    time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "start of year",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "end of year",
			input:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfYear(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEndOfYear(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "middle of year",
			input:    time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "start of year",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "end of year",
			input:    time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
			expected: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfYear(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected bool
	}{
		{"monday", time.Date(2024, 3, 11, 12, 0, 0, 0, time.UTC), false},
		{"tuesday", time.Date(2024, 3, 12, 12, 0, 0, 0, time.UTC), false},
		{"wednesday", time.Date(2024, 3, 13, 12, 0, 0, 0, time.UTC), false},
		{"thursday", time.Date(2024, 3, 14, 12, 0, 0, 0, time.UTC), false},
		{"friday", time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC), false},
		{"saturday", time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC), true},
		{"sunday", time.Date(2024, 3, 17, 12, 0, 0, 0, time.UTC), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeekend(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWeekday(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected bool
	}{
		{"monday", time.Date(2024, 3, 11, 12, 0, 0, 0, time.UTC), true},
		{"tuesday", time.Date(2024, 3, 12, 12, 0, 0, 0, time.UTC), true},
		{"wednesday", time.Date(2024, 3, 13, 12, 0, 0, 0, time.UTC), true},
		{"thursday", time.Date(2024, 3, 14, 12, 0, 0, 0, time.UTC), true},
		{"friday", time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC), true},
		{"saturday", time.Date(2024, 3, 16, 12, 0, 0, 0, time.UTC), false},
		{"sunday", time.Date(2024, 3, 17, 12, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeekday(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkStartOfDay(b *testing.B) {
	input := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)
	for i := 0; i < b.N; i++ {
		StartOfDay(input)
	}
}

func BenchmarkEndOfMonth(b *testing.B) {
	input := time.Date(2024, 3, 15, 14, 30, 45, 123456789, time.UTC)
	for i := 0; i < b.N; i++ {
		EndOfMonth(input)
	}
}