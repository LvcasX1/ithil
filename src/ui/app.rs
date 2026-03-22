//! Main TUI application logic.
//!
//! This module contains the core application struct and event loop
//! that drives the Ithil user interface. It manages the application state,
//! handles key events, processes Telegram updates, and renders the UI.
//!
//! # Architecture
//!
//! The application follows a state machine pattern with four main states:
//! - **Loading**: Initial connection to Telegram
//! - **Auth**: Authentication flow (phone, code, 2FA)
//! - **Main**: The main three-pane chat interface
//! - **Settings**: Configuration screen
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::App;
//! use ithil::app::Config;
//! use ithil::telegram::TelegramClient;
//! use ithil::cache::new_shared_cache;
//! use std::sync::Arc;
//!
//! # async fn example() -> Result<(), Box<dyn std::error::Error>> {
//! let config = Config::default();
//! let cache = new_shared_cache(100);
//! let telegram = Arc::new(TelegramClient::new(
//!     12345,
//!     "api_hash".to_string(),
//!     "session.session".to_string(),
//!     cache.clone(),
//! ));
//!
//! let mut app = App::new(config, telegram, cache);
//! // Run with a terminal...
//! # Ok(())
//! # }
//! ```

use std::sync::Arc;
use std::time::Duration;

use anyhow::Result;
use crossterm::event::{self, Event, KeyEvent, KeyEventKind};
use ratatui::{
    layout::{Alignment, Constraint, Direction, Layout, Rect},
    text::{Line, Span},
    widgets::{Block, Borders, Clear, Paragraph},
    Frame, Terminal,
};
use tokio::sync::mpsc;

use crate::app::Config;
use crate::cache::SharedCache;
use crate::telegram::TelegramClient;
use crate::types::{AuthState, Update, UpdateType};

use super::components::{
    AuthAction, AuthModel, ChatListAction, ChatListModel, ConversationAction, ConversationModel,
    ConversationWidget,
};
use super::keys::{Action, KeyMap};
use super::styles::Styles;

/// Which pane is currently focused in the main view.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum FocusedPane {
    /// Chat list pane (left)
    #[default]
    ChatList,
    /// Conversation pane (center)
    Conversation,
    /// Sidebar/info pane (right)
    Sidebar,
    /// Message input field
    Input,
}

impl std::fmt::Display for FocusedPane {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::ChatList => write!(f, "Chats"),
            Self::Conversation => write!(f, "Conversation"),
            Self::Sidebar => write!(f, "Info"),
            Self::Input => write!(f, "Input"),
        }
    }
}

/// Application state representing the current screen/mode.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum AppState {
    /// Initial loading state - connecting to Telegram
    #[default]
    Loading,
    /// Authentication flow
    Auth,
    /// Main chat interface
    Main,
    /// Settings screen
    Settings,
}

impl std::fmt::Display for AppState {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Loading => write!(f, "Loading"),
            Self::Auth => write!(f, "Authentication"),
            Self::Main => write!(f, "Main"),
            Self::Settings => write!(f, "Settings"),
        }
    }
}

/// Actions that flow out of the app for external handling.
#[derive(Debug, Clone)]
pub enum AppAction {
    /// Application should quit
    Quit,
    /// Action should be forwarded to the appropriate sub-view
    Forward(Action),
    /// Authentication action that requires async handling
    Auth(AuthAction),
    /// A chat was selected and should be opened
    ChatSelected(i64),
    /// Send a message to the current chat
    SendMessage(i64, String, Option<i64>),
    /// Edit an existing message
    EditMessage(i64, i64, String),
    /// Delete a message
    DeleteMessage(i64, i64),
    /// Open media (download if needed and open with system viewer)
    OpenMedia(i64, i64),
}

/// The main TUI application.
///
/// This struct holds all application state including configuration,
/// the Telegram client, cache, and UI state.
pub struct App {
    /// Current application state
    pub state: AppState,

    /// Which pane is currently focused
    pub focused_pane: FocusedPane,

    /// Whether the sidebar is visible
    pub show_sidebar: bool,

    /// Whether the help overlay is visible
    pub show_help: bool,

    /// Whether the application should quit
    pub should_quit: bool,

    /// Application configuration
    pub config: Config,

    /// Key bindings
    pub keymap: KeyMap,

    /// Telegram client
    pub telegram: Arc<TelegramClient>,

    /// Shared cache for Telegram data
    pub cache: SharedCache,

    /// Channel receiver for Telegram updates
    pub update_rx: Option<mpsc::Receiver<Update>>,

    /// Current authentication state
    pub auth_state: AuthState,

    /// Authentication UI model
    auth_model: AuthModel,

    /// Chat list UI model
    chat_list_model: ChatListModel,

    /// Conversation UI model
    conversation_model: ConversationModel,

    /// Currently selected chat ID (for conversation view)
    selected_chat_id: Option<i64>,

    /// Status message to display (for errors/info)
    pub status_message: Option<String>,
}

impl App {
    /// Create a new application instance.
    ///
    /// # Arguments
    ///
    /// * `config` - Application configuration
    /// * `telegram` - Telegram client (wrapped in Arc for sharing)
    /// * `cache` - Shared cache for Telegram data
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// use ithil::ui::App;
    /// use ithil::app::Config;
    /// use ithil::telegram::TelegramClient;
    /// use ithil::cache::new_shared_cache;
    /// use std::sync::Arc;
    ///
    /// let config = Config::default();
    /// let cache = new_shared_cache(100);
    /// let telegram = Arc::new(TelegramClient::new(
    ///     12345,
    ///     "api_hash".to_string(),
    ///     "session.session".to_string(),
    ///     cache.clone(),
    /// ));
    ///
    /// let app = App::new(config, telegram, cache);
    /// ```
    #[must_use]
    pub fn new(config: Config, telegram: Arc<TelegramClient>, cache: SharedCache) -> Self {
        let vim_mode = config.ui.keyboard.vim_mode;
        let show_sidebar = config.ui.layout.show_info_pane;
        let chat_list_model = ChatListModel::new(cache.clone());
        let conversation_model = ConversationModel::new();

        Self {
            state: AppState::Loading,
            focused_pane: FocusedPane::ChatList,
            show_sidebar,
            show_help: false,
            should_quit: false,
            config,
            keymap: KeyMap::new(vim_mode),
            telegram,
            cache,
            update_rx: None,
            auth_state: AuthState::WaitPhoneNumber,
            auth_model: AuthModel::new(),
            chat_list_model,
            conversation_model,
            selected_chat_id: None,
            status_message: None,
        }
    }

    /// Set the update receiver channel.
    ///
    /// Telegram updates will be received through this channel.
    pub fn set_update_receiver(&mut self, rx: mpsc::Receiver<Update>) {
        self.update_rx = Some(rx);
    }

    /// Set a status message to display.
    pub fn set_status_message(&mut self, message: impl Into<String>) {
        self.status_message = Some(message.into());
    }

    /// Clear the status message.
    pub fn clear_status_message(&mut self) {
        self.status_message = None;
    }

    /// Updates the authentication state in both App and `AuthModel`.
    ///
    /// Call this when the Telegram client's auth state changes.
    pub fn update_auth_state(&mut self, state: AuthState) {
        self.auth_state = state;
        self.auth_model.set_auth_state(state);

        // Transition app state based on auth state
        match state {
            AuthState::Ready => {
                self.state = AppState::Main;
            },
            AuthState::Closed => {
                self.should_quit = true;
            },
            _ => {
                if self.state != AppState::Auth {
                    self.state = AppState::Auth;
                }
            },
        }
    }

    /// Sets an error message in the auth model.
    pub fn set_auth_error(&mut self, message: impl Into<String>) {
        self.auth_model.set_error(message);
    }

    /// Sets the auth model loading state.
    pub fn set_auth_loading(&mut self, loading: bool) {
        self.auth_model.set_loading(loading);
    }

    /// Refreshes the chat list from the cache.
    ///
    /// Call this after loading dialogs from Telegram.
    pub fn refresh_chat_list(&mut self) {
        self.chat_list_model.refresh_from_cache();
    }

    /// Updates a single chat in the chat list.
    pub fn update_chat(&mut self, chat: crate::types::Chat) {
        self.chat_list_model.update_chat(chat);
    }

    /// Returns the currently selected chat ID.
    #[must_use]
    pub fn get_selected_chat_id(&self) -> Option<i64> {
        self.selected_chat_id
    }

    /// Sets the size of the chat list pane.
    ///
    /// This should be called when the terminal is resized.
    pub fn set_chat_list_size(&mut self, width: u16, height: u16) {
        self.chat_list_model.set_size(width, height);
    }

    /// Run the main application loop.
    ///
    /// This method will block until the application is closed.
    ///
    /// # Errors
    ///
    /// Returns an error if terminal rendering or event handling fails.
    pub fn run<B: ratatui::backend::Backend>(&mut self, terminal: &mut Terminal<B>) -> Result<()> {
        let tick_rate = Duration::from_millis(100);

        loop {
            // Render the UI
            terminal.draw(|frame| self.render(frame))?;

            // Handle events
            if event::poll(tick_rate)? {
                if let Event::Key(key) = event::read()? {
                    // Only handle key press events, not release
                    if key.kind == KeyEventKind::Press {
                        self.handle_key(key);
                    }
                }
            }

            // Process any pending Telegram updates (sync version, no mark-as-read)
            self.process_updates_sync();

            // Check if we should quit
            if self.should_quit {
                break;
            }
        }

        Ok(())
    }

    /// Run the main application loop with async support.
    ///
    /// This is the async version of [`run`](Self::run) that properly handles
    /// async operations like Telegram authentication and API calls.
    ///
    /// # Errors
    ///
    /// Returns an error if terminal rendering or event handling fails.
    pub async fn run_async<B: ratatui::backend::Backend>(
        &mut self,
        terminal: &mut Terminal<B>,
    ) -> Result<()> {
        let tick_rate = Duration::from_millis(100);

        loop {
            // Render the UI
            terminal.draw(|frame| self.render(frame))?;

            // Handle events (poll is non-blocking with timeout)
            if event::poll(tick_rate)? {
                if let Event::Key(key) = event::read()? {
                    // Only handle key press events, not release
                    if key.kind == KeyEventKind::Press {
                        if let Some(action) = self.handle_key(key) {
                            self.handle_app_action(action).await;
                        }
                    }
                }
            }

            // Process any pending Telegram updates
            self.process_updates().await;

            // Check if we should quit
            if self.should_quit {
                break;
            }
        }

        Ok(())
    }

    /// Run the application loop with a background connection task.
    ///
    /// This method runs the UI event loop while a Telegram connection is being
    /// established in the background. Once the connection completes, it checks
    /// the auth state and transitions accordingly.
    ///
    /// # Arguments
    ///
    /// * `terminal` - The terminal to render to
    /// * `connect_handle` - A `JoinHandle` for the background connection task
    ///
    /// # Errors
    ///
    /// Returns an error if terminal rendering fails or the connection fails.
    pub async fn run_async_with_connection<B: ratatui::backend::Backend>(
        &mut self,
        terminal: &mut Terminal<B>,
        connect_handle: tokio::task::JoinHandle<Result<(), crate::telegram::TelegramError>>,
    ) -> Result<()> {
        use tokio::time::{interval, Duration as TokioDuration};
        use tracing::{error, info};

        let tick_rate = TokioDuration::from_millis(50);
        let mut tick_interval = interval(tick_rate);
        let mut connection_complete = false;

        // Pin the connect_handle so we can poll it
        tokio::pin!(connect_handle);

        loop {
            // Render the UI
            terminal.draw(|frame| self.render(frame))?;

            // Use tokio::select to handle multiple async sources
            tokio::select! {
                // Poll for terminal events (with short timeout to stay responsive)
                _ = tick_interval.tick() => {
                    // Check for terminal events (non-blocking)
                    while event::poll(Duration::from_millis(0))? {
                        if let Event::Key(key) = event::read()? {
                            if key.kind == KeyEventKind::Press {
                                if let Some(action) = self.handle_key(key) {
                                    self.handle_app_action(action).await;
                                }
                            }
                        }
                    }

                    // Process any pending Telegram updates
                    self.process_updates().await;
                }

                // Poll the connection handle (only if not already complete)
                result = &mut connect_handle, if !connection_complete => {
                    connection_complete = true;

                    match result {
                        Ok(Ok(())) => {
                            info!("Connection established, checking auth state");

                            // Check auth state and transition
                            let auth_state = self.telegram.get_auth_state().await;
                            self.update_auth_state(auth_state);

                            if auth_state == AuthState::Ready {
                                info!("Already authorized, loading initial data");
                                self.on_authorized().await;
                            } else {
                                info!("Auth required, state: {:?}", auth_state);
                            }
                        }
                        Ok(Err(e)) => {
                            error!("Connection failed: {e}");
                            self.set_status_message(format!("Connection failed: {e}"));
                            // Stay in loading state but show error
                            // User can quit with 'q' or Ctrl+C
                        }
                        Err(e) => {
                            error!("Connection task panicked: {e}");
                            self.set_status_message(format!("Connection error: {e}"));
                        }
                    }
                }
            }

            // Check if we should quit
            if self.should_quit {
                break;
            }
        }

        Ok(())
    }

    /// Handle app actions that may require async operations.
    async fn handle_app_action(&mut self, action: AppAction) {
        match action {
            AppAction::Auth(auth_action) => {
                self.handle_auth_action(auth_action).await;
            },
            AppAction::ChatSelected(chat_id) => {
                self.handle_chat_selected(chat_id).await;
            },
            AppAction::SendMessage(chat_id, text, reply_to) => {
                self.handle_send_message(chat_id, text, reply_to).await;
            },
            AppAction::EditMessage(chat_id, message_id, text) => {
                self.handle_edit_message(chat_id, message_id, text).await;
            },
            AppAction::DeleteMessage(chat_id, message_id) => {
                self.handle_delete_message(chat_id, message_id).await;
            },
            AppAction::OpenMedia(chat_id, message_id) => {
                self.handle_open_media(chat_id, message_id).await;
            },
            // Quit and Forward are already handled by setting should_quit in handle_key
            AppAction::Quit | AppAction::Forward(_) => {},
        }
    }

    /// Converts a conversation action to an app action.
    fn handle_conversation_action(&self, action: ConversationAction) -> Option<AppAction> {
        let chat_id = self.selected_chat_id?;

        match action {
            ConversationAction::SendMessage(text, reply_to) => {
                Some(AppAction::SendMessage(chat_id, text, reply_to))
            },
            ConversationAction::EditMessage(message_id, text) => {
                Some(AppAction::EditMessage(chat_id, message_id, text))
            },
            ConversationAction::DeleteMessage(message_id) => {
                Some(AppAction::DeleteMessage(chat_id, message_id))
            },
            ConversationAction::ForwardMessage(_message_id) => {
                // Forward not yet implemented
                None
            },
        }
    }

    /// Handle sending a message.
    async fn handle_send_message(&mut self, chat_id: i64, text: String, reply_to: Option<i64>) {
        match self.telegram.send_message(chat_id, &text, reply_to).await {
            Ok(message) => {
                // Add the sent message to the conversation
                self.conversation_model.add_message(message);
            },
            Err(e) => {
                self.set_status_message(format!("Failed to send message: {e}"));
            },
        }
    }

    /// Handle editing a message.
    async fn handle_edit_message(&mut self, chat_id: i64, message_id: i64, text: String) {
        match self.telegram.edit_message(chat_id, message_id, &text).await {
            Ok(message) => {
                self.conversation_model.update_message(message);
            },
            Err(e) => {
                self.set_status_message(format!("Failed to edit message: {e}"));
            },
        }
    }

    /// Handle deleting a message.
    async fn handle_delete_message(&mut self, chat_id: i64, message_id: i64) {
        // revoke=true means delete for everyone (if allowed)
        match self
            .telegram
            .delete_messages(chat_id, &[message_id], true)
            .await
        {
            Ok(()) => {
                self.conversation_model.delete_message(message_id);
            },
            Err(e) => {
                self.set_status_message(format!("Failed to delete message: {e}"));
            },
        }
    }

    /// Handle opening media from a message.
    ///
    /// Downloads the media if not already downloaded, then opens it with the system viewer.
    async fn handle_open_media(&mut self, chat_id: i64, message_id: i64) {
        use crate::telegram::TelegramClient;
        use crate::types::MessageType;

        // Get the message from cache
        let message = self
            .cache
            .get_messages(chat_id)
            .into_iter()
            .find(|m| m.id == message_id);

        let Some(message) = message else {
            self.set_status_message("Message not found".to_string());
            return;
        };

        // Check if it's a photo message
        if message.content.content_type != MessageType::Photo {
            self.set_status_message("Selected message is not a photo".to_string());
            return;
        }

        // Get the media directory from config (clone to avoid borrow issues)
        let media_dir = self.config.cache.media_directory.clone();

        // Download if needed and open
        self.set_status_message("Downloading photo...".to_string());

        match self
            .telegram
            .download_photo_if_needed(&message, &media_dir)
            .await
        {
            Ok(path) => {
                self.clear_status_message();
                // Open the file with system viewer
                if let Err(e) = TelegramClient::open_media_file(&path).await {
                    self.set_status_message(format!("Failed to open photo: {e}"));
                }
            },
            Err(e) => {
                self.set_status_message(format!("Failed to download photo: {e}"));
            },
        }
    }

    /// Handle authentication actions asynchronously.
    async fn handle_auth_action(&mut self, action: AuthAction) {
        self.set_auth_loading(true);
        self.clear_status_message();

        let result: Result<(), crate::telegram::TelegramError> = match action {
            AuthAction::SubmitPhoneNumber(phone) => self.telegram.request_login_code(&phone).await,
            AuthAction::SubmitCode(code) => self.telegram.sign_in(&code).await,
            AuthAction::SubmitPassword(password) => self.telegram.check_password(&password).await,
            AuthAction::SubmitRegistration(_) => {
                // Sign up via third-party apps is not supported by Telegram
                Err(crate::telegram::TelegramError::SignUpRequired)
            },
            AuthAction::Quit => {
                self.should_quit = true;
                return;
            },
        };

        self.set_auth_loading(false);

        match result {
            Ok(()) => {
                let new_state = self.telegram.get_auth_state().await;
                self.update_auth_state(new_state);

                // If we just became authorized, load initial data
                if new_state == AuthState::Ready {
                    self.on_authorized().await;
                }
            },
            Err(e) => {
                // Handle specific errors that change auth state
                match &e {
                    crate::telegram::TelegramError::PasswordRequired
                    | crate::telegram::TelegramError::SignUpRequired => {
                        let new_state = self.telegram.get_auth_state().await;
                        self.update_auth_state(new_state);
                    },
                    _ => {},
                }
                self.set_auth_error(e.to_string());
            },
        }
    }

    /// Called when the user becomes authorized.
    ///
    /// Loads initial data and prepares the main view.
    async fn on_authorized(&mut self) {
        // Load dialogs
        if let Err(e) = self.telegram.get_dialogs().await {
            self.set_status_message(format!("Failed to load chats: {e}"));
        } else {
            self.refresh_chat_list();
        }

        // Start the update loop if not already running
        if !self.telegram.is_update_loop_running() {
            let telegram = self.telegram.clone();
            tokio::spawn(async move {
                if let Err(e) = Box::pin(telegram.run_update_loop()).await {
                    tracing::error!("Update loop error: {e}");
                }
            });
        }
    }

    /// Handle chat selection - load messages for the selected chat.
    async fn handle_chat_selected(&mut self, chat_id: i64) {
        tracing::info!("Chat selected: {}", chat_id);

        // Get the chat from cache and set it on the conversation model
        if let Some(chat) = self.cache.get_chat(chat_id) {
            tracing::info!("Found chat in cache: {}", chat.title);
            self.conversation_model.set_chat(chat);
        } else {
            tracing::warn!("Chat {} not found in cache", chat_id);
        }

        // Load messages for the selected chat
        tracing::info!("Loading messages for chat {}", chat_id);
        match self.telegram.get_messages(chat_id, 50, None).await {
            Ok(messages) => {
                tracing::info!("Loaded {} messages for chat {}", messages.len(), chat_id);
                // Set messages on the conversation model
                self.conversation_model.set_messages(messages);
            },
            Err(e) => {
                tracing::error!("Failed to load messages for chat {}: {}", chat_id, e);
                self.set_status_message(format!("Failed to load messages: {e}"));
            },
        }

        // Mark chat as read
        if let Err(e) = self.telegram.mark_as_read(chat_id).await {
            tracing::warn!("Failed to mark chat {} as read: {}", chat_id, e);
        }
        self.refresh_chat_list();
    }

    /// Handle a key event.
    ///
    /// Returns an optional [`AppAction`] if the key triggered an action
    /// that needs external handling.
    pub fn handle_key(&mut self, key: KeyEvent) -> Option<AppAction> {
        // Handle auth state separately - forward all keys to AuthModel
        if self.state == AppState::Auth {
            if let Some(auth_action) = self.auth_model.handle_input(key) {
                return match auth_action {
                    AuthAction::Quit => {
                        self.should_quit = true;
                        Some(AppAction::Quit)
                    },
                    _ => Some(AppAction::Auth(auth_action)),
                };
            }
            return None;
        }

        // Handle chat list input when focused
        if self.state == AppState::Main && self.focused_pane == FocusedPane::ChatList {
            match self.chat_list_model.handle_input(key) {
                ChatListAction::OpenChat(chat_id) => {
                    self.selected_chat_id = Some(chat_id);
                    self.chat_list_model.clear_new_message(chat_id);
                    self.chat_list_model.set_focused(false);
                    self.focused_pane = FocusedPane::Conversation;
                    return Some(AppAction::ChatSelected(chat_id));
                },
                ChatListAction::None => {
                    // Key was handled by chat list (navigation, search, etc.)
                    // Check if it was a navigation key that was consumed
                    if self.chat_list_model.is_search_mode() {
                        return None;
                    }
                },
            }
        }

        // Handle conversation input when focused
        if self.state == AppState::Main && self.focused_pane == FocusedPane::Conversation {
            if let Some(action) = self.keymap.get_action(&key) {
                // Forward navigation and conversation-specific actions to conversation model
                match action {
                    Action::Up
                    | Action::Down
                    | Action::ScrollUp
                    | Action::ScrollDown
                    | Action::PageUp
                    | Action::PageDown
                    | Action::Home
                    | Action::End
                    | Action::Reply
                    | Action::Edit
                    | Action::Delete
                    | Action::Forward
                    | Action::CancelAction => {
                        let _ = self.conversation_model.handle_action(action);
                        return None;
                    },
                    Action::FocusInput | Action::OpenChat => {
                        // Focus the input - sync both the model and the pane
                        self.conversation_model.input.set_focused(true);
                        self.focused_pane = FocusedPane::Input;
                        return None;
                    },
                    Action::OpenMedia => {
                        // Get the selected message ID and open media
                        if let (Some(chat_id), Some(message)) = (
                            self.selected_chat_id,
                            self.conversation_model.selected_message(),
                        ) {
                            return Some(AppAction::OpenMedia(chat_id, message.id));
                        }
                        return None;
                    },
                    // Global actions should be handled by handle_action
                    _ => return self.handle_action(action),
                }
            }
            return None;
        }

        // Handle message input when focused
        if self.state == AppState::Main && self.focused_pane == FocusedPane::Input {
            // Check for special keys first
            if let Some(action) = self.keymap.get_action(&key) {
                match action {
                    // Enter key (OpenChat) sends message when in input mode
                    Action::SendMessage | Action::OpenChat => {
                        // Handle send message action
                        if let Some(conv_action) =
                            self.conversation_model.handle_action(Action::SendMessage)
                        {
                            return self.handle_conversation_action(conv_action);
                        }
                        return None;
                    },
                    Action::CancelAction => {
                        // Unfocus input and return to conversation
                        self.conversation_model.input.set_focused(false);
                        self.conversation_model.clear_action_state();
                        self.focused_pane = FocusedPane::Conversation;
                        return None;
                    },
                    Action::NewLine => {
                        self.conversation_model.input.insert_char('\n');
                        return None;
                    },
                    // Only Quit should work while typing (Ctrl+Q)
                    // Help (?) should be typed as a character
                    Action::Quit => {
                        return self.handle_action(action);
                    },
                    _ => {},
                }
            }

            // Forward raw key events to the input component
            self.conversation_model.input.handle_input(key);
            return None;
        }

        // Get action from keymap for other states
        if let Some(action) = self.keymap.get_action(&key) {
            return self.handle_action(action);
        }

        // Handle raw key codes for state-specific behavior
        match self.state {
            AppState::Settings => {
                if key.code == crossterm::event::KeyCode::Esc {
                    self.state = AppState::Main;
                }
            },
            AppState::Loading | AppState::Main | AppState::Auth => {},
        }

        None
    }

    /// Handle an action from the keymap.
    fn handle_action(&mut self, action: Action) -> Option<AppAction> {
        match action {
            Action::Quit => {
                self.should_quit = true;
                Some(AppAction::Quit)
            },
            Action::Help => {
                self.show_help = !self.show_help;
                None
            },
            Action::ToggleSidebar => {
                self.show_sidebar = !self.show_sidebar;
                // If we were focused on sidebar and it's now hidden, move focus
                if !self.show_sidebar && self.focused_pane == FocusedPane::Sidebar {
                    self.focused_pane = FocusedPane::Conversation;
                }
                None
            },
            Action::NextPane => {
                self.cycle_pane(1);
                None
            },
            Action::PreviousPane => {
                self.cycle_pane(-1);
                None
            },
            Action::FocusChatList => {
                self.focused_pane = FocusedPane::ChatList;
                self.chat_list_model.set_focused(true);
                None
            },
            Action::FocusConversation => {
                self.focused_pane = FocusedPane::Conversation;
                self.chat_list_model.set_focused(false);
                None
            },
            Action::FocusSidebar => {
                if self.show_sidebar {
                    self.focused_pane = FocusedPane::Sidebar;
                    self.chat_list_model.set_focused(false);
                }
                None
            },
            Action::FocusInput => {
                self.focused_pane = FocusedPane::Input;
                self.chat_list_model.set_focused(false);
                None
            },
            Action::OpenSettings => {
                self.state = AppState::Settings;
                None
            },
            Action::CancelAction => {
                match self.state {
                    AppState::Settings => self.state = AppState::Main,
                    AppState::Auth => {
                        self.should_quit = true;
                        return Some(AppAction::Quit);
                    },
                    _ => {
                        // Clear help overlay if visible
                        if self.show_help {
                            self.show_help = false;
                        }
                    },
                }
                None
            },
            // Forward other actions to sub-views
            _ => Some(AppAction::Forward(action)),
        }
    }

    /// Process pending Telegram updates (sync version, no mark-as-read).
    fn process_updates_sync(&mut self) {
        let updates: Vec<Update> = self.update_rx.as_mut().map_or_else(Vec::new, |rx| {
            let mut collected = Vec::new();
            while let Ok(update) = rx.try_recv() {
                collected.push(update);
            }
            collected
        });
        for update in updates {
            self.handle_update(update);
        }
    }

    /// Process pending Telegram updates from the channel.
    async fn process_updates(&mut self) {
        // Collect updates first to avoid borrowing issues
        let updates: Vec<Update> = self.update_rx.as_mut().map_or_else(Vec::new, |rx| {
            let mut collected = Vec::new();
            while let Ok(update) = rx.try_recv() {
                collected.push(update);
            }
            collected
        });

        // Track whether we received new messages for the active chat
        let mut should_mark_read = false;

        // Now process all collected updates
        for update in &updates {
            if update.update_type == UpdateType::NewMessage
                && self.selected_chat_id == Some(update.chat_id)
                && self.focused_pane != FocusedPane::ChatList
            {
                should_mark_read = true;
            }
        }

        for update in updates {
            self.handle_update(update);
        }

        // Mark active chat as read if we got new messages while viewing it
        if should_mark_read {
            if let Some(chat_id) = self.selected_chat_id {
                if let Err(e) = self.telegram.mark_as_read(chat_id).await {
                    tracing::warn!("Failed to mark chat {} as read: {}", chat_id, e);
                }
                self.refresh_chat_list();
            }
        }
    }

    /// Handle a single Telegram update.
    pub fn handle_update(&mut self, update: Update) {
        let is_selected_chat = self.selected_chat_id == Some(update.chat_id);

        match update.update_type {
            UpdateType::NewMessage => {
                if let Some(msg) = update.message {
                    let msg = *msg;
                    self.cache.add_message(update.chat_id, msg.clone());
                    // Update conversation view if this is the active chat
                    if is_selected_chat {
                        self.conversation_model.add_message(msg);
                    }
                    // Refresh chat list to update last message / order
                    self.refresh_chat_list();
                }
            },
            UpdateType::MessageEdited => {
                if let Some(msg) = update.message {
                    let msg = *msg;
                    self.cache.update_message(update.chat_id, msg.clone());
                    if is_selected_chat {
                        self.conversation_model.update_message(msg);
                    }
                }
            },
            UpdateType::MessageDeleted => {
                if let crate::types::UpdateData::Integer(msg_id) = update.data {
                    self.cache.delete_message(update.chat_id, msg_id);
                    if is_selected_chat {
                        self.conversation_model.delete_message(msg_id);
                    }
                }
            },
            UpdateType::NewChat => {
                if let crate::types::UpdateData::Chat(chat) = update.data {
                    self.cache.set_chat(*chat);
                    self.refresh_chat_list();
                }
            },
            UpdateType::UserStatus => {
                if let crate::types::UpdateData::User(user) = update.data {
                    self.cache.set_user(*user);
                }
            },
            _ => {
                // Other update types will be handled in future phases
            },
        }
    }

    /// Cycle focus between panes.
    #[allow(clippy::cast_possible_truncation, clippy::cast_possible_wrap)]
    fn cycle_pane(&mut self, direction: i32) {
        let panes = if self.show_sidebar {
            vec![
                FocusedPane::ChatList,
                FocusedPane::Conversation,
                FocusedPane::Sidebar,
            ]
        } else {
            vec![FocusedPane::ChatList, FocusedPane::Conversation]
        };

        if let Some(idx) = panes.iter().position(|&p| p == self.focused_pane) {
            let len = panes.len() as i32;
            let new_idx = (idx as i32 + direction).rem_euclid(len) as usize;
            self.focused_pane = panes[new_idx];
        } else {
            // If current pane not in list (e.g., Input), go to first pane
            self.focused_pane = panes[0];
        }

        // Update chat list focus state
        self.chat_list_model
            .set_focused(self.focused_pane == FocusedPane::ChatList);
    }

    /// Render the application.
    pub fn render(&mut self, frame: &mut Frame) {
        match self.state {
            AppState::Loading => self.render_loading(frame),
            AppState::Auth => self.render_auth(frame),
            AppState::Main => self.render_main(frame),
            AppState::Settings => self.render_settings(frame),
        }

        // Render help overlay if visible
        if self.show_help {
            self.render_help_overlay(frame);
        }
    }

    /// Render the loading screen.
    fn render_loading(&self, frame: &mut Frame) {
        let area = frame.area();
        let _ = self; // Suppress unused_self warning - will use self in future phases

        let block = Block::default()
            .borders(Borders::ALL)
            .border_style(Styles::border_focused())
            .title(" Ithil ")
            .title_alignment(Alignment::Center);

        let inner = block.inner(area);
        frame.render_widget(block, area);

        // Create centered content
        let chunks = Layout::vertical([
            Constraint::Percentage(40),
            Constraint::Length(3),
            Constraint::Length(1),
            Constraint::Length(1),
            Constraint::Min(0),
        ])
        .split(inner);

        // App name with styling
        let title = Paragraph::new(Line::from(vec![Span::styled("Ithil", Styles::highlight())]))
            .alignment(Alignment::Center);
        frame.render_widget(title, chunks[1]);

        // Loading message
        let loading = Paragraph::new(Line::from(vec![Span::styled(
            "Connecting to Telegram...",
            Styles::text_muted(),
        )]))
        .alignment(Alignment::Center);
        frame.render_widget(loading, chunks[2]);

        // Version info
        let version = Paragraph::new(Line::from(vec![Span::styled(
            format!("v{}", env!("CARGO_PKG_VERSION")),
            Styles::text_muted(),
        )]))
        .alignment(Alignment::Center);
        frame.render_widget(version, chunks[3]);
    }

    /// Render the authentication screen.
    fn render_auth(&self, frame: &mut Frame) {
        let area = frame.area();
        self.auth_model.render(frame, area);
    }

    /// Render the main chat interface.
    fn render_main(&mut self, frame: &mut Frame) {
        let area = frame.area();

        // Calculate layout based on config
        let constraints = self.calculate_layout_constraints();

        let chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints(constraints)
            .split(area);

        // Render chat list
        self.render_chat_list_pane(frame, chunks[0]);

        // Render conversation
        self.render_conversation_pane(frame, chunks[1]);

        // Render sidebar if visible
        if self.show_sidebar && chunks.len() > 2 {
            self.render_sidebar_pane(frame, chunks[2]);
        }
    }

    /// Calculate layout constraints based on configuration.
    fn calculate_layout_constraints(&self) -> Vec<Constraint> {
        let layout = &self.config.ui.layout;

        if self.show_sidebar {
            vec![
                Constraint::Percentage(u16::from(layout.chat_list_width)),
                Constraint::Percentage(u16::from(layout.conversation_width)),
                Constraint::Percentage(u16::from(layout.info_width)),
            ]
        } else {
            // Redistribute sidebar width to conversation
            let total = u16::from(layout.conversation_width) + u16::from(layout.info_width);
            vec![
                Constraint::Percentage(u16::from(layout.chat_list_width)),
                Constraint::Percentage(total),
            ]
        }
    }

    /// Render the chat list pane.
    fn render_chat_list_pane(&mut self, frame: &mut Frame, area: Rect) {
        self.chat_list_model.render(frame, area);
    }

    /// Render the conversation pane.
    fn render_conversation_pane(&self, frame: &mut Frame, area: Rect) {
        let is_focused = self.focused_pane == FocusedPane::Conversation
            || self.focused_pane == FocusedPane::Input;

        // Create a closure to look up sender names from the cache
        let cache = &self.cache;
        let get_sender_name = |user_id: i64| -> String {
            cache
                .get_user(user_id)
                .map(|u| u.get_display_name())
                .unwrap_or_else(|| format!("User {}", user_id))
        };

        let widget =
            ConversationWidget::new(&self.conversation_model, get_sender_name).focused(is_focused);

        frame.render_widget(widget, area);
    }

    /// Render the sidebar pane.
    fn render_sidebar_pane(&self, frame: &mut Frame, area: Rect) {
        let is_focused = self.focused_pane == FocusedPane::Sidebar;
        let border_style = if is_focused {
            Styles::border_focused()
        } else {
            Styles::border()
        };

        let block = Block::default()
            .title(" Info ")
            .borders(Borders::ALL)
            .border_style(border_style);

        let inner = block.inner(area);
        frame.render_widget(block, area);

        // Placeholder content - will be implemented in Phase 8
        let cache_stats = self.cache.stats();
        let content = Paragraph::new(vec![
            Line::from(Span::styled("Sidebar", Styles::text_bright())),
            Line::from(""),
            Line::from(Span::styled(
                format!("Cached users: {}", cache_stats.1),
                Styles::text_muted(),
            )),
            Line::from(""),
            Line::from(Span::styled("(Phase 8)", Styles::text_muted())),
        ])
        .alignment(Alignment::Center);

        frame.render_widget(content, inner);
    }

    /// Render the settings screen.
    fn render_settings(&self, frame: &mut Frame) {
        let area = frame.area();

        let block = Block::default()
            .borders(Borders::ALL)
            .border_style(Styles::info())
            .title(" Settings ")
            .title_alignment(Alignment::Center);

        let inner = block.inner(area);
        frame.render_widget(block, area);

        let vim_status = if self.config.ui.keyboard.vim_mode {
            "Enabled"
        } else {
            "Disabled"
        };

        let content = Paragraph::new(vec![
            Line::from(""),
            Line::from(Span::styled("Settings", Styles::highlight())),
            Line::from(""),
            Line::from(vec![
                Span::styled("Vim Mode: ", Styles::text()),
                Span::styled(vim_status, Styles::text_accent()),
            ]),
            Line::from(vec![
                Span::styled("Theme: ", Styles::text()),
                Span::styled(&self.config.ui.theme, Styles::text_accent()),
            ]),
            Line::from(vec![
                Span::styled("Show Sidebar: ", Styles::text()),
                Span::styled(
                    if self.show_sidebar { "Yes" } else { "No" },
                    Styles::text_accent(),
                ),
            ]),
            Line::from(""),
            Line::from(Span::styled(
                "(Full settings in Phase 8)",
                Styles::text_muted(),
            )),
            Line::from(""),
            Line::from(Span::styled("Press Esc to go back", Styles::text_muted())),
        ])
        .alignment(Alignment::Center);

        frame.render_widget(content, inner);
    }

    /// Render the help overlay.
    fn render_help_overlay(&self, frame: &mut Frame) {
        let area = frame.area();

        // Calculate centered help box dimensions
        let help_width = 50.min(area.width.saturating_sub(4));
        let help_height = 20.min(area.height.saturating_sub(4));
        let x = (area.width.saturating_sub(help_width)) / 2;
        let y = (area.height.saturating_sub(help_height)) / 2;

        let help_area = Rect::new(x, y, help_width, help_height);

        // Clear the background
        frame.render_widget(Clear, help_area);

        // Build help text
        let help_items = self.keymap.get_help_text();
        let mut lines: Vec<Line> = Vec::with_capacity(help_items.len() + 4);

        lines.push(Line::from(""));

        for (key, desc) in help_items {
            lines.push(Line::from(vec![
                Span::styled(format!("{key:12}"), Styles::text_accent()),
                Span::styled(desc, Styles::text()),
            ]));
        }

        lines.push(Line::from(""));
        lines.push(Line::from(Span::styled(
            "Press ? or Esc to close",
            Styles::text_muted(),
        )));

        let block = Block::default()
            .title(" Help ")
            .borders(Borders::ALL)
            .border_style(Styles::border_focused())
            .style(Styles::modal_background());

        let paragraph = Paragraph::new(lines).block(block);

        frame.render_widget(paragraph, help_area);
    }
}

impl std::fmt::Debug for App {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("App")
            .field("state", &self.state)
            .field("focused_pane", &self.focused_pane)
            .field("show_sidebar", &self.show_sidebar)
            .field("show_help", &self.show_help)
            .field("should_quit", &self.should_quit)
            .field("auth_state", &self.auth_state)
            .finish_non_exhaustive()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::cache::new_shared_cache;

    fn create_test_app() -> App {
        let config = Config::default();
        let cache = new_shared_cache(100);
        let telegram = Arc::new(TelegramClient::new(
            12345,
            "test_hash".to_string(),
            "test.session".to_string(),
            cache.clone(),
        ));
        App::new(config, telegram, cache)
    }

    #[test]
    fn test_new_app_default_state() {
        let app = create_test_app();
        assert_eq!(app.state, AppState::Loading);
        assert_eq!(app.focused_pane, FocusedPane::ChatList);
        assert!(!app.should_quit);
        assert!(!app.show_help);
    }

    #[test]
    fn test_cycle_pane_forward() {
        let mut app = create_test_app();
        app.show_sidebar = true;

        assert_eq!(app.focused_pane, FocusedPane::ChatList);

        app.cycle_pane(1);
        assert_eq!(app.focused_pane, FocusedPane::Conversation);

        app.cycle_pane(1);
        assert_eq!(app.focused_pane, FocusedPane::Sidebar);

        app.cycle_pane(1);
        assert_eq!(app.focused_pane, FocusedPane::ChatList);
    }

    #[test]
    fn test_cycle_pane_backward() {
        let mut app = create_test_app();
        app.show_sidebar = true;

        app.cycle_pane(-1);
        assert_eq!(app.focused_pane, FocusedPane::Sidebar);

        app.cycle_pane(-1);
        assert_eq!(app.focused_pane, FocusedPane::Conversation);
    }

    #[test]
    fn test_cycle_pane_without_sidebar() {
        let mut app = create_test_app();
        app.show_sidebar = false;

        assert_eq!(app.focused_pane, FocusedPane::ChatList);

        app.cycle_pane(1);
        assert_eq!(app.focused_pane, FocusedPane::Conversation);

        app.cycle_pane(1);
        assert_eq!(app.focused_pane, FocusedPane::ChatList);
    }

    #[test]
    fn test_toggle_help() {
        let mut app = create_test_app();
        assert!(!app.show_help);

        app.handle_action(Action::Help);
        assert!(app.show_help);

        app.handle_action(Action::Help);
        assert!(!app.show_help);
    }

    #[test]
    fn test_toggle_sidebar() {
        let mut app = create_test_app();
        let initial = app.show_sidebar;

        app.handle_action(Action::ToggleSidebar);
        assert_eq!(app.show_sidebar, !initial);

        app.handle_action(Action::ToggleSidebar);
        assert_eq!(app.show_sidebar, initial);
    }

    #[test]
    fn test_toggle_sidebar_moves_focus() {
        let mut app = create_test_app();
        app.show_sidebar = true;
        app.focused_pane = FocusedPane::Sidebar;

        app.handle_action(Action::ToggleSidebar);

        assert!(!app.show_sidebar);
        assert_eq!(app.focused_pane, FocusedPane::Conversation);
    }

    #[test]
    fn test_quit_action() {
        let mut app = create_test_app();
        assert!(!app.should_quit);

        let result = app.handle_action(Action::Quit);

        assert!(app.should_quit);
        assert!(matches!(result, Some(AppAction::Quit)));
    }

    #[test]
    fn test_open_settings() {
        let mut app = create_test_app();
        app.state = AppState::Main;

        app.handle_action(Action::OpenSettings);

        assert_eq!(app.state, AppState::Settings);
    }

    #[test]
    fn test_cancel_returns_from_settings() {
        let mut app = create_test_app();
        app.state = AppState::Settings;

        app.handle_action(Action::CancelAction);

        assert_eq!(app.state, AppState::Main);
    }

    #[test]
    fn test_focus_pane_actions() {
        let mut app = create_test_app();
        app.show_sidebar = true;

        app.handle_action(Action::FocusConversation);
        assert_eq!(app.focused_pane, FocusedPane::Conversation);

        app.handle_action(Action::FocusSidebar);
        assert_eq!(app.focused_pane, FocusedPane::Sidebar);

        app.handle_action(Action::FocusChatList);
        assert_eq!(app.focused_pane, FocusedPane::ChatList);

        app.handle_action(Action::FocusInput);
        assert_eq!(app.focused_pane, FocusedPane::Input);
    }

    #[test]
    fn test_focus_sidebar_ignored_when_hidden() {
        let mut app = create_test_app();
        app.show_sidebar = false;
        app.focused_pane = FocusedPane::ChatList;

        app.handle_action(Action::FocusSidebar);

        // Should not change focus when sidebar is hidden
        assert_eq!(app.focused_pane, FocusedPane::ChatList);
    }

    #[test]
    fn test_forward_action() {
        let mut app = create_test_app();

        let result = app.handle_action(Action::Up);

        assert!(matches!(result, Some(AppAction::Forward(Action::Up))));
    }

    #[test]
    fn test_status_message() {
        let mut app = create_test_app();
        assert!(app.status_message.is_none());

        app.set_status_message("Test message");
        assert_eq!(app.status_message, Some("Test message".to_string()));

        app.clear_status_message();
        assert!(app.status_message.is_none());
    }

    #[test]
    fn test_debug_impl() {
        let app = create_test_app();
        let debug = format!("{:?}", app);
        assert!(debug.contains("App"));
        assert!(debug.contains("state"));
    }
}
