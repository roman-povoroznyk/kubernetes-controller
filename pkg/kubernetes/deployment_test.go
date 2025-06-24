package kubernetes

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		age      time.Duration
		expected string
	}{
		{
			name:     "30 seconds",
			age:      30 * time.Second,
			expected: "30s",
		},
		{
			name:     "5 minutes",
			age:      5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "5 minutes 30 seconds",
			age:      5*time.Minute + 30*time.Second,
			expected: "5m30s",
		},
		{
			name:     "2 hours",
			age:      2 * time.Hour,
			expected: "2h",
		},
		{
			name:     "2 hours 30 minutes",
			age:      2*time.Hour + 30*time.Minute,
			expected: "2h30m",
		},
		{
			name:     "1 day",
			age:      24 * time.Hour,
			expected: "1d",
		},
		{
			name:     "1 day 5 hours",
			age:      24*time.Hour + 5*time.Hour,
			expected: "1d5h",
		},
		{
			name:     "3 days",
			age:      3 * 24 * time.Hour,
			expected: "3d",
		},
		{
			name:     "3 days 2 hours",
			age:      3*24*time.Hour + 2*time.Hour,
			expected: "3d2h",
		},
		{
			name:     "7 days",
			age:      7 * 24 * time.Hour,
			expected: "7d",
		},
		{
			name:     "10 days",
			age:      10 * 24 * time.Hour,
			expected: "10d",
		},
		{
			name:     "30 days",
			age:      30 * 24 * time.Hour,
			expected: "30d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := now.Add(-tt.age)
			result := FormatAge(testTime)
			if result != tt.expected {
				t.Errorf("FormatAge() = %v, want %v", result, tt.expected)
			}
		})
	}
}
