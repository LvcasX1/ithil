// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/internal/utils"
	"github.com/lvcasx1/ithil/pkg/types"
)

// ChatItemComponent represents a single chat in the chat list.
type ChatItemComponent struct {
	Chat       *types.Chat
	IsSelected bool
	Width      int
	ShowPreview bool
}

// NewChatItemComponent creates a new chat item component.
func NewChatItemComponent(chat *types.Chat, isSelected bool, width int) *ChatItemComponent {
	return &ChatItemComponent{
		Chat:       chat,
		IsSelected: isSelected,
		Width:      width,
		ShowPreview: true,
	}
}

// Render renders the chat item component.
func (c *ChatItemComponent) Render() string {
	if c.Chat == nil {
		return ""
	}

	// Build the chat item content
	content := c.buildContent()

	// Both selected and unselected have the same background
	// Only the border changes to indicate selection
	// c.Width is viewport width (already accounts for pane borders)
	// lipgloss Width() sets the CONTENT width, then adds padding and border on top
	// So we need to subtract: left padding (1) + right padding (1) + left border (1) + right border (1) = 4
	// This ensures the total rendered width (content + padding + border) equals c.Width
	contentWidth := c.Width - 4
	if contentWidth < 1 {
		contentWidth = 1
	}

	var itemStyle lipgloss.Style
	if c.IsSelected {
		// Selected: bright blue border, transparent background, compact padding
		itemStyle = lipgloss.NewStyle().
			Padding(0, 1). // Compact padding for more vertical space
			Border(lipgloss.ThickBorder()). // Thick border for emphasis
			BorderForeground(lipgloss.Color(styles.BorderFocused)). // Bright blue for selection
			Width(contentWidth) // Content width: viewport width minus padding and borders
	} else {
		// Unselected: subtle border, transparent background, compact padding
		itemStyle = lipgloss.NewStyle().
			Padding(0, 1). // Compact padding for more vertical space
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.BorderNormal)). // Subtle border
			Width(contentWidth) // Content width: viewport width minus padding and borders
	}

	return itemStyle.Render(content)
}

// buildContent builds the chat item content.
func (c *ChatItemComponent) buildContent() string {
	var sb strings.Builder

	// Avatar (first letter of title in a circle)
	avatar := c.buildAvatar()

	// First line: title and indicators
	firstLine := c.buildFirstLine()

	// Combine avatar with first line
	avatarAndTitle := lipgloss.JoinHorizontal(lipgloss.Top, avatar+" ", firstLine)
	sb.WriteString(avatarAndTitle)

	// Second line: message preview (if enabled and available)
	if c.ShowPreview && c.Chat.LastMessage != nil {
		sb.WriteString("\n")
		// Add spacing for alignment with title (avatar width + 1 space)
		spacing := "   "
		preview := c.buildPreview()
		sb.WriteString(spacing + preview)
	}

	// Third line: additional metadata (chat type, online status, message count)
	metadata := c.buildMetadata()
	if metadata != "" {
		sb.WriteString("\n")
		spacing := "   "
		sb.WriteString(spacing + metadata)
	}

	return sb.String()
}

// buildAvatar creates an avatar with the first letter of the chat title.
func (c *ChatItemComponent) buildAvatar() string {
	title := c.Chat.Title
	if title == "" {
		title = "?"
	}

	firstLetter := string([]rune(title)[0])

	// Choose color based on chat type
	var avatarColor string
	switch c.Chat.Type {
	case types.ChatTypePrivate:
		avatarColor = styles.AccentCyan // Cyan for private
	case types.ChatTypeGroup, types.ChatTypeSupergroup:
		avatarColor = styles.AccentYellow // Yellow for groups
	case types.ChatTypeChannel:
		avatarColor = styles.AccentMagenta // Magenta for channels
	default:
		avatarColor = styles.AccentBlue // Default
	}

	avatarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(avatarColor)).
		Bold(true)

	return avatarStyle.Render(firstLetter)
}

// buildFirstLine builds the first line of the chat item (title + indicators).
func (c *ChatItemComponent) buildFirstLine() string {
	var parts []string

	// Chat title
	title := c.Chat.Title
	if title == "" {
		title = fmt.Sprintf("Chat %d", c.Chat.ID)
	}

	var titleStyle lipgloss.Style
	if c.IsSelected {
		titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextBright)).
			Bold(true)
	} else {
		titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextBright)).
			Bold(true)
	}
	parts = append(parts, titleStyle.Render(utils.TruncateString(title, c.Width-20)))

	// Status indicators as badges
	var badges []string
	if c.Chat.IsPinned {
		badges = append(badges, "📌")
	}
	if c.Chat.IsMuted {
		badges = append(badges, "🔕")
	}
	// Online status for private chats
	if c.Chat.Type == types.ChatTypePrivate {
		if c.Chat.UserStatus == types.UserStatusOnline {
			badges = append(badges, "🟢")
		}
	}

	// Unread badge (right-aligned)
	var rightParts []string
	if c.Chat.UnreadCount > 0 {
		unreadText := fmt.Sprintf("%d", c.Chat.UnreadCount)
		if c.Chat.UnreadCount > 99 {
			unreadText = "99+"
		}
		unreadBadge := lipgloss.NewStyle().
			Background(lipgloss.Color(styles.AccentRed)).
			Foreground(lipgloss.Color(styles.TextBright)).
			Padding(0, 1).
			Bold(true).
			Render(unreadText)
		rightParts = append(rightParts, unreadBadge)
	}

	// Timestamp of last message
	if c.Chat.LastMessage != nil {
		timestamp := utils.FormatTimestamp(c.Chat.LastMessage.Date, true)
		var timestampStyle lipgloss.Style
		if c.IsSelected {
			timestampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextPrimary))
		} else {
			timestampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextSecondary))
		}
		rightParts = append(rightParts, timestampStyle.Render(timestamp))
	}

	// Add badges to parts
	if len(badges) > 0 {
		parts = append(parts, strings.Join(badges, " "))
	}

	// Combine left and right parts
	leftPart := strings.Join(parts, " ")
	rightPart := strings.Join(rightParts, " ")

	// Calculate spacing (account for avatar width + space = 3 chars)
	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)
	spacingLen := c.Width - leftLen - rightLen - 12 // Account for padding, borders, avatar

	if spacingLen < 1 {
		spacingLen = 1
	}
	spacing := strings.Repeat(" ", spacingLen)

	return leftPart + spacing + rightPart
}

// buildPreview builds the message preview line.
func (c *ChatItemComponent) buildPreview() string {
	if c.Chat.LastMessage == nil {
		return ""
	}

	msg := c.Chat.LastMessage
	var preview string

	// Add sender prefix for outgoing messages
	if msg.IsOutgoing {
		preview = "You: "
	}

	// Get message preview based on content type
	switch msg.Content.Type {
	case types.MessageTypeText:
		preview += msg.Content.Text
	case types.MessageTypePhoto:
		preview += "Photo"
		if msg.Content.Caption != "" {
			preview += ": " + msg.Content.Caption
		}
	case types.MessageTypeVideo:
		preview += "Video"
		if msg.Content.Caption != "" {
			preview += ": " + msg.Content.Caption
		}
	case types.MessageTypeVoice:
		preview += "Voice message"
	case types.MessageTypeAudio:
		preview += "Audio"
	case types.MessageTypeDocument:
		preview += "Document"
		if msg.Content.Document != nil && msg.Content.Document.FileName != "" {
			preview += ": " + msg.Content.Document.FileName
		}
	case types.MessageTypeSticker:
		preview += "Sticker"
	case types.MessageTypeAnimation:
		preview += "GIF"
	case types.MessageTypeLocation:
		preview += "Location"
	case types.MessageTypeContact:
		preview += "Contact"
	case types.MessageTypePoll:
		preview += "Poll"
		if msg.Content.Poll != nil {
			preview += ": " + msg.Content.Poll.Question
		}
	default:
		preview += "[Message]"
	}

	// Truncate preview to fit width (account for avatar spacing)
	maxLen := c.Width - 10 // Account for padding, borders, avatar
	preview = utils.TruncateString(preview, maxLen)

	// Style the preview with appropriate color
	var previewStyle lipgloss.Style
	if c.IsSelected {
		previewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextPrimary)).
			Italic(true)
	} else {
		previewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextSecondary)).
			Italic(true)
	}

	return previewStyle.Render(preview)
}

// buildMetadata builds the metadata line showing chat type, online status, etc.
func (c *ChatItemComponent) buildMetadata() string {
	var parts []string

	// Chat type indicator
	var chatTypeStr string
	switch c.Chat.Type {
	case types.ChatTypePrivate:
		chatTypeStr = "Private"
		// Show detailed online status for private chats
		if c.Chat.UserStatus == types.UserStatusOnline {
			chatTypeStr += " • Online"
		} else if c.Chat.UserStatus == types.UserStatusRecently {
			chatTypeStr += " • Recently"
		} else if c.Chat.UserStatus == types.UserStatusLastWeek {
			chatTypeStr += " • Last week"
		} else if c.Chat.UserStatus == types.UserStatusLastMonth {
			chatTypeStr += " • Last month"
		}
	case types.ChatTypeGroup:
		chatTypeStr = "Group"
	case types.ChatTypeSupergroup:
		chatTypeStr = "Supergroup"
	case types.ChatTypeChannel:
		chatTypeStr = "Channel"
	}

	if chatTypeStr != "" {
		var typeStyle lipgloss.Style
		if c.IsSelected {
			typeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.AccentBlue)).
				Bold(false)
		} else {
			typeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.TextSecondary)).
				Bold(false)
		}
		parts = append(parts, typeStyle.Render(chatTypeStr))
	}

	// Message count if available (total messages in chat)
	if c.Chat.LastMessage != nil && c.Chat.LastMessage.ID > 0 {
		msgCountStr := fmt.Sprintf("%d messages", c.Chat.LastMessage.ID)
		var msgStyle lipgloss.Style
		if c.IsSelected {
			msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.AccentBlue))
		} else {
			msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextSecondary))
		}
		parts = append(parts, msgStyle.Render(msgCountStr))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " • ")
}

// SetSelected sets the selected state of the chat item.
func (c *ChatItemComponent) SetSelected(selected bool) {
	c.IsSelected = selected
}

// SetWidth sets the width of the chat item.
func (c *ChatItemComponent) SetWidth(width int) {
	c.Width = width
}

// SetShowPreview sets whether to show message preview.
func (c *ChatItemComponent) SetShowPreview(show bool) {
	c.ShowPreview = show
}
