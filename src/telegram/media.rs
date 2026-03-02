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

impl TelegramClient {
    /// Downloads a photo from a message to the specified directory.
    ///
    /// The photo is saved with a filename based on the photo ID.
    /// Returns the path to the downloaded file.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat containing the message
    /// * `message_id` - ID of the message with the photo
    /// * `download_dir` - Directory to save the photo
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The client is not connected or authorized
    /// - The message is not found
    /// - The message doesn't contain a photo
    /// - The download fails
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # use std::path::Path;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// let path = client.download_photo(
    ///     123456789,
    ///     42,
    ///     Path::new("/tmp/media"),
    /// ).await?;
    /// println!("Photo saved to: {}", path.display());
    /// # Ok(())
    /// # }
    /// ```
    pub async fn download_photo(
        &self,
        chat_id: i64,
        message_id: i64,
        download_dir: &Path,
    ) -> Result<PathBuf, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        debug!(
            "Downloading photo from message {} in chat {}",
            message_id, chat_id
        );

        // Fetch the specific message
        #[allow(clippy::cast_possible_truncation)]
        let message_id_i32 = message_id as i32;

        let mut iter = client.iter_messages(peer_ref);
        iter = iter.offset_id(message_id_i32 + 1).limit(1);

        let msg = iter
            .next()
            .await
            .map_err(TelegramError::from)?
            .ok_or_else(|| TelegramError::MessageNotFound(message_id))?;

        // Verify this is the message we want
        if i64::from(msg.id()) != message_id {
            return Err(TelegramError::MessageNotFound(message_id));
        }

        // Get the media
        let media = msg
            .media()
            .ok_or_else(|| TelegramError::NoMedia(message_id))?;

        // Verify it's a photo
        if !matches!(media, grammers_client::media::Media::Photo(_)) {
            return Err(TelegramError::NotAPhoto(message_id));
        }

        // Ensure download directory exists
        fs::create_dir_all(download_dir)
            .await
            .map_err(|e| TelegramError::Io(e.to_string()))?;

        // Generate filename based on photo ID and chat
        let filename = format!("photo_{}_{}.jpg", chat_id, message_id);
        let file_path = download_dir.join(&filename);

        // Download the media
        client
            .download_media(&media, &file_path)
            .await
            .map_err(TelegramError::from)?;

        info!(
            "Downloaded photo from message {} to {}",
            message_id,
            file_path.display()
        );

        Ok(file_path)
    }

    /// Downloads a photo and returns the path, checking if it's already downloaded.
    ///
    /// If the photo is already downloaded at the expected path, returns that path
    /// without re-downloading.
    ///
    /// # Arguments
    ///
    /// * `message` - The message containing the photo
    /// * `download_dir` - Directory to save the photo
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The message doesn't contain a photo
    /// - The download fails
    pub async fn download_photo_if_needed(
        &self,
        message: &Message,
        download_dir: &Path,
    ) -> Result<PathBuf, TelegramError> {
        use crate::types::MessageType;

        // Verify it's a photo message
        if message.content.content_type != MessageType::Photo {
            return Err(TelegramError::NotAPhoto(message.id));
        }

        // Check if already downloaded
        let filename = format!("photo_{}_{}.jpg", message.chat_id, message.id);
        let file_path = download_dir.join(&filename);

        if file_path.exists() {
            debug!(
                "Photo for message {} already exists at {}",
                message.id,
                file_path.display()
            );
            return Ok(file_path);
        }

        // Download the photo
        self.download_photo(message.chat_id, message.id, download_dir)
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
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_photo_filename_format() {
        let chat_id: i64 = 123456789;
        let message_id: i64 = 42;
        let filename = format!("photo_{}_{}.jpg", chat_id, message_id);
        assert_eq!(filename, "photo_123456789_42.jpg");
    }
}
