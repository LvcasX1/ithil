//! Theme-aware styles for Ithil.
//!
//! This module provides runtime theme switching with predefined color palettes
//! (Nord, Dracula, Tokyo Night, etc.) and a System theme that inherits the
//! user's terminal colors.
//!
//! # Usage
//!
//! ```rust,no_run
//! use ithil::ui::styles::{colors, Styles, Theme};
//! use ratatui::widgets::Paragraph;
//!
//! // Set the active theme
//! Theme::Nord.apply();
//!
//! let text = Paragraph::new("Hello, world!")
//!     .style(Styles::text());
//!
//! let error = Paragraph::new("Error occurred!")
//!     .style(Styles::error());
//! ```

use ratatui::style::{Color, Modifier, Style};
use std::sync::atomic::{AtomicU8, Ordering};

// =========================================================================
// Theme enum and palette
// =========================================================================

/// Global theme index.
static CURRENT_THEME: AtomicU8 = AtomicU8::new(0);

/// A color palette with all semantic color slots.
#[derive(Debug, Clone, Copy)]
pub struct ThemePalette {
    pub bg_primary: Color,
    pub bg_elevated: Color,
    pub bg_interactive: Color,
    pub fg_primary: Color,
    pub fg_muted: Color,
    pub fg_bright: Color,
    pub accent_primary: Color,
    pub accent_secondary: Color,
    pub status_error: Color,
    pub status_warning: Color,
    pub status_attention: Color,
    pub status_success: Color,
    pub decorative: Color,
}

/// Available color themes.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
#[repr(u8)]
pub enum Theme {
    /// Inherit colors from the terminal emulator.
    #[default]
    System = 0,
    /// Nord - cold, bluish palette.
    Nord = 1,
    /// Dracula - dark purple palette.
    Dracula = 2,
    /// Tokyo Night - muted blue/purple palette.
    TokyoNight = 3,
    /// Gruvbox - retro warm palette.
    Gruvbox = 4,
    /// Catppuccin Mocha - pastel dark palette.
    CatppuccinMocha = 5,
    /// Solarized Dark - precision dark palette.
    SolarizedDark = 6,
    /// One Dark - Atom editor palette.
    OneDark = 7,
}

impl Theme {
    /// All available themes in display order.
    pub const ALL: [Self; 8] = [
        Self::System,
        Self::Nord,
        Self::Dracula,
        Self::TokyoNight,
        Self::Gruvbox,
        Self::CatppuccinMocha,
        Self::SolarizedDark,
        Self::OneDark,
    ];

    /// Display name for UI.
    #[must_use]
    pub const fn name(&self) -> &'static str {
        match self {
            Self::System => "System",
            Self::Nord => "Nord",
            Self::Dracula => "Dracula",
            Self::TokyoNight => "Tokyo Night",
            Self::Gruvbox => "Gruvbox",
            Self::CatppuccinMocha => "Catppuccin Mocha",
            Self::SolarizedDark => "Solarized Dark",
            Self::OneDark => "One Dark",
        }
    }

    /// Parse from config string.
    #[must_use]
    pub fn from_config_str(s: &str) -> Self {
        match s.to_lowercase().as_str() {
            "nord" => Self::Nord,
            "dracula" => Self::Dracula,
            "tokyo-night" | "tokyo_night" | "tokyonight" => Self::TokyoNight,
            "gruvbox" => Self::Gruvbox,
            "catppuccin-mocha" | "catppuccin_mocha" | "catppuccinmocha" | "catppuccin" => {
                Self::CatppuccinMocha
            },
            "solarized-dark" | "solarized_dark" | "solarizeddark" | "solarized" => {
                Self::SolarizedDark
            },
            "one-dark" | "one_dark" | "onedark" => Self::OneDark,
            // "system", "default", "dark", "light", and anything unknown
            _ => Self::System,
        }
    }

    /// Serialize to config string.
    #[must_use]
    pub const fn to_config_str(&self) -> &'static str {
        match self {
            Self::System => "system",
            Self::Nord => "nord",
            Self::Dracula => "dracula",
            Self::TokyoNight => "tokyo-night",
            Self::Gruvbox => "gruvbox",
            Self::CatppuccinMocha => "catppuccin-mocha",
            Self::SolarizedDark => "solarized-dark",
            Self::OneDark => "one-dark",
        }
    }

    /// Set this theme as the active global theme.
    pub fn apply(self) {
        CURRENT_THEME.store(self as u8, Ordering::Relaxed);
    }

    /// Get the currently active theme.
    #[must_use]
    pub fn current() -> Self {
        Self::from_u8(CURRENT_THEME.load(Ordering::Relaxed))
    }

    const fn from_u8(v: u8) -> Self {
        match v {
            1 => Self::Nord,
            2 => Self::Dracula,
            3 => Self::TokyoNight,
            4 => Self::Gruvbox,
            5 => Self::CatppuccinMocha,
            6 => Self::SolarizedDark,
            7 => Self::OneDark,
            // 0 and anything unknown
            _ => Self::System,
        }
    }

    /// Returns the color palette for this theme.
    #[must_use]
    pub const fn palette(&self) -> ThemePalette {
        match self {
            Self::System => ThemePalette {
                bg_primary: Color::Reset,
                bg_elevated: Color::Reset,
                bg_interactive: Color::DarkGray,
                fg_primary: Color::Reset,
                fg_muted: Color::DarkGray,
                fg_bright: Color::White,
                accent_primary: Color::Cyan,
                accent_secondary: Color::Blue,
                status_error: Color::Red,
                status_warning: Color::Yellow,
                status_attention: Color::Yellow,
                status_success: Color::Green,
                decorative: Color::Magenta,
            },
            Self::Nord => ThemePalette {
                bg_primary: Color::Rgb(46, 52, 64),
                bg_elevated: Color::Rgb(59, 66, 82),
                bg_interactive: Color::Rgb(67, 76, 94),
                fg_primary: Color::Rgb(216, 222, 233),
                fg_muted: Color::Rgb(76, 86, 106),
                fg_bright: Color::Rgb(236, 239, 244),
                accent_primary: Color::Rgb(136, 192, 208),
                accent_secondary: Color::Rgb(129, 161, 193),
                status_error: Color::Rgb(191, 97, 106),
                status_warning: Color::Rgb(208, 135, 112),
                status_attention: Color::Rgb(235, 203, 139),
                status_success: Color::Rgb(163, 190, 140),
                decorative: Color::Rgb(180, 142, 173),
            },
            Self::Dracula => ThemePalette {
                bg_primary: Color::Rgb(40, 42, 54),
                bg_elevated: Color::Rgb(68, 71, 90),
                bg_interactive: Color::Rgb(68, 71, 90),
                fg_primary: Color::Rgb(248, 248, 242),
                fg_muted: Color::Rgb(98, 114, 164),
                fg_bright: Color::Rgb(255, 255, 255),
                accent_primary: Color::Rgb(139, 233, 253),
                accent_secondary: Color::Rgb(189, 147, 249),
                status_error: Color::Rgb(255, 85, 85),
                status_warning: Color::Rgb(255, 184, 108),
                status_attention: Color::Rgb(241, 250, 140),
                status_success: Color::Rgb(80, 250, 123),
                decorative: Color::Rgb(255, 121, 198),
            },
            Self::TokyoNight => ThemePalette {
                bg_primary: Color::Rgb(26, 27, 38),
                bg_elevated: Color::Rgb(36, 40, 59),
                bg_interactive: Color::Rgb(41, 46, 66),
                fg_primary: Color::Rgb(192, 202, 245),
                fg_muted: Color::Rgb(86, 95, 137),
                fg_bright: Color::Rgb(220, 228, 255),
                accent_primary: Color::Rgb(125, 207, 255),
                accent_secondary: Color::Rgb(122, 162, 247),
                status_error: Color::Rgb(247, 118, 142),
                status_warning: Color::Rgb(224, 175, 104),
                status_attention: Color::Rgb(224, 175, 104),
                status_success: Color::Rgb(158, 206, 106),
                decorative: Color::Rgb(187, 154, 247),
            },
            Self::Gruvbox => ThemePalette {
                bg_primary: Color::Rgb(40, 40, 40),
                bg_elevated: Color::Rgb(60, 56, 54),
                bg_interactive: Color::Rgb(80, 73, 69),
                fg_primary: Color::Rgb(235, 219, 178),
                fg_muted: Color::Rgb(146, 131, 116),
                fg_bright: Color::Rgb(253, 244, 193),
                accent_primary: Color::Rgb(131, 165, 152),
                accent_secondary: Color::Rgb(69, 133, 136),
                status_error: Color::Rgb(204, 36, 29),
                status_warning: Color::Rgb(254, 128, 25),
                status_attention: Color::Rgb(250, 189, 47),
                status_success: Color::Rgb(152, 151, 26),
                decorative: Color::Rgb(177, 98, 134),
            },
            Self::CatppuccinMocha => ThemePalette {
                bg_primary: Color::Rgb(30, 30, 46),
                bg_elevated: Color::Rgb(49, 50, 68),
                bg_interactive: Color::Rgb(69, 71, 90),
                fg_primary: Color::Rgb(205, 214, 244),
                fg_muted: Color::Rgb(108, 112, 134),
                fg_bright: Color::Rgb(255, 255, 255),
                accent_primary: Color::Rgb(137, 220, 235),
                accent_secondary: Color::Rgb(116, 199, 236),
                status_error: Color::Rgb(243, 139, 168),
                status_warning: Color::Rgb(250, 179, 135),
                status_attention: Color::Rgb(249, 226, 175),
                status_success: Color::Rgb(166, 227, 161),
                decorative: Color::Rgb(203, 166, 247),
            },
            Self::SolarizedDark => ThemePalette {
                bg_primary: Color::Rgb(0, 43, 54),
                bg_elevated: Color::Rgb(7, 54, 66),
                bg_interactive: Color::Rgb(7, 54, 66),
                fg_primary: Color::Rgb(131, 148, 150),
                fg_muted: Color::Rgb(88, 110, 117),
                fg_bright: Color::Rgb(238, 232, 213),
                accent_primary: Color::Rgb(42, 161, 152),
                accent_secondary: Color::Rgb(38, 139, 210),
                status_error: Color::Rgb(220, 50, 47),
                status_warning: Color::Rgb(203, 75, 22),
                status_attention: Color::Rgb(181, 137, 0),
                status_success: Color::Rgb(133, 153, 0),
                decorative: Color::Rgb(108, 113, 196),
            },
            Self::OneDark => ThemePalette {
                bg_primary: Color::Rgb(40, 44, 52),
                bg_elevated: Color::Rgb(50, 55, 65),
                bg_interactive: Color::Rgb(62, 68, 81),
                fg_primary: Color::Rgb(171, 178, 191),
                fg_muted: Color::Rgb(92, 99, 112),
                fg_bright: Color::Rgb(220, 223, 228),
                accent_primary: Color::Rgb(86, 182, 194),
                accent_secondary: Color::Rgb(97, 175, 239),
                status_error: Color::Rgb(224, 108, 117),
                status_warning: Color::Rgb(209, 154, 102),
                status_attention: Color::Rgb(229, 192, 123),
                status_success: Color::Rgb(152, 195, 121),
                decorative: Color::Rgb(198, 120, 221),
            },
        }
    }
}

/// Returns the palette for the currently active theme.
#[must_use]
fn current_palette() -> ThemePalette {
    Theme::current().palette()
}

/// Semantic color accessors that read from the active theme.
pub mod colors {
    use super::current_palette;
    use ratatui::style::Color;

    // Backgrounds
    #[must_use]
    pub fn bg_primary() -> Color {
        current_palette().bg_primary
    }
    #[must_use]
    pub fn bg_elevated() -> Color {
        current_palette().bg_elevated
    }
    #[must_use]
    pub fn bg_interactive() -> Color {
        current_palette().bg_interactive
    }

    // Foregrounds
    #[must_use]
    pub fn fg_primary() -> Color {
        current_palette().fg_primary
    }
    #[must_use]
    pub fn fg_muted() -> Color {
        current_palette().fg_muted
    }
    #[must_use]
    pub fn fg_bright() -> Color {
        current_palette().fg_bright
    }

    // Accents
    #[must_use]
    pub fn accent_primary() -> Color {
        current_palette().accent_primary
    }
    #[must_use]
    pub fn accent_secondary() -> Color {
        current_palette().accent_secondary
    }

    // Status
    #[must_use]
    pub fn status_error() -> Color {
        current_palette().status_error
    }
    #[must_use]
    pub fn status_warning() -> Color {
        current_palette().status_warning
    }
    #[must_use]
    pub fn status_attention() -> Color {
        current_palette().status_attention
    }
    #[must_use]
    pub fn status_success() -> Color {
        current_palette().status_success
    }
    #[must_use]
    pub fn decorative() -> Color {
        current_palette().decorative
    }
}

/// Pre-built styles for common UI elements.
///
/// These methods read from the active theme, so changing the theme
/// immediately updates all styles across the application.
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
    pub fn text() -> Style {
        Style::new().fg(colors::fg_primary())
    }

    /// Muted text style for less important content.
    #[must_use]
    pub fn text_muted() -> Style {
        Style::new().fg(colors::fg_muted())
    }

    /// Bright text style for emphasis.
    #[must_use]
    pub fn text_bright() -> Style {
        Style::new().fg(colors::fg_bright())
    }

    /// Accent text style.
    #[must_use]
    pub fn text_accent() -> Style {
        Style::new().fg(colors::accent_primary())
    }

    // =========================================================================
    // Status Styles
    // =========================================================================

    /// Success/positive style.
    #[must_use]
    pub fn success() -> Style {
        Style::new().fg(colors::status_success())
    }

    /// Warning style.
    #[must_use]
    pub fn warning() -> Style {
        Style::new().fg(colors::status_warning())
    }

    /// Error style.
    #[must_use]
    pub fn error() -> Style {
        Style::new().fg(colors::status_error())
    }

    /// Informational style.
    #[must_use]
    pub fn info() -> Style {
        Style::new().fg(colors::accent_secondary())
    }

    // =========================================================================
    // Selection/Highlight Styles
    // =========================================================================

    /// Selected item style.
    #[must_use]
    pub fn selected() -> Style {
        Style::new()
            .bg(colors::bg_interactive())
            .fg(colors::fg_bright())
    }

    /// Highlight style with bold modifier.
    #[must_use]
    pub fn highlight() -> Style {
        Style::new()
            .fg(colors::accent_primary())
            .add_modifier(Modifier::BOLD)
    }

    // =========================================================================
    // Border Styles
    // =========================================================================

    /// Default border style.
    #[must_use]
    pub fn border() -> Style {
        Style::new().fg(colors::fg_muted())
    }

    /// Focused border style.
    #[must_use]
    pub fn border_focused() -> Style {
        Style::new().fg(colors::accent_primary())
    }

    // =========================================================================
    // Chat List Styles
    // =========================================================================

    /// Unread chat style.
    #[must_use]
    pub fn chat_unread() -> Style {
        Style::new()
            .fg(colors::accent_primary())
            .add_modifier(Modifier::BOLD)
    }

    /// Pinned chat style.
    #[must_use]
    pub fn chat_pinned() -> Style {
        Style::new().fg(colors::status_attention())
    }

    /// Muted chat style.
    #[must_use]
    pub fn chat_muted() -> Style {
        Style::new().fg(colors::fg_muted())
    }

    // =========================================================================
    // Message Styles
    // =========================================================================

    /// Outgoing message style.
    #[must_use]
    pub fn message_outgoing() -> Style {
        Style::new().fg(colors::status_success())
    }

    /// Incoming message style.
    #[must_use]
    pub fn message_incoming() -> Style {
        Style::new().fg(colors::fg_primary())
    }

    /// System message style.
    #[must_use]
    pub fn message_system() -> Style {
        Style::new()
            .fg(colors::fg_muted())
            .add_modifier(Modifier::ITALIC)
    }

    /// Timestamp style.
    #[must_use]
    pub fn timestamp() -> Style {
        Style::new().fg(colors::fg_muted())
    }

    /// Username style.
    #[must_use]
    pub fn username() -> Style {
        Style::new()
            .fg(colors::accent_secondary())
            .add_modifier(Modifier::BOLD)
    }

    // =========================================================================
    // Status Bar Styles
    // =========================================================================

    /// Status bar background style.
    #[must_use]
    pub fn status_bar() -> Style {
        Style::new()
            .bg(colors::bg_interactive())
            .fg(colors::fg_primary())
    }

    /// Online status indicator style.
    #[must_use]
    pub fn status_online() -> Style {
        Style::new().fg(colors::status_success())
    }

    /// Offline status indicator style.
    #[must_use]
    pub fn status_offline() -> Style {
        Style::new().fg(colors::status_error())
    }

    // =========================================================================
    // Input Styles
    // =========================================================================

    /// Input field text style.
    #[must_use]
    pub fn input() -> Style {
        Style::new().fg(colors::fg_primary())
    }

    /// Input cursor style.
    #[must_use]
    pub fn input_cursor() -> Style {
        Style::new()
            .fg(colors::bg_primary())
            .bg(colors::accent_primary())
    }

    /// Input placeholder style.
    #[must_use]
    pub fn input_placeholder() -> Style {
        Style::new().fg(colors::fg_muted())
    }

    // =========================================================================
    // Modal/Dialog Styles
    // =========================================================================

    /// Modal background style.
    #[must_use]
    pub fn modal_background() -> Style {
        Style::new().bg(colors::bg_interactive())
    }

    /// Modal title style.
    #[must_use]
    pub fn modal_title() -> Style {
        Style::new()
            .fg(colors::fg_bright())
            .add_modifier(Modifier::BOLD)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_system_theme_uses_ansi() {
        Theme::System.apply();
        assert_eq!(colors::bg_primary(), Color::Reset);
        assert_eq!(colors::fg_primary(), Color::Reset);
        assert_eq!(colors::accent_primary(), Color::Cyan);
        assert_eq!(colors::status_error(), Color::Red);
        assert_eq!(colors::status_success(), Color::Green);
    }

    #[test]
    fn test_nord_theme_uses_rgb() {
        Theme::Nord.apply();
        assert_eq!(colors::bg_primary(), Color::Rgb(46, 52, 64));
        assert_eq!(colors::accent_primary(), Color::Rgb(136, 192, 208));
        // Restore
        Theme::System.apply();
    }

    #[test]
    fn test_theme_switching() {
        Theme::System.apply();
        assert_eq!(Theme::current(), Theme::System);

        Theme::Dracula.apply();
        assert_eq!(Theme::current(), Theme::Dracula);
        assert_eq!(colors::bg_primary(), Color::Rgb(40, 42, 54));

        // Restore
        Theme::System.apply();
    }

    #[test]
    fn test_theme_from_config_str() {
        assert_eq!(Theme::from_config_str("system"), Theme::System);
        assert_eq!(Theme::from_config_str("dark"), Theme::System);
        assert_eq!(Theme::from_config_str("nord"), Theme::Nord);
        assert_eq!(Theme::from_config_str("dracula"), Theme::Dracula);
        assert_eq!(Theme::from_config_str("tokyo-night"), Theme::TokyoNight);
        assert_eq!(Theme::from_config_str("tokyonight"), Theme::TokyoNight);
        assert_eq!(Theme::from_config_str("gruvbox"), Theme::Gruvbox);
        assert_eq!(Theme::from_config_str("catppuccin-mocha"), Theme::CatppuccinMocha);
        assert_eq!(Theme::from_config_str("catppuccin"), Theme::CatppuccinMocha);
        assert_eq!(Theme::from_config_str("solarized-dark"), Theme::SolarizedDark);
        assert_eq!(Theme::from_config_str("solarized"), Theme::SolarizedDark);
        assert_eq!(Theme::from_config_str("one-dark"), Theme::OneDark);
        assert_eq!(Theme::from_config_str("unknown"), Theme::System);
    }

    #[test]
    fn test_theme_to_config_str() {
        assert_eq!(Theme::System.to_config_str(), "system");
        assert_eq!(Theme::Nord.to_config_str(), "nord");
        assert_eq!(Theme::TokyoNight.to_config_str(), "tokyo-night");
    }

    #[test]
    fn test_theme_all() {
        assert_eq!(Theme::ALL.len(), 8);
        assert_eq!(Theme::ALL[0], Theme::System);
        assert_eq!(Theme::ALL[7], Theme::OneDark);
    }

    #[test]
    fn test_theme_name() {
        assert_eq!(Theme::System.name(), "System");
        assert_eq!(Theme::CatppuccinMocha.name(), "Catppuccin Mocha");
        assert_eq!(Theme::SolarizedDark.name(), "Solarized Dark");
    }

    #[test]
    fn test_system_palette_colors() {
        let p = Theme::System.palette();
        assert_eq!(p.fg_primary, Color::Reset);
        assert_eq!(p.fg_bright, Color::White);
        assert_eq!(p.bg_interactive, Color::DarkGray);
        assert_eq!(p.accent_primary, Color::Cyan);
        assert_eq!(p.accent_secondary, Color::Blue);
        assert_eq!(p.status_success, Color::Green);
        assert_eq!(p.status_warning, Color::Yellow);
        assert_eq!(p.status_error, Color::Red);
        assert_eq!(p.fg_muted, Color::DarkGray);
    }

    #[test]
    fn test_highlight_style_has_bold() {
        let style = Styles::highlight();
        assert!(style.add_modifier.contains(Modifier::BOLD));
    }

    #[test]
    fn test_all_themes_have_valid_palettes() {
        for theme in Theme::ALL {
            let p = theme.palette();
            // Just verify we can access all fields without panic
            let _ = (
                p.bg_primary,
                p.bg_elevated,
                p.bg_interactive,
                p.fg_primary,
                p.fg_muted,
                p.fg_bright,
                p.accent_primary,
                p.accent_secondary,
                p.status_error,
                p.status_warning,
                p.status_attention,
                p.status_success,
                p.decorative,
            );
        }
    }
}
