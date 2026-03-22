//! Message operations for the Telegram client.
//!
//! This module provides methods for working with messages:
//! - Fetching message history
//! - Sending messages
//! - Editing messages
//! - Deleting messages
//! - Forwarding messages

use grammers_client::message::InputMessage;
use grammers_client::tl;
use grammers_session::types::PeerKind;
use tracing::{debug, info};

use super::chats::{grammers_message_to_message, grammers_peer_to_user};
use super::client::TelegramClient;
use super::error::TelegramError;
use crate::types::Message;

impl TelegramClient {
    /// Gets message history for a chat.
    ///
    /// Messages are returned in reverse chronological order (newest first).
    /// Results are cached automatically.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to get messages from
    /// * `limit` - Maximum number of messages to retrieve
    /// * `offset_id` - Optional message ID to start from (for pagination)
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// or the chat is not found.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// // Get the last 50 messages
    /// let messages = client.get_messages(123456789, 50, None).await?;
    ///
    /// // Get 50 more messages before message ID 100
    /// let older = client.get_messages(123456789, 50, Some(100)).await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn get_messages(
        &self,
        chat_id: i64,
        limit: usize,
        offset_id: Option<i64>,
    ) -> Result<Vec<Message>, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        debug!(
            "Fetching {} messages from chat {}, offset: {:?}",
            limit, chat_id, offset_id
        );

        let mut iter = client.iter_messages(peer_ref);

        // Set the offset if provided
        if let Some(id) = offset_id {
            // We need to convert to i32 for grammers
            #[allow(clippy::cast_possible_truncation)]
            let id_i32 = id as i32;
            iter = iter.offset_id(id_i32);
        }

        // Limit the number of messages
        iter = iter.limit(limit);

        let mut messages = Vec::with_capacity(limit);

        while let Some(msg) = iter.next().await.map_err(TelegramError::from)? {
            // Cache the sender as a user if available
            if let Some(sender_peer) = msg.sender() {
                if let Some(user) = grammers_peer_to_user(sender_peer) {
                    self.cache().set_user(user);
                }
            }

            let message = grammers_message_to_message(&msg);

            // Cache the message
            self.cache().add_message(chat_id, message.clone());

            messages.push(message);

            if messages.len() >= limit {
                break;
            }
        }

        debug!("Fetched {} messages from chat {}", messages.len(), chat_id);
        Ok(messages)
    }

    /// Sends a text message to a chat.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to send the message to
    /// * `text` - Message text
    /// * `reply_to` - Optional message ID to reply to
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// the chat is not found, or sending fails.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// // Send a simple message
    /// let msg = client.send_message(123456789, "Hello!", None).await?;
    ///
    /// // Reply to a message
    /// let reply = client.send_message(123456789, "This is a reply", Some(42)).await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn send_message(
        &self,
        chat_id: i64,
        text: &str,
        reply_to: Option<i64>,
    ) -> Result<Message, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!("Sending message to chat {}", chat_id);

        let mut input_message = InputMessage::new().text(text);

        if let Some(reply_id) = reply_to {
            // Convert to i32 for grammers
            #[allow(clippy::cast_possible_truncation)]
            let reply_id_i32 = reply_id as i32;
            input_message = input_message.reply_to(Some(reply_id_i32));
        }

        let sent = client
            .send_message(peer_ref, input_message)
            .await
            .map_err(TelegramError::from)?;

        let message = grammers_message_to_message(&sent);

        // Cache the sent message
        self.cache().add_message(chat_id, message.clone());

        debug!("Sent message {} to chat {}", message.id, chat_id);
        Ok(message)
    }

    /// Edits an existing message.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat containing the message
    /// * `message_id` - ID of the message to edit
    /// * `new_text` - New message text
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// the chat/message is not found, or editing fails (e.g., not your message).
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// // Edit a message
    /// let edited = client.edit_message(123456789, 42, "Updated text").await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn edit_message(
        &self,
        chat_id: i64,
        message_id: i64,
        new_text: &str,
    ) -> Result<Message, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!("Editing message {} in chat {}", message_id, chat_id);

        // Convert to i32 for grammers
        #[allow(clippy::cast_possible_truncation)]
        let message_id_i32 = message_id as i32;

        let input_message = InputMessage::new().text(new_text);

        client
            .edit_message(peer_ref, message_id_i32, input_message)
            .await
            .map_err(TelegramError::from)?;

        // Get the updated message - we need to fetch it since edit doesn't return the message
        // For now, create an updated version based on what we sent
        let mut message = self
            .cache()
            .get_messages(chat_id)
            .into_iter()
            .find(|m| m.id == message_id)
            .unwrap_or_else(|| Message {
                id: message_id,
                chat_id,
                ..Default::default()
            });

        message.content.text = new_text.to_string();
        message.is_edited = true;
        message.edit_date = Some(chrono::Utc::now());

        // Update cache
        self.cache().update_message(chat_id, message.clone());

        debug!("Edited message {} in chat {}", message_id, chat_id);
        Ok(message)
    }

    /// Deletes messages from a chat.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat containing the messages
    /// * `message_ids` - IDs of messages to delete
    /// * `revoke` - If `true`, delete for everyone; if `false`, delete only for self
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// the chat is not found, or deletion fails.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// // Delete messages for everyone
    /// client.delete_messages(123456789, &[42, 43, 44], true).await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn delete_messages(
        &self,
        chat_id: i64,
        message_ids: &[i64],
        revoke: bool,
    ) -> Result<(), TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        info!(
            "Deleting {} messages from chat {}, revoke: {}",
            message_ids.len(),
            chat_id,
            revoke
        );

        // Convert message IDs to i32
        #[allow(clippy::cast_possible_truncation)]
        let ids: Vec<i32> = message_ids.iter().map(|&id| id as i32).collect();

        // Different API for channels/supergroups vs regular chats
        match peer_ref.id.kind() {
            PeerKind::Channel => {
                // For channels/supergroups, use channels.DeleteMessages
                client
                    .invoke(&tl::functions::channels::DeleteMessages {
                        channel: tl::types::InputChannel {
                            channel_id: peer_ref.id.bare_id(),
                            access_hash: peer_ref.auth.hash(),
                        }
                        .into(),
                        id: ids,
                    })
                    .await
                    .map_err(TelegramError::from)?;
            },
            PeerKind::User | PeerKind::UserSelf | PeerKind::Chat => {
                // For private chats and basic groups, use messages.DeleteMessages
                client
                    .invoke(&tl::functions::messages::DeleteMessages { revoke, id: ids })
                    .await
                    .map_err(TelegramError::from)?;
            },
        }

        // Update cache
        for &msg_id in message_ids {
            self.cache().delete_message(chat_id, msg_id);
        }

        debug!(
            "Deleted {} messages from chat {}",
            message_ids.len(),
            chat_id
        );
        Ok(())
    }

    /// Forwards messages to another chat.
    ///
    /// # Arguments
    ///
    /// * `from_chat_id` - ID of the source chat
    /// * `to_chat_id` - ID of the destination chat
    /// * `message_ids` - IDs of messages to forward
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// either chat is not found, or forwarding fails.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// // Forward messages to another chat
    /// let forwarded = client.forward_messages(123456789, 987654321, &[42, 43]).await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn forward_messages(
        &self,
        from_chat_id: i64,
        to_chat_id: i64,
        message_ids: &[i64],
    ) -> Result<Vec<Message>, TelegramError> {
        let client = self.require_authorized().await?;
        let from_peer_ref = self.get_peer_ref(from_chat_id).await?;
        let to_peer_ref = self.get_peer_ref(to_chat_id).await?;

        info!(
            "Forwarding {} messages from {} to {}",
            message_ids.len(),
            from_chat_id,
            to_chat_id
        );

        // Convert message IDs to i32
        #[allow(clippy::cast_possible_truncation)]
        let ids: Vec<i32> = message_ids.iter().map(|&id| id as i32).collect();

        let forwarded = client
            .forward_messages(to_peer_ref, &ids, from_peer_ref)
            .await
            .map_err(TelegramError::from)?;

        let messages: Vec<Message> = forwarded
            .into_iter()
            .flatten()
            .map(|msg| {
                let message = grammers_message_to_message(&msg);
                // Cache the forwarded message
                self.cache().add_message(to_chat_id, message.clone());
                message
            })
            .collect();

        debug!(
            "Forwarded {} messages from {} to {}",
            messages.len(),
            from_chat_id,
            to_chat_id
        );
        Ok(messages)
    }

    /// Searches messages in a chat.
    ///
    /// # Arguments
    ///
    /// * `chat_id` - ID of the chat to search in
    /// * `query` - Search query string
    /// * `limit` - Maximum number of messages to return
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected, not authorized,
    /// or the chat is not found.
    pub async fn search_messages(
        &self,
        chat_id: i64,
        query: &str,
        limit: usize,
    ) -> Result<Vec<Message>, TelegramError> {
        let client = self.require_authorized().await?;
        let peer_ref = self.get_peer_ref(chat_id).await?;

        debug!(
            "Searching for '{}' in chat {}, limit: {}",
            query, chat_id, limit
        );

        let mut iter = client.search_messages(peer_ref).query(query).limit(limit);

        let mut messages = Vec::with_capacity(limit);

        while let Some(msg) = iter.next().await.map_err(TelegramError::from)? {
            let message = grammers_message_to_message(&msg);
            messages.push(message);

            if messages.len() >= limit {
                break;
            }
        }

        debug!(
            "Found {} messages matching '{}' in chat {}",
            messages.len(),
            query,
            chat_id
        );
        Ok(messages)
    }
}

#[cfg(test)]
mod tests {
    use crate::types::MessageType;

    #[test]
    fn test_message_type_display() {
        assert_eq!(format!("{}", MessageType::Text), "Text");
        assert_eq!(format!("{}", MessageType::Photo), "Photo");
        assert_eq!(format!("{}", MessageType::Video), "Video");
    }
}
