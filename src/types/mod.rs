//! Core domain types for the Ithil Telegram client.
//!
//! This module provides shared type definitions used throughout the application,
//! including users, chats, messages, and various content types.
//!
//! # Design Decisions
//!
//! - All types derive `Debug`, `Clone`, and implement `Default` where sensible
//! - `Option<T>` is used instead of pointer types for nullable fields
//! - `chrono::DateTime<Utc>` is used for all timestamp fields
//! - Enums use explicit discriminants for clarity and future serialization

// Allow these lints for types that faithfully port the Go structs
#![allow(clippy::struct_excessive_bools)]
#![allow(clippy::cast_precision_loss)]

use chrono::{DateTime, Utc};
use std::fmt;
use std::time::Duration;

// ============================================================================
// User Types
// ============================================================================

/// Represents the online status of a Telegram user.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum UserStatus {
    /// User is currently online
    Online,
    /// User is offline (with optional last seen time)
    #[default]
    Offline,
    /// User was seen recently (within 1-3 days)
    Recently,
    /// User was seen within the last week
    LastWeek,
    /// User was seen within the last month
    LastMonth,
}

impl fmt::Display for UserStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Online => write!(f, "online"),
            Self::Offline => write!(f, "offline"),
            Self::Recently => write!(f, "recently"),
            Self::LastWeek => write!(f, "last week"),
            Self::LastMonth => write!(f, "last month"),
        }
    }
}

/// Represents a Telegram user.
#[derive(Debug, Clone, Default)]
pub struct User {
    /// Unique user identifier
    pub id: i64,
    /// User's first name
    pub first_name: String,
    /// User's last name (may be empty)
    pub last_name: String,
    /// User's username without @ (may be empty)
    pub username: String,
    /// User's phone number (may be empty)
    pub phone_number: String,
    /// ID of the user's profile photo (may be empty)
    pub profile_photo_id: String,
    /// User's online status
    pub status: UserStatus,
    /// Whether this user is a bot
    pub is_bot: bool,
    /// Whether this user is in the current user's contacts
    pub is_contact: bool,
    /// Whether this is a mutual contact
    pub is_mutual_contact: bool,
    /// Whether this user is verified by Telegram
    pub is_verified: bool,
    /// Whether this user has Telegram Premium
    pub is_premium: bool,
}

impl User {
    /// Returns the best available display name for the user.
    ///
    /// Priority: `FirstName + LastName` > `FirstName` > `Username` > empty string
    ///
    /// # Examples
    ///
    /// ```
    /// use ithil::types::User;
    ///
    /// let user = User {
    ///     first_name: "John".to_string(),
    ///     last_name: "Doe".to_string(),
    ///     ..Default::default()
    /// };
    /// assert_eq!(user.get_display_name(), "John Doe");
    ///
    /// let user = User {
    ///     first_name: "Jane".to_string(),
    ///     ..Default::default()
    /// };
    /// assert_eq!(user.get_display_name(), "Jane");
    ///
    /// let user = User {
    ///     username: "cooluser".to_string(),
    ///     ..Default::default()
    /// };
    /// assert_eq!(user.get_display_name(), "cooluser");
    /// ```
    #[must_use]
    pub fn get_display_name(&self) -> String {
        if !self.first_name.is_empty() {
            if !self.last_name.is_empty() {
                return format!("{} {}", self.first_name, self.last_name);
            }
            return self.first_name.clone();
        }
        if !self.username.is_empty() {
            return self.username.clone();
        }
        String::new()
    }
}

// ============================================================================
// Chat Types
// ============================================================================

/// Represents the type of a Telegram chat.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum ChatType {
    /// Private one-on-one conversation
    #[default]
    Private,
    /// Basic group chat
    Group,
    /// Supergroup (large group with advanced features)
    Supergroup,
    /// Channel (broadcast)
    Channel,
    /// Secret chat (end-to-end encrypted)
    Secret,
}

impl fmt::Display for ChatType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Private => write!(f, "Private"),
            Self::Group => write!(f, "Group"),
            Self::Supergroup => write!(f, "Supergroup"),
            Self::Channel => write!(f, "Channel"),
            Self::Secret => write!(f, "Secret"),
        }
    }
}

/// Represents notification settings for a chat.
#[derive(Debug, Clone, Default)]
pub struct NotificationSettings {
    /// Duration to mute notifications (in seconds, 0 = not muted)
    pub mute_for: i32,
    /// Custom notification sound identifier
    pub sound: String,
    /// Whether to show message preview in notifications
    pub show_preview: bool,
    /// Whether to use the default notification sound
    pub use_default_sound: bool,
    /// Whether to disable notifications for pinned messages
    pub disable_pinned: bool,
    /// Whether to disable notifications for mentions
    pub disable_mention: bool,
}

/// Represents a draft message in a chat.
#[derive(Debug, Clone, Default)]
pub struct Draft {
    /// ID of the message being replied to (0 if not a reply)
    pub reply_to_message_id: i64,
    /// When the draft was last updated
    pub date: DateTime<Utc>,
    /// The draft text content
    pub text: String,
}

/// Represents a custom chat folder/filter.
#[derive(Debug, Clone, Default)]
pub struct ChatFilter {
    /// Unique filter identifier
    pub id: i32,
    /// Display title of the filter
    pub title: String,
    /// Icon name for the filter
    pub icon_name: String,
    /// Chat IDs explicitly included in this filter
    pub included_chat_ids: Vec<i64>,
    /// Chat IDs explicitly excluded from this filter
    pub excluded_chat_ids: Vec<i64>,
    /// Whether to include contacts
    pub include_contacts: bool,
    /// Whether to include non-contacts
    pub include_non_contacts: bool,
    /// Whether to include groups
    pub include_groups: bool,
    /// Whether to include channels
    pub include_channels: bool,
    /// Whether to include bots
    pub include_bots: bool,
}

/// Represents a Telegram chat (private, group, supergroup, or channel).
#[derive(Debug, Clone, Default)]
pub struct Chat {
    /// Unique chat identifier
    pub id: i64,
    /// Type of chat
    pub chat_type: ChatType,
    /// Chat title (for groups/channels) or user name (for private chats)
    pub title: String,
    /// Chat username (without @)
    pub username: String,
    /// ID of the chat photo
    pub photo_id: String,
    /// The last message in the chat
    pub last_message: Option<Box<Message>>,
    /// Number of unread messages
    pub unread_count: i32,
    /// Whether this chat is pinned
    pub is_pinned: bool,
    /// Order of pinned chats (lower = higher priority, 0 = not pinned)
    pub pin_order: i32,
    /// Whether notifications are muted
    pub is_muted: bool,
    /// Draft message text
    pub draft_message: String,
    /// ID of the last read incoming message
    pub last_read_inbox_id: i64,
    /// ID of the last read outgoing message
    pub last_read_outbox_id: i64,
    /// Access hash required for API calls
    pub access_hash: i64,
    /// Online status (for private chats with users)
    pub user_status: UserStatus,
    /// Notification settings for this chat
    pub notification_settings: Option<NotificationSettings>,
    /// Indicates if chat has received a new message (for visual highlighting)
    pub has_new_message: bool,
}

// ============================================================================
// Message Types
// ============================================================================

/// Represents the type of message content.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum MessageType {
    /// Plain text message
    #[default]
    Text,
    /// Photo message
    Photo,
    /// Video message
    Video,
    /// Voice message
    Voice,
    /// Video note (round video)
    VideoNote,
    /// Audio file
    Audio,
    /// Document/file
    Document,
    /// Sticker
    Sticker,
    /// Animation (GIF)
    Animation,
    /// Location
    Location,
    /// Shared contact
    Contact,
    /// Poll
    Poll,
    /// Venue
    Venue,
    /// Game
    Game,
}

impl fmt::Display for MessageType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Text => write!(f, "Text"),
            Self::Photo => write!(f, "Photo"),
            Self::Video => write!(f, "Video"),
            Self::Voice => write!(f, "Voice"),
            Self::VideoNote => write!(f, "Video Note"),
            Self::Audio => write!(f, "Audio"),
            Self::Document => write!(f, "Document"),
            Self::Sticker => write!(f, "Sticker"),
            Self::Animation => write!(f, "Animation"),
            Self::Location => write!(f, "Location"),
            Self::Contact => write!(f, "Contact"),
            Self::Poll => write!(f, "Poll"),
            Self::Venue => write!(f, "Venue"),
            Self::Game => write!(f, "Game"),
        }
    }
}

/// Represents the type of text entity.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum EntityType {
    /// Bold text
    #[default]
    Bold,
    /// Italic text
    Italic,
    /// Monospace/code text
    Code,
    /// Preformatted/code block
    Pre,
    /// Text with embedded URL
    TextUrl,
    /// @mention
    Mention,
    /// #hashtag
    Hashtag,
    /// $cashtag
    Cashtag,
    /// `/bot_command`
    BotCommand,
    /// Clickable URL
    Url,
    /// Email address
    Email,
    /// Phone number
    PhoneNumber,
    /// Spoiler text (hidden until clicked)
    Spoiler,
    /// ~~Strikethrough~~ text
    Strikethrough,
    /// Underlined text
    Underline,
}

/// Represents a text entity (bold, italic, link, etc.).
#[derive(Debug, Clone, Default)]
pub struct MessageEntity {
    /// Type of entity
    pub entity_type: EntityType,
    /// Offset in UTF-16 code units to the start of the entity
    pub offset: i32,
    /// Length in UTF-16 code units
    pub length: i32,
    /// URL for `TextUrl` entities
    pub url: String,
    /// User ID for `Mention` entities
    pub user_id: i64,
}

/// Represents a photo size variant (thumbnail, medium, large, etc.).
#[derive(Debug, Clone, Default)]
pub struct PhotoSize {
    /// Size type identifier ("s", "m", "x", "y", "w", etc.)
    pub size_type: String,
    /// Width in pixels
    pub width: i32,
    /// Height in pixels
    pub height: i32,
    /// Size in bytes
    pub size: i32,
}

/// Represents a thumbnail for media content.
#[derive(Debug, Clone, Default)]
pub struct Thumbnail {
    /// Width in pixels
    pub width: i32,
    /// Height in pixels
    pub height: i32,
    /// Local file path (if downloaded)
    pub path: String,
}

/// Represents the current download status of media.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum DownloadStatus {
    /// Not yet downloaded
    #[default]
    NotDownloaded,
    /// Currently downloading
    Downloading,
    /// Successfully downloaded
    Downloaded,
    /// Download failed
    Failed,
}

/// Represents download progress information.
#[derive(Debug, Clone)]
pub struct DownloadProgress {
    /// Current download status
    pub status: DownloadStatus,
    /// Total size in bytes
    pub bytes_total: i64,
    /// Downloaded size in bytes
    pub bytes_loaded: i64,
    /// Error message if failed
    pub error: Option<String>,
    /// When the download started
    pub start_time: DateTime<Utc>,
    /// When the progress was last updated
    pub last_update: DateTime<Utc>,
}

impl Default for DownloadProgress {
    fn default() -> Self {
        Self {
            status: DownloadStatus::default(),
            bytes_total: 0,
            bytes_loaded: 0,
            error: None,
            start_time: Utc::now(),
            last_update: Utc::now(),
        }
    }
}

impl DownloadProgress {
    /// Returns download progress as percentage (0.0-100.0).
    #[must_use]
    pub fn get_percentage(&self) -> f64 {
        if self.bytes_total <= 0 {
            return 0.0;
        }
        let percentage = (self.bytes_loaded as f64 / self.bytes_total as f64) * 100.0;
        percentage.min(100.0)
    }

    /// Returns download speed in bytes per second.
    ///
    /// Returns 0.0 if no time has elapsed since start.
    #[must_use]
    pub fn get_speed(&self) -> f64 {
        let elapsed = (self.last_update - self.start_time).num_milliseconds() as f64 / 1000.0;
        if elapsed <= 0.0 {
            return 0.0;
        }
        self.bytes_loaded as f64 / elapsed
    }

    /// Returns estimated time remaining based on current speed.
    ///
    /// Returns `Duration::ZERO` if speed is 0 or all bytes are downloaded.
    #[must_use]
    pub fn get_eta(&self) -> Duration {
        if self.bytes_loaded >= self.bytes_total {
            return Duration::ZERO;
        }

        let speed = self.get_speed();
        if speed <= 0.0 {
            return Duration::ZERO;
        }

        let remaining_bytes = self.bytes_total - self.bytes_loaded;
        let eta_seconds = remaining_bytes as f64 / speed;
        Duration::from_secs_f64(eta_seconds)
    }
}

/// Represents media content in a message.
#[derive(Debug, Clone, Default)]
pub struct Media {
    /// Unique media identifier
    pub id: String,
    /// Width in pixels
    pub width: i32,
    /// Height in pixels
    pub height: i32,
    /// Duration in seconds (for video/audio)
    pub duration: i32,
    /// File size in bytes
    pub size: i64,
    /// MIME type
    pub mime_type: String,
    /// Thumbnail for the media
    pub thumbnail: Option<Thumbnail>,
    /// Local file path (if downloaded)
    pub local_path: String,
    /// Remote file path/ID
    pub remote_path: String,
    /// Whether the file is downloaded
    pub is_downloaded: bool,
    /// Access hash for Telegram API
    pub access_hash: i64,
    /// File reference for Telegram API
    pub file_reference: Vec<u8>,
    /// Available photo sizes (for photos)
    pub photo_sizes: Vec<PhotoSize>,
    /// Current download status
    pub download_status: DownloadStatus,
    /// Download progress information
    pub download_progress: Option<DownloadProgress>,
}

/// Represents a geographical location.
#[derive(Debug, Clone, Default)]
pub struct Location {
    /// Latitude
    pub latitude: f64,
    /// Longitude
    pub longitude: f64,
}

/// Represents a contact shared in a message.
#[derive(Debug, Clone, Default)]
pub struct Contact {
    /// Phone number
    pub phone_number: String,
    /// Contact's first name
    pub first_name: String,
    /// Contact's last name
    pub last_name: String,
    /// User ID (if registered on Telegram)
    pub user_id: i64,
    /// vCard string
    pub vcard: String,
}

/// Represents the type of poll.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum PollType {
    /// Regular poll (multiple choice)
    #[default]
    Regular,
    /// Quiz (single correct answer)
    Quiz,
}

/// Represents an option in a poll.
#[derive(Debug, Clone, Default)]
pub struct PollOption {
    /// Option text
    pub text: String,
    /// Number of voters who chose this option
    pub voter_count: i32,
    /// Whether the current user chose this option
    pub is_chosen: bool,
}

/// Represents a poll in a message.
#[derive(Debug, Clone, Default)]
pub struct Poll {
    /// Unique poll identifier
    pub id: String,
    /// Poll question
    pub question: String,
    /// Poll options
    pub options: Vec<PollOption>,
    /// Total number of voters
    pub total_voter_count: i32,
    /// Whether the poll is closed
    pub is_closed: bool,
    /// Whether the poll is anonymous
    pub is_anonymous: bool,
    /// Type of poll
    pub poll_type: PollType,
    /// Time in seconds before the poll auto-closes (0 = no limit)
    pub open_period: i32,
    /// When the poll closes
    pub close_date: Option<DateTime<Utc>>,
}

/// Represents a sticker in a message.
#[derive(Debug, Clone, Default)]
pub struct Sticker {
    /// Sticker set ID
    pub set_id: i64,
    /// Width in pixels
    pub width: i32,
    /// Height in pixels
    pub height: i32,
    /// Associated emoji
    pub emoji: String,
    /// Whether this is an animated sticker (TGS)
    pub is_animated: bool,
    /// Whether this is a video sticker (WEBM)
    pub is_video: bool,
    /// Sticker thumbnail
    pub thumbnail: Option<Thumbnail>,
    /// Sticker file
    pub file: Option<Box<Media>>,
}

/// Represents an animation (GIF) in a message.
#[derive(Debug, Clone, Default)]
pub struct Animation {
    /// Width in pixels
    pub width: i32,
    /// Height in pixels
    pub height: i32,
    /// Duration in seconds
    pub duration: i32,
    /// File name
    pub file_name: String,
    /// MIME type
    pub mime_type: String,
    /// Animation thumbnail
    pub thumbnail: Option<Thumbnail>,
    /// Animation file
    pub file: Option<Box<Media>>,
}

/// Represents a document in a message.
#[derive(Debug, Clone, Default)]
pub struct Document {
    /// File name
    pub file_name: String,
    /// MIME type
    pub mime_type: String,
    /// Document thumbnail
    pub thumbnail: Option<Thumbnail>,
    /// Document file
    pub file: Option<Box<Media>>,
}

/// Represents the origin of a forwarded message.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum ForwardOrigin {
    /// Forwarded from a user
    #[default]
    User,
    /// Forwarded from a chat
    Chat,
    /// Forwarded from a channel
    Channel,
    /// Forwarded from a user who hid their identity
    HiddenUser,
}

/// Contains information about forwarded messages.
#[derive(Debug, Clone, Default)]
pub struct ForwardInfo {
    /// Origin type of the forward
    pub origin: ForwardOrigin,
    /// Original chat ID (for channel/chat forwards)
    pub from_chat_id: i64,
    /// Original user ID (for user forwards)
    pub from_user_id: i64,
    /// Original message ID
    pub message_id: i64,
    /// When the original message was sent
    pub date: DateTime<Utc>,
    /// Author signature (for channel posts)
    pub author_signature: String,
}

/// Represents the content of a message.
#[derive(Debug, Clone, Default)]
pub struct MessageContent {
    /// Type of content
    pub content_type: MessageType,
    /// Text content (for text messages)
    pub text: String,
    /// Caption (for media messages)
    pub caption: String,
    /// Text entities (formatting, links, etc.)
    pub entities: Vec<MessageEntity>,
    /// Media content (photos, videos, etc.)
    pub media: Option<Box<Media>>,
    /// Location data
    pub location: Option<Location>,
    /// Contact data
    pub contact: Option<Contact>,
    /// Poll data
    pub poll: Option<Poll>,
    /// Sticker data
    pub sticker: Option<Box<Sticker>>,
    /// Animation data
    pub animation: Option<Box<Animation>>,
    /// Document data
    pub document: Option<Box<Document>>,
}

/// Represents a Telegram message.
#[derive(Debug, Clone, Default)]
pub struct Message {
    /// Unique message identifier within the chat
    pub id: i64,
    /// Chat ID where the message was sent
    pub chat_id: i64,
    /// Sender's user ID
    pub sender_id: i64,
    /// Message content
    pub content: MessageContent,
    /// When the message was sent
    pub date: DateTime<Utc>,
    /// When the message was last edited
    pub edit_date: Option<DateTime<Utc>>,
    /// Whether this message was sent by the current user
    pub is_outgoing: bool,
    /// Whether this is a channel post
    pub is_channel_post: bool,
    /// Whether this message is pinned
    pub is_pinned: bool,
    /// Whether this message has been edited
    pub is_edited: bool,
    /// Whether this message is forwarded
    pub is_forwarded: bool,
    /// ID of the message being replied to (0 if not a reply)
    pub reply_to_message_id: i64,
    /// Forward information (if forwarded)
    pub forward_info: Option<ForwardInfo>,
    /// View count (for channel posts)
    pub views: i32,
    /// Media album ID (for grouped media)
    pub media_album_id: i64,
}

// ============================================================================
// Authentication Types
// ============================================================================

/// Represents the authentication state.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum AuthState {
    /// Waiting for phone number input
    #[default]
    WaitPhoneNumber,
    /// Waiting for verification code
    WaitCode,
    /// Waiting for 2FA password
    WaitPassword,
    /// Waiting for registration (new account)
    WaitRegistration,
    /// Successfully authenticated
    Ready,
    /// Connection closed
    Closed,
}

impl fmt::Display for AuthState {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::WaitPhoneNumber => write!(f, "Waiting for phone number"),
            Self::WaitCode => write!(f, "Waiting for verification code"),
            Self::WaitPassword => write!(f, "Waiting for password"),
            Self::WaitRegistration => write!(f, "Waiting for registration"),
            Self::Ready => write!(f, "Ready"),
            Self::Closed => write!(f, "Closed"),
        }
    }
}

// ============================================================================
// Update Types
// ============================================================================

/// Represents the type of Telegram update.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum UpdateType {
    /// New message received
    #[default]
    NewMessage,
    /// Message content was updated
    MessageContent,
    /// Message was edited
    MessageEdited,
    /// Message was deleted
    MessageDeleted,
    /// Chat's last message changed
    ChatLastMessage,
    /// Chat title changed
    ChatTitle,
    /// Chat photo changed
    ChatPhoto,
    /// Chat read inbox updated
    ChatReadInbox,
    /// Chat read outbox updated
    ChatReadOutbox,
    /// Chat unread count changed
    ChatUnreadCount,
    /// Chat draft message changed
    ChatDraftMessage,
    /// User status changed
    UserStatus,
    /// New chat appeared
    NewChat,
    /// Chat position/order changed
    ChatPosition,
    /// File update
    File,
    /// File download progress update
    FileDownload,
}

/// Represents any data that can be attached to an update.
#[derive(Debug, Clone, Default)]
pub enum UpdateData {
    /// No additional data
    #[default]
    None,
    /// String data (titles, etc.)
    String(String),
    /// Integer data (counts, etc.)
    Integer(i64),
    /// User data
    User(Box<User>),
    /// Chat data
    Chat(Box<Chat>),
    /// Message data
    Message(Box<Message>),
    /// File download data
    FileDownload(Box<FileDownload>),
}

/// Represents a Telegram update event.
#[derive(Debug, Clone, Default)]
pub struct Update {
    /// Type of update
    pub update_type: UpdateType,
    /// Chat ID associated with this update
    pub chat_id: i64,
    /// Message associated with this update (if any)
    pub message: Option<Box<Message>>,
    /// Additional data for this update
    pub data: UpdateData,
}

// ============================================================================
// File Download Types
// ============================================================================

/// Represents the state of a file download.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Hash)]
pub enum FileDownloadState {
    /// Download is pending
    #[default]
    Pending,
    /// Currently downloading
    Downloading,
    /// Download completed
    Completed,
    /// Download failed
    Failed,
    /// Download was cancelled
    Cancelled,
}

/// Tracks the progress of a file download.
#[derive(Debug, Clone, Default)]
pub struct FileDownload {
    /// Unique file identifier
    pub file_id: String,
    /// Current download state
    pub state: FileDownloadState,
    /// Bytes downloaded so far
    pub downloaded_size: i64,
    /// Total file size in bytes
    pub total_size: i64,
    /// Local path where file is saved
    pub local_path: String,
    /// Error message (if failed)
    pub error: Option<String>,
}

#[cfg(test)]
mod tests {
    use super::*;

    mod user_tests {
        use super::*;

        #[test]
        fn get_display_name_with_full_name() {
            let user = User {
                first_name: "John".to_string(),
                last_name: "Doe".to_string(),
                ..Default::default()
            };
            assert_eq!(user.get_display_name(), "John Doe");
        }

        #[test]
        fn get_display_name_with_first_name_only() {
            let user = User {
                first_name: "Jane".to_string(),
                ..Default::default()
            };
            assert_eq!(user.get_display_name(), "Jane");
        }

        #[test]
        fn get_display_name_with_username_only() {
            let user = User {
                username: "cooluser".to_string(),
                ..Default::default()
            };
            assert_eq!(user.get_display_name(), "cooluser");
        }

        #[test]
        fn get_display_name_empty() {
            let user = User::default();
            assert_eq!(user.get_display_name(), "");
        }

        #[test]
        fn get_display_name_priority() {
            // First name takes priority over username
            let user = User {
                first_name: "John".to_string(),
                username: "johndoe".to_string(),
                ..Default::default()
            };
            assert_eq!(user.get_display_name(), "John");

            // Full name takes priority over everything
            let user = User {
                first_name: "John".to_string(),
                last_name: "Doe".to_string(),
                username: "johndoe".to_string(),
                ..Default::default()
            };
            assert_eq!(user.get_display_name(), "John Doe");
        }
    }

    mod download_progress_tests {
        use super::*;

        #[test]
        fn get_percentage_zero_total() {
            let progress = DownloadProgress {
                bytes_total: 0,
                bytes_loaded: 100,
                ..Default::default()
            };
            assert!((progress.get_percentage() - 0.0).abs() < f64::EPSILON);
        }

        #[test]
        fn get_percentage_half() {
            let progress = DownloadProgress {
                bytes_total: 1000,
                bytes_loaded: 500,
                ..Default::default()
            };
            assert!((progress.get_percentage() - 50.0).abs() < 0.001);
        }

        #[test]
        fn get_percentage_capped_at_100() {
            let progress = DownloadProgress {
                bytes_total: 100,
                bytes_loaded: 200,
                ..Default::default()
            };
            assert!((progress.get_percentage() - 100.0).abs() < 0.001);
        }

        #[test]
        fn get_eta_completed() {
            let progress = DownloadProgress {
                bytes_total: 100,
                bytes_loaded: 100,
                ..Default::default()
            };
            assert_eq!(progress.get_eta(), Duration::ZERO);
        }
    }

    mod enum_display_tests {
        use super::*;

        #[test]
        fn user_status_display() {
            assert_eq!(format!("{}", UserStatus::Online), "online");
            assert_eq!(format!("{}", UserStatus::Offline), "offline");
            assert_eq!(format!("{}", UserStatus::Recently), "recently");
        }

        #[test]
        fn chat_type_display() {
            assert_eq!(format!("{}", ChatType::Private), "Private");
            assert_eq!(format!("{}", ChatType::Channel), "Channel");
        }

        #[test]
        fn auth_state_display() {
            assert_eq!(
                format!("{}", AuthState::WaitPhoneNumber),
                "Waiting for phone number"
            );
            assert_eq!(format!("{}", AuthState::Ready), "Ready");
        }
    }
}
