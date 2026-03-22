//! Time formatting utilities.
//!
//! This module provides functions for formatting timestamps and durations
//! in human-readable formats.

use chrono::{DateTime, Duration, Local, Utc};

/// Formats a timestamp for display.
///
/// # Arguments
///
/// * `time` - The timestamp to format
/// * `relative` - If true, returns relative time (e.g., "5m ago"); otherwise absolute
///
/// # Examples
///
/// ```
/// use chrono::Utc;
/// use ithil::utils::format_timestamp;
///
/// let now = Utc::now();
/// let result = format_timestamp(now, true);
/// assert_eq!(result, "just now");
/// ```
#[must_use]
pub fn format_timestamp(time: DateTime<Utc>, relative: bool) -> String {
    if relative {
        format_relative_time(time)
    } else {
        format_absolute_time(time)
    }
}

/// Formats a time as a relative string (e.g., "5m ago", "2h ago").
///
/// # Arguments
///
/// * `time` - The timestamp to format
///
/// # Returns
///
/// A human-readable relative time string.
///
/// # Examples
///
/// ```
/// use chrono::{Utc, Duration};
/// use ithil::utils::format_relative_time;
///
/// let now = Utc::now();
/// assert_eq!(format_relative_time(now), "just now");
///
/// let five_min_ago = now - Duration::minutes(5);
/// assert_eq!(format_relative_time(five_min_ago), "5m ago");
/// ```
#[must_use]
pub fn format_relative_time(time: DateTime<Utc>) -> String {
    let now = Utc::now();
    let diff = now.signed_duration_since(time);

    // Future times
    if diff < Duration::zero() {
        let abs_diff = -diff;
        return format_future_time(abs_diff, time);
    }

    // Past times
    format_past_time(diff, time)
}

/// Format a time in the future.
fn format_future_time(diff: Duration, time: DateTime<Utc>) -> String {
    if diff < Duration::minutes(1) {
        "in a moment".to_string()
    } else if diff < Duration::hours(1) {
        format!("in {}m", diff.num_minutes())
    } else if diff < Duration::hours(24) {
        format!("in {}h", diff.num_hours())
    } else {
        format_absolute_time(time)
    }
}

/// Format a time in the past.
fn format_past_time(diff: Duration, time: DateTime<Utc>) -> String {
    if diff < Duration::minutes(1) {
        "just now".to_string()
    } else if diff < Duration::hours(1) {
        let minutes = diff.num_minutes();
        if minutes == 1 {
            "1m ago".to_string()
        } else {
            format!("{minutes}m ago")
        }
    } else if diff < Duration::hours(24) {
        let hours = diff.num_hours();
        if hours == 1 {
            "1h ago".to_string()
        } else {
            format!("{hours}h ago")
        }
    } else if diff < Duration::days(7) {
        let days = diff.num_days();
        if days == 1 {
            "yesterday".to_string()
        } else {
            format!("{days}d ago")
        }
    } else if diff < Duration::days(30) {
        let weeks = diff.num_weeks();
        if weeks == 1 {
            "1w ago".to_string()
        } else {
            format!("{weeks}w ago")
        }
    } else {
        format_absolute_time(time)
    }
}

/// Formats a time as an absolute string.
///
/// Returns different formats based on how recent the time is:
/// - Today: "15:04"
/// - Yesterday: "Yesterday 15:04"
/// - This year: "Jan 02 15:04"
/// - Other: "Jan 02, 2006 15:04"
#[must_use]
pub fn format_absolute_time(time: DateTime<Utc>) -> String {
    let local_time = time.with_timezone(&Local);
    let now = Local::now();

    if is_today(&local_time, &now) {
        return local_time.format("%H:%M").to_string();
    }

    if is_yesterday(&local_time, &now) {
        return format!("Yesterday {}", local_time.format("%H:%M"));
    }

    if local_time.format("%Y").to_string() == now.format("%Y").to_string() {
        return local_time.format("%b %d %H:%M").to_string();
    }

    local_time.format("%b %d, %Y %H:%M").to_string()
}

/// Formats a duration in a human-readable way.
///
/// # Arguments
///
/// * `duration` - The duration to format
///
/// # Returns
///
/// A formatted string like "5m", "2h 30m", etc.
///
/// # Examples
///
/// ```
/// use chrono::Duration;
/// use ithil::utils::format_duration;
///
/// assert_eq!(format_duration(Duration::seconds(30)), "30s");
/// assert_eq!(format_duration(Duration::minutes(5)), "5m");
/// assert_eq!(format_duration(Duration::hours(2)), "2h");
/// ```
#[must_use]
pub fn format_duration(duration: Duration) -> String {
    let total_seconds = duration.num_seconds();

    if total_seconds < 60 {
        return format!("{total_seconds}s");
    }

    if total_seconds < 3600 {
        let minutes = duration.num_minutes();
        let seconds = total_seconds % 60;
        if seconds == 0 {
            return format!("{minutes}m");
        }
        return format!("{minutes}m {seconds}s");
    }

    let hours = duration.num_hours();
    let minutes = (total_seconds / 60) % 60;
    if minutes == 0 {
        return format!("{hours}h");
    }
    format!("{hours}h {minutes}m")
}

/// Checks if a datetime is today.
fn is_today<Tz: chrono::TimeZone>(time: &DateTime<Tz>, now: &DateTime<Local>) -> bool {
    let time_local = time.with_timezone(&Local);
    time_local.date_naive() == now.date_naive()
}

/// Checks if a datetime is yesterday.
fn is_yesterday<Tz: chrono::TimeZone>(time: &DateTime<Tz>, now: &DateTime<Local>) -> bool {
    let time_local = time.with_timezone(&Local);
    let yesterday = now
        .date_naive()
        .pred_opt()
        .unwrap_or_else(|| now.date_naive());
    time_local.date_naive() == yesterday
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn just_now() {
        let now = Utc::now();
        assert_eq!(format_relative_time(now), "just now");
    }

    #[test]
    fn minutes_ago() {
        let time = Utc::now() - Duration::minutes(5);
        assert_eq!(format_relative_time(time), "5m ago");
    }

    #[test]
    fn one_minute_ago() {
        let time = Utc::now() - Duration::minutes(1);
        assert_eq!(format_relative_time(time), "1m ago");
    }

    #[test]
    fn hours_ago() {
        let time = Utc::now() - Duration::hours(3);
        assert_eq!(format_relative_time(time), "3h ago");
    }

    #[test]
    fn one_hour_ago() {
        let time = Utc::now() - Duration::hours(1);
        assert_eq!(format_relative_time(time), "1h ago");
    }

    #[test]
    fn yesterday() {
        let time = Utc::now() - Duration::days(1);
        assert_eq!(format_relative_time(time), "yesterday");
    }

    #[test]
    fn days_ago() {
        let time = Utc::now() - Duration::days(3);
        assert_eq!(format_relative_time(time), "3d ago");
    }

    #[test]
    fn weeks_ago() {
        let time = Utc::now() - Duration::weeks(2);
        assert_eq!(format_relative_time(time), "2w ago");
    }

    #[test]
    fn future_time() {
        // Use 10 minutes and check for a range to avoid race conditions
        // at minute boundaries
        let time = Utc::now() + Duration::minutes(10);
        let result = format_relative_time(time);
        // Should be "in 9m" or "in 10m" depending on exact timing
        assert!(
            result == "in 9m" || result == "in 10m",
            "Expected 'in 9m' or 'in 10m', got '{result}'"
        );
    }

    #[test]
    fn format_duration_seconds() {
        assert_eq!(format_duration(Duration::seconds(30)), "30s");
        assert_eq!(format_duration(Duration::seconds(59)), "59s");
    }

    #[test]
    fn format_duration_minutes() {
        assert_eq!(format_duration(Duration::minutes(5)), "5m");
        assert_eq!(format_duration(Duration::seconds(90)), "1m 30s");
    }

    #[test]
    fn format_duration_hours() {
        assert_eq!(format_duration(Duration::hours(2)), "2h");
        assert_eq!(format_duration(Duration::minutes(150)), "2h 30m");
    }
}
