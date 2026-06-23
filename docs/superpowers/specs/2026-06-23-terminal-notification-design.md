# Terminal notification on unfocused new message

**Date:** 2026-06-23
**Status:** Approved (design)

## Goal

When a new incoming message arrives and the ithil terminal is **not focused**, fire a
desktop notification (sender + message preview) using the **current terminal's** native
notification facility (OSC 9 escape sequence). No external desktop-notification dependency.

## Non-goals

- No platform-specific desktop library (notify-rust, libnotify, etc.). Notifications must
  travel through the terminal in use, per requirement.
- No notification grouping/coalescing — one notification per incoming message.
- No notification when ithil is focused.
- No new UI surface, modal, or keybinding.

## Background / current state

- `src/telegram/updates.rs::handle_update` has two arms that BOTH emit `UpdateType::NewMessage`:
  the incoming arm (`NewMessage(msg) if !msg.outgoing()`) and the outgoing-confirmation arm
  (`NewMessage(msg) if msg.outgoing()`). They collapse to the same `UpdateType`, so the UI-side
  `NewMessage` handler fires for the user's own sent messages too. The trigger MUST therefore
  guard on `!msg.is_outgoing` itself (`Message.is_outgoing: bool`, types/mod.rs:720) — incoming
  is NOT guaranteed by the time the update reaches the UI.
- The app consumes updates in two methods that both delegate to `App::handle_update`
  (`src/ui/app.rs`, a synchronous `pub fn` at ~line 1203): `process_updates` (async, ~1164) and
  `process_updates_sync` (~1150). The `UpdateType::NewMessage` match arm lives inside
  `handle_update` (~1207), not in the collectors. `selected_chat_id` holds the open chat.
- The live event loop is `App::run_async_with_connection`. It reads terminal events with
  `event::read()` and currently only matches `Event::Key`. Mouse capture is enabled in
  `main.rs`; **focus reporting is not**.
- `crossterm` 0.28 supports `event::EnableFocusChange` / `DisableFocusChange` and emits
  `Event::FocusGained` / `Event::FocusLost` on terminals that support DEC mode 1004.
- Config already has `NotificationConfig { enabled, sound, desktop, muted_chats: Vec<i64> }`
  (defaults `enabled=true, sound=true, desktop=false`). `Chat` has `is_muted: bool`.

## Design

### 1. Focus tracking

- `main.rs`: add `crossterm::event::EnableFocusChange` to the terminal-setup `execute!` block
  and `crossterm::event::DisableFocusChange` to the teardown block.
- `App`: add field `terminal_focused: bool`, initialized to `true`. Rationale for the default:
  terminals without focus reporting never emit `FocusLost`, so the value stays `true` and the
  app never produces spurious notifications. The feature is silently inert on unsupported
  terminals rather than misfiring.
- `run_async_with_connection`: the loop currently reads events with
  `if let Event::Key(key) = event::read()?`. Convert that to a `match event::read()?` so focus
  events can be handled alongside keys:
  - `Event::FocusGained` → `self.terminal_focused = true`
  - `Event::FocusLost`  → `self.terminal_focused = false`
  - `Event::Key(key)` → existing key handling.
  - Neither focus arm requires a redraw.

### 2. Notification emit — `src/utils/notify.rs` (new module)

- `pub fn send_notification(text: &str, sound: bool)`:
  - Build the escape string `\x1b]9;{sanitized}\x07`. If `sound`, append a BEL (`\x07`).
  - Note: ithil renders in the alternate screen, so this escape interleaves with Ratatui's
    draw cycle on the same stdout. OSC 9 is a self-contained, cursor-neutral sequence; emitting
    it between frames is safe and does not corrupt the rendered UI. The manual iTerm2/kitty test
    is the regression guard for this interleave.
  - **Sanitize** `text` before writing (security: message content is untrusted Telegram input
    going straight to the terminal):
    - Strip ESC (`\x1b`), BEL (`\x07`), and other C0 control characters.
    - Collapse newlines/tabs to spaces.
    - Truncate to a fixed max (120 chars), appending `…` if truncated.
  - Best-effort write to stdout via `crossterm::execute!` / direct write; ignore errors.
- `fn sanitize(text: &str) -> String` — pure, unit-tested.

### 3. "Should notify" predicate — pure, testable

- `fn should_notify(focused: bool, cfg: &NotificationConfig, chat_id: i64, chat_muted: bool) -> bool`:
  - returns `!focused && cfg.enabled && cfg.desktop && !chat_muted
    && !cfg.muted_chats.contains(&chat_id)`.
- Unit-tested across the focus × enabled × desktop × muted matrix without a terminal.

### 4. Trigger wiring

- In `App::handle_update` (covers both the async and sync collector paths), inside the
  `UpdateType::NewMessage` arm where `msg` is already destructured: fire only when
  `!msg.is_outgoing` AND `should_notify(...)` is true, then format `"{sender}: {preview}"` and
  call `send_notification(text, cfg.sound)`.
  - `!msg.is_outgoing` guard is REQUIRED — the same `UpdateType::NewMessage` also carries the
    user's own outgoing-message confirmations (see Background).
  - Sender: `Message` has no name field — resolve via cache. Prefer
    `cache.get_user(msg.sender_id)` (first/last name) for the human sender; fall back to
    `cache.get_chat(chat_id).title` (e.g. group/channel name) when no user is found.
  - Preview: message text truncated using existing `appearance.message_preview_length`.
  - `chat_muted` derived from the cached `Chat.is_muted` for `chat_id`.

### 5. Config change

- Flip `NotificationConfig` default `desktop: false → true` so the feature is active out of the
  box. Existing user configs with an explicit `desktop: false` are still respected.
- `enabled && desktop` gate the OSC 9 emission; `sound` controls the appended BEL.
- No bell when focused — bell rides along with the notification, which only fires when unfocused.

## Error handling

- All terminal writes are best-effort; failures are ignored (a missed notification must never
  crash or stall the update loop).
- Unsupported terminals: OSC 9 / focus events are simply absent — no error path needed.

## Testing

- `notify::sanitize` — strips ESC/BEL/control chars, collapses whitespace, truncates with `…`.
  Includes an injection case (message containing `\x1b]9;evil\x07`).
- `should_notify` — truth table over focused, enabled, desktop, chat_muted, muted_chats.
- Manual: run in iTerm2/kitty, background the terminal, send a message from another device,
  confirm a notification appears with sender + preview; confirm none appears while focused.
