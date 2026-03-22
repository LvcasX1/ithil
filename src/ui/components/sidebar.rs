//! Sidebar (info panel) component.
//!
//! This module provides the sidebar model and widget for displaying
//! information about the selected chat, including:
//! - Chat title and type
//! - User information (for private chats)
//! - Member counts (for groups/channels)
//! - Chat settings (pinned, muted, unread count)
//!
//! # Architecture
//!
//! The sidebar follows the Model-View-Update pattern:
//! - [`SidebarModel`]: Holds the current chat and user information
//! - [`SidebarWidget`]: Renders the model to the terminal
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::types::{Chat, User};
//! use ithil::ui::components::sidebar::{SidebarModel, SidebarWidget};
//!
//! let mut model = SidebarModel::new();
//! model.set_chat(Chat::default(), None);
//!
//! // In render function:
//! // let widget = SidebarWidget::new(&model).focused(true);
//! ```

use ratatui::{
    buffer::Buffer,
    layout::Rect,
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph, Widget},
};

use crate::types::{Chat, ChatType, User, UserStatus};
use crate::ui::styles::Styles;

/// Model for the sidebar (info panel).
///
/// This struct holds information about the currently selected chat
/// for display in the sidebar.
#[derive(Debug, Clone, Default)]
pub struct SidebarModel {
    /// Currently displayed chat
    pub chat: Option<Chat>,
    /// User information (for private chats)
    pub user: Option<User>,
    /// Member count (for groups/channels)
    pub member_count: Option<i32>,
    /// Online member count (for groups)
    pub online_count: Option<i32>,
    /// Chat description/bio
    pub description: Option<String>,
}

impl SidebarModel {
    /// Creates a new sidebar model.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::sidebar::SidebarModel;
    ///
    /// let model = SidebarModel::new();
    /// assert!(model.chat.is_none());
    /// ```
    #[must_use]
    pub const fn new() -> Self {
        Self {
            chat: None,
            user: None,
            member_count: None,
            online_count: None,
            description: None,
        }
    }

    /// Sets the current chat to display.
    ///
    /// This resets the group info (member count, online count, description)
    /// which should be populated separately after fetching.
    ///
    /// # Arguments
    ///
    /// * `chat` - The chat to display
    /// * `user` - Optional user information (for private chats)
    pub fn set_chat(&mut self, chat: Chat, user: Option<User>) {
        self.chat = Some(chat);
        self.user = user;
        // Reset stats - will be populated later via API call
        self.member_count = None;
        self.online_count = None;
        self.description = None;
    }

    /// Sets the group/channel information.
    ///
    /// This should be called after fetching the full chat info from the API.
    ///
    /// # Arguments
    ///
    /// * `member_count` - Total number of members
    /// * `online_count` - Number of online members (if available)
    /// * `description` - Chat description/bio
    pub fn set_group_info(
        &mut self,
        member_count: i32,
        online_count: Option<i32>,
        description: Option<String>,
    ) {
        self.member_count = Some(member_count);
        self.online_count = online_count;
        self.description = description;
    }

    /// Clears all sidebar information.
    pub fn clear(&mut self) {
        self.chat = None;
        self.user = None;
        self.member_count = None;
        self.online_count = None;
        self.description = None;
    }

    /// Returns `true` if a chat is currently set.
    #[must_use]
    pub const fn has_chat(&self) -> bool {
        self.chat.is_some()
    }
}

/// Widget for rendering the sidebar.
///
/// Displays information about the currently selected chat including
/// title, type, user status, member counts, and chat settings.
pub struct SidebarWidget<'a> {
    /// Reference to the sidebar model
    model: &'a SidebarModel,
    /// Whether this pane is focused
    is_focused: bool,
}

impl<'a> SidebarWidget<'a> {
    /// Creates a new sidebar widget.
    ///
    /// # Arguments
    ///
    /// * `model` - Reference to the sidebar model
    #[must_use]
    pub const fn new(model: &'a SidebarModel) -> Self {
        Self {
            model,
            is_focused: false,
        }
    }

    /// Sets whether this pane is focused.
    #[must_use]
    pub const fn focused(mut self, focused: bool) -> Self {
        self.is_focused = focused;
        self
    }

    /// Builds the lines to display for the current chat.
    fn build_content_lines(&self) -> Vec<Line<'static>> {
        let Some(chat) = self.model.chat.as_ref() else {
            return vec![Line::from(Span::styled(
                "Select a chat to see info",
                Styles::text_muted(),
            ))];
        };

        let mut lines: Vec<Line<'static>> = Vec::new();

        // Title
        lines.push(Line::from(vec![Span::styled(
            chat.title.clone(),
            Styles::highlight(),
        )]));
        lines.push(Line::from("")); // spacer

        // Chat type
        let type_str = match chat.chat_type {
            ChatType::Private => "Private Chat",
            ChatType::Group => "Group",
            ChatType::Supergroup => "Supergroup",
            ChatType::Channel => "Channel",
            ChatType::Secret => "Secret Chat",
        };
        lines.push(Line::from(vec![
            Span::styled("Type: ", Styles::text_muted()),
            Span::styled(type_str, Styles::text()),
        ]));

        // Username if available
        if !chat.username.is_empty() {
            lines.push(Line::from(vec![
                Span::styled("Username: ", Styles::text_muted()),
                Span::styled(format!("@{}", chat.username), Styles::text_accent()),
            ]));
        }

        // For private chats, show user info
        if chat.chat_type == ChatType::Private {
            self.add_user_info_lines(&mut lines);
        } else {
            // For groups/channels, show member count
            self.add_group_info_lines(&mut lines, chat);
        }

        // Description if available
        if let Some(ref desc) = self.model.description {
            if !desc.is_empty() {
                lines.push(Line::from("")); // spacer
                lines.push(Line::from(vec![Span::styled(
                    "About:",
                    Styles::text_muted(),
                )]));
                // Word-wrap description (take first 5 lines)
                for line in desc.lines().take(5) {
                    lines.push(Line::from(vec![Span::styled(
                        line.to_string(),
                        Styles::text(),
                    )]));
                }
            }
        }

        // Chat settings
        lines.push(Line::from("")); // spacer
        lines.push(Line::from(vec![Span::styled(
            "─── Settings ───",
            Styles::text_muted(),
        )]));

        if chat.is_pinned {
            lines.push(Line::from(vec![Span::styled(
                "📌 Pinned",
                Styles::chat_pinned(),
            )]));
        }
        if chat.is_muted {
            lines.push(Line::from(vec![Span::styled(
                "🔇 Muted",
                Styles::chat_muted(),
            )]));
        }

        // Unread count
        if chat.unread_count > 0 {
            lines.push(Line::from(vec![Span::styled(
                format!("📬 {} unread", chat.unread_count),
                Styles::chat_unread(),
            )]));
        }

        lines
    }

    /// Adds user-specific information lines for private chats.
    fn add_user_info_lines(&self, lines: &mut Vec<Line<'static>>) {
        let Some(ref user) = self.model.user else {
            return;
        };

        lines.push(Line::from("")); // spacer

        // Phone if available
        if !user.phone_number.is_empty() {
            lines.push(Line::from(vec![
                Span::styled("Phone: ", Styles::text_muted()),
                Span::styled(user.phone_number.clone(), Styles::text()),
            ]));
        }

        // Status
        let (status_str, status_style) = match user.status {
            UserStatus::Online => ("Online", Styles::status_online()),
            UserStatus::Offline => ("Offline", Styles::status_offline()),
            UserStatus::Recently => ("Last seen recently", Styles::text_muted()),
            UserStatus::LastWeek => ("Last seen within a week", Styles::text_muted()),
            UserStatus::LastMonth => ("Last seen within a month", Styles::text_muted()),
        };
        lines.push(Line::from(vec![
            Span::styled("Status: ", Styles::text_muted()),
            Span::styled(status_str, status_style),
        ]));

        // Badges
        let mut badges = Vec::new();
        if user.is_verified {
            badges.push("✓ Verified");
        }
        if user.is_premium {
            badges.push("⭐ Premium");
        }
        if user.is_bot {
            badges.push("🤖 Bot");
        }
        if !badges.is_empty() {
            lines.push(Line::from(vec![Span::styled(
                badges.join(" "),
                Styles::text_accent(),
            )]));
        }
    }

    /// Adds group/channel-specific information lines.
    fn add_group_info_lines(&self, lines: &mut Vec<Line<'static>>, chat: &Chat) {
        if let Some(count) = self.model.member_count {
            lines.push(Line::from("")); // spacer

            let label = if chat.chat_type == ChatType::Channel {
                "Subscribers"
            } else {
                "Members"
            };

            lines.push(Line::from(vec![
                Span::styled(format!("{label}: "), Styles::text_muted()),
                Span::styled(count.to_string(), Styles::text()),
            ]));

            if let Some(online) = self.model.online_count {
                lines.push(Line::from(vec![
                    Span::styled("Online: ", Styles::text_muted()),
                    Span::styled(online.to_string(), Styles::status_online()),
                ]));
            }
        }
    }
}

impl Widget for SidebarWidget<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        let border_style = if self.is_focused {
            Styles::border_focused()
        } else {
            Styles::border()
        };

        let block = Block::default()
            .title(" Info ")
            .borders(Borders::ALL)
            .border_style(border_style);

        let inner = block.inner(area);
        block.render(area, buf);

        let lines = self.build_content_lines();
        let paragraph = Paragraph::new(lines);
        paragraph.render(inner, buf);
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::ChatType;

    fn create_test_chat(id: i64, title: &str, chat_type: ChatType) -> Chat {
        Chat {
            id,
            title: title.to_string(),
            chat_type,
            ..Default::default()
        }
    }

    fn create_test_user(id: i64, first_name: &str) -> User {
        User {
            id,
            first_name: first_name.to_string(),
            status: UserStatus::Online,
            ..Default::default()
        }
    }

    #[test]
    fn test_new_model() {
        let model = SidebarModel::new();

        assert!(model.chat.is_none());
        assert!(model.user.is_none());
        assert!(model.member_count.is_none());
        assert!(model.online_count.is_none());
        assert!(model.description.is_none());
    }

    #[test]
    fn test_default_model() {
        let model = SidebarModel::default();

        assert!(model.chat.is_none());
        assert!(!model.has_chat());
    }

    #[test]
    fn test_set_chat_private() {
        let mut model = SidebarModel::new();
        let chat = create_test_chat(1, "Test User", ChatType::Private);
        let user = create_test_user(1, "Test");

        model.set_chat(chat, Some(user));

        assert!(model.has_chat());
        assert_eq!(model.chat.as_ref().unwrap().title, "Test User");
        assert_eq!(model.user.as_ref().unwrap().first_name, "Test");
        // Stats should be reset
        assert!(model.member_count.is_none());
    }

    #[test]
    fn test_set_chat_group() {
        let mut model = SidebarModel::new();
        let chat = create_test_chat(1, "Test Group", ChatType::Supergroup);

        model.set_chat(chat, None);

        assert!(model.has_chat());
        assert_eq!(model.chat.as_ref().unwrap().chat_type, ChatType::Supergroup);
        assert!(model.user.is_none());
    }

    #[test]
    fn test_set_group_info() {
        let mut model = SidebarModel::new();
        let chat = create_test_chat(1, "Test Group", ChatType::Supergroup);

        model.set_chat(chat, None);
        model.set_group_info(100, Some(25), Some("A test group".to_string()));

        assert_eq!(model.member_count, Some(100));
        assert_eq!(model.online_count, Some(25));
        assert_eq!(model.description.as_ref().unwrap(), "A test group");
    }

    #[test]
    fn test_clear() {
        let mut model = SidebarModel::new();
        let chat = create_test_chat(1, "Test", ChatType::Private);
        let user = create_test_user(1, "Test");

        model.set_chat(chat, Some(user));
        model.set_group_info(10, Some(5), Some("Description".to_string()));

        model.clear();

        assert!(model.chat.is_none());
        assert!(model.user.is_none());
        assert!(model.member_count.is_none());
        assert!(model.online_count.is_none());
        assert!(model.description.is_none());
        assert!(!model.has_chat());
    }

    #[test]
    fn test_set_chat_resets_group_info() {
        let mut model = SidebarModel::new();
        let chat1 = create_test_chat(1, "Group 1", ChatType::Supergroup);

        model.set_chat(chat1, None);
        model.set_group_info(100, Some(25), Some("Description".to_string()));

        // Set a new chat
        let chat2 = create_test_chat(2, "Group 2", ChatType::Supergroup);
        model.set_chat(chat2, None);

        // Group info should be reset
        assert!(model.member_count.is_none());
        assert!(model.online_count.is_none());
        assert!(model.description.is_none());
    }

    #[test]
    fn test_widget_focused() {
        let model = SidebarModel::new();
        let widget = SidebarWidget::new(&model);

        assert!(!widget.is_focused);

        let widget = widget.focused(true);
        assert!(widget.is_focused);
    }

    #[test]
    fn test_widget_no_chat_shows_placeholder() {
        let model = SidebarModel::new();
        let widget = SidebarWidget::new(&model);

        let lines = widget.build_content_lines();

        assert_eq!(lines.len(), 1);
        // The line should contain the placeholder text
    }

    #[test]
    fn test_widget_with_private_chat() {
        let mut model = SidebarModel::new();
        let mut chat = create_test_chat(1, "John Doe", ChatType::Private);
        chat.username = "johndoe".to_string();
        let mut user = create_test_user(1, "John");
        user.phone_number = "+1234567890".to_string();
        user.is_verified = true;
        user.is_premium = true;

        model.set_chat(chat, Some(user));

        let widget = SidebarWidget::new(&model);
        let lines = widget.build_content_lines();

        // Should have title, type, username, phone, status, badges, settings section
        assert!(lines.len() >= 5);
    }

    #[test]
    fn test_widget_with_group() {
        let mut model = SidebarModel::new();
        let mut chat = create_test_chat(1, "Test Group", ChatType::Supergroup);
        chat.is_pinned = true;
        chat.unread_count = 5;

        model.set_chat(chat, None);
        model.set_group_info(150, Some(30), Some("Group description".to_string()));

        let widget = SidebarWidget::new(&model);
        let lines = widget.build_content_lines();

        // Should include member count, online count, description, pinned indicator, unread count
        assert!(lines.len() >= 8);
    }

    #[test]
    fn test_widget_with_channel() {
        let mut model = SidebarModel::new();
        let chat = create_test_chat(1, "News Channel", ChatType::Channel);

        model.set_chat(chat, None);
        model.set_group_info(10000, None, None);

        let widget = SidebarWidget::new(&model);
        let lines = widget.build_content_lines();

        // Should show "Subscribers" label instead of "Members"
        assert!(lines.len() >= 4);
    }

    #[test]
    fn test_widget_muted_chat() {
        let mut model = SidebarModel::new();
        let mut chat = create_test_chat(1, "Muted Group", ChatType::Group);
        chat.is_muted = true;

        model.set_chat(chat, None);

        let widget = SidebarWidget::new(&model);
        let lines = widget.build_content_lines();

        // Should include muted indicator in the settings section
        assert!(lines.len() >= 5);
    }
}
