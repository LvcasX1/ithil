//! Configuration management for Ithil.
//!
//! This module provides configuration loading, validation, and serialization
//! matching the Go version's YAML format for compatibility.

use std::collections::HashMap;
use std::fs;
use std::path::{Path, PathBuf};

use anyhow::{Context, Result};
use directories::ProjectDirs;
use serde::{Deserialize, Serialize};
use thiserror::Error;

/// Configuration errors.
#[derive(Error, Debug)]
pub enum ConfigError {
    #[error("Configuration file not found: {0}")]
    NotFound(PathBuf),

    #[error("Failed to parse configuration: {0}")]
    ParseError(#[from] serde_yaml::Error),

    #[error("Invalid configuration: {0}")]
    ValidationError(String),

    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}

/// Main configuration structure matching the Go version's YAML format.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(default)]
pub struct Config {
    /// General application settings
    pub app: AppConfig,

    /// Telegram API settings
    pub telegram: TelegramConfig,

    /// User interface settings
    pub ui: UiConfig,

    /// Notification settings
    pub notifications: NotificationConfig,

    /// Privacy settings
    pub privacy: PrivacyConfig,

    /// Cache settings
    pub cache: CacheConfig,

    /// Logging settings
    pub logging: LoggingConfig,
}

/// General application settings.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct AppConfig {
    /// Application name
    pub name: String,

    /// Application version
    pub version: String,
}

/// Telegram API configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct TelegramConfig {
    /// Use built-in credentials (default: true)
    pub use_default_credentials: bool,

    /// Custom API ID (optional, only used if `use_default_credentials` is false)
    pub api_id: String,

    /// Custom API Hash (optional, only used if `use_default_credentials` is false)
    pub api_hash: String,

    /// Path to session file
    pub session_file: PathBuf,

    /// Path to database directory
    pub database_directory: PathBuf,
}

/// User interface configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct UiConfig {
    /// Theme: "dark", "light", or "nord"
    pub theme: String,

    /// Layout settings
    pub layout: LayoutConfig,

    /// Appearance settings
    pub appearance: AppearanceConfig,

    /// Behavior settings
    pub behavior: BehaviorConfig,

    /// Keyboard settings
    pub keyboard: KeyboardConfig,
}

/// Layout configuration defining pane widths.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct LayoutConfig {
    /// Chat list pane width percentage
    pub chat_list_width: u8,

    /// Conversation pane width percentage
    pub conversation_width: u8,

    /// Info pane width percentage
    pub info_width: u8,

    /// Whether to show the info pane
    pub show_info_pane: bool,
}

/// Appearance configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct AppearanceConfig {
    /// Show user avatars
    pub show_avatars: bool,

    /// Show status bar
    pub show_status_bar: bool,

    /// Date format: "12h" or "24h"
    pub date_format: String,

    /// Use relative timestamps (e.g., "2 hours ago")
    pub relative_timestamps: bool,

    /// Maximum length of message preview in chat list
    pub message_preview_length: usize,
}

/// Behavior configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct BehaviorConfig {
    /// Send message on Enter (false for Ctrl+Enter)
    pub send_on_enter: bool,

    /// Auto-download limit in bytes
    pub auto_download_limit: u64,

    /// Mark messages as read when scrolling
    pub mark_read_on_scroll: bool,

    /// Emoji style: "unicode" or "ascii"
    pub emoji_style: String,
}

/// Keyboard configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct KeyboardConfig {
    /// Enable vim-style keybindings (j/k navigation)
    pub vim_mode: bool,

    /// Custom key bindings
    pub custom_bindings: HashMap<String, String>,
}

/// Notification configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct NotificationConfig {
    /// Enable notifications
    pub enabled: bool,

    /// Enable notification sounds
    pub sound: bool,

    /// Enable desktop notifications
    pub desktop: bool,

    /// List of muted chat IDs
    pub muted_chats: Vec<i64>,
}

/// Privacy configuration.
///
/// Note: This struct contains multiple boolean fields which is intentional
/// for matching the Go configuration format and providing granular privacy controls.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
#[allow(clippy::struct_excessive_bools)]
pub struct PrivacyConfig {
    /// Show online status to others
    pub show_online_status: bool,

    /// Send read receipts
    pub show_read_receipts: bool,

    /// Show typing indicator
    pub show_typing: bool,

    /// Stealth mode (disables read receipts and typing indicators)
    pub stealth_mode: bool,
}

/// Cache configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct CacheConfig {
    /// Maximum messages to cache per chat
    pub max_messages_per_chat: usize,

    /// Maximum media file size in bytes
    pub max_media_size: u64,

    /// Directory for cached media files
    pub media_directory: PathBuf,
}

/// Logging configuration.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct LoggingConfig {
    /// Log level: "trace", "debug", "info", "warn", "error"
    pub level: String,

    /// Path to log file
    pub file: PathBuf,
}

// Default implementations

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            name: "Ithil".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
        }
    }
}

impl Default for TelegramConfig {
    fn default() -> Self {
        let config_dir = get_config_dir();
        Self {
            use_default_credentials: true,
            api_id: String::new(),
            api_hash: String::new(),
            // Use .session extension for grammers SQLite session format
            // Different from Go version's session.json to avoid conflicts
            session_file: config_dir.join("ithil.session"),
            database_directory: config_dir.join("tdlib"),
        }
    }
}

impl Default for UiConfig {
    fn default() -> Self {
        Self {
            theme: "dark".to_string(),
            layout: LayoutConfig::default(),
            appearance: AppearanceConfig::default(),
            behavior: BehaviorConfig::default(),
            keyboard: KeyboardConfig::default(),
        }
    }
}

impl Default for LayoutConfig {
    fn default() -> Self {
        Self {
            chat_list_width: 25,
            conversation_width: 50,
            info_width: 25,
            show_info_pane: true,
        }
    }
}

impl Default for AppearanceConfig {
    fn default() -> Self {
        Self {
            show_avatars: true,
            show_status_bar: true,
            date_format: "12h".to_string(),
            relative_timestamps: true,
            message_preview_length: 50,
        }
    }
}

impl Default for BehaviorConfig {
    fn default() -> Self {
        Self {
            send_on_enter: true,
            auto_download_limit: 5_242_880, // 5MB
            mark_read_on_scroll: true,
            emoji_style: "unicode".to_string(),
        }
    }
}

impl Default for KeyboardConfig {
    fn default() -> Self {
        Self {
            vim_mode: true,
            custom_bindings: HashMap::new(),
        }
    }
}

impl Default for NotificationConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            sound: true,
            desktop: false,
            muted_chats: Vec::new(),
        }
    }
}

impl Default for PrivacyConfig {
    fn default() -> Self {
        Self {
            show_online_status: true,
            show_read_receipts: true,
            show_typing: true,
            stealth_mode: false,
        }
    }
}

impl Default for CacheConfig {
    fn default() -> Self {
        let cache_dir = get_cache_dir();
        Self {
            max_messages_per_chat: 1000,
            max_media_size: 104_857_600, // 100MB
            media_directory: cache_dir.join("media"),
        }
    }
}

impl Default for LoggingConfig {
    fn default() -> Self {
        let config_dir = get_config_dir();
        Self {
            level: "info".to_string(),
            file: config_dir.join("ithil.log"),
        }
    }
}

impl Config {
    /// Load configuration from the specified path or default locations.
    ///
    /// Search order:
    /// 1. Specified path (if provided)
    /// 2. `./config.yaml`
    /// 3. `~/.config/ithil/config.yaml`
    /// 4. `~/.ithil.yaml`
    ///
    /// If no config file is found, returns the default configuration.
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The specified path doesn't exist
    /// - The config file cannot be read
    /// - The config file contains invalid YAML
    pub fn load(path: Option<&Path>) -> Result<Self> {
        // Try specified path first
        if let Some(p) = path {
            let expanded = expand_tilde(p);
            if expanded.exists() {
                return Self::load_from_file(&expanded);
            }
            // If specified path doesn't exist, return error
            return Err(ConfigError::NotFound(expanded).into());
        }

        // Try default locations
        let locations = Self::config_search_paths();

        for location in locations {
            if location.exists() {
                tracing::debug!("Loading config from: {}", location.display());
                return Self::load_from_file(&location);
            }
        }

        // No config file found, use defaults
        tracing::info!("No config file found, using defaults");
        Ok(Self::default())
    }

    /// Get the list of paths to search for config files.
    fn config_search_paths() -> Vec<PathBuf> {
        let mut paths = vec![PathBuf::from("config.yaml")];

        if let Some(home) = dirs::home_dir() {
            paths.push(home.join(".config").join("ithil").join("config.yaml"));
            paths.push(home.join(".ithil.yaml"));
        }

        paths
    }

    /// Load configuration from a specific file.
    fn load_from_file(path: &Path) -> Result<Self> {
        let content = fs::read_to_string(path)
            .with_context(|| format!("Failed to read config file: {}", path.display()))?;

        let mut config: Self = serde_yaml::from_str(&content)
            .with_context(|| format!("Failed to parse config file: {}", path.display()))?;

        // Expand paths
        config.expand_paths();

        Ok(config)
    }

    /// Save configuration to a file.
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The parent directory cannot be created
    /// - The configuration cannot be serialized
    /// - The file cannot be written
    pub fn save(&self, path: &Path) -> Result<()> {
        let expanded = expand_tilde(path);

        // Create parent directory if needed
        if let Some(parent) = expanded.parent() {
            fs::create_dir_all(parent).with_context(|| {
                format!("Failed to create config directory: {}", parent.display())
            })?;
        }

        let content = serde_yaml::to_string(self).context("Failed to serialize configuration")?;

        fs::write(&expanded, content)
            .with_context(|| format!("Failed to write config file: {}", expanded.display()))?;

        Ok(())
    }

    /// Validate the configuration.
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - Custom credentials are enabled but not properly configured
    /// - Layout widths don't sum to 100%
    pub fn validate(&self) -> Result<(), ConfigError> {
        // Validate custom credentials if not using defaults
        if !self.telegram.use_default_credentials {
            if self.telegram.api_id.is_empty() || self.telegram.api_id == "YOUR_API_ID" {
                return Err(ConfigError::ValidationError(
                    "Telegram API ID is not configured".to_string(),
                ));
            }

            if self.telegram.api_hash.is_empty() || self.telegram.api_hash == "YOUR_API_HASH" {
                return Err(ConfigError::ValidationError(
                    "Telegram API hash is not configured".to_string(),
                ));
            }
        }

        // Validate layout widths sum to 100
        let total_width = u16::from(self.ui.layout.chat_list_width)
            + u16::from(self.ui.layout.conversation_width)
            + u16::from(self.ui.layout.info_width);

        if total_width != 100 {
            return Err(ConfigError::ValidationError(format!(
                "Layout widths must sum to 100%, got {total_width}%"
            )));
        }

        Ok(())
    }

    /// Expand tilde in all path fields.
    fn expand_paths(&mut self) {
        self.telegram.session_file = expand_tilde(&self.telegram.session_file);
        self.telegram.database_directory = expand_tilde(&self.telegram.database_directory);
        self.cache.media_directory = expand_tilde(&self.cache.media_directory);
        self.logging.file = expand_tilde(&self.logging.file);
    }

    /// Ensure all required directories exist.
    ///
    /// # Errors
    ///
    /// Returns an error if any directory cannot be created.
    pub fn ensure_directories(&self) -> Result<()> {
        let dirs = [
            self.telegram.session_file.parent(),
            Some(self.telegram.database_directory.as_path()),
            Some(self.cache.media_directory.as_path()),
            self.logging.file.parent(),
        ];

        for dir in dirs.into_iter().flatten() {
            fs::create_dir_all(dir)
                .with_context(|| format!("Failed to create directory: {}", dir.display()))?;
        }

        Ok(())
    }
}

/// Get the default config directory for Ithil.
fn get_config_dir() -> PathBuf {
    ProjectDirs::from("", "", "ithil").map_or_else(
        || {
            dirs::home_dir().map_or_else(
                || PathBuf::from(".config/ithil"),
                |h| h.join(".config").join("ithil"),
            )
        },
        |dirs| dirs.config_dir().to_path_buf(),
    )
}

/// Get the default cache directory for Ithil.
fn get_cache_dir() -> PathBuf {
    ProjectDirs::from("", "", "ithil").map_or_else(
        || {
            dirs::home_dir().map_or_else(
                || PathBuf::from(".cache/ithil"),
                |h| h.join(".cache").join("ithil"),
            )
        },
        |dirs| dirs.cache_dir().to_path_buf(),
    )
}

/// Expand tilde (~) to home directory in a path.
fn expand_tilde(path: &Path) -> PathBuf {
    let path_str = path.to_string_lossy();

    if let Some(stripped) = path_str.strip_prefix('~') {
        if let Some(home) = dirs::home_dir() {
            let rest = stripped.strip_prefix('/').unwrap_or(stripped);
            return home.join(rest);
        }
    }

    path.to_path_buf()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = Config::default();
        assert_eq!(config.app.name, "Ithil");
        assert!(config.telegram.use_default_credentials);
        assert_eq!(config.ui.layout.chat_list_width, 25);
        assert_eq!(config.ui.layout.conversation_width, 50);
        assert_eq!(config.ui.layout.info_width, 25);
    }

    #[test]
    fn test_config_validation_layout_widths() {
        let mut config = Config::default();
        config.ui.layout.chat_list_width = 30;
        // Total is now 30 + 50 + 25 = 105, should fail

        let result = config.validate();
        assert!(result.is_err());
    }

    #[test]
    fn test_config_validation_custom_credentials() {
        let mut config = Config::default();
        config.telegram.use_default_credentials = false;
        // Empty credentials should fail

        let result = config.validate();
        assert!(result.is_err());
    }

    #[test]
    fn test_expand_tilde() {
        let path = Path::new("~/test/path");
        let expanded = expand_tilde(path);

        if let Some(home) = dirs::home_dir() {
            assert!(expanded.starts_with(&home));
            assert!(expanded.ends_with("test/path"));
        }
    }
}
