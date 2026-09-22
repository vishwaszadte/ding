package timeparser

import (
	"testing"
	"time"
)

// Table-Driven Tests: An idiomatic Go pattern.
// We declare a slice of anonymous structs defining test scenarios,
// then iterate over them using `t.Run(tc.name, ...)`.
func TestParse(t *testing.T) {
	// Fixed reference time for reproducible testing:
	// Wednesday, 2026-09-23 10:00:00 AM Local
	refTime := time.Date(2026, 9, 23, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name            string
		input           string
		expectError     bool
		expectedNextRun time.Time
		isRecurring     bool
		intervalSeconds int64
	}{
		{
			name:            "Relative minutes: in 15m",
			input:           "in 15m",
			expectedNextRun: refTime.Add(15 * time.Minute),
			isRecurring:     false,
		},
		{
			name:            "Relative hours without 'in': 2h",
			input:           "2h",
			expectedNextRun: refTime.Add(2 * time.Hour),
			isRecurring:     false,
		},
		{
			name:            "Relative seconds: in 30s",
			input:           "in 30s",
			expectedNextRun: refTime.Add(30 * time.Second),
			isRecurring:     false,
		},
		{
			name:            "Recurring every hour",
			input:           "every hour",
			expectedNextRun: refTime.Add(1 * time.Hour),
			isRecurring:     true,
			intervalSeconds: 3600,
		},
		{
			name:            "Recurring every 30m",
			input:           "every 30m",
			expectedNextRun: refTime.Add(30 * time.Minute),
			isRecurring:     true,
			intervalSeconds: 1800,
		},
		{
			name:            "Clock time: today 4pm",
			input:           "today 4pm",
			expectedNextRun: time.Date(2026, 9, 23, 16, 0, 0, 0, time.Local),
			isRecurring:     false,
		},
		{
			name:            "Clock time: tomorrow 9am",
			input:           "tomorrow 9am",
			expectedNextRun: time.Date(2026, 9, 24, 9, 0, 0, 0, time.Local),
			isRecurring:     false,
		},
		{
			name:            "Recurring every day 8pm",
			input:           "every day 8pm",
			expectedNextRun: time.Date(2026, 9, 23, 20, 0, 0, 0, time.Local),
			isRecurring:     true,
		},
		{
			name:            "Recurring every weekday 9am (today is Wed 10am, so next is Thu 9am)",
			input:           "every weekday 9am",
			expectedNextRun: time.Date(2026, 9, 24, 9, 0, 0, 0, time.Local),
			isRecurring:     true,
		},
		{
			name:        "Past time today should error: today 8am",
			input:       "today 8am",
			expectError: true,
		},
		{
			name:        "Invalid gibberish expression",
			input:       "sometime whenever",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse(tc.input, refTime)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error for input %q, but got nil", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}

			if !result.NextRunAt.Equal(tc.expectedNextRun) {
				t.Errorf("expected NextRunAt %v, got %v", tc.expectedNextRun, result.NextRunAt)
			}

			if result.IsRecurring != tc.isRecurring {
				t.Errorf("expected IsRecurring %v, got %v", tc.isRecurring, result.IsRecurring)
			}

			if tc.intervalSeconds > 0 && result.IntervalSeconds != tc.intervalSeconds {
				t.Errorf("expected IntervalSeconds %d, got %d", tc.intervalSeconds, result.IntervalSeconds)
			}
		})
	}
}
