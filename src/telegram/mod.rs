//! Telegram client module providing `MTProto` communication via grammers.
//!
//! This module wraps the grammers library to provide a high-level interface
//! for Telegram operations including authentication, chat management,
//! message handling, and real-time update streaming.
//!
//! # Architecture
//!
//! The [`TelegramClient`] is the main entry point. It manages:
//! - Connection lifecycle (connect/disconnect)
//! - Authentication flow (phone → code → optional 2FA password)
//! - Dialog/chat operations
//! - Message sending and history retrieval
//! - Real-time update streaming to the UI via tokio channels
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::telegram::TelegramClient;
//! use ithil::cache::new_shared_cache;
//!
//! # async fn example() -> Result<(), Box<dyn std::error::Error>> {
//! let cache = new_shared_cache(100);
//! let client = TelegramClient::new(
//!     12345,                    // API ID
//!     "your_api_hash".into(),   // API Hash
//!     "session.session".into(), // Session file path
//!     cache,
//! );
//!
//! client.connect().await?;
//!
//! if client.get_auth_state().await == ithil::types::AuthState::WaitPhoneNumber {
//!     client.request_login_code("+1234567890").await?;
//!     // ... get code from user ...
//!     client.sign_in("12345").await?;
//! }
//! # Ok(())
//! # }
//! ```

pub mod auth;
pub mod chats;
pub mod client;
pub mod error;
pub mod media;
pub mod messages;
pub mod updates;

pub use client::TelegramClient;
pub use error::TelegramError;
