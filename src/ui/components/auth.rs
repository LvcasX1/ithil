//! Authentication flow UI component.
//!
//! This component handles the entire authentication flow for Telegram:
//! 1. Phone number input
//! 2. Verification code input
//! 3. 2FA password input (if enabled)
//! 4. Registration (for new accounts - typically redirects to official app)
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::components::AuthModel;
//! use ithil::types::AuthState;
//!
//! let mut auth = AuthModel::new();
//! auth.set_auth_state(AuthState::WaitPhoneNumber);
//! // Handle input, render, etc.
//! ```

use crossterm::event::{KeyCode, KeyEvent};
use ratatui::{
    layout::{Alignment, Constraint, Layout, Rect},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::types::AuthState;
use crate::ui::styles::Styles;

use super::input::{EchoMode, InputComponent};

/// Actions that the auth model can produce.
///
/// These are returned to the caller for handling (e.g., making API calls).
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum AuthAction {
    /// User submitted phone number
    SubmitPhoneNumber(String),
    /// User submitted verification code
    SubmitCode(String),
    /// User submitted 2FA password
    SubmitPassword(String),
    /// User submitted registration info
    SubmitRegistration(String),
    /// User requested to quit
    Quit,
}

/// Authentication flow UI component.
///
/// This component manages the authentication state machine and renders
/// the appropriate UI for each authentication step.
#[derive(Debug, Clone)]
pub struct AuthModel {
    /// Current authentication state
    auth_state: AuthState,
    /// Text input field
    input: InputComponent,
    /// Error message to display (if any)
    error_message: Option<String>,
    /// Whether an operation is in progress
    loading: bool,
    /// Component dimensions
    width: u16,
    height: u16,
}

impl Default for AuthModel {
    fn default() -> Self {
        Self::new()
    }
}

impl AuthModel {
    /// Creates a new authentication model.
    ///
    /// Starts in the `WaitPhoneNumber` state.
    #[must_use]
    pub fn new() -> Self {
        let mut model = Self {
            auth_state: AuthState::WaitPhoneNumber,
            input: InputComponent::new("Phone number (e.g., +1234567890)").with_char_limit(20),
            error_message: None,
            loading: false,
            width: 80,
            height: 24,
        };
        model.input.set_focused(true);
        model
    }

    /// Gets the current authentication state.
    #[must_use]
    pub const fn auth_state(&self) -> AuthState {
        self.auth_state
    }

    /// Sets the authentication state and updates the input field accordingly.
    pub fn set_auth_state(&mut self, state: AuthState) {
        self.auth_state = state;
        self.loading = false;
        self.update_input_for_state();
    }

    /// Sets an error message to display.
    pub fn set_error(&mut self, message: impl Into<String>) {
        self.error_message = Some(message.into());
        self.loading = false;
    }

    /// Clears the error message.
    pub fn clear_error(&mut self) {
        self.error_message = None;
    }

    /// Sets the loading state.
    pub fn set_loading(&mut self, loading: bool) {
        self.loading = loading;
    }

    /// Returns `true` if currently loading.
    #[must_use]
    pub const fn is_loading(&self) -> bool {
        self.loading
    }

    /// Sets the component size.
    pub fn set_size(&mut self, width: u16, height: u16) {
        self.width = width;
        self.height = height;
    }

    /// Handles a key event.
    ///
    /// Returns `Some(AuthAction)` if the user performed an action
    /// that requires external handling (like submitting credentials).
    pub fn handle_input(&mut self, key: KeyEvent) -> Option<AuthAction> {
        // Don't process input while loading
        if self.loading {
            return None;
        }

        match key.code {
            KeyCode::Enter => {
                return self.handle_submit();
            },
            KeyCode::Esc => {
                return Some(AuthAction::Quit);
            },
            _ => {
                // Clear error on any input
                if self.error_message.is_some() {
                    self.clear_error();
                }
                // Forward to input component
                self.input.handle_input(key);
            },
        }

        None
    }

    /// Handles the submit action (Enter key).
    fn handle_submit(&mut self) -> Option<AuthAction> {
        let value = self.input.value().trim().to_string();

        if value.is_empty() {
            self.set_error("Please enter a value");
            return None;
        }

        self.loading = true;
        self.clear_error();

        match self.auth_state {
            AuthState::WaitPhoneNumber => Some(AuthAction::SubmitPhoneNumber(value)),
            AuthState::WaitCode => Some(AuthAction::SubmitCode(value)),
            AuthState::WaitPassword => Some(AuthAction::SubmitPassword(value)),
            AuthState::WaitRegistration => Some(AuthAction::SubmitRegistration(value)),
            AuthState::Ready | AuthState::Closed => None,
        }
    }

    /// Updates the input field configuration for the current state.
    fn update_input_for_state(&mut self) {
        self.input.clear();
        self.clear_error();

        match self.auth_state {
            AuthState::WaitPhoneNumber => {
                self.input
                    .set_placeholder("Phone number (e.g., +1234567890)");
                self.input.set_char_limit(20);
                self.input.set_echo_mode(EchoMode::Normal);
            },
            AuthState::WaitCode => {
                self.input.set_placeholder("Verification code");
                self.input.set_char_limit(10);
                self.input.set_echo_mode(EchoMode::Normal);
            },
            AuthState::WaitPassword => {
                self.input.set_placeholder("2FA password");
                self.input.set_char_limit(100);
                self.input.set_echo_mode(EchoMode::Password);
            },
            AuthState::WaitRegistration => {
                self.input.set_placeholder("First name");
                self.input.set_char_limit(50);
                self.input.set_echo_mode(EchoMode::Normal);
            },
            AuthState::Ready | AuthState::Closed => {
                // No input needed
            },
        }
    }

    /// Returns the prompt text for the current state.
    const fn get_prompt(&self) -> &'static str {
        match self.auth_state {
            AuthState::WaitPhoneNumber => "Please enter your phone number:",
            AuthState::WaitCode => "Please enter the verification code:",
            AuthState::WaitPassword => "Please enter your 2FA password:",
            AuthState::WaitRegistration => "Please enter your name:",
            AuthState::Ready => "Authentication successful!",
            AuthState::Closed => "Connection closed.",
        }
    }

    /// Returns the loading message for the current state.
    const fn get_loading_message(&self) -> &'static str {
        match self.auth_state {
            AuthState::WaitPhoneNumber => "Sending verification code...",
            AuthState::WaitCode => "Verifying code...",
            AuthState::WaitPassword => "Verifying password...",
            AuthState::WaitRegistration => "Registering account...",
            AuthState::Ready => "Authenticated!",
            AuthState::Closed => "Disconnected.",
        }
    }

    /// Returns the title for the current state.
    const fn get_title(&self) -> &'static str {
        match self.auth_state {
            AuthState::WaitPhoneNumber => "Sign In",
            AuthState::WaitCode => "Verification",
            AuthState::WaitPassword => "Two-Factor Authentication",
            AuthState::WaitRegistration => "Registration",
            AuthState::Ready => "Welcome",
            AuthState::Closed => "Disconnected",
        }
    }

    /// Renders the authentication screen.
    pub fn render(&self, frame: &mut Frame, area: Rect) {
        // Calculate centered content area
        let center_width = 50.min(area.width.saturating_sub(4));
        let center_height = 16.min(area.height.saturating_sub(2));
        let x = (area.width.saturating_sub(center_width)) / 2;
        let y = (area.height.saturating_sub(center_height)) / 2;

        let content_area = Rect::new(area.x + x, area.y + y, center_width, center_height);

        // Render outer border
        let outer_block = Block::default()
            .borders(Borders::ALL)
            .border_style(Styles::border_focused())
            .title(format!(" {} ", self.get_title()))
            .title_alignment(Alignment::Center);

        let inner_area = outer_block.inner(content_area);
        frame.render_widget(outer_block, content_area);

        // Split inner area into sections
        let chunks = Layout::vertical([
            Constraint::Length(2), // Title spacing
            Constraint::Length(2), // Prompt or loading
            Constraint::Length(1), // Spacing
            Constraint::Length(3), // Input field
            Constraint::Length(1), // Spacing
            Constraint::Length(2), // Error message
            Constraint::Min(0),    // Help text
        ])
        .split(inner_area);

        // Render app title
        let title = Paragraph::new(Line::from(vec![
            Span::styled("Welcome to ", Styles::text()),
            Span::styled("Ithil", Styles::highlight()),
        ]))
        .alignment(Alignment::Center);
        frame.render_widget(title, chunks[0]);

        if self.loading {
            // Show loading indicator
            self.render_loading(frame, chunks[1]);
        } else {
            // Show prompt
            let prompt =
                Paragraph::new(Line::from(Span::styled(self.get_prompt(), Styles::text())))
                    .alignment(Alignment::Center);
            frame.render_widget(prompt, chunks[1]);

            // Show input field (skip for Ready/Closed states)
            if self.auth_state != AuthState::Ready && self.auth_state != AuthState::Closed {
                self.render_input(frame, chunks[3]);
            }
        }

        // Show error message if present
        if let Some(ref error) = self.error_message {
            let error_para = Paragraph::new(Line::from(Span::styled(error, Styles::error())))
                .alignment(Alignment::Center);
            frame.render_widget(error_para, chunks[5]);
        }

        // Show help text
        let help_text = if self.loading {
            "Please wait..."
        } else {
            "Enter: Submit • Esc: Quit"
        };
        let help = Paragraph::new(Line::from(Span::styled(help_text, Styles::text_muted())))
            .alignment(Alignment::Center);
        frame.render_widget(help, chunks[6]);
    }

    /// Renders the loading indicator.
    fn render_loading(&self, frame: &mut Frame, area: Rect) {
        let loading_text = self.get_loading_message();

        let loading = Paragraph::new(Line::from(Span::styled(loading_text, Styles::info())))
            .alignment(Alignment::Center);
        frame.render_widget(loading, area);
    }

    /// Renders the input field.
    fn render_input(&self, frame: &mut Frame, area: Rect) {
        // Calculate centered input area
        let input_width = 35.min(area.width.saturating_sub(4));
        let x = (area.width.saturating_sub(input_width)) / 2;

        let _input_area = Rect::new(area.x + x, area.y, input_width, 1);

        // Create input border
        let input_block =
            Block::default()
                .borders(Borders::ALL)
                .border_style(if self.input.is_focused() {
                    Styles::border_focused()
                } else {
                    Styles::border()
                });

        // Calculate the inner area for the actual text
        let block_area = Rect::new(area.x + x, area.y, input_width, 3.min(area.height));
        let text_area = input_block.inner(block_area);

        frame.render_widget(input_block, block_area);

        // Render input text
        let (paragraph, cursor_pos) = self.input.render_paragraph();
        frame.render_widget(paragraph, text_area);

        // Set cursor position
        if let Some((cx, _cy)) = cursor_pos {
            let cursor_x = text_area.x + cx;
            let cursor_y = text_area.y;

            // Ensure cursor is within bounds
            if cursor_x < text_area.x + text_area.width {
                frame.set_cursor_position((cursor_x, cursor_y));
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_auth_model() {
        let model = AuthModel::new();
        assert_eq!(model.auth_state(), AuthState::WaitPhoneNumber);
        assert!(!model.is_loading());
        assert!(model.error_message.is_none());
    }

    #[test]
    fn test_set_auth_state() {
        let mut model = AuthModel::new();

        model.set_auth_state(AuthState::WaitCode);
        assert_eq!(model.auth_state(), AuthState::WaitCode);
        assert!(model.input.is_empty()); // Input should be cleared
    }

    #[test]
    fn test_set_error() {
        let mut model = AuthModel::new();

        model.set_error("Test error");
        assert_eq!(model.error_message, Some("Test error".to_string()));
        assert!(!model.is_loading()); // Loading should be cleared

        model.clear_error();
        assert!(model.error_message.is_none());
    }

    #[test]
    fn test_submit_empty_value() {
        let mut model = AuthModel::new();

        let action = model.handle_submit();
        assert!(action.is_none());
        assert!(model.error_message.is_some());
    }

    #[test]
    fn test_submit_phone_number() {
        let mut model = AuthModel::new();
        model.input.set_value("+1234567890");

        let action = model.handle_submit();
        assert_eq!(
            action,
            Some(AuthAction::SubmitPhoneNumber("+1234567890".to_string()))
        );
        assert!(model.is_loading());
    }

    #[test]
    fn test_submit_code() {
        let mut model = AuthModel::new();
        model.set_auth_state(AuthState::WaitCode);
        model.input.set_value("12345");

        let action = model.handle_submit();
        assert_eq!(action, Some(AuthAction::SubmitCode("12345".to_string())));
    }

    #[test]
    fn test_submit_password() {
        let mut model = AuthModel::new();
        model.set_auth_state(AuthState::WaitPassword);
        model.input.set_value("secret");

        let action = model.handle_submit();
        assert_eq!(
            action,
            Some(AuthAction::SubmitPassword("secret".to_string()))
        );
    }

    #[test]
    fn test_esc_quits() {
        let mut model = AuthModel::new();

        let key = KeyEvent::from(KeyCode::Esc);
        let action = model.handle_input(key);
        assert_eq!(action, Some(AuthAction::Quit));
    }

    #[test]
    fn test_loading_blocks_input() {
        let mut model = AuthModel::new();
        model.set_loading(true);

        let key = KeyEvent::from(KeyCode::Char('a'));
        let action = model.handle_input(key);
        assert!(action.is_none());
        assert!(model.input.is_empty()); // Input should not have changed
    }

    #[test]
    fn test_input_clears_error() {
        let mut model = AuthModel::new();
        model.set_error("Some error");

        let key = KeyEvent::from(KeyCode::Char('a'));
        model.handle_input(key);

        assert!(model.error_message.is_none());
    }

    #[test]
    fn test_prompts() {
        let model = AuthModel::new();
        assert_eq!(model.get_prompt(), "Please enter your phone number:");

        let mut model = AuthModel::new();
        model.set_auth_state(AuthState::WaitCode);
        assert_eq!(model.get_prompt(), "Please enter the verification code:");

        model.set_auth_state(AuthState::WaitPassword);
        assert_eq!(model.get_prompt(), "Please enter your 2FA password:");
    }

    #[test]
    fn test_loading_messages() {
        let model = AuthModel::new();
        assert_eq!(model.get_loading_message(), "Sending verification code...");

        let mut model = AuthModel::new();
        model.set_auth_state(AuthState::WaitCode);
        assert_eq!(model.get_loading_message(), "Verifying code...");
    }

    #[test]
    fn test_password_mode_for_2fa() {
        let mut model = AuthModel::new();
        model.set_auth_state(AuthState::WaitPassword);

        // Input characters
        model.input.set_value("password");

        // The echo mode should be password
        let (_paragraph, _) = model.input.render_paragraph();
        // We can't easily check the rendered output, but we can verify
        // the echo mode was set by checking the model
    }
}
