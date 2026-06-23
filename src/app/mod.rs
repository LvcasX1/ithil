//! Application-level functionality including configuration and credentials.
//!
//! This module provides:
//! - Configuration loading and management
//! - Default API credentials handling
//! - Application state management

mod config;
mod credentials;

pub use config::{Config, NotificationConfig};
pub use credentials::Credentials;
