//! Conversation view component.
//!
//! This module provides the conversation model and widget for displaying
//! messages in a selected chat, with support for:
//! - Message list with scrolling and selection
//! - Input area for composing messages
//! - Reply and edit modes
//! - Keyboard navigation
//!
//! # Architecture
//!
//! The conversation follows the Model-View-Update pattern:
//! - [`ConversationModel`]: Holds all state (messages, selection, input)
//! - [`ConversationWidget`]: Renders the model to the terminal
//! - [`ConversationAction`]: Actions that can be triggered from user input
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::types::Chat;
//! use ithil::ui::components::conversation::{ConversationModel, ConversationWidget};
//!
//! let mut model = ConversationModel::new();
//! model.set_chat(Chat::default());
//!
//! // In render function:
//! // let widget = ConversationWidget::new(&model, |id| "User".to_string())
//! //     .focused(true);
//! ```

use ratatui::{
    buffer::Buffer,
    layout::{Constraint, Direction, Layout, Rect},
    text::Span,
    widgets::{Block, Borders, Paragraph, Widget},
};

use crate::types::{Chat, Message};
use crate::ui::components::InputComponent;
use crate::ui::keys::Action;
use crate::ui::styles::Styles;

use super::message::MessageWidget;

/// Input mode for the conversation.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum InputMode {
    /// Normal message input
    #[default]
    Normal,
    /// Replying to a message
    Reply,
    /// Editing an existing message
    Edit,
}

/// Model for the conversation view.
///
/// This struct holds all state for the conversation pane, including
/// the current chat, messages, selection state, and input handling.
#[derive(Debug)]
pub struct ConversationModel {
    /// Currently selected chat
    pub chat: Option<Chat>,
    /// Messages in the current chat
    pub messages: Vec<Message>,
    /// Index of the selected message
    pub selected_index: usize,
    /// Scroll offset for the message list
    pub scroll_offset: usize,
    /// Input component for message composition
    pub input: InputComponent,
    /// Message ID being replied to
    pub reply_to: Option<i64>,
    /// Message ID being edited
    pub editing: Option<i64>,
    /// Current input mode
    pub input_mode: InputMode,
    /// Visible height of the message area (in lines)
    visible_height: usize,
}

impl Default for ConversationModel {
    fn default() -> Self {
        Self::new()
    }
}

impl ConversationModel {
    /// Creates a new conversation model.
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::conversation::ConversationModel;
    ///
    /// let model = ConversationModel::new();
    /// assert!(model.chat.is_none());
    /// assert!(model.messages.is_empty());
    /// ```
    #[must_use]
    pub fn new() -> Self {
        let mut input = InputComponent::new("Type a message...");
        input.set_focused(false); // Start with input unfocused; focus on message list
        Self {
            chat: None,
            messages: Vec::new(),
            selected_index: 0,
            scroll_offset: 0,
            input,
            reply_to: None,
            editing: None,
            input_mode: InputMode::Normal,
            visible_height: 20,
        }
    }

    /// Sets the current chat.
    ///
    /// This clears the message list and resets selection state.
    pub fn set_chat(&mut self, chat: Chat) {
        self.chat = Some(chat);
        self.messages.clear();
        self.selected_index = 0;
        self.scroll_offset = 0;
        self.clear_action_state();
    }

    /// Clears the current chat.
    pub fn clear_chat(&mut self) {
        self.chat = None;
        self.messages.clear();
        self.selected_index = 0;
        self.scroll_offset = 0;
        self.clear_action_state();
    }

    /// Sets the messages for the current chat.
    ///
    /// Messages from Telegram come in reverse chronological order (newest first),
    /// so we reverse them to display oldest at top, newest at bottom.
    /// The selection is anchored to the most recent (bottom) message.
    pub fn set_messages(&mut self, mut messages: Vec<Message>) {
        // Reverse so oldest is first, newest is last (at bottom)
        messages.reverse();
        self.messages = messages;
        // Select the most recent message (at the bottom) and scroll to show it
        if !self.messages.is_empty() {
            self.selected_index = self.messages.len() - 1;
            // Set scroll offset to show the bottom messages
            self.scroll_to_bottom();
        }
    }

    /// Adds a new message to the chat.
    ///
    /// If the user was viewing the latest message, auto-scrolls to the new one.
    pub fn add_message(&mut self, message: Message) {
        let was_at_bottom = self.selected_index == self.messages.len().saturating_sub(1);
        self.messages.push(message);

        // Auto-scroll to new message if at bottom
        if was_at_bottom && !self.messages.is_empty() {
            self.selected_index = self.messages.len() - 1;
            self.ensure_selected_visible();
        }
    }

    /// Updates an existing message.
    ///
    /// Finds the message by ID and replaces it.
    pub fn update_message(&mut self, message: Message) {
        if let Some(idx) = self.messages.iter().position(|m| m.id == message.id) {
            self.messages[idx] = message;
        }
    }

    /// Deletes a message from the chat.
    pub fn delete_message(&mut self, message_id: i64) {
        if let Some(idx) = self.messages.iter().position(|m| m.id == message_id) {
            self.messages.remove(idx);
            // Adjust selection if necessary
            if self.selected_index >= self.messages.len() && !self.messages.is_empty() {
                self.selected_index = self.messages.len() - 1;
            }
            if self.messages.is_empty() {
                self.selected_index = 0;
            }
        }
    }

    /// Handles an action from the key bindings.
    ///
    /// Returns a [`ConversationAction`] if the action triggers an external
    /// operation (like sending a message).
    pub fn handle_action(&mut self, action: Action) -> Option<ConversationAction> {
        // If input is focused, handle input actions
        if self.input.is_focused() {
            return self.handle_input_action(action);
        }

        match action {
            Action::Up | Action::ScrollUp => {
                self.select_previous();
                None
            },
            Action::Down | Action::ScrollDown => {
                self.select_next();
                None
            },
            Action::PageUp => {
                self.page_up();
                None
            },
            Action::PageDown => {
                self.page_down();
                None
            },
            Action::Home => {
                self.select_first();
                None
            },
            Action::End => {
                self.select_last();
                None
            },
            Action::FocusInput | Action::OpenChat => {
                self.input.set_focused(true);
                None
            },
            Action::Reply => {
                // Extract values before mutation to avoid borrow issues
                let msg_id = self.selected_message().map(|m| m.id);
                if let Some(id) = msg_id {
                    self.reply_to = Some(id);
                    self.input_mode = InputMode::Reply;
                    self.input.set_focused(true);
                    self.input.set_placeholder("Reply...");
                }
                None
            },
            Action::Edit => {
                // Extract needed values before mutation to avoid borrow issues
                let edit_info = self.selected_message().and_then(|msg| {
                    if msg.is_outgoing {
                        Some((msg.id, msg.content.text.clone()))
                    } else {
                        None
                    }
                });
                if let Some((id, text)) = edit_info {
                    self.editing = Some(id);
                    self.input_mode = InputMode::Edit;
                    self.input.set_value(&text);
                    self.input.set_focused(true);
                    self.input.set_placeholder("Edit message...");
                }
                None
            },
            Action::Delete => self
                .selected_message()
                .map(|msg| ConversationAction::DeleteMessage(msg.id)),
            Action::Forward => self
                .selected_message()
                .map(|msg| ConversationAction::ForwardMessage(msg.id)),
            Action::CancelAction => {
                self.clear_action_state();
                None
            },
            _ => None,
        }
    }

    /// Handles actions when the input is focused.
    fn handle_input_action(&mut self, action: Action) -> Option<ConversationAction> {
        match action {
            Action::SendMessage => self.submit_input(),
            Action::CancelAction => {
                self.input.set_focused(false);
                self.clear_action_state();
                None
            },
            Action::NewLine => {
                self.input.insert_char('\n');
                None
            },
            _ => None,
        }
    }

    /// Submits the current input.
    fn submit_input(&mut self) -> Option<ConversationAction> {
        let text = self.input.value().trim().to_string();
        if text.is_empty() {
            return None;
        }

        let action = if let Some(edit_id) = self.editing {
            ConversationAction::EditMessage(edit_id, text)
        } else {
            ConversationAction::SendMessage(text, self.reply_to)
        };

        self.input.clear();
        self.clear_action_state();
        Some(action)
    }

    /// Clears the reply/edit state.
    fn clear_action_state(&mut self) {
        self.reply_to = None;
        self.editing = None;
        self.input_mode = InputMode::Normal;
        self.input.set_placeholder("Type a message...");
    }

    /// Returns the currently selected message.
    #[must_use]
    pub fn selected_message(&self) -> Option<&Message> {
        self.messages.get(self.selected_index)
    }

    /// Returns true if there are no messages.
    #[must_use]
    pub fn is_empty(&self) -> bool {
        self.messages.is_empty()
    }

    /// Returns the number of messages.
    #[must_use]
    pub fn message_count(&self) -> usize {
        self.messages.len()
    }

    /// Selects the previous message.
    fn select_previous(&mut self) {
        if self.selected_index > 0 {
            self.selected_index -= 1;
            self.ensure_selected_visible();
        }
    }

    /// Selects the next message.
    fn select_next(&mut self) {
        if self.selected_index < self.messages.len().saturating_sub(1) {
            self.selected_index += 1;
            self.ensure_selected_visible();
        }
    }

    /// Selects the first message.
    fn select_first(&mut self) {
        self.selected_index = 0;
        self.ensure_selected_visible();
    }

    /// Selects the last message.
    fn select_last(&mut self) {
        if !self.messages.is_empty() {
            self.selected_index = self.messages.len() - 1;
            self.ensure_selected_visible();
        }
    }

    /// Scrolls up by a page.
    fn page_up(&mut self) {
        self.selected_index = self.selected_index.saturating_sub(self.visible_height);
        self.ensure_selected_visible();
    }

    /// Scrolls down by a page.
    fn page_down(&mut self) {
        self.selected_index =
            (self.selected_index + self.visible_height).min(self.messages.len().saturating_sub(1));
        self.ensure_selected_visible();
    }

    /// Ensures the selected message is visible.
    fn ensure_selected_visible(&mut self) {
        if self.selected_index < self.scroll_offset {
            self.scroll_offset = self.selected_index;
        } else if self.selected_index >= self.scroll_offset + self.visible_height {
            self.scroll_offset = self.selected_index.saturating_sub(self.visible_height) + 1;
        }
    }

    /// Scrolls to show the bottom (most recent) messages.
    fn scroll_to_bottom(&mut self) {
        if self.messages.len() > self.visible_height {
            self.scroll_offset = self.messages.len().saturating_sub(self.visible_height);
        } else {
            self.scroll_offset = 0;
        }
    }

    /// Sets the visible height for pagination.
    pub fn set_visible_height(&mut self, height: usize) {
        // Account for borders and input area
        self.visible_height = height.saturating_sub(5);
    }
}

/// Actions that can be triggered from the conversation.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ConversationAction {
    /// Send a new message (text, optional `reply_to` message ID)
    SendMessage(String, Option<i64>),
    /// Edit an existing message (`message_id`, `new_text`)
    EditMessage(i64, String),
    /// Delete a message
    DeleteMessage(i64),
    /// Forward a message
    ForwardMessage(i64),
}

/// Widget for rendering the conversation.
///
/// This widget renders the message list and input area.
pub struct ConversationWidget<'a, F>
where
    F: Fn(i64) -> String,
{
    /// Reference to the conversation model
    model: &'a ConversationModel,
    /// Whether this pane is focused
    is_focused: bool,
    /// Function to get sender name from user ID
    get_sender_name: F,
}

impl<'a, F> ConversationWidget<'a, F>
where
    F: Fn(i64) -> String,
{
    /// Creates a new conversation widget.
    ///
    /// # Arguments
    ///
    /// * `model` - Reference to the conversation model
    /// * `get_sender_name` - Function to resolve user IDs to display names
    #[must_use]
    #[allow(clippy::missing_const_for_fn)] // Can't be const due to closure parameter
    pub fn new(model: &'a ConversationModel, get_sender_name: F) -> Self {
        Self {
            model,
            is_focused: false,
            get_sender_name,
        }
    }

    /// Sets whether this pane is focused.
    #[must_use]
    pub const fn focused(mut self, focused: bool) -> Self {
        self.is_focused = focused;
        self
    }
}

impl<F> Widget for ConversationWidget<'_, F>
where
    F: Fn(i64) -> String,
{
    fn render(self, area: Rect, buf: &mut Buffer) {
        // Split into messages area and input area
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Min(3),    // Messages
                Constraint::Length(3), // Input
            ])
            .split(area);

        let messages_area = chunks[0];
        let input_area = chunks[1];

        // Render messages area
        let border_style = if self.is_focused && !self.model.input.is_focused() {
            Styles::border_focused()
        } else {
            Styles::border()
        };

        let title = self.model.chat.as_ref().map_or_else(
            || " No chat selected ".to_string(),
            |chat| format!(" {} ", chat.title),
        );

        let block = Block::default()
            .title(Span::styled(title, Styles::text_bright()))
            .borders(Borders::ALL)
            .border_style(border_style);

        let inner_area = block.inner(messages_area);
        block.render(messages_area, buf);

        if self.model.messages.is_empty() {
            let empty = Paragraph::new("No messages").style(Styles::text_muted());
            empty.render(inner_area, buf);
        } else {
            self.render_messages(inner_area, buf);
        }

        // Render input area
        self.render_input(input_area, buf);
    }
}

impl<F> ConversationWidget<'_, F>
where
    F: Fn(i64) -> String,
{
    /// Renders the message list.
    fn render_messages(&self, area: Rect, buf: &mut Buffer) {
        let mut y = area.y;
        let max_y = area.y + area.height;
        let message_spacing: u16 = 1; // Blank line between messages

        for (idx, msg) in self
            .model
            .messages
            .iter()
            .enumerate()
            .skip(self.model.scroll_offset)
        {
            if y >= max_y {
                break;
            }

            let sender_name = (self.get_sender_name)(msg.sender_id);
            let is_selected = idx == self.model.selected_index;

            let msg_widget = MessageWidget::new(msg, sender_name)
                .selected(is_selected)
                .width(area.width);

            let msg_height = msg_widget.height().min(max_y - y);
            let msg_area = Rect::new(area.x, y, area.width, msg_height);

            msg_widget.render(msg_area, buf);
            y += msg_height + message_spacing; // Add spacing after each message
        }
    }

    /// Renders the input area.
    fn render_input(&self, area: Rect, buf: &mut Buffer) {
        let input_border_style = if self.model.input.is_focused() {
            Styles::border_focused()
        } else {
            Styles::border()
        };

        let input_title = match self.model.input_mode {
            InputMode::Edit => " Edit message (Esc to cancel) ",
            InputMode::Reply => " Reply (Esc to cancel) ",
            InputMode::Normal => " Message ",
        };

        let input_block = Block::default()
            .title(Span::styled(input_title, Styles::text()))
            .borders(Borders::ALL)
            .border_style(input_border_style);

        let input_inner = input_block.inner(area);
        input_block.render(area, buf);

        // Render input text
        let (paragraph, _cursor_pos) = self.model.input.render_paragraph();
        paragraph.render(input_inner, buf);

        // Show cursor if focused
        if self.model.input.is_focused() {
            #[allow(clippy::cast_possible_truncation)]
            let cursor_x = input_inner.x + self.model.input.cursor() as u16;
            let cursor_y = input_inner.y;
            if cursor_x < input_inner.x + input_inner.width {
                buf[(cursor_x, cursor_y)].set_style(Styles::input_cursor());
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::{MessageContent, MessageType};
    use chrono::Utc;

    fn create_test_message(id: i64, text: &str, is_outgoing: bool) -> Message {
        Message {
            id,
            chat_id: 100,
            sender_id: if is_outgoing { 1 } else { 42 },
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

    fn create_test_chat(id: i64, title: &str) -> Chat {
        Chat {
            id,
            title: title.to_string(),
            ..Default::default()
        }
    }

    #[test]
    fn test_new_model() {
        let model = ConversationModel::new();

        assert!(model.chat.is_none());
        assert!(model.messages.is_empty());
        assert_eq!(model.selected_index, 0);
        assert_eq!(model.scroll_offset, 0);
        assert!(model.reply_to.is_none());
        assert!(model.editing.is_none());
    }

    #[test]
    fn test_set_chat() {
        let mut model = ConversationModel::new();
        model.add_message(create_test_message(1, "Test", false));
        model.reply_to = Some(1);

        model.set_chat(create_test_chat(100, "Test Chat"));

        assert!(model.chat.is_some());
        assert!(model.messages.is_empty()); // Messages cleared
        assert!(model.reply_to.is_none()); // Reply state cleared
    }

    #[test]
    fn test_set_messages() {
        let mut model = ConversationModel::new();
        let messages = vec![
            create_test_message(1, "First", false),
            create_test_message(2, "Second", false),
            create_test_message(3, "Third", false),
        ];

        model.set_messages(messages);

        assert_eq!(model.messages.len(), 3);
        assert_eq!(model.selected_index, 2); // Last message selected
    }

    #[test]
    fn test_add_message_auto_scroll() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![
            create_test_message(1, "First", false),
            create_test_message(2, "Second", false),
        ]);

        // Selection should be at last message
        assert_eq!(model.selected_index, 1);

        // Add new message - should auto-scroll
        model.add_message(create_test_message(3, "Third", false));

        assert_eq!(model.messages.len(), 3);
        assert_eq!(model.selected_index, 2);
    }

    #[test]
    fn test_update_message() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![create_test_message(1, "Original", false)]);

        let updated = create_test_message(1, "Updated", false);
        model.update_message(updated);

        assert_eq!(model.messages[0].content.text, "Updated");
    }

    #[test]
    fn test_delete_message() {
        let mut model = ConversationModel::new();
        // Messages come from Telegram in reverse chronological order (newest first)
        // After set_messages reverses them, order will be [1, 2, 3]
        model.set_messages(vec![
            create_test_message(3, "Third", false),
            create_test_message(2, "Second", false),
            create_test_message(1, "First", false),
        ]);

        model.delete_message(2);

        assert_eq!(model.messages.len(), 2);
        assert_eq!(model.messages[0].id, 1);
        assert_eq!(model.messages[1].id, 3);
    }

    #[test]
    fn test_delete_last_message_adjusts_selection() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![
            create_test_message(1, "First", false),
            create_test_message(2, "Second", false),
        ]);

        assert_eq!(model.selected_index, 1);

        model.delete_message(2);

        assert_eq!(model.selected_index, 0); // Adjusted to last valid index
    }

    #[test]
    fn test_navigation_up_down() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![
            create_test_message(1, "First", false),
            create_test_message(2, "Second", false),
            create_test_message(3, "Third", false),
        ]);

        assert_eq!(model.selected_index, 2);

        model.handle_action(Action::Up);
        assert_eq!(model.selected_index, 1);

        model.handle_action(Action::Up);
        assert_eq!(model.selected_index, 0);

        model.handle_action(Action::Up);
        assert_eq!(model.selected_index, 0); // Can't go below 0

        model.handle_action(Action::Down);
        assert_eq!(model.selected_index, 1);
    }

    #[test]
    fn test_home_end_navigation() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![
            create_test_message(1, "First", false),
            create_test_message(2, "Second", false),
            create_test_message(3, "Third", false),
        ]);

        model.handle_action(Action::Home);
        assert_eq!(model.selected_index, 0);

        model.handle_action(Action::End);
        assert_eq!(model.selected_index, 2);
    }

    #[test]
    fn test_reply_action() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![create_test_message(1, "Message", false)]);

        model.handle_action(Action::Reply);

        assert_eq!(model.reply_to, Some(1));
        assert_eq!(model.input_mode, InputMode::Reply);
        assert!(model.input.is_focused());
    }

    #[test]
    fn test_edit_action_outgoing_only() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![create_test_message(1, "Incoming", false)]);

        model.handle_action(Action::Edit);

        // Can't edit incoming messages
        assert!(model.editing.is_none());
        assert_eq!(model.input_mode, InputMode::Normal);
    }

    #[test]
    fn test_edit_action_outgoing() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![
            create_test_message(1, "My message", true), // Outgoing
        ]);

        model.handle_action(Action::Edit);

        assert_eq!(model.editing, Some(1));
        assert_eq!(model.input_mode, InputMode::Edit);
        assert_eq!(model.input.value(), "My message");
    }

    #[test]
    fn test_cancel_action_clears_state() {
        let mut model = ConversationModel::new();
        model.set_messages(vec![create_test_message(1, "Message", false)]);

        model.handle_action(Action::Reply);
        assert!(model.reply_to.is_some());

        model.handle_action(Action::CancelAction);

        assert!(model.reply_to.is_none());
        assert_eq!(model.input_mode, InputMode::Normal);
    }

    #[test]
    fn test_focus_input_action() {
        let mut model = ConversationModel::new();
        model.input.set_focused(false);

        model.handle_action(Action::FocusInput);

        assert!(model.input.is_focused());
    }

    #[test]
    fn test_selected_message() {
        let mut model = ConversationModel::new();
        // Messages come from Telegram in reverse chronological order (newest first)
        // After set_messages reverses them, order will be [1, 2] with indices 0, 1
        model.set_messages(vec![
            create_test_message(2, "Second", false),
            create_test_message(1, "First", false),
        ]);

        model.selected_index = 0;
        assert_eq!(model.selected_message().map(|m| m.id), Some(1));

        model.selected_index = 1;
        assert_eq!(model.selected_message().map(|m| m.id), Some(2));
    }

    #[test]
    fn test_is_empty() {
        let model = ConversationModel::new();
        assert!(model.is_empty());

        let mut model = ConversationModel::new();
        model.add_message(create_test_message(1, "Test", false));
        assert!(!model.is_empty());
    }

    #[test]
    fn test_message_count() {
        let mut model = ConversationModel::new();
        assert_eq!(model.message_count(), 0);

        model.set_messages(vec![
            create_test_message(1, "First", false),
            create_test_message(2, "Second", false),
        ]);
        assert_eq!(model.message_count(), 2);
    }

    #[test]
    fn test_conversation_action_variants() {
        let send = ConversationAction::SendMessage("Hello".to_string(), Some(42));
        let edit = ConversationAction::EditMessage(1, "Updated".to_string());
        let delete = ConversationAction::DeleteMessage(1);
        let forward = ConversationAction::ForwardMessage(1);

        // Just verify they can be created and matched
        assert!(matches!(send, ConversationAction::SendMessage(_, _)));
        assert!(matches!(edit, ConversationAction::EditMessage(_, _)));
        assert!(matches!(delete, ConversationAction::DeleteMessage(_)));
        assert!(matches!(forward, ConversationAction::ForwardMessage(_)));
    }

    #[test]
    fn test_set_visible_height() {
        let mut model = ConversationModel::new();
        model.set_visible_height(30);

        // visible_height should be 30 - 5 = 25 (accounting for borders/input)
        assert_eq!(model.visible_height, 25);
    }

    #[test]
    fn test_clear_chat() {
        let mut model = ConversationModel::new();
        model.set_chat(create_test_chat(1, "Test"));
        model.set_messages(vec![create_test_message(1, "Test", false)]);
        model.reply_to = Some(1);

        model.clear_chat();

        assert!(model.chat.is_none());
        assert!(model.messages.is_empty());
        assert!(model.reply_to.is_none());
    }
}
