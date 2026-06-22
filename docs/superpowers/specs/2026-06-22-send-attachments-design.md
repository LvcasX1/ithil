# Send Messages with Attachments

**Issue:** #4 — "Attachments in Ithil". Users can view received attachments but cannot send them.
**Date:** 2026-06-22

## Goal

Let the user attach a file from the local filesystem to an outgoing message: pick a file
through a browser modal, stage it with an optional caption, then send it as a Telegram photo
(for images) or a document (for everything else).

## Background

The codebase already receives and renders media. The domain types (`Media`, `Document`,
`MessageContent`, `MessageType`) fully model attachments, and `telegram/media.rs` downloads
them. The send path is the only gap: `telegram/messages.rs::send_message` builds
`InputMessage::new().text(text)` and never attaches a file.

grammers-client 0.9 supports the full upload path:

- `Client::upload_file<P: AsRef<Path>>(path) -> Result<Uploaded, io::Error>`
- `InputMessage::photo(Uploaded)`, `::document(Uploaded)`, `::file(Uploaded)`,
  plus `::mime_type`, `::attribute`
- The message text doubles as the media caption.

So the work is plumbing: pick a file, upload it, attach it to an `InputMessage`, send.

## User flow

```
Ctrl+T (conversation focused)
  -> App opens FilePicker overlay
  -> arrow keys browse, Enter descends a dir or selects a file, Esc cancels
  -> selected path -> ConversationModel.pending_attachment
  -> banner "[attach] name (size)  Esc to remove" renders above the input
  -> user types an optional caption, presses Enter
  -> ConversationAction::SendMessageWithAttachment(text, path, reply_to)
  -> AppAction::SendMessageWithAttachment(chat_id, text, path, reply_to)
  -> telegram::send_file(...) uploads and sends
```

## Components

### FilePicker (new) — `src/ui/components/file_picker.rs`

A modal file browser, rendered as an overlay (mirrors `HelpModal`).

State:
- `current_dir: PathBuf` — starts at `$HOME` (fallback: current working dir)
- `entries: Vec<PathBuf>` — directories and files in `current_dir`, sorted (dirs first), with a
  synthetic `..` parent entry when not at filesystem root
- `selected: usize` — highlighted index

Behavior:
- Up / Down move the selection (clamped)
- Enter on a directory descends into it and reloads `entries`
- Enter on `..` ascends to the parent
- Enter on a file returns that `PathBuf` to the caller and closes the modal
- Esc closes the modal without selecting

The directory-listing and navigation logic is separated from rendering so it can be unit-tested
against a temporary directory tree. Hidden files (dotfiles) are listed; no filtering in v1.
Unreadable directories produce an empty list rather than an error/panic.

### App overlay — `src/ui/app.rs`

- Add `file_picker: Option<FilePicker>` to `App`, following the existing modal-overlay pattern.
- When `Some`, the picker intercepts key events and renders on top of the main UI.
- `Action::AttachFile` (when the conversation/input is focused) sets `file_picker = Some(..)`.
- When the picker returns a file, store it in the conversation model's `pending_attachment` and
  close the picker.
- New `AppAction::SendMessageWithAttachment(chat_id, text, path, reply_to)` routes to a new
  `handle_send_message_with_attachment` method, which calls `telegram::send_file` and reuses the
  existing `Ok`/`Err` status-bar handling from `handle_send_message`.

### ConversationModel — `src/ui/components/conversation.rs`

- Add `pending_attachment: Option<PathBuf>`.
- `submit_input()` branches: if `pending_attachment.is_some()`, emit
  `ConversationAction::SendMessageWithAttachment(text, path, reply_to)`; otherwise emit the
  existing `SendMessage(text, reply_to)`. Clear the pending attachment after submit.
- Esc precedence: if a `pending_attachment` is set, Esc clears it first (before clearing reply or
  edit state).
- Render a one-line banner above the input when an attachment is pending:
  `[attach] <file_name> (<human_size>)  Esc to remove`.

### telegram::send_file — `src/telegram/messages.rs`

```rust
pub async fn send_file(
    &self,
    chat_id: i64,
    text: &str,
    path: &Path,
    reply_to: Option<i64>,
) -> Result<Message, TelegramError> {
    let uploaded = client.upload_file(path).await?;
    let mut msg = InputMessage::new().text(text);
    msg = if is_image(path) { msg.photo(uploaded) } else { msg.document(uploaded) };
    if let Some(r) = reply_to {
        msg = msg.reply_to(Some(r as i32));
    }
    let sent = client.send_message(peer_ref, msg).await?;
    // map to domain Message exactly as send_message does
}
```

`upload_file` returns `io::Error`; map it into `TelegramError` (add a variant or reuse an
existing IO/wrapped variant as the codebase already does for other IO).

### is_image helper

```rust
fn is_image(path: &Path) -> bool {
    matches!(
        path.extension().and_then(|e| e.to_str()).map(str::to_ascii_lowercase).as_deref(),
        Some("jpg" | "jpeg" | "png" | "webp" | "bmp")
    )
}
```

Images send as photos; everything else (including `.gif`, videos, audio, docs) sends as a
document. Pure function — unit-tested directly.

### Keymap — `src/ui/keys.rs`

- Add `Action::AttachFile`.
- Bind `Ctrl+T` to `Action::AttachFile` in both standard and vim binding sets (`Ctrl+T` is
  currently unbound; `Ctrl+f/r/e/o/s/p` are taken).

## Error handling

- `upload_file` / `send_message` errors flow through the same `match` arm as `handle_send_message`,
  surfacing in the status bar. No new error UI.
- The browser only yields existing regular files, so missing-path / is-a-directory cases do not
  arise from normal use; `send_file` still returns an error (not a panic) if the path is bad.
- Upload is awaited inline, matching the current `send_message` behavior. While in flight the
  status bar shows `Uploading...`. A progress bar is out of scope (see below).

## Testing

- **Unit — `is_image`**: image extensions (mixed case) -> true; `.gif`, `.mp4`, `.txt`, no
  extension -> false.
- **Unit — FilePicker navigation**: over a temp directory tree, assert descending into a dir,
  ascending via `..`, selection clamping, and that selecting a file returns its path.
- **Unit — ConversationModel**: pending-attachment lifecycle; `submit_input` emits
  `SendMessageWithAttachment` when an attachment is staged and `SendMessage` otherwise; Esc clears
  the pending attachment before reply/edit state.
- **Manual**: `send_file` is thin glue over the live grammers client; verify end-to-end with a
  real Telegram account (send an image -> arrives as photo; send a `.txt`/`.pdf` -> arrives as a
  document; caption and reply preserved).

## Scope

**In v1:** single file per message; auto photo-vs-document by extension; caption; reply with
attachment; Esc to cancel a staged attachment; `$HOME`-rooted file browser.

**Out (YAGNI):** albums / multi-select, drag-and-drop, clipboard paste, upload progress bar,
voice-note recording, sending stickers/animations as their native types.

## Decisions

- `.gif` sends as a document (Telegram treats GIFs as animations; documents are the safe, simple
  path) rather than a photo.
- Upload blocks the event loop inline, like the current text-send path; a non-blocking upload with
  a progress indicator is deferred.
