package client

import (
	"testing"
	"time"
)

func TestDailyTitle(t *testing.T) {
	tests := []struct {
		date string
		want string
	}{
		{"2026-01-01", "January 1st, 2026"},
		{"2026-01-02", "January 2nd, 2026"},
		{"2026-01-03", "January 3rd, 2026"},
		{"2026-01-04", "January 4th, 2026"},
		{"2026-01-11", "January 11th, 2026"},
		{"2026-01-12", "January 12th, 2026"},
		{"2026-01-13", "January 13th, 2026"},
		{"2026-01-21", "January 21st, 2026"},
		{"2026-01-22", "January 22nd, 2026"},
		{"2026-01-23", "January 23rd, 2026"},
		{"2026-01-31", "January 31st, 2026"},
		{"2026-02-14", "February 14th, 2026"},
		{"2026-03-04", "March 4th, 2026"},
		{"2026-03-14", "March 14th, 2026"},
		{"2026-12-25", "December 25th, 2026"},
	}
	for _, tt := range tests {
		d, _ := time.Parse("2006-01-02", tt.date)
		got := DailyTitle(d)
		if got != tt.want {
			t.Errorf("DailyTitle(%s) = %q, want %q", tt.date, got, tt.want)
		}
	}
}
