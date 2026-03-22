//! Generic modal dialog component.
//!
//! Provides a reusable modal dialog that can be configured with custom
//! title, content, and buttons for confirmations, alerts, and other dialogs.
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::components::{Modal, ModalWidget};
//!
//! // Create a confirmation dialog
//! let modal = Modal::confirm("Delete Message", "Are you sure you want to delete this message?");
//!
//! // Or an alert
//! let alert = Modal::alert("Error", "Failed to send message.");
//! ```

use ratatui::{
    buffer::Buffer,
    layout::{Alignment, Constraint, Direction, Layout, Rect},
    text::{Line, Span},
    widgets::{Block, Borders, Clear, Paragraph, Widget, Wrap},
};

use crate::ui::styles::Styles;

/// A generic modal dialog.
///
/// Modals are centered overlays that display a title, content, and one or more
/// buttons. They support keyboard navigation between buttons.
#[derive(Debug, Clone)]
pub struct Modal {
    /// Modal title displayed in the border
    pub title: String,
    /// Main content/message of the modal
    pub content: String,
    /// List of button labels
    pub buttons: Vec<String>,
    /// Index of the currently selected button
    pub selected_button: usize,
    /// Width of the modal in characters
    pub width: u16,
    /// Height of the modal in lines
    pub height: u16,
}

impl Modal {
    /// Creates a new modal with the given title and content.
    ///
    /// By default, creates a modal with a single "OK" button.
    ///
    /// # Arguments
    ///
    /// * `title` - Title displayed in the modal border
    /// * `content` - Main message content
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::Modal;
    ///
    /// let modal = Modal::new("Info", "Operation completed successfully.");
    /// assert_eq!(modal.buttons.len(), 1);
    /// assert_eq!(modal.buttons[0], "OK");
    /// ```
    #[must_use]
    pub fn new(title: impl Into<String>, content: impl Into<String>) -> Self {
        Self {
            title: title.into(),
            content: content.into(),
            buttons: vec!["OK".to_string()],
            selected_button: 0,
            width: 50,
            height: 10,
        }
    }

    /// Sets custom buttons for the modal.
    ///
    /// # Arguments
    ///
    /// * `buttons` - List of button labels
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::Modal;
    ///
    /// let modal = Modal::new("Save", "Do you want to save changes?")
    ///     .with_buttons(vec!["Save".to_string(), "Don't Save".to_string(), "Cancel".to_string()]);
    /// assert_eq!(modal.buttons.len(), 3);
    /// ```
    #[must_use]
    pub fn with_buttons(mut self, buttons: Vec<String>) -> Self {
        self.buttons = buttons;
        self
    }

    /// Sets the size of the modal.
    ///
    /// # Arguments
    ///
    /// * `width` - Width in characters
    /// * `height` - Height in lines
    #[must_use]
    pub const fn with_size(mut self, width: u16, height: u16) -> Self {
        self.width = width;
        self.height = height;
        self
    }

    /// Creates a confirmation dialog with "Yes" and "No" buttons.
    ///
    /// # Arguments
    ///
    /// * `title` - Title for the confirmation
    /// * `message` - Question or message to display
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::Modal;
    ///
    /// let modal = Modal::confirm("Delete", "Are you sure?");
    /// assert_eq!(modal.buttons, vec!["Yes", "No"]);
    /// ```
    #[must_use]
    pub fn confirm(title: impl Into<String>, message: impl Into<String>) -> Self {
        Self::new(title, message).with_buttons(vec!["Yes".to_string(), "No".to_string()])
    }

    /// Creates an alert dialog with a single "OK" button.
    ///
    /// # Arguments
    ///
    /// * `title` - Title for the alert
    /// * `message` - Message to display
    #[must_use]
    pub fn alert(title: impl Into<String>, message: impl Into<String>) -> Self {
        Self::new(title, message).with_buttons(vec!["OK".to_string()])
    }

    /// Selects the next button (moves right).
    ///
    /// Does nothing if already at the last button.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::Modal;
    ///
    /// let mut modal = Modal::confirm("Test", "Message");
    /// assert_eq!(modal.selected_button, 0);
    /// modal.select_next();
    /// assert_eq!(modal.selected_button, 1);
    /// modal.select_next(); // Already at last, no change
    /// assert_eq!(modal.selected_button, 1);
    /// ```
    pub fn select_next(&mut self) {
        if self.selected_button < self.buttons.len().saturating_sub(1) {
            self.selected_button += 1;
        }
    }

    /// Selects the previous button (moves left).
    ///
    /// Does nothing if already at the first button.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::Modal;
    ///
    /// let mut modal = Modal::confirm("Test", "Message");
    /// modal.selected_button = 1;
    /// modal.select_previous();
    /// assert_eq!(modal.selected_button, 0);
    /// modal.select_previous(); // Already at first, no change
    /// assert_eq!(modal.selected_button, 0);
    /// ```
    pub fn select_previous(&mut self) {
        if self.selected_button > 0 {
            self.selected_button -= 1;
        }
    }

    /// Returns the text of the currently selected button.
    ///
    /// Returns `None` if there are no buttons (shouldn't happen in normal use).
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::Modal;
    ///
    /// let modal = Modal::confirm("Test", "Message");
    /// assert_eq!(modal.selected_button_text(), Some("Yes"));
    /// ```
    #[must_use]
    pub fn selected_button_text(&self) -> Option<&str> {
        self.buttons.get(self.selected_button).map(String::as_str)
    }

    /// Returns the index of the currently selected button.
    #[must_use]
    pub const fn selected_button_index(&self) -> usize {
        self.selected_button
    }

    /// Returns true if the selected button is "Yes" (for confirm dialogs).
    #[must_use]
    pub fn is_confirmed(&self) -> bool {
        self.selected_button_text() == Some("Yes")
    }

    /// Returns true if the selected button is "No" or "Cancel".
    #[must_use]
    pub fn is_cancelled(&self) -> bool {
        matches!(self.selected_button_text(), Some("No" | "Cancel" | "Close"))
    }
}

impl Default for Modal {
    fn default() -> Self {
        Self::new("", "")
    }
}

/// Widget for rendering a modal dialog.
pub struct ModalWidget<'a> {
    modal: &'a Modal,
}

impl<'a> ModalWidget<'a> {
    /// Creates a new modal widget.
    #[must_use]
    pub const fn new(modal: &'a Modal) -> Self {
        Self { modal }
    }

    /// Calculates a centered rect within the given area.
    fn get_centered_rect(&self, area: Rect) -> Rect {
        let width = self.modal.width.min(area.width.saturating_sub(4));
        let height = self.modal.height.min(area.height.saturating_sub(4));
        let x = area.x + (area.width.saturating_sub(width)) / 2;
        let y = area.y + (area.height.saturating_sub(height)) / 2;
        Rect::new(x, y, width, height)
    }
}

impl Widget for ModalWidget<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        let modal_area = self.get_centered_rect(area);

        // Clear the background area
        Clear.render(modal_area, buf);

        // Apply modal background
        buf.set_style(modal_area, Styles::modal_background());

        // Render modal border with title
        let block = Block::default()
            .title(format!(" {} ", self.modal.title))
            .title_style(Styles::modal_title())
            .borders(Borders::ALL)
            .border_style(Styles::border_focused());

        let inner = block.inner(modal_area);
        block.render(modal_area, buf);

        // Split inner area into content and buttons
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Min(2),    // Content
                Constraint::Length(1), // Buttons
            ])
            .split(inner);

        // Render content with word wrap
        let content = Paragraph::new(self.modal.content.as_str())
            .style(Styles::text())
            .wrap(Wrap { trim: true });
        content.render(chunks[0], buf);

        // Render buttons
        let button_spans: Vec<Span> = self
            .modal
            .buttons
            .iter()
            .enumerate()
            .flat_map(|(idx, btn)| {
                let style = if idx == self.modal.selected_button {
                    Styles::selected()
                } else {
                    Styles::text_muted()
                };
                vec![Span::styled(format!(" [{btn}] "), style)]
            })
            .collect();

        let buttons = Paragraph::new(Line::from(button_spans)).alignment(Alignment::Center);
        buttons.render(chunks[1], buf);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_modal_new() {
        let modal = Modal::new("Title", "Content");
        assert_eq!(modal.title, "Title");
        assert_eq!(modal.content, "Content");
        assert_eq!(modal.buttons, vec!["OK"]);
        assert_eq!(modal.selected_button, 0);
    }

    #[test]
    fn test_modal_with_buttons() {
        let modal =
            Modal::new("Title", "Content").with_buttons(vec!["A".to_string(), "B".to_string()]);
        assert_eq!(modal.buttons, vec!["A", "B"]);
    }

    #[test]
    fn test_modal_with_size() {
        let modal = Modal::new("Title", "Content").with_size(60, 15);
        assert_eq!(modal.width, 60);
        assert_eq!(modal.height, 15);
    }

    #[test]
    fn test_modal_confirm() {
        let modal = Modal::confirm("Delete", "Are you sure?");
        assert_eq!(modal.title, "Delete");
        assert_eq!(modal.content, "Are you sure?");
        assert_eq!(modal.buttons, vec!["Yes", "No"]);
    }

    #[test]
    fn test_modal_alert() {
        let modal = Modal::alert("Error", "Something went wrong");
        assert_eq!(modal.buttons, vec!["OK"]);
    }

    #[test]
    fn test_select_next() {
        let mut modal = Modal::confirm("Test", "Message");
        assert_eq!(modal.selected_button, 0);

        modal.select_next();
        assert_eq!(modal.selected_button, 1);

        // Should not go past last button
        modal.select_next();
        assert_eq!(modal.selected_button, 1);
    }

    #[test]
    fn test_select_previous() {
        let mut modal = Modal::confirm("Test", "Message");
        modal.selected_button = 1;

        modal.select_previous();
        assert_eq!(modal.selected_button, 0);

        // Should not go below 0
        modal.select_previous();
        assert_eq!(modal.selected_button, 0);
    }

    #[test]
    fn test_selected_button_text() {
        let modal = Modal::confirm("Test", "Message");
        assert_eq!(modal.selected_button_text(), Some("Yes"));

        let mut modal = modal;
        modal.select_next();
        assert_eq!(modal.selected_button_text(), Some("No"));
    }

    #[test]
    fn test_is_confirmed() {
        let modal = Modal::confirm("Test", "Message");
        assert!(modal.is_confirmed());

        let mut modal = modal;
        modal.select_next();
        assert!(!modal.is_confirmed());
    }

    #[test]
    fn test_is_cancelled() {
        let mut modal = Modal::confirm("Test", "Message");
        assert!(!modal.is_cancelled());

        modal.select_next();
        assert!(modal.is_cancelled());

        let modal = Modal::new("Test", "Message").with_buttons(vec!["Cancel".to_string()]);
        assert!(modal.is_cancelled());
    }

    #[test]
    fn test_empty_buttons_selected_button_text() {
        let modal = Modal {
            buttons: vec![],
            ..Default::default()
        };
        assert_eq!(modal.selected_button_text(), None);
    }

    #[test]
    fn test_modal_widget_creation() {
        let modal = Modal::new("Test", "Content");
        let _widget = ModalWidget::new(&modal);
        // Widget creation should not panic
    }

    #[test]
    fn test_centered_rect_calculation() {
        let modal = Modal::new("Test", "Content").with_size(40, 10);
        let widget = ModalWidget::new(&modal);

        let area = Rect::new(0, 0, 100, 50);
        let centered = widget.get_centered_rect(area);

        // Should be centered
        assert_eq!(centered.width, 40);
        assert_eq!(centered.height, 10);
        assert_eq!(centered.x, 30); // (100 - 40) / 2
        assert_eq!(centered.y, 20); // (50 - 10) / 2
    }

    #[test]
    fn test_centered_rect_clamp_to_area() {
        let modal = Modal::new("Test", "Content").with_size(200, 100);
        let widget = ModalWidget::new(&modal);

        let area = Rect::new(0, 0, 50, 30);
        let centered = widget.get_centered_rect(area);

        // Should be clamped to available area minus padding
        assert!(centered.width <= 50);
        assert!(centered.height <= 30);
    }
}
