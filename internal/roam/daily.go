package roam

import (
	"fmt"
	"time"
)

// DailyTitle returns the Roam Research daily page title for the given date.
// Format: "January 2nd, 2006" (English month, day with ordinal suffix, comma, 4-digit year).
func DailyTitle(t time.Time) string {
	return fmt.Sprintf("%s %d%s, %d", t.Month().String(), t.Day(), ordinalSuffix(t.Day()), t.Year())
}

func ordinalSuffix(day int) string {
	if day >= 11 && day <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
