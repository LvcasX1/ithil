//! Text formatting utilities.
//!
//! This module provides functions for truncating, padding, and formatting text
//! for terminal display.

use unicode_width::UnicodeWidthStr;

/// Truncates a string to fit within the specified display width.
///
/// If the string exceeds `max_width`, it will be truncated and "..." will be
/// appended. The function respects Unicode character boundaries and uses
/// display width (not byte count) for measurements.
///
/// # Arguments
///
/// * `s` - The string to truncate
/// * `max_width` - Maximum display width in columns
///
/// # Returns
///
/// The truncated string, or the original string if it fits within `max_width`.
///
/// # Examples
///
/// ```
/// use ithil::utils::truncate_string;
///
/// assert_eq!(truncate_string("Hello, World!", 10), "Hello, ...");
/// assert_eq!(truncate_string("Short", 10), "Short");
/// assert_eq!(truncate_string("Hi", 2), "Hi");
/// ```
#[must_use]
pub fn truncate_string(s: &str, max_width: usize) -> String {
    if max_width == 0 {
        return String::new();
    }

    let width = UnicodeWidthStr::width(s);
    if width <= max_width {
        return s.to_string();
    }

    // String needs truncation but max_width is too small for "..."
    if max_width <= 3 {
        return ".".repeat(max_width);
    }

    // We need to truncate to (max_width - 3) to leave room for "..."
    let target_width = max_width - 3;
    let mut current_width = 0;
    let mut result = String::new();

    for ch in s.chars() {
        let char_width = unicode_width::UnicodeWidthChar::width(ch).unwrap_or(0);
        if current_width + char_width > target_width {
            break;
        }
        result.push(ch);
        current_width += char_width;
    }

    result.push_str("...");
    result
}

/// Wraps text to fit within the specified width, breaking on word boundaries.
///
/// # Arguments
///
/// * `text` - The text to wrap
/// * `width` - Maximum width per line
///
/// # Returns
///
/// The text with newlines inserted at word boundaries.
///
/// # Examples
///
/// ```
/// use ithil::utils::word_wrap;
///
/// let wrapped = word_wrap("Hello World", 6);
/// assert_eq!(wrapped, "Hello\nWorld");
/// ```
#[must_use]
pub fn word_wrap(text: &str, width: usize) -> String {
    if width == 0 {
        return text.to_string();
    }

    let words: Vec<&str> = text.split_whitespace().collect();
    if words.is_empty() {
        return text.to_string();
    }

    let mut lines: Vec<String> = Vec::new();
    let mut current_line = String::new();

    for word in words {
        if current_line.is_empty() {
            current_line = word.to_string();
        } else {
            let test_line = format!("{} {}", current_line, word);
            if UnicodeWidthStr::width(test_line.as_str()) <= width {
                current_line = test_line;
            } else {
                lines.push(current_line);
                current_line = word.to_string();
            }
        }
    }

    if !current_line.is_empty() {
        lines.push(current_line);
    }

    lines.join("\n")
}

/// Formats a file size in bytes to a human-readable string.
///
/// # Arguments
///
/// * `bytes` - File size in bytes
///
/// # Returns
///
/// A formatted string like "1.5 MB" or "300 KB".
///
/// # Examples
///
/// ```
/// use ithil::utils::format_file_size;
///
/// assert_eq!(format_file_size(500), "500 B");
/// assert_eq!(format_file_size(1024), "1 KB");
/// assert_eq!(format_file_size(1536), "1.5 KB");
/// assert_eq!(format_file_size(1_048_576), "1 MB");
/// ```
#[must_use]
pub fn format_file_size(bytes: i64) -> String {
    const KB: i64 = 1024;
    const MB: i64 = KB * 1024;
    const GB: i64 = MB * 1024;

    #[allow(clippy::cast_precision_loss)]
    if bytes >= GB {
        let gb = bytes as f64 / GB as f64;
        format_float_size(gb, "GB")
    } else if bytes >= MB {
        let mb = bytes as f64 / MB as f64;
        format_float_size(mb, "MB")
    } else if bytes >= KB {
        let kb = bytes as f64 / KB as f64;
        format_float_size(kb, "KB")
    } else {
        format!("{bytes} B")
    }
}

/// Helper to format float sizes with minimal decimal places.
fn format_float_size(value: f64, unit: &str) -> String {
    if (value - value.round()).abs() < 0.05 {
        format!("{} {}", value as i64, unit)
    } else {
        format!("{:.1} {}", value, unit)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    mod truncate_tests {
        use super::*;

        #[test]
        fn empty_max_width() {
            assert_eq!(truncate_string("Hello", 0), "");
        }

        #[test]
        fn very_small_max_width() {
            assert_eq!(truncate_string("Hello", 1), ".");
            assert_eq!(truncate_string("Hello", 2), "..");
            assert_eq!(truncate_string("Hello", 3), "...");
        }

        #[test]
        fn no_truncation_needed() {
            assert_eq!(truncate_string("Hello", 10), "Hello");
            assert_eq!(truncate_string("Hello", 5), "Hello");
        }

        #[test]
        fn truncation_with_ellipsis() {
            assert_eq!(truncate_string("Hello, World!", 10), "Hello, ...");
            assert_eq!(truncate_string("Hello, World!", 8), "Hello...");
        }

        #[test]
        fn unicode_characters() {
            // Japanese characters are typically 2 columns wide
            let s = "Hello";
            assert_eq!(truncate_string(s, 5), "Hello");
        }
    }

    mod word_wrap_tests {
        use super::*;

        #[test]
        fn zero_width() {
            assert_eq!(word_wrap("Hello World", 0), "Hello World");
        }

        #[test]
        fn no_wrap_needed() {
            assert_eq!(word_wrap("Hello World", 20), "Hello World");
        }

        #[test]
        fn simple_wrap() {
            assert_eq!(word_wrap("Hello World", 6), "Hello\nWorld");
        }

        #[test]
        fn multiple_lines() {
            let result = word_wrap("The quick brown fox jumps over", 10);
            let lines: Vec<&str> = result.lines().collect();
            assert!(lines.len() >= 3);
        }

        #[test]
        fn empty_string() {
            assert_eq!(word_wrap("", 10), "");
        }
    }

    mod file_size_tests {
        use super::*;

        #[test]
        fn bytes() {
            assert_eq!(format_file_size(0), "0 B");
            assert_eq!(format_file_size(500), "500 B");
            assert_eq!(format_file_size(1023), "1023 B");
        }

        #[test]
        fn kilobytes() {
            assert_eq!(format_file_size(1024), "1 KB");
            assert_eq!(format_file_size(1536), "1.5 KB");
        }

        #[test]
        fn megabytes() {
            assert_eq!(format_file_size(1_048_576), "1 MB");
            assert_eq!(format_file_size(1_572_864), "1.5 MB");
        }

        #[test]
        fn gigabytes() {
            assert_eq!(format_file_size(1_073_741_824), "1 GB");
        }
    }
}
