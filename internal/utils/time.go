// Package utils provides utility functions for the Ithil application.
package utils

import (
	"fmt"
	"time"
)

// FormatTimestamp formats a timestamp for display.
// If relative is true, it returns a relative time string (e.g., "5m ago").
// Otherwise, it returns an absolute time string.
func FormatTimestamp(t time.Time, relative bool) string {
	if relative {
		return FormatRelativeTime(t)
	}
	return FormatAbsoluteTime(t)
}

// FormatRelativeTime formats a time as a relative string (e.g., "5m ago", "2h ago").
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	// Future times
	if diff < 0 {
		diff = -diff
		switch {
		case diff < time.Minute:
			return "in a moment"
		case diff < time.Hour:
			minutes := int(diff.Minutes())
			return fmt.Sprintf("in %dm", minutes)
		case diff < 24*time.Hour:
			hours := int(diff.Hours())
			return fmt.Sprintf("in %dh", hours)
		default:
			return FormatAbsoluteTime(t)
		}
	}

	// Past times
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%dd ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", weeks)
	default:
		return FormatAbsoluteTime(t)
	}
}

// FormatAbsoluteTime formats a time as an absolute string.
func FormatAbsoluteTime(t time.Time) string {
	now := time.Now()

	// If today, show only time
	if isToday(t, now) {
		return t.Format("15:04")
	}

	// If yesterday
	if isYesterday(t, now) {
		return "Yesterday " + t.Format("15:04")
	}

	// If this year, show date without year
	if t.Year() == now.Year() {
		return t.Format("Jan 02 15:04")
	}

	// Otherwise, show full date
	return t.Format("Jan 02, 2006 15:04")
}

// FormatTime12Hour formats a time in 12-hour format.
func FormatTime12Hour(t time.Time) string {
	return t.Format("3:04 PM")
}

// FormatTime24Hour formats a time in 24-hour format.
func FormatTime24Hour(t time.Time) string {
	return t.Format("15:04")
}

// FormatDate formats a date.
func FormatDate(t time.Time) string {
	now := time.Now()

	if isToday(t, now) {
		return "Today"
	}

	if isYesterday(t, now) {
		return "Yesterday"
	}

	if t.Year() == now.Year() {
		return t.Format("January 2")
	}

	return t.Format("January 2, 2006")
}

// FormatDateTime formats a date and time.
func FormatDateTime(t time.Time) string {
	now := time.Now()

	dateStr := FormatDate(t)
	timeStr := t.Format("15:04")

	if isToday(t, now) {
		return timeStr
	}

	return dateStr + " at " + timeStr
}

// FormatDuration formats a duration in a human-readable way.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		seconds := int(d.Seconds())
		return fmt.Sprintf("%ds", seconds)
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// FormatLastSeen formats the last seen time for a user.
func FormatLastSeen(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "online"
	}

	if diff < 5*time.Minute {
		return "last seen recently"
	}

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("last seen %d minutes ago", minutes)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "last seen 1 hour ago"
		}
		return fmt.Sprintf("last seen %d hours ago", hours)
	}

	return "last seen " + FormatDate(t)
}

// isToday checks if a time is today.
func isToday(t, now time.Time) bool {
	tYear, tMonth, tDay := t.Date()
	nowYear, nowMonth, nowDay := now.Date()
	return tYear == nowYear && tMonth == nowMonth && tDay == nowDay
}

// isYesterday checks if a time is yesterday.
func isYesterday(t, now time.Time) bool {
	yesterday := now.AddDate(0, 0, -1)
	tYear, tMonth, tDay := t.Date()
	yYear, yMonth, yDay := yesterday.Date()
	return tYear == yYear && tMonth == yMonth && tDay == yDay
}

// ParseDuration parses a duration string like "5m", "2h", "1d".
func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
