// Package models provides Bubbletea models for the Ithil TUI.
package models

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/cache"
	"github.com/lvcasx1/ithil/internal/telegram"
	"github.com/lvcasx1/ithil/internal/ui/components"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/pkg/types"
)

// ConversationModel represents the conversation/message view pane.
type ConversationModel struct {
	client         *telegram.Client
	cache          *cache.Cache
	currentChat    *types.Chat
	messages       []*types.Message
	width          int
	height         int
	focused        bool
	viewport       viewport.Model
	input          *components.InputComponent
	replyToMessage *types.Message                   // Message being replied to
	editingMessage *types.Message                   // Message being edited
	sendingMessage bool                             // Flag to show sending state
	typingUsers    map[int64]time.Time              // userID -> when they started typing
	mediaViewer    *components.MediaViewerComponent // Media viewer modal
	selectedMsgIdx int                              // Index of selected message for media viewing
}

// NewConversationModel creates a new conversation model.
func NewConversationModel(client *telegram.Client, cache *cache.Cache) *ConversationModel {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = false

	return &ConversationModel{
		client:         client,
		cache:          cache,
		messages:       []*types.Message{},
		viewport:       vp,
		input:          components.NewInputComponent(100, 5),
		typingUsers:    make(map[int64]time.Time),
		mediaViewer:    components.NewMediaViewerComponent(80, 30),
		selectedMsgIdx: -1,
	}
}

// Init initializes the conversation model.
func (m *ConversationModel) Init() tea.Cmd {
	return m.input.Init()
}

// Update handles conversation model updates.
func (m *ConversationModel) Update(msg tea.Msg) (*ConversationModel, tea.Cmd) {
	var cmd tea.Cmd

	// If media viewer is visible, let it handle updates first
	if m.mediaViewer.IsVisible() {
		var viewerCmd tea.Cmd
		m.mediaViewer, viewerCmd = m.mediaViewer.Update(msg)
		if viewerCmd != nil {
			return m, viewerCmd
		}
		// Check if viewer was dismissed
		if !m.mediaViewer.IsVisible() {
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// CRITICAL: Only handle keys if this pane is focused
		// If input is focused, let it handle the keys exclusively
		if m.input.Focused && m.focused {
			return m.handleInputKeys(msg)
		}

		// Otherwise handle navigation keys only when pane is focused
		if m.focused && !m.input.Focused && !m.mediaViewer.IsVisible() {
			switch msg.String() {
			case "up", "k":
				m.viewport.LineUp(1)
				return m, nil
			case "down", "j":
				m.viewport.LineDown(1)
				return m, nil
			case "pgup":
				m.viewport.ViewUp()
				return m, nil
			case "pgdown":
				m.viewport.ViewDown()
				return m, nil
			case "g":
				m.viewport.GotoTop()
				return m, nil
			case "G":
				m.viewport.GotoBottom()
				return m, nil
			case "i", "a":
				// Focus input to type new message
				m.ClearReplyEdit()
				return m, m.input.Focus()
			case "r":
				// Reply to last message (since we don't have message selection in viewport mode)
				if len(m.messages) > 0 {
					m.SetReplyTo(m.messages[len(m.messages)-1])
					return m, m.input.Focus()
				}
				return m, nil
			case "e":
				// Edit last outgoing message (since we don't have message selection in viewport mode)
				for i := len(m.messages) - 1; i >= 0; i-- {
					if m.messages[i].IsOutgoing {
						m.SetEditing(m.messages[i])
						return m, m.input.Focus()
						break
					}
				}
				return m, nil
			case "enter":
				// Open media viewer for last message with media
				return m, m.openMediaViewer()
			}
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case chatSelectedMsg:
		m.currentChat = msg.chat
		m.loadMessages()
		return m, nil

	case messagesUpdatedMsg:
		m.messages = msg.messages
		m.updateViewport()
		return m, nil

	case components.MediaViewerDismissedMsg:
		m.mediaViewer.Hide()
		return m, nil

	case components.MediaDownloadRequestMsg:
		// Handle media download request
		return m, m.downloadMediaForViewer(msg.Message)

	case components.MediaDownloadedMsg:
		// Pass to media viewer
		var viewerCmd tea.Cmd
		m.mediaViewer, viewerCmd = m.mediaViewer.Update(msg)
		return m, viewerCmd
	}

	// CRITICAL: Only update viewport when this pane is focused
	// This prevents navigation keys from affecting the viewport when the pane is not focused
	if m.focused {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	// Only update input component if pane is focused and input is focused
	// This prevents input from capturing keys when it shouldn't
	var inputCmd tea.Cmd
	if m.focused && m.input.Focused {
		m.input, inputCmd = m.input.Update(msg)
	}

	return m, tea.Batch(cmd, inputCmd)
}

// handleInputKeys handles keys when input is focused.
func (m *ConversationModel) handleInputKeys(msg tea.KeyMsg) (*ConversationModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Clear reply/edit state and unfocus input
		m.ClearReplyEdit()
		m.input.Blur()
		return m, nil
	case "ctrl+a":
		// Attach file - prompt for file path
		return m, func() tea.Msg {
			return fileAttachRequestMsg{}
		}
	case "ctrl+x":
		// Remove attachment
		m.input.ClearAttachment()
		return m, nil
	case "enter":
		// Check if we're editing or sending
		if m.editingMessage != nil {
			return m, m.editMessage()
		}
		return m, m.sendMessage()
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

// View renders the conversation model.
func (m *ConversationModel) View() string {
	// Render base content
	var baseView string
	if m.currentChat == nil {
		baseView = m.renderEmpty()
	} else {
		baseView = m.renderConversationView()
	}

	// If media viewer is visible, overlay it
	if m.mediaViewer.IsVisible() {
		viewerView := m.mediaViewer.View()
		// Center the media viewer
		overlay := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			viewerView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
		return overlay
	}

	return baseView
}

// renderConversationView renders the main conversation view.
func (m *ConversationModel) renderConversationView() string {
	// Typing indicator
	typingView := m.renderTypingIndicator()

	// Determine border colors based on focus
	var borderColor lipgloss.Color
	var titleColor lipgloss.Color

	if m.focused {
		borderColor = lipgloss.Color(styles.BorderFocused)
		titleColor = lipgloss.Color(styles.AccentCyan)
	} else {
		borderColor = lipgloss.Color(styles.BorderNormal)
		titleColor = lipgloss.Color(styles.TextBright)
	}

	// Get viewport content
	viewportContent := m.viewport.View()

	// Create custom top border with embedded title
	title := " " + m.currentChat.Title + " "
	// Use UTF-8 rune count for accurate visual width (handles multi-byte characters like á, ñ, etc.)
	titleLen := utf8.RuneCountInString(title)

	// Truncate title if too long
	maxTitleLen := m.width - 6 // Reserve space for "┌─", "─┐", and some dashes
	if titleLen > maxTitleLen {
		if maxTitleLen > 7 {
			// Convert to runes for proper truncation of UTF-8 strings
			titleRunes := []rune(m.currentChat.Title)
			truncatedLen := maxTitleLen - 7
			if truncatedLen > len(titleRunes) {
				truncatedLen = len(titleRunes)
			}
			title = " " + string(titleRunes[:truncatedLen]) + "... "
			titleLen = utf8.RuneCountInString(title)
		} else {
			title = " ... "
			titleLen = 5
		}
	}

	// Calculate remaining border width
	remainingWidth := m.width - titleLen - 3
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Build top border with title
	topBorderLeft := lipgloss.NewStyle().Foreground(borderColor).Render("┌─")
	topBorderTitle := lipgloss.NewStyle().Foreground(titleColor).Render(title)
	topBorderRight := lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", remainingWidth) + "┐")
	topBorder := topBorderLeft + topBorderTitle + topBorderRight

	// Create side borders
	leftBorder := lipgloss.NewStyle().Foreground(borderColor).Render("│")
	rightBorder := lipgloss.NewStyle().Foreground(borderColor).Render("│")

	// Wrap viewport content with side borders
	contentLines := strings.Split(viewportContent, "\n")
	var borderedLines []string
	for _, line := range contentLines {
		lineWidth := lipgloss.Width(line)
		padding := ""
		if lineWidth < m.width-4 {
			padding = strings.Repeat(" ", m.width-4-lineWidth)
		}
		borderedLines = append(borderedLines, leftBorder+" "+line+padding+" "+rightBorder)
	}

	// Create bottom border
	bottomBorder := lipgloss.NewStyle().Foreground(borderColor).Render("└" + strings.Repeat("─", m.width-2) + "┘")

	// Combine border parts
	messagesWithBorder := topBorder + "\n" + strings.Join(borderedLines, "\n") + "\n" + bottomBorder

	// Input box
	inputView := m.input.View()

	// Combine all elements
	var parts []string
	parts = append(parts, messagesWithBorder)
	if typingView != "" {
		parts = append(parts, typingView)
	}
	parts = append(parts, inputView)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderEmpty renders an empty state.
func (m *ConversationModel) renderEmpty() string {
	emptyText := "Select a chat to start messaging"

	// Border style
	borderColor := lipgloss.Color(styles.BorderNormal)
	titleColor := lipgloss.Color(styles.TextBright)

	// Create centered empty text content
	// renderEmpty just returns the bordered box (no input, no typing indicator)
	// Layout: topBorder (1) + contentLines + bottomBorder (1) = height
	// Therefore: contentHeight = height - 2
	contentHeight := m.height - 2
	var contentLines []string

	// Add empty lines to center the text vertically
	emptyLinesBefore := (contentHeight - 1) / 2
	for i := 0; i < emptyLinesBefore; i++ {
		contentLines = append(contentLines, "")
	}

	// Add the centered empty text
	emptyTextStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Width(m.width - 4).
		Align(lipgloss.Center).
		Render(emptyText)
	contentLines = append(contentLines, emptyTextStyled)

	// Fill remaining space
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}

	// Create custom top border with embedded title
	title := " 💬 CONVERSATION "
	// Use lipgloss.Width() for accurate display width
	titleLen := lipgloss.Width(title)

	// Calculate remaining border width
	// Total: "┌─" (2) + title + dashes + "┐" (1) = m.width
	// Therefore: 2 + titleLen + dashCount + 1 = m.width
	// dashCount = m.width - titleLen - 3
	remainingWidth := m.width - titleLen - 3
	if remainingWidth < 0 {
		remainingWidth = 0
		if m.width < 10 {
			title = ""
			titleLen = 0
			remainingWidth = m.width - 3
		}
	}

	// Build top border with title
	topBorderLeft := lipgloss.NewStyle().Foreground(borderColor).Render("┌─")
	topBorderTitle := lipgloss.NewStyle().Foreground(titleColor).Render(title)
	topBorderRight := lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", remainingWidth) + "┐")
	topBorder := topBorderLeft + topBorderTitle + topBorderRight

	// Create side borders
	leftBorder := lipgloss.NewStyle().Foreground(borderColor).Render("│")
	rightBorder := lipgloss.NewStyle().Foreground(borderColor).Render("│")

	// Wrap content with side borders
	var borderedLines []string
	for _, line := range contentLines {
		lineWidth := lipgloss.Width(line)
		padding := ""
		if lineWidth < m.width-4 {
			padding = strings.Repeat(" ", m.width-4-lineWidth)
		}
		borderedLines = append(borderedLines, leftBorder+" "+line+padding+" "+rightBorder)
	}

	// Create bottom border
	bottomBorder := lipgloss.NewStyle().Foreground(borderColor).Render("└" + strings.Repeat("─", m.width-2) + "┘")

	// Combine all parts
	result := topBorder + "\n" + strings.Join(borderedLines, "\n") + "\n" + bottomBorder

	// DEBUG: Log first line to see if title is there
	f, _ := os.OpenFile("/tmp/ithil-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		firstLine := strings.Split(result, "\n")[0]
		fmt.Fprintf(f, "Conversation renderEmpty() details:\n")
		fmt.Fprintf(f, "  width=%d, height=%d\n", m.width, m.height)
		fmt.Fprintf(f, "  title=%q, titleLen=%d, remainingWidth=%d\n", title, titleLen, remainingWidth)
		fmt.Fprintf(f, "  topBorder=%q\n", topBorder)
		fmt.Fprintf(f, "  lipgloss.Width(topBorder)=%d\n", lipgloss.Width(topBorder))
		fmt.Fprintf(f, "  firstLine=%q\n", firstLine)
		fmt.Fprintf(f, "  Total lines in result=%d\n\n", len(strings.Split(result, "\n")))
		f.Close()
	}

	return result
}

// renderAllMessages renders all messages for the viewport.
func (m *ConversationModel) renderAllMessages() string {
	if len(m.messages) == 0 {
		return styles.DimStyle.Render("No messages yet. Start the conversation!")
	}

	// Get current user ID
	currentUserID := m.client.GetCurrentUserID()

	var messageViews []string
	for _, message := range m.messages {
		// Get sender name from cache for incoming messages
		senderName := ""
		if !message.IsOutgoing {
			if m.currentChat != nil && m.currentChat.Type == types.ChatTypePrivate {
				// For private chats, use the chat title as sender name
				senderName = m.currentChat.Title
			} else {
				// For group chats, try to get user name from cache
				if user, exists := m.cache.GetUser(message.SenderID); exists {
					senderName = user.GetDisplayName()
				}
				// If senderName is still empty (user not in cache or GetDisplayName returned empty),
				// fall back to "User ID" in message component by leaving senderName empty
			}
		}

		// Use viewport width for message rendering so messages use full available space
		messageComponent := components.NewMessageComponentWithUser(message, m.viewport.Width, currentUserID, senderName)
		messageViews = append(messageViews, messageComponent.Render())
	}

	return strings.Join(messageViews, "\n\n")
}

// updateViewport updates the viewport content and auto-scrolls to bottom.
func (m *ConversationModel) updateViewport() {
	content := m.renderAllMessages()
	m.viewport.SetContent(content)

	// Auto-scroll to bottom to show newest messages
	m.viewport.GotoBottom()
}

// renderTypingIndicator renders the typing indicator if someone is typing.
func (m *ConversationModel) renderTypingIndicator() string {
	typingUsers := m.GetTypingUsers()
	if len(typingUsers) == 0 {
		return ""
	}

	// Show typing indicator
	typingText := "typing..."
	return styles.TypingIndicatorStyle.Render(typingText)
}

// loadMessages loads messages for the current chat.
func (m *ConversationModel) loadMessages() {
	if m.currentChat == nil {
		m.messages = []*types.Message{}
		return
	}

	// Load messages from cache
	m.messages = m.cache.GetMessages(m.currentChat.ID)

	// TODO: Load more messages from server if needed
	// For now, we just use cached messages

	m.updateViewport()
}

// sendMessage sends the current input as a message or media.
func (m *ConversationModel) sendMessage() tea.Cmd {
	if m.currentChat == nil {
		return nil
	}

	// Check if we have an attachment
	hasAttachment := m.input.HasAttachment()
	hasText := !m.input.IsEmpty()

	// Must have either text or attachment
	if !hasAttachment && !hasText {
		return nil
	}

	text := m.input.GetValue()
	attachment := m.input.GetAttachment()
	currentChat := m.currentChat
	replyToID := int64(0)
	if m.replyToMessage != nil {
		replyToID = m.replyToMessage.ID
	}

	// Clear input and reply state immediately for better UX
	m.input.Clear()
	m.ClearReplyEdit()
	m.sendingMessage = true

	return func() tea.Msg {
		var message *types.Message
		var err error

		// Send media or text
		if hasAttachment {
			// Send media with optional caption
			message, err = m.client.SendMediaMessage(currentChat, attachment, text, replyToID)
		} else {
			// Send text message
			message, err = m.client.SendMessage(currentChat, text, replyToID)
		}

		if err != nil {
			return sendErrorMsg{error: err.Error()}
		}
		return messageSentMsg{
			chatID:  currentChat.ID,
			message: message,
		}
	}
}

// editMessage edits the current message.
func (m *ConversationModel) editMessage() tea.Cmd {
	if m.currentChat == nil || m.editingMessage == nil || m.input.IsEmpty() {
		return nil
	}

	text := m.input.GetValue()
	currentChat := m.currentChat
	messageID := m.editingMessage.ID

	// Clear input and edit state immediately
	m.input.Clear()
	m.ClearReplyEdit()

	return func() tea.Msg {
		err := m.client.EditMessage(currentChat, messageID, text)
		if err != nil {
			return sendErrorMsg{error: err.Error()}
		}
		return messageEditedMsg{
			chatID:    currentChat.ID,
			messageID: messageID,
			newText:   text,
		}
	}
}

// SetCurrentChat sets the current chat.
func (m *ConversationModel) SetCurrentChat(chat *types.Chat) {
	m.currentChat = chat
	// Clear input state and unfocus when switching chats
	m.input.Blur()
	m.input.Clear()
	m.ClearReplyEdit()
	m.loadMessages()
}

// SetMessages sets the messages for the current chat.
func (m *ConversationModel) SetMessages(messages []*types.Message) {
	m.messages = messages
	m.updateViewport()
}

// SetSize sets the size of the conversation view.
func (m *ConversationModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate viewport height
	// Layout breakdown (from renderConversationView):
	// - Top border with header: 1 line
	// - Viewport content with side borders: viewportHeight lines
	// - Bottom border: 1 line
	// - Typing indicator: 0-1 lines (reserve 1 to prevent layout shift)
	// - Input box: 4 lines (includes border only, no padding)
	// Total: 1 + viewportHeight + 1 + 1 + 4 = viewportHeight + 7
	// Therefore: viewportHeight = height - 7

	inputHeight := 4   // Input box total height (border + 3 content lines)
	borderLines := 2   // Top border (1) + Bottom border (1)
	typingHeight := 1  // Reserve space for typing indicator to prevent overflow

	viewportHeight := height - inputHeight - borderLines - typingHeight
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.viewport.Width = width - 4 // Account for borders and horizontal padding
	m.viewport.Height = viewportHeight
	m.input.SetWidth(width - 2)
	m.input.SetHeight(inputHeight)

	// Update media viewer size
	viewerWidth := width - 20
	viewerHeight := height - 10
	if viewerWidth < 60 {
		viewerWidth = 60
	}
	if viewerHeight < 30 {
		viewerHeight = 30
	}
	m.mediaViewer.SetSize(viewerWidth, viewerHeight)

	m.updateViewport()
}

// SetFocused sets the focused state.
func (m *ConversationModel) SetFocused(focused bool) {
	m.focused = focused
	// When pane loses focus, ensure input is also unfocused
	if !focused && m.input.Focused {
		m.input.Blur()
	}
}

// FocusInput focuses the input field.
func (m *ConversationModel) FocusInput() tea.Cmd {
	return m.input.Focus()
}

// SetReplyTo sets the message to reply to.
func (m *ConversationModel) SetReplyTo(message *types.Message) {
	m.replyToMessage = message
	m.editingMessage = nil
	m.input.SetReplyTo(message.ID)
}

// SetEditing sets the message to edit.
func (m *ConversationModel) SetEditing(message *types.Message) {
	m.editingMessage = message
	m.replyToMessage = nil
	m.input.SetEdit(message.ID, message.Content.Text)
}

// ClearReplyEdit clears reply and edit state.
func (m *ConversationModel) ClearReplyEdit() {
	m.replyToMessage = nil
	m.editingMessage = nil
	m.input.CancelReplyOrEdit()
}

// AddMessage adds a message to the conversation.
func (m *ConversationModel) AddMessage(message *types.Message) {
	m.messages = append(m.messages, message)
	m.updateViewport()
}

// UpdateMessage updates an existing message in the conversation.
func (m *ConversationModel) UpdateMessage(messageID int64, newText string) {
	for i, msg := range m.messages {
		if msg.ID == messageID {
			m.messages[i].Content.Text = newText
			m.messages[i].IsEdited = true
			m.updateViewport()
			break
		}
	}
}

// RemoveMessage removes a message from the conversation.
func (m *ConversationModel) RemoveMessage(messageID int64) {
	for i, msg := range m.messages {
		if msg.ID == messageID {
			m.messages = append(m.messages[:i], m.messages[i+1:]...)
			m.updateViewport()
			return
		}
	}
}

// MarkMessagesRead marks all messages up to maxID as read.
func (m *ConversationModel) MarkMessagesRead(maxID int64) {
	// Update visual read indicators
	// This could update message styling to show read receipts
	for _, msg := range m.messages {
		if msg.ID <= maxID && msg.IsOutgoing {
			// Mark as read (could update message component state)
			// For now, this is a placeholder for future read receipt visualization
		}
	}
}

// SetUserTyping sets a user as typing in the current conversation.
func (m *ConversationModel) SetUserTyping(userID int64) {
	if m.typingUsers == nil {
		m.typingUsers = make(map[int64]time.Time)
	}
	m.typingUsers[userID] = time.Now()
}

// GetTypingUsers returns the list of users currently typing.
func (m *ConversationModel) GetTypingUsers() []int64 {
	now := time.Now()
	var typing []int64
	for userID, startTime := range m.typingUsers {
		// Typing indicator expires after 5 seconds
		if now.Sub(startTime) < 5*time.Second {
			typing = append(typing, userID)
		} else {
			delete(m.typingUsers, userID)
		}
	}
	return typing
}

// openMediaViewer opens the media viewer for the last message with media.
func (m *ConversationModel) openMediaViewer() tea.Cmd {
	// Find the last message with media (going backwards)
	for i := len(m.messages) - 1; i >= 0; i-- {
		message := m.messages[i]
		if m.hasViewableMedia(message) {
			m.selectedMsgIdx = i
			mediaPath := ""
			if message.Content.Media != nil && message.Content.Media.IsDownloaded {
				mediaPath = message.Content.Media.LocalPath
			}
			return m.mediaViewer.ShowMedia(message, mediaPath)
		}
	}
	return nil
}

// hasViewableMedia checks if a message has viewable media.
func (m *ConversationModel) hasViewableMedia(message *types.Message) bool {
	switch message.Content.Type {
	case types.MessageTypePhoto, types.MessageTypeVideo, types.MessageTypeAudio,
		types.MessageTypeVoice, types.MessageTypeDocument:
		return message.Content.Media != nil
	default:
		return false
	}
}

// downloadMediaForViewer downloads media for the viewer.
func (m *ConversationModel) downloadMediaForViewer(message *types.Message) tea.Cmd {
	return func() tea.Msg {
		// Download the media using the telegram client
		// This is a placeholder - actual implementation depends on your media manager
		if m.currentChat == nil || message.Content.Media == nil {
			return components.MediaDownloadedMsg{Path: ""}
		}

		// TODO: Implement actual media download using MediaManager
		// For now, we'll just return the existing path if available
		if message.Content.Media.IsDownloaded && message.Content.Media.LocalPath != "" {
			return components.MediaDownloadedMsg{
				Path: message.Content.Media.LocalPath,
			}
		}

		// Return empty path to indicate download failure
		return components.MediaDownloadedMsg{Path: ""}
	}
}

// Messages for conversation updates.
type messagesUpdatedMsg struct {
	messages []*types.Message
}

type messageSentMsg struct {
	chatID  int64
	message *types.Message
}

type messageEditedMsg struct {
	chatID    int64
	messageID int64
	newText   string
}

type sendErrorMsg struct {
	error string
}

// fileAttachRequestMsg requests a file picker dialog
type fileAttachRequestMsg struct{}

// fileAttachedMsg indicates a file has been attached
type fileAttachedMsg struct {
	filePath string
}
