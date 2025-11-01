// Package models provides Bubbletea models for the Ithil TUI.
package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/app"
	"github.com/lvcasx1/ithil/internal/cache"
	"github.com/lvcasx1/ithil/internal/telegram"
	"github.com/lvcasx1/ithil/internal/ui/components"
	"github.com/lvcasx1/ithil/internal/ui/keys"
	"github.com/lvcasx1/ithil/pkg/types"
)

// FocusPane represents which pane is currently focused.
type FocusPane int

const (
	FocusChatList FocusPane = iota
	FocusConversation
	FocusSidebar
)

// MainModel is the root Bubbletea model for the application.
type MainModel struct {
	config       *app.Config
	client       *telegram.Client
	cache        *cache.Cache
	keyMap       keys.KeyMap
	width        int
	height       int
	ready        bool
	authenticated bool
	authChecked  bool  // Track if we've checked auth status yet

	// Sub-models
	auth         *AuthModel
	chatList     *ChatListModel
	conversation *ConversationModel
	sidebar      *SidebarModel
	statusBar    *components.StatusBarComponent

	// State
	focusPane    FocusPane
	currentChat  *types.Chat
	showHelp     bool
	errorMessage string
}

// NewMainModel creates a new main model.
func NewMainModel(config *app.Config, client *telegram.Client) *MainModel {
	cache := cache.New(config.Cache.MaxMessagesPerChat)

	// Set the cache on the client so the update handler can cache users
	client.SetCache(cache)

	// Create key map based on config
	var keyMap keys.KeyMap
	if config.UI.Keyboard.VimMode {
		keyMap = keys.VimKeyMap()
	} else {
		keyMap = keys.DefaultKeyMap()
	}

	// Don't check auth status here - it's not ready yet (client connecting asynchronously)
	// Will be set via checkAuthStatus() in Init()

	return &MainModel{
		config:       config,
		client:       client,
		cache:        cache,
		keyMap:       keyMap,
		authenticated: false,  // Will be updated via authStateMsg
		authChecked:  false,   // Will be set to true when we get auth status
		focusPane:    FocusChatList,

		// Initialize sub-models
		auth:         NewAuthModel(client),
		chatList:     NewChatListModel(cache),
		conversation: NewConversationModel(client, cache),
		sidebar:      NewSidebarModel(cache),
		statusBar:    components.NewStatusBarComponent(100),
	}
}

// Init initializes the main model.
func (m *MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.auth.Init(),
		m.chatList.Init(),
		m.conversation.Init(),
		m.sidebar.Init(),
		m.subscribeToAuthStateChanges(),  // Listen for auth state changes from client
		m.tickStatusBar(),
		// NOTE: subscribeToUpdates() is called when authenticated (see authStateMsg handler)
	)
}

// Update handles main model updates.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "ctrl+s":
			m.sidebar.ToggleVisible()
			return m, m.recalculateLayout()
		case "S":
			// Toggle stealth mode
			m.config.Privacy.StealthMode = !m.config.Privacy.StealthMode
			// Update status bar or show notification
			stealthStatus := "disabled"
			if m.config.Privacy.StealthMode {
				stealthStatus = "enabled"
			}
			m.statusBar.SetMessage("Stealth mode " + stealthStatus)
			// Clear message after 3 seconds
			return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
				return clearStatusMsg{}
			})
		}

		// If not authenticated, let auth model handle it
		if !m.authenticated {
			var cmd tea.Cmd
			m.auth, cmd = m.auth.Update(msg)
			return m, cmd
		}

		// Handle pane switching
		switch msg.String() {
		case "tab":
			m.nextPane()
			return m, nil
		case "shift+tab":
			m.prevPane()
			return m, nil
		case "ctrl+1":
			m.focusPane = FocusChatList
			m.updatePaneFocus()
			return m, nil
		case "ctrl+2":
			m.focusPane = FocusConversation
			m.updatePaneFocus()
			return m, nil
		case "ctrl+3":
			if m.sidebar.IsVisible() {
				m.focusPane = FocusSidebar
				m.updatePaneFocus()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, m.recalculateLayout()

	case authStateMsg:
		m.authChecked = true  // Mark that we've received auth status
		if msg.state == types.AuthStateReady {
			m.authenticated = true
			m.statusBar.SetConnectionStatus(components.StatusConnected)
			// Start update subscription when authenticated (runs in background goroutine)
			return m, tea.Batch(m.loadChats(), m.subscribeToAuthStateChanges(), m.waitForUpdate())
		}
		m.authenticated = false  // Explicitly set to false for unauthenticated states
		var cmd tea.Cmd
		m.auth, cmd = m.auth.Update(msg)
		return m, tea.Batch(cmd, m.subscribeToAuthStateChanges())

	case chatSelectedMsg:
		m.currentChat = msg.chat
		m.statusBar.SetCurrentChat(msg.chat.Title)
		m.focusPane = FocusConversation
		m.updatePaneFocus()

		// Update all sub-models
		m.conversation.SetCurrentChat(msg.chat)
		m.sidebar.SetCurrentChat(msg.chat)

		// Mark chat as read and clear unread count (only if stealth mode is disabled)
		if msg.chat.UnreadCount > 0 && !m.config.Privacy.StealthMode {
			msg.chat.UnreadCount = 0
			m.cache.SetChat(msg.chat)
			m.chatList.UpdateChat(msg.chat)
			m.updateUnreadCount()

			// Mark as read on server (async)
			go func() {
				// Silently mark as read - errors are not critical for UX
				_ = m.client.MarkChatAsRead(msg.chat)
			}()
		}

		// Load messages for the selected chat
		return m, m.loadMessages(msg.chat)

	case chatsLoadedMsg:
		m.chatList.SetChats(msg.chats)
		m.updateUnreadCount()
		return m, nil

	case messagesLoadedMsg:
		m.conversation.SetMessages(msg.messages)
		return m, nil

	case messageSentMsg:
		// Add the sent message to conversation
		m.conversation.AddMessage(msg.message)
		// Update cache
		m.cache.AddMessage(msg.chatID, msg.message)
		// Update chat last message in cache
		if chat, exists := m.cache.GetChat(msg.chatID); exists {
			chat.LastMessage = msg.message
			m.cache.SetChat(chat)
		}
		return m, nil

	case messageEditedMsg:
		// Update the message in conversation
		m.conversation.UpdateMessage(msg.messageID, msg.newText)
		// Update cache
		if cachedMsg, exists := m.cache.GetMessage(msg.chatID, msg.messageID); exists {
			cachedMsg.Content.Text = msg.newText
			cachedMsg.IsEdited = true
		}
		return m, nil

	case sendErrorMsg:
		m.errorMessage = msg.error
		// Show error to user (could add error display in UI)
		return m, nil

	case tickMsg:
		return m, m.tickStatusBar()

	case renderUpdateMsg:
		// Force all sub-models to update by sending them the message
		var cmd tea.Cmd
		m.chatList, cmd = m.chatList.Update(msg)
		m.conversation, cmd = m.conversation.Update(msg)
		m.sidebar, cmd = m.sidebar.Update(msg)
		// Return to trigger a full re-render
		return m, cmd

	case telegramUpdateMsg:
		// Process the update - returns commands for UI updates
		updateCmds := m.processUpdate(msg.update)

		// Immediately spawn another goroutine to wait for the next update
		// This ensures we're always listening without blocking
		updateCmds = append(updateCmds, m.waitForUpdate())

		return m, tea.Batch(updateCmds...)

	case clearStatusMsg:
		m.statusBar.ClearMessage()
		return m, nil
	}

	// Update sub-models based on authentication state
	if !m.authenticated {
		var cmd tea.Cmd
		m.auth, cmd = m.auth.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		// Update all panes
		var cmd tea.Cmd

		m.chatList, cmd = m.chatList.Update(msg)
		cmds = append(cmds, cmd)

		m.conversation, cmd = m.conversation.Update(msg)
		cmds = append(cmds, cmd)

		m.sidebar, cmd = m.sidebar.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the main model.
func (m *MainModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Wait for auth status check before deciding what to show
	if !m.authChecked {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Connecting to Telegram...")
	}

	// If not authenticated, show auth screen
	if !m.authenticated {
		return m.auth.View()
	}

	// Show main UI with three panes
	return m.renderMainUI()
}

// renderMainUI renders the main three-pane UI.
func (m *MainModel) renderMainUI() string {
	// Render panes
	chatListView := m.chatList.View()
	conversationView := m.conversation.View()
	sidebarView := ""
	if m.sidebar.IsVisible() {
		sidebarView = m.sidebar.View()
	}

	// Combine panes horizontally
	var content string
	if m.sidebar.IsVisible() {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			chatListView,
			conversationView,
			sidebarView,
		)
	} else {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			chatListView,
			conversationView,
		)
	}

	// Status bar
	statusBar := m.statusBar.Render()

	// Combine vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		statusBar,
	)
}

// recalculateLayout recalculates the layout of all panes.
func (m *MainModel) recalculateLayout() tea.Cmd {
	if !m.ready {
		return nil
	}

	// Calculate pane dimensions
	chatListWidth := m.width * m.config.UI.Layout.ChatListWidth / 100
	conversationWidth := m.width * m.config.UI.Layout.ConversationWidth / 100
	sidebarWidth := m.width * m.config.UI.Layout.InfoWidth / 100

	if !m.sidebar.IsVisible() {
		conversationWidth = m.width - chatListWidth
		sidebarWidth = 0
	}

	statusBarHeight := 1
	contentHeight := m.height - statusBarHeight

	// Update all sub-models
	m.auth.SetSize(m.width, m.height)
	m.chatList.SetSize(chatListWidth, contentHeight)
	m.conversation.SetSize(conversationWidth, contentHeight)
	m.sidebar.SetSize(sidebarWidth, contentHeight)
	m.statusBar.SetWidth(m.width)

	return nil
}

// updatePaneFocus updates the focus state of all panes.
func (m *MainModel) updatePaneFocus() {
	m.chatList.SetFocused(m.focusPane == FocusChatList)
	m.conversation.SetFocused(m.focusPane == FocusConversation)
	m.sidebar.SetFocused(m.focusPane == FocusSidebar)
}

// nextPane moves focus to the next pane.
func (m *MainModel) nextPane() {
	switch m.focusPane {
	case FocusChatList:
		m.focusPane = FocusConversation
	case FocusConversation:
		if m.sidebar.IsVisible() {
			m.focusPane = FocusSidebar
		} else {
			m.focusPane = FocusChatList
		}
	case FocusSidebar:
		m.focusPane = FocusChatList
	}
	m.updatePaneFocus()
}

// prevPane moves focus to the previous pane.
func (m *MainModel) prevPane() {
	switch m.focusPane {
	case FocusChatList:
		if m.sidebar.IsVisible() {
			m.focusPane = FocusSidebar
		} else {
			m.focusPane = FocusConversation
		}
	case FocusConversation:
		m.focusPane = FocusChatList
	case FocusSidebar:
		m.focusPane = FocusConversation
	}
	m.updatePaneFocus()
}

// checkAuthStatus checks the authentication status.
func (m *MainModel) checkAuthStatus() tea.Cmd {
	return func() tea.Msg {
		if m.client.IsAuthenticated() {
			return authStateMsg{state: types.AuthStateReady}
		}
		return authStateMsg{state: m.client.GetAuthState()}
	}
}

// loadChats loads the chat list.
func (m *MainModel) loadChats() tea.Cmd {
	return func() tea.Msg {
		chats, err := m.client.GetChats(100)
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		// Cache the chats
		for _, chat := range chats {
			m.cache.SetChat(chat)
		}

		return chatsLoadedMsg{chats: chats}
	}
}

// updateUnreadCount updates the total unread count in the status bar.
func (m *MainModel) updateUnreadCount() {
	totalUnread := 0
	chats := m.cache.GetChats()
	for _, chat := range chats {
		totalUnread += chat.UnreadCount
	}
	m.statusBar.SetUnreadCount(totalUnread)
}

// tickStatusBar creates a tick message for status bar updates.
func (m *MainModel) tickStatusBar() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// loadMessages loads messages for the selected chat.
func (m *MainModel) loadMessages(chat *types.Chat) tea.Cmd {
	return func() tea.Msg {
		messages, err := m.client.GetMessages(chat, 50, 0)
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		// Cache the messages
		for _, msg := range messages {
			m.cache.AddMessage(chat.ID, msg)
		}

		return messagesLoadedMsg{chatID: chat.ID, messages: messages}
	}
}

// waitForUpdate waits for a single update and returns it as a command.
// This runs in its own goroutine and doesn't block the UI.
func (m *MainModel) waitForUpdate() tea.Cmd {
	return func() tea.Msg {
		// DEBUG: Log subscription attempt
		if f, err := os.OpenFile("/tmp/ithil-updates.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "[%s] Waiting for update in goroutine...\n", time.Now().Format("15:04:05"))
			f.Close()
		}

		updateChan := m.client.Updates()

		// This blocks in its own goroutine, not the main event loop
		update, ok := <-updateChan
		if !ok {
			// Channel closed
			if f, err := os.OpenFile("/tmp/ithil-updates.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				fmt.Fprintf(f, "[%s] Update channel closed\n", time.Now().Format("15:04:05"))
				f.Close()
			}
			return nil
		}

		// DEBUG: Log received update
		if f, err := os.OpenFile("/tmp/ithil-updates.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "[%s] Update received: type=%d, chatID=%d\n", time.Now().Format("15:04:05"), update.Type, update.ChatID)
			f.Close()
		}

		return telegramUpdateMsg{update: update}
	}
}

// subscribeToAuthStateChanges subscribes to auth state changes.
func (m *MainModel) subscribeToAuthStateChanges() tea.Cmd {
	return func() tea.Msg {
		// Wait for next auth state change from Telegram client
		authStateChan := m.client.AuthStateChanges()

		// This will block until an auth state change arrives
		state, ok := <-authStateChan
		if !ok {
			// Channel closed, stop listening
			return nil
		}

		// Convert to Bubbletea message
		return authStateMsg{state: state}
	}
}

// processUpdate processes a Telegram update and returns commands to execute.
func (m *MainModel) processUpdate(update *types.Update) []tea.Cmd {
	var cmds []tea.Cmd

	// DEBUG: Log all updates to a file for debugging
	if f, err := os.OpenFile("/tmp/ithil-updates.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "[%s] Update received: type=%d\n", time.Now().Format("15:04:05"), update.Type)
		f.Close()
	}

	switch update.Type {
	case types.UpdateTypeNewMessage:
		// New message arrived
		message := update.Message
		chatID := update.ChatID

		// Update cache
		m.cache.AddMessage(chatID, message)

		// If this chat is currently open, trigger a message to update the conversation
		if m.currentChat != nil && m.currentChat.ID == chatID {
			// IMPORTANT: Skip adding outgoing messages to the conversation UI
			// because they were already added optimistically when we sent them.
			// Only add incoming messages (from other users) to avoid duplicates.
			if !message.IsOutgoing {
				m.conversation.AddMessage(message)
			}

			// Update sidebar with new message
			m.sidebar.SetCurrentChat(m.currentChat)

			// Return a command that sends a no-op message to trigger re-render
			cmds = append(cmds, func() tea.Msg {
				return renderUpdateMsg{}
			})
		}

		// Update chat list
		if chat, exists := m.cache.GetChat(chatID); exists {
			chat.LastMessage = message
			if !message.IsOutgoing {
				chat.UnreadCount++
			}
			m.chatList.UpdateChat(chat)

			// Update sidebar if this is the current chat
			if m.currentChat != nil && m.currentChat.ID == chatID {
				m.currentChat = chat
				m.sidebar.SetCurrentChat(chat)
			}
		}

		// Update unread count in status bar
		m.updateUnreadCount()

	case types.UpdateTypeMessageEdited:
		// Message was edited
		message := update.Message
		if message != nil {
			chatID := update.ChatID

			// Update in conversation if open
			if m.currentChat != nil && m.currentChat.ID == chatID {
				m.conversation.UpdateMessage(message.ID, message.Content.Text)
			}

			// Update in cache
			if msg, exists := m.cache.GetMessage(chatID, message.ID); exists {
				msg.Content.Text = message.Content.Text
				msg.IsEdited = true
			}
		}

	case types.UpdateTypeMessageDeleted:
		// Message was deleted
		if messageID, ok := update.Data.(int64); ok {
			chatID := update.ChatID

			// Remove from conversation if open
			if m.currentChat != nil && m.currentChat.ID == chatID {
				m.conversation.RemoveMessage(messageID)
			}

			// Remove from cache
			m.cache.DeleteMessage(chatID, messageID)
		}

	case types.UpdateTypeChatReadInbox:
		// Messages were read
		if maxID, ok := update.Data.(int64); ok {
			chatID := update.ChatID

			// Update unread count in chat list
			if chat, exists := m.cache.GetChat(chatID); exists {
				chat.LastReadInboxID = maxID
				// Recalculate unread count (simple approach - set to 0)
				chat.UnreadCount = 0
				m.chatList.UpdateChat(chat)
			}

			m.updateUnreadCount()
		}

	case types.UpdateTypeChatReadOutbox:
		// Our messages were read
		if maxID, ok := update.Data.(int64); ok {
			chatID := update.ChatID

			// Update read receipts in conversation
			if m.currentChat != nil && m.currentChat.ID == chatID {
				m.conversation.MarkMessagesRead(maxID)
			}
		}

	case types.UpdateTypeUserStatus:
		// User online/offline status changed or typing
		if data, ok := update.Data.(map[string]interface{}); ok {
			// Check if it's a typing update
			if typing, exists := data["typing"]; exists && typing == true {
				if userID, ok := data["user_id"].(int64); ok {
					chatID := update.ChatID

					// If this chat is open, show typing indicator
					if m.currentChat != nil && m.currentChat.ID == chatID {
						m.conversation.SetUserTyping(userID)
					}
				}
			} else if userID, ok := data["user_id"].(int64); ok {
				// User status changed (online/offline)
				// Update in chat list if this user has a chat
				for _, chat := range m.cache.GetChats() {
					if chat.Type == types.ChatTypePrivate && chat.ID == userID {
						// Update user status from gotd status
						if status, statusOk := data["status"]; statusOk {
							chat.UserStatus = convertUserStatus(status)
						}
						// Update chat list and sidebar
						m.chatList.UpdateChat(chat)
						if m.currentChat != nil && m.currentChat.ID == chat.ID {
							m.currentChat.UserStatus = chat.UserStatus
							m.sidebar.SetCurrentChat(chat)
						}
						break
					}
				}
			}
		} else if update.Data == "typing" {
			// Simple typing indicator
			chatID := update.ChatID
			if m.currentChat != nil && m.currentChat.ID == chatID {
				m.conversation.SetUserTyping(chatID)
			}
		}

	case types.UpdateTypeNewChat:
		// New chat created (someone messaged you)
		if data, ok := update.Data.(map[string]interface{}); ok {
			if chat, ok := data["chat"].(*types.Chat); ok {
				m.cache.SetChat(chat)
				m.chatList.AddChat(chat)
			}
		}
	}

	return cmds
}

// Messages for main model.
type chatsLoadedMsg struct {
	chats []*types.Chat
}

type messagesLoadedMsg struct {
	chatID   int64
	messages []*types.Message
}

type errorMsg struct {
	error string
}

type tickMsg time.Time

type telegramUpdateMsg struct {
	update *types.Update
}

type renderUpdateMsg struct{}

type clearStatusMsg struct{}

// convertUserStatus converts gotd user status to our UserStatus type.
func convertUserStatus(status interface{}) types.UserStatus {
	// gotd uses tg.UserStatus types, we need to convert them
	// For now, we'll use a simple check on the type name
	statusStr := fmt.Sprintf("%T", status)

	if strings.Contains(statusStr, "UserStatusOnline") {
		return types.UserStatusOnline
	} else if strings.Contains(statusStr, "UserStatusOffline") {
		return types.UserStatusOffline
	} else if strings.Contains(statusStr, "UserStatusRecently") {
		return types.UserStatusRecently
	} else if strings.Contains(statusStr, "UserStatusLastWeek") {
		return types.UserStatusLastWeek
	} else if strings.Contains(statusStr, "UserStatusLastMonth") {
		return types.UserStatusLastMonth
	}

	return types.UserStatusOffline
}
