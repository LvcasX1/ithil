//! Error types for Telegram client operations.
//!
//! This module provides a unified error type for all Telegram-related operations,
//! converting from grammers library errors into a more ergonomic interface.

use thiserror::Error;

/// Unified error type for Telegram client operations.
///
/// This enum covers all possible failure modes when interacting with Telegram,
/// from connection issues to authentication failures to API errors.
#[derive(Error, Debug)]
pub enum TelegramError {
    /// The client is not connected to Telegram servers.
    ///
    /// Call [`TelegramClient::connect`] before attempting operations.
    #[error("Not connected to Telegram")]
    NotConnected,

    /// Authentication is required before this operation can proceed.
    ///
    /// The user needs to complete the sign-in flow.
    #[error("Authentication required")]
    AuthRequired,

    /// The provided phone number is invalid or incorrectly formatted.
    ///
    /// Phone numbers should be in international format (e.g., "+1234567890").
    #[error("Invalid phone number: {0}")]
    InvalidPhoneNumber(String),

    /// The verification code entered is incorrect.
    #[error("Invalid verification code")]
    InvalidCode,

    /// The 2FA password entered is incorrect.
    #[error("Invalid password")]
    InvalidPassword,

    /// Sign-up is required because the phone number is not registered.
    ///
    /// The user needs to complete registration via [`TelegramClient::sign_up`].
    #[error("Sign up required for this phone number")]
    SignUpRequired,

    /// A 2FA password is required to complete authentication.
    ///
    /// Call [`TelegramClient::check_password`] with the user's password.
    #[error("Two-factor authentication password required")]
    PasswordRequired,

    /// Failed to load or save the session file.
    #[error("Session error: {0}")]
    Session(String),

    /// Network-related error during communication with Telegram.
    #[error("Network error: {0}")]
    Network(String),

    /// API error returned by Telegram servers.
    #[error("Telegram API error: {0}")]
    Api(String),

    /// Flood wait error - too many requests.
    ///
    /// The client should wait the specified number of seconds before retrying.
    #[error("Flood wait: retry after {0} seconds")]
    FloodWait(i32),

    /// The operation timed out.
    #[error("Operation timed out")]
    Timeout,

    /// The requested chat was not found.
    #[error("Chat not found: {0}")]
    ChatNotFound(i64),

    /// The requested message was not found.
    #[error("Message not found: {0}")]
    MessageNotFound(i64),

    /// The message does not contain any media.
    #[error("Message {0} does not contain media")]
    NoMedia(i64),

    /// The message does not contain a photo.
    #[error("Message {0} is not a photo")]
    NotAPhoto(i64),

    /// The requested file was not found.
    #[error("File not found: {0}")]
    FileNotFound(std::path::PathBuf),

    /// IO error during file operations.
    #[error("IO error: {0}")]
    Io(String),

    /// Internal error that should not occur in normal operation.
    #[error("Internal error: {0}")]
    Internal(String),
}

impl TelegramError {
    /// Returns `true` if this error is recoverable by retrying.
    ///
    /// Network errors and timeouts are generally recoverable.
    #[must_use]
    pub const fn is_recoverable(&self) -> bool {
        matches!(self, Self::Network(_) | Self::Timeout | Self::FloodWait(_))
    }

    /// Returns `true` if this error requires user action to resolve.
    ///
    /// Authentication-related errors require user input.
    #[must_use]
    pub const fn requires_user_action(&self) -> bool {
        matches!(
            self,
            Self::AuthRequired
                | Self::InvalidPhoneNumber(_)
                | Self::InvalidCode
                | Self::InvalidPassword
                | Self::SignUpRequired
                | Self::PasswordRequired
        )
    }
}

impl From<grammers_client::InvocationError> for TelegramError {
    fn from(err: grammers_client::InvocationError) -> Self {
        use grammers_client::InvocationError;

        match err {
            InvocationError::Rpc(rpc_error) => {
                let error_message = rpc_error.name.as_str();

                // Handle flood wait
                if error_message.starts_with("FLOOD_WAIT_") {
                    if let Some(seconds_str) = error_message.strip_prefix("FLOOD_WAIT_") {
                        if let Ok(seconds) = seconds_str.parse::<i32>() {
                            return Self::FloodWait(seconds);
                        }
                    }
                }

                // Handle specific error codes
                match error_message {
                    "PHONE_NUMBER_INVALID" => {
                        Self::InvalidPhoneNumber("Phone number format is invalid".into())
                    },
                    "PHONE_CODE_INVALID" | "PHONE_CODE_EXPIRED" | "PHONE_CODE_EMPTY" => {
                        Self::InvalidCode
                    },
                    "PASSWORD_HASH_INVALID" => Self::InvalidPassword,
                    "SESSION_PASSWORD_NEEDED" => Self::PasswordRequired,
                    "AUTH_KEY_UNREGISTERED" | "USER_DEACTIVATED" | "USER_DEACTIVATED_BAN" => {
                        Self::AuthRequired
                    },
                    _ => Self::Api(error_message.to_string()),
                }
            },
            InvocationError::Io(io_err) => Self::Network(io_err.to_string()),
            InvocationError::Dropped => Self::Internal("Request was dropped".into()),
            InvocationError::Deserialize(err) => {
                Self::Internal(format!("Deserialize error: {}", err))
            },
            InvocationError::Transport(err) => Self::Network(format!("Transport error: {}", err)),
            InvocationError::InvalidDc => Self::Internal("Invalid datacenter".into()),
            InvocationError::Authentication(err) => {
                Self::Internal(format!("Authentication error: {}", err))
            },
        }
    }
}

impl From<grammers_client::SignInError> for TelegramError {
    fn from(err: grammers_client::SignInError) -> Self {
        use grammers_client::SignInError;

        match err {
            SignInError::SignUpRequired => Self::SignUpRequired,
            SignInError::PasswordRequired(_) => Self::PasswordRequired,
            SignInError::InvalidCode => Self::InvalidCode,
            SignInError::InvalidPassword(_) => Self::InvalidPassword,
            SignInError::Other(invocation_error) => invocation_error.into(),
        }
    }
}

impl From<std::io::Error> for TelegramError {
    fn from(err: std::io::Error) -> Self {
        Self::Io(err.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_recoverable() {
        assert!(TelegramError::Network("timeout".into()).is_recoverable());
        assert!(TelegramError::Timeout.is_recoverable());
        assert!(TelegramError::FloodWait(30).is_recoverable());

        assert!(!TelegramError::InvalidCode.is_recoverable());
        assert!(!TelegramError::AuthRequired.is_recoverable());
    }

    #[test]
    fn test_requires_user_action() {
        assert!(TelegramError::AuthRequired.requires_user_action());
        assert!(TelegramError::InvalidCode.requires_user_action());
        assert!(TelegramError::InvalidPassword.requires_user_action());
        assert!(TelegramError::PasswordRequired.requires_user_action());

        assert!(!TelegramError::Network("timeout".into()).requires_user_action());
        assert!(!TelegramError::Timeout.requires_user_action());
    }
}
