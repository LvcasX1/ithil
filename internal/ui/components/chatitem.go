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
	Chat        *types.Chat
	IsSelected  bool
	IsFocused   bool
	Width       int
	ShowPreview bool
}

// NewChatItemComponent creates a new chat item component.
func NewChatItemComponent(chat *types.Chat, isSelected bool, isFocused bool, width int) *ChatItemComponent {
	return &ChatItemComponent{
		Chat:        chat,
		IsSelected:  isSelected,
		IsFocused:   isFocused,
		Width:       width,
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

	// Ultra-compact design: maximize horizontal space usage
	// - Selected chats get a visible border for visual emphasis
	// - Unselected chats have an INVISIBLE border (same border style, but transparent)
	// - This ensures ALL chats occupy exactly the same dimensions (height and width)
	// - Borders add 2 lines of vertical height, so unselected items need the same border
	// - This prevents visual shifting during navigation up/down
	// - All chats use the same width calculation for maximum space efficiency
	contentWidth := c.Width
	if contentWidth < 1 {
		contentWidth = 1
	}

	var itemStyle lipgloss.Style
	if c.IsSelected && c.IsFocused {
		// Selected AND focused: border only for emphasis (NO background)
		itemStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).                       // Full card border
			BorderForeground(lipgloss.Color(styles.BorderFocused)). // Bright blue border
			Padding(0, 1).                                          // Minimal padding
			Margin(0, 0).                                           // Equal top and bottom margins (0, 0)
			Width(contentWidth - 4)                                 // Account for borders (2) + padding (2)
	} else if c.IsSelected {
		// Selected but not focused: border only with dimmer color (NO background)
		itemStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).                    // Full card border
			BorderForeground(lipgloss.Color(styles.AccentCyan)). // Cyan border
			Padding(0, 1).                                       // Minimal padding
			Margin(0, 0).                                        // Equal top and bottom margins (0, 0)
			Width(contentWidth - 4)                              // Account for borders (2) + padding (2)
	} else {
		// Unselected: INVISIBLE border with SAME padding as selected to maintain consistent height
		// This is critical - the border adds vertical space even when invisible
		// Without this, unselected items are 2 lines shorter, causing visual shifting during navigation
		itemStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).                     // Same border type as selected
			BorderForeground(lipgloss.Color(styles.ColorBlack)).  // Make border invisible (black on black)
			Padding(0, 1).                                        // Same padding as selected (0 vertical, 1 horizontal)
			Margin(0, 0).                                         // Equal top and bottom margins (0, 0)
			Width(contentWidth - 4)                               // Same width calculation as selected for consistency
	}

	return itemStyle.Render(content)
}

// buildContent builds the chat item content.
func (c *ChatItemComponent) buildContent() string {
	var sb strings.Builder

	// Avatar (first letter of title in a circle)
	avatar := c.buildAvatar()

	// First line: title, indicators, and timestamp (compact)
	firstLine := c.buildFirstLine()

	// Combine avatar with first line
	avatarAndTitle := lipgloss.JoinHorizontal(lipgloss.Top, avatar+" ", firstLine)
	sb.WriteString(avatarAndTitle)

	// Second line: preview + metadata combined (more space-efficient)
	if c.ShowPreview && c.Chat.LastMessage != nil {
		sb.WriteString("\n")
		preview := c.buildPreview()
		metadata := c.buildMetadata()

		// Combine preview and metadata on same line with separator
		if metadata != "" {
			secondLine := preview + " " + lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.TextSecondary)).
				Render("•") + " " + metadata
			sb.WriteString("  " + secondLine) // Small indent for alignment
		} else {
			sb.WriteString("  " + preview)
		}
	} else {
		// If no preview, show metadata on same line as title
		metadata := c.buildMetadata()
		if metadata != "" {
			sb.WriteString("\n  " + metadata)
		}
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

	// Title is always bright and bold, selection is shown via card border
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextBright)).
		Bold(true)
	// Maximum space available: Width - avatar(2) - badges(~12) - timestamp(~8) - spacing(~3)
	maxTitleWidth := c.Width - 25
	if maxTitleWidth < 10 {
		maxTitleWidth = 10 // Ensure minimum title width
	}
	parts = append(parts, titleStyle.Render(utils.TruncateString(title, maxTitleWidth)))

	// Status indicators as badges
	var badges []string
	// NEW badge for chats with new messages (always first)
	if c.Chat.HasNewMessage {
		newBadge := lipgloss.NewStyle().
			Background(lipgloss.Color(styles.AccentCyan)).
			Foreground(lipgloss.Color(styles.ColorBlack)).
			Padding(0, 1).
			Bold(true).
			Render("NEW")
		badges = append(badges, newBadge)
	}
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
		// Use brighter color for new message unread count
		bgColor := styles.AccentRed
		if c.Chat.HasNewMessage {
			bgColor = styles.AccentCyan // Cyan background for new messages
		}
		unreadBadge := lipgloss.NewStyle().
			Background(lipgloss.Color(bgColor)).
			Foreground(lipgloss.Color(styles.ColorBlack)).
			Padding(0, 1).
			Bold(true).
			Render(unreadText)
		rightParts = append(rightParts, unreadBadge)
	}

	// Timestamp of last message
	if c.Chat.LastMessage != nil {
		timestamp := utils.FormatTimestamp(c.Chat.LastMessage.Date, true)
		// Replace spaces with non-breaking spaces to prevent wrapping within the timestamp
		// This ensures "34m ago" stays on one line instead of wrapping "ago" to the next line
		timestamp = strings.ReplaceAll(timestamp, " ", "\u00A0")
		// Timestamp color depends on whether chat list is focused AND selected
		var timestampStyle lipgloss.Style
		if c.IsSelected && c.IsFocused {
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

	// Calculate spacing - maximize horizontal space usage
	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)

	// Account for padding - now consistent for both selected and unselected
	// Both use Width(contentWidth - 4) which accounts for borders/padding
	paddingAdjustment := 4 // Border + padding space (or equivalent padding for unselected)

	spacingLen := c.Width - leftLen - rightLen - paddingAdjustment
	// Ensure minimum spacing of 2 to prevent timestamp from getting too close to left content
	if spacingLen < 2 {
		spacingLen = 2
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

	// Truncate preview to fit width - maximum horizontal space
	// Reserve space for metadata if it will be on the same line (separator + status text)
	// Since we only show status for private chats now, we can use more space
	maxLen := c.Width - 15 // Minimal reservation for separator + status (if any)
	if maxLen < 20 {
		maxLen = 20 // Ensure minimum preview length
	}
	preview = utils.TruncateString(preview, maxLen)

	// Style the preview with appropriate color
	// Only brighten if chat list is focused AND selected
	var previewStyle lipgloss.Style
	if c.IsSelected && c.IsFocused {
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

// buildMetadata builds the metadata line showing essential info (online status for private chats).
func (c *ChatItemComponent) buildMetadata() string {
	var parts []string

	// Only show online status for private chats (no chat type label)
	if c.Chat.Type == types.ChatTypePrivate {
		var statusStr string
		switch c.Chat.UserStatus {
		case types.UserStatusOnline:
			statusStr = "Online"
		case types.UserStatusRecently:
			statusStr = "Recently"
		case types.UserStatusLastWeek:
			statusStr = "Last week"
		case types.UserStatusLastMonth:
			statusStr = "Last month"
		}

		if statusStr != "" {
			var statusStyle lipgloss.Style
			// Only brighten if chat list is focused AND selected
			if c.IsSelected && c.IsFocused {
				statusStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(styles.AccentBlue)).
					Bold(false)
			} else {
				statusStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(styles.TextSecondary)).
					Bold(false)
			}
			parts = append(parts, statusStyle.Render(statusStr))
		}
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
