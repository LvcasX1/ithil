package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFormatRelativeTime_JustNow(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-30 * time.Second))
	if result != "just now" {
		t.Errorf("Expected 'just now', got '%s'", result)
	}
}

func TestFormatRelativeTime_Minutes(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-5 * time.Minute))
	if result != "5m ago" {
		t.Errorf("Expected '5m ago', got '%s'", result)
	}
}

func TestFormatRelativeTime_OneMinute(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-1 * time.Minute))
	if result != "1m ago" {
		t.Errorf("Expected '1m ago', got '%s'", result)
	}
}

func TestFormatRelativeTime_Hours(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-3 * time.Hour))
	if result != "3h ago" {
		t.Errorf("Expected '3h ago', got '%s'", result)
	}
}

func TestFormatRelativeTime_OneHour(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-1 * time.Hour))
	if result != "1h ago" {
		t.Errorf("Expected '1h ago', got '%s'", result)
	}
}

func TestFormatRelativeTime_Yesterday(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-24 * time.Hour))
	if result != "yesterday" {
		t.Errorf("Expected 'yesterday', got '%s'", result)
	}
}

func TestFormatRelativeTime_Days(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-3 * 24 * time.Hour))
	if result != "3d ago" {
		t.Errorf("Expected '3d ago', got '%s'", result)
	}
}

func TestFormatRelativeTime_Weeks(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(-14 * 24 * time.Hour))
	if result != "2w ago" {
		t.Errorf("Expected '2w ago', got '%s'", result)
	}
}

func TestFormatRelativeTime_Future(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(5 * time.Minute))
	// Allow for slight timing variations (4m or 5m are both acceptable)
	if result != "in 5m" && result != "in 4m" {
		t.Errorf("Expected 'in 5m' or 'in 4m', got '%s'", result)
	}
}

func TestFormatRelativeTime_FutureNow(t *testing.T) {
	now := time.Now()
	result := FormatRelativeTime(now.Add(30 * time.Second))
	if result != "in a moment" {
		t.Errorf("Expected 'in a moment', got '%s'", result)
	}
}

func TestFormatAbsoluteTime_Today(t *testing.T) {
	now := time.Now()
	result := FormatAbsoluteTime(now)
	// Should only show time (HH:MM)
	if !strings.Contains(result, ":") || len(result) != 5 {
		t.Errorf("Expected time format HH:MM, got '%s'", result)
	}
}

func TestFormatAbsoluteTime_Yesterday(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	result := FormatAbsoluteTime(yesterday)
	if !strings.HasPrefix(result, "Yesterday ") {
		t.Errorf("Expected 'Yesterday ...' prefix, got '%s'", result)
	}
}

func TestFormatAbsoluteTime_ThisYear(t *testing.T) {
	// Date 30 days ago but same year
	pastDate := time.Now().AddDate(0, 0, -30)
	result := FormatAbsoluteTime(pastDate)
	// Should contain month and day but not year
	if strings.Contains(result, "2006") || strings.Contains(result, fmt.Sprint(time.Now().Year())) {
		t.Errorf("Expected format without year, got '%s'", result)
	}
}

func TestFormatTime12Hour(t *testing.T) {
	testTime := time.Date(2025, 1, 1, 14, 30, 0, 0, time.UTC)
	result := FormatTime12Hour(testTime)
	if result != "2:30 PM" {
		t.Errorf("Expected '2:30 PM', got '%s'", result)
	}
}

func TestFormatTime24Hour(t *testing.T) {
	testTime := time.Date(2025, 1, 1, 14, 30, 0, 0, time.UTC)
	result := FormatTime24Hour(testTime)
	if result != "14:30" {
		t.Errorf("Expected '14:30', got '%s'", result)
	}
}

func TestFormatDate_Today(t *testing.T) {
	now := time.Now()
	result := FormatDate(now)
	if result != "Today" {
		t.Errorf("Expected 'Today', got '%s'", result)
	}
}

func TestFormatDate_Yesterday(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	result := FormatDate(yesterday)
	if result != "Yesterday" {
		t.Errorf("Expected 'Yesterday', got '%s'", result)
	}
}

func TestFormatDuration_Seconds(t *testing.T) {
	result := FormatDuration(30 * time.Second)
	if result != "30s" {
		t.Errorf("Expected '30s', got '%s'", result)
	}
}

func TestFormatDuration_Minutes(t *testing.T) {
	result := FormatDuration(5 * time.Minute)
	if result != "5m" {
		t.Errorf("Expected '5m', got '%s'", result)
	}
}

func TestFormatDuration_MinutesAndSeconds(t *testing.T) {
	result := FormatDuration(5*time.Minute + 30*time.Second)
	if result != "5m 30s" {
		t.Errorf("Expected '5m 30s', got '%s'", result)
	}
}

func TestFormatDuration_Hours(t *testing.T) {
	result := FormatDuration(2 * time.Hour)
	if result != "2h" {
		t.Errorf("Expected '2h', got '%s'", result)
	}
}

func TestFormatDuration_HoursAndMinutes(t *testing.T) {
	result := FormatDuration(2*time.Hour + 30*time.Minute)
	if result != "2h 30m" {
		t.Errorf("Expected '2h 30m', got '%s'", result)
	}
}

func TestFormatLastSeen_Online(t *testing.T) {
	now := time.Now()
	result := FormatLastSeen(now.Add(-30 * time.Second))
	if result != "online" {
		t.Errorf("Expected 'online', got '%s'", result)
	}
}

func TestFormatLastSeen_Recently(t *testing.T) {
	now := time.Now()
	result := FormatLastSeen(now.Add(-3 * time.Minute))
	if result != "last seen recently" {
		t.Errorf("Expected 'last seen recently', got '%s'", result)
	}
}

func TestFormatLastSeen_Minutes(t *testing.T) {
	now := time.Now()
	result := FormatLastSeen(now.Add(-15 * time.Minute))
	if result != "last seen 15 minutes ago" {
		t.Errorf("Expected 'last seen 15 minutes ago', got '%s'", result)
	}
}

func TestFormatLastSeen_OneHour(t *testing.T) {
	now := time.Now()
	result := FormatLastSeen(now.Add(-1 * time.Hour))
	if result != "last seen 1 hour ago" {
		t.Errorf("Expected 'last seen 1 hour ago', got '%s'", result)
	}
}

func TestFormatLastSeen_Hours(t *testing.T) {
	now := time.Now()
	result := FormatLastSeen(now.Add(-5 * time.Hour))
	if result != "last seen 5 hours ago" {
		t.Errorf("Expected 'last seen 5 hours ago', got '%s'", result)
	}
}

func TestFormatTimestamp_Relative(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-5 * time.Minute)
	result := FormatTimestamp(pastTime, true)
	if result != "5m ago" {
		t.Errorf("Expected '5m ago', got '%s'", result)
	}
}

func TestFormatTimestamp_Absolute(t *testing.T) {
	now := time.Now()
	result := FormatTimestamp(now, false)
	// Should be in time format for today
	if !strings.Contains(result, ":") {
		t.Errorf("Expected time format, got '%s'", result)
	}
}

func TestIsToday(t *testing.T) {
	now := time.Now()
	if !isToday(now, now) {
		t.Error("Expected now to be today")
	}

	yesterday := now.Add(-24 * time.Hour)
	if isToday(yesterday, now) {
		t.Error("Expected yesterday not to be today")
	}
}

func TestIsYesterday(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	if !isYesterday(yesterday, now) {
		t.Error("Expected time to be yesterday")
	}

	if isYesterday(now, now) {
		t.Error("Expected now not to be yesterday")
	}
}
