//! Key bindings system for Ithil.
//!
//! This module provides a configurable key binding system that supports
//! both standard and Vim-style key mappings. Actions are decoupled from
//! specific key combinations, allowing users to customize their experience.
//!
//! # Architecture
//!
//! The system consists of:
//! - [`Action`]: An enum representing all possible user actions
//! - [`KeyMap`]: A mapping from [`KeyEvent`] to [`Action`]
//!
//! # Vim Mode
//!
//! When vim mode is enabled, navigation uses `h/j/k/l` instead of arrow keys,
//! and several actions use single-letter commands (`i` for input, `/` for search).
//!
//! # Example
//!
//! ```rust
//! use ithil::ui::keys::{KeyMap, Action};
//! use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
//!
//! let keymap = KeyMap::new(true); // vim mode
//!
//! let key = KeyEvent::new(KeyCode::Char('j'), KeyModifiers::NONE);
//! assert_eq!(keymap.get_action(&key), Some(Action::Down));
//! ```

use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use std::collections::HashMap;

/// Actions that can be triggered by key bindings.
///
/// These actions are abstract and decoupled from specific key combinations,
/// allowing the same action to be bound to different keys.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Action {
    // =========================================================================
    // Global Actions
    // =========================================================================
    /// Exit the application
    Quit,
    /// Toggle help overlay
    Help,
    /// Move focus to the next pane
    NextPane,
    /// Move focus to the previous pane
    PreviousPane,
    /// Focus the chat list pane
    FocusChatList,
    /// Focus the conversation pane
    FocusConversation,
    /// Focus the sidebar pane
    FocusSidebar,
    /// Toggle sidebar visibility
    ToggleSidebar,
    /// Open settings screen
    OpenSettings,

    // =========================================================================
    // Navigation Actions
    // =========================================================================
    /// Navigate up
    Up,
    /// Navigate down
    Down,
    /// Navigate left
    Left,
    /// Navigate right
    Right,
    /// Page up (large scroll)
    PageUp,
    /// Page down (large scroll)
    PageDown,
    /// Go to the beginning
    Home,
    /// Go to the end
    End,

    // =========================================================================
    // Chat List Actions
    // =========================================================================
    /// Open the selected chat
    OpenChat,
    /// Search chats
    SearchChats,
    /// Pin/unpin the selected chat
    PinChat,
    /// Mute/unmute the selected chat
    MuteChat,
    /// Archive the selected chat
    ArchiveChat,
    /// Mark the selected chat as read
    MarkAsRead,

    // =========================================================================
    // Conversation Actions
    // =========================================================================
    /// Focus the message input field
    FocusInput,
    /// Send the current message
    SendMessage,
    /// Insert a new line in the input
    NewLine,
    /// Reply to the selected message
    Reply,
    /// Edit the selected message
    Edit,
    /// Delete the selected message
    Delete,
    /// Forward the selected message
    Forward,
    /// Cancel the current action
    CancelAction,
    /// Open/view media (photo, video, document)
    OpenMedia,

    // =========================================================================
    // Input Actions
    // =========================================================================
    /// Delete character before cursor
    Backspace,
    /// Delete character at cursor
    DeleteChar,

    // =========================================================================
    // Scroll Actions
    // =========================================================================
    /// Scroll up one line
    ScrollUp,
    /// Scroll down one line
    ScrollDown,

    // =========================================================================
    // Settings Actions
    // =========================================================================
    /// Save settings to config file
    SaveSettings,
}

impl std::fmt::Display for Action {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Quit => write!(f, "Quit"),
            Self::Help => write!(f, "Help"),
            Self::NextPane => write!(f, "Next Pane"),
            Self::PreviousPane => write!(f, "Previous Pane"),
            Self::FocusChatList => write!(f, "Focus Chat List"),
            Self::FocusConversation => write!(f, "Focus Conversation"),
            Self::FocusSidebar => write!(f, "Focus Sidebar"),
            Self::ToggleSidebar => write!(f, "Toggle Sidebar"),
            Self::OpenSettings => write!(f, "Open Settings"),
            Self::Up => write!(f, "Up"),
            Self::Down => write!(f, "Down"),
            Self::Left => write!(f, "Left"),
            Self::Right => write!(f, "Right"),
            Self::PageUp => write!(f, "Page Up"),
            Self::PageDown => write!(f, "Page Down"),
            Self::Home => write!(f, "Home"),
            Self::End => write!(f, "End"),
            Self::OpenChat => write!(f, "Open Chat"),
            Self::SearchChats => write!(f, "Search Chats"),
            Self::PinChat => write!(f, "Pin Chat"),
            Self::MuteChat => write!(f, "Mute Chat"),
            Self::ArchiveChat => write!(f, "Archive Chat"),
            Self::MarkAsRead => write!(f, "Mark As Read"),
            Self::FocusInput => write!(f, "Focus Input"),
            Self::SendMessage => write!(f, "Send Message"),
            Self::NewLine => write!(f, "New Line"),
            Self::Reply => write!(f, "Reply"),
            Self::Edit => write!(f, "Edit"),
            Self::Delete => write!(f, "Delete"),
            Self::Forward => write!(f, "Forward"),
            Self::CancelAction => write!(f, "Cancel"),
            Self::OpenMedia => write!(f, "Open Media"),
            Self::Backspace => write!(f, "Backspace"),
            Self::DeleteChar => write!(f, "Delete Char"),
            Self::ScrollUp => write!(f, "Scroll Up"),
            Self::ScrollDown => write!(f, "Scroll Down"),
            Self::SaveSettings => write!(f, "Save Settings"),
        }
    }
}

/// Key binding configuration.
///
/// Maps [`KeyEvent`]s to [`Action`]s, supporting both standard and Vim-style
/// key bindings based on configuration.
#[derive(Debug, Clone)]
pub struct KeyMap {
    bindings: HashMap<KeyEvent, Action>,
    vim_mode: bool,
}

impl Default for KeyMap {
    fn default() -> Self {
        Self::new(false)
    }
}

impl KeyMap {
    /// Create a new key map with the specified mode.
    ///
    /// # Arguments
    ///
    /// * `vim_mode` - If `true`, enables Vim-style key bindings (h/j/k/l navigation)
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::keys::KeyMap;
    ///
    /// let standard = KeyMap::new(false);
    /// let vim = KeyMap::new(true);
    /// ```
    #[must_use]
    pub fn new(vim_mode: bool) -> Self {
        let mut bindings = HashMap::new();

        // =====================================================================
        // Global bindings (same for both modes)
        // =====================================================================
        bindings.insert(key(KeyCode::Char('q'), ctrl()), Action::Quit);
        bindings.insert(key(KeyCode::Char('c'), ctrl()), Action::Quit);
        bindings.insert(key(KeyCode::Char('?'), none()), Action::Help);
        bindings.insert(key(KeyCode::Tab, none()), Action::NextPane);
        bindings.insert(key(KeyCode::BackTab, shift()), Action::PreviousPane);
        bindings.insert(key(KeyCode::Char('1'), ctrl()), Action::FocusChatList);
        bindings.insert(key(KeyCode::Char('2'), ctrl()), Action::FocusConversation);
        bindings.insert(key(KeyCode::Char('3'), ctrl()), Action::FocusSidebar);
        bindings.insert(key(KeyCode::Char('s'), ctrl()), Action::ToggleSidebar);
        bindings.insert(key(KeyCode::Char(','), ctrl()), Action::OpenSettings);
        bindings.insert(key(KeyCode::F(12), none()), Action::OpenSettings);

        // =====================================================================
        // Arrow key navigation (both modes)
        // =====================================================================
        bindings.insert(key(KeyCode::Up, none()), Action::Up);
        bindings.insert(key(KeyCode::Down, none()), Action::Down);
        bindings.insert(key(KeyCode::Left, none()), Action::Left);
        bindings.insert(key(KeyCode::Right, none()), Action::Right);
        bindings.insert(key(KeyCode::PageUp, none()), Action::PageUp);
        bindings.insert(key(KeyCode::PageDown, none()), Action::PageDown);
        bindings.insert(key(KeyCode::Home, none()), Action::Home);
        bindings.insert(key(KeyCode::End, none()), Action::End);

        // =====================================================================
        // Common actions (both modes)
        // =====================================================================
        bindings.insert(key(KeyCode::Enter, none()), Action::OpenChat);
        bindings.insert(key(KeyCode::Enter, shift()), Action::NewLine);
        bindings.insert(key(KeyCode::Esc, none()), Action::CancelAction);
        bindings.insert(key(KeyCode::Backspace, none()), Action::Backspace);
        bindings.insert(key(KeyCode::Delete, none()), Action::DeleteChar);

        // =====================================================================
        // Mode-specific bindings
        // =====================================================================
        if vim_mode {
            Self::add_vim_bindings(&mut bindings);
        } else {
            Self::add_standard_bindings(&mut bindings);
        }

        Self { bindings, vim_mode }
    }

    /// Add Vim-style key bindings.
    fn add_vim_bindings(bindings: &mut HashMap<KeyEvent, Action>) {
        // Navigation
        bindings.insert(key(KeyCode::Char('j'), none()), Action::Down);
        bindings.insert(key(KeyCode::Char('k'), none()), Action::Up);
        bindings.insert(key(KeyCode::Char('h'), none()), Action::Left);
        bindings.insert(key(KeyCode::Char('l'), none()), Action::Right);
        bindings.insert(key(KeyCode::Char('g'), none()), Action::Home);
        bindings.insert(key(KeyCode::Char('G'), shift()), Action::End);
        bindings.insert(key(KeyCode::Char('u'), ctrl()), Action::PageUp);
        bindings.insert(key(KeyCode::Char('d'), ctrl()), Action::PageDown);

        // Actions
        bindings.insert(key(KeyCode::Char('i'), none()), Action::FocusInput);
        bindings.insert(key(KeyCode::Char('/'), none()), Action::SearchChats);
        bindings.insert(key(KeyCode::Char('p'), none()), Action::PinChat);
        bindings.insert(key(KeyCode::Char('m'), none()), Action::MuteChat);
        bindings.insert(key(KeyCode::Char('a'), none()), Action::ArchiveChat);
        bindings.insert(key(KeyCode::Char('r'), none()), Action::Reply);
        bindings.insert(key(KeyCode::Char('e'), none()), Action::Edit);
        bindings.insert(key(KeyCode::Char('x'), none()), Action::Delete);
        bindings.insert(key(KeyCode::Char('f'), none()), Action::Forward);
        bindings.insert(key(KeyCode::Char('o'), none()), Action::OpenMedia);
    }

    /// Add standard key bindings.
    fn add_standard_bindings(bindings: &mut HashMap<KeyEvent, Action>) {
        bindings.insert(key(KeyCode::Char('f'), ctrl()), Action::SearchChats);
        bindings.insert(key(KeyCode::Char('r'), ctrl()), Action::Reply);
        bindings.insert(key(KeyCode::Char('e'), ctrl()), Action::Edit);
        bindings.insert(key(KeyCode::Char('o'), ctrl()), Action::OpenMedia);
        bindings.insert(key(KeyCode::F(5), none()), Action::MarkAsRead);
        bindings.insert(key(KeyCode::F(2), none()), Action::PinChat);
        bindings.insert(key(KeyCode::F(3), none()), Action::MuteChat);
    }

    /// Get the action for a key event.
    ///
    /// Returns `None` if no action is bound to the key.
    ///
    /// # Arguments
    ///
    /// * `key` - The key event to look up
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::keys::{KeyMap, Action};
    /// use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
    ///
    /// let keymap = KeyMap::new(false);
    /// let key = KeyEvent::new(KeyCode::Tab, KeyModifiers::NONE);
    /// assert_eq!(keymap.get_action(&key), Some(Action::NextPane));
    /// ```
    #[must_use]
    pub fn get_action(&self, key: &KeyEvent) -> Option<Action> {
        // Normalize the key event to ignore release events and other fields
        let normalized = KeyEvent::new(key.code, key.modifiers);
        self.bindings.get(&normalized).copied()
    }

    /// Check if Vim mode is enabled.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::keys::KeyMap;
    ///
    /// let vim = KeyMap::new(true);
    /// assert!(vim.is_vim_mode());
    ///
    /// let standard = KeyMap::new(false);
    /// assert!(!standard.is_vim_mode());
    /// ```
    #[must_use]
    pub const fn is_vim_mode(&self) -> bool {
        self.vim_mode
    }

    /// Get help text for the current key bindings.
    ///
    /// Returns a vector of (key description, action description) tuples
    /// suitable for display in a help overlay.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::keys::KeyMap;
    ///
    /// let keymap = KeyMap::new(true);
    /// for (key, desc) in keymap.get_help_text() {
    ///     println!("{:12} {}", key, desc);
    /// }
    /// ```
    #[must_use]
    pub fn get_help_text(&self) -> Vec<(&'static str, &'static str)> {
        if self.vim_mode {
            vec![
                ("j/k", "Navigate up/down"),
                ("h/l", "Navigate left/right"),
                ("g/G", "Go to start/end"),
                ("Enter", "Open chat / Edit value"),
                ("i", "Focus input"),
                ("/", "Search"),
                ("r", "Reply"),
                ("e", "Edit"),
                ("x", "Delete"),
                ("f", "Forward"),
                ("o", "Open media"),
                ("p", "Pin/unpin"),
                ("m", "Mute/unmute"),
                ("Tab", "Next pane"),
                ("Shift+Tab", "Previous pane"),
                ("Ctrl+S", "Toggle sidebar / Save"),
                ("Ctrl+,/F12", "Open settings"),
                ("S", "Toggle stealth mode"),
                ("?", "Toggle help"),
                ("Esc", "Back / Cancel"),
                ("Ctrl+Q", "Quit"),
            ]
        } else {
            vec![
                ("↑/↓", "Navigate up/down"),
                ("←/→", "Navigate left/right"),
                ("Home/End", "Go to start/end"),
                ("Enter", "Open / Edit value"),
                ("Ctrl+F", "Search"),
                ("Ctrl+R", "Reply"),
                ("Ctrl+E", "Edit"),
                ("Ctrl+O", "Open media"),
                ("F2", "Pin/unpin"),
                ("F3", "Mute/unmute"),
                ("F5", "Mark as read"),
                ("Tab", "Next pane"),
                ("Shift+Tab", "Previous pane"),
                ("Ctrl+S", "Toggle sidebar / Save"),
                ("Ctrl+,/F12", "Open settings"),
                ("S", "Toggle stealth mode"),
                ("?", "Toggle help"),
                ("Esc", "Back / Cancel"),
                ("Ctrl+Q", "Quit"),
            ]
        }
    }

    /// Get all key bindings as a vector.
    ///
    /// Returns a vector of (key event, action) tuples.
    #[must_use]
    pub fn all_bindings(&self) -> Vec<(KeyEvent, Action)> {
        self.bindings.iter().map(|(k, a)| (*k, *a)).collect()
    }

    /// Check if a key is bound to any action.
    #[must_use]
    pub fn is_bound(&self, key: &KeyEvent) -> bool {
        let normalized = KeyEvent::new(key.code, key.modifiers);
        self.bindings.contains_key(&normalized)
    }
}

// =============================================================================
// Helper Functions
// =============================================================================

/// Create a key event with the specified code and modifiers.
#[inline]
const fn key(code: KeyCode, modifiers: KeyModifiers) -> KeyEvent {
    KeyEvent::new(code, modifiers)
}

/// No modifiers.
#[inline]
const fn none() -> KeyModifiers {
    KeyModifiers::NONE
}

/// Control modifier.
#[inline]
const fn ctrl() -> KeyModifiers {
    KeyModifiers::CONTROL
}

/// Shift modifier.
#[inline]
const fn shift() -> KeyModifiers {
    KeyModifiers::SHIFT
}

/// Alt modifier (unused but available for future use).
#[inline]
#[allow(dead_code)]
const fn alt() -> KeyModifiers {
    KeyModifiers::ALT
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_standard_mode() {
        let keymap = KeyMap::new(false);
        assert!(!keymap.is_vim_mode());
    }

    #[test]
    fn test_vim_mode() {
        let keymap = KeyMap::new(true);
        assert!(keymap.is_vim_mode());
    }

    #[test]
    fn test_global_bindings_same_in_both_modes() {
        let standard = KeyMap::new(false);
        let vim = KeyMap::new(true);

        // Ctrl+Q should quit in both modes
        let quit_key = KeyEvent::new(KeyCode::Char('q'), KeyModifiers::CONTROL);
        assert_eq!(standard.get_action(&quit_key), Some(Action::Quit));
        assert_eq!(vim.get_action(&quit_key), Some(Action::Quit));

        // Tab should switch panes in both modes
        let tab_key = KeyEvent::new(KeyCode::Tab, KeyModifiers::NONE);
        assert_eq!(standard.get_action(&tab_key), Some(Action::NextPane));
        assert_eq!(vim.get_action(&tab_key), Some(Action::NextPane));
    }

    #[test]
    fn test_vim_navigation() {
        let keymap = KeyMap::new(true);

        let j = KeyEvent::new(KeyCode::Char('j'), KeyModifiers::NONE);
        let k = KeyEvent::new(KeyCode::Char('k'), KeyModifiers::NONE);
        let h = KeyEvent::new(KeyCode::Char('h'), KeyModifiers::NONE);
        let l = KeyEvent::new(KeyCode::Char('l'), KeyModifiers::NONE);

        assert_eq!(keymap.get_action(&j), Some(Action::Down));
        assert_eq!(keymap.get_action(&k), Some(Action::Up));
        assert_eq!(keymap.get_action(&h), Some(Action::Left));
        assert_eq!(keymap.get_action(&l), Some(Action::Right));
    }

    #[test]
    fn test_arrow_keys_work_in_both_modes() {
        let standard = KeyMap::new(false);
        let vim = KeyMap::new(true);

        let up = KeyEvent::new(KeyCode::Up, KeyModifiers::NONE);
        let down = KeyEvent::new(KeyCode::Down, KeyModifiers::NONE);

        assert_eq!(standard.get_action(&up), Some(Action::Up));
        assert_eq!(standard.get_action(&down), Some(Action::Down));
        assert_eq!(vim.get_action(&up), Some(Action::Up));
        assert_eq!(vim.get_action(&down), Some(Action::Down));
    }

    #[test]
    fn test_unbound_key_returns_none() {
        let keymap = KeyMap::new(false);

        let unbound = KeyEvent::new(KeyCode::Char('z'), KeyModifiers::NONE);
        assert_eq!(keymap.get_action(&unbound), None);
    }

    #[test]
    fn test_help_text_different_for_modes() {
        let standard = KeyMap::new(false);
        let vim = KeyMap::new(true);

        let standard_help = standard.get_help_text();
        let vim_help = vim.get_help_text();

        // Vim mode should have j/k navigation
        assert!(vim_help.iter().any(|(k, _)| k.contains("j/k")));

        // Standard mode should have arrow key navigation
        assert!(standard_help.iter().any(|(k, _)| k.contains("↑/↓")));
    }

    #[test]
    fn test_is_bound() {
        let keymap = KeyMap::new(false);

        let bound = KeyEvent::new(KeyCode::Tab, KeyModifiers::NONE);
        let unbound = KeyEvent::new(KeyCode::Char('z'), KeyModifiers::NONE);

        assert!(keymap.is_bound(&bound));
        assert!(!keymap.is_bound(&unbound));
    }

    #[test]
    fn test_action_display() {
        assert_eq!(format!("{}", Action::Quit), "Quit");
        assert_eq!(format!("{}", Action::NextPane), "Next Pane");
        assert_eq!(format!("{}", Action::FocusInput), "Focus Input");
    }

    #[test]
    fn test_default_keymap() {
        let keymap = KeyMap::default();
        assert!(!keymap.is_vim_mode());
    }
}
