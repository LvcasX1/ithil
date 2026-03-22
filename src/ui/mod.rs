//! User interface module for Ithil.
//!
//! This module provides the TUI components and rendering logic
//! using Ratatui and crossterm.
//!
//! # Modules
//!
//! - [`app`]: Main application state machine and rendering
//! - [`components`]: Reusable UI components (input, auth, etc.)
//! - [`keys`]: Key bindings system with Vim/standard mode support
//! - [`styles`]: Theme-aware color palettes and pre-built styles
//!
//! # Quick Start
//!
//! ```rust,no_run
//! use ithil::ui::{App, AppState, FocusedPane};
//! use ithil::ui::{KeyMap, Action};
//! use ithil::ui::{colors, Styles};
//! use ithil::ui::components::{InputComponent, AuthModel};
//!
//! // Styles provide consistent theming
//! let border_style = Styles::border_focused();
//! let text_style = Styles::text();
//!
//! // KeyMap handles keyboard input
//! let keymap = KeyMap::new(true); // Vim mode
//!
//! // Input component for text entry
//! let input = InputComponent::new("Enter text...");
//! ```

pub mod app;
pub mod components;
pub mod keys;
pub mod styles;

pub use app::{App, AppAction, AppState, FocusedPane};
pub use components::{AuthAction, AuthModel, InputComponent};
pub use keys::{Action, KeyMap};
pub use styles::{colors, Styles, Theme};
