//! Text input component with cursor support.
//!
//! This component provides a text input field with:
//! - Cursor navigation (left/right, home/end)
//! - Text insertion and deletion
//! - Password masking mode
//! - Placeholder text support
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::ui::components::InputComponent;
//!
//! let mut input = InputComponent::new("Enter phone number...");
//! input.insert_char('1');
//! input.insert_char('2');
//! input.insert_char('3');
//!
//! assert_eq!(input.value(), "123");
//! ```

use crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use ratatui::{
    buffer::Buffer,
    layout::Rect,
    text::{Line, Span},
    widgets::{Paragraph, StatefulWidget, Widget},
};
use unicode_width::UnicodeWidthStr;

use crate::ui::styles::Styles;

/// Input echo mode for password fields.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum EchoMode {
    /// Display characters normally
    #[default]
    Normal,
    /// Display asterisks instead of characters
    Password,
}

/// A text input component with cursor support.
///
/// This component handles text input, cursor navigation, and rendering.
/// It supports both normal text input and password masking modes.
#[derive(Debug, Clone)]
pub struct InputComponent {
    /// The current input value
    value: String,
    /// Cursor position (character index)
    cursor: usize,
    /// Placeholder text shown when empty
    placeholder: String,
    /// Maximum character limit (0 = unlimited)
    char_limit: usize,
    /// Echo mode (normal or password)
    echo_mode: EchoMode,
    /// Whether the input is focused
    focused: bool,
    /// Visible width for rendering
    width: u16,
}

impl Default for InputComponent {
    fn default() -> Self {
        Self::new("")
    }
}

impl InputComponent {
    /// Creates a new input component with the given placeholder text.
    ///
    /// # Arguments
    ///
    /// * `placeholder` - Text to display when the input is empty
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::ui::components::InputComponent;
    ///
    /// let input = InputComponent::new("Enter your name...");
    /// assert!(input.value().is_empty());
    /// ```
    #[must_use]
    pub fn new(placeholder: impl Into<String>) -> Self {
        Self {
            value: String::new(),
            cursor: 0,
            placeholder: placeholder.into(),
            char_limit: 0,
            echo_mode: EchoMode::Normal,
            focused: true,
            width: 30,
        }
    }

    /// Creates a new password input.
    ///
    /// Characters will be displayed as asterisks.
    #[must_use]
    pub fn password(placeholder: impl Into<String>) -> Self {
        let mut input = Self::new(placeholder);
        input.echo_mode = EchoMode::Password;
        input
    }

    /// Sets the character limit.
    ///
    /// Set to 0 for unlimited.
    #[must_use]
    pub const fn with_char_limit(mut self, limit: usize) -> Self {
        self.char_limit = limit;
        self
    }

    /// Sets the visible width for rendering.
    #[must_use]
    pub const fn with_width(mut self, width: u16) -> Self {
        self.width = width;
        self
    }

    /// Sets the echo mode.
    #[must_use]
    pub const fn with_echo_mode(mut self, mode: EchoMode) -> Self {
        self.echo_mode = mode;
        self
    }

    /// Gets the current input value.
    #[must_use]
    pub fn value(&self) -> &str {
        &self.value
    }

    /// Sets the input value and resets cursor to end.
    pub fn set_value(&mut self, value: impl Into<String>) {
        self.value = value.into();
        self.cursor = self.value.chars().count();
    }

    /// Clears the input value.
    pub fn clear(&mut self) {
        self.value.clear();
        self.cursor = 0;
    }

    /// Returns `true` if the input is empty.
    #[must_use]
    pub fn is_empty(&self) -> bool {
        self.value.is_empty()
    }

    /// Sets whether the input is focused.
    pub fn set_focused(&mut self, focused: bool) {
        self.focused = focused;
    }

    /// Returns `true` if the input is focused.
    #[must_use]
    pub const fn is_focused(&self) -> bool {
        self.focused
    }

    /// Sets the placeholder text.
    pub fn set_placeholder(&mut self, placeholder: impl Into<String>) {
        self.placeholder = placeholder.into();
    }

    /// Sets the echo mode.
    pub fn set_echo_mode(&mut self, mode: EchoMode) {
        self.echo_mode = mode;
    }

    /// Sets the character limit.
    pub fn set_char_limit(&mut self, limit: usize) {
        self.char_limit = limit;
    }

    /// Returns the current cursor position.
    #[must_use]
    pub const fn cursor(&self) -> usize {
        self.cursor
    }

    /// Handles a key event.
    ///
    /// Returns `true` if the event was handled, `false` otherwise.
    pub fn handle_input(&mut self, key: KeyEvent) -> bool {
        if !self.focused {
            return false;
        }

        match key.code {
            KeyCode::Char(c) => {
                // Check for Ctrl+A (select all / go to start)
                if key.modifiers.contains(KeyModifiers::CONTROL) && c == 'a' {
                    self.cursor = 0;
                    return true;
                }
                // Check for Ctrl+E (go to end)
                if key.modifiers.contains(KeyModifiers::CONTROL) && c == 'e' {
                    self.cursor = self.value.chars().count();
                    return true;
                }
                // Check for Ctrl+U (clear line)
                if key.modifiers.contains(KeyModifiers::CONTROL) && c == 'u' {
                    self.clear();
                    return true;
                }
                // Check for Ctrl+W (delete word)
                if key.modifiers.contains(KeyModifiers::CONTROL) && c == 'w' {
                    self.delete_word_backward();
                    return true;
                }

                self.insert_char(c);
                true
            }
            KeyCode::Backspace => {
                self.delete_char_backward();
                true
            }
            KeyCode::Delete => {
                self.delete_char_forward();
                true
            }
            KeyCode::Left => {
                self.move_cursor_left();
                true
            }
            KeyCode::Right => {
                self.move_cursor_right();
                true
            }
            KeyCode::Home => {
                self.cursor = 0;
                true
            }
            KeyCode::End => {
                self.cursor = self.value.chars().count();
                true
            }
            _ => false,
        }
    }

    /// Inserts a character at the cursor position.
    pub fn insert_char(&mut self, c: char) {
        // Check character limit
        if self.char_limit > 0 && self.value.chars().count() >= self.char_limit {
            return;
        }

        // Insert at cursor position
        let byte_index = self.cursor_byte_index();
        self.value.insert(byte_index, c);
        self.cursor += 1;
    }

    /// Deletes the character before the cursor.
    pub fn delete_char_backward(&mut self) {
        if self.cursor == 0 {
            return;
        }

        // Find byte range of character before cursor
        let byte_index = self.cursor_byte_index();
        let prev_char_byte_index = self
            .value
            .char_indices()
            .take(self.cursor)
            .last()
            .map_or(0, |(i, _)| i);

        self.value.drain(prev_char_byte_index..byte_index);
        self.cursor -= 1;
    }

    /// Deletes the character after the cursor.
    pub fn delete_char_forward(&mut self) {
        let char_count = self.value.chars().count();
        if self.cursor >= char_count {
            return;
        }

        let byte_index = self.cursor_byte_index();

        // Find byte index of next character
        let next_byte_index = self
            .value
            .char_indices()
            .nth(self.cursor + 1)
            .map_or(self.value.len(), |(i, _)| i);

        self.value.drain(byte_index..next_byte_index);
    }

    /// Deletes the word before the cursor.
    pub fn delete_word_backward(&mut self) {
        if self.cursor == 0 {
            return;
        }

        let chars: Vec<char> = self.value.chars().collect();
        let mut new_cursor = self.cursor;

        // Skip trailing whitespace
        while new_cursor > 0 && chars[new_cursor - 1].is_whitespace() {
            new_cursor -= 1;
        }

        // Skip word characters
        while new_cursor > 0 && !chars[new_cursor - 1].is_whitespace() {
            new_cursor -= 1;
        }

        // Calculate byte indices
        let start_byte = chars
            .iter()
            .take(new_cursor)
            .map(|c| c.len_utf8())
            .sum::<usize>();
        let end_byte = self.cursor_byte_index();

        self.value.drain(start_byte..end_byte);
        self.cursor = new_cursor;
    }

    /// Moves the cursor left by one character.
    pub fn move_cursor_left(&mut self) {
        if self.cursor > 0 {
            self.cursor -= 1;
        }
    }

    /// Moves the cursor right by one character.
    pub fn move_cursor_right(&mut self) {
        let char_count = self.value.chars().count();
        if self.cursor < char_count {
            self.cursor += 1;
        }
    }

    /// Returns the byte index of the cursor position.
    fn cursor_byte_index(&self) -> usize {
        self.value
            .char_indices()
            .nth(self.cursor)
            .map_or(self.value.len(), |(i, _)| i)
    }

    /// Renders the input as a paragraph widget.
    ///
    /// Returns the paragraph along with the cursor position for rendering.
    #[allow(clippy::cast_possible_truncation)]
    #[must_use]
    pub fn render_paragraph(&self) -> (Paragraph<'_>, Option<(u16, u16)>) {
        let display_value = if self.value.is_empty() {
            self.placeholder.clone()
        } else {
            match self.echo_mode {
                EchoMode::Normal => self.value.clone(),
                EchoMode::Password => "*".repeat(self.value.chars().count()),
            }
        };

        let style = if self.value.is_empty() {
            Styles::input_placeholder()
        } else {
            Styles::input()
        };

        let paragraph = Paragraph::new(Line::from(Span::styled(display_value, style)));

        // Calculate cursor position for display
        let cursor_pos = if self.focused && !self.value.is_empty() {
            let display_text = match self.echo_mode {
                EchoMode::Normal => &self.value[..self.cursor_byte_index()],
                EchoMode::Password => &"*".repeat(self.cursor),
            };
            Some((display_text.width() as u16, 0))
        } else if self.focused && self.value.is_empty() {
            Some((0, 0))
        } else {
            None
        };

        (paragraph, cursor_pos)
    }
}

/// State for the input widget (used with `StatefulWidget` pattern).
#[derive(Debug, Default)]
pub struct InputState {
    /// Cursor position in screen coordinates
    pub cursor_position: Option<(u16, u16)>,
}

impl StatefulWidget for &InputComponent {
    type State = InputState;

    #[allow(clippy::cast_possible_truncation)]
    fn render(self, area: Rect, buf: &mut Buffer, state: &mut Self::State) {
        let display_value = if self.value.is_empty() {
            &self.placeholder
        } else {
            match self.echo_mode {
                EchoMode::Normal => &self.value,
                EchoMode::Password => {
                    // For password mode, we need a temporary string
                    // Since we can't return a reference to a local variable,
                    // we handle this differently in render
                    &self.value
                }
            }
        };

        let style = if self.value.is_empty() {
            Styles::input_placeholder()
        } else {
            Styles::input()
        };

        // Create the text to display
        let text = if self.echo_mode == EchoMode::Password && !self.value.is_empty() {
            "*".repeat(self.value.chars().count())
        } else {
            display_value.clone()
        };

        let paragraph = Paragraph::new(Line::from(Span::styled(text, style)));
        paragraph.render(area, buf);

        // Calculate and store cursor position
        if self.focused {
            let cursor_text = if self.echo_mode == EchoMode::Password {
                "*".repeat(self.cursor)
            } else {
                self.value.chars().take(self.cursor).collect::<String>()
            };

            let cursor_x = area.x + cursor_text.width() as u16;
            let cursor_y = area.y;

            // Ensure cursor is within area bounds
            if cursor_x < area.x + area.width {
                state.cursor_position = Some((cursor_x, cursor_y));
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_input() {
        let input = InputComponent::new("placeholder");
        assert!(input.is_empty());
        assert_eq!(input.cursor(), 0);
        assert!(input.is_focused());
    }

    #[test]
    fn test_insert_char() {
        let mut input = InputComponent::new("");
        input.insert_char('a');
        input.insert_char('b');
        input.insert_char('c');

        assert_eq!(input.value(), "abc");
        assert_eq!(input.cursor(), 3);
    }

    #[test]
    fn test_delete_char_backward() {
        let mut input = InputComponent::new("");
        input.set_value("abc");

        input.delete_char_backward();
        assert_eq!(input.value(), "ab");
        assert_eq!(input.cursor(), 2);

        input.delete_char_backward();
        assert_eq!(input.value(), "a");
        assert_eq!(input.cursor(), 1);
    }

    #[test]
    fn test_delete_char_backward_at_start() {
        let mut input = InputComponent::new("");
        input.set_value("abc");
        input.cursor = 0;

        input.delete_char_backward();
        assert_eq!(input.value(), "abc");
        assert_eq!(input.cursor(), 0);
    }

    #[test]
    fn test_delete_char_forward() {
        let mut input = InputComponent::new("");
        input.set_value("abc");
        input.cursor = 0;

        input.delete_char_forward();
        assert_eq!(input.value(), "bc");
        assert_eq!(input.cursor(), 0);
    }

    #[test]
    fn test_cursor_navigation() {
        let mut input = InputComponent::new("");
        input.set_value("abc");
        assert_eq!(input.cursor(), 3);

        input.move_cursor_left();
        assert_eq!(input.cursor(), 2);

        input.move_cursor_left();
        input.move_cursor_left();
        assert_eq!(input.cursor(), 0);

        // Can't go below 0
        input.move_cursor_left();
        assert_eq!(input.cursor(), 0);

        input.move_cursor_right();
        assert_eq!(input.cursor(), 1);
    }

    #[test]
    fn test_char_limit() {
        let mut input = InputComponent::new("").with_char_limit(3);
        input.insert_char('a');
        input.insert_char('b');
        input.insert_char('c');
        input.insert_char('d'); // Should be ignored

        assert_eq!(input.value(), "abc");
    }

    #[test]
    fn test_clear() {
        let mut input = InputComponent::new("");
        input.set_value("hello");
        input.clear();

        assert!(input.is_empty());
        assert_eq!(input.cursor(), 0);
    }

    #[test]
    fn test_unicode_handling() {
        let mut input = InputComponent::new("");
        input.insert_char('你');
        input.insert_char('好');

        assert_eq!(input.value(), "你好");
        assert_eq!(input.cursor(), 2);

        input.delete_char_backward();
        assert_eq!(input.value(), "你");
        assert_eq!(input.cursor(), 1);
    }

    #[test]
    fn test_password_mode() {
        let input = InputComponent::password("password").with_char_limit(100);
        assert_eq!(input.echo_mode, EchoMode::Password);
    }

    #[test]
    fn test_delete_word_backward() {
        let mut input = InputComponent::new("");
        input.set_value("hello world");

        input.delete_word_backward();
        assert_eq!(input.value(), "hello ");
        assert_eq!(input.cursor(), 6);

        input.delete_word_backward();
        assert_eq!(input.value(), "");
        assert_eq!(input.cursor(), 0);
    }

    #[test]
    fn test_insert_in_middle() {
        let mut input = InputComponent::new("");
        input.set_value("ac");
        input.cursor = 1; // Position after 'a'

        input.insert_char('b');
        assert_eq!(input.value(), "abc");
        assert_eq!(input.cursor(), 2);
    }

    #[test]
    fn test_focused_state() {
        let mut input = InputComponent::new("");
        assert!(input.is_focused());

        input.set_focused(false);
        assert!(!input.is_focused());

        // Key input should be ignored when not focused
        let key = KeyEvent::from(KeyCode::Char('a'));
        assert!(!input.handle_input(key));
    }
}
