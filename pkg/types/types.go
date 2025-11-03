// Package types provides shared type definitions used throughout the Ithil application.
package types

import "time"

// User represents a Telegram user.
type User struct {
	ID              int64
	FirstName       string
	LastName        string
	Username        string
	PhoneNumber     string
	ProfilePhotoID  string
	Status          UserStatus
	IsBot           bool
	IsContact       bool
	IsMutualContact bool
	IsVerified      bool
	IsPremium       bool
}

// GetDisplayName returns the best available display name for the user.
// Priority: FirstName + LastName > FirstName > Username > "User {ID}"
func (u *User) GetDisplayName() string {
	if u.FirstName != "" {
		if u.LastName != "" {
			return u.FirstName + " " + u.LastName
		}
		return u.FirstName
	}
	if u.Username != "" {
		return u.Username
	}
	return ""
}

// UserStatus represents the online status of a user.
type UserStatus int

const (
	UserStatusOnline UserStatus = iota
	UserStatusOffline
	UserStatusRecently
	UserStatusLastWeek
	UserStatusLastMonth
)

// Chat represents a Telegram chat (private, group, supergroup, or channel).
type Chat struct {
	ID                   int64
	Type                 ChatType
	Title                string
	Username             string
	PhotoID              string
	LastMessage          *Message
	UnreadCount          int
	IsPinned             bool
	PinOrder             int  // Order of pinned chats (lower = higher priority, 0 = not pinned)
	IsMuted              bool
	DraftMessage         string
	LastReadInboxID      int64
	LastReadOutboxID     int64
	AccessHash           int64      // Required for API calls to users and channels
	UserStatus           UserStatus // Online status for private chats
	NotificationSettings *NotificationSettings
	HasNewMessage        bool // Indicates if chat has received a new message (for visual highlighting)
}

// ChatType represents the type of chat.
type ChatType int

const (
	ChatTypePrivate ChatType = iota
	ChatTypeGroup
	ChatTypeSupergroup
	ChatTypeChannel
	ChatTypeSecret
)

// Message represents a Telegram message.
type Message struct {
	ID               int64
	ChatID           int64
	SenderID         int64
	Content          MessageContent
	Date             time.Time
	EditDate         time.Time
	IsOutgoing       bool
	IsChannelPost    bool
	IsPinned         bool
	IsEdited         bool
	IsForwarded      bool
	ReplyToMessageID int64
	ForwardInfo      *ForwardInfo
	Views            int
	MediaAlbumID     int64
}

// MessageContent represents the content of a message.
type MessageContent struct {
	Type      MessageType
	Text      string
	Caption   string
	Entities  []MessageEntity
	Media     *Media
	Location  *Location
	Contact   *Contact
	Poll      *Poll
	Sticker   *Sticker
	Animation *Animation
	Document  *Document
}

// MessageType represents the type of message content.
type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypePhoto
	MessageTypeVideo
	MessageTypeVoice
	MessageTypeVideoNote
	MessageTypeAudio
	MessageTypeDocument
	MessageTypeSticker
	MessageTypeAnimation
	MessageTypeLocation
	MessageTypeContact
	MessageTypePoll
	MessageTypeVenue
	MessageTypeGame
)

// MessageEntity represents a text entity (bold, italic, link, etc.).
type MessageEntity struct {
	Type   EntityType
	Offset int
	Length int
	URL    string
	UserID int64
}

// EntityType represents the type of text entity.
type EntityType int

const (
	EntityTypeBold EntityType = iota
	EntityTypeItalic
	EntityTypeCode
	EntityTypePre
	EntityTypeTextURL
	EntityTypeMention
	EntityTypeHashtag
	EntityTypeCashtag
	EntityTypeBotCommand
	EntityTypeURL
	EntityTypeEmail
	EntityTypePhoneNumber
	EntityTypeSpoiler
	EntityTypeStrikethrough
	EntityTypeUnderline
)

// PhotoSize represents a photo size variant (thumbnail, medium, large, etc.)
type PhotoSize struct {
	Type   string // Size type: "s", "m", "x", "y", "w", etc.
	Width  int
	Height int
	Size   int // Size in bytes
}

// Media represents media content in a message.
type Media struct {
	ID            string
	Width         int
	Height        int
	Duration      int
	Size          int64
	MimeType      string
	Thumbnail     *Thumbnail
	LocalPath     string
	RemotePath    string
	IsDownloaded  bool
	AccessHash    int64        // Telegram AccessHash for downloading
	FileReference []byte       // Telegram FileReference for downloading
	PhotoSizes    []PhotoSize  // Photo sizes for photos (needed for download)
}

// Thumbnail represents a thumbnail for media content.
type Thumbnail struct {
	Width  int
	Height int
	Path   string
}

// ForwardInfo contains information about forwarded messages.
type ForwardInfo struct {
	Origin          ForwardOrigin
	FromChatID      int64
	FromUserID      int64
	MessageID       int64
	Date            time.Time
	AuthorSignature string
}

// ForwardOrigin represents the origin of a forwarded message.
type ForwardOrigin int

const (
	ForwardOriginUser ForwardOrigin = iota
	ForwardOriginChat
	ForwardOriginChannel
	ForwardOriginHiddenUser
)

// Location represents a geographical location.
type Location struct {
	Latitude  float64
	Longitude float64
}

// Contact represents a contact shared in a message.
type Contact struct {
	PhoneNumber string
	FirstName   string
	LastName    string
	UserID      int64
	VCard       string
}

// Poll represents a poll in a message.
type Poll struct {
	ID              string
	Question        string
	Options         []PollOption
	TotalVoterCount int
	IsClosed        bool
	IsAnonymous     bool
	Type            PollType
	OpenPeriod      int
	CloseDate       time.Time
}

// PollOption represents an option in a poll.
type PollOption struct {
	Text       string
	VoterCount int
	IsChosen   bool
}

// PollType represents the type of poll.
type PollType int

const (
	PollTypeRegular PollType = iota
	PollTypeQuiz
)

// Sticker represents a sticker in a message.
type Sticker struct {
	SetID      int64
	Width      int
	Height     int
	Emoji      string
	IsAnimated bool
	IsVideo    bool
	Thumbnail  *Thumbnail
	File       *Media
}

// Animation represents an animation (GIF) in a message.
type Animation struct {
	Width     int
	Height    int
	Duration  int
	FileName  string
	MimeType  string
	Thumbnail *Thumbnail
	File      *Media
}

// Document represents a document in a message.
type Document struct {
	FileName  string
	MimeType  string
	Thumbnail *Thumbnail
	File      *Media
}

// NotificationSettings represents notification settings for a chat.
type NotificationSettings struct {
	MuteFor         int
	Sound           string
	ShowPreview     bool
	UseDefaultSound bool
	DisablePinned   bool
	DisableMention  bool
}

// Draft represents a draft message in a chat.
type Draft struct {
	ReplyToMessageID int64
	Date             time.Time
	Text             string
}

// ChatFilter represents a custom chat folder/filter.
type ChatFilter struct {
	ID                 int32
	Title              string
	IconName           string
	IncludedChatIDs    []int64
	ExcludedChatIDs    []int64
	IncludeContacts    bool
	IncludeNonContacts bool
	IncludeGroups      bool
	IncludeChannels    bool
	IncludeBots        bool
}

// AuthState represents the authentication state.
type AuthState int

const (
	AuthStateWaitPhoneNumber AuthState = iota
	AuthStateWaitCode
	AuthStateWaitPassword
	AuthStateWaitRegistration
	AuthStateReady
	AuthStateClosed
)

// UpdateType represents the type of Telegram update.
type UpdateType int

const (
	UpdateTypeNewMessage UpdateType = iota
	UpdateTypeMessageContent
	UpdateTypeMessageEdited
	UpdateTypeMessageDeleted
	UpdateTypeChatLastMessage
	UpdateTypeChatTitle
	UpdateTypeChatPhoto
	UpdateTypeChatReadInbox
	UpdateTypeChatReadOutbox
	UpdateTypeChatUnreadCount
	UpdateTypeChatDraftMessage
	UpdateTypeUserStatus
	UpdateTypeNewChat
	UpdateTypeChatPosition
	UpdateTypeFile
	UpdateTypeFileDownload
)

// Update represents a Telegram update event.
type Update struct {
	Type    UpdateType
	ChatID  int64
	Message *Message
	Data    interface{}
}

// FileDownloadState represents the state of a file download.
type FileDownloadState int

const (
	FileDownloadStatePending FileDownloadState = iota
	FileDownloadStateDownloading
	FileDownloadStateCompleted
	FileDownloadStateFailed
	FileDownloadStateCancelled
)

// FileDownload tracks the progress of a file download.
type FileDownload struct {
	FileID         string
	State          FileDownloadState
	DownloadedSize int64
	TotalSize      int64
	LocalPath      string
	Error          error
}
