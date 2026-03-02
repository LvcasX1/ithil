//! Message rendering component.
//!
//! This module provides a widget for rendering individual Telegram messages
//! with proper formatting for different message types, selection state,
//! timestamps, and reply indicators.
//!
//! # Example
//!
//! ```rust,no_run
//! use ratatui::prelude::*;
//! use ithil::types::Message;
//! use ithil::ui::components::message::MessageWidget;
//!
//! let message = Message::default();
//! let widget = MessageWidget::new(&message, "Alice".to_string())
//!     .selected(true)
//!     .width(80);
//! ```

use ratatui::{
    buffer::Buffer,
    layout::Rect,
    text::{Line, Span},
    widgets::{Paragraph, Widget, Wrap},
};

use crate::types::{Message, MessageType};
use crate::ui::styles::Styles;
use crate::utils::format_timestamp;

/// A widget that renders a single message.
///
/// This widget handles the visual representation of a Telegram message,
/// including sender name, timestamp, content, and various indicators
/// for replies, edits, and selection state.
pub struct MessageWidget<'a> {
    /// The message to render
    message: &'a Message,
    /// Display name of the sender
    sender_name: String,
    /// Whether this message is currently selected
    is_selected: bool,
    /// Whether to show the timestamp
    show_timestamp: bool,
    /// Available width for rendering
    width: u16,
}

impl<'a> MessageWidget<'a> {
    /// Creates a new message widget.
    ///
    /// # Arguments
    ///
    /// * `message` - The message to render
    /// * `sender_name` - Display name of the message sender
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// use ithil::types::Message;
    /// use ithil::ui::components::message::MessageWidget;
    ///
    /// let message = Message::default();
    /// let widget = MessageWidget::new(&message, "Bob".to_string());
    /// ```
    #[must_use]
    #[allow(clippy::missing_const_for_fn)] // String parameter prevents const
    pub fn new(message: &'a Message, sender_name: String) -> Self {
        Self {
            message,
            sender_name,
            is_selected: false,
            show_timestamp: true,
            width: 80,
        }
    }

    /// Sets whether this message is selected.
    ///
    /// When selected, the message will have a selection marker and
    /// different styling.
    #[must_use]
    pub const fn selected(mut self, selected: bool) -> Self {
        self.is_selected = selected;
        self
    }

    /// Sets the available width for rendering.
    ///
    /// This affects text wrapping calculations.
    #[must_use]
    pub const fn width(mut self, width: u16) -> Self {
        self.width = width;
        self
    }

    /// Sets whether to show the timestamp.
    #[must_use]
    #[allow(dead_code)]
    pub const fn show_timestamp(mut self, show: bool) -> Self {
        self.show_timestamp = show;
        self
    }

    /// Calculates the height needed to render this message.
    ///
    /// This takes into account:
    /// - Header line (sender + timestamp)
    /// - Content lines (with wrapping)
    /// - Optional reply indicator
    ///
    /// # Returns
    ///
    /// The number of terminal rows needed to render this message.
    #[must_use]
    #[allow(clippy::cast_possible_truncation)]
    pub fn height(&self) -> u16 {
        // Header line (sender + timestamp)
        let mut lines: u16 = 1;

        // Content
        let content = self.get_content_text();
        let content_width = self.width.saturating_sub(4) as usize; // Account for padding
        if content_width > 0 && !content.is_empty() {
            // Count lines in content, accounting for wrapping
            for line in content.lines() {
                let line_count = if line.is_empty() {
                    1
                } else {
                    (line.len().saturating_sub(1) / content_width + 1) as u16
                };
                lines = lines.saturating_add(line_count);
            }
            // If no newlines in content, count as single wrapped block
            if !content.contains('\n') && content.lines().count() == 1 {
                lines = 1 + (content.len().saturating_sub(1) / content_width + 1) as u16;
            }
        } else {
            lines = lines.saturating_add(1); // At least one content line
        }

        // Reply indicator
        if self.message.reply_to_message_id > 0 {
            lines = lines.saturating_add(1);
        }

        lines.max(2) // Minimum 2 lines
    }

    /// Gets the text content to display for this message.
    ///
    /// This handles different message types and returns appropriate
    /// text representations.
    fn get_content_text(&self) -> String {
        match self.message.content.content_type {
            MessageType::Text => self.message.content.text.clone(),
            MessageType::Photo => {
                if self.message.content.caption.is_empty() {
                    "[Photo]".to_string()
                } else {
                    format!("[Photo] {}", self.message.content.caption)
                }
            },
            MessageType::Video => {
                if self.message.content.caption.is_empty() {
                    "[Video]".to_string()
                } else {
                    format!("[Video] {}", self.message.content.caption)
                }
            },
            MessageType::Voice => "[Voice message]".to_string(),
            MessageType::VideoNote => "[Video note]".to_string(),
            MessageType::Audio => {
                if self.message.content.caption.is_empty() {
                    "[Audio]".to_string()
                } else {
                    format!("[Audio] {}", self.message.content.caption)
                }
            },
            MessageType::Document => self.message.content.document.as_ref().map_or_else(
                || "[Document]".to_string(),
                |doc| format!("[Document: {}]", doc.file_name),
            ),
            MessageType::Sticker => self.message.content.sticker.as_ref().map_or_else(
                || "[Sticker]".to_string(),
                |sticker| format!("[Sticker: {}]", sticker.emoji),
            ),
            MessageType::Animation => "[GIF]".to_string(),
            MessageType::Location => "[Location]".to_string(),
            MessageType::Contact => "[Contact]".to_string(),
            MessageType::Poll => self.message.content.poll.as_ref().map_or_else(
                || "[Poll]".to_string(),
                |poll| format!("[Poll: {}]", poll.question),
            ),
            MessageType::Venue => "[Venue]".to_string(),
            MessageType::Game => "[Game]".to_string(),
        }
    }

    /// Builds the lines to render for this message.
    fn build_lines(&self) -> Vec<Line<'static>> {
        let mut lines = Vec::new();

        // Selection indicator
        let selection_marker = if self.is_selected { "▶ " } else { "  " };

        // Header: sender name + timestamp
        let timestamp = if self.show_timestamp {
            format_timestamp(self.message.date, true)
        } else {
            String::new()
        };

        let header_style = if self.message.is_outgoing {
            Styles::message_outgoing()
        } else {
            Styles::username()
        };

        let mut header_spans = vec![
            Span::styled(
                selection_marker.to_string(),
                if self.is_selected {
                    Styles::highlight()
                } else {
                    Styles::text()
                },
            ),
            Span::styled(self.sender_name.clone(), header_style),
        ];

        if !timestamp.is_empty() {
            header_spans.push(Span::raw(" "));
            header_spans.push(Span::styled(timestamp, Styles::timestamp()));
        }

        if self.message.is_edited {
            header_spans.push(Span::styled(" (edited)".to_string(), Styles::text_muted()));
        }

        lines.push(Line::from(header_spans));

        // Reply indicator
        if self.message.reply_to_message_id > 0 {
            lines.push(Line::from(vec![
                Span::raw("  "),
                Span::styled("↩ Reply to message".to_string(), Styles::text_muted()),
            ]));
        }

        // Content
        let content = self.get_content_text();
        let content_style = if self.is_selected {
            Styles::selected()
        } else {
            Styles::text()
        };

        if content.is_empty() {
            lines.push(Line::from(vec![
                Span::raw("  "),
                Span::styled(String::new(), content_style),
            ]));
        } else {
            for line in content.lines() {
                lines.push(Line::from(vec![
                    Span::raw("  "),
                    Span::styled(line.to_string(), content_style),
                ]));
            }
            // Handle case where content has no newlines
            if !content.contains('\n') && content.lines().count() == 0 {
                lines.push(Line::from(vec![
                    Span::raw("  "),
                    Span::styled(content, content_style),
                ]));
            }
        }

        lines
    }
}

impl Widget for MessageWidget<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        let lines = self.build_lines();
        let paragraph = Paragraph::new(lines).wrap(Wrap { trim: false });
        paragraph.render(area, buf);
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::MessageContent;
    use chrono::Utc;

    fn create_test_message(text: &str, is_outgoing: bool) -> Message {
        Message {
            id: 1,
            chat_id: 100,
            sender_id: 42,
            content: MessageContent {
                content_type: MessageType::Text,
                text: text.to_string(),
                ..Default::default()
            },
            date: Utc::now(),
            is_outgoing,
            ..Default::default()
        }
    }

    #[test]
    fn test_new_message_widget() {
        let msg = create_test_message("Hello, world!", false);
        let widget = MessageWidget::new(&msg, "Alice".to_string());

        assert!(!widget.is_selected);
        assert!(widget.show_timestamp);
        assert_eq!(widget.width, 80);
    }

    #[test]
    fn test_selected_message() {
        let msg = create_test_message("Hello", false);
        let widget = MessageWidget::new(&msg, "Bob".to_string()).selected(true);

        assert!(widget.is_selected);
    }

    #[test]
    fn test_width_setting() {
        let msg = create_test_message("Test", false);
        let widget = MessageWidget::new(&msg, "Charlie".to_string()).width(120);

        assert_eq!(widget.width, 120);
    }

    #[test]
    fn test_height_calculation() {
        let msg = create_test_message("Short message", false);
        let widget = MessageWidget::new(&msg, "Dan".to_string()).width(80);

        let height = widget.height();
        assert!(height >= 2); // At least header + content
    }

    #[test]
    fn test_height_with_reply() {
        let mut msg = create_test_message("Reply message", false);
        msg.reply_to_message_id = 123;

        let widget = MessageWidget::new(&msg, "Eve".to_string()).width(80);
        let height = widget.height();

        assert!(height >= 3); // Header + reply indicator + content
    }

    #[test]
    fn test_content_text_for_text_message() {
        let msg = create_test_message("Hello, world!", false);
        let widget = MessageWidget::new(&msg, "Frank".to_string());

        assert_eq!(widget.get_content_text(), "Hello, world!");
    }

    #[test]
    fn test_content_text_for_photo() {
        let msg = Message {
            content: MessageContent {
                content_type: MessageType::Photo,
                caption: "Nice photo".to_string(),
                ..Default::default()
            },
            ..Default::default()
        };
        let widget = MessageWidget::new(&msg, "Grace".to_string());

        assert_eq!(widget.get_content_text(), "[Photo] Nice photo");
    }

    #[test]
    fn test_content_text_for_sticker_with_emoji() {
        let msg = Message {
            content: MessageContent {
                content_type: MessageType::Sticker,
                sticker: Some(Box::new(crate::types::Sticker {
                    emoji: "😀".to_string(),
                    ..Default::default()
                })),
                ..Default::default()
            },
            ..Default::default()
        };
        let widget = MessageWidget::new(&msg, "Henry".to_string());

        assert_eq!(widget.get_content_text(), "[Sticker: 😀]");
    }

    #[test]
    fn test_build_lines_basic() {
        let msg = create_test_message("Test message", false);
        let widget = MessageWidget::new(&msg, "Ivan".to_string());

        let lines = widget.build_lines();
        assert!(!lines.is_empty());
        assert!(lines.len() >= 2); // Header + content
    }

    #[test]
    fn test_build_lines_with_selection() {
        let msg = create_test_message("Selected", false);
        let widget = MessageWidget::new(&msg, "Julia".to_string()).selected(true);

        let lines = widget.build_lines();
        // First line should contain selection marker
        let first_line_text: String = lines[0].spans.iter().map(|s| s.content.as_ref()).collect();
        assert!(first_line_text.starts_with('▶'));
    }

    #[test]
    fn test_build_lines_with_edit_indicator() {
        let mut msg = create_test_message("Edited message", false);
        msg.is_edited = true;

        let widget = MessageWidget::new(&msg, "Kate".to_string());
        let lines = widget.build_lines();

        // First line should contain "(edited)"
        let first_line_text: String = lines[0].spans.iter().map(|s| s.content.as_ref()).collect();
        assert!(first_line_text.contains("(edited)"));
    }
}
