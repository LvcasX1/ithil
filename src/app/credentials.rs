//! Default API credentials handling for Telegram.
//!
//! This module provides placeholder credentials for development.
//! In production, users should provide their own credentials from <https://my.telegram.org>.

use crate::app::Config;

/// Telegram API credentials.
#[derive(Debug, Clone)]
pub struct Credentials {
    /// API ID (numeric)
    pub api_id: i32,

    /// API Hash (32 hex characters)
    pub api_hash: String,
}

impl Credentials {
    /// Default (placeholder) API credentials.
    ///
    /// These are placeholder values for development.
    /// Users should obtain their own credentials from <https://my.telegram.org>.
    ///
    /// # Security Note
    ///
    /// In a production build, these would be replaced with actual default credentials
    /// or the application would require users to provide their own.
    pub const DEFAULT_API_ID: i32 = 0;
    pub const DEFAULT_API_HASH: &'static str = "";

    /// Create credentials from configuration.
    ///
    /// If `use_default_credentials` is true, returns the built-in defaults.
    /// Otherwise, uses the custom credentials from the config.
    #[must_use]
    pub fn from_config(config: &Config) -> Self {
        if config.telegram.use_default_credentials {
            Self::default()
        } else {
            Self {
                api_id: config
                    .telegram
                    .api_id
                    .parse()
                    .unwrap_or(Self::DEFAULT_API_ID),
                api_hash: config.telegram.api_hash.clone(),
            }
        }
    }

    /// Check if these are valid credentials.
    ///
    /// Returns `true` if both API ID and API Hash are non-empty/non-zero.
    #[must_use]
    pub fn is_valid(&self) -> bool {
        self.api_id > 0 && !self.api_hash.is_empty()
    }
}

impl Default for Credentials {
    fn default() -> Self {
        Self {
            api_id: Self::DEFAULT_API_ID,
            api_hash: Self::DEFAULT_API_HASH.to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_credentials() {
        let creds = Credentials::default();
        // Default placeholder credentials should be invalid
        assert!(!creds.is_valid());
    }

    #[test]
    fn test_credentials_from_config_default() {
        let config = Config::default();
        let creds = Credentials::from_config(&config);

        // Should use default credentials
        assert_eq!(creds.api_id, Credentials::DEFAULT_API_ID);
    }

    #[test]
    fn test_credentials_from_config_custom() {
        let mut config = Config::default();
        config.telegram.use_default_credentials = false;
        config.telegram.api_id = "12345678".to_string();
        config.telegram.api_hash = "abcdef1234567890abcdef1234567890".to_string();

        let creds = Credentials::from_config(&config);

        assert_eq!(creds.api_id, 12345678);
        assert_eq!(creds.api_hash, "abcdef1234567890abcdef1234567890");
        assert!(creds.is_valid());
    }
}
