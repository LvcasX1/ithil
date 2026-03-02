//! Real-time update handling for the Telegram client.
//!
//! This module provides the update loop that streams Telegram events
//! to the UI via tokio channels. It handles:
//! - New messages
//! - Message edits
//! - Message deletions
//! - Chat updates
//! - User status changes

use std::ops::Deref;

use grammers_client::client::UpdateStream;
use grammers_client::update::Update as GrammersUpdate;
use tracing::{debug, error, info, trace, warn};

use super::chats::grammers_message_to_message;
use super::client::TelegramClient;
use super::error::TelegramError;
use crate::types::{Update, UpdateData, UpdateType};

impl TelegramClient {
    /// Starts the update loop.
    ///
    /// This method runs indefinitely, fetching updates from Telegram and
    /// sending them to the UI via the channel set with [`set_update_channel`](Self::set_update_channel).
    ///
    /// The loop will:
    /// 1. Create an update stream using `stream_updates()`
    /// 2. Fetch updates via `stream.next()`
    /// 3. Convert them to our `Update` type
    /// 4. Update the cache with new data
    /// 5. Send the update through the channel
    ///
    /// # Cancellation Safety
    ///
    /// This method is cancellation-safe. If the task is cancelled, the update
    /// loop will stop cleanly. Call [`stop_update_loop`](Self::stop_update_loop)
    /// to signal the loop to stop.
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected or a network error occurs.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # use tokio::sync::mpsc;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// let (tx, mut rx) = mpsc::channel(100);
    /// client.set_update_channel(tx).await;
    ///
    /// // Spawn update loop in a separate task
    /// let client_clone = client.clone();
    /// tokio::spawn(async move {
    ///     if let Err(e) = client_clone.run_update_loop().await {
    ///         eprintln!("Update loop error: {}", e);
    ///     }
    /// });
    ///
    /// // Process updates
    /// while let Some(update) = rx.recv().await {
    ///     println!("Got update: {:?}", update.update_type);
    /// }
    /// # Ok(())
    /// # }
    /// ```
    pub async fn run_update_loop(&self) -> Result<(), TelegramError> {
        if self.is_update_loop_running() {
            warn!("Update loop is already running");
            return Ok(());
        }

        // Create the update stream (this takes the updates_receiver)
        let mut stream: UpdateStream = self.create_update_stream().await?;

        self.set_update_loop_running(true);
        info!("Starting update loop");

        loop {
            // Check if we should stop
            if !self.is_update_loop_running() {
                info!("Update loop stopped");
                // Sync update state before exiting
                stream.sync_update_state().await;
                break;
            }

            // Fetch next update using stream.next()
            match stream.next().await {
                Ok(update) => {
                    if let Some(our_update) = self.handle_update(update).await {
                        // Send to UI
                        if let Some(tx) = self.get_update_sender().await {
                            if tx.send(our_update).await.is_err() {
                                warn!("Update channel closed, stopping update loop");
                                self.set_update_loop_running(false);
                                stream.sync_update_state().await;
                                break;
                            }
                        }
                    }
                }
                Err(e) => {
                    error!("Error fetching update: {}", e);

                    // Check if this is a recoverable error
                    let telegram_error = TelegramError::from(e);
                    if !telegram_error.is_recoverable() {
                        self.set_update_loop_running(false);
                        stream.sync_update_state().await;
                        return Err(telegram_error);
                    }

                    // For recoverable errors, wait a bit and continue
                    tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
                }
            }
        }

        Ok(())
    }

    /// Stops the update loop.
    ///
    /// This signals the update loop to stop at its next iteration.
    /// The loop may not stop immediately if it's waiting for an update.
    pub async fn stop_update_loop(&self) {
        if self.is_update_loop_running() {
            info!("Stopping update loop...");
            self.set_update_loop_running(false);
        }
    }

    /// Handles a single update from grammers.
    ///
    /// Converts the grammers update to our Update type and updates the cache.
    async fn handle_update(&self, update: GrammersUpdate) -> Option<Update> {
        match update {
            GrammersUpdate::NewMessage(msg) if !msg.outgoing() => {
                trace!("Received new message: {}", msg.id());

                // The update::Message derefs to message::Message
                let message = grammers_message_to_message(msg.deref());
                let chat_id = message.chat_id;

                // Update cache
                self.cache().add_message(chat_id, message.clone());

                // Update chat's has_new_message flag
                if let Some(mut chat) = self.cache().get_chat(chat_id) {
                    chat.has_new_message = true;
                    chat.last_message = Some(Box::new(message.clone()));
                    self.cache().set_chat(chat);
                }

                Some(Update {
                    update_type: UpdateType::NewMessage,
                    chat_id,
                    message: Some(Box::new(message)),
                    data: UpdateData::None,
                })
            }

            GrammersUpdate::NewMessage(msg) if msg.outgoing() => {
                trace!("Received outgoing message confirmation: {}", msg.id());

                let message = grammers_message_to_message(msg.deref());
                let chat_id = message.chat_id;

                // Update cache - the message might already be there from send_message
                self.cache().add_message(chat_id, message.clone());

                // Update chat's last message
                if let Some(mut chat) = self.cache().get_chat(chat_id) {
                    chat.last_message = Some(Box::new(message.clone()));
                    self.cache().set_chat(chat);
                }

                Some(Update {
                    update_type: UpdateType::NewMessage,
                    chat_id,
                    message: Some(Box::new(message)),
                    data: UpdateData::None,
                })
            }

            GrammersUpdate::MessageEdited(msg) => {
                trace!("Received message edit: {}", msg.id());

                let message = grammers_message_to_message(msg.deref());
                let chat_id = message.chat_id;

                // Update cache
                self.cache().update_message(chat_id, message.clone());

                Some(Update {
                    update_type: UpdateType::MessageEdited,
                    chat_id,
                    message: Some(Box::new(message)),
                    data: UpdateData::None,
                })
            }

            GrammersUpdate::MessageDeleted(deletion) => {
                debug!("Received message deletion");

                // Get the message IDs
                let message_ids = deletion.messages();
                let channel_id = deletion.channel_id();

                // We need to determine the chat_id
                // For channels/supergroups, we have the channel_id
                // For private chats, we don't have the chat_id in the deletion
                let chat_id = channel_id.unwrap_or(0);

                // Update cache - remove deleted messages
                for msg_id in message_ids {
                    self.cache().delete_message(chat_id, i64::from(*msg_id));
                }

                Some(Update {
                    update_type: UpdateType::MessageDeleted,
                    chat_id,
                    message: None,
                    data: UpdateData::None,
                })
            }

            GrammersUpdate::Raw(raw_update) => self.handle_raw_update(raw_update.raw).await,

            _ => {
                trace!("Ignoring unhandled update type");
                None
            }
        }
    }

    /// Handles raw updates that aren't directly supported by grammers' high-level API.
    async fn handle_raw_update(
        &self,
        update: grammers_client::tl::enums::Update,
    ) -> Option<Update> {
        use grammers_client::tl::enums::Update as TlUpdate;
        use grammers_client::tl::types;

        match update {
            TlUpdate::ReadHistoryInbox(types::UpdateReadHistoryInbox {
                peer,
                max_id,
                still_unread_count,
                ..
            }) => {
                let chat_id = peer_to_chat_id(&peer);
                debug!(
                    "Read inbox update for chat {}: max_id={}, unread={}",
                    chat_id, max_id, still_unread_count
                );

                // Update cache
                if let Some(mut chat) = self.cache().get_chat(chat_id) {
                    chat.last_read_inbox_id = i64::from(max_id);
                    chat.unread_count = still_unread_count;
                    self.cache().set_chat(chat);
                }

                Some(Update {
                    update_type: UpdateType::ChatReadInbox,
                    chat_id,
                    message: None,
                    data: UpdateData::Integer(i64::from(max_id)),
                })
            }

            TlUpdate::ReadHistoryOutbox(types::UpdateReadHistoryOutbox { peer, max_id, .. }) => {
                let chat_id = peer_to_chat_id(&peer);
                debug!("Read outbox update for chat {}: max_id={}", chat_id, max_id);

                // Update cache
                if let Some(mut chat) = self.cache().get_chat(chat_id) {
                    chat.last_read_outbox_id = i64::from(max_id);
                    self.cache().set_chat(chat);
                }

                Some(Update {
                    update_type: UpdateType::ChatReadOutbox,
                    chat_id,
                    message: None,
                    data: UpdateData::Integer(i64::from(max_id)),
                })
            }

            TlUpdate::ReadChannelInbox(types::UpdateReadChannelInbox {
                channel_id,
                max_id,
                still_unread_count,
                ..
            }) => {
                let chat_id = channel_id;
                debug!(
                    "Read channel inbox for {}: max_id={}, unread={}",
                    chat_id, max_id, still_unread_count
                );

                // Update cache
                if let Some(mut chat) = self.cache().get_chat(chat_id) {
                    chat.last_read_inbox_id = i64::from(max_id);
                    chat.unread_count = still_unread_count;
                    self.cache().set_chat(chat);
                }

                Some(Update {
                    update_type: UpdateType::ChatReadInbox,
                    chat_id,
                    message: None,
                    data: UpdateData::Integer(i64::from(max_id)),
                })
            }

            TlUpdate::UserStatus(types::UpdateUserStatus { user_id, status }) => {
                debug!("User {} status changed", user_id);

                let new_status = tl_status_to_user_status(&status);

                // Update user in cache if we have them
                if let Some(mut user) = self.cache().get_user(user_id) {
                    user.status = new_status;
                    self.cache().set_user(user);
                }

                Some(Update {
                    update_type: UpdateType::UserStatus,
                    chat_id: user_id,
                    message: None,
                    data: UpdateData::None,
                })
            }

            TlUpdate::ChatParticipants(_) => {
                debug!("Chat participants update");
                None // We don't track participants yet
            }

            TlUpdate::DraftMessage(types::UpdateDraftMessage { peer, draft, .. }) => {
                let chat_id = peer_to_chat_id(&peer);
                debug!("Draft message update for chat {}", chat_id);

                // Extract draft text
                let draft_text = match draft {
                    grammers_client::tl::enums::DraftMessage::Message(m) => m.message,
                    grammers_client::tl::enums::DraftMessage::Empty(_) => String::new(),
                };

                // Update cache
                if let Some(mut chat) = self.cache().get_chat(chat_id) {
                    chat.draft_message = draft_text.clone();
                    self.cache().set_chat(chat);
                }

                Some(Update {
                    update_type: UpdateType::ChatDraftMessage,
                    chat_id,
                    message: None,
                    data: UpdateData::String(draft_text),
                })
            }

            TlUpdate::PinnedDialogs(_) => {
                debug!("Pinned dialogs update");
                // Refresh dialogs to get the new order
                Some(Update {
                    update_type: UpdateType::ChatPosition,
                    chat_id: 0,
                    message: None,
                    data: UpdateData::None,
                })
            }

            _ => {
                trace!("Ignoring unhandled raw update");
                None
            }
        }
    }
}

/// Converts a TL Peer to a chat ID.
fn peer_to_chat_id(peer: &grammers_client::tl::enums::Peer) -> i64 {
    use grammers_client::tl::enums::Peer;

    match peer {
        Peer::User(u) => u.user_id,
        Peer::Chat(c) => c.chat_id,
        Peer::Channel(c) => c.channel_id,
    }
}

/// Converts a TL UserStatus to our UserStatus type.
fn tl_status_to_user_status(
    status: &grammers_client::tl::enums::UserStatus,
) -> crate::types::UserStatus {
    use crate::types::UserStatus;
    use grammers_client::tl::enums::UserStatus as TlStatus;

    match status {
        TlStatus::Online(_) => UserStatus::Online,
        TlStatus::Offline(_) => UserStatus::Offline,
        TlStatus::Recently(_) => UserStatus::Recently,
        TlStatus::LastWeek(_) => UserStatus::LastWeek,
        TlStatus::LastMonth(_) => UserStatus::LastMonth,
        TlStatus::Empty => UserStatus::Offline,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::UserStatus;

    #[test]
    fn test_peer_to_chat_id() {
        use grammers_client::tl::types;

        let user_peer =
            grammers_client::tl::enums::Peer::User(types::PeerUser { user_id: 12345 });
        assert_eq!(peer_to_chat_id(&user_peer), 12345);

        let chat_peer =
            grammers_client::tl::enums::Peer::Chat(types::PeerChat { chat_id: 67890 });
        assert_eq!(peer_to_chat_id(&chat_peer), 67890);

        let channel_peer =
            grammers_client::tl::enums::Peer::Channel(types::PeerChannel { channel_id: 11111 });
        assert_eq!(peer_to_chat_id(&channel_peer), 11111);
    }

    #[test]
    fn test_tl_status_to_user_status() {
        use grammers_client::tl::types;

        let online =
            grammers_client::tl::enums::UserStatus::Online(types::UserStatusOnline { expires: 0 });
        assert_eq!(tl_status_to_user_status(&online), UserStatus::Online);

        let offline = grammers_client::tl::enums::UserStatus::Offline(types::UserStatusOffline {
            was_online: 0,
        });
        assert_eq!(tl_status_to_user_status(&offline), UserStatus::Offline);

        let recently =
            grammers_client::tl::enums::UserStatus::Recently(types::UserStatusRecently {
                by_me: false,
            });
        assert_eq!(tl_status_to_user_status(&recently), UserStatus::Recently);

        let last_week =
            grammers_client::tl::enums::UserStatus::LastWeek(types::UserStatusLastWeek {
                by_me: false,
            });
        assert_eq!(tl_status_to_user_status(&last_week), UserStatus::LastWeek);

        let last_month =
            grammers_client::tl::enums::UserStatus::LastMonth(types::UserStatusLastMonth {
                by_me: false,
            });
        assert_eq!(tl_status_to_user_status(&last_month), UserStatus::LastMonth);

        let empty = grammers_client::tl::enums::UserStatus::Empty;
        assert_eq!(tl_status_to_user_status(&empty), UserStatus::Offline);
    }
}
