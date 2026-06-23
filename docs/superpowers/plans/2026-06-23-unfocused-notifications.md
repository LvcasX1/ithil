# Terminal Notification on Unfocused New Message — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fire a terminal-native OSC 9 desktop notification (sender + message preview) when a new incoming Telegram message arrives while the ithil terminal is not focused.

**Architecture:** Track terminal focus via crossterm focus-reporting events (`FocusGained`/`FocusLost`) held on `App.terminal_focused`. On each incoming `UpdateType::NewMessage` in `App::handle_update`, a pure `should_notify` predicate gates emission; a new `src/utils/notify.rs` module writes a sanitized OSC 9 escape to stdout. Notification text reuses a shared `MessageContent::preview()` helper.

**Tech Stack:** Rust, crossterm 0.28, ratatui, tokio, grammers. Existing `NotificationConfig` and `Chat.is_muted` are reused.

**Spec:** `docs/superpowers/specs/2026-06-23-terminal-notification-design.md`

**Branch:** `feat/unfocused-notifications` (already checked out)

---

## File Structure

| File | Responsibility | Action |
|---|---|---|
| `src/types/mod.rs` | Add `MessageContent::preview()` — body text for a message, no sender prefix | Modify |
| `src/ui/components/chat_item.rs` | Refactor `get_preview_text` to reuse `MessageContent::preview()` (DRY) | Modify |
| `src/utils/notify.rs` | OSC 9 emission + `sanitize` + `should_notify` predicate | Create |
| `src/utils/mod.rs` | Register `notify` module, re-export `send_notification`, `should_notify` | Modify |
| `src/app/config.rs` | Flip `NotificationConfig` default `desktop: false → true` | Modify |
| `src/main.rs` | Enable/disable terminal focus reporting | Modify |
| `src/ui/app.rs` | `terminal_focused` field; focus event arms; notification trigger in `handle_update` | Modify |

---

## Task 1: Shared message-preview helper

DRY: `chat_item::get_preview_text` hard-codes the per-`MessageType` body text. Extract the body (without the "You: " prefix) into `MessageContent::preview()` so the notifier and chat list share one source of truth.

**Files:**
- Modify: `src/types/mod.rs` (impl block for `MessageContent`, near its definition ~line 679)
- Modify: `src/ui/components/chat_item.rs:269-330` (`get_preview_text`)

- [ ] **Step 1: Write the failing test**

Add to the `#[cfg(test)]` module in `src/types/mod.rs` (create one if absent):

```rust
#[test]
fn message_content_preview_text() {
    let mut c = MessageContent::default();
    c.content_type = MessageType::Text;
    c.text = "hello world".to_string();
    assert_eq!(c.preview(), "hello world");
}

#[test]
fn message_content_preview_photo_with_caption() {
    let mut c = MessageContent::default();
    c.content_type = MessageType::Photo;
    c.caption = "beach".to_string();
    assert_eq!(c.preview(), "📷 Photo: beach");
}

#[test]
fn message_content_preview_voice_no_caption() {
    let mut c = MessageContent::default();
    c.content_type = MessageType::Voice;
    assert_eq!(c.preview(), "🎤 Voice message");
}
```

> If `MessageContent` has no `Default`, construct it explicitly with all fields instead of `..Default::default()`. Check the struct first.

- [ ] **Step 2: Run test, verify it fails**

Run: `cargo test message_content_preview -- --nocapture`
Expected: FAIL — `no method named preview found for struct MessageContent`.

- [ ] **Step 3: Implement `MessageContent::preview()`**

Add an `impl MessageContent` block in `src/types/mod.rs`. Port the match body from `chat_item::get_preview_text` verbatim (no "You: " prefix):

```rust
impl MessageContent {
    /// Human-readable one-line preview of this message's body (no sender prefix).
    #[must_use]
    pub fn preview(&self) -> String {
        let mut preview = String::new();
        match self.content_type {
            MessageType::Text => preview.push_str(&self.text),
            MessageType::Photo => {
                preview.push_str("📷 Photo");
                if !self.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&self.caption);
                }
            },
            MessageType::Video => {
                preview.push_str("🎬 Video");
                if !self.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&self.caption);
                }
            },
            MessageType::Voice => preview.push_str("🎤 Voice message"),
            MessageType::VideoNote => preview.push_str("📹 Video message"),
            MessageType::Audio => preview.push_str("🎵 Audio"),
            MessageType::Document => {
                preview.push_str("📎 Document");
                if let Some(ref doc) = self.document {
                    if !doc.file_name.is_empty() {
                        preview.push_str(": ");
                        preview.push_str(&doc.file_name);
                    }
                }
                if !self.caption.is_empty() {
                    preview.push_str(": ");
                    preview.push_str(&self.caption);
                }
            },
            MessageType::Sticker => preview.push_str("🎨 Sticker"),
            MessageType::Animation => preview.push_str("GIF"),
            MessageType::Location => preview.push_str("📍 Location"),
            MessageType::Contact => preview.push_str("👤 Contact"),
            MessageType::Poll => {
                preview.push_str("📊 Poll");
                if let Some(ref poll) = self.poll {
                    preview.push_str(": ");
                    preview.push_str(&poll.question);
                }
            },
            MessageType::Venue => preview.push_str("📍 Venue"),
            MessageType::Game => preview.push_str("🎮 Game"),
        }
        preview
    }
}
```

- [ ] **Step 4: Run test, verify pass**

Run: `cargo test message_content_preview`
Expected: PASS.

- [ ] **Step 5: Refactor `chat_item::get_preview_text` to reuse it**

Replace the whole `match msg.content.content_type { … }` block in `get_preview_text` (src/ui/components/chat_item.rs) with:

```rust
preview.push_str(&msg.content.preview());
```

Keep the `"You: "` prefix logic that precedes it untouched.

- [ ] **Step 6: Verify chat_item still builds and its tests pass**

Run: `cargo test --lib chat_item && cargo build`
Expected: PASS / builds clean.

- [ ] **Step 7: Commit**

```bash
git add src/types/mod.rs src/ui/components/chat_item.rs
git commit -m "refactor: extract MessageContent::preview() shared by chat list and notifier"
```

---

## Task 2: `sanitize` — escape-injection guard

Telegram message text is untrusted and goes straight to the terminal. Strip control bytes / escape sequences and truncate.

**Files:**
- Create: `src/utils/notify.rs`
- Modify: `src/utils/mod.rs`

- [ ] **Step 1: Create the module with a failing test**

Create `src/utils/notify.rs`:

```rust
//! Terminal-native desktop notifications via the OSC 9 escape sequence.

/// Max characters in a notification body before truncation.
const MAX_LEN: usize = 120;

/// Strip control/escape bytes and collapse whitespace so untrusted message
/// content cannot inject terminal escape sequences. Truncates to `MAX_LEN`.
#[must_use]
pub fn sanitize(text: &str) -> String {
    let mut out = String::with_capacity(text.len().min(MAX_LEN));
    for ch in text.chars() {
        // Drop C0 controls (incl. ESC 0x1b and BEL 0x07) and DEL.
        if ch.is_control() {
            // Represent line breaks / tabs as a single space for readability.
            if matches!(ch, '\n' | '\r' | '\t') {
                if !out.ends_with(' ') {
                    out.push(' ');
                }
            }
            continue;
        }
        out.push(ch);
    }
    let trimmed = out.trim();
    if trimmed.chars().count() > MAX_LEN {
        let truncated: String = trimmed.chars().take(MAX_LEN - 1).collect();
        format!("{truncated}…")
    } else {
        trimmed.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn strips_escape_injection() {
        // A crafted OSC 9 payload must not survive.
        let evil = "hi\x1b]9;you are pwned\x07 there";
        let clean = sanitize(evil);
        assert!(!clean.contains('\x1b'));
        assert!(!clean.contains('\x07'));
        assert_eq!(clean, "hi]9;you are pwned there");
    }

    #[test]
    fn collapses_newlines_to_space() {
        assert_eq!(sanitize("a\n\nb"), "a b");
    }

    #[test]
    fn truncates_long_text() {
        let long = "x".repeat(200);
        let s = sanitize(&long);
        assert_eq!(s.chars().count(), 120);
        assert!(s.ends_with('…'));
    }

    #[test]
    fn passes_plain_text_through() {
        assert_eq!(sanitize("Alice: hey there"), "Alice: hey there");
    }
}
```

- [ ] **Step 2: Register the module**

In `src/utils/mod.rs` add `mod notify;` and extend the re-export line (add the items as Task 3/notify functions land — for now export sanitize is internal, so just register the module):

```rust
mod notify;

pub use notify::{send_notification, should_notify};
```

> `send_notification` and `should_notify` are added in Tasks 3-4; if this line fails to compile now, temporarily export only what exists and finalize in Task 4. Simpler: defer adding this `pub use` line until Task 4 and add only `mod notify;` here.

So for this step, add ONLY:

```rust
mod notify;
```

- [ ] **Step 3: Run test, verify fail then pass**

Run: `cargo test --lib notify::tests::strips_escape_injection`
Expected: PASS (function is implemented above). Run the whole `cargo test --lib notify` group; all four sanitize tests PASS.

- [ ] **Step 4: Commit**

```bash
git add src/utils/notify.rs src/utils/mod.rs
git commit -m "feat(notify): add sanitize for terminal escape-injection guard"
```

---

## Task 3: `should_notify` predicate

Pure, terminal-free gate. Tested as a truth table.

**Files:**
- Modify: `src/utils/notify.rs`

- [ ] **Step 1: Add failing tests**

Append to the `tests` module in `src/utils/notify.rs`:

```rust
use crate::app::config::NotificationConfig;

fn cfg(enabled: bool, desktop: bool, muted: Vec<i64>) -> NotificationConfig {
    NotificationConfig { enabled, sound: true, desktop, muted_chats: muted }
}

#[test]
fn notifies_when_unfocused_enabled_desktop_unmuted() {
    assert!(should_notify(false, &cfg(true, true, vec![]), 42, false));
}

#[test]
fn no_notify_when_focused() {
    assert!(!should_notify(true, &cfg(true, true, vec![]), 42, false));
}

#[test]
fn no_notify_when_disabled() {
    assert!(!should_notify(false, &cfg(false, true, vec![]), 42, false));
}

#[test]
fn no_notify_when_desktop_off() {
    assert!(!should_notify(false, &cfg(true, false, vec![]), 42, false));
}

#[test]
fn no_notify_when_chat_muted_flag() {
    assert!(!should_notify(false, &cfg(true, true, vec![]), 42, true));
}

#[test]
fn no_notify_when_chat_in_muted_list() {
    assert!(!should_notify(false, &cfg(true, true, vec![42]), 42, false));
}
```

> Confirm `NotificationConfig` field names/order against `src/app/config.rs` and adjust the struct literal if needed. If `NotificationConfig` is not constructible from this module's path, import via its actual path (e.g. `crate::app::config::NotificationConfig` or `crate::Config`-relative).

- [ ] **Step 2: Run test, verify it fails**

Run: `cargo test --lib notify::tests::notifies_when`
Expected: FAIL — `cannot find function should_notify`.

- [ ] **Step 3: Implement `should_notify`**

Add to `src/utils/notify.rs` (above the tests module):

```rust
use crate::app::config::NotificationConfig;

/// Decide whether an incoming message should raise a notification.
/// Pure — no terminal or I/O. The caller is responsible for the
/// `!msg.is_outgoing` check (see plan Task 5).
#[must_use]
pub fn should_notify(
    focused: bool,
    cfg: &NotificationConfig,
    chat_id: i64,
    chat_muted: bool,
) -> bool {
    !focused
        && cfg.enabled
        && cfg.desktop
        && !chat_muted
        && !cfg.muted_chats.contains(&chat_id)
}
```

> Ensure `NotificationConfig` is `pub` and reachable. If `config` module is private, make the type path correct (it is used as `config::NotificationConfig` in `config.rs`; expose as needed).

- [ ] **Step 4: Run test, verify pass**

Run: `cargo test --lib notify`
Expected: all sanitize + should_notify tests PASS.

- [ ] **Step 5: Commit**

```bash
git add src/utils/notify.rs
git commit -m "feat(notify): add should_notify focus/config/mute gate"
```

---

## Task 4: `send_notification` — OSC 9 emit

**Files:**
- Modify: `src/utils/notify.rs`, `src/utils/mod.rs`

- [ ] **Step 1: Implement `send_notification`**

Add to `src/utils/notify.rs`:

```rust
use std::io::Write;

/// Emit an OSC 9 desktop notification through the current terminal.
/// `text` is sanitized first. When `sound` is true a BEL is appended so
/// terminals that map it to an alert will also chime. Best-effort: any
/// I/O error is swallowed (a missed notification must never disrupt the UI).
pub fn send_notification(text: &str, sound: bool) {
    let body = sanitize(text);
    if body.is_empty() {
        return;
    }
    let mut seq = format!("\x1b]9;{body}\x07");
    if sound {
        seq.push('\x07');
    }
    let mut stdout = std::io::stdout();
    let _ = stdout.write_all(seq.as_bytes());
    let _ = stdout.flush();
}
```

- [ ] **Step 2: Export from utils**

In `src/utils/mod.rs`, replace the bare `mod notify;` with:

```rust
mod notify;

pub use notify::{send_notification, should_notify};
```

- [ ] **Step 3: Build**

Run: `cargo build`
Expected: builds clean (no dead-code warnings once Task 5 wires it; if `send_notification`/`should_notify` warn as unused here, that is acceptable until Task 5 — do NOT add `#[allow(dead_code)]`, the next task consumes them).

- [ ] **Step 4: Commit**

```bash
git add src/utils/notify.rs src/utils/mod.rs
git commit -m "feat(notify): add OSC 9 send_notification emitter"
```

---

## Task 5: Config default — `desktop: true`

**Files:**
- Modify: `src/app/config.rs:317-326` (`impl Default for NotificationConfig`)

- [ ] **Step 1: Add failing test**

Add to the test module in `src/app/config.rs` (create if absent):

```rust
#[test]
fn notification_desktop_defaults_on() {
    assert!(NotificationConfig::default().desktop);
}
```

- [ ] **Step 2: Run test, verify it fails**

Run: `cargo test --lib notification_desktop_defaults_on`
Expected: FAIL (current default is `false`).

- [ ] **Step 3: Flip the default**

In `impl Default for NotificationConfig`, change `desktop: false,` to `desktop: true,`.

- [ ] **Step 4: Run test, verify pass**

Run: `cargo test --lib notification_desktop_defaults_on`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/app/config.rs
git commit -m "feat(config): enable desktop notifications by default"
```

---

## Task 6: Terminal focus tracking

**Files:**
- Modify: `src/main.rs:119-124` (setup `execute!`), `src/main.rs:183-188` (teardown `execute!`)
- Modify: `src/ui/app.rs` — add `terminal_focused` field (struct ~line 160, init in `new` ~line 232/251); event loop in `run_async_with_connection` (~line 456)

- [ ] **Step 1: Enable focus reporting in main.rs**

In the setup `execute!` block add `crossterm::event::EnableFocusChange` after `EnableMouseCapture`:

```rust
crossterm::execute!(
    stdout,
    crossterm::terminal::EnterAlternateScreen,
    crossterm::event::EnableMouseCapture,
    crossterm::event::EnableFocusChange
)
.context("Failed to set up terminal")?;
```

In the teardown `execute!` block add `crossterm::event::DisableFocusChange`:

```rust
crossterm::execute!(
    terminal.backend_mut(),
    crossterm::terminal::LeaveAlternateScreen,
    crossterm::event::DisableMouseCapture,
    crossterm::event::DisableFocusChange
)
.context("Failed to restore terminal")?;
```

- [ ] **Step 2: Add `terminal_focused` field to `App`**

In the `App` struct (`src/ui/app.rs`) add:

```rust
/// Whether the terminal is currently focused. Starts true so terminals
/// without focus reporting never produce spurious notifications.
terminal_focused: bool,
```

Initialize it to `true` in `App::new` (the struct-literal return, ~line 251 alongside `selected_chat_id: None,`):

```rust
terminal_focused: true,
```

- [ ] **Step 3: Handle focus events in the live event loop**

In `run_async_with_connection`, the poll branch currently reads:

```rust
if let Event::Key(key) = event::read()? {
    // ... existing key handling ...
}
```

Convert to a `match`:

```rust
match event::read()? {
    Event::FocusGained => self.terminal_focused = true,
    Event::FocusLost => self.terminal_focused = false,
    Event::Key(key) => {
        // ... existing key handling, unchanged ...
    },
    _ => {}
}
```

> Preserve the exact existing key-handling body. If the existing arm already sits inside a broader `match`, just add the two focus arms. Check whether `run_async` (the non-connection loop, ~line 395) needs the same arms — if it is used in production paths, mirror the change; if it is test-only, leave it.

- [ ] **Step 4: Build**

Run: `cargo build`
Expected: builds clean.

- [ ] **Step 5: Commit**

```bash
git add src/main.rs src/ui/app.rs
git commit -m "feat(ui): track terminal focus via crossterm focus events"
```

---

## Task 7: Wire notification trigger into `handle_update`

**Files:**
- Modify: `src/ui/app.rs` — `handle_update` `UpdateType::NewMessage` arm (~line 1207)

- [ ] **Step 1: Add the trigger**

Inside the `UpdateType::NewMessage` arm, where `msg` is destructured (`let msg = *msg;`), after the message is added to cache and before/after the chat-list refresh, add:

```rust
if !msg.is_outgoing
    && crate::utils::should_notify(
        self.terminal_focused,
        &self.config.notifications,
        update.chat_id,
        self.cache
            .get_chat(update.chat_id)
            .is_some_and(|c| c.is_muted),
    )
{
    let sender = self
        .cache
        .get_user(msg.sender_id)
        .map(|u| {
            let name = format!("{} {}", u.first_name, u.last_name);
            name.trim().to_string()
        })
        .filter(|n| !n.is_empty())
        .or_else(|| self.cache.get_chat(update.chat_id).map(|c| c.title))
        .unwrap_or_else(|| "New message".to_string());

    let preview = msg.content.preview();
    let limit = self.config.ui.appearance.message_preview_length;
    let preview = crate::utils::truncate_string(&preview, limit);
    crate::utils::send_notification(
        &format!("{sender}: {preview}"),
        self.config.notifications.sound,
    );
}
```

> Verify `truncate_string` signature (`src/utils/formatting.rs`) — adjust the call if it takes `(usize, &str)` or returns differently. If `msg` was moved into `self.conversation_model.add_message(msg)` earlier in the arm, compute `sender`/`preview` and call `send_notification` BEFORE that move, or clone what you need. Order the code so no use-after-move occurs.

- [ ] **Step 2: Build + clippy**

Run: `cargo build && cargo clippy --all-targets -- -D warnings`
Expected: clean. Fix any borrow/move errors per the note above.

- [ ] **Step 3: Full test + format**

Run: `cargo fmt && cargo test`
Expected: all tests PASS, no diff left by fmt that breaks build.

- [ ] **Step 4: Commit**

```bash
git add src/ui/app.rs
git commit -m "feat(ui): notify on incoming message when terminal unfocused"
```

---

## Task 8: Final verification

- [ ] **Step 1: Full CI-equivalent gate**

Run:
```bash
cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test
```
Expected: all green.

- [ ] **Step 2: Manual smoke test**

In iTerm2 / kitty / WezTerm / Ghostty:
1. `cargo run`, log in, reach the chat list.
2. Switch focus to another window (background the terminal).
3. From another device, send yourself a message.
4. Confirm a desktop notification appears reading `Sender: preview`, and (if `sound`) an alert.
5. Bring ithil back to focus; send another message; confirm NO notification fires.
6. Set `notifications.desktop: false` in `~/.config/ithil/config.yaml`; confirm notifications stop.

> Document the terminal tested in the PR description. Terminals without OSC 9 / focus reporting simply produce nothing — not an error.

- [ ] **Step 3: Push branch**

```bash
git push -u origin feat/unfocused-notifications
```

---

## Notes for the implementer

- **DRY:** Task 1 is the only place message-preview text is defined after this change — do not re-implement the match anywhere.
- **Security:** `sanitize` is the trust boundary for terminal output. Never write raw message text to the terminal escape sequence — always through `send_notification`, which sanitizes.
- **YAGNI:** No coalescing, no per-message throttle, no async — `send_notification` is best-effort sync I/O on a single short string.
- **Best-effort:** notification I/O errors are intentionally ignored; never propagate them into the update loop.
