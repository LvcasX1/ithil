// Package models provides Bubbletea models for the Ithil TUI.
package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/cache"
	"github.com/lvcasx1/ithil/internal/media"
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
	input           *components.InputComponent
	replyToMessage  *types.Message                   // Message being replied to
	editingMessage  *types.Message                   // Message being edited
	sendingMessage  bool                             // Flag to show sending state
	typingUsers     map[int64]time.Time              // userID -> when they started typing
	mediaViewer     *components.MediaViewerComponent // Media viewer modal
	selectedMsgIdx  int                              // Index of selected message for navigation and media viewing
	messageLinesMap  map[int]int                      // Maps message index to starting line position
	deleteModal      *components.ModalComponent       // Delete confirmation modal
	pendingDeleteID  int64                            // ID of message pending deletion
	forwardMode       bool                             // Forward mode active
	forwardMessageID  int64                            // ID of message to forward
	forwardModal      *components.ModalComponent       // Forward chat selector modal
	reactionMode      bool                             // Reaction picker mode active
	reactionMessageID int64                            // ID of message to react to
	reactionModal     *components.ModalComponent       // Reaction picker modal
}

// NewConversationModel creates a new conversation model.
func NewConversationModel(client *telegram.Client, cache *cache.Cache) *ConversationModel {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = false

	return &ConversationModel{
		client:          client,
		cache:           cache,
		messages:        []*types.Message{},
		viewport:        vp,
		input:           components.NewInputComponent(100, 5),
		typingUsers:     make(map[int64]time.Time),
		mediaViewer:     components.NewMediaViewerComponent(80, 30),
		selectedMsgIdx:  -1,
		messageLinesMap:  make(map[int]int),
		deleteModal:       components.NewModalComponent("Confirm Delete", 50, 8),
		pendingDeleteID:   0,
		forwardMode:       false,
		forwardMessageID:  0,
		forwardModal:      components.NewModalComponent("Forward Message", 60, 15),
		reactionMode:      false,
		reactionMessageID: 0,
		reactionModal:     components.NewModalComponent("React to Message", 50, 12),
	}
}

// Init initializes the conversation model.
func (m *ConversationModel) Init() tea.Cmd {
	return m.input.Init()
}

// Update handles conversation model updates.
func (m *ConversationModel) Update(msg tea.Msg) (*ConversationModel, tea.Cmd) {
	var cmd tea.Cmd

	// If reaction modal is visible, handle it first
	if m.reactionModal.IsVisible() {
		return m.handleReactionModal(msg)
	}

	// If forward modal is visible, handle it first
	if m.forwardModal.IsVisible() {
		return m.handleForwardModal(msg)
	}

	// If delete modal is visible, handle it first
	if m.deleteModal.IsVisible() {
		return m.handleDeleteModal(msg)
	}

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
			case "up":
				// Select previous message (single step for precision)
				return m, m.selectPreviousMessage(1)
			case "down":
				// Select next message (single step for precision)
				return m, m.selectNextMessage(1)
			case "k":
				// Fast navigation - skip 3 messages up
				return m, m.selectPreviousMessage(3)
			case "j":
				// Fast navigation - skip 3 messages down
				return m, m.selectNextMessage(3)
			case "ctrl+u":
				// Scroll up half page (faster navigation)
				m.viewport.HalfViewUp()
				return m, nil
			case "ctrl+d":
				// Scroll down half page (faster navigation)
				m.viewport.HalfViewDown()
				return m, nil
			case "pgup", "ctrl+b":
				m.viewport.ViewUp()
				return m, nil
			case "pgdown", "ctrl+f":
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
				// Reply to selected message if available, otherwise last message
				if m.selectedMsgIdx >= 0 && m.selectedMsgIdx < len(m.messages) {
					m.SetReplyTo(m.messages[m.selectedMsgIdx])
					return m, m.input.Focus()
				} else if len(m.messages) > 0 {
					m.SetReplyTo(m.messages[len(m.messages)-1])
					return m, m.input.Focus()
				}
				return m, nil
			case "e":
				// Edit selected outgoing message if available, otherwise last outgoing message
				if m.selectedMsgIdx >= 0 && m.selectedMsgIdx < len(m.messages) && m.messages[m.selectedMsgIdx].IsOutgoing {
					m.SetEditing(m.messages[m.selectedMsgIdx])
					return m, m.input.Focus()
				} else {
					// Fallback: find last outgoing message
					for i := len(m.messages) - 1; i >= 0; i-- {
						if m.messages[i].IsOutgoing {
							m.SetEditing(m.messages[i])
							return m, m.input.Focus()
						}
					}
				}
				return m, nil
			case "enter":
				// Open media viewer for selected message if available
				return m, m.openMediaViewer()
			case "d":
				// Delete selected message if available
				if m.selectedMsgIdx >= 0 && m.selectedMsgIdx < len(m.messages) {
					msg := m.messages[m.selectedMsgIdx]
					// Only allow deleting outgoing messages
					if msg.IsOutgoing {
						m.pendingDeleteID = msg.ID
						m.deleteModal.SetContent(fmt.Sprintf("Are you sure you want to delete this message?\n\nPress 'y' to confirm or ESC to cancel"))
						m.deleteModal.Show()
						return m, nil
					}
				}
				return m, nil
			case "f":
				// Forward selected message if available
				if m.selectedMsgIdx >= 0 && m.selectedMsgIdx < len(m.messages) {
					msg := m.messages[m.selectedMsgIdx]
					m.forwardMessageID = msg.ID
					return m, m.showForwardChatSelector()
				}
				return m, nil
			case "x":
				// React to selected message if available
				if m.selectedMsgIdx >= 0 && m.selectedMsgIdx < len(m.messages) {
					msg := m.messages[m.selectedMsgIdx]
					m.reactionMessageID = msg.ID
					return m, m.showReactionPicker()
				}
				return m, nil
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

	case messageDeletedMsg:
		// Remove message from local state
		m.RemoveMessage(msg.messageID)
		m.pendingDeleteID = 0
		return m, nil

	case messageForwardedMsg:
		// Message forwarded successfully
		// TODO: Show success feedback to user
		return m, nil

	case messageReactedMsg:
		// Reaction sent successfully
		// TODO: Show success feedback to user
		return m, nil

	case sendErrorMsg:
		// TODO: Show error to user (could use a status bar or notification)
		return m, nil
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

	// If reaction modal is visible, overlay it
	if m.reactionModal.IsVisible() {
		modalView := m.reactionModal.View()
		// Center the modal
		overlay := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
		return overlay
	}

	// If forward modal is visible, overlay it
	if m.forwardModal.IsVisible() {
		modalView := m.forwardModal.View()
		// Center the modal
		overlay := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
		return overlay
	}

	// If delete modal is visible, overlay it
	if m.deleteModal.IsVisible() {
		modalView := m.deleteModal.View()
		// Center the modal
		overlay := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
		return overlay
	}

	// If media viewer is visible, overlay it
	if m.mediaViewer.IsVisible() {
		// For fullscreen mode, return raw view (though MainModel should handle this)
		if m.mediaViewer.IsFullscreenMode() {
			return m.mediaViewer.ViewFullscreen()
		}

		// For modal mode, center it with Lipgloss
		viewerView := m.mediaViewer.View()
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
	// Use lipgloss.Width() for accurate display width (handles emojis, wide chars, etc.)
	titleLen := lipgloss.Width(title)

	// Truncate title if too long
	maxTitleLen := m.width - 6 // Reserve space for "┌─", "─┐", and some dashes
	if titleLen > maxTitleLen {
		if maxTitleLen > 7 {
			// Truncate by gradually removing runes until display width fits
			titleRunes := []rune(m.currentChat.Title)
			for len(titleRunes) > 0 && lipgloss.Width(" "+string(titleRunes)+"... ") > maxTitleLen {
				titleRunes = titleRunes[:len(titleRunes)-1]
			}
			title = " " + string(titleRunes) + "... "
			titleLen = lipgloss.Width(title)
		} else {
			title = " ... "
			titleLen = lipgloss.Width(title)
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

	// Reset the message lines map
	m.messageLinesMap = make(map[int]int)

	var messageViews []string
	currentLine := 0

	for i, message := range m.messages {
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

		// Check if this message is selected
		isSelected := i == m.selectedMsgIdx

		// Store the starting line for this message
		m.messageLinesMap[i] = currentLine

		// Use viewport width for message rendering so messages use full available space
		messageComponent := components.NewMessageComponentWithSelection(message, m.viewport.Width, currentUserID, senderName, isSelected)
		renderedMessage := messageComponent.Render()
		messageViews = append(messageViews, renderedMessage)

		// Count the lines in the rendered message
		messageLineCount := strings.Count(renderedMessage, "\n") + 1
		currentLine += messageLineCount

		// Add 2 lines for the spacing between messages ("\n\n")
		if i < len(m.messages)-1 {
			currentLine += 2
		}
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
		m.selectedMsgIdx = -1
		return
	}

	// Load messages from cache
	m.messages = m.cache.GetMessages(m.currentChat.ID)

	// Initialize selection to last message
	if len(m.messages) > 0 {
		m.selectedMsgIdx = len(m.messages) - 1
	} else {
		m.selectedMsgIdx = -1
	}

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
	// Reset selection to last message
	if len(m.messages) > 0 {
		m.selectedMsgIdx = len(m.messages) - 1
	} else {
		m.selectedMsgIdx = -1
	}
}

// SetMessages sets the messages for the current chat.
func (m *ConversationModel) SetMessages(messages []*types.Message) {
	m.messages = messages
	// Keep selection valid, default to last message if invalid
	if m.selectedMsgIdx < 0 || m.selectedMsgIdx >= len(m.messages) {
		if len(m.messages) > 0 {
			m.selectedMsgIdx = len(m.messages) - 1
		} else {
			m.selectedMsgIdx = -1
		}
	}
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
	// Keep selection on the same message by index (it will shift as new messages arrive)
	// If selection is at last message, move it to the new last message
	if m.selectedMsgIdx == len(m.messages)-2 {
		m.selectedMsgIdx = len(m.messages) - 1
	}
	// Ensure selection is valid
	if m.selectedMsgIdx < 0 || m.selectedMsgIdx >= len(m.messages) {
		m.selectedMsgIdx = len(m.messages) - 1
	}
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

// openMediaViewer opens the media viewer for the selected message or last message with media.
func (m *ConversationModel) openMediaViewer() tea.Cmd {
	// If a message is selected and has viewable media, open it
	if m.selectedMsgIdx >= 0 && m.selectedMsgIdx < len(m.messages) {
		message := m.messages[m.selectedMsgIdx]
		if m.hasViewableMedia(message) {
			mediaPath := ""
			if message.Content.Media != nil && message.Content.Media.IsDownloaded {
				// DEFENSIVE CHECK: Verify the file actually exists
				localPath := message.Content.Media.LocalPath

				if localPath != "" {
					if _, err := os.Stat(localPath); err == nil {
						// File exists, use the path
						mediaPath = localPath
					} else {
						// File doesn't exist or error, pass empty string to trigger download
						mediaPath = ""
					}
				}
			}
			return m.mediaViewer.ShowMedia(message, mediaPath)
		}
		// Selected message has no viewable media - do nothing (user feedback could be added here)
		return nil
	}

	// Fallback: Find the last message with media (going backwards)
	for i := len(m.messages) - 1; i >= 0; i-- {
		message := m.messages[i]
		if m.hasViewableMedia(message) {
			m.selectedMsgIdx = i
			mediaPath := ""
			if message.Content.Media != nil && message.Content.Media.IsDownloaded {
				// DEFENSIVE CHECK: Verify the file actually exists
				localPath := message.Content.Media.LocalPath

				if localPath != "" {
					if _, err := os.Stat(localPath); err == nil {
						// File exists, use the path
						mediaPath = localPath
					} else {
						// File doesn't exist or error, pass empty string to trigger download
						mediaPath = ""
					}
				}
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
		if m.currentChat == nil || message.Content.Media == nil {
			return components.MediaDownloadedMsg{Path: ""}
		}

		// Use the client's DownloadMedia method
		localPath, err := m.client.DownloadMedia(message)
		if err != nil {
			m.client.GetLogger().Error("Failed to download media for viewer", "error", err)
			return components.MediaDownloadedMsg{Path: ""}
		}

		// Update the message in cache with the downloaded path
		if m.cache != nil {
			message.Content.Media.LocalPath = localPath
			message.Content.Media.IsDownloaded = true
			m.cache.SetMessage(m.currentChat.ID, message)
		}

		return components.MediaDownloadedMsg{
			Path: localPath,
		}
	}
}

// selectPreviousMessage selects the previous message in the conversation.
// The step parameter controls how many messages to skip (e.g., 1 for single step, 3 for fast navigation).
func (m *ConversationModel) selectPreviousMessage(step int) tea.Cmd {
	if len(m.messages) == 0 {
		return nil
	}

	oldIdx := m.selectedMsgIdx

	// If no selection, start at last message
	if m.selectedMsgIdx < 0 {
		m.selectedMsgIdx = len(m.messages) - 1
	} else {
		// Move up by 'step' messages
		m.selectedMsgIdx -= step

		// Handle wrapping
		if m.selectedMsgIdx < 0 {
			m.selectedMsgIdx = 0
		}
	}

	// Only re-render if selection actually changed
	if oldIdx != m.selectedMsgIdx {
		m.updateViewport()
		m.scrollToSelectedMessage()
	}
	return nil
}

// selectNextMessage selects the next message in the conversation.
// The step parameter controls how many messages to skip (e.g., 1 for single step, 3 for fast navigation).
func (m *ConversationModel) selectNextMessage(step int) tea.Cmd {
	if len(m.messages) == 0 {
		return nil
	}

	oldIdx := m.selectedMsgIdx

	// If no selection, start at first message
	if m.selectedMsgIdx < 0 {
		m.selectedMsgIdx = 0
	} else {
		// Move down by 'step' messages
		m.selectedMsgIdx += step

		// Handle wrapping
		if m.selectedMsgIdx >= len(m.messages) {
			m.selectedMsgIdx = len(m.messages) - 1
		}
	}

	// Only re-render if selection actually changed
	if oldIdx != m.selectedMsgIdx {
		m.updateViewport()
		m.scrollToSelectedMessage()
	}
	return nil
}

// scrollToSelectedMessage scrolls the viewport to keep the selected message centered.
// Uses actual rendered line positions from messageLinesMap for accurate scrolling.
// Keeps the selected message in the middle of the viewport for best visibility.
func (m *ConversationModel) scrollToSelectedMessage() {
	if m.selectedMsgIdx < 0 || m.selectedMsgIdx >= len(m.messages) {
		return
	}

	// Get the actual starting line position of the selected message
	messageStartLine, exists := m.messageLinesMap[m.selectedMsgIdx]
	if !exists {
		// Fallback: rebuild viewport if message position not found
		m.updateViewport()
		messageStartLine, exists = m.messageLinesMap[m.selectedMsgIdx]
		if !exists {
			return
		}
	}

	// Strategy: Keep the selected message centered in the viewport
	// This makes it always visible and easy to see what's selected

	// Calculate where the message should be positioned (centered)
	targetOffset := messageStartLine - (m.viewport.Height / 2)

	// Don't scroll past the beginning
	if targetOffset < 0 {
		targetOffset = 0
	}

	// Get total content height
	totalLines := 0
	if len(m.messageLinesMap) > 0 {
		// Find the last message's position
		lastMsgIdx := len(m.messages) - 1
		if lastPos, hasLast := m.messageLinesMap[lastMsgIdx]; hasLast {
			// Estimate total height (last message start + some buffer for the message itself)
			totalLines = lastPos + 15 // Add buffer for last message height
		}
	}

	// Don't scroll past the end (if content is shorter than viewport, stay at top)
	maxOffset := totalLines - m.viewport.Height
	if maxOffset < 0 {
		maxOffset = 0
	}

	if targetOffset > maxOffset {
		targetOffset = maxOffset
	}

	// Apply the new offset
	m.viewport.SetYOffset(targetOffset)
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

// messageDeletedMsg indicates a message was deleted
type messageDeletedMsg struct {
	chatID    int64
	messageID int64
}

// handleDeleteModal handles delete modal interactions
func (m *ConversationModel) handleDeleteModal(msg tea.Msg) (*ConversationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Confirm deletion
			m.deleteModal.Hide()
			return m, m.deleteMessage(m.pendingDeleteID)
		case "esc", "n", "N":
			// Cancel deletion
			m.deleteModal.Hide()
			m.pendingDeleteID = 0
			return m, nil
		}
	}
	return m, nil
}

// deleteMessage deletes a message from the current chat
func (m *ConversationModel) deleteMessage(messageID int64) tea.Cmd {
	if m.currentChat == nil || messageID == 0 {
		return nil
	}

	currentChat := m.currentChat
	return func() tea.Msg {
		err := m.client.DeleteMessage(currentChat, messageID)
		if err != nil {
			return sendErrorMsg{error: fmt.Sprintf("Failed to delete message: %v", err)}
		}
		return messageDeletedMsg{
			chatID:    currentChat.ID,
			messageID: messageID,
		}
	}
}

// showForwardChatSelector shows the forward chat selector modal
func (m *ConversationModel) showForwardChatSelector() tea.Cmd {
	// Get all chats from cache
	allChats := m.cache.GetChats()
	if len(allChats) == 0 {
		m.forwardModal.SetContent("No chats available to forward to.\n\nPress ESC to cancel")
		m.forwardModal.Show()
		return nil
	}

	// Build chat list display (show first 9 chats with numbers)
	var chatList strings.Builder
	chatList.WriteString("Select a chat to forward to:\n\n")

	maxChats := 9
	if len(allChats) > maxChats {
		allChats = allChats[:maxChats]
	}

	for i, chat := range allChats {
		chatList.WriteString(fmt.Sprintf("%d. %s\n", i+1, chat.Title))
	}

	chatList.WriteString("\nPress 1-9 to select, or ESC to cancel")

	m.forwardModal.SetContent(chatList.String())
	m.forwardModal.Show()
	m.forwardMode = true

	return nil
}

// handleForwardModal handles forward modal interactions
func (m *ConversationModel) handleForwardModal(msg tea.Msg) (*ConversationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Select chat by number
			chatIdx := int(msg.String()[0] - '1') // Convert '1'-'9' to 0-8

			allChats := m.cache.GetChats()
			if chatIdx >= 0 && chatIdx < len(allChats) && chatIdx < 9 {
				toChat := allChats[chatIdx]
				m.forwardModal.Hide()
				m.forwardMode = false
				return m, m.forwardMessage(toChat)
			}
			return m, nil
		case "esc":
			// Cancel forward
			m.forwardModal.Hide()
			m.forwardMode = false
			m.forwardMessageID = 0
			return m, nil
		}
	}
	return m, nil
}

// forwardMessage forwards a message to another chat
func (m *ConversationModel) forwardMessage(toChat *types.Chat) tea.Cmd {
	if m.currentChat == nil || m.forwardMessageID == 0 {
		return nil
	}

	fromChat := m.currentChat
	messageID := m.forwardMessageID
	m.forwardMessageID = 0

	return func() tea.Msg {
		err := m.client.ForwardMessage(fromChat, toChat, []int64{messageID})
		if err != nil {
			return sendErrorMsg{error: fmt.Sprintf("Failed to forward message: %v", err)}
		}
		return messageForwardedMsg{
			toChat: toChat.Title,
		}
	}
}

// messageForwardedMsg indicates a message was forwarded
type messageForwardedMsg struct {
	toChat string
}

// showReactionPicker shows the reaction picker modal
func (m *ConversationModel) showReactionPicker() tea.Cmd {
	// Common emoji reactions
	reactions := []string{
		"1. 👍 Thumbs up",
		"2. 👎 Thumbs down",
		"3. ❤️ Heart",
		"4. 🔥 Fire",
		"5. 👏 Clap",
		"6. 😂 Laugh",
		"7. 😮 Surprised",
		"8. 😢 Sad",
	}

	var pickerContent strings.Builder
	pickerContent.WriteString("Select a reaction:\n\n")
	for _, reaction := range reactions {
		pickerContent.WriteString(reaction + "\n")
	}
	pickerContent.WriteString("\nPress 1-8 to select, or ESC to cancel")

	m.reactionModal.SetContent(pickerContent.String())
	m.reactionModal.Show()
	m.reactionMode = true

	return nil
}

// handleReactionModal handles reaction modal interactions
func (m *ConversationModel) handleReactionModal(msg tea.Msg) (*ConversationModel, tea.Cmd) {
	// Map of number keys to emoji
	reactionMap := map[string]string{
		"1": "👍",
		"2": "👎",
		"3": "❤️",
		"4": "🔥",
		"5": "👏",
		"6": "😂",
		"7": "😮",
		"8": "😢",
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5", "6", "7", "8":
			// Select reaction by number
			emoji, exists := reactionMap[msg.String()]
			if exists {
				m.reactionModal.Hide()
				m.reactionMode = false
				return m, m.reactToMessage(emoji)
			}
			return m, nil
		case "esc":
			// Cancel reaction
			m.reactionModal.Hide()
			m.reactionMode = false
			m.reactionMessageID = 0
			return m, nil
		}
	}
	return m, nil
}

// reactToMessage sends a reaction to a message
func (m *ConversationModel) reactToMessage(emoji string) tea.Cmd {
	if m.currentChat == nil || m.reactionMessageID == 0 {
		return nil
	}

	currentChat := m.currentChat
	messageID := m.reactionMessageID
	m.reactionMessageID = 0

	return func() tea.Msg {
		err := m.client.ReactToMessage(currentChat, messageID, emoji)
		if err != nil {
			return sendErrorMsg{error: fmt.Sprintf("Failed to react to message: %v", err)}
		}
		return messageReactedMsg{
			reaction: emoji,
		}
	}
}

// messageReactedMsg indicates a reaction was sent
type messageReactedMsg struct {
	reaction string
}

// IsMediaViewerFullscreen returns true if the media viewer is visible and in fullscreen mode.
// This is used by the main model to short-circuit normal rendering.
func (m *ConversationModel) IsMediaViewerFullscreen() bool {
	return m.mediaViewer.IsVisible() && m.mediaViewer.IsFullscreenMode()
}

// GetMediaViewerFullscreenView returns the raw fullscreen view from the media viewer.
// This should only be called when IsMediaViewerFullscreen() returns true.
func (m *ConversationModel) GetMediaViewerFullscreenView() string {
	return m.mediaViewer.ViewFullscreen()
}

// SetExternalAudioPlayer sets the external audio player on the media viewer for background playback.
func (m *ConversationModel) SetExternalAudioPlayer(player *media.AudioPlayer) {
	m.mediaViewer.SetExternalAudioPlayer(player)
}
