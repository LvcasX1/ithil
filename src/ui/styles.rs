//! Nord color theme styles for Ithil.
//!
//! This module provides the complete Nord color palette and pre-built styles
//! for common UI elements. The Nord theme is a cold, bluish color palette
//! that provides excellent readability in terminal environments.
//!
//! # Color Palette Structure
//!
//! - **Polar Night** (NORD0-3): Dark backgrounds, ranging from darkest to lightest
//! - **Snow Storm** (NORD4-6): Light foreground colors for text
//! - **Frost** (NORD7-10): Blue/cyan accent colors
//! - **Aurora** (NORD11-15): Status/semantic colors (red, orange, yellow, green, purple)
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

/// Nord color palette constants.
///
/// See <https://www.nordtheme.com/docs/colors-and-palettes> for reference.
pub mod colors {
    use ratatui::style::Color;

    // =========================================================================
    // Polar Night - Dark backgrounds
    // =========================================================================

    /// NORD0 - Darkest background (#2E3440)
    ///
    /// Used as the main background color for the application.
    pub const NORD0: Color = Color::Rgb(46, 52, 64);

    /// NORD1 - Slightly lighter dark (#3B4252)
    ///
    /// Used for elevated surfaces like status bars and secondary backgrounds.
    pub const NORD1: Color = Color::Rgb(59, 66, 82);

    /// NORD2 - Medium dark (#434C5E)
    ///
    /// Used for selection backgrounds and interactive states.
    pub const NORD2: Color = Color::Rgb(67, 76, 94);

    /// NORD3 - Lightest dark (#4C566A)
    ///
    /// Used for comments, muted text, and subtle borders.
    pub const NORD3: Color = Color::Rgb(76, 86, 106);

    // =========================================================================
    // Snow Storm - Light foregrounds
    // =========================================================================

    /// NORD4 - Dark white (#D8DEE9)
    ///
    /// Primary text color for most content.
    pub const NORD4: Color = Color::Rgb(216, 222, 233);

    /// NORD5 - Medium white (#E5E9F0)
    ///
    /// Secondary text color for slightly brighter content.
    pub const NORD5: Color = Color::Rgb(229, 233, 240);

    /// NORD6 - Brightest white (#ECEFF4)
    ///
    /// Used for headlines, important text, and emphasis.
    pub const NORD6: Color = Color::Rgb(236, 239, 244);

    // =========================================================================
    // Frost - Blue/Cyan accents
    // =========================================================================

    /// NORD7 - Teal (#8FBCBB)
    ///
    /// Used for class names and type annotations.
    pub const NORD7: Color = Color::Rgb(143, 188, 187);

    /// NORD8 - Light blue (#88C0D0)
    ///
    /// Primary accent color, used for highlights and focus states.
    pub const NORD8: Color = Color::Rgb(136, 192, 208);

    /// NORD9 - Blue (#81A1C1)
    ///
    /// Used for keywords, usernames, and secondary accents.
    pub const NORD9: Color = Color::Rgb(129, 161, 193);

    /// NORD10 - Darker blue (#5E81AC)
    ///
    /// Used for functions and methods.
    pub const NORD10: Color = Color::Rgb(94, 129, 172);

    // =========================================================================
    // Aurora - Status/Semantic colors
    // =========================================================================

    /// NORD11 - Red (#BF616A)
    ///
    /// Error states, destructive actions, and critical warnings.
    pub const NORD11: Color = Color::Rgb(191, 97, 106);

    /// NORD12 - Orange (#D08770)
    ///
    /// Warning states and caution indicators.
    pub const NORD12: Color = Color::Rgb(208, 135, 112);

    /// NORD13 - Yellow (#EBCB8B)
    ///
    /// Attention-grabbing elements, pinned items.
    pub const NORD13: Color = Color::Rgb(235, 203, 139);

    /// NORD14 - Green (#A3BE8C)
    ///
    /// Success states, online indicators, outgoing messages.
    pub const NORD14: Color = Color::Rgb(163, 190, 140);

    /// NORD15 - Purple (#B48EAD)
    ///
    /// Decorative accents, numbers, and constants.
    pub const NORD15: Color = Color::Rgb(180, 142, 173);
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
    ///
    /// Uses NORD4 for good readability on dark backgrounds.
    #[must_use]
    pub const fn text() -> Style {
        Style::new().fg(colors::NORD4)
    }

    /// Muted text style for less important content.
    ///
    /// Uses NORD3 for subtle, de-emphasized text.
    #[must_use]
    pub const fn text_muted() -> Style {
        Style::new().fg(colors::NORD3)
    }

    /// Bright text style for emphasis.
    ///
    /// Uses NORD6 for headlines and important text.
    #[must_use]
    pub const fn text_bright() -> Style {
        Style::new().fg(colors::NORD6)
    }

    /// Accent text style.
    ///
    /// Uses NORD8 (light blue) for highlighted text.
    #[must_use]
    pub const fn text_accent() -> Style {
        Style::new().fg(colors::NORD8)
    }

    // =========================================================================
    // Status Styles
    // =========================================================================

    /// Success/positive style.
    ///
    /// Uses NORD14 (green) for success messages and confirmations.
    #[must_use]
    pub const fn success() -> Style {
        Style::new().fg(colors::NORD14)
    }

    /// Warning style.
    ///
    /// Uses NORD12 (orange) for warnings and cautions.
    #[must_use]
    pub const fn warning() -> Style {
        Style::new().fg(colors::NORD12)
    }

    /// Error style.
    ///
    /// Uses NORD11 (red) for errors and critical issues.
    #[must_use]
    pub const fn error() -> Style {
        Style::new().fg(colors::NORD11)
    }

    /// Informational style.
    ///
    /// Uses NORD9 (blue) for informational messages.
    #[must_use]
    pub const fn info() -> Style {
        Style::new().fg(colors::NORD9)
    }

    // =========================================================================
    // Selection/Highlight Styles
    // =========================================================================

    /// Selected item style.
    ///
    /// Uses NORD2 background with NORD6 foreground for clear selection indication.
    #[must_use]
    pub const fn selected() -> Style {
        Style::new().bg(colors::NORD2).fg(colors::NORD6)
    }

    /// Highlight style with bold modifier.
    ///
    /// Uses NORD8 with bold for emphasis without changing background.
    #[must_use]
    pub const fn highlight() -> Style {
        Style::new().fg(colors::NORD8).add_modifier(Modifier::BOLD)
    }

    // =========================================================================
    // Border Styles
    // =========================================================================

    /// Default border style.
    ///
    /// Uses NORD3 for subtle, non-intrusive borders.
    #[must_use]
    pub const fn border() -> Style {
        Style::new().fg(colors::NORD3)
    }

    /// Focused border style.
    ///
    /// Uses NORD8 (light blue) to clearly indicate focus.
    #[must_use]
    pub const fn border_focused() -> Style {
        Style::new().fg(colors::NORD8)
    }

    // =========================================================================
    // Chat List Styles
    // =========================================================================

    /// Unread chat style.
    ///
    /// Uses NORD8 with bold to highlight chats with unread messages.
    #[must_use]
    pub const fn chat_unread() -> Style {
        Style::new().fg(colors::NORD8).add_modifier(Modifier::BOLD)
    }

    /// Pinned chat style.
    ///
    /// Uses NORD13 (yellow) to indicate pinned chats.
    #[must_use]
    pub const fn chat_pinned() -> Style {
        Style::new().fg(colors::NORD13)
    }

    /// Muted chat style.
    ///
    /// Uses NORD3 to de-emphasize muted chats.
    #[must_use]
    pub const fn chat_muted() -> Style {
        Style::new().fg(colors::NORD3)
    }

    // =========================================================================
    // Message Styles
    // =========================================================================

    /// Outgoing message style.
    ///
    /// Uses NORD14 (green) to distinguish sent messages.
    #[must_use]
    pub const fn message_outgoing() -> Style {
        Style::new().fg(colors::NORD14)
    }

    /// Incoming message style.
    ///
    /// Uses NORD4 for standard received messages.
    #[must_use]
    pub const fn message_incoming() -> Style {
        Style::new().fg(colors::NORD4)
    }

    /// System message style.
    ///
    /// Uses NORD3 with italic for system notifications.
    #[must_use]
    pub const fn message_system() -> Style {
        Style::new()
            .fg(colors::NORD3)
            .add_modifier(Modifier::ITALIC)
    }

    /// Timestamp style.
    ///
    /// Uses NORD3 for subtle timestamp display.
    #[must_use]
    pub const fn timestamp() -> Style {
        Style::new().fg(colors::NORD3)
    }

    /// Username style.
    ///
    /// Uses NORD9 with bold to highlight sender names.
    #[must_use]
    pub const fn username() -> Style {
        Style::new().fg(colors::NORD9).add_modifier(Modifier::BOLD)
    }

    // =========================================================================
    // Status Bar Styles
    // =========================================================================

    /// Status bar background style.
    ///
    /// Uses NORD1 background with NORD4 text.
    #[must_use]
    pub const fn status_bar() -> Style {
        Style::new().bg(colors::NORD1).fg(colors::NORD4)
    }

    /// Online status indicator style.
    ///
    /// Uses NORD14 (green) for online status.
    #[must_use]
    pub const fn status_online() -> Style {
        Style::new().fg(colors::NORD14)
    }

    /// Offline status indicator style.
    ///
    /// Uses NORD11 (red) for offline status.
    #[must_use]
    pub const fn status_offline() -> Style {
        Style::new().fg(colors::NORD11)
    }

    // =========================================================================
    // Input Styles
    // =========================================================================

    /// Input field text style.
    ///
    /// Uses NORD4 for standard input text.
    #[must_use]
    pub const fn input() -> Style {
        Style::new().fg(colors::NORD4)
    }

    /// Input cursor style.
    ///
    /// Uses NORD0 on NORD8 background for a visible cursor.
    #[must_use]
    pub const fn input_cursor() -> Style {
        Style::new().fg(colors::NORD0).bg(colors::NORD8)
    }

    /// Input placeholder style.
    ///
    /// Uses NORD3 for subtle placeholder text.
    #[must_use]
    pub const fn input_placeholder() -> Style {
        Style::new().fg(colors::NORD3)
    }

    // =========================================================================
    // Modal/Dialog Styles
    // =========================================================================

    /// Modal background style.
    ///
    /// Uses NORD1 background for modal overlays.
    #[must_use]
    pub const fn modal_background() -> Style {
        Style::new().bg(colors::NORD1)
    }

    /// Modal title style.
    ///
    /// Uses NORD6 with bold for modal titles.
    #[must_use]
    pub const fn modal_title() -> Style {
        Style::new().fg(colors::NORD6).add_modifier(Modifier::BOLD)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use ratatui::style::Color;

    #[test]
    fn test_nord_colors_are_correct() {
        // Verify a few key colors match the Nord specification
        assert_eq!(colors::NORD0, Color::Rgb(46, 52, 64));
        assert_eq!(colors::NORD8, Color::Rgb(136, 192, 208));
        assert_eq!(colors::NORD11, Color::Rgb(191, 97, 106));
        assert_eq!(colors::NORD14, Color::Rgb(163, 190, 140));
    }

    #[test]
    fn test_text_style() {
        let style = Styles::text();
        assert_eq!(style.fg, Some(colors::NORD4));
    }

    #[test]
    fn test_selected_style() {
        let style = Styles::selected();
        assert_eq!(style.fg, Some(colors::NORD6));
        assert_eq!(style.bg, Some(colors::NORD2));
    }

    #[test]
    fn test_highlight_style_has_bold() {
        let style = Styles::highlight();
        assert!(style.add_modifier.contains(Modifier::BOLD));
    }

    #[test]
    fn test_status_styles() {
        assert_eq!(Styles::success().fg, Some(colors::NORD14));
        assert_eq!(Styles::warning().fg, Some(colors::NORD12));
        assert_eq!(Styles::error().fg, Some(colors::NORD11));
        assert_eq!(Styles::info().fg, Some(colors::NORD9));
    }

    #[test]
    fn test_border_styles() {
        assert_eq!(Styles::border().fg, Some(colors::NORD3));
        assert_eq!(Styles::border_focused().fg, Some(colors::NORD8));
    }
}
