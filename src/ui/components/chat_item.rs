//! Chat item component for rendering individual chats in the list.
//!
//! This module provides the [`ChatItemComponent`] which renders a single chat
//! entry showing the chat title, last message preview, timestamps, and status
//! indicators (pinned, muted, unread count, online status).

use ratatui::{
    buffer::Buffer,
    layout::Rect,
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph, Widget},
};
use unicode_width::UnicodeWidthStr;

use crate::types::{Chat, ChatType, MessageType, UserStatus};
use crate::ui::styles::{colors, Styles};
use crate::utils::{format_timestamp, truncate_string};

/// Configuration for how a chat item should be rendered.
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

/// A component that renders a single chat in the chat list.
///
/// The chat item displays:
/// - Chat title with truncation
/// - Status badges (NEW, pinned, muted, online)
/// - Unread count badge
/// - Timestamp of last message
/// - Message preview (optional)
/// - User online status (for private chats)
///
/// # Visual Design
///
/// Selected items are highlighted with a border. The border color indicates
/// focus state:
/// - Selected + Focused: Bright blue border
/// - Selected + Not focused: Cyan border  
/// - Not selected: Invisible/dark border (maintains consistent height)
///
/// # Example
///
/// ```rust,no_run
/// use ithil::ui::components::{ChatItemComponent, ChatItemConfig};
/// use ithil::types::Chat;
///
/// let chat = Chat::default();
/// let config = ChatItemConfig {
///     is_selected: true,
///     is_focused: true,
///     width: 40,
///     show_preview: true,
/// };
///
/// let item = ChatItemComponent::new(&chat, config);
/// // Render with ratatui...
/// ```
#[derive(Debug, Clone)]
pub struct ChatItemComponent<'a> {
    chat: &'a Chat,
    config: ChatItemConfig,
}

impl<'a> ChatItemComponent<'a> {
    /// Creates a new chat item component.
    ///
    /// # Arguments
    ///
    /// * `chat` - Reference to the chat data
    /// * `config` - Rendering configuration
    #[must_use]
    pub fn new(chat: &'a Chat, config: ChatItemConfig) -> Self {
        Self { chat, config }
    }

    /// Builds the first line containing title, badges, and timestamp.
    fn build_first_line(&self) -> Line<'a> {
        let mut spans = Vec::new();
        let width = self.config.width as usize;

        // Calculate space for title
        // Reserve: badges (~12), timestamp (~10), spacing (~3)
        let max_title_width = width.saturating_sub(25).max(10);

        // Chat title
        let title = if self.chat.title.is_empty() {
            format!("Chat {}", self.chat.id)
        } else {
            self.chat.title.clone()
        };
        let truncated_title = truncate_string(&title, max_title_width);

        let title_style = Styles::text_bright().add_modifier(Modifier::BOLD);
        spans.push(Span::styled(truncated_title, title_style));

        // Status badges
        let badges = self.build_badges();
        if !badges.is_empty() {
            spans.push(Span::raw(" "));
            spans.extend(badges);
        }

        // Calculate spacing for right-aligned content
        let left_content: String = spans.iter().map(|s| s.content.as_ref()).collect();
        let left_width = UnicodeWidthStr::width(left_content.as_str());

        // Right side: unread badge + timestamp
        let right_spans = self.build_right_side();
        let right_content: String = right_spans.iter().map(|s| s.content.as_ref()).collect();
        let right_width = UnicodeWidthStr::width(right_content.as_str());

        // Calculate padding (account for borders: 4 chars)
        let padding_width = width.saturating_sub(left_width + right_width + 4);
        let spacing = if padding_width > 0 {
            " ".repeat(padding_width)
        } else {
            " ".to_string()
        };

        spans.push(Span::raw(spacing));
        spans.extend(right_spans);

        Line::from(spans)
    }

    /// Builds badge spans for status indicators.
    fn build_badges(&self) -> Vec<Span<'a>> {
        let mut badges = Vec::new();

        // NEW badge for chats with new messages
        if self.chat.has_new_message {
            badges.push(Span::styled(
                " NEW ",
                Style::default()
                    .bg(colors::NORD8)
                    .fg(colors::NORD0)
                    .add_modifier(Modifier::BOLD),
            ));
        }

        // Pinned indicator
        if self.chat.is_pinned {
            if !badges.is_empty() {
                badges.push(Span::raw(" "));
            }
            badges.push(Span::styled("PIN", Styles::chat_pinned()));
        }

        // Muted indicator
        if self.chat.is_muted {
            if !badges.is_empty() {
                badges.push(Span::raw(" "));
            }
            badges.push(Span::styled("MUTE", Styles::chat_muted()));
        }

        // Online status for private chats
        if self.chat.chat_type == ChatType::Private && self.chat.user_status == UserStatus::Online {
            if !badges.is_empty() {
                badges.push(Span::raw(" "));
            }
            badges.push(Span::styled("*", Styles::status_online()));
        }

        badges
    }

    /// Builds the right side content (unread badge + timestamp).
    fn build_right_side(&self) -> Vec<Span<'a>> {
        let mut spans = Vec::new();

        // Unread count badge
        if self.chat.unread_count > 0 {
            let unread_text = if self.chat.unread_count > 99 {
                "99+".to_string()
            } else {
                self.chat.unread_count.to_string()
            };

            // Use cyan for new messages, red otherwise
            let bg_color = if self.chat.has_new_message {
                colors::NORD8
            } else {
                colors::NORD11
            };

            spans.push(Span::styled(
                format!(" {} ", unread_text),
                Style::default()
                    .bg(bg_color)
                    .fg(colors::NORD0)
                    .add_modifier(Modifier::BOLD),
            ));
            spans.push(Span::raw(" "));
        }

        // Timestamp
        if let Some(ref last_message) = self.chat.last_message {
            let timestamp = format_timestamp(last_message.date, true);
            // Replace spaces with non-breaking spaces
            let timestamp = timestamp.replace(' ', "\u{00A0}");

            let style = if self.config.is_selected && self.config.is_focused {
                Styles::text()
            } else {
                Styles::text_muted()
            };

            spans.push(Span::styled(timestamp, style));
        }

        spans
    }

    /// Builds the second line containing message preview and metadata.
    fn build_second_line(&self) -> Option<Line<'a>> {
        if !self.config.show_preview {
            return None;
        }

        let mut spans = Vec::new();
        let width = self.config.width as usize;

        // Indent
        spans.push(Span::raw("  "));

        // Message preview
        if let Some(preview) = self.build_preview() {
            spans.push(preview);
        }

        // Metadata (user status for private chats)
        if let Some(metadata) = self.build_metadata() {
            spans.push(Span::styled(" * ", Styles::text_muted()));
            spans.push(metadata);
        }

        // Truncate the entire line if needed
        let content: String = spans.iter().map(|s| s.content.as_ref()).collect();
        if UnicodeWidthStr::width(content.as_str()) > width.saturating_sub(6) {
            // Simplified: just return preview truncated
            if self.chat.last_message.is_some() {
                let preview_text = self.get_preview_text();
                let max_len = width.saturating_sub(10);
                let truncated = truncate_string(&preview_text, max_len);

                let style = if self.config.is_selected && self.config.is_focused {
                    Styles::text().add_modifier(Modifier::ITALIC)
                } else {
                    Styles::text_muted().add_modifier(Modifier::ITALIC)
                };

                return Some(Line::from(vec![
                    Span::raw("  "),
                    Span::styled(truncated, style),
                ]));
            }
        }

        if spans.len() > 1 {
            Some(Line::from(spans))
        } else {
            None
        }
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
                preview.push_str("Photo");
                if !msg.content.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&msg.content.caption);
                }
            }
            MessageType::Video => {
                preview.push_str("Video");
                if !msg.content.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&msg.content.caption);
                }
            }
            MessageType::Voice => preview.push_str("Voice message"),
            MessageType::VideoNote => preview.push_str("Video message"),
            MessageType::Audio => preview.push_str("Audio"),
            MessageType::Document => {
                preview.push_str("Document");
                if let Some(ref doc) = msg.content.document {
                    if !doc.file_name.is_empty() {
                        preview.push_str(": ");
                        preview.push_str(&doc.file_name);
                    }
                }
            }
            MessageType::Sticker => preview.push_str("Sticker"),
            MessageType::Animation => preview.push_str("GIF"),
            MessageType::Location => preview.push_str("Location"),
            MessageType::Contact => preview.push_str("Contact"),
            MessageType::Poll => {
                preview.push_str("Poll");
                if let Some(ref poll) = msg.content.poll {
                    preview.push_str(": ");
                    preview.push_str(&poll.question);
                }
            }
            MessageType::Venue => preview.push_str("Venue"),
            MessageType::Game => preview.push_str("Game"),
        }

        preview
    }

    /// Builds the preview span.
    fn build_preview(&self) -> Option<Span<'a>> {
        if self.chat.last_message.is_none() {
            return None;
        }

        let preview_text = self.get_preview_text();
        let max_len = (self.config.width as usize).saturating_sub(20);
        let truncated = truncate_string(&preview_text, max_len);

        let style = if self.config.is_selected && self.config.is_focused {
            Styles::text().add_modifier(Modifier::ITALIC)
        } else {
            Styles::text_muted().add_modifier(Modifier::ITALIC)
        };

        Some(Span::styled(truncated, style))
    }

    /// Builds metadata span (online status for private chats).
    fn build_metadata(&self) -> Option<Span<'a>> {
        if self.chat.chat_type != ChatType::Private {
            return None;
        }

        let status_text = match self.chat.user_status {
            UserStatus::Online => return None, // Already shown as badge
            UserStatus::Recently => "Recently",
            UserStatus::LastWeek => "Last week",
            UserStatus::LastMonth => "Last month",
            UserStatus::Offline => return None,
        };

        let style = if self.config.is_selected && self.config.is_focused {
            Styles::info()
        } else {
            Styles::text_muted()
        };

        Some(Span::styled(status_text, style))
    }

    /// Returns the height in lines this item will occupy.
    #[must_use]
    pub fn height(&self) -> u16 {
        // Border (2) + title line (1) + optional preview (1)
        if self.config.show_preview && self.chat.last_message.is_some() {
            4
        } else {
            3
        }
    }
}

impl Widget for ChatItemComponent<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        if area.width < 10 || area.height < 2 {
            return;
        }

        // Build content lines
        let mut lines = vec![self.build_first_line()];
        if let Some(second_line) = self.build_second_line() {
            lines.push(second_line);
        }

        // Determine border style based on selection state
        let (border_style, block_style) = if self.config.is_selected && self.config.is_focused {
            (Style::default().fg(colors::NORD8), Style::default())
        } else if self.config.is_selected {
            (Style::default().fg(colors::NORD7), Style::default())
        } else {
            (Style::default().fg(colors::NORD0), Style::default())
        };

        let block = Block::default()
            .borders(Borders::ALL)
            .border_style(border_style)
            .style(block_style);

        let paragraph = Paragraph::new(lines).block(block);
        paragraph.render(area, buf);
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
    fn test_chat_item_creation() {
        let chat = create_test_chat();
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        assert!(!item.config.is_selected);
        assert!(!item.config.is_focused);
    }

    #[test]
    fn test_preview_text() {
        let chat = create_test_chat();
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        let preview = item.get_preview_text();
        assert!(preview.contains("Hello, world!"));
    }

    #[test]
    fn test_height_with_preview() {
        let chat = create_test_chat();
        let config = ChatItemConfig {
            show_preview: true,
            ..Default::default()
        };
        let item = ChatItemComponent::new(&chat, config);

        assert_eq!(item.height(), 4);
    }

    #[test]
    fn test_height_without_preview() {
        let mut chat = create_test_chat();
        chat.last_message = None;
        let config = ChatItemConfig {
            show_preview: true,
            ..Default::default()
        };
        let item = ChatItemComponent::new(&chat, config);

        assert_eq!(item.height(), 3);
    }

    #[test]
    fn test_unread_badge() {
        let mut chat = create_test_chat();
        chat.unread_count = 150;
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        let right_spans = item.build_right_side();
        let text: String = right_spans.iter().map(|s| s.content.as_ref()).collect();
        assert!(text.contains("99+"));
    }

    #[test]
    fn test_outgoing_message_preview() {
        let mut chat = create_test_chat();
        if let Some(ref mut msg) = chat.last_message {
            msg.is_outgoing = true;
        }
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        let preview = item.get_preview_text();
        assert!(preview.starts_with("You: "));
    }

    #[test]
    fn test_media_preview_types() {
        let mut chat = create_test_chat();
        if let Some(ref mut msg) = chat.last_message {
            msg.content.content_type = MessageType::Photo;
            msg.content.text = String::new();
        }
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        let preview = item.get_preview_text();
        assert_eq!(preview, "Photo");
    }

    #[test]
    fn test_online_badge() {
        let mut chat = create_test_chat();
        chat.chat_type = ChatType::Private;
        chat.user_status = UserStatus::Online;
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        let badges = item.build_badges();
        let text: String = badges.iter().map(|s| s.content.as_ref()).collect();
        assert!(text.contains("*"));
    }

    #[test]
    fn test_group_no_online_badge() {
        let mut chat = create_test_chat();
        chat.chat_type = ChatType::Group;
        chat.user_status = UserStatus::Online;
        let config = ChatItemConfig::default();
        let item = ChatItemComponent::new(&chat, config);

        let badges = item.build_badges();
        let text: String = badges.iter().map(|s| s.content.as_ref()).collect();
        assert!(!text.contains("*"));
    }
}
