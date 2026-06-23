//! Terminal-native desktop notifications via the OSC 9 escape sequence.

use std::io::Write;

use crate::app::NotificationConfig;

/// Max characters in a notification body before truncation.
const MAX_LEN: usize = 120;

/// Returns `true` for characters that must never reach the terminal/OS.
///
/// Covers control and escape bytes (incl. ESC 0x1b and BEL 0x07, plus C1 and
/// DEL via [`char::is_control`]) as well as Unicode bidi/format characters that
/// can spoof or reverse the displayed text (LRM/RLM, embeddings/overrides,
/// isolates, and the BOM / zero-width no-break space).
fn is_unwanted(ch: char) -> bool {
    ch.is_control()
        || matches!(ch,
            '\u{200E}' | '\u{200F}'           // LRM / RLM
            | '\u{202A}'..='\u{202E}'         // embeddings + overrides
            | '\u{2066}'..='\u{2069}'         // isolates
            | '\u{FEFF}')                     // BOM / ZWNBSP
}

/// Sanitize untrusted message content for safe display in a notification.
///
/// Strips control/escape bytes and Unicode bidi/format characters, and
/// collapses whitespace, so the content cannot inject terminal escape
/// sequences or spoof the displayed text via bidi overrides. Truncates to
/// `MAX_LEN`.
#[must_use]
pub fn sanitize(text: &str) -> String {
    let mut out = String::with_capacity(text.len().min(MAX_LEN));
    for ch in text.chars() {
        // Drop controls and bidi/format characters.
        if is_unwanted(ch) {
            // Represent line breaks / tabs as a single space for readability;
            // bidi/format chars are dropped outright with no replacement.
            if matches!(ch, '\n' | '\r' | '\t') && !out.ends_with(' ') {
                out.push(' ');
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

/// Decide whether an incoming message should raise a notification.
/// Pure — no terminal or I/O. The caller is responsible for the
/// `!msg.is_outgoing` check (see plan Task 5/Task 7).
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

/// Emit an OSC 9 desktop notification through the current terminal.
///
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

    #[test]
    fn strips_bidi_override() {
        let s = sanitize("Alice: \u{202E}reversed");
        assert!(!s.contains('\u{202E}'));
        assert_eq!(s, "Alice: reversed");
    }

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
}
