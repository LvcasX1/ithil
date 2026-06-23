//! Media operations for the Telegram client.
//!
//! This module provides methods for downloading and managing media files:
//! - Downloading photos
//! - Downloading documents (future)
//! - Opening media files with system viewer

use std::path::{Path, PathBuf};

use tokio::fs;
use tracing::{debug, info};

use super::client::TelegramClient;
use super::error::TelegramError;
use crate::types::Message;

/// Builds a deterministic local filename for a downloaded media item.
///
/// Photos keep their historical `photo_<chat>_<msg>.jpg` name. Documents keep
/// their original filename (sanitised, prefixed with the message id to avoid
/// collisions); other media fall back to a generic name with an extension
/// guessed from the MIME type.
fn media_file_name(chat_id: i64, message_id: i64, media: &grammers_client::media::Media) -> String {
    use grammers_client::media::Media as GMedia;

    match media {
        GMedia::Photo(_) => format!("photo_{chat_id}_{message_id}.jpg"),
        GMedia::Document(doc) => document_file_name(chat_id, message_id, doc.name(), doc.mime_type()),
        _ => format!("media_{chat_id}_{message_id}.bin"),
    }
}

/// Pure filename logic for document attachments (testable without a client).
fn document_file_name(
    chat_id: i64,
    message_id: i64,
    name: Option<&str>,
    mime: Option<&str>,
) -> String {
    if let Some(name) = name.map(str::trim).filter(|n| !n.is_empty()) {
        format!("{chat_id}_{message_id}_{}", sanitize_filename(name))
    } else {
        let ext = mime.and_then(ext_from_mime).unwrap_or("bin");
        format!("file_{chat_id}_{message_id}.{ext}")
    }
}

/// Replaces path separators and control characters so a server-supplied
/// filename can't escape the download directory.
fn sanitize_filename(name: &str) -> String {
    name.chars()
        .map(|c| {
            if c == '/' || c == '\\' || c.is_control() {
                '_'
            } else {
                c
            }
        })
        .collect()
}

/// Maps a MIME type to a sensible file extension, ignoring any `; charset=...`
/// suffix. Returns `None` for unknown types.
fn ext_from_mime(mime: &str) -> Option<&'static str> {
    let mime = mime.split(';').next().unwrap_or(mime).trim();
    Some(match mime {
        "image/jpeg" => "jpg",
        "image/png" => "png",
        "image/gif" => "gif",
        "image/webp" => "webp",
        "video/mp4" => "mp4",
        "video/quicktime" => "mov",
        "video/webm" => "webm",
        "audio/mpeg" => "mp3",
        "audio/ogg" | "audio/opus" => "ogg",
        "audio/mp4" | "audio/x-m4a" => "m4a",
        "application/pdf" => "pdf",
        "application/zip" => "zip",
        "text/plain" => "txt",
        _ => return None,
    })
}

impl TelegramClient {
    /// Downloads the media from a message to the specified directory.
    ///
    /// Handles any attachment type (photo, document, video, audio, ...). The
    /// file is saved with a deterministic, viewer-friendly name and an existing
    /// local copy is reused. Returns the path to the downloaded file.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat containing the message
    /// * `message_id` - ID of the message with the media
    /// * `download_dir` - Directory to save the file
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The client is not connected or authorized
    /// - The message is not found
    /// - The message doesn't contain any media
    /// - The download fails
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # use std::path::Path;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// let path = client.download_media(
    ///     123456789,
    ///     42,
    ///     Path::new("/tmp/media"),
    /// ).await?;
    /// println!("Media saved to: {}", path.display());
    /// # Ok(())
    /// # }
    /// ```
    pub async fn download_media(
        &self,
        chat_id: i64,
        message_id: i64,
        download_dir: &Path,
    ) -> Result<PathBuf, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        debug!(
            "Downloading media from message {} in chat {}",
            message_id, chat_id
        );

        // Fetch the specific message. Re-fetching yields a fresh file reference
        // (stored references expire), so we always download from the live message.
        #[allow(clippy::cast_possible_truncation)]
        let message_id_i32 = message_id as i32;

        let mut iter = client.iter_messages(peer_ref);
        iter = iter.offset_id(message_id_i32 + 1).limit(1);

        let msg = iter
            .next()
            .await
            .map_err(TelegramError::from)?
            .ok_or(TelegramError::MessageNotFound(message_id))?;

        // Verify this is the message we want
        if i64::from(msg.id()) != message_id {
            return Err(TelegramError::MessageNotFound(message_id));
        }

        // Get the media (any type: photo, document, video, audio, ...)
        let media = msg.media().ok_or(TelegramError::NoMedia(message_id))?;

        // Ensure download directory exists
        fs::create_dir_all(download_dir)
            .await
            .map_err(|e| TelegramError::Io(e.to_string()))?;

        // Derive a deterministic, viewer-friendly filename from the media.
        let filename = media_file_name(chat_id, message_id, &media);
        let file_path = download_dir.join(&filename);

        // Reuse an already-downloaded file instead of fetching it again.
        if file_path.exists() {
            debug!(
                "Media for message {} already exists at {}",
                message_id,
                file_path.display()
            );
            return Ok(file_path);
        }

        // Download the media
        client
            .download_media(&media, &file_path)
            .await
            .map_err(TelegramError::from)?;

        info!(
            "Downloaded media from message {} to {}",
            message_id,
            file_path.display()
        );

        Ok(file_path)
    }

    /// Downloads the attachment of a message, reusing a local copy if present.
    ///
    /// Works for any attachment type (photo, document, video, audio, voice,
    /// animation, sticker). Returns the path to the downloaded file.
    ///
    /// # Arguments
    ///
    /// * `message` - The message containing the attachment
    /// * `download_dir` - Directory to save the file
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The message has no downloadable attachment
    /// - The download fails
    pub async fn download_media_if_needed(
        &self,
        message: &Message,
        download_dir: &Path,
    ) -> Result<PathBuf, TelegramError> {
        if !message.content.content_type.is_downloadable() {
            return Err(TelegramError::NoMedia(message.id));
        }

        self.download_media(message.chat_id, message.id, download_dir)
            .await
    }

    /// Opens a media file with the system's default application.
    ///
    /// On macOS, this uses `open`. On Linux, it uses `xdg-open`.
    /// On Windows, it uses `start`.
    ///
    /// # Arguments
    ///
    /// * `path` - Path to the media file to open
    ///
    /// # Errors
    ///
    /// Returns an error if the file doesn't exist or the open command fails.
    #[allow(clippy::unused_async)]
    pub async fn open_media_file(path: &Path) -> Result<(), TelegramError> {
        if !path.exists() {
            return Err(TelegramError::FileNotFound(path.to_path_buf()));
        }

        info!("Opening media file: {}", path.display());

        #[cfg(target_os = "macos")]
        let result = tokio::process::Command::new("open").arg(path).spawn();

        #[cfg(target_os = "linux")]
        let result = tokio::process::Command::new("xdg-open").arg(path).spawn();

        #[cfg(target_os = "windows")]
        let result = tokio::process::Command::new("cmd")
            .args(["/C", "start", "", path.to_str().unwrap_or("")])
            .spawn();

        result.map_err(|e| TelegramError::Io(e.to_string()))?;

        Ok(())
    }

    /// Opens a URL with the system's default browser.
    ///
    /// On macOS this uses `open`, on Linux `xdg-open`, and on Windows `start`.
    ///
    /// # Errors
    ///
    /// Returns an error if the open command cannot be spawned.
    #[allow(clippy::unused_async)]
    pub async fn open_url(url: &str) -> Result<(), TelegramError> {
        info!("Opening URL: {}", url);

        #[cfg(target_os = "macos")]
        let result = tokio::process::Command::new("open").arg(url).spawn();

        #[cfg(target_os = "linux")]
        let result = tokio::process::Command::new("xdg-open").arg(url).spawn();

        #[cfg(target_os = "windows")]
        let result = tokio::process::Command::new("cmd")
            .args(["/C", "start", "", url])
            .spawn();

        result.map_err(|e| TelegramError::Io(e.to_string()))?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::{document_file_name, ext_from_mime, sanitize_filename};

    #[test]
    fn test_document_keeps_original_name() {
        assert_eq!(
            document_file_name(123, 42, Some("report.pdf"), Some("application/pdf")),
            "123_42_report.pdf"
        );
    }

    #[test]
    fn test_document_without_name_uses_mime_extension() {
        assert_eq!(
            document_file_name(123, 42, None, Some("video/mp4")),
            "file_123_42.mp4"
        );
    }

    #[test]
    fn test_document_without_name_or_known_mime_falls_back_to_bin() {
        assert_eq!(
            document_file_name(123, 42, None, Some("application/x-weird")),
            "file_123_42.bin"
        );
        assert_eq!(document_file_name(123, 42, None, None), "file_123_42.bin");
    }

    #[test]
    fn test_blank_name_is_treated_as_missing() {
        assert_eq!(
            document_file_name(123, 42, Some("   "), Some("text/plain")),
            "file_123_42.txt"
        );
    }

    #[test]
    fn test_sanitize_strips_path_separators() {
        assert_eq!(sanitize_filename("../../etc/passwd"), ".._.._etc_passwd");
        assert_eq!(sanitize_filename("a\\b/c"), "a_b_c");
        assert_eq!(sanitize_filename("clean.txt"), "clean.txt");
    }

    #[test]
    fn test_ext_from_mime_ignores_charset_suffix() {
        assert_eq!(ext_from_mime("text/plain; charset=utf-8"), Some("txt"));
        assert_eq!(ext_from_mime("image/png"), Some("png"));
        assert_eq!(ext_from_mime("application/octet-stream"), None);
    }
}
