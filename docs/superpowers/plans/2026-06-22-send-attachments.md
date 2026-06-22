# Send Message Attachments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the user attach a local file to an outgoing Telegram message — pick it via a browser modal, stage it with an optional caption, and send images as photos and everything else as documents.

**Architecture:** A new `FilePicker` overlay component (mirrors how `show_help` gates rendering/keys) returns a chosen path. The path is staged on `ConversationModel.pending_attachment` and shown as a banner above the input. On Enter, the conversation emits a new `SendMessageWithAttachment` action that flows through `App` to a new `TelegramClient::send_file`, which uploads via grammers and attaches the file to an `InputMessage`.

**Tech Stack:** Rust, ratatui, crossterm, grammers-client 0.9 (`upload_file` + `InputMessage::photo/document`), tokio.

**Spec:** `docs/superpowers/specs/2026-06-22-send-attachments-design.md`

---

## File Structure

| File | Change | Responsibility |
|---|---|---|
| `src/telegram/messages.rs` | Modify | Add `send_file` + private `is_image` classifier (+ unit tests) |
| `src/ui/components/file_picker.rs` | Create | Modal file-browser component: state, navigation, render, key handling |
| `src/ui/components/mod.rs` | Modify | Register + export the file picker |
| `src/ui/components/conversation.rs` | Modify | `pending_attachment` field, `submit_input` branch, Esc precedence, banner render, new `ConversationAction` variant |
| `src/ui/keys.rs` | Modify | `Action::AttachFile`, `Ctrl+T` binding (both modes), `Display`, help text |
| `src/ui/app.rs` | Modify | `file_picker` field, `AppAction::SendMessageWithAttachment`, picker open/intercept, overlay render, `handle_send_message_with_attachment` |

Conventions to follow (already in the codebase): every `messages.rs` method opens with `let client = self.require_authorized().await?;` then `let peer_ref = self.get_peer_ref(chat_id).await?;` and ends by caching the sent message with `self.cache().add_message(chat_id, message.clone())`. The `reply_to` i32 cast uses `#[allow(clippy::cast_possible_truncation)]`. Run `cargo fmt` before every commit and keep `cargo clippy` clean.

---

## Task 1: `is_image` classifier + `send_file` in the Telegram client

**Files:**
- Modify: `src/telegram/messages.rs`
- Test: `src/telegram/messages.rs` (`#[cfg(test)]` module at end of file)

- [ ] **Step 1: Write the failing test for the classifier**

Add to the bottom of `src/telegram/messages.rs`:

```rust
#[cfg(test)]
mod tests {
    use super::is_image;
    use std::path::Path;

    #[test]
    fn images_classify_as_photo() {
        for p in ["a.jpg", "a.jpeg", "a.PNG", "dir/sub/photo.WebP", "x.bmp"] {
            assert!(is_image(Path::new(p)), "{p} should be an image");
        }
    }

    #[test]
    fn non_images_classify_as_document() {
        for p in ["a.gif", "clip.mp4", "notes.txt", "report.pdf", "noext", "a.tar.gz"] {
            assert!(!is_image(Path::new(p)), "{p} should NOT be an image");
        }
    }
}
```

- [ ] **Step 2: Run it to confirm it fails to compile (no `is_image` yet)**

Run: `cargo test -p ithil is_image 2>&1 | tail -20` (or `cargo test is_image`)
Expected: compile error — `cannot find function is_image`.

- [ ] **Step 3: Implement `is_image`**

Add near the top of `src/telegram/messages.rs`, after the `use` block (module-level, not inside `impl`):

```rust
/// Returns `true` when the file extension indicates an image that Telegram
/// should receive as a compressed photo. Everything else is sent as a document.
fn is_image(path: &std::path::Path) -> bool {
    matches!(
        path.extension()
            .and_then(|e| e.to_str())
            .map(str::to_ascii_lowercase)
            .as_deref(),
        Some("jpg" | "jpeg" | "png" | "webp" | "bmp")
    )
}
```

- [ ] **Step 4: Run the test to confirm it passes**

Run: `cargo test is_image`
Expected: 2 tests pass.

- [ ] **Step 5: Add the `send_file` method**

Insert into the `impl TelegramClient` block in `src/telegram/messages.rs`, right after `send_message` (which ends ~line 161). It mirrors `send_message` exactly, swapping the `InputMessage` construction and adding the upload:

```rust
    /// Sends a file (photo or document) with an optional caption to a chat.
    ///
    /// Images (by extension) are sent as compressed photos; every other file
    /// type is sent as a document. The `text` becomes the media caption.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to send to
    /// * `text` - Caption (may be empty)
    /// * `path` - Path to the local file to upload
    /// * `reply_to` - Optional message ID to reply to
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not authorized, the chat is not found,
    /// the file cannot be read/uploaded, or sending fails.
    pub async fn send_file(
        &self,
        chat_id: i64,
        text: &str,
        path: &std::path::Path,
        reply_to: Option<i64>,
    ) -> Result<Message, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!("Uploading file to chat {}: {}", chat_id, path.display());

        let uploaded = client.upload_file(path).await?;

        let mut input_message = InputMessage::new().text(text);
        input_message = if is_image(path) {
            input_message.photo(uploaded)
        } else {
            input_message.document(uploaded)
        };

        if let Some(reply_id) = reply_to {
            #[allow(clippy::cast_possible_truncation)]
            let reply_id_i32 = reply_id as i32;
            input_message = input_message.reply_to(Some(reply_id_i32));
        }

        let sent = client
            .send_message(peer_ref, input_message)
            .await
            .map_err(TelegramError::from)?;

        let message = grammers_message_to_message(&sent);
        self.cache().add_message(chat_id, message.clone());

        debug!("Sent file message {} to chat {}", message.id, chat_id);
        Ok(message)
    }
```

Note: `client.upload_file(path).await?` works because `From<std::io::Error> for TelegramError` already exists (`error.rs:190`).

- [ ] **Step 6: Verify it compiles and tests pass**

Run: `cargo test is_image && cargo clippy --all-targets 2>&1 | tail -5`
Expected: tests pass; no new clippy warnings.

- [ ] **Step 7: Commit**

```bash
cargo fmt
git add src/telegram/messages.rs
git commit -m "Add send_file with photo/document classification to Telegram client"
```

---

## Task 2: FilePicker component

**Files:**
- Create: `src/ui/components/file_picker.rs`
- Modify: `src/ui/components/mod.rs`
- Test: `src/ui/components/file_picker.rs` (`#[cfg(test)]` module)

The component separates navigation logic (testable) from rendering. `new()` starts at `$HOME`; `with_dir()` exists for tests.

- [ ] **Step 1: Write the failing navigation tests**

Create `src/ui/components/file_picker.rs` with ONLY this content first (tests + minimal skeleton will fail to compile until Step 3):

```rust
#[cfg(test)]
mod tests {
    use super::{FilePicker, FilePickerAction};
    use std::fs;

    fn temp_tree() -> std::path::PathBuf {
        // Unique dir under the OS temp dir. No Date/rand needed: use the test
        // binary's pid + a counter via an atomic.
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
        let picker = FilePicker::with_dir(dir.clone());
        let names = picker.entry_names();
        // first entry is the ".." parent
        assert_eq!(names.first().map(String::as_str), Some(".."));
        // "subdir/" appears before "file_a.txt"
        let sub = names.iter().position(|n| n.starts_with("subdir")).unwrap();
        let file = names.iter().position(|n| n == "file_a.txt").unwrap();
        assert!(sub < file);
    }

    #[test]
    fn descends_into_directory_then_selects_file() {
        let dir = temp_tree();
        let mut picker = FilePicker::with_dir(dir.clone());
        // move selection onto "subdir" (index 1, after "..")
        picker.select_next();
        assert_eq!(picker.activate(), FilePickerAction::None); // descended, no file yet
        // now inside subdir: ".." then "nested.txt"; select the file
        picker.select_next();
        match picker.activate() {
            FilePickerAction::Selected(p) => assert!(p.ends_with("nested.txt")),
            other => panic!("expected Selected, got {other:?}"),
        }
    }

    #[test]
    fn parent_entry_ascends() {
        let dir = temp_tree();
        let mut picker = FilePicker::with_dir(dir.join("subdir"));
        // selection starts at "..", activating it goes up to `dir`
        assert_eq!(picker.activate(), FilePickerAction::None);
        assert!(picker.entry_names().iter().any(|n| n == "file_a.txt"));
    }

    #[test]
    fn selection_clamps_at_bounds() {
        let dir = temp_tree();
        let mut picker = FilePicker::with_dir(dir);
        picker.select_previous(); // already at top, stays at 0
        assert_eq!(picker.selected_index(), 0);
        for _ in 0..50 {
            picker.select_next();
        }
        assert_eq!(picker.selected_index(), picker.entry_names().len() - 1);
    }
}
```

- [ ] **Step 2: Run tests to confirm they fail**

Run: `cargo test file_picker 2>&1 | tail -20`
Expected: compile error — `FilePicker` / `FilePickerAction` not found.

- [ ] **Step 3: Implement the component**

Prepend this above the test module in `src/ui/components/file_picker.rs`:

```rust
//! Modal file browser for selecting a file to attach to a message.
//!
//! Rendered as an overlay (like the help overlay). Navigation logic is kept
//! separate from rendering so it can be unit-tested against a temp directory.

use std::path::{Path, PathBuf};

use ratatui::{
    layout::Rect,
    text::{Line, Span},
    widgets::{Block, Borders, Clear, List, ListItem, ListState},
    Frame,
};

use crate::ui::styles::Styles;

/// Result of activating the current selection or cancelling.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum FilePickerAction {
    /// Nothing to report (navigated, descended into a dir, or no-op).
    None,
    /// User pressed Esc — close without choosing.
    Cancelled,
    /// User chose this file.
    Selected(PathBuf),
}

/// A single browsable entry: either the synthetic parent, a directory, or a file.
#[derive(Debug, Clone)]
struct Entry {
    /// Full path this entry points at (parent dir for "..").
    path: PathBuf,
    /// Display label, e.g. "..", "subdir/", "file.txt".
    label: String,
    is_dir: bool,
    is_parent: bool,
}

/// Modal file-browser state.
#[derive(Debug)]
pub struct FilePicker {
    current_dir: PathBuf,
    entries: Vec<Entry>,
    selected: usize,
}

impl FilePicker {
    /// Creates a picker rooted at the user's home directory (cwd as fallback).
    #[must_use]
    pub fn new() -> Self {
        let start = dirs::home_dir()
            .or_else(|| std::env::current_dir().ok())
            .unwrap_or_else(|| PathBuf::from("."));
        Self::with_dir(start)
    }

    /// Creates a picker rooted at a specific directory (used by tests).
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

    /// Rebuilds `entries` from `current_dir`. An unreadable dir yields an empty
    /// listing (plus the parent entry) rather than an error.
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
                    is_parent: false,
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
                is_parent: true,
            });
        }
        entries.extend(dirs);
        entries.extend(files);

        self.entries = entries;
        self.selected = 0;
    }

    /// Moves the selection up one (clamped).
    pub fn select_previous(&mut self) {
        self.selected = self.selected.saturating_sub(1);
    }

    /// Moves the selection down one (clamped).
    pub fn select_next(&mut self) {
        if !self.entries.is_empty() {
            self.selected = (self.selected + 1).min(self.entries.len() - 1);
        }
    }

    /// Activates the current entry: descend/ascend a directory, or select a file.
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

    /// Entry labels (test/inspection helper).
    #[must_use]
    pub fn entry_names(&self) -> Vec<String> {
        self.entries.iter().map(|e| e.label.clone()).collect()
    }

    /// Current selection index (test helper).
    #[must_use]
    pub const fn selected_index(&self) -> usize {
        self.selected
    }

    /// Renders the picker as a centered overlay.
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

        let list = List::default()
            .block(block)
            .items(items)
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
```

If `List::default().items(...)` / `.highlight_style(...)` does not match the installed ratatui API, mirror exactly how `chat_list.rs` builds its `List` (check that file for the correct constructor and stateful-render call). The render method is not unit-tested, so get it compiling and visually correct against the existing list code.

- [ ] **Step 4: Register the module**

In `src/ui/components/mod.rs`, add `mod file_picker;` with the other `mod` lines and export it:

```rust
pub use file_picker::{FilePicker, FilePickerAction};
```

- [ ] **Step 5: Run the navigation tests**

Run: `cargo test file_picker`
Expected: 4 tests pass.

- [ ] **Step 6: Verify clippy + compile**

Run: `cargo clippy --all-targets 2>&1 | tail -5`
Expected: no new warnings.

- [ ] **Step 7: Commit**

```bash
cargo fmt
git add src/ui/components/file_picker.rs src/ui/components/mod.rs
git commit -m "Add FilePicker modal component with directory navigation"
```

---

## Task 3: Stage attachments in ConversationModel

**Files:**
- Modify: `src/ui/components/conversation.rs`
- Test: `src/ui/components/conversation.rs` (existing `#[cfg(test)]` module)

- [ ] **Step 1: Write failing tests**

Add to the test module at the bottom of `conversation.rs`:

```rust
    #[test]
    fn submit_with_attachment_emits_attachment_action() {
        use std::path::PathBuf;
        let mut model = ConversationModel::new();
        model.input.set_focused(true);
        model.set_pending_attachment(PathBuf::from("/tmp/cat.png"));
        model.input.set_value("look");
        let action = model.handle_action(Action::SendMessage);
        assert_eq!(
            action,
            Some(ConversationAction::SendMessageWithAttachment(
                "look".to_string(),
                PathBuf::from("/tmp/cat.png"),
                None
            ))
        );
        assert!(model.pending_attachment().is_none(), "cleared after send");
    }

    #[test]
    fn submit_attachment_with_empty_caption_still_sends() {
        use std::path::PathBuf;
        let mut model = ConversationModel::new();
        model.input.set_focused(true);
        model.set_pending_attachment(PathBuf::from("/tmp/cat.png"));
        // no text typed
        let action = model.handle_action(Action::SendMessage);
        assert!(matches!(
            action,
            Some(ConversationAction::SendMessageWithAttachment(_, _, _))
        ));
    }

    #[test]
    fn submit_without_attachment_emits_plain_send() {
        let mut model = ConversationModel::new();
        model.input.set_focused(true);
        model.input.set_value("hi");
        let action = model.handle_action(Action::SendMessage);
        assert_eq!(
            action,
            Some(ConversationAction::SendMessage("hi".to_string(), None))
        );
    }

    #[test]
    fn esc_clears_pending_attachment_first() {
        use std::path::PathBuf;
        let mut model = ConversationModel::new();
        model.input.set_focused(true);
        model.reply_to = Some(5);
        model.set_pending_attachment(PathBuf::from("/tmp/cat.png"));
        model.handle_action(Action::CancelAction);
        assert!(model.pending_attachment().is_none());
        assert_eq!(model.reply_to, Some(5), "reply preserved on first Esc");
    }
```

Note: `set_value`/`set_focused` on `InputComponent` are already public (used elsewhere in this module). If `input` field access from the test needs it, it is already `pub`.

- [ ] **Step 2: Run tests to confirm they fail**

Run: `cargo test -p ithil --lib conversation 2>&1 | tail -20`
Expected: compile errors — `set_pending_attachment` / `pending_attachment` / `SendMessageWithAttachment` not found.

- [ ] **Step 3: Add the field**

In the `ConversationModel` struct (after `input_mode: InputMode,`):

```rust
    /// Path of a file staged to send with the next message, if any.
    pub pending_attachment: Option<std::path::PathBuf>,
```

Initialize it in `ConversationModel::new()` (in the `Self { ... }` literal):

```rust
            pending_attachment: None,
```

- [ ] **Step 4: Add accessors**

Add these methods inside `impl ConversationModel` (near `clear_action_state`):

```rust
    /// Stages a file to be sent with the next message.
    pub fn set_pending_attachment(&mut self, path: std::path::PathBuf) {
        self.pending_attachment = Some(path);
    }

    /// Returns the staged attachment path, if any.
    #[must_use]
    pub fn pending_attachment(&self) -> Option<&std::path::PathBuf> {
        self.pending_attachment.as_ref()
    }
```

- [ ] **Step 5: Add the action variant**

In `enum ConversationAction`, add:

```rust
    /// Send a message with a file attachment (caption, file path, optional reply_to)
    SendMessageWithAttachment(String, std::path::PathBuf, Option<i64>),
```

- [ ] **Step 6: Branch `submit_input`**

Replace the body of `submit_input` (currently lines ~292-307) with:

```rust
    fn submit_input(&mut self) -> Option<ConversationAction> {
        let text = self.input.value().trim().to_string();

        // With an attachment, an empty caption is allowed. Without one, keep the
        // existing guard that drops empty submissions.
        if text.is_empty() && self.pending_attachment.is_none() {
            return None;
        }

        let action = if let Some(path) = self.pending_attachment.take() {
            ConversationAction::SendMessageWithAttachment(text, path, self.reply_to)
        } else if let Some(edit_id) = self.editing {
            ConversationAction::EditMessage(edit_id, text)
        } else {
            ConversationAction::SendMessage(text, self.reply_to)
        };

        self.input.clear();
        self.clear_action_state();
        Some(action)
    }
```

- [ ] **Step 7: Make Esc clear the attachment first**

In `handle_input_action`, change the `Action::CancelAction` arm so a staged attachment is cleared before unfocusing/clearing other state:

```rust
            Action::CancelAction => {
                if self.pending_attachment.take().is_some() {
                    return None;
                }
                self.input.set_focused(false);
                self.clear_action_state();
                None
            },
```

(Leave the non-input `Action::CancelAction` arm in `handle_action` as-is.)

- [ ] **Step 8: Render the banner**

In `render_input` (the `ConversationWidget` impl), the input occupies a 3-line area. When an attachment is pending, draw a one-line banner at the top of that area and shrink the input box below it. Replace the start of `render_input` so it splits the area when needed:

```rust
    fn render_input(&self, area: Rect, buf: &mut Buffer) {
        // Reserve a banner line for a staged attachment.
        let (banner_area, area) = if let Some(path) = self.model.pending_attachment.as_ref() {
            let rows = Layout::default()
                .direction(Direction::Vertical)
                .constraints([Constraint::Length(1), Constraint::Min(2)])
                .split(area);
            let name = path
                .file_name()
                .map_or_else(|| path.display().to_string(), |n| n.to_string_lossy().into_owned());
            let banner = Paragraph::new(Line::from(vec![
                Span::styled(format!("📎 {name}"), Styles::text_accent()),
                Span::styled("  Esc to remove", Styles::text_muted()),
            ]));
            banner.render(rows[0], buf);
            (Some(rows[0]), rows[1])
        } else {
            (None, area)
        };
        let _ = banner_area;

        // ... existing render_input body continues unchanged, using `area` ...
```

Imports: `Constraint`, `Direction`, `Layout`, `Paragraph`, and `Span` are already imported in `conversation.rs`. **`Line` is NOT imported** — add it to the `text::{...}` group of the existing `use ratatui::{...}` block (i.e. `text::{Line, Span}`), or the banner's `Paragraph::new(Line::from(vec![...]))` will not compile.

- [ ] **Step 9: Run tests**

Run: `cargo test -p ithil --lib conversation`
Expected: the 4 new tests pass, existing conversation tests still pass.

- [ ] **Step 10: Verify clippy + compile**

Run: `cargo clippy --all-targets 2>&1 | tail -5`
Expected: no new warnings.

- [ ] **Step 11: Commit**

```bash
cargo fmt
git add src/ui/components/conversation.rs
git commit -m "Stage attachments in ConversationModel and branch submit_input"
```

---

## Task 4: AttachFile action + keybinding

**Files:**
- Modify: `src/ui/keys.rs`

- [ ] **Step 1: Add the `Action` variant**

In `enum Action`, under the "Conversation Actions" section (after `OpenMedia`):

```rust
    /// Open the file picker to attach a file to the message
    AttachFile,
```

- [ ] **Step 2: Add the `Display` arm**

In `impl std::fmt::Display for Action`, alongside the others:

```rust
            Self::AttachFile => write!(f, "Attach File"),
```

- [ ] **Step 3: Bind `Ctrl+T` in both modes**

In `KeyMap::new`, in the common bindings (so it works in vim and standard modes), add near the other `Ctrl+*` global bindings:

```rust
        bindings.insert(key(KeyCode::Char('t'), ctrl()), Action::AttachFile);
```

Place it with the global bindings block (around line 232) so both `add_vim_bindings` and `add_standard_bindings` inherit it. `Ctrl+T` is currently unbound.

- [ ] **Step 4: Add to help text (if applicable)**

Check `get_help_text` in `keys.rs`. If it enumerates bindings manually, add an entry like `("Ctrl+T", "Attach file")`. If it derives from the binding map, no change is needed.

- [ ] **Step 5: Verify compile + existing key tests**

Run: `cargo test -p ithil --lib keys && cargo clippy --all-targets 2>&1 | tail -5`
Expected: pass, no new warnings.

- [ ] **Step 6: Commit**

```bash
cargo fmt
git add src/ui/keys.rs
git commit -m "Add AttachFile action bound to Ctrl+T"
```

---

## Task 5: Wire the picker and send path into App

**Files:**
- Modify: `src/ui/app.rs`

This task is integration glue (no new unit tests; verified by compile + the manual checklist in Task 6).

- [ ] **Step 1: Add the field + import**

Add to the `App` struct (after `status_bar: StatusBar,` or near the other UI models):

```rust
    /// Active file picker overlay, when attaching a file.
    file_picker: Option<crate::ui::components::FilePicker>,
```

Initialize it in `App::new`'s `Self { ... }` literal:

```rust
            file_picker: None,
```

Ensure `FilePicker` / `FilePickerAction` are reachable — add to the existing components `use` in `app.rs` (e.g. `use crate::ui::components::{... , FilePicker, FilePickerAction};`).

- [ ] **Step 2: Add the AppAction variant**

In `enum AppAction`, after `SendMessage`:

```rust
    /// Send a message with a file attachment (chat_id, caption, file path, optional reply_to)
    SendMessageWithAttachment(i64, String, std::path::PathBuf, Option<i64>),
```

- [ ] **Step 3: Route the conversation action**

In `handle_conversation_action`, add an arm:

```rust
            ConversationAction::SendMessageWithAttachment(text, path, reply_to) => {
                Some(AppAction::SendMessageWithAttachment(chat_id, text, path, reply_to))
            },
```

- [ ] **Step 4: Dispatch the app action**

In `handle_app_action`'s `match`, add:

```rust
            AppAction::SendMessageWithAttachment(chat_id, text, path, reply_to) => {
                self.handle_send_message_with_attachment(chat_id, text, path, reply_to)
                    .await;
            },
```

- [ ] **Step 5: Add the handler**

Add after `handle_send_message`:

```rust
    /// Handle sending a message with a file attachment.
    async fn handle_send_message_with_attachment(
        &mut self,
        chat_id: i64,
        text: String,
        path: std::path::PathBuf,
        reply_to: Option<i64>,
    ) {
        self.set_status_message("Uploading…".to_string());
        match self.telegram.send_file(chat_id, &text, &path, reply_to).await {
            Ok(message) => {
                self.conversation_model.add_message(message);
                self.status_message = None;
            },
            Err(e) => {
                self.set_status_message(format!("Failed to send file: {e}"));
            },
        }
    }
```

(Confirm `set_status_message` exists and takes `String` — it is used throughout `app.rs`. The status text shows before the await; the UI repaints on the next loop tick, matching how other status messages behave.)

- [ ] **Step 6: Intercept keys when the picker is open**

At the **top** of `handle_key` (before the auth check at line ~749), add:

```rust
        // File picker overlay captures all keys while open.
        if self.file_picker.is_some() {
            return self.handle_file_picker_key(key);
        }
```

Then add the helper method:

```rust
    /// Handle key events while the file picker overlay is open.
    fn handle_file_picker_key(&mut self, key: KeyEvent) -> Option<AppAction> {
        use crate::ui::keys::Action;
        let Some(picker) = self.file_picker.as_mut() else {
            return None;
        };
        match self.keymap.get_action(&key) {
            Some(Action::Up) => picker.select_previous(),
            Some(Action::Down) => picker.select_next(),
            Some(Action::CancelAction) => {
                self.file_picker = None;
            },
            Some(Action::OpenChat | Action::SendMessage) => {
                // Enter activates the current entry.
                match picker.activate() {
                    FilePickerAction::Selected(path) => {
                        self.conversation_model.set_pending_attachment(path);
                        self.file_picker = None;
                        // Focus the input so the user can type a caption.
                        self.conversation_model.input.set_focused(true);
                        self.focused_pane = FocusedPane::Input;
                    },
                    // `activate()` never returns Cancelled.
                    FilePickerAction::Cancelled | FilePickerAction::None => {},
                }
            },
            _ => {},
        }
        None
    }
```

- [ ] **Step 7: Open the picker on AttachFile**

`Action::AttachFile` must open the picker from both the Conversation-focused and Input-focused branches of `handle_key`. In the Conversation branch `match action { ... }` (around line 792), add an arm before the `_ =>` fallback:

```rust
                    Action::AttachFile => {
                        self.file_picker = Some(crate::ui::components::FilePicker::new());
                        return None;
                    },
```

In the Input branch `match action { ... }` (around line 836), add the same arm before its `_ => {}`:

```rust
                    Action::AttachFile => {
                        self.file_picker = Some(crate::ui::components::FilePicker::new());
                        return None;
                    },
```

- [ ] **Step 8: Render the overlay**

In `render`, after the help-overlay block (line ~1190-1192), add:

```rust
        // Render file picker overlay if open
        if let Some(picker) = &self.file_picker {
            picker.render(frame);
        }
```

- [ ] **Step 9: Update the Debug impl (optional)**

In `impl Debug for App`, optionally add `.field("file_picker_open", &self.file_picker.is_some())` to keep the debug output informative. Not required.

- [ ] **Step 10: Compile + clippy + full test suite**

Run: `cargo build 2>&1 | tail -15 && cargo clippy --all-targets 2>&1 | tail -8 && cargo test 2>&1 | tail -15`
Expected: builds, no new warnings, all tests pass.

- [ ] **Step 11: Commit**

```bash
cargo fmt
git add src/ui/app.rs
git commit -m "Wire file picker overlay and attachment send path into App"
```

---

## Task 6: Final verification

**Files:** none (verification only)

- [ ] **Step 1: Full check**

Run:
```bash
cargo fmt --check
cargo clippy --all-targets
cargo test
```
Expected: all clean / green. If `fmt --check` fails, run `cargo fmt` and amend the relevant commit.

- [ ] **Step 2: Manual end-to-end verification (real account)**

`send_file` is thin glue over the live grammers client and is not unit-tested, so verify by hand with a real Telegram login:

1. Run `cargo run`, log in, open a chat.
2. Press `Ctrl+T` → file picker opens at `$HOME`. Arrow keys move; Enter on a folder descends; `..` ascends.
3. Select an image (e.g. a `.png`). Picker closes; banner `📎 name  Esc to remove` shows above the input; input is focused.
4. Type a caption, press Enter → message sends; the image arrives in Telegram as a **photo** with the caption.
5. Repeat selecting a `.txt`/`.pdf` → arrives as a **document**.
6. Attach a file, press Enter with no caption → still sends.
7. Attach a file, press Esc → banner clears, no send; a second Esc unfocuses the input as before.
8. Reply to a message (`Ctrl+R`/`r`), then attach + send → arrives as a reply with the file.

Record the results in the PR description.

- [ ] **Step 3: (No commit)** — verification only.

---

## Notes / Out of Scope (per spec)

- Single file per message only — no albums/multi-select.
- No drag-and-drop, clipboard paste, upload progress bar, or voice recording.
- `.gif` is sent as a document by design.
- Upload blocks the event loop inline, matching the existing text-send path.
