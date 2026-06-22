//! Modal file browser for selecting a file to attach to a message.
//!
//! Rendered as an overlay (like the help overlay). Navigation logic is kept
//! separate from rendering so it can be unit-tested against a temp directory.

use std::path::PathBuf;

use ratatui::{
    layout::Rect,
    text::{Line, Span},
    widgets::{Block, Borders, Clear, List, ListItem, ListState},
    Frame,
};

use crate::ui::styles::Styles;

/// Result of activating the current selection in the file picker.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum FilePickerAction {
    /// Navigated or descended into a directory — nothing to report.
    None,
    /// User confirmed a file; contains the absolute path to the selected file.
    Selected(PathBuf),
}

#[derive(Debug, Clone)]
struct Entry {
    path: PathBuf,
    label: String,
    is_dir: bool,
}

/// Modal file browser overlay for selecting a file to attach.
///
/// Displays a navigable directory listing. Directories are listed before
/// files. A `..` entry is prepended whenever a parent directory exists.
#[derive(Debug)]
pub struct FilePicker {
    current_dir: PathBuf,
    entries: Vec<Entry>,
    selected: usize,
}

impl FilePicker {
    /// Creates a new picker rooted at the user's home directory (falling back
    /// to the current working directory).
    #[must_use]
    pub fn new() -> Self {
        let start = dirs::home_dir()
            .or_else(|| std::env::current_dir().ok())
            .unwrap_or_else(|| PathBuf::from("."));
        Self::with_dir(start)
    }

    /// Creates a picker rooted at `dir` and loads its initial listing.
    #[must_use]
    pub fn with_dir(dir: PathBuf) -> Self {
        let mut picker = Self {
            current_dir: dir,
            entries: Vec::new(),
            selected: 0,
        };
        picker.reload();
        picker
    }

    // Reads `current_dir` from disk and rebuilds `entries` (dirs first, then files).
    fn reload(&mut self) {
        let mut dirs: Vec<Entry> = Vec::new();
        let mut files: Vec<Entry> = Vec::new();

        if let Ok(read) = std::fs::read_dir(&self.current_dir) {
            for dent in read.flatten() {
                let path = dent.path();
                let is_dir = path.is_dir();
                let name = dent.file_name().to_string_lossy().into_owned();
                let label = if is_dir { format!("{name}/") } else { name };
                let entry = Entry {
                    path,
                    label,
                    is_dir,
                };
                if is_dir {
                    dirs.push(entry);
                } else {
                    files.push(entry);
                }
            }
        }

        dirs.sort_by(|a, b| a.label.to_lowercase().cmp(&b.label.to_lowercase()));
        files.sort_by(|a, b| a.label.to_lowercase().cmp(&b.label.to_lowercase()));

        let mut entries = Vec::with_capacity(dirs.len() + files.len() + 1);
        if let Some(parent) = self.current_dir.parent() {
            entries.push(Entry {
                path: parent.to_path_buf(),
                label: "..".to_string(),
                is_dir: true,
            });
        }
        entries.extend(dirs);
        entries.extend(files);

        self.entries = entries;
        self.selected = 0;
    }

    /// Moves the selection up by one row; clamps at the first entry.
    pub fn select_previous(&mut self) {
        self.selected = self.selected.saturating_sub(1);
    }

    /// Moves the selection down by one row; clamps at the last entry.
    pub fn select_next(&mut self) {
        if !self.entries.is_empty() {
            self.selected = (self.selected + 1).min(self.entries.len() - 1);
        }
    }

    /// Activates the currently highlighted entry.
    ///
    /// If the entry is a directory (including `..`), descends into it and
    /// returns [`FilePickerAction::None`]. If it is a file, returns
    /// [`FilePickerAction::Selected`] with the file's path.
    pub fn activate(&mut self) -> FilePickerAction {
        let Some(entry) = self.entries.get(self.selected) else {
            return FilePickerAction::None;
        };
        if entry.is_dir {
            self.current_dir = entry.path.clone();
            self.reload();
            FilePickerAction::None
        } else {
            FilePickerAction::Selected(entry.path.clone())
        }
    }

    /// Returns the display labels of all entries in the current directory listing.
    #[must_use]
    pub fn entry_names(&self) -> Vec<String> {
        self.entries.iter().map(|e| e.label.clone()).collect()
    }

    /// Returns the zero-based index of the currently highlighted entry.
    #[must_use]
    pub const fn selected_index(&self) -> usize {
        self.selected
    }

    /// Renders the file picker as a centered modal overlay on the given frame.
    pub fn render(&self, frame: &mut Frame) {
        let area = frame.area();
        let w = 60.min(area.width.saturating_sub(4));
        let h = 20.min(area.height.saturating_sub(4));
        let x = (area.width.saturating_sub(w)) / 2;
        let y = (area.height.saturating_sub(h)) / 2;
        let modal = Rect::new(x, y, w, h);

        frame.render_widget(Clear, modal);

        let title = format!(" Attach file — {} ", self.current_dir.display());
        let block = Block::default()
            .title(Span::styled(title, Styles::text_bright()))
            .borders(Borders::ALL)
            .border_style(Styles::border_focused())
            .style(Styles::modal_background());

        let items: Vec<ListItem> = self
            .entries
            .iter()
            .map(|e| {
                let style = if e.is_dir {
                    Styles::text_accent()
                } else {
                    Styles::text()
                };
                ListItem::new(Line::from(Span::styled(e.label.clone(), style)))
            })
            .collect();

        let list = List::new(items)
            .block(block)
            .highlight_style(Styles::highlight());

        let mut state = ListState::default();
        if !self.entries.is_empty() {
            state.select(Some(self.selected));
        }
        frame.render_stateful_widget(list, modal, &mut state);
    }
}

impl Default for FilePicker {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::{FilePicker, FilePickerAction};
    use std::fs;

    fn temp_tree() -> std::path::PathBuf {
        use std::sync::atomic::{AtomicU32, Ordering};
        static N: AtomicU32 = AtomicU32::new(0);
        let base = std::env::temp_dir().join(format!(
            "ithil_fp_test_{}_{}",
            std::process::id(),
            N.fetch_add(1, Ordering::Relaxed)
        ));
        fs::create_dir_all(base.join("subdir")).unwrap();
        fs::write(base.join("file_a.txt"), b"a").unwrap();
        fs::write(base.join("subdir").join("nested.txt"), b"b").unwrap();
        base
    }

    #[test]
    fn lists_dirs_before_files_and_includes_parent() {
        let dir = temp_tree();
        let picker = FilePicker::with_dir(dir);
        let names = picker.entry_names();
        assert_eq!(names.first().map(String::as_str), Some(".."));
        let sub = names
            .iter()
            .position(|n: &String| n.starts_with("subdir"))
            .unwrap();
        let file = names
            .iter()
            .position(|n: &String| n == "file_a.txt")
            .unwrap();
        assert!(sub < file);
    }

    #[test]
    fn descends_into_directory_then_selects_file() {
        let dir = temp_tree();
        let mut picker = FilePicker::with_dir(dir);
        picker.select_next();
        assert_eq!(picker.activate(), FilePickerAction::None);
        picker.select_next();
        match picker.activate() {
            FilePickerAction::Selected(p) => assert!(p.ends_with("nested.txt")),
            other @ FilePickerAction::None => panic!("expected Selected, got {other:?}"),
        }
    }

    #[test]
    fn parent_entry_ascends() {
        let dir = temp_tree();
        let mut picker = FilePicker::with_dir(dir.join("subdir"));
        assert_eq!(picker.activate(), FilePickerAction::None);
        assert!(picker
            .entry_names()
            .iter()
            .any(|n: &String| n == "file_a.txt"));
    }

    #[test]
    fn selection_clamps_at_bounds() {
        let dir = temp_tree();
        let mut picker = FilePicker::with_dir(dir);
        picker.select_previous();
        assert_eq!(picker.selected_index(), 0);
        for _ in 0..50 {
            picker.select_next();
        }
        assert_eq!(picker.selected_index(), picker.entry_names().len() - 1);
    }
}
