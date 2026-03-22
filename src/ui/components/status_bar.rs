//! Status bar component for Ithil.
//!
//! Displays connection status, current user information, and unread counts
//! at the bottom of the application window.
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::components::{StatusBar, StatusBarWidget, ConnectionStatus};
//!
//! let mut status = StatusBar::new();
//! status.set_connection_status(ConnectionStatus::Connected);
//! status.set_unread_count(5);
//!
//! // Render with StatusBarWidget::new(&status)
//! ```

use ratatui::{
    buffer::Buffer,
    layout::{Alignment, Constraint, Direction, Layout, Rect},
    text::{Line, Span},
    widgets::{Paragraph, Widget},
};

use crate::types::User;
use crate::ui::styles::Styles;

/// Connection status indicator.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum ConnectionStatus {
    /// Not connected to Telegram
    #[default]
    Disconnected,
    /// Attempting to establish connection
    Connecting,
    /// Successfully connected
    Connected,
    /// Connection lost, attempting to reconnect
    Reconnecting,
}

impl std::fmt::Display for ConnectionStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Disconnected => write!(f, "Disconnected"),
            Self::Connecting => write!(f, "Connecting"),
            Self::Connected => write!(f, "Connected"),
            Self::Reconnecting => write!(f, "Reconnecting"),
        }
    }
}

/// Status bar model containing the current application state.
///
/// The status bar is displayed at the bottom of the screen and shows:
/// - Connection status indicator (left)
/// - Current user name (left)
/// - Status message or app name (center)
/// - Unread message count (right)
/// - Vim mode indicator (right)
#[derive(Debug, Clone, Default)]
pub struct StatusBar {
    /// Current connection state
    pub connection_status: ConnectionStatus,
    /// Currently authenticated user
    pub current_user: Option<User>,
    /// Total unread message count across all chats
    pub total_unread: i32,
    /// Temporary status message to display
    pub status_message: Option<String>,
    /// Whether vim keybindings are active
    pub vim_mode: bool,
}

impl StatusBar {
    /// Creates a new status bar with default values.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::StatusBar;
    ///
    /// let status = StatusBar::new();
    /// assert_eq!(status.total_unread, 0);
    /// ```
    #[must_use]
    pub fn new() -> Self {
        Self::default()
    }

    /// Sets the connection status.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::{StatusBar, ConnectionStatus};
    ///
    /// let mut status = StatusBar::new();
    /// status.set_connection_status(ConnectionStatus::Connected);
    /// assert_eq!(status.connection_status, ConnectionStatus::Connected);
    /// ```
    pub fn set_connection_status(&mut self, status: ConnectionStatus) {
        self.connection_status = status;
    }

    /// Sets the current user.
    ///
    /// Pass `None` to clear the user (e.g., on logout).
    pub fn set_user(&mut self, user: Option<User>) {
        self.current_user = user;
    }

    /// Sets the total unread message count.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::StatusBar;
    ///
    /// let mut status = StatusBar::new();
    /// status.set_unread_count(42);
    /// assert_eq!(status.total_unread, 42);
    /// ```
    pub fn set_unread_count(&mut self, count: i32) {
        self.total_unread = count;
    }

    /// Sets a temporary status message.
    ///
    /// Pass `None` to clear the message and show the default app name.
    pub fn set_status_message(&mut self, msg: Option<String>) {
        self.status_message = msg;
    }

    /// Enables or disables vim mode indicator.
    pub fn set_vim_mode(&mut self, enabled: bool) {
        self.vim_mode = enabled;
    }
}

/// Widget for rendering the status bar.
///
/// This widget renders the status bar with three sections:
/// - Left: Connection indicator and user name
/// - Center: Status message or app name
/// - Right: Unread count and vim mode indicator
pub struct StatusBarWidget<'a> {
    model: &'a StatusBar,
}

impl<'a> StatusBarWidget<'a> {
    /// Creates a new status bar widget.
    ///
    /// # Arguments
    ///
    /// * `model` - Reference to the status bar model
    #[must_use]
    pub const fn new(model: &'a StatusBar) -> Self {
        Self { model }
    }
}

impl Widget for StatusBarWidget<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        // Apply status bar background style to the entire area
        buf.set_style(area, Styles::status_bar());

        // Split into left, center, right sections
        let chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(30), // Left: connection + user
                Constraint::Percentage(40), // Center: status message
                Constraint::Percentage(30), // Right: unread + vim mode
            ])
            .split(area);

        // Left section: connection status indicator + user name
        let (conn_icon, conn_style) = match self.model.connection_status {
            ConnectionStatus::Connected => ("●", Styles::status_online()),
            ConnectionStatus::Connecting => ("◐", Styles::warning()),
            ConnectionStatus::Reconnecting => ("↻", Styles::warning()),
            ConnectionStatus::Disconnected => ("○", Styles::status_offline()),
        };

        let user_name = self
            .model
            .current_user
            .as_ref()
            .map(User::get_display_name)
            .unwrap_or_default();

        let left = Line::from(vec![
            Span::raw(" "),
            Span::styled(conn_icon, conn_style),
            Span::raw(" "),
            Span::styled(user_name, Styles::text()),
        ]);
        Paragraph::new(left).render(chunks[0], buf);

        // Center section: status message or key hints
        let center_text = self
            .model
            .status_message
            .as_deref()
            .unwrap_or("? Help  Ctrl+P Settings");
        let center = Line::from(vec![Span::styled(center_text, Styles::text_muted())]);
        Paragraph::new(center)
            .alignment(Alignment::Center)
            .render(chunks[1], buf);

        // Right section: unread count + vim mode + version
        let mut right_spans = Vec::new();

        if self.model.total_unread > 0 {
            right_spans.push(Span::styled(
                format!("[{}] ", self.model.total_unread),
                Styles::chat_unread(),
            ));
        }

        if self.model.vim_mode {
            right_spans.push(Span::styled("[VIM] ", Styles::text_accent()));
        }

        right_spans.push(Span::styled(
            concat!("v", env!("CARGO_PKG_VERSION")),
            Styles::text_muted(),
        ));
        right_spans.push(Span::raw(" "));

        let right = Line::from(right_spans);
        Paragraph::new(right)
            .alignment(Alignment::Right)
            .render(chunks[2], buf);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_status_bar_default() {
        let status = StatusBar::new();
        assert_eq!(status.connection_status, ConnectionStatus::Disconnected);
        assert!(status.current_user.is_none());
        assert_eq!(status.total_unread, 0);
        assert!(status.status_message.is_none());
        assert!(!status.vim_mode);
    }

    #[test]
    fn test_set_connection_status() {
        let mut status = StatusBar::new();

        status.set_connection_status(ConnectionStatus::Connecting);
        assert_eq!(status.connection_status, ConnectionStatus::Connecting);

        status.set_connection_status(ConnectionStatus::Connected);
        assert_eq!(status.connection_status, ConnectionStatus::Connected);

        status.set_connection_status(ConnectionStatus::Reconnecting);
        assert_eq!(status.connection_status, ConnectionStatus::Reconnecting);

        status.set_connection_status(ConnectionStatus::Disconnected);
        assert_eq!(status.connection_status, ConnectionStatus::Disconnected);
    }

    #[test]
    fn test_set_user() {
        let mut status = StatusBar::new();
        assert!(status.current_user.is_none());

        let user = User {
            first_name: "John".to_string(),
            last_name: "Doe".to_string(),
            ..Default::default()
        };
        status.set_user(Some(user));
        assert!(status.current_user.is_some());
        assert_eq!(
            status.current_user.as_ref().unwrap().get_display_name(),
            "John Doe"
        );

        status.set_user(None);
        assert!(status.current_user.is_none());
    }

    #[test]
    fn test_set_unread_count() {
        let mut status = StatusBar::new();
        assert_eq!(status.total_unread, 0);

        status.set_unread_count(10);
        assert_eq!(status.total_unread, 10);

        status.set_unread_count(0);
        assert_eq!(status.total_unread, 0);
    }

    #[test]
    fn test_set_status_message() {
        let mut status = StatusBar::new();
        assert!(status.status_message.is_none());

        status.set_status_message(Some("Loading...".to_string()));
        assert_eq!(status.status_message.as_deref(), Some("Loading..."));

        status.set_status_message(None);
        assert!(status.status_message.is_none());
    }

    #[test]
    fn test_set_vim_mode() {
        let mut status = StatusBar::new();
        assert!(!status.vim_mode);

        status.set_vim_mode(true);
        assert!(status.vim_mode);

        status.set_vim_mode(false);
        assert!(!status.vim_mode);
    }

    #[test]
    fn test_connection_status_display() {
        assert_eq!(
            format!("{}", ConnectionStatus::Disconnected),
            "Disconnected"
        );
        assert_eq!(format!("{}", ConnectionStatus::Connecting), "Connecting");
        assert_eq!(format!("{}", ConnectionStatus::Connected), "Connected");
        assert_eq!(
            format!("{}", ConnectionStatus::Reconnecting),
            "Reconnecting"
        );
    }

    #[test]
    fn test_status_bar_widget_creation() {
        let status = StatusBar::new();
        let _widget = StatusBarWidget::new(&status);
        // Widget creation should not panic
    }
}
