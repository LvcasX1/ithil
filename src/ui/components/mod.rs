//! Reusable UI components for Ithil.
//!
//! This module provides stateful, reusable UI elements that can be
//! composed to build the application interface.
//!
//! # Components
//!
//! - [`InputComponent`]: Text input field with cursor handling
//! - [`AuthModel`]: Authentication flow UI (phone, code, password)
//! - [`ChatItemComponent`]: Single chat entry in the chat list
//! - [`ChatListModel`]: Chat list pane with selection and search
//! - [`ConversationModel`]: Conversation view with message list and input
//! - [`MessageWidget`]: Individual message rendering
//! - [`SidebarModel`]: Info panel showing chat details
//! - [`SettingsModel`]: Application settings view
//! - [`StatusBar`]: Status bar showing connection and user info
//! - [`Modal`]: Generic modal dialog for confirmations and alerts
//! - [`HelpModal`]: Help overlay showing keyboard shortcuts
//!
//! # Design Pattern
//!
//! Components follow a Model-View-Update pattern:
//! - Struct holds state (model)
//! - `handle_input()` processes events (update)
//! - `render()` draws to the terminal (view)

mod auth;
mod chat_item;
mod chat_list;
pub mod conversation;
mod help_modal;
mod input;
pub mod message;
mod modal;
pub mod settings;
pub mod sidebar;
mod status_bar;

pub use auth::{AuthAction, AuthModel};
pub use chat_item::{ChatItemBuilder, ChatItemComponent, ChatItemConfig};
pub use chat_list::{ChatListAction, ChatListModel, ChatListState};
pub use conversation::{ConversationAction, ConversationModel, ConversationWidget, InputMode};
pub use help_modal::{HelpModal, HelpModalWidget};
pub use input::InputComponent;
pub use message::MessageWidget;
pub use modal::{Modal, ModalWidget};
pub use settings::{SettingsAction, SettingsModel, SettingsSection, SettingsWidget};
pub use sidebar::{SidebarModel, SidebarWidget};
pub use status_bar::{ConnectionStatus, StatusBar, StatusBarWidget};
