//! Thread-safe in-memory cache for Telegram data.
//!
//! This module provides a cache system for storing chats, messages, and users
//! with thread-safe access using `RwLock` for concurrent reads.
//!
//! # Design Decisions
//!
//! - Uses `RwLock` to allow multiple concurrent readers
//! - Messages are stored per-chat with a configurable limit (FIFO eviction)
//! - All operations return cloned data to avoid lock contention
//! - Provides a `SharedCache` type alias for `Arc<Cache>` convenience

// The significant_drop_tightening lint gives false positives for our use case
// where we need to hold the lock for the entire operation duration.
#![allow(clippy::significant_drop_tightening)]

use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use crate::types::{Chat, Message, User};

/// A thread-safe cache for storing Telegram data.
///
/// The cache stores chats, messages (per-chat), and users with thread-safe
/// access. Messages are limited per chat to prevent unbounded memory growth.
///
/// # Examples
///
/// ```
/// use ithil::cache::{Cache, new_shared_cache};
/// use ithil::types::{User, Chat, Message};
///
/// // Create a shared cache with a limit of 100 messages per chat
/// let cache = new_shared_cache(100);
///
/// // Store a user
/// let user = User {
///     id: 12345,
///     first_name: "John".to_string(),
///     ..Default::default()
/// };
/// cache.set_user(user);
///
/// // Retrieve the user
/// if let Some(user) = cache.get_user(12345) {
///     println!("Found user: {}", user.get_display_name());
/// }
/// ```
#[derive(Debug)]
pub struct Cache {
    /// Chat storage: `chat_id` -> `Chat`
    chats: RwLock<HashMap<i64, Chat>>,
    /// Message storage: `chat_id` -> `Vec<Message>`
    /// Messages are stored in chronological order (oldest first)
    messages: RwLock<HashMap<i64, Vec<Message>>>,
    /// User storage: `user_id` -> `User`
    users: RwLock<HashMap<i64, User>>,
    /// Maximum number of messages to store per chat
    max_messages_per_chat: usize,
}

impl Cache {
    /// Creates a new cache with the specified message limit per chat.
    ///
    /// # Arguments
    ///
    /// * `max_messages_per_chat` - Maximum number of messages to store per chat.
    ///   When this limit is exceeded, the oldest messages are removed.
    ///
    /// # Examples
    ///
    /// ```
    /// use ithil::cache::Cache;
    ///
    /// let cache = Cache::new(100);
    /// ```
    #[must_use]
    pub fn new(max_messages_per_chat: usize) -> Self {
        Self {
            chats: RwLock::new(HashMap::new()),
            messages: RwLock::new(HashMap::new()),
            users: RwLock::new(HashMap::new()),
            max_messages_per_chat,
        }
    }

    // ========================================================================
    // Chat Methods
    // ========================================================================

    /// Retrieves a chat by ID.
    ///
    /// Returns `None` if the chat is not in the cache.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn get_chat(&self, id: i64) -> Option<Chat> {
        self.chats
            .read()
            .expect("chats lock poisoned")
            .get(&id)
            .cloned()
    }

    /// Stores or updates a chat in the cache.
    ///
    /// If a chat with the same ID already exists, it will be replaced.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn set_chat(&self, chat: Chat) {
        self.chats
            .write()
            .expect("chats lock poisoned")
            .insert(chat.id, chat);
    }

    /// Retrieves all chats from the cache.
    ///
    /// The returned vector is a clone of the cached data.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn get_all_chats(&self) -> Vec<Chat> {
        self.chats
            .read()
            .expect("chats lock poisoned")
            .values()
            .cloned()
            .collect()
    }

    /// Removes a chat from the cache.
    ///
    /// Also removes all messages associated with the chat.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn remove_chat(&self, id: i64) {
        self.chats.write().expect("chats lock poisoned").remove(&id);
        self.messages
            .write()
            .expect("messages lock poisoned")
            .remove(&id);
    }

    // ========================================================================
    // Message Methods
    // ========================================================================

    /// Retrieves all messages for a chat.
    ///
    /// Returns an empty vector if no messages are cached for the chat.
    /// Messages are returned in chronological order (oldest first).
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn get_messages(&self, chat_id: i64) -> Vec<Message> {
        self.messages
            .read()
            .expect("messages lock poisoned")
            .get(&chat_id)
            .cloned()
            .unwrap_or_default()
    }

    /// Adds a message to a chat's message list.
    ///
    /// If the message limit is exceeded, the oldest messages are removed.
    /// Messages are inserted in sorted order by ID (assumed to be chronological).
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn add_message(&self, chat_id: i64, message: Message) {
        let mut messages = self.messages.write().expect("messages lock poisoned");
        let chat_messages = messages.entry(chat_id).or_default();

        // Insert in sorted order by message ID
        let insert_pos = chat_messages
            .binary_search_by_key(&message.id, |m| m.id)
            .unwrap_or_else(|pos| pos);

        // Check if message already exists
        if insert_pos < chat_messages.len() && chat_messages[insert_pos].id == message.id {
            // Update existing message
            chat_messages[insert_pos] = message;
        } else {
            // Insert new message
            chat_messages.insert(insert_pos, message);

            // Enforce message limit (remove from the beginning - oldest messages)
            while chat_messages.len() > self.max_messages_per_chat {
                chat_messages.remove(0);
            }
        }
    }

    /// Updates an existing message in the cache.
    ///
    /// If the message doesn't exist, it will be added.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn update_message(&self, chat_id: i64, message: Message) {
        let mut messages = self.messages.write().expect("messages lock poisoned");
        let chat_messages = messages.entry(chat_id).or_default();

        // Find and update the message
        if let Some(existing) = chat_messages.iter_mut().find(|m| m.id == message.id) {
            *existing = message;
        } else {
            // Message not found, add it
            drop(messages);
            self.add_message(chat_id, message);
        }
    }

    /// Deletes a message from the cache.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn delete_message(&self, chat_id: i64, message_id: i64) {
        let mut messages = self.messages.write().expect("messages lock poisoned");
        if let Some(chat_messages) = messages.get_mut(&chat_id) {
            chat_messages.retain(|m| m.id != message_id);
        }
    }

    /// Returns the number of cached messages for a chat.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn message_count(&self, chat_id: i64) -> usize {
        self.messages
            .read()
            .expect("messages lock poisoned")
            .get(&chat_id)
            .map_or(0, Vec::len)
    }

    // ========================================================================
    // User Methods
    // ========================================================================

    /// Retrieves a user by ID.
    ///
    /// Returns `None` if the user is not in the cache.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn get_user(&self, id: i64) -> Option<User> {
        self.users
            .read()
            .expect("users lock poisoned")
            .get(&id)
            .cloned()
    }

    /// Stores or updates a user in the cache.
    ///
    /// If a user with the same ID already exists, it will be replaced.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn set_user(&self, user: User) {
        self.users
            .write()
            .expect("users lock poisoned")
            .insert(user.id, user);
    }

    /// Retrieves all users from the cache.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn get_all_users(&self) -> Vec<User> {
        self.users
            .read()
            .expect("users lock poisoned")
            .values()
            .cloned()
            .collect()
    }

    /// Removes a user from the cache.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn remove_user(&self, id: i64) {
        self.users.write().expect("users lock poisoned").remove(&id);
    }

    // ========================================================================
    // General Methods
    // ========================================================================

    /// Clears all data from the cache.
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    pub fn clear(&self) {
        self.chats.write().expect("chats lock poisoned").clear();
        self.messages
            .write()
            .expect("messages lock poisoned")
            .clear();
        self.users.write().expect("users lock poisoned").clear();
    }

    /// Returns the total number of cached items (chats + users + messages).
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn total_items(&self) -> usize {
        let chats_count = self.chats.read().expect("chats lock poisoned").len();
        let users_count = self.users.read().expect("users lock poisoned").len();
        let messages_count: usize = self
            .messages
            .read()
            .expect("messages lock poisoned")
            .values()
            .map(Vec::len)
            .sum();
        chats_count + users_count + messages_count
    }

    /// Returns cache statistics as a tuple: (chats, users, messages).
    ///
    /// # Panics
    ///
    /// Panics if the internal lock is poisoned (another thread panicked while holding it).
    #[must_use]
    pub fn stats(&self) -> (usize, usize, usize) {
        let chats_count = self.chats.read().expect("chats lock poisoned").len();
        let users_count = self.users.read().expect("users lock poisoned").len();
        let messages_count: usize = self
            .messages
            .read()
            .expect("messages lock poisoned")
            .values()
            .map(Vec::len)
            .sum();
        (chats_count, users_count, messages_count)
    }
}

impl Default for Cache {
    fn default() -> Self {
        Self::new(100) // Default to 100 messages per chat
    }
}

/// A thread-safe shared reference to a [`Cache`].
///
/// This is a convenience type alias for `Arc<Cache>`.
pub type SharedCache = Arc<Cache>;

/// Creates a new shared cache with the specified message limit.
///
/// # Arguments
///
/// * `max_messages_per_chat` - Maximum number of messages to store per chat.
///
/// # Examples
///
/// ```
/// use ithil::cache::new_shared_cache;
///
/// let cache = new_shared_cache(200);
/// // cache can now be cloned and shared across threads
/// ```
#[must_use]
pub fn new_shared_cache(max_messages_per_chat: usize) -> SharedCache {
    Arc::new(Cache::new(max_messages_per_chat))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::{ChatType, MessageContent, MessageType};

    fn create_test_user(id: i64, name: &str) -> User {
        User {
            id,
            first_name: name.to_string(),
            ..Default::default()
        }
    }

    fn create_test_chat(id: i64, title: &str) -> Chat {
        Chat {
            id,
            title: title.to_string(),
            chat_type: ChatType::Private,
            ..Default::default()
        }
    }

    fn create_test_message(id: i64, chat_id: i64, text: &str) -> Message {
        Message {
            id,
            chat_id,
            content: MessageContent {
                content_type: MessageType::Text,
                text: text.to_string(),
                ..Default::default()
            },
            ..Default::default()
        }
    }

    mod user_cache_tests {
        use super::*;

        #[test]
        fn set_and_get_user() {
            let cache = Cache::new(100);
            let user = create_test_user(1, "Alice");

            cache.set_user(user.clone());

            let retrieved = cache.get_user(1);
            assert!(retrieved.is_some());
            assert_eq!(retrieved.unwrap().first_name, "Alice");
        }

        #[test]
        fn get_nonexistent_user() {
            let cache = Cache::new(100);
            assert!(cache.get_user(999).is_none());
        }

        #[test]
        fn update_user() {
            let cache = Cache::new(100);
            let user1 = create_test_user(1, "Alice");
            cache.set_user(user1);

            let user2 = create_test_user(1, "Alice Updated");
            cache.set_user(user2);

            let retrieved = cache.get_user(1).unwrap();
            assert_eq!(retrieved.first_name, "Alice Updated");
        }

        #[test]
        fn remove_user() {
            let cache = Cache::new(100);
            let user = create_test_user(1, "Alice");
            cache.set_user(user);

            cache.remove_user(1);

            assert!(cache.get_user(1).is_none());
        }

        #[test]
        fn get_all_users() {
            let cache = Cache::new(100);
            cache.set_user(create_test_user(1, "Alice"));
            cache.set_user(create_test_user(2, "Bob"));
            cache.set_user(create_test_user(3, "Charlie"));

            let users = cache.get_all_users();
            assert_eq!(users.len(), 3);
        }
    }

    mod chat_cache_tests {
        use super::*;

        #[test]
        fn set_and_get_chat() {
            let cache = Cache::new(100);
            let chat = create_test_chat(1, "Test Chat");

            cache.set_chat(chat);

            let retrieved = cache.get_chat(1);
            assert!(retrieved.is_some());
            assert_eq!(retrieved.unwrap().title, "Test Chat");
        }

        #[test]
        fn get_nonexistent_chat() {
            let cache = Cache::new(100);
            assert!(cache.get_chat(999).is_none());
        }

        #[test]
        fn get_all_chats() {
            let cache = Cache::new(100);
            cache.set_chat(create_test_chat(1, "Chat 1"));
            cache.set_chat(create_test_chat(2, "Chat 2"));

            let chats = cache.get_all_chats();
            assert_eq!(chats.len(), 2);
        }

        #[test]
        fn remove_chat_also_removes_messages() {
            let cache = Cache::new(100);
            cache.set_chat(create_test_chat(1, "Test Chat"));
            cache.add_message(1, create_test_message(1, 1, "Hello"));
            cache.add_message(1, create_test_message(2, 1, "World"));

            assert_eq!(cache.message_count(1), 2);

            cache.remove_chat(1);

            assert!(cache.get_chat(1).is_none());
            assert_eq!(cache.message_count(1), 0);
        }
    }

    mod message_cache_tests {
        use super::*;

        #[test]
        fn add_and_get_messages() {
            let cache = Cache::new(100);
            cache.add_message(1, create_test_message(1, 1, "Hello"));
            cache.add_message(1, create_test_message(2, 1, "World"));

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 2);
            assert_eq!(messages[0].content.text, "Hello");
            assert_eq!(messages[1].content.text, "World");
        }

        #[test]
        fn get_messages_empty_chat() {
            let cache = Cache::new(100);
            let messages = cache.get_messages(999);
            assert!(messages.is_empty());
        }

        #[test]
        fn messages_sorted_by_id() {
            let cache = Cache::new(100);
            // Add out of order
            cache.add_message(1, create_test_message(3, 1, "Third"));
            cache.add_message(1, create_test_message(1, 1, "First"));
            cache.add_message(1, create_test_message(2, 1, "Second"));

            let messages = cache.get_messages(1);
            assert_eq!(messages[0].id, 1);
            assert_eq!(messages[1].id, 2);
            assert_eq!(messages[2].id, 3);
        }

        #[test]
        fn update_existing_message() {
            let cache = Cache::new(100);
            cache.add_message(1, create_test_message(1, 1, "Original"));

            cache.update_message(1, create_test_message(1, 1, "Updated"));

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 1);
            assert_eq!(messages[0].content.text, "Updated");
        }

        #[test]
        fn update_adds_if_not_exists() {
            let cache = Cache::new(100);
            cache.update_message(1, create_test_message(1, 1, "New Message"));

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 1);
            assert_eq!(messages[0].content.text, "New Message");
        }

        #[test]
        fn delete_message() {
            let cache = Cache::new(100);
            cache.add_message(1, create_test_message(1, 1, "Hello"));
            cache.add_message(1, create_test_message(2, 1, "World"));

            cache.delete_message(1, 1);

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 1);
            assert_eq!(messages[0].id, 2);
        }

        #[test]
        fn delete_nonexistent_message() {
            let cache = Cache::new(100);
            cache.add_message(1, create_test_message(1, 1, "Hello"));

            // Should not panic
            cache.delete_message(1, 999);
            cache.delete_message(999, 1);

            assert_eq!(cache.message_count(1), 1);
        }

        #[test]
        fn message_limit_enforcement() {
            let cache = Cache::new(3); // Limit to 3 messages
            cache.add_message(1, create_test_message(1, 1, "First"));
            cache.add_message(1, create_test_message(2, 1, "Second"));
            cache.add_message(1, create_test_message(3, 1, "Third"));
            cache.add_message(1, create_test_message(4, 1, "Fourth"));
            cache.add_message(1, create_test_message(5, 1, "Fifth"));

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 3);
            // Should have the newest messages (3, 4, 5)
            assert_eq!(messages[0].id, 3);
            assert_eq!(messages[1].id, 4);
            assert_eq!(messages[2].id, 5);
        }

        #[test]
        fn message_limit_with_out_of_order_insertion() {
            let cache = Cache::new(3);
            cache.add_message(1, create_test_message(5, 1, "Fifth"));
            cache.add_message(1, create_test_message(3, 1, "Third"));
            cache.add_message(1, create_test_message(1, 1, "First"));
            cache.add_message(1, create_test_message(4, 1, "Fourth"));
            cache.add_message(1, create_test_message(2, 1, "Second"));

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 3);
            // Messages should be in order but we keep the 3 newest
            // After adding in order 5,3,1,4,2, the sorted list is 1,2,3,4,5
            // With limit 3, we keep 3,4,5
            assert_eq!(messages[0].id, 3);
            assert_eq!(messages[1].id, 4);
            assert_eq!(messages[2].id, 5);
        }

        #[test]
        fn duplicate_message_id_updates() {
            let cache = Cache::new(100);
            cache.add_message(1, create_test_message(1, 1, "Original"));
            cache.add_message(1, create_test_message(1, 1, "Updated"));

            let messages = cache.get_messages(1);
            assert_eq!(messages.len(), 1);
            assert_eq!(messages[0].content.text, "Updated");
        }
    }

    mod general_cache_tests {
        use super::*;

        #[test]
        fn clear_cache() {
            let cache = Cache::new(100);
            cache.set_user(create_test_user(1, "Alice"));
            cache.set_chat(create_test_chat(1, "Chat"));
            cache.add_message(1, create_test_message(1, 1, "Hello"));

            cache.clear();

            assert!(cache.get_user(1).is_none());
            assert!(cache.get_chat(1).is_none());
            assert!(cache.get_messages(1).is_empty());
        }

        #[test]
        fn total_items() {
            let cache = Cache::new(100);
            cache.set_user(create_test_user(1, "Alice"));
            cache.set_user(create_test_user(2, "Bob"));
            cache.set_chat(create_test_chat(1, "Chat"));
            cache.add_message(1, create_test_message(1, 1, "Hello"));
            cache.add_message(1, create_test_message(2, 1, "World"));

            assert_eq!(cache.total_items(), 5); // 2 users + 1 chat + 2 messages
        }

        #[test]
        fn stats() {
            let cache = Cache::new(100);
            cache.set_user(create_test_user(1, "Alice"));
            cache.set_user(create_test_user(2, "Bob"));
            cache.set_chat(create_test_chat(1, "Chat"));
            cache.add_message(1, create_test_message(1, 1, "Hello"));
            cache.add_message(1, create_test_message(2, 1, "World"));
            cache.add_message(2, create_test_message(3, 2, "Other chat"));

            let (chats, users, messages) = cache.stats();
            assert_eq!(chats, 1);
            assert_eq!(users, 2);
            assert_eq!(messages, 3);
        }

        #[test]
        fn default_cache() {
            let cache = Cache::default();
            // Should have default limit of 100
            for i in 0..150 {
                cache.add_message(1, create_test_message(i, 1, "Message"));
            }
            assert_eq!(cache.message_count(1), 100);
        }
    }

    mod shared_cache_tests {
        use super::*;
        use std::thread;

        #[test]
        fn shared_cache_across_threads() {
            let cache = new_shared_cache(100);

            let cache1 = Arc::clone(&cache);
            let cache2 = Arc::clone(&cache);

            let handle1 = thread::spawn(move || {
                for i in 0..100 {
                    cache1.set_user(create_test_user(i, &format!("User{i}")));
                }
            });

            let handle2 = thread::spawn(move || {
                for i in 100..200 {
                    cache2.set_user(create_test_user(i, &format!("User{i}")));
                }
            });

            handle1.join().unwrap();
            handle2.join().unwrap();

            assert_eq!(cache.get_all_users().len(), 200);
        }

        #[test]
        fn concurrent_reads_and_writes() {
            let cache = new_shared_cache(100);

            // Pre-populate
            for i in 0..50 {
                cache.set_user(create_test_user(i, &format!("User{i}")));
            }

            let cache1 = Arc::clone(&cache);
            let cache2 = Arc::clone(&cache);

            // Writer thread
            let writer = thread::spawn(move || {
                for i in 50..100 {
                    cache1.set_user(create_test_user(i, &format!("User{i}")));
                }
            });

            // Reader thread
            let reader = thread::spawn(move || {
                let mut count = 0;
                for _ in 0..100 {
                    count = cache2.get_all_users().len();
                }
                count
            });

            writer.join().unwrap();
            let final_read = reader.join().unwrap();

            // Final state should have all 100 users
            assert_eq!(cache.get_all_users().len(), 100);
            // Reader should have seen at least 50 (initial) users
            assert!(final_read >= 50);
        }
    }
}
