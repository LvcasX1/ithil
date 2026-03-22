//! Chat item component for rendering individual chats in the list.
//!
//! This module provides the [`ChatItemBuilder`] which creates styled [`ListItem`]
//! entries for display in a Ratatui [`List`] widget. Each chat entry shows the
//! chat title, last message preview, timestamps, and status indicators.
//!
//! # Ratatui Standards
//!
//! This component follows modern Ratatui patterns:
//! - Uses [`ListItem`] for proper integration with the [`List`] widget
//! - Leverages [`Text`] and [`Line`] for multi-line items
//! - Returns owned `ListItem<'static>` for flexible use with stateful widgets
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::components::ChatItemBuilder;
//! use ithil::types::Chat;
//! use ratatui::widgets::{List, ListItem};
//!
//! let chats: Vec<Chat> = vec![]; // Your chats
//! let items: Vec<ListItem> = chats
//!     .iter()
//!     .map(|chat| ChatItemBuilder::new(chat, 40).build())
//!     .collect();
//!
//! let list = List::new(items)
//!     .highlight_symbol("▌ ")
//!     .highlight_style(ratatui::style::Style::default());
//! ```

use ratatui::{
    style::{Modifier, Style},
    text::{Line, Span, Text},
    widgets::ListItem,
};
use unicode_width::UnicodeWidthStr;

use crate::types::{Chat, ChatType, MessageType, UserStatus};
use crate::ui::styles::{colors, Styles};
use crate::utils::{format_timestamp, truncate_string};

/// Builder for creating styled [`ListItem`] entries from chat data.
///
/// This follows the builder pattern for flexible configuration while
/// producing standard Ratatui widgets. The builder produces owned
/// `ListItem<'static>` instances to allow flexible use with stateful
/// list rendering.
///
/// # Visual Layout
///
/// Each chat item displays two lines:
///
/// ```text
/// ┌────────────────────────────────────────┐
/// │ Chat Title  📌 ●              [3] 12:30 │
/// │   Last message preview...              │
/// └────────────────────────────────────────┘
/// ```
///
/// Where:
/// - `📌` appears for pinned chats
/// - `●` appears for online users (private chats)
/// - `[3]` is the unread count badge
/// - `12:30` is the timestamp
#[derive(Debug, Clone)]
pub struct ChatItemBuilder<'a> {
    chat: &'a Chat,
    width: u16,
    show_preview: bool,
}

impl<'a> ChatItemBuilder<'a> {
    /// Creates a new chat item builder.
    ///
    /// # Arguments
    ///
    /// * `chat` - Reference to the chat data
    /// * `width` - Available width for rendering
    #[must_use]
    pub const fn new(chat: &'a Chat, width: u16) -> Self {
        Self {
            chat,
            width,
            show_preview: true,
        }
    }

    /// Sets whether to show the message preview line.
    #[must_use]
    pub const fn show_preview(mut self, show: bool) -> Self {
        self.show_preview = show;
        self
    }

    /// Builds the [`ListItem`] for this chat.
    ///
    /// The returned item is fully owned (`'static` lifetime) and can be used
    /// directly with Ratatui's [`List`] widget, including stateful rendering.
    #[must_use]
    pub fn build(self) -> ListItem<'static> {
        let mut lines = Vec::new();

        // Title line
        lines.push(self.build_title_line());

        // Preview line (if enabled and message exists)
        if self.show_preview {
            if let Some(preview_line) = self.build_preview_line() {
                lines.push(preview_line);
            }
        }

        // Add a blank line at the bottom for visual separation between items
        lines.push(Line::default());

        ListItem::new(Text::from(lines))
    }

    /// Builds the title line with chat name, badges, and timestamp.
    fn build_title_line(&self) -> Line<'static> {
        let mut spans: Vec<Span<'static>> = Vec::new();
        let width = self.width as usize;

        // Calculate available space for title
        // Reserve space for: badges (~8), timestamp (~8), unread (~4), padding (~4)
        let reserved_width = 24_usize;
        let max_title_width = width.saturating_sub(reserved_width).max(8);

        // Chat title (owned string)
        let title = if self.chat.title.is_empty() {
            format!("Chat {}", self.chat.id)
        } else {
            self.chat.title.clone()
        };
        let truncated_title = truncate_string(&title, max_title_width);

        // Title styling: bold, and highlighted if has new messages
        let title_style = if self.chat.has_new_message {
            Style::default()
                .fg(colors::NORD6)
                .add_modifier(Modifier::BOLD)
        } else {
            Style::default()
                .fg(colors::NORD4)
                .add_modifier(Modifier::BOLD)
        };
        spans.push(Span::styled(truncated_title, title_style));

        // Add status badges inline
        self.append_badges(&mut spans);

        // Calculate current content width
        let left_content: String = spans.iter().map(|s| s.content.as_ref()).collect();
        let left_width = UnicodeWidthStr::width(left_content.as_str());

        // Build right side (unread badge + timestamp)
        let right_spans = self.build_right_content();
        let right_content: String = right_spans.iter().map(|s| s.content.as_ref()).collect();
        let right_width = UnicodeWidthStr::width(right_content.as_str());

        // Calculate padding to right-align
        let padding = width.saturating_sub(left_width + right_width);
        if padding > 0 {
            spans.push(Span::raw(" ".repeat(padding)));
        }

        spans.extend(right_spans);

        Line::from(spans)
    }

    /// Appends status badge spans to the given vector.
    fn append_badges(&self, spans: &mut Vec<Span<'static>>) {
        // Pinned indicator with icon
        if self.chat.is_pinned {
            spans.push(Span::raw(" "));
            spans.push(Span::styled(
                "📌".to_string(),
                Style::default().fg(colors::NORD13),
            ));
        }

        // Muted indicator
        if self.chat.is_muted {
            spans.push(Span::raw(" "));
            spans.push(Span::styled(
                "🔇".to_string(),
                Style::default().fg(colors::NORD3),
            ));
        }

        // Online status indicator for private chats
        if self.chat.chat_type == ChatType::Private && self.chat.user_status == UserStatus::Online {
            spans.push(Span::raw(" "));
            spans.push(Span::styled(
                "●".to_string(),
                Style::default().fg(colors::NORD14),
            ));
        }
    }

    /// Builds the right-side content (unread badge + timestamp).
    fn build_right_content(&self) -> Vec<Span<'static>> {
        let mut spans: Vec<Span<'static>> = Vec::new();

        // Unread count badge
        if self.chat.unread_count > 0 {
            let unread_text = if self.chat.unread_count > 99 {
                "99+".to_string()
            } else {
                self.chat.unread_count.to_string()
            };

            // Style based on importance
            let badge_style = if self.chat.has_new_message {
                Style::default()
                    .bg(colors::NORD8)
                    .fg(colors::NORD0)
                    .add_modifier(Modifier::BOLD)
            } else if self.chat.is_muted {
                Style::default().bg(colors::NORD3).fg(colors::NORD0)
            } else {
                Style::default()
                    .bg(colors::NORD11)
                    .fg(colors::NORD0)
                    .add_modifier(Modifier::BOLD)
            };

            spans.push(Span::styled(format!(" {unread_text} "), badge_style));
            spans.push(Span::raw(" "));
        }

        // Timestamp
        if let Some(ref last_message) = self.chat.last_message {
            let timestamp = format_timestamp(last_message.date, true);
            // Use non-breaking spaces to prevent wrapping
            let timestamp = timestamp.replace(' ', "\u{00A0}");
            spans.push(Span::styled(timestamp, Styles::text_muted()));
        }

        spans
    }

    /// Builds the preview line showing the last message.
    fn build_preview_line(&self) -> Option<Line<'static>> {
        let preview_text = self.get_preview_text();
        if preview_text.is_empty() {
            return None;
        }

        let max_len = (self.width as usize).saturating_sub(4);
        let truncated = truncate_string(&preview_text, max_len);

        let style = Style::default()
            .fg(colors::NORD3)
            .add_modifier(Modifier::ITALIC);

        Some(Line::from(vec![
            Span::raw("  ".to_string()), // Indent for visual hierarchy
            Span::styled(truncated, style),
        ]))
    }

    /// Gets the preview text for the last message.
    fn get_preview_text(&self) -> String {
        let Some(ref msg) = self.chat.last_message else {
            return String::new();
        };

        let mut preview = if msg.is_outgoing {
            "You: ".to_string()
        } else {
            String::new()
        };

        match msg.content.content_type {
            MessageType::Text => preview.push_str(&msg.content.text),
            MessageType::Photo => {
                preview.push_str("📷 Photo");
                if !msg.content.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&msg.content.caption);
                }
            },
            MessageType::Video => {
                preview.push_str("🎬 Video");
                if !msg.content.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&msg.content.caption);
                }
            },
            MessageType::Voice => preview.push_str("🎤 Voice message"),
            MessageType::VideoNote => preview.push_str("📹 Video message"),
            MessageType::Audio => preview.push_str("🎵 Audio"),
            MessageType::Document => {
                preview.push_str("📎 Document");
                if let Some(ref doc) = msg.content.document {
                    if !doc.file_name.is_empty() {
                        preview.push_str(": ");
                        preview.push_str(&doc.file_name);
                    }
                }
            },
            MessageType::Sticker => preview.push_str("🎨 Sticker"),
            MessageType::Animation => preview.push_str("GIF"),
            MessageType::Location => preview.push_str("📍 Location"),
            MessageType::Contact => preview.push_str("👤 Contact"),
            MessageType::Poll => {
                preview.push_str("📊 Poll");
                if let Some(ref poll) = msg.content.poll {
                    preview.push_str(": ");
                    preview.push_str(&poll.question);
                }
            },
            MessageType::Venue => preview.push_str("📍 Venue"),
            MessageType::Game => preview.push_str("🎮 Game"),
        }

        preview
    }

    /// Returns the expected height of this item in lines.
    #[must_use]
    pub const fn height(&self) -> u16 {
        if self.show_preview && self.chat.last_message.is_some() {
            3 // Title + preview + spacing
        } else {
            2 // Title + spacing
        }
    }
}

// ============================================================================
// Legacy compatibility - ChatItemComponent and ChatItemConfig
// ============================================================================

/// Configuration for how a chat item should be rendered.
///
/// This is kept for backward compatibility but the new approach
/// uses [`ChatItemBuilder`] with Ratatui's standard [`List`] widget.
#[derive(Debug, Clone)]
pub struct ChatItemConfig {
    /// Whether this chat is currently selected
    pub is_selected: bool,
    /// Whether the chat list pane is focused
    pub is_focused: bool,
    /// Available width for rendering
    pub width: u16,
    /// Whether to show the message preview
    pub show_preview: bool,
}

impl Default for ChatItemConfig {
    fn default() -> Self {
        Self {
            is_selected: false,
            is_focused: false,
            width: 40,
            show_preview: true,
        }
    }
}

/// Legacy component wrapper for backward compatibility.
///
/// Prefer using [`ChatItemBuilder`] directly with Ratatui's [`List`] widget.
#[derive(Debug, Clone)]
pub struct ChatItemComponent<'a> {
    chat: &'a Chat,
    config: ChatItemConfig,
}

impl<'a> ChatItemComponent<'a> {
    /// Creates a new chat item component.
    #[must_use]
    pub const fn new(chat: &'a Chat, config: ChatItemConfig) -> Self {
        Self { chat, config }
    }

    /// Converts to a [`ListItem`] for use with Ratatui's [`List`] widget.
    #[must_use]
    pub fn to_list_item(&self) -> ListItem<'static> {
        ChatItemBuilder::new(self.chat, self.config.width)
            .show_preview(self.config.show_preview)
            .build()
    }

    /// Returns the height in lines this item will occupy.
    #[must_use]
    pub const fn height(&self) -> u16 {
        ChatItemBuilder::new(self.chat, self.config.width)
            .show_preview(self.config.show_preview)
            .height()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::{Message, MessageContent};
    use chrono::Utc;

    fn create_test_chat() -> Chat {
        Chat {
            id: 12345,
            title: "Test Chat".to_string(),
            chat_type: ChatType::Private,
            unread_count: 5,
            is_pinned: true,
            last_message: Some(Box::new(Message {
                id: 1,
                content: MessageContent {
                    text: "Hello, world!".to_string(),
                    ..Default::default()
                },
                date: Utc::now(),
                ..Default::default()
            })),
            ..Default::default()
        }
    }

    #[test]
    fn test_builder_creation() {
        let chat = create_test_chat();
        let builder = ChatItemBuilder::new(&chat, 40);
        assert_eq!(builder.width, 40);
        assert!(builder.show_preview);
    }

    #[test]
    fn test_builder_produces_list_item() {
        let chat = create_test_chat();
        let item = ChatItemBuilder::new(&chat, 40).build();
        // ListItem should have non-zero height (lines are present)
        assert!(item.height() > 0);
    }

    #[test]
    fn test_preview_text_with_message() {
        let chat = create_test_chat();
        let builder = ChatItemBuilder::new(&chat, 40);
        let preview = builder.get_preview_text();
        assert!(preview.contains("Hello, world!"));
    }

    #[test]
    fn test_preview_text_outgoing() {
        let mut chat = create_test_chat();
        if let Some(ref mut msg) = chat.last_message {
            msg.is_outgoing = true;
        }
        let builder = ChatItemBuilder::new(&chat, 40);
        let preview = builder.get_preview_text();
        assert!(preview.starts_with("You: "));
    }

    #[test]
    fn test_height_with_preview() {
        let chat = create_test_chat();
        let builder = ChatItemBuilder::new(&chat, 40).show_preview(true);
        assert_eq!(builder.height(), 3);
    }

    #[test]
    fn test_height_without_message() {
        let mut chat = create_test_chat();
        chat.last_message = None;
        let builder = ChatItemBuilder::new(&chat, 40).show_preview(true);
        assert_eq!(builder.height(), 2);
    }

    #[test]
    fn test_media_preview_types() {
        let mut chat = create_test_chat();
        if let Some(ref mut msg) = chat.last_message {
            msg.content.content_type = MessageType::Photo;
            msg.content.text = String::new();
        }
        let builder = ChatItemBuilder::new(&chat, 40);
        let preview = builder.get_preview_text();
        assert!(preview.contains("Photo"));
    }

    #[test]
    fn test_legacy_component_compatibility() {
        let chat = create_test_chat();
        let config = ChatItemConfig::default();
        let component = ChatItemComponent::new(&chat, config);
        let _item = component.to_list_item();
        assert!(component.height() >= 2);
    }

    #[test]
    fn test_unread_badge_capped() {
        let mut chat = create_test_chat();
        chat.unread_count = 150;
        let builder = ChatItemBuilder::new(&chat, 60);
        let right_spans = builder.build_right_content();
        let text: String = right_spans.iter().map(|s| s.content.as_ref()).collect();
        assert!(text.contains("99+"));
    }
}
