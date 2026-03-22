//! Chat and dialog operations for the Telegram client.
//!
//! This module provides methods for working with chats (dialogs):
//! - Fetching all dialogs
//! - Searching chats
//! - Pinning/unpinning chats
//! - Muting/unmuting chats
//! - Archiving/unarchiving chats
//! - Marking chats as read

use grammers_client::peer::{Dialog, Peer as GrammersPeer};
use grammers_client::tl;
use grammers_session::types::PeerRef;
use tracing::{debug, info};

use super::client::TelegramClient;
use super::error::TelegramError;
use crate::types::{Chat, ChatType, Message, UserStatus};

impl TelegramClient {
    /// Fetches all dialogs (chats) from Telegram.
    ///
    /// This retrieves all chats the user has, including private chats,
    /// groups, supergroups, and channels. Results are cached automatically.
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected or not authorized.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// let chats = client.get_dialogs().await?;
    /// for chat in &chats {
    ///     println!("{}: {}", chat.title, chat.unread_count);
    /// }
    /// # Ok(())
    /// # }
    /// ```
    pub async fn get_dialogs(&self) -> Result<Vec<Chat>, TelegramError> {
        let client = self.require_authorized().await?;

        info!("Fetching dialogs...");

        let mut dialogs = client.iter_dialogs();
        let mut result = Vec::new();

        while let Some(dialog) = dialogs.next().await.map_err(TelegramError::from)? {
            // Cache the peer as a user if it's a private chat
            if let Some(user) = grammers_peer_to_user(&dialog.peer()) {
                self.cache().set_user(user);
            }

            let chat = dialog_to_chat(&dialog);

            // Cache the chat
            self.cache().set_chat(chat.clone());

            result.push(chat);
        }

        info!("Fetched {} dialogs", result.len());
        Ok(result)
    }

    /// Searches chats by query string.
    ///
    /// This searches through the user's dialogs for chats matching the query.
    /// The search is performed locally on cached data if available.
    ///
    /// # Arguments
    ///
    /// * `query` - Search query string
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected or not authorized.
    pub async fn search_chats(&self, query: &str) -> Result<Vec<Chat>, TelegramError> {
        // First, ensure we have dialogs cached
        let dialogs = self.get_dialogs().await?;

        let query_lower = query.to_lowercase();

        let results: Vec<Chat> = dialogs
            .into_iter()
            .filter(|chat| {
                chat.title.to_lowercase().contains(&query_lower)
                    || chat.username.to_lowercase().contains(&query_lower)
            })
            .collect();

        debug!("Search for '{}' found {} results", query, results.len());
        Ok(results)
    }

    /// Pins or unpins a chat.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to pin/unpin
    /// * `pin` - `true` to pin, `false` to unpin
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// or the chat is not found.
    pub async fn pin_chat(&self, chat_id: i64, pin: bool) -> Result<(), TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!(
            "{} chat {}",
            if pin { "Pinning" } else { "Unpinning" },
            chat_id
        );

        client
            .invoke(&tl::functions::messages::ToggleDialogPin {
                pinned: pin,
                peer: tl::enums::InputDialogPeer::Peer(tl::types::InputDialogPeer {
                    peer: tl::enums::InputPeer::from(peer_ref),
                }),
            })
            .await
            .map_err(TelegramError::from)?;

        // Update cache
        if let Some(mut chat) = self.cache().get_chat(chat_id) {
            chat.is_pinned = pin;
            self.cache().set_chat(chat);
        }

        Ok(())
    }

    /// Mutes or unmutes a chat.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to mute/unmute
    /// * `mute` - `true` to mute, `false` to unmute
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// or the chat is not found.
    pub async fn mute_chat(&self, chat_id: i64, mute: bool) -> Result<(), TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!(
            "{} chat {}",
            if mute { "Muting" } else { "Unmuting" },
            chat_id
        );

        // Mute for a very long time (1 year in seconds) or unmute
        let mute_until = if mute { i32::MAX } else { 0 };

        client
            .invoke(&tl::functions::account::UpdateNotifySettings {
                peer: tl::enums::InputNotifyPeer::Peer(tl::types::InputNotifyPeer {
                    peer: tl::enums::InputPeer::from(peer_ref),
                }),
                settings: tl::enums::InputPeerNotifySettings::Settings(
                    tl::types::InputPeerNotifySettings {
                        show_previews: None,
                        silent: Some(mute),
                        mute_until: Some(mute_until),
                        sound: None,
                        stories_muted: None,
                        stories_hide_sender: None,
                        stories_sound: None,
                    },
                ),
            })
            .await
            .map_err(TelegramError::from)?;

        // Update cache
        if let Some(mut chat) = self.cache().get_chat(chat_id) {
            chat.is_muted = mute;
            self.cache().set_chat(chat);
        }

        Ok(())
    }

    /// Archives or unarchives a chat.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to archive/unarchive
    /// * `archive` - `true` to archive, `false` to unarchive
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// or the chat is not found.
    pub async fn archive_chat(&self, chat_id: i64, archive: bool) -> Result<(), TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!(
            "{} chat {}",
            if archive { "Archiving" } else { "Unarchiving" },
            chat_id
        );

        // Folder ID 1 is the Archive folder
        let folder_id = if archive { 1 } else { 0 };

        client
            .invoke(&tl::functions::folders::EditPeerFolders {
                folder_peers: vec![tl::types::InputFolderPeer {
                    peer: tl::enums::InputPeer::from(peer_ref),
                    folder_id,
                }
                .into()],
            })
            .await
            .map_err(TelegramError::from)?;

        Ok(())
    }

    /// Marks all messages in a chat as read.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to mark as read
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// or the chat is not found.
    pub async fn mark_as_read(&self, chat_id: i64) -> Result<(), TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        debug!("Marking chat {} as read", chat_id);

        // Use the high-level mark_as_read method
        client
            .mark_as_read(peer_ref)
            .await
            .map_err(TelegramError::from)?;

        // Update cache
        if let Some(mut chat) = self.cache().get_chat(chat_id) {
            chat.unread_count = 0;
            self.cache().set_chat(chat);
        }

        Ok(())
    }

    /// Resolves a chat ID to a `PeerRef` for API calls.
    ///
    /// First checks the session cache, then fetches dialogs if not found.
    pub(crate) async fn get_peer_ref(&self, chat_id: i64) -> Result<PeerRef, TelegramError> {
        let client = self.client().await?;

        // Try to resolve from the session
        // We need to construct a PeerId from the chat_id
        // This is tricky because we don't know the peer type from just the ID
        // We'll try to find it in dialogs

        // Fetch dialogs to populate the session cache
        let mut dialogs = client.iter_dialogs();
        while let Some(dialog) = dialogs.next().await.map_err(TelegramError::from)? {
            let peer = dialog.peer();
            if peer.id().bare_id() == chat_id {
                if let Some(peer_ref) = peer.to_ref().await {
                    return Ok(peer_ref);
                }
            }
        }

        Err(TelegramError::ChatNotFound(chat_id))
    }
}

/// Converts a grammers Dialog to our Chat type.
fn dialog_to_chat(dialog: &Dialog) -> Chat {
    let peer = dialog.peer();
    let last_message = dialog
        .last_message
        .as_ref()
        .map(grammers_message_to_message);

    // Extract dialog-specific info from raw
    let (unread_count, is_pinned, draft_message) = extract_dialog_info(&dialog.raw);

    // Get peer_ref for access_hash
    let peer_ref = dialog.peer_ref();
    let access_hash = peer_ref.auth.hash();

    Chat {
        id: peer.id().bare_id(),
        chat_type: grammers_peer_type(&peer),
        title: peer.name().unwrap_or("").to_string(),
        username: peer.username().map(ToString::to_string).unwrap_or_default(),
        photo_id: String::new(), // Photo handling requires additional work
        last_message: last_message.map(Box::new),
        unread_count,
        is_pinned,
        pin_order: 0,
        is_muted: false, // Would need to check notification settings
        draft_message,
        last_read_inbox_id: 0,
        last_read_outbox_id: 0,
        access_hash,
        user_status: UserStatus::Offline,
        notification_settings: None,
        has_new_message: false,
    }
}

/// Extracts dialog-specific information from raw dialog data.
fn extract_dialog_info(raw: &tl::enums::Dialog) -> (i32, bool, String) {
    match raw {
        tl::enums::Dialog::Dialog(d) => {
            let draft = d
                .draft
                .as_ref()
                .and_then(|draft| {
                    if let tl::enums::DraftMessage::Message(m) = draft {
                        Some(m.message.clone())
                    } else {
                        None
                    }
                })
                .unwrap_or_default();

            (d.unread_count, d.pinned, draft)
        },
        tl::enums::Dialog::Folder(_) => (0, false, String::new()),
    }
}

/// Maps grammers Peer to our ChatType.
fn grammers_peer_type(peer: &GrammersPeer) -> ChatType {
    use grammers_session::types::ChannelKind;

    match peer {
        GrammersPeer::User(_) => ChatType::Private,
        GrammersPeer::Group(_) => ChatType::Group,
        GrammersPeer::Channel(c) => match c.kind() {
            Some(ChannelKind::Megagroup) | Some(ChannelKind::Gigagroup) => ChatType::Supergroup,
            Some(ChannelKind::Broadcast) | None => ChatType::Channel,
        },
    }
}

/// Converts a grammers Message to our Message type.
pub(crate) fn grammers_message_to_message(msg: &grammers_client::message::Message) -> Message {
    use crate::types::{DownloadStatus, Media, MessageContent, MessageType, PhotoSize};

    let sender_id = msg.sender().map_or(0, |s| s.id().bare_id());
    let chat_id = msg.peer_id().bare_id();

    // Determine message type, text/caption, and media based on media presence
    let (content_type, text, caption, media) = if let Some(grammers_media) = msg.media() {
        // Has media - determine type and extract metadata
        match grammers_media {
            grammers_client::media::Media::Photo(ref photo) => {
                // Extract photo metadata from grammers PhotoSize variants
                let photo_sizes: Vec<PhotoSize> = photo
                    .thumbs()
                    .iter()
                    .filter_map(|thumb| {
                        // Extract dimensions from PhotoSize variants that have them
                        match thumb {
                            grammers_client::media::PhotoSize::Size(size) => Some(PhotoSize {
                                size_type: thumb.photo_type(),
                                width: size.width,
                                height: size.height,
                                size: size.size,
                            }),
                            grammers_client::media::PhotoSize::Progressive(prog) => {
                                Some(PhotoSize {
                                    size_type: thumb.photo_type(),
                                    width: prog.width,
                                    height: prog.height,
                                    // Use the largest size from progressive sizes
                                    size: prog.sizes.iter().max().copied().unwrap_or(0),
                                })
                            },
                            grammers_client::media::PhotoSize::Cached(cached) => Some(PhotoSize {
                                size_type: thumb.photo_type(),
                                width: cached.width,
                                height: cached.height,
                                #[allow(clippy::cast_possible_truncation)]
                                size: cached.bytes.len() as i32,
                            }),
                            // Skip variants without dimension info
                            _ => None,
                        }
                    })
                    .collect();

                // Find the largest photo size for dimensions
                let (width, height) = photo_sizes
                    .iter()
                    .max_by_key(|s| s.width * s.height)
                    .map(|s| (s.width, s.height))
                    .unwrap_or((0, 0));

                // Calculate total size from largest photo
                let size = photo.size().map(|s| s as i64).unwrap_or_else(|| {
                    photo_sizes
                        .iter()
                        .max_by_key(|s| s.size)
                        .map(|s| i64::from(s.size))
                        .unwrap_or(0)
                });

                let media = Media {
                    id: photo.id().to_string(),
                    width,
                    height,
                    duration: 0,
                    size,
                    mime_type: "image/jpeg".to_string(),
                    thumbnail: None,
                    local_path: String::new(),
                    remote_path: String::new(),
                    is_downloaded: false,
                    access_hash: 0, // Not needed - grammers handles download internally
                    file_reference: Vec::new(),
                    photo_sizes,
                    download_status: DownloadStatus::NotDownloaded,
                    download_progress: None,
                };

                (
                    MessageType::Photo,
                    String::new(),
                    msg.text().to_string(),
                    Some(Box::new(media)),
                )
            },
            grammers_client::media::Media::Document(ref doc) => {
                // Check document MIME type and attributes for specific types
                let mime = doc.mime_type().unwrap_or("");
                let content_type = if mime.starts_with("audio/ogg") || mime == "audio/opus" {
                    // Voice messages are typically ogg/opus
                    MessageType::Voice
                } else if mime.starts_with("video/") {
                    MessageType::Video
                } else if mime.starts_with("audio/") {
                    MessageType::Audio
                } else {
                    MessageType::Document
                };
                (content_type, msg.text().to_string(), String::new(), None)
            },
            grammers_client::media::Media::Sticker(_) => (
                MessageType::Sticker,
                msg.text().to_string(),
                String::new(),
                None,
            ),
            grammers_client::media::Media::Contact(_) => (
                MessageType::Contact,
                msg.text().to_string(),
                String::new(),
                None,
            ),
            grammers_client::media::Media::Geo(_) => (
                MessageType::Location,
                msg.text().to_string(),
                String::new(),
                None,
            ),
            grammers_client::media::Media::GeoLive(_) => (
                MessageType::Location,
                msg.text().to_string(),
                String::new(),
                None,
            ),
            grammers_client::media::Media::Venue(_) => (
                MessageType::Venue,
                msg.text().to_string(),
                String::new(),
                None,
            ),
            grammers_client::media::Media::Poll(_) => (
                MessageType::Poll,
                msg.text().to_string(),
                String::new(),
                None,
            ),
            // Game media type doesn't exist in grammers 0.9
            _ => (
                MessageType::Document,
                msg.text().to_string(),
                String::new(),
                None,
            ),
        }
    } else {
        (
            MessageType::Text,
            msg.text().to_string(),
            String::new(),
            None,
        )
    };

    // Use the public date() method which returns DateTime<Utc>
    let date = msg.date();

    // edit_date() already returns Option<DateTime<Utc>>
    let edit_date = msg.edit_date();

    Message {
        id: i64::from(msg.id()),
        chat_id,
        sender_id,
        content: MessageContent {
            content_type,
            text,
            caption,
            entities: Vec::new(), // Would need to convert entities
            media,
            location: None,
            contact: None,
            poll: None,
            sticker: None,
            animation: None,
            document: None,
        },
        date,
        edit_date,
        is_outgoing: msg.outgoing(),
        is_channel_post: msg.post(),
        is_pinned: msg.pinned(),
        is_edited: edit_date.is_some(),
        is_forwarded: msg.forward_header().is_some(),
        reply_to_message_id: msg.reply_to_message_id().map(i64::from).unwrap_or(0),
        forward_info: None, // Would need to convert forward info
        views: msg.view_count().unwrap_or(0),
        media_album_id: msg.grouped_id().unwrap_or(0),
    }
}

/// Extracts a User from a grammers Peer if it's a user type.
///
/// Returns None for groups/channels since they aren't users.
pub(crate) fn grammers_peer_to_user(peer: &GrammersPeer) -> Option<crate::types::User> {
    match peer {
        GrammersPeer::User(user) => {
            // Use first_name and last_name from grammers User
            let first_name = user.first_name().unwrap_or("").to_string();
            let last_name = user.last_name().unwrap_or("").to_string();

            Some(crate::types::User {
                id: user.id().bare_id(),
                first_name,
                last_name,
                username: user.username().map(ToString::to_string).unwrap_or_default(),
                phone_number: String::new(), // Not available from peer
                profile_photo_id: String::new(),
                status: UserStatus::Offline, // Would need separate query
                is_bot: user.is_bot(),
                is_contact: false, // Not available from peer
                is_mutual_contact: false,
                is_verified: user.verified(),
                is_premium: user.is_premium(),
            })
        },
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_chat_type_mapping() {
        // We can't easily test without mocking grammers types,
        // but we can verify our ChatType enum works correctly
        assert_eq!(format!("{}", ChatType::Private), "Private");
        assert_eq!(format!("{}", ChatType::Group), "Group");
        assert_eq!(format!("{}", ChatType::Supergroup), "Supergroup");
        assert_eq!(format!("{}", ChatType::Channel), "Channel");
    }
}
