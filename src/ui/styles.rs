//! Terminal-native styles for Ithil.
//!
//! This module uses ANSI named colors so the application automatically inherits
//! the user's terminal color scheme (Tokyo Night, Dracula, Solarized, etc.).
//!
//! # Color Palette Structure
//!
//! - **Backgrounds**: Transparent (`Reset`) to inherit the terminal background
//! - **Foregrounds**: Default terminal fg, dimmed, and bright white
//! - **Accents**: Cyan and Blue for focus/highlights
//! - **Status**: Red, Yellow, Green, Magenta for semantic states
//!
//! # Usage
//!
//! ```rust,no_run
//! use ithil::ui::styles::{colors, Styles};
//! use ratatui::widgets::Paragraph;
//!
//! let text = Paragraph::new("Hello, world!")
//!     .style(Styles::text());
//!
//! let error = Paragraph::new("Error occurred!")
//!     .style(Styles::error());
//! ```

use ratatui::style::{Modifier, Style};

/// Terminal-native color constants using ANSI named colors.
///
/// These colors inherit from the user's terminal theme automatically.
pub mod colors {
    use ratatui::style::Color;

    // =========================================================================
    // Backgrounds
    // =========================================================================

    /// Primary background - transparent, inherits terminal background.
    pub const BG_PRIMARY: Color = Color::Reset;

    /// Elevated background - transparent, inherits terminal background.
    pub const BG_ELEVATED: Color = Color::Reset;

    /// Interactive background - used for selections and highlights.
    pub const BG_INTERACTIVE: Color = Color::DarkGray;

    // =========================================================================
    // Foregrounds
    // =========================================================================

    /// Primary text - default terminal foreground.
    pub const FG_PRIMARY: Color = Color::Reset;

    /// Muted text - secondary, de-emphasized content.
    pub const FG_MUTED: Color = Color::DarkGray;

    /// Bright text - emphasis, headlines.
    pub const FG_BRIGHT: Color = Color::White;

    // =========================================================================
    // Accents
    // =========================================================================

    /// Primary accent - focus states, highlights, interactive elements.
    pub const ACCENT_PRIMARY: Color = Color::Cyan;

    /// Secondary accent - usernames, keywords.
    pub const ACCENT_SECONDARY: Color = Color::Blue;

    // =========================================================================
    // Status/Semantic colors
    // =========================================================================

    /// Error states, destructive actions, offline indicators.
    pub const STATUS_ERROR: Color = Color::Red;

    /// Warning states and caution indicators.
    pub const STATUS_WARNING: Color = Color::Yellow;

    /// Attention-grabbing elements, pinned items.
    pub const STATUS_ATTENTION: Color = Color::Yellow;

    /// Success states, online indicators, outgoing messages.
    pub const STATUS_SUCCESS: Color = Color::Green;

    /// Decorative accents.
    pub const DECORATIVE: Color = Color::Magenta;
}

/// Pre-built styles for common UI elements.
///
/// This struct provides static methods that return configured [`Style`] objects
/// for consistent theming across the application.
///
/// # Example
///
/// ```rust,no_run
/// use ithil::ui::styles::Styles;
/// use ratatui::widgets::Paragraph;
///
/// let paragraph = Paragraph::new("Status: Online")
///     .style(Styles::success());
/// ```
pub struct Styles;

impl Styles {
    // =========================================================================
    // Text Styles
    // =========================================================================

    /// Standard text style.
    #[must_use]
    pub const fn text() -> Style {
        Style::new().fg(colors::FG_PRIMARY)
    }

    /// Muted text style for less important content.
    #[must_use]
    pub const fn text_muted() -> Style {
        Style::new()
            .fg(colors::FG_MUTED)
            .add_modifier(Modifier::DIM)
    }

    /// Bright text style for emphasis.
    #[must_use]
    pub const fn text_bright() -> Style {
        Style::new().fg(colors::FG_BRIGHT)
    }

    /// Accent text style.
    #[must_use]
    pub const fn text_accent() -> Style {
        Style::new().fg(colors::ACCENT_PRIMARY)
    }

    // =========================================================================
    // Status Styles
    // =========================================================================

    /// Success/positive style.
    #[must_use]
    pub const fn success() -> Style {
        Style::new().fg(colors::STATUS_SUCCESS)
    }

    /// Warning style.
    #[must_use]
    pub const fn warning() -> Style {
        Style::new().fg(colors::STATUS_WARNING)
    }

    /// Error style.
    #[must_use]
    pub const fn error() -> Style {
        Style::new().fg(colors::STATUS_ERROR)
    }

    /// Informational style.
    #[must_use]
    pub const fn info() -> Style {
        Style::new().fg(colors::ACCENT_SECONDARY)
    }

    // =========================================================================
    // Selection/Highlight Styles
    // =========================================================================

    /// Selected item style.
    #[must_use]
    pub const fn selected() -> Style {
        Style::new()
            .bg(colors::BG_INTERACTIVE)
            .fg(colors::FG_BRIGHT)
    }

    /// Highlight style with bold modifier.
    #[must_use]
    pub const fn highlight() -> Style {
        Style::new()
            .fg(colors::ACCENT_PRIMARY)
            .add_modifier(Modifier::BOLD)
    }

    // =========================================================================
    // Border Styles
    // =========================================================================

    /// Default border style.
    #[must_use]
    pub const fn border() -> Style {
        Style::new().fg(colors::FG_MUTED)
    }

    /// Focused border style.
    #[must_use]
    pub const fn border_focused() -> Style {
        Style::new().fg(colors::ACCENT_PRIMARY)
    }

    // =========================================================================
    // Chat List Styles
    // =========================================================================

    /// Unread chat style.
    #[must_use]
    pub const fn chat_unread() -> Style {
        Style::new()
            .fg(colors::ACCENT_PRIMARY)
            .add_modifier(Modifier::BOLD)
    }

    /// Pinned chat style.
    #[must_use]
    pub const fn chat_pinned() -> Style {
        Style::new().fg(colors::STATUS_ATTENTION)
    }

    /// Muted chat style.
    #[must_use]
    pub const fn chat_muted() -> Style {
        Style::new()
            .fg(colors::FG_MUTED)
            .add_modifier(Modifier::DIM)
    }

    // =========================================================================
    // Message Styles
    // =========================================================================

    /// Outgoing message style.
    #[must_use]
    pub const fn message_outgoing() -> Style {
        Style::new().fg(colors::STATUS_SUCCESS)
    }

    /// Incoming message style.
    #[must_use]
    pub const fn message_incoming() -> Style {
        Style::new().fg(colors::FG_PRIMARY)
    }

    /// System message style.
    #[must_use]
    pub const fn message_system() -> Style {
        Style::new()
            .fg(colors::FG_MUTED)
            .add_modifier(Modifier::ITALIC)
    }

    /// Timestamp style.
    #[must_use]
    pub const fn timestamp() -> Style {
        Style::new()
            .fg(colors::FG_MUTED)
            .add_modifier(Modifier::DIM)
    }

    /// Username style.
    #[must_use]
    pub const fn username() -> Style {
        Style::new()
            .fg(colors::ACCENT_SECONDARY)
            .add_modifier(Modifier::BOLD)
    }

    // =========================================================================
    // Status Bar Styles
    // =========================================================================

    /// Status bar background style.
    #[must_use]
    pub const fn status_bar() -> Style {
        Style::new()
            .bg(colors::BG_INTERACTIVE)
            .fg(colors::FG_PRIMARY)
    }

    /// Online status indicator style.
    #[must_use]
    pub const fn status_online() -> Style {
        Style::new().fg(colors::STATUS_SUCCESS)
    }

    /// Offline status indicator style.
    #[must_use]
    pub const fn status_offline() -> Style {
        Style::new().fg(colors::STATUS_ERROR)
    }

    // =========================================================================
    // Input Styles
    // =========================================================================

    /// Input field text style.
    #[must_use]
    pub const fn input() -> Style {
        Style::new().fg(colors::FG_PRIMARY)
    }

    /// Input cursor style.
    #[must_use]
    pub const fn input_cursor() -> Style {
        Style::new()
            .fg(colors::BG_PRIMARY)
            .bg(colors::ACCENT_PRIMARY)
    }

    /// Input placeholder style.
    #[must_use]
    pub const fn input_placeholder() -> Style {
        Style::new()
            .fg(colors::FG_MUTED)
            .add_modifier(Modifier::DIM)
    }

    // =========================================================================
    // Modal/Dialog Styles
    // =========================================================================

    /// Modal background style.
    #[must_use]
    pub const fn modal_background() -> Style {
        Style::new().bg(colors::BG_INTERACTIVE)
    }

    /// Modal title style.
    #[must_use]
    pub const fn modal_title() -> Style {
        Style::new()
            .fg(colors::FG_BRIGHT)
            .add_modifier(Modifier::BOLD)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use ratatui::style::Color;

    #[test]
    fn test_colors_are_ansi() {
        assert_eq!(colors::BG_PRIMARY, Color::Reset);
        assert_eq!(colors::FG_PRIMARY, Color::Reset);
        assert_eq!(colors::ACCENT_PRIMARY, Color::Cyan);
        assert_eq!(colors::STATUS_ERROR, Color::Red);
        assert_eq!(colors::STATUS_SUCCESS, Color::Green);
    }

    #[test]
    fn test_text_style() {
        let style = Styles::text();
        assert_eq!(style.fg, Some(colors::FG_PRIMARY));
    }

    #[test]
    fn test_selected_style() {
        let style = Styles::selected();
        assert_eq!(style.fg, Some(colors::FG_BRIGHT));
        assert_eq!(style.bg, Some(colors::BG_INTERACTIVE));
    }

    #[test]
    fn test_highlight_style_has_bold() {
        let style = Styles::highlight();
        assert!(style.add_modifier.contains(Modifier::BOLD));
    }

    #[test]
    fn test_status_styles() {
        assert_eq!(Styles::success().fg, Some(colors::STATUS_SUCCESS));
        assert_eq!(Styles::warning().fg, Some(colors::STATUS_WARNING));
        assert_eq!(Styles::error().fg, Some(colors::STATUS_ERROR));
        assert_eq!(Styles::info().fg, Some(colors::ACCENT_SECONDARY));
    }

    #[test]
    fn test_border_styles() {
        assert_eq!(Styles::border().fg, Some(colors::FG_MUTED));
        assert_eq!(Styles::border_focused().fg, Some(colors::ACCENT_PRIMARY));
    }

    #[test]
    fn test_muted_styles_have_dim() {
        assert!(Styles::text_muted().add_modifier.contains(Modifier::DIM));
        assert!(Styles::timestamp().add_modifier.contains(Modifier::DIM));
    }
}
