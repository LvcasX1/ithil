//! Help modal component showing keyboard shortcuts.
//!
//! Displays a list of available keyboard shortcuts based on the current
//! key binding mode (standard or Vim).
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::components::{HelpModal, HelpModalWidget};
//!
//! let help = HelpModal::new(true); // vim mode
//! // Render with HelpModalWidget::new(&help)
//! ```

use ratatui::{
    buffer::Buffer,
    layout::Rect,
    text::{Line, Span},
    widgets::{Block, Borders, Clear, Paragraph, Widget},
};

use crate::ui::keys::KeyMap;
use crate::ui::styles::Styles;

/// Help modal displaying keyboard shortcuts.
///
/// The help modal shows a list of available keyboard shortcuts. The shortcuts
/// displayed depend on whether Vim mode is enabled.
#[derive(Debug, Clone)]
pub struct HelpModal {
    /// Key map containing all available shortcuts
    keymap: KeyMap,
}

impl HelpModal {
    /// Creates a new help modal for the given key mode.
    ///
    /// # Arguments
    ///
    /// * `vim_mode` - If `true`, shows Vim-style shortcuts; otherwise standard shortcuts
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::HelpModal;
    ///
    /// let help = HelpModal::new(false); // Standard mode
    /// let vim_help = HelpModal::new(true); // Vim mode
    /// ```
    #[must_use]
    pub fn new(vim_mode: bool) -> Self {
        Self {
            keymap: KeyMap::new(vim_mode),
        }
    }

    /// Returns whether vim mode is enabled.
    #[must_use]
    pub const fn is_vim_mode(&self) -> bool {
        self.keymap.is_vim_mode()
    }

    /// Returns the help text entries.
    ///
    /// Each entry is a tuple of (key description, action description).
    #[must_use]
    pub fn get_help_entries(&self) -> Vec<(&'static str, &'static str)> {
        self.keymap.get_help_text()
    }
}

impl Default for HelpModal {
    fn default() -> Self {
        Self::new(false)
    }
}

/// Widget for rendering the help modal.
pub struct HelpModalWidget<'a> {
    modal: &'a HelpModal,
}

impl<'a> HelpModalWidget<'a> {
    /// Creates a new help modal widget.
    #[must_use]
    pub const fn new(modal: &'a HelpModal) -> Self {
        Self { modal }
    }
}

impl Widget for HelpModalWidget<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        let help_items = self.modal.keymap.get_help_text();
        // Calculate height: items + borders (2) + title line + footer (2 lines)
        #[allow(clippy::cast_possible_truncation)]
        let height = (help_items.len() as u16).saturating_add(5);
        let width = 45u16;

        // Center the modal
        let x = area.x + area.width.saturating_sub(width) / 2;
        let y = area.y + area.height.saturating_sub(height) / 2;
        let modal_area = Rect::new(x, y, width.min(area.width), height.min(area.height));

        // Clear background
        Clear.render(modal_area, buf);

        // Apply background style
        buf.set_style(modal_area, Styles::modal_background());

        // Render border with title
        let title = if self.modal.is_vim_mode() {
            " Keyboard Shortcuts (Vim) "
        } else {
            " Keyboard Shortcuts "
        };

        let block = Block::default()
            .title(title)
            .title_style(Styles::modal_title())
            .borders(Borders::ALL)
            .border_style(Styles::border_focused());

        let inner = block.inner(modal_area);
        block.render(modal_area, buf);

        // Build help text lines
        let mut lines: Vec<Line> = help_items
            .iter()
            .map(|(key, desc)| {
                Line::from(vec![
                    Span::styled(format!("{key:14}"), Styles::text_accent()),
                    Span::styled(*desc, Styles::text()),
                ])
            })
            .collect();

        // Add footer
        lines.push(Line::from(""));
        lines.push(Line::from(vec![Span::styled(
            "Press ? to close",
            Styles::text_muted(),
        )]));

        let paragraph = Paragraph::new(lines);
        paragraph.render(inner, buf);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_help_modal_standard_mode() {
        let help = HelpModal::new(false);
        assert!(!help.is_vim_mode());
    }

    #[test]
    fn test_help_modal_vim_mode() {
        let help = HelpModal::new(true);
        assert!(help.is_vim_mode());
    }

    #[test]
    fn test_help_modal_default() {
        let help = HelpModal::default();
        assert!(!help.is_vim_mode());
    }

    #[test]
    fn test_get_help_entries_standard() {
        let help = HelpModal::new(false);
        let entries = help.get_help_entries();

        // Should have multiple entries
        assert!(!entries.is_empty());

        // Should contain arrow key navigation in standard mode
        let has_arrows = entries.iter().any(|(k, _)| k.contains("↑/↓"));
        assert!(has_arrows, "Standard mode should show arrow key navigation");
    }

    #[test]
    fn test_get_help_entries_vim() {
        let help = HelpModal::new(true);
        let entries = help.get_help_entries();

        // Should have multiple entries
        assert!(!entries.is_empty());

        // Should contain j/k navigation in vim mode
        let has_vim_nav = entries.iter().any(|(k, _)| k.contains("j/k"));
        assert!(has_vim_nav, "Vim mode should show j/k navigation");
    }

    #[test]
    fn test_help_entries_have_descriptions() {
        let help = HelpModal::new(false);
        let entries = help.get_help_entries();

        for (key, desc) in entries {
            assert!(!key.is_empty(), "Key should not be empty");
            assert!(!desc.is_empty(), "Description should not be empty");
        }
    }

    #[test]
    fn test_help_modal_widget_creation() {
        let help = HelpModal::new(false);
        let _widget = HelpModalWidget::new(&help);
        // Widget creation should not panic
    }
}
