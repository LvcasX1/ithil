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
