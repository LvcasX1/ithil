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

// MessageComponent represents a single message in the conversation view.
type MessageComponent struct {
	Message       *types.Message
	Width         int
	CurrentUserID int64
	SenderName    string  // Cached sender name for display
}

// NewMessageComponent creates a new message component.
func NewMessageComponent(message *types.Message, width int) *MessageComponent {
	return &MessageComponent{
		Message: message,
		Width:   width,
	}
}

// NewMessageComponentWithUser creates a new message component with user info.
func NewMessageComponentWithUser(message *types.Message, width int, currentUserID int64, senderName string) *MessageComponent {
	return &MessageComponent{
		Message:       message,
		Width:         width,
		CurrentUserID: currentUserID,
		SenderName:    senderName,
	}
}

// Render renders the message component.
func (m *MessageComponent) Render() string {
	if m.Message == nil {
		return ""
	}

	// Build message content
	var contentBuilder strings.Builder

	// Build message header (sender name + timestamp)
	header := m.renderHeader()
	if header != "" {
		contentBuilder.WriteString(header)
		contentBuilder.WriteString("\n")
	}

	// Render reply if present
	if m.Message.ReplyToMessageID != 0 {
		contentBuilder.WriteString(m.renderReply())
		contentBuilder.WriteString("\n")
	}

	// Render main content
	content := m.renderContent()
	contentBuilder.WriteString(content)

	// Render footer (edited indicator, views, etc.)
	footer := m.renderFooter()
	if footer != "" {
		contentBuilder.WriteString("\n")
		contentBuilder.WriteString(footer)
	}

	messageContent := contentBuilder.String()

	// Choose style based on message direction
	var messageStyle lipgloss.Style

	// Message bubble max width (use nearly full width with small margins)
	// Leave 4 chars total (2 on each side) for equal margins
	bubbleMaxWidth := m.Width - 4
	if bubbleMaxWidth < 30 {
		bubbleMaxWidth = 30
	}

	// Get user-specific border color
	borderColor := getUserColor(m.Message.SenderID)

	if m.Message.IsOutgoing {
		// Outgoing messages: left-aligned, adapt to content width
		messageStyle = styles.MessageOutgoingStyle.
			MaxWidth(bubbleMaxWidth).
			BorderForeground(borderColor)
	} else {
		// Incoming messages: left-aligned, adapt to content width
		messageStyle = styles.MessageIncomingStyle.
			MaxWidth(bubbleMaxWidth).
			BorderForeground(borderColor)
	}

	// Render the message bubble
	bubble := messageStyle.Render(messageContent)

	// Add small left margin for equal spacing on both sides
	container := lipgloss.NewStyle().
		PaddingLeft(2)

	return container.Render(bubble)
}

// renderHeader renders the message header (sender name and timestamp).
func (m *MessageComponent) renderHeader() string {
	var parts []string

	// Add sender name
	var senderName string
	if m.Message.IsOutgoing {
		// Show "Me" for own messages
		senderName = "Me"
	} else {
		// Use provided sender name or fallback to User ID
		if m.SenderName != "" {
			senderName = m.SenderName
		} else {
			senderName = fmt.Sprintf("User %d", m.Message.SenderID)
		}
	}
	parts = append(parts, styles.SenderNameStyle.Render(senderName))

	// Add timestamp
	timestamp := utils.FormatTimestamp(m.Message.Date, true)
	parts = append(parts, styles.TimestampStyle.Render(timestamp))

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// renderContent renders the main message content.
func (m *MessageComponent) renderContent() string {
	content := m.Message.Content

	switch content.Type {
	case types.MessageTypeText:
		return m.renderTextContent(content.Text, content.Entities)

	case types.MessageTypePhoto:
		return m.renderMediaContent("Photo", content.Caption, content.Entities)

	case types.MessageTypeVideo:
		return m.renderMediaContent("Video", content.Caption, content.Entities)

	case types.MessageTypeVoice:
		return m.renderMediaContent("Voice Message", "", nil)

	case types.MessageTypeAudio:
		return m.renderMediaContent("Audio", content.Caption, content.Entities)

	case types.MessageTypeDocument:
		fileName := "Document"
		if content.Document != nil {
			fileName = content.Document.FileName
		}
		return m.renderMediaContent(fileName, content.Caption, content.Entities)

	case types.MessageTypeSticker:
		return m.renderMediaContent("Sticker", "", nil)

	case types.MessageTypeAnimation:
		return m.renderMediaContent("GIF", "", nil)

	case types.MessageTypeLocation:
		return m.renderMediaContent("Location", "", nil)

	case types.MessageTypeContact:
		if content.Contact != nil {
			return m.renderMediaContent(
				fmt.Sprintf("Contact: %s %s", content.Contact.FirstName, content.Contact.LastName),
				"",
				nil,
			)
		}
		return m.renderMediaContent("Contact", "", nil)

	case types.MessageTypePoll:
		if content.Poll != nil {
			return m.renderPoll(content.Poll)
		}
		return m.renderMediaContent("Poll", "", nil)

	default:
		return styles.DimStyle.Render("[Unsupported message type]")
	}
}

// renderTextContent renders text content with formatting entities.
func (m *MessageComponent) renderTextContent(text string, entities []types.MessageEntity) string {
	if len(entities) == 0 {
		return text
	}

	// TODO: Implement proper entity rendering
	// For now, just return the plain text
	return text
}

// renderMediaContent renders media content with caption.
func (m *MessageComponent) renderMediaContent(mediaType, caption string, entities []types.MessageEntity) string {
	var sb strings.Builder

	// Media type indicator
	sb.WriteString(styles.InfoStyle.Render(fmt.Sprintf("[%s]", mediaType)))

	// Caption if present
	if caption != "" {
		sb.WriteString("\n")
		sb.WriteString(m.renderTextContent(caption, entities))
	}

	return sb.String()
}

// renderPoll renders a poll message.
func (m *MessageComponent) renderPoll(poll *types.Poll) string {
	var sb strings.Builder

	// Poll question
	sb.WriteString(styles.HighlightStyle.Render(poll.Question))
	sb.WriteString("\n\n")

	// Poll options
	for i, option := range poll.Options {
		prefix := fmt.Sprintf("%d. ", i+1)
		if option.IsChosen {
			prefix = "✓ " + prefix
		}

		sb.WriteString(prefix)
		sb.WriteString(option.Text)
		sb.WriteString(fmt.Sprintf(" (%d votes)", option.VoterCount))
		sb.WriteString("\n")
	}

	// Poll info
	info := fmt.Sprintf("\nTotal votes: %d", poll.TotalVoterCount)
	if poll.IsClosed {
		info += " (Closed)"
	}
	sb.WriteString(styles.DimStyle.Render(info))

	return sb.String()
}

// renderReply renders a reply preview.
func (m *MessageComponent) renderReply() string {
	// TODO: Fetch the replied message and render a preview
	replyText := fmt.Sprintf("Reply to message %d", m.Message.ReplyToMessageID)
	return styles.ReplyStyle.Render(replyText)
}

// renderFooter renders the message footer (edited, views, etc.).
func (m *MessageComponent) renderFooter() string {
	var parts []string

	// Edited indicator
	if m.Message.IsEdited {
		parts = append(parts, styles.EditedStyle.Render("edited"))
	}

	// Views for channel posts
	if m.Message.Views > 0 {
		views := fmt.Sprintf("%d views", m.Message.Views)
		parts = append(parts, styles.DimStyle.Render(views))
	}

	// Pinned indicator
	if m.Message.IsPinned {
		parts = append(parts, styles.PinnedStyle.Render("pinned"))
	}

	if len(parts) == 0 {
		return ""
	}

	return styles.DimStyle.Render(strings.Join(parts, " • "))
}

// SetWidth sets the width of the message component.
func (m *MessageComponent) SetWidth(width int) {
	m.Width = width
}

// getUserColor returns a consistent color for a given user ID.
// Each user gets a unique color based on their ID, ensuring visual distinction.
func getUserColor(userID int64) lipgloss.Color {
	// Color palette for user distinction
	colors := []string{
		styles.AccentCyan,       // 0: Cyan
		styles.AccentGreen,      // 1: Green
		styles.AccentYellow,     // 2: Yellow
		styles.AccentMagenta,    // 3: Magenta
		styles.AccentBlue,       // 4: Blue
		styles.AccentRed,        // 5: Red
		"10",                    // 6: Bright Green
		"11",                    // 7: Bright Yellow
		"13",                    // 8: Bright Magenta
		"14",                    // 9: Bright Cyan
	}

	// Use modulo to map user ID to a color index
	colorIndex := int(userID % int64(len(colors)))
	return lipgloss.Color(colors[colorIndex])
}
