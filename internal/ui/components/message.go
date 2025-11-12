// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/media"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/internal/utils"
	"github.com/lvcasx1/ithil/pkg/types"
)

// MessageComponent represents a single message in the conversation view.
type MessageComponent struct {
	Message       *types.Message
	Width         int
	CurrentUserID int64
	SenderName    string // Cached sender name for display
	IsSelected    bool   // Whether this message is currently selected
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
		IsSelected:    false,
	}
}

// NewMessageComponentWithSelection creates a new message component with selection state.
func NewMessageComponentWithSelection(message *types.Message, width int, currentUserID int64, senderName string, isSelected bool) *MessageComponent {
	return &MessageComponent{
		Message:       message,
		Width:         width,
		CurrentUserID: currentUserID,
		SenderName:    senderName,
		IsSelected:    isSelected,
	}
}

// Render renders the message component.
func (m *MessageComponent) Render() string {
	if m.Message == nil {
		return ""
	}

	var contentBuilder strings.Builder

	// Add selection indicator if selected
	selectionMarker := ""
	if m.IsSelected {
		selectionMarker = styles.HighlightStyle.Render("▶ ")
	} else {
		selectionMarker = "  " // Two spaces to align with arrow
	}

	// Build message header (sender name + timestamp + edited indicator)
	header := m.renderHeader()
	if header != "" {
		contentBuilder.WriteString(selectionMarker)
		contentBuilder.WriteString(header)
		contentBuilder.WriteString("\n")
	}

	// Render reply if present
	if m.Message.ReplyToMessageID != 0 {
		reply := m.renderReply()
		// Indent reply slightly (add 2 spaces for selection marker alignment)
		indentedReply := indentText(reply, 4)
		contentBuilder.WriteString(indentedReply)
		contentBuilder.WriteString("\n")
	}

	// Render main content (indented, add 2 spaces for selection marker alignment)
	content := m.renderContent()
	indentedContent := indentText(content, 4)
	contentBuilder.WriteString(indentedContent)

	// Render footer (views, pinned, etc. - but NOT edited, that's in header)
	footer := m.renderFooter()
	if footer != "" {
		contentBuilder.WriteString("\n")
		// Indent footer to align with content (add 2 spaces for selection marker alignment)
		indentedFooter := indentText(footer, 4)
		contentBuilder.WriteString(indentedFooter)
	}

	// If selected, wrap the entire message in a subtle border or background
	result := contentBuilder.String()
	if m.IsSelected {
		// Add a subtle left border to highlight the selected message
		borderStyle := lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(styles.AccentCyan)).
			PaddingLeft(0)
		result = borderStyle.Render(result)
	}

	return result
}

// renderHeader renders the message header (sender name and timestamp).
func (m *MessageComponent) renderHeader() string {
	// Determine sender name
	var senderName string
	if m.Message.IsOutgoing {
		// Show "You" for own messages (matching screenshot)
		senderName = "You"
	} else {
		// Use provided sender name or fallback to User ID
		if m.SenderName != "" {
			senderName = m.SenderName
		} else {
			senderName = fmt.Sprintf("User %d", m.Message.SenderID)
		}
	}

	// Format timestamp
	timestamp := utils.FormatTimestamp(m.Message.Date, true)

	// Get user-specific color for the sender name
	userColor := getUserColor(m.Message.SenderID)
	senderStyle := lipgloss.NewStyle().
		Foreground(userColor).
		Bold(true)

	// Build header: "Sender • HH:MM" or "Sender • HH:MM (edited)"
	header := fmt.Sprintf("%s • %s",
		senderStyle.Render(senderName),
		styles.TimestampStyle.Render(timestamp))

	// Add edited indicator to header if message is edited
	if m.Message.IsEdited {
		header += " " + styles.EditedStyle.Render("(edited)")
	}

	return header
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

	case types.MessageTypeVideoNote:
		return m.renderMediaContent("Video Message", "", nil)

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

// renderMediaContent renders media content with caption and preview.
func (m *MessageComponent) renderMediaContent(mediaType, caption string, entities []types.MessageEntity) string {
	var sb strings.Builder

	// Media type indicator with icon
	var icon string
	switch mediaType {
	case "Photo":
		icon = "📷"
	case "Video":
		icon = "🎥"
	case "Voice Message":
		icon = "🎤"
	case "Audio":
		icon = "🎵"
	case "Sticker":
		icon = "🎨"
	case "GIF":
		icon = "🎬"
	case "Location":
		icon = "📍"
	case "Contact":
		icon = "👤"
	default:
		icon = "📎"
	}

	mediaIndicator := fmt.Sprintf("%s %s", icon, mediaType)
	sb.WriteString(styles.InfoStyle.Render(mediaIndicator))

	// Show media info if available
	if m.Message != nil && m.Message.Content.Media != nil {
		media := m.Message.Content.Media
		var mediaInfo []string

		// Size
		if media.Size > 0 {
			sizeStr := formatFileSize(media.Size)
			mediaInfo = append(mediaInfo, sizeStr)
		}

		// Dimensions
		if media.Width > 0 && media.Height > 0 {
			mediaInfo = append(mediaInfo, fmt.Sprintf("%dx%d", media.Width, media.Height))
		}

		// Duration
		if media.Duration > 0 {
			mediaInfo = append(mediaInfo, formatDuration(media.Duration))
		}

		if len(mediaInfo) > 0 {
			sb.WriteString("\n")
			sb.WriteString(styles.DimStyle.Render(strings.Join(mediaInfo, " • ")))
		}

		// Download status
		if media.LocalPath != "" && media.IsDownloaded {
			sb.WriteString("\n")
			sb.WriteString(styles.SuccessStyle.Render("✓ Downloaded"))

			// Render inline preview if downloaded
			preview := m.renderMediaPreview(mediaType, media.LocalPath)
			if preview != "" {
				sb.WriteString("\n\n")
				sb.WriteString(preview)
			}
		} else {
			sb.WriteString("\n")
			sb.WriteString(styles.DimStyle.Render("Press Enter to download and view"))
		}
	}

	// Caption if present
	if caption != "" {
		sb.WriteString("\n")
		sb.WriteString(m.renderTextContent(caption, entities))
	}

	return sb.String()
}

// renderMediaPreview renders a small inline preview of media.
func (m *MessageComponent) renderMediaPreview(mediaType, localPath string) string {
	// Calculate preview dimensions based on message width
	// Keep thumbnails small for inline preview - users can press Enter for fullscreen
	maxPreviewWidth := m.Width - 8 // Leave margin for indentation
	if maxPreviewWidth > 50 {
		maxPreviewWidth = 50 // Cap at reasonable inline size
	}
	if maxPreviewWidth < 30 {
		maxPreviewWidth = 30 // Minimum size for visibility
	}

	previewWidth := maxPreviewWidth
	previewHeight := 12 // Compact height for inline preview

	switch mediaType {
	case "Photo":
		// Render using mosaic (Unicode half-blocks) for better quality
		mosaicRenderer := media.NewMosaicRenderer(previewWidth, previewHeight, true)
		preview, err := mosaicRenderer.RenderImageFile(localPath)
		if err != nil {
			return styles.DimStyle.Render(fmt.Sprintf("Preview error: %s", err.Error()))
		}
		// Add a visual border around the image for better separation
		borderTop := styles.DimStyle.Render(strings.Repeat("─", min(previewWidth, lipgloss.Width(preview))))
		borderBottom := styles.DimStyle.Render(strings.Repeat("─", min(previewWidth, lipgloss.Width(preview))))
		return borderTop + "\n" + preview + "\n" + borderBottom

	case "Audio", "Voice Message":
		// Render enhanced audio waveform preview
		audioRenderer := media.NewAudioRenderer(previewWidth)
		var preview string
		var err error
		if mediaType == "Voice Message" {
			preview, err = audioRenderer.RenderVoicePreview(localPath, m.Message.Content.Media)
		} else {
			preview, err = audioRenderer.RenderAudioPreview(localPath, m.Message.Content.Media)
		}
		if err != nil {
			return styles.DimStyle.Render(fmt.Sprintf("Preview error: %s", err.Error()))
		}
		// Add playback hint
		playbackHint := styles.DimStyle.Render("▶ Press Enter to view full audio player")
		return preview + "\n" + playbackHint

	case "Video":
		// Show a simple placeholder for videos
		return styles.DimStyle.Render("🎥 Video file (press Enter to view details)")

	default:
		return ""
	}
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

// renderFooter renders the message footer (views, pinned, etc.).
// Note: edited indicator is now shown in the header, not footer
func (m *MessageComponent) renderFooter() string {
	var parts []string

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
		styles.AccentCyan,    // 0: Cyan
		styles.AccentGreen,   // 1: Green
		styles.AccentYellow,  // 2: Yellow
		styles.AccentMagenta, // 3: Magenta
		styles.AccentBlue,    // 4: Blue
		styles.AccentRed,     // 5: Red
		"10",                 // 6: Bright Green
		"11",                 // 7: Bright Yellow
		"13",                 // 8: Bright Magenta
		"14",                 // 9: Bright Cyan
	}

	// Use modulo to map user ID to a color index
	colorIndex := int(userID % int64(len(colors)))
	return lipgloss.Color(colors[colorIndex])
}

// indentText adds left indentation to each line of text.
func indentText(text string, spaces int) string {
	if text == "" {
		return text
	}
	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i := range lines {
		if lines[i] != "" { // Don't indent empty lines
			lines[i] = indent + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}
