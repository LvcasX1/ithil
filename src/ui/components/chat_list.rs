//! Chat list model for managing and displaying the list of chats.
//!
//! This module provides the [`ChatListModel`] which manages the chat list pane,
//! including selection, scrolling, search/filtering, and rendering.
//!
//! # Ratatui Standards
//!
//! This component follows modern Ratatui patterns:
//! - Uses the standard [`List`] widget with [`ListState`] for selection
//! - Leverages [`ListItem`] created by [`ChatItemBuilder`] for consistent styling
//! - Applies highlight styles via the `List` widget's built-in methods

use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use ratatui::{
    layout::Rect,
    style::{Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, HighlightSpacing, List, ListState, Paragraph},
};

use crate::cache::SharedCache;
use crate::types::Chat;
use crate::ui::styles::{colors, Styles};

use super::chat_item::ChatItemBuilder;

/// Actions that can result from chat list input handling.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ChatListAction {
    /// A chat was selected and should be opened
    OpenChat(i64),
    /// No action needed
    None,
}

/// The chat list model managing selection, search, and display.
///
/// # Features
///
/// - Navigation with up/down keys (or j/k in vim mode)
/// - Fast navigation with Ctrl+U/Ctrl+D (half page)
/// - Jump to top/bottom with g/G or Home/End
/// - Search mode activated with `/`
/// - Quick jump to chats 1-9 with number keys
///
/// # Example
///
/// ```rust,no_run
/// use ithil::cache::new_shared_cache;
/// use ithil::ui::components::ChatListModel;
///
/// let cache = new_shared_cache(100);
/// let mut model = ChatListModel::new(cache);
///
/// // Add some chats...
/// model.refresh_from_cache();
/// ```
#[derive(Debug)]
pub struct ChatListModel {
    /// Cache for fetching chat data
    cache: SharedCache,
    /// List of chats (sorted by recency)
    chats: Vec<Chat>,
    /// Ratatui `ListState` for selection and scrolling
    list_state: ListState,
    /// Available width
    width: u16,
    /// Available height
    height: u16,
    /// Whether this pane has focus
    focused: bool,
    /// Search mode active
    search_mode: bool,
    /// Current search query
    search_query: String,
    /// Filtered chats (when in search mode)
    filtered_chats: Vec<Chat>,
}

impl ChatListModel {
    /// Creates a new chat list model.
    #[must_use]
    pub fn new(cache: SharedCache) -> Self {
        Self {
            cache,
            chats: Vec::new(),
            list_state: ListState::default(),
            width: 0,
            height: 0,
            focused: true,
            search_mode: false,
            search_query: String::new(),
            filtered_chats: Vec::new(),
        }
    }

    /// Sets the size of the chat list pane.
    pub fn set_size(&mut self, width: u16, height: u16) {
        self.width = width;
        self.height = height;
    }

    /// Sets the focus state.
    pub fn set_focused(&mut self, focused: bool) {
        self.focused = focused;
    }

    /// Returns whether the chat list is focused.
    #[must_use]
    pub const fn is_focused(&self) -> bool {
        self.focused
    }

    /// Refreshes the chat list from the cache.
    pub fn refresh_from_cache(&mut self) {
        let chats = self.cache.get_all_chats();
        self.set_chats(chats);
    }

    /// Sets the list of chats and sorts them.
    pub fn set_chats(&mut self, mut chats: Vec<Chat>) {
        // Remember selected chat ID
        let selected_chat_id = self.get_selected_chat().map(|c| c.id);

        // Sort chats by recency (pinned first, then by last message date)
        Self::sort_chats(&mut chats);
        self.chats = chats;

        // Try to maintain selection on the same chat
        if let Some(chat_id) = selected_chat_id {
            if let Some(new_idx) = self.chats.iter().position(|c| c.id == chat_id) {
                self.list_state.select(Some(new_idx));
            } else {
                // Chat not found, select first if available
                self.select_first_if_available();
            }
        } else {
            self.select_first_if_available();
        }
    }

    /// Selects the first chat if the list is not empty.
    fn select_first_if_available(&mut self) {
        if self.get_active_chats().is_empty() {
            self.list_state.select(None);
        } else {
            self.list_state.select(Some(0));
        }
    }

    /// Sorts chats: pinned first (by pin order), then by last message date.
    fn sort_chats(chats: &mut [Chat]) {
        chats.sort_by(|a, b| {
            // Pinned chats first
            match (a.is_pinned, b.is_pinned) {
                (true, false) => std::cmp::Ordering::Less,
                (false, true) => std::cmp::Ordering::Greater,
                (true, true) => a.pin_order.cmp(&b.pin_order),
                (false, false) => {
                    // Sort by last message date (most recent first)
                    let a_date = a.last_message.as_ref().map(|m| m.date);
                    let b_date = b.last_message.as_ref().map(|m| m.date);
                    b_date.cmp(&a_date)
                },
            }
        });
    }

    /// Updates a single chat in the list.
    pub fn update_chat(&mut self, chat: Chat) {
        if let Some(idx) = self.chats.iter().position(|c| c.id == chat.id) {
            self.chats[idx] = chat;
        } else {
            self.chats.push(chat);
        }
        Self::sort_chats(&mut self.chats);
    }

    /// Marks a chat as having a new message and moves it to top.
    pub fn mark_new_message(&mut self, chat_id: i64) {
        if let Some(chat) = self.chats.iter_mut().find(|c| c.id == chat_id) {
            chat.has_new_message = true;
        }
        Self::sort_chats(&mut self.chats);
    }

    /// Clears the new message flag for a chat.
    pub fn clear_new_message(&mut self, chat_id: i64) {
        if let Some(chat) = self.chats.iter_mut().find(|c| c.id == chat_id) {
            chat.has_new_message = false;
        }
    }

    /// Returns the currently selected chat.
    #[must_use]
    pub fn get_selected_chat(&self) -> Option<&Chat> {
        let chats = self.get_active_chats();
        self.list_state.selected().and_then(|i| chats.get(i))
    }

    /// Returns the selected chat ID.
    #[must_use]
    pub fn get_selected_chat_id(&self) -> Option<i64> {
        self.get_selected_chat().map(|c| c.id)
    }

    /// Returns the active chats (filtered if in search mode, all otherwise).
    fn get_active_chats(&self) -> &[Chat] {
        if self.search_mode && !self.filtered_chats.is_empty() {
            &self.filtered_chats
        } else if self.search_mode {
            &[]
        } else {
            &self.chats
        }
    }

    /// Returns the current selection index.
    fn selected_index(&self) -> usize {
        self.list_state.selected().unwrap_or(0)
    }

    /// Handles key input.
    ///
    /// Returns a [`ChatListAction`] indicating what action should be taken.
    pub fn handle_input(&mut self, key: KeyEvent) -> ChatListAction {
        if !self.focused {
            return ChatListAction::None;
        }

        if self.search_mode {
            return self.handle_search_input(key);
        }

        self.handle_normal_input(key)
    }

    /// Handles input in search mode.
    fn handle_search_input(&mut self, key: KeyEvent) -> ChatListAction {
        match key.code {
            KeyCode::Esc => {
                self.exit_search_mode();
                ChatListAction::None
            },
            KeyCode::Enter => {
                if let Some(chat) = self.get_selected_chat() {
                    let chat_id = chat.id;
                    self.exit_search_mode();
                    ChatListAction::OpenChat(chat_id)
                } else {
                    ChatListAction::None
                }
            },
            KeyCode::Backspace => {
                if !self.search_query.is_empty() {
                    self.search_query.pop();
                    self.filter_chats();
                }
                ChatListAction::None
            },
            KeyCode::Char(c) => {
                self.search_query.push(c);
                self.filter_chats();
                ChatListAction::None
            },
            KeyCode::Up => {
                self.move_up();
                ChatListAction::None
            },
            KeyCode::Down => {
                self.move_down();
                ChatListAction::None
            },
            _ => ChatListAction::None,
        }
    }

    /// Handles input in normal (non-search) mode.
    fn handle_normal_input(&mut self, key: KeyEvent) -> ChatListAction {
        match key.code {
            KeyCode::Up | KeyCode::Char('k') => {
                self.move_up();
                ChatListAction::None
            },
            KeyCode::Down | KeyCode::Char('j') => {
                self.move_down();
                ChatListAction::None
            },
            KeyCode::Char('u') if key.modifiers.contains(KeyModifiers::CONTROL) => {
                // Move up 5 items
                for _ in 0..5 {
                    if self.selected_index() == 0 {
                        break;
                    }
                    self.move_up();
                }
                ChatListAction::None
            },
            KeyCode::Char('d') if key.modifiers.contains(KeyModifiers::CONTROL) => {
                // Move down 5 items
                let max = self.get_active_chats().len().saturating_sub(1);
                for _ in 0..5 {
                    if self.selected_index() >= max {
                        break;
                    }
                    self.move_down();
                }
                ChatListAction::None
            },
            KeyCode::Home | KeyCode::Char('g') => {
                if !self.get_active_chats().is_empty() {
                    self.list_state.select(Some(0));
                }
                ChatListAction::None
            },
            KeyCode::End | KeyCode::Char('G') => {
                let chats = self.get_active_chats();
                if !chats.is_empty() {
                    self.list_state.select(Some(chats.len() - 1));
                }
                ChatListAction::None
            },
            KeyCode::Enter | KeyCode::Char('l') | KeyCode::Right => self.open_selected_chat(),
            KeyCode::Char('/') => {
                self.enter_search_mode();
                ChatListAction::None
            },
            KeyCode::Char(c @ '1'..='9') => {
                // Quick jump to chat by number
                let idx = (c as usize) - ('1' as usize);
                let chats = self.get_active_chats();
                if idx < chats.len() {
                    self.list_state.select(Some(idx));
                    self.open_selected_chat()
                } else {
                    ChatListAction::None
                }
            },
            _ => ChatListAction::None,
        }
    }

    /// Opens the currently selected chat.
    fn open_selected_chat(&self) -> ChatListAction {
        if let Some(chat) = self.get_selected_chat() {
            ChatListAction::OpenChat(chat.id)
        } else {
            ChatListAction::None
        }
    }

    /// Moves selection up.
    fn move_up(&mut self) {
        let selected = self.list_state.selected().unwrap_or(0);
        if selected > 0 {
            self.list_state.select(Some(selected - 1));
        }
    }

    /// Moves selection down.
    fn move_down(&mut self) {
        let chats = self.get_active_chats();
        let selected = self.list_state.selected().unwrap_or(0);
        if selected < chats.len().saturating_sub(1) {
            self.list_state.select(Some(selected + 1));
        }
    }

    /// Enters search mode.
    fn enter_search_mode(&mut self) {
        self.search_mode = true;
        self.search_query.clear();
        self.filtered_chats = self.chats.clone();
        self.list_state.select(Some(0));
    }

    /// Exits search mode.
    fn exit_search_mode(&mut self) {
        self.search_mode = false;
        self.search_query.clear();
        self.filtered_chats.clear();
        self.select_first_if_available();
    }

    /// Filters chats based on search query.
    fn filter_chats(&mut self) {
        if self.search_query.is_empty() {
            self.filtered_chats = self.chats.clone();
            self.list_state.select(Some(0));
            return;
        }

        let query = self.search_query.to_lowercase();
        self.filtered_chats = self
            .chats
            .iter()
            .filter(|chat| {
                // Search in title
                if chat.title.to_lowercase().contains(&query) {
                    return true;
                }
                // Search in username
                if !chat.username.is_empty() && chat.username.to_lowercase().contains(&query) {
                    return true;
                }
                // Search in last message text
                if let Some(ref msg) = chat.last_message {
                    if msg.content.text.to_lowercase().contains(&query) {
                        return true;
                    }
                }
                false
            })
            .cloned()
            .collect();

        self.list_state.select(Some(0));
    }

    /// Returns the number of chats.
    #[must_use]
    pub fn chat_count(&self) -> usize {
        self.get_active_chats().len()
    }

    /// Returns true if in search mode.
    #[must_use]
    pub const fn is_search_mode(&self) -> bool {
        self.search_mode
    }

    /// Renders the chat list.
    pub fn render(&mut self, frame: &mut ratatui::Frame<'_>, area: Rect) {
        // Update size
        self.width = area.width;
        self.height = area.height;

        // Build the title
        let title = self.build_title();

        // Determine border style based on focus
        let border_style = if self.focused {
            Style::default().fg(colors::accent_primary())
        } else {
            Style::default().fg(colors::fg_muted())
        };

        let block = Block::default()
            .title(title)
            .borders(Borders::ALL)
            .border_style(border_style);

        let inner_area = block.inner(area);

        // Get active chats
        let chats = self.get_active_chats();

        if chats.is_empty() {
            // Render empty state
            frame.render_widget(block, area);
            let empty_text = if self.search_mode {
                "No chats match your search"
            } else {
                "No chats yet"
            };
            let paragraph = Paragraph::new(empty_text)
                .style(Styles::text_muted())
                .alignment(ratatui::layout::Alignment::Center);

            // Center vertically
            let y_offset = inner_area.height / 2;
            if y_offset > 0 && inner_area.height > 1 {
                let centered_area =
                    Rect::new(inner_area.x, inner_area.y + y_offset, inner_area.width, 1);
                frame.render_widget(paragraph, centered_area);
            }
            return;
        }

        // Build list items using ChatItemBuilder
        let items: Vec<_> = chats
            .iter()
            .map(|chat| {
                ChatItemBuilder::new(chat, inner_area.width.saturating_sub(4))
                    .show_preview(true)
                    .build()
            })
            .collect();

        // Create the List widget with Ratatui's standard patterns
        let highlight_style = if self.focused {
            Style::default()
                .bg(colors::bg_interactive())
                .fg(colors::fg_bright())
                .add_modifier(Modifier::BOLD)
        } else {
            Style::default().bg(colors::bg_interactive()).fg(colors::fg_primary())
        };

        let list = List::new(items)
            .block(block)
            .highlight_symbol("▌ ")
            .highlight_style(highlight_style)
            .highlight_spacing(HighlightSpacing::Always)
            .repeat_highlight_symbol(true);

        // Render the list with state
        frame.render_stateful_widget(list, area, &mut self.list_state);
    }

    /// Builds the title for the chat list pane.
    ///
    /// Returns an owned `Line` to avoid lifetime conflicts with `list_state`.
    fn build_title(&self) -> Line<'static> {
        if self.search_mode {
            Line::from(vec![
                Span::styled(" 🔍 ", Styles::text_accent()),
                Span::styled(self.search_query.clone(), Styles::text_bright()),
                Span::styled("_", Styles::text_accent()),
                Span::raw(" "),
            ])
        } else {
            Line::from(vec![Span::styled(" Chats ", Styles::text_bright())])
        }
    }
}

// Keep ChatListState for potential external use, but it's no longer used internally
/// State for the chat list widget (legacy compatibility).
#[derive(Debug, Default)]
pub struct ChatListState {
    /// Current scroll offset (in items, not pixels)
    pub scroll_offset: usize,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::cache::new_shared_cache;
    use crate::types::{ChatType, Message, MessageContent};
    use chrono::Utc;

    fn create_test_model() -> ChatListModel {
        let cache = new_shared_cache(100);
        ChatListModel::new(cache)
    }

    fn create_test_chat(id: i64, title: &str) -> Chat {
        Chat {
            id,
            title: title.to_string(),
            chat_type: ChatType::Private,
            last_message: Some(Box::new(Message {
                id: 1,
                content: MessageContent {
                    text: "Test message".to_string(),
                    ..Default::default()
                },
                date: Utc::now(),
                ..Default::default()
            })),
            ..Default::default()
        }
    }

    #[test]
    fn test_new_model() {
        let model = create_test_model();
        assert_eq!(model.chat_count(), 0);
        assert!(!model.is_search_mode());
        assert!(model.is_focused());
    }

    #[test]
    fn test_set_chats() {
        let mut model = create_test_model();
        let chats = vec![create_test_chat(1, "Chat 1"), create_test_chat(2, "Chat 2")];
        model.set_chats(chats);
        assert_eq!(model.chat_count(), 2);
    }

    #[test]
    fn test_navigation() {
        let mut model = create_test_model();
        model.set_chats(vec![
            create_test_chat(1, "Chat 1"),
            create_test_chat(2, "Chat 2"),
            create_test_chat(3, "Chat 3"),
        ]);

        assert_eq!(model.selected_index(), 0);

        model.move_down();
        assert_eq!(model.selected_index(), 1);

        model.move_down();
        assert_eq!(model.selected_index(), 2);

        // Can't go past the end
        model.move_down();
        assert_eq!(model.selected_index(), 2);

        model.move_up();
        assert_eq!(model.selected_index(), 1);
    }

    #[test]
    fn test_search_mode() {
        let mut model = create_test_model();
        model.set_chats(vec![
            create_test_chat(1, "Alice"),
            create_test_chat(2, "Bob"),
            create_test_chat(3, "Charlie"),
        ]);

        model.enter_search_mode();
        assert!(model.is_search_mode());

        model.search_query = "ali".to_string();
        model.filter_chats();
        assert_eq!(model.chat_count(), 1);

        model.exit_search_mode();
        assert!(!model.is_search_mode());
        assert_eq!(model.chat_count(), 3);
    }

    #[test]
    fn test_pinned_sorting() {
        let mut model = create_test_model();
        let chat1 = create_test_chat(1, "Unpinned");
        let mut chat2 = create_test_chat(2, "Pinned 1");
        chat2.is_pinned = true;
        chat2.pin_order = 1;
        let mut chat3 = create_test_chat(3, "Pinned 2");
        chat3.is_pinned = true;
        chat3.pin_order = 2;

        model.set_chats(vec![chat1, chat2, chat3]);

        // Pinned chats should be first, in pin order
        let chats = model.get_active_chats();
        assert_eq!(chats[0].title, "Pinned 1");
        assert_eq!(chats[1].title, "Pinned 2");
        assert_eq!(chats[2].title, "Unpinned");
    }

    #[test]
    fn test_open_chat_action() {
        let mut model = create_test_model();
        model.set_chats(vec![create_test_chat(123, "Test")]);

        let action = model.open_selected_chat();
        assert_eq!(action, ChatListAction::OpenChat(123));
    }

    #[test]
    fn test_quick_jump() {
        let mut model = create_test_model();

        // Create chats with different timestamps to ensure deterministic sort order
        let now = Utc::now();
        let mut chat1 = create_test_chat(1, "Chat 1");
        let mut chat2 = create_test_chat(2, "Chat 2");
        let mut chat3 = create_test_chat(3, "Chat 3");

        // Set timestamps: chat1 most recent, then chat2, then chat3
        if let Some(ref mut msg) = chat1.last_message {
            msg.date = now;
        }
        if let Some(ref mut msg) = chat2.last_message {
            msg.date = now - chrono::Duration::seconds(1);
        }
        if let Some(ref mut msg) = chat3.last_message {
            msg.date = now - chrono::Duration::seconds(2);
        }

        model.set_chats(vec![chat1, chat2, chat3]);

        // Press '2' to jump to the chat at index 1 (second chat)
        let action = model.handle_input(KeyEvent::from(KeyCode::Char('2')));
        assert_eq!(action, ChatListAction::OpenChat(2));
    }

    #[test]
    fn test_mark_new_message() {
        let mut model = create_test_model();
        model.set_chats(vec![create_test_chat(1, "Test")]);

        model.mark_new_message(1);
        assert!(model.chats[0].has_new_message);

        model.clear_new_message(1);
        assert!(!model.chats[0].has_new_message);
    }

    #[test]
    fn test_list_state_integration() {
        let mut model = create_test_model();
        model.set_chats(vec![
            create_test_chat(1, "Chat 1"),
            create_test_chat(2, "Chat 2"),
        ]);

        // Should have first item selected by default
        assert_eq!(model.list_state.selected(), Some(0));

        model.move_down();
        assert_eq!(model.list_state.selected(), Some(1));

        model.move_up();
        assert_eq!(model.list_state.selected(), Some(0));
    }
}
