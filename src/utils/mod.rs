//! Utility functions for Ithil.
//!
//! This module provides common utility functions for text formatting,
//! time handling, and other helper operations.

mod formatting;
mod notify;
mod time;

pub use formatting::{first_url, format_file_size, truncate_string, word_wrap};
pub use notify::sanitize;
pub use time::{format_duration, format_relative_time, format_timestamp};
