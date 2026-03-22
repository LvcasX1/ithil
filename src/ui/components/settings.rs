//! Settings view component.
//!
//! This module provides the settings model and widget for configuring
//! the application, with support for:
//! - Multiple settings sections (General, Appearance, Keyboard, Privacy, Credentials)
//! - Inline editing of configuration values
//! - Navigation between sections and items
//!
//! # Architecture
//!
//! The settings view follows the Model-View-Update pattern:
//! - [`SettingsModel`]: Holds the configuration and editing state
//! - [`SettingsWidget`]: Renders the settings UI
//! - [`SettingsAction`]: Actions triggered by user input
//!
//! # Example
//!
//! ```rust,no_run
//! use ithil::app::Config;
//! use ithil::ui::components::settings::{SettingsModel, SettingsWidget};
//!
//! let config = Config::default();
//! let mut model = SettingsModel::new(config);
//!
//! // In render function:
//! // let widget = SettingsWidget::new(&model);
//! ```

use ratatui::{
    buffer::Buffer,
    layout::{Constraint, Direction, Layout, Rect},
    text::{Line, Span},
    widgets::{Block, Borders, List, ListItem, Paragraph, Widget},
};

use crate::app::Config;
use crate::ui::keys::Action;
use crate::ui::styles::Styles;

/// Settings section identifier.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum SettingsSection {
    /// General application settings
    #[default]
    General,
    /// Appearance settings (theme, layout, etc.)
    Appearance,
    /// Keyboard settings (vim mode, bindings)
    Keyboard,
    /// Privacy settings
    Privacy,
    /// Telegram credentials
    Credentials,
}

impl SettingsSection {
    /// Returns all sections in order.
    #[must_use]
    pub const fn all() -> [Self; 5] {
        [
            Self::General,
            Self::Appearance,
            Self::Keyboard,
            Self::Privacy,
            Self::Credentials,
        ]
    }

    /// Returns the display name for this section.
    #[must_use]
    pub const fn name(&self) -> &'static str {
        match self {
            Self::General => "General",
            Self::Appearance => "Appearance",
            Self::Keyboard => "Keyboard",
            Self::Privacy => "Privacy",
            Self::Credentials => "Credentials",
        }
    }

    /// Returns the next section (wrapping).
    #[must_use]
    pub const fn next(self) -> Self {
        match self {
            Self::General => Self::Appearance,
            Self::Appearance => Self::Keyboard,
            Self::Keyboard => Self::Privacy,
            Self::Privacy => Self::Credentials,
            Self::Credentials => Self::General,
        }
    }

    /// Returns the previous section (wrapping).
    #[must_use]
    pub const fn previous(self) -> Self {
        match self {
            Self::General => Self::Credentials,
            Self::Appearance => Self::General,
            Self::Keyboard => Self::Appearance,
            Self::Privacy => Self::Keyboard,
            Self::Credentials => Self::Privacy,
        }
    }
}

/// Model for the settings view.
///
/// This struct holds the configuration being edited and tracks
/// the current editing state.
#[derive(Debug, Clone)]
pub struct SettingsModel {
    /// Current configuration
    pub config: Config,
    /// Currently selected section
    pub current_section: SettingsSection,
    /// Currently selected item index within the section
    pub selected_item: usize,
    /// Whether we're in edit mode
    pub editing: bool,
    /// Current edit value
    pub edit_value: String,
    /// Whether there are unsaved changes
    pub has_changes: bool,
}

impl SettingsModel {
    /// Creates a new settings model with the given configuration.
    ///
    /// # Arguments
    ///
    /// * `config` - The configuration to edit
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::app::Config;
    /// use ithil::ui::components::settings::SettingsModel;
    ///
    /// let config = Config::default();
    /// let model = SettingsModel::new(config);
    /// assert!(!model.has_changes);
    /// ```
    #[must_use]
    pub fn new(config: Config) -> Self {
        Self {
            config,
            current_section: SettingsSection::default(),
            selected_item: 0,
            editing: false,
            edit_value: String::new(),
            has_changes: false,
        }
    }

    /// Handles an action from the key bindings.
    ///
    /// Returns a [`SettingsAction`] if the action triggers an external
    /// operation (like closing the settings).
    pub fn handle_action(&mut self, action: Action) -> Option<SettingsAction> {
        if self.editing {
            return self.handle_edit_action(action);
        }

        match action {
            Action::Up => {
                self.select_previous_item();
                None
            },
            Action::Down => {
                self.select_next_item();
                None
            },
            Action::Left => {
                self.previous_section();
                None
            },
            Action::Right => {
                self.next_section();
                None
            },
            Action::OpenChat | Action::FocusInput => {
                self.start_editing();
                None
            },
            Action::CancelAction => Some(SettingsAction::Close),
            _ => None,
        }
    }

    /// Handles actions while in edit mode.
    fn handle_edit_action(&mut self, action: Action) -> Option<SettingsAction> {
        match action {
            Action::CancelAction => {
                self.cancel_editing();
                None
            },
            Action::SendMessage | Action::OpenChat => {
                self.apply_edit();
                None
            },
            Action::Backspace => {
                self.edit_value.pop();
                None
            },
            _ => None,
        }
    }

    /// Handles character input while editing.
    ///
    /// # Arguments
    ///
    /// * `c` - The character to insert
    pub fn handle_char(&mut self, c: char) {
        if self.editing {
            self.edit_value.push(c);
        }
    }

    /// Moves to the previous section.
    fn previous_section(&mut self) {
        self.current_section = self.current_section.previous();
        self.selected_item = 0;
    }

    /// Moves to the next section.
    fn next_section(&mut self) {
        self.current_section = self.current_section.next();
        self.selected_item = 0;
    }

    /// Selects the previous item in the current section.
    fn select_previous_item(&mut self) {
        if self.selected_item > 0 {
            self.selected_item -= 1;
        }
    }

    /// Selects the next item in the current section.
    fn select_next_item(&mut self) {
        let items = self.get_section_items();
        if self.selected_item < items.len().saturating_sub(1) {
            self.selected_item += 1;
        }
    }

    /// Starts editing the current item.
    fn start_editing(&mut self) {
        self.editing = true;
        self.edit_value = self.get_current_value();
    }

    /// Cancels the current edit.
    fn cancel_editing(&mut self) {
        self.editing = false;
        self.edit_value.clear();
    }

    /// Applies the current edit to the configuration.
    fn apply_edit(&mut self) {
        if !self.edit_value.is_empty() {
            self.set_current_value(self.edit_value.clone());
            self.has_changes = true;
        }
        self.editing = false;
        self.edit_value.clear();
    }

    /// Gets the current value of the selected item.
    fn get_current_value(&self) -> String {
        match self.current_section {
            SettingsSection::General => match self.selected_item {
                0 => self.config.app.name.clone(),
                1 => self.config.app.version.clone(),
                _ => String::new(),
            },
            SettingsSection::Appearance => match self.selected_item {
                0 => self.config.ui.theme.clone(),
                1 => self.config.ui.layout.chat_list_width.to_string(),
                2 => self.config.ui.layout.conversation_width.to_string(),
                3 => self.config.ui.layout.info_width.to_string(),
                4 => self.config.ui.appearance.date_format.clone(),
                5 => self.config.ui.appearance.show_avatars.to_string(),
                6 => self.config.ui.appearance.show_status_bar.to_string(),
                7 => self.config.ui.appearance.relative_timestamps.to_string(),
                _ => String::new(),
            },
            SettingsSection::Keyboard => match self.selected_item {
                0 => self.config.ui.keyboard.vim_mode.to_string(),
                _ => String::new(),
            },
            SettingsSection::Privacy => match self.selected_item {
                0 => self.config.privacy.show_online_status.to_string(),
                1 => self.config.privacy.show_read_receipts.to_string(),
                2 => self.config.privacy.show_typing.to_string(),
                3 => self.config.privacy.stealth_mode.to_string(),
                _ => String::new(),
            },
            SettingsSection::Credentials => match self.selected_item {
                0 => self.config.telegram.use_default_credentials.to_string(),
                1 => self.config.telegram.api_id.clone(),
                2 => "[hidden]".to_string(), // Don't show API hash
                _ => String::new(),
            },
        }
    }

    /// Sets the value of the selected item.
    fn set_current_value(&mut self, value: String) {
        match self.current_section {
            SettingsSection::General => {
                if self.selected_item == 0 {
                    self.config.app.name = value;
                }
                // Version (item 1) is read-only
            },
            SettingsSection::Appearance => match self.selected_item {
                0 => self.config.ui.theme = value,
                1 => {
                    if let Ok(v) = value.parse() {
                        self.config.ui.layout.chat_list_width = v;
                    }
                },
                2 => {
                    if let Ok(v) = value.parse() {
                        self.config.ui.layout.conversation_width = v;
                    }
                },
                3 => {
                    if let Ok(v) = value.parse() {
                        self.config.ui.layout.info_width = v;
                    }
                },
                4 => self.config.ui.appearance.date_format = value,
                5 => self.config.ui.appearance.show_avatars = value.to_lowercase() == "true",
                6 => self.config.ui.appearance.show_status_bar = value.to_lowercase() == "true",
                7 => {
                    self.config.ui.appearance.relative_timestamps = value.to_lowercase() == "true";
                },
                _ => {},
            },
            SettingsSection::Keyboard => {
                if self.selected_item == 0 {
                    self.config.ui.keyboard.vim_mode = value.to_lowercase() == "true";
                }
            },
            SettingsSection::Privacy => match self.selected_item {
                0 => self.config.privacy.show_online_status = value.to_lowercase() == "true",
                1 => self.config.privacy.show_read_receipts = value.to_lowercase() == "true",
                2 => self.config.privacy.show_typing = value.to_lowercase() == "true",
                3 => self.config.privacy.stealth_mode = value.to_lowercase() == "true",
                _ => {},
            },
            SettingsSection::Credentials => match self.selected_item {
                0 => {
                    self.config.telegram.use_default_credentials = value.to_lowercase() == "true";
                },
                1 => self.config.telegram.api_id = value,
                2 => self.config.telegram.api_hash = value,
                _ => {},
            },
        }
    }

    /// Returns the modified configuration.
    #[must_use]
    pub fn get_modified_config(&self) -> Config {
        self.config.clone()
    }

    /// Returns the items for the current section.
    #[must_use]
    pub fn get_section_items(&self) -> Vec<(&'static str, String)> {
        match self.current_section {
            SettingsSection::General => vec![
                ("App Name", self.config.app.name.clone()),
                ("Version", self.config.app.version.clone()),
            ],
            SettingsSection::Appearance => vec![
                ("Theme", self.config.ui.theme.clone()),
                (
                    "Chat List Width %",
                    self.config.ui.layout.chat_list_width.to_string(),
                ),
                (
                    "Conversation Width %",
                    self.config.ui.layout.conversation_width.to_string(),
                ),
                (
                    "Sidebar Width %",
                    self.config.ui.layout.info_width.to_string(),
                ),
                ("Date Format", self.config.ui.appearance.date_format.clone()),
                (
                    "Show Avatars",
                    self.config.ui.appearance.show_avatars.to_string(),
                ),
                (
                    "Show Status Bar",
                    self.config.ui.appearance.show_status_bar.to_string(),
                ),
                (
                    "Relative Timestamps",
                    self.config.ui.appearance.relative_timestamps.to_string(),
                ),
            ],
            SettingsSection::Keyboard => {
                vec![("Vim Mode", self.config.ui.keyboard.vim_mode.to_string())]
            },
            SettingsSection::Privacy => vec![
                (
                    "Show Online Status",
                    self.config.privacy.show_online_status.to_string(),
                ),
                (
                    "Show Read Receipts",
                    self.config.privacy.show_read_receipts.to_string(),
                ),
                ("Show Typing", self.config.privacy.show_typing.to_string()),
                ("Stealth Mode", self.config.privacy.stealth_mode.to_string()),
            ],
            SettingsSection::Credentials => vec![
                (
                    "Use Default Credentials",
                    self.config.telegram.use_default_credentials.to_string(),
                ),
                ("API ID", self.config.telegram.api_id.clone()),
                (
                    "API Hash",
                    if self.config.telegram.api_hash.is_empty() {
                        "[not set]".to_string()
                    } else {
                        "[hidden]".to_string()
                    },
                ),
            ],
        }
    }

    /// Returns `true` if the settings view is in edit mode.
    #[must_use]
    pub const fn is_editing(&self) -> bool {
        self.editing
    }

    /// Resets the model to its initial state with the given config.
    pub fn reset(&mut self, config: Config) {
        self.config = config;
        self.current_section = SettingsSection::default();
        self.selected_item = 0;
        self.editing = false;
        self.edit_value.clear();
        self.has_changes = false;
    }
}

/// Actions that can be triggered from the settings view.
#[derive(Debug, Clone)]
pub enum SettingsAction {
    /// Close the settings view without saving
    Close,
    /// Save and close the settings view
    SaveAndClose(Box<Config>),
}

impl PartialEq for SettingsAction {
    fn eq(&self, other: &Self) -> bool {
        matches!(
            (self, other),
            (Self::Close, Self::Close) | (Self::SaveAndClose(_), Self::SaveAndClose(_))
        )
    }
}

/// Widget for rendering the settings view.
pub struct SettingsWidget<'a> {
    /// Reference to the settings model
    model: &'a SettingsModel,
}

impl<'a> SettingsWidget<'a> {
    /// Creates a new settings widget.
    ///
    /// # Arguments
    ///
    /// * `model` - Reference to the settings model
    #[must_use]
    pub const fn new(model: &'a SettingsModel) -> Self {
        Self { model }
    }
}

impl Widget for SettingsWidget<'_> {
    fn render(self, area: Rect, buf: &mut Buffer) {
        // Split into sections nav and content
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Section tabs
                Constraint::Min(5),    // Content
                Constraint::Length(2), // Help
            ])
            .split(area);

        self.render_tabs(chunks[0], buf);
        self.render_content(chunks[1], buf);
        self.render_help(chunks[2], buf);
    }
}

impl SettingsWidget<'_> {
    /// Renders the section tabs.
    fn render_tabs(&self, area: Rect, buf: &mut Buffer) {
        let mut spans: Vec<Span> = Vec::new();

        for section in SettingsSection::all() {
            let style = if section == self.model.current_section {
                Styles::highlight()
            } else {
                Styles::text_muted()
            };

            spans.push(Span::styled(format!(" {} ", section.name()), style));
            spans.push(Span::raw("|"));
        }

        // Remove trailing separator
        if !spans.is_empty() {
            spans.pop();
        }

        let tabs_line = Line::from(spans);
        let tabs_block = Block::default()
            .title(" Settings (←/→ to switch, Esc to close) ")
            .borders(Borders::ALL)
            .border_style(Styles::border_focused());

        let tabs_para = Paragraph::new(tabs_line).block(tabs_block);
        tabs_para.render(area, buf);
    }

    /// Renders the settings content.
    fn render_content(&self, area: Rect, buf: &mut Buffer) {
        let content_block = Block::default()
            .borders(Borders::ALL)
            .border_style(Styles::border());

        let inner = content_block.inner(area);
        content_block.render(area, buf);

        let items = self.model.get_section_items();
        let list_items: Vec<ListItem> = items
            .iter()
            .enumerate()
            .map(|(idx, (label, value))| {
                let is_selected = idx == self.model.selected_item;
                let style = if is_selected {
                    Styles::selected()
                } else {
                    Styles::text()
                };

                let display_value = if self.model.editing && is_selected {
                    format!("{}▏", self.model.edit_value)
                } else {
                    value.clone()
                };

                let line = Line::from(vec![
                    Span::styled(format!("{label}: "), Styles::text_muted()),
                    Span::styled(display_value, style),
                ]);
                ListItem::new(line)
            })
            .collect();

        let list = List::new(list_items);
        list.render(inner, buf);
    }

    /// Renders the help line.
    fn render_help(&self, area: Rect, buf: &mut Buffer) {
        let help = if self.model.editing {
            "Enter to save, Esc to cancel"
        } else if self.model.has_changes {
            "Changes pending - Enter to edit, Esc to close (changes will be lost)"
        } else {
            "Enter to edit, ←/→ section, Esc to close"
        };

        let help_para = Paragraph::new(help).style(Styles::text_muted());
        help_para.render(area, buf);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_settings_section_all() {
        let sections = SettingsSection::all();
        assert_eq!(sections.len(), 5);
        assert_eq!(sections[0], SettingsSection::General);
        assert_eq!(sections[4], SettingsSection::Credentials);
    }

    #[test]
    fn test_settings_section_next() {
        assert_eq!(SettingsSection::General.next(), SettingsSection::Appearance);
        assert_eq!(
            SettingsSection::Appearance.next(),
            SettingsSection::Keyboard
        );
        assert_eq!(SettingsSection::Keyboard.next(), SettingsSection::Privacy);
        assert_eq!(
            SettingsSection::Privacy.next(),
            SettingsSection::Credentials
        );
        assert_eq!(
            SettingsSection::Credentials.next(),
            SettingsSection::General
        ); // wraps
    }

    #[test]
    fn test_settings_section_previous() {
        assert_eq!(
            SettingsSection::General.previous(),
            SettingsSection::Credentials
        ); // wraps
        assert_eq!(
            SettingsSection::Appearance.previous(),
            SettingsSection::General
        );
        assert_eq!(
            SettingsSection::Keyboard.previous(),
            SettingsSection::Appearance
        );
        assert_eq!(
            SettingsSection::Privacy.previous(),
            SettingsSection::Keyboard
        );
        assert_eq!(
            SettingsSection::Credentials.previous(),
            SettingsSection::Privacy
        );
    }

    #[test]
    fn test_settings_section_name() {
        assert_eq!(SettingsSection::General.name(), "General");
        assert_eq!(SettingsSection::Appearance.name(), "Appearance");
        assert_eq!(SettingsSection::Keyboard.name(), "Keyboard");
        assert_eq!(SettingsSection::Privacy.name(), "Privacy");
        assert_eq!(SettingsSection::Credentials.name(), "Credentials");
    }

    #[test]
    fn test_new_model() {
        let config = Config::default();
        let model = SettingsModel::new(config);

        assert_eq!(model.current_section, SettingsSection::General);
        assert_eq!(model.selected_item, 0);
        assert!(!model.editing);
        assert!(model.edit_value.is_empty());
        assert!(!model.has_changes);
    }

    #[test]
    fn test_section_navigation() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        assert_eq!(model.current_section, SettingsSection::General);

        model.handle_action(Action::Right);
        assert_eq!(model.current_section, SettingsSection::Appearance);
        assert_eq!(model.selected_item, 0); // Reset on section change

        model.handle_action(Action::Left);
        assert_eq!(model.current_section, SettingsSection::General);
    }

    #[test]
    fn test_item_navigation() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        // Move to Appearance section which has more items
        model.handle_action(Action::Right);
        assert_eq!(model.current_section, SettingsSection::Appearance);

        assert_eq!(model.selected_item, 0);

        model.handle_action(Action::Down);
        assert_eq!(model.selected_item, 1);

        model.handle_action(Action::Down);
        assert_eq!(model.selected_item, 2);

        model.handle_action(Action::Up);
        assert_eq!(model.selected_item, 1);

        model.handle_action(Action::Up);
        assert_eq!(model.selected_item, 0);

        // Can't go below 0
        model.handle_action(Action::Up);
        assert_eq!(model.selected_item, 0);
    }

    #[test]
    fn test_start_editing() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        assert!(!model.editing);

        model.handle_action(Action::OpenChat);

        assert!(model.editing);
        assert!(!model.edit_value.is_empty()); // Should have current value
    }

    #[test]
    fn test_cancel_editing() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        model.handle_action(Action::OpenChat);
        assert!(model.editing);

        model.handle_action(Action::CancelAction);

        assert!(!model.editing);
        assert!(model.edit_value.is_empty());
    }

    #[test]
    fn test_handle_char() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        // Characters are ignored when not editing
        model.handle_char('x');
        assert!(model.edit_value.is_empty());

        // Start editing
        model.handle_action(Action::OpenChat);
        let original_len = model.edit_value.len();

        model.handle_char('!');

        assert_eq!(model.edit_value.len(), original_len + 1);
        assert!(model.edit_value.ends_with('!'));
    }

    #[test]
    fn test_apply_edit() {
        let mut config = Config::default();
        config.app.name = "Original".to_string();
        let mut model = SettingsModel::new(config);

        // Start editing
        model.handle_action(Action::OpenChat);
        model.edit_value = "NewName".to_string();

        // Apply
        model.handle_action(Action::SendMessage);

        assert!(!model.editing);
        assert!(model.has_changes);
        assert_eq!(model.config.app.name, "NewName");
    }

    #[test]
    fn test_backspace_in_edit() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        model.handle_action(Action::OpenChat);
        model.edit_value = "Test".to_string();

        model.handle_action(Action::Backspace);

        assert_eq!(model.edit_value, "Tes");
    }

    #[test]
    fn test_close_action() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        let action = model.handle_action(Action::CancelAction);

        assert_eq!(action, Some(SettingsAction::Close));
    }

    #[test]
    fn test_get_section_items() {
        let config = Config::default();
        let model = SettingsModel::new(config);

        let items = model.get_section_items();

        assert!(!items.is_empty());
        assert_eq!(items[0].0, "App Name");
    }

    #[test]
    fn test_get_modified_config() {
        let mut config = Config::default();
        config.app.name = "Original".to_string();
        let mut model = SettingsModel::new(config);

        model.config.app.name = "Modified".to_string();

        let modified = model.get_modified_config();
        assert_eq!(modified.app.name, "Modified");
    }

    #[test]
    fn test_reset() {
        let config = Config::default();
        let mut model = SettingsModel::new(config.clone());

        // Make some changes
        model.handle_action(Action::Right);
        model.handle_action(Action::Down);
        model.handle_action(Action::OpenChat);
        model.edit_value = "test".to_string();
        model.handle_action(Action::SendMessage);

        assert!(model.has_changes);
        assert_ne!(model.current_section, SettingsSection::General);

        // Reset
        model.reset(config);

        assert_eq!(model.current_section, SettingsSection::General);
        assert_eq!(model.selected_item, 0);
        assert!(!model.editing);
        assert!(model.edit_value.is_empty());
        assert!(!model.has_changes);
    }

    #[test]
    fn test_is_editing() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        assert!(!model.is_editing());

        model.handle_action(Action::OpenChat);

        assert!(model.is_editing());
    }

    #[test]
    fn test_privacy_settings() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        // Navigate to Privacy section
        model.current_section = SettingsSection::Privacy;
        model.selected_item = 0;

        let items = model.get_section_items();
        assert_eq!(items.len(), 4);
        assert_eq!(items[0].0, "Show Online Status");
        assert_eq!(items[3].0, "Stealth Mode");
    }

    #[test]
    fn test_credentials_hides_api_hash() {
        let mut config = Config::default();
        config.telegram.api_hash = "secret123".to_string();
        let mut model = SettingsModel::new(config);

        model.current_section = SettingsSection::Credentials;

        let items = model.get_section_items();
        assert_eq!(items[2].1, "[hidden]");
    }

    #[test]
    fn test_credentials_shows_not_set() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        model.current_section = SettingsSection::Credentials;

        let items = model.get_section_items();
        assert_eq!(items[2].1, "[not set]");
    }

    #[test]
    fn test_settings_action_variants() {
        let close = SettingsAction::Close;
        let save = SettingsAction::SaveAndClose(Box::new(Config::default()));

        assert!(matches!(close, SettingsAction::Close));
        assert!(matches!(save, SettingsAction::SaveAndClose(_)));
    }

    #[test]
    fn test_item_bounds_checking() {
        let config = Config::default();
        let mut model = SettingsModel::new(config);

        // Navigate to General section (only 2 items)
        model.current_section = SettingsSection::General;

        // Try to go beyond bounds
        model.selected_item = 0;
        model.handle_action(Action::Up);
        assert_eq!(model.selected_item, 0); // Should stay at 0

        model.selected_item = 1;
        model.handle_action(Action::Down);
        assert_eq!(model.selected_item, 1); // Should stay at max
    }
}
