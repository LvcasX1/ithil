//! Ithil - A Terminal User Interface for Telegram
//!
//! This library provides the core functionality for the Ithil TUI client,
//! including configuration management, UI components, and Telegram integration.
//!
//! # Modules
//!
//! - [`app`]: Application-level functionality including configuration and credentials
//! - [`cache`]: Thread-safe in-memory cache for Telegram data
//! - [`telegram`]: Telegram client wrapper using grammers for MTProto communication
//! - [`types`]: Core domain types (User, Chat, Message, etc.)
//! - [`ui`]: User interface components and rendering
//!
//! # Quick Start
//!
//! ```rust,no_run
//! use ithil::telegram::TelegramClient;
//! use ithil::cache::new_shared_cache;
//!
//! # async fn example() -> Result<(), Box<dyn std::error::Error>> {
//! // Create a shared cache
//! let cache = new_shared_cache(100);
//!
//! // Create and connect the Telegram client
//! let client = TelegramClient::new(
//!     12345,                    // API ID from my.telegram.org
//!     "api_hash".to_string(),   // API Hash
//!     "session.session".into(), // Session file path
//!     cache,
//! );
//!
//! client.connect().await?;
//!
//! // Check if authentication is needed
//! if client.get_auth_state().await == ithil::types::AuthState::Ready {
//!     // Already authenticated - fetch dialogs
//!     let chats = client.get_dialogs().await?;
//!     println!("Found {} chats", chats.len());
//! }
//! # Ok(())
//! # }
//! ```

#![warn(clippy::all, clippy::pedantic, clippy::nursery)]
#![allow(clippy::module_name_repetitions)]

pub mod app;
pub mod cache;
pub mod telegram;
pub mod types;
pub mod ui;
pub mod utils;
