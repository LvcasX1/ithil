// Package models provides Bubbletea models for the Ithil TUI.
package models

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/cache"
	"github.com/lvcasx1/ithil/internal/ui/components"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/pkg/types"
)

// ChatListModel represents the chat list pane.
type ChatListModel struct {
	cache         *cache.Cache
	chats         []*types.Chat
	selectedIndex int
	width         int
	height        int
	focused       bool
	viewport      viewport.Model
	searchMode    bool   // Whether search/filter mode is active
	searchQuery   string // Current search query
	filteredChats []*types.Chat
}

// NewChatListModel creates a new chat list model.
func NewChatListModel(cache *cache.Cache) *ChatListModel {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = false

	return &ChatListModel{
		cache:         cache,
		chats:         []*types.Chat{},
		selectedIndex: 0,
		focused:       true,
		viewport:      vp,
	}
}

// Init initializes the chat list model.
func (m *ChatListModel) Init() tea.Cmd {
	return nil
}

// Update handles chat list model updates.
func (m *ChatListModel) Update(msg tea.Msg) (*ChatListModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		// Handle search mode separately
		if m.searchMode {
			switch msg.String() {
			case "esc":
				// Exit search mode
				m.exitSearchMode()
				return m, nil
			case "enter":
				// Open selected chat and exit search
				m.exitSearchMode()
				return m, m.openSelectedChat()
			case "backspace":
				// Remove last character from search
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterChats()
				}
				return m, nil
			default:
				// Add character to search query
				if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
					m.searchQuery += msg.String()
					m.filterChats()
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "up", "k":
			m.moveUp()
			return m, nil
		case "down", "j":
			m.moveDown()
			return m, nil
		case "ctrl+u":
			// Move up 5 chats for faster navigation
			for i := 0; i < 5 && m.selectedIndex > 0; i++ {
				m.moveUp()
			}
			return m, nil
		case "ctrl+d":
			// Move down 5 chats for faster navigation
			for i := 0; i < 5 && m.selectedIndex < len(m.chats)-1; i++ {
				m.moveDown()
			}
			return m, nil
		case "home", "g":
			m.selectedIndex = 0
			m.updateViewport()
			return m, nil
		case "end", "G":
			if len(m.chats) > 0 {
				m.selectedIndex = len(m.chats) - 1
				m.updateViewport()
			}
			return m, nil
		case "pgup":
			m.viewport.ViewUp()
			return m, nil
		case "pgdown":
			m.viewport.ViewDown()
			return m, nil
		case "enter", "l", "right":
			return m, m.openSelectedChat()
		case "/":
			// Enter search mode
			m.enterSearchMode()
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Quick jump to chat by number (1-9)
			idx := int(msg.String()[0] - '1')
			if idx < len(m.chats) {
				m.selectedIndex = idx
				m.updateViewport()
				return m, m.openSelectedChat()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case chatsUpdatedMsg:
		m.chats = msg.chats
		if m.selectedIndex >= len(m.chats) {
			m.selectedIndex = len(m.chats) - 1
		}
		if m.selectedIndex < 0 && len(m.chats) > 0 {
			m.selectedIndex = 0
		}
		m.updateViewport()
		return m, nil
	}

	// CRITICAL: Only update viewport when this pane is focused
	// This prevents navigation keys from affecting the viewport when the pane is not focused
	if m.focused {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// View renders the chat list model.
func (m *ChatListModel) View() string {
	if len(m.chats) == 0 {
		return m.renderEmpty()
	}

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

	// Render viewport content
	viewportContent := m.viewport.View()

	// Create custom top border with embedded title
	title := " 📋 CHATS "
	if m.searchMode {
		title = " 🔍 SEARCH: " + m.searchQuery + "_ "
	}
	// Use lipgloss.Width() for accurate display width
	titleLen := lipgloss.Width(title)

	// Calculate remaining border width after title
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

	// Build top border with title: ┌─ Chats ─────...─┐
	topBorderLeft := lipgloss.NewStyle().Foreground(borderColor).Render("┌─")
	topBorderTitle := lipgloss.NewStyle().Foreground(titleColor).Render(title)
	topBorderRight := lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", remainingWidth) + "┐")
	topBorder := topBorderLeft + topBorderTitle + topBorderRight

	// Create side borders for content
	leftBorder := lipgloss.NewStyle().Foreground(borderColor).Render("│")
	rightBorder := lipgloss.NewStyle().Foreground(borderColor).Render("│")

	// Wrap viewport content with side borders - minimal padding for max space
	contentLines := strings.Split(viewportContent, "\n")
	var borderedLines []string
	for _, line := range contentLines {
		// Ensure line is padded to full width
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
		fmt.Fprintf(f, "ChatList View() details:\n")
		fmt.Fprintf(f, "  width=%d, height=%d, focused=%v\n", m.width, m.height, m.focused)
		fmt.Fprintf(f, "  title=%q, titleLen=%d, remainingWidth=%d\n", title, titleLen, remainingWidth)
		fmt.Fprintf(f, "  topBorderLeft=%q\n", topBorderLeft)
		fmt.Fprintf(f, "  topBorderTitle=%q\n", topBorderTitle)
		fmt.Fprintf(f, "  topBorderRight=%q\n", topBorderRight)
		fmt.Fprintf(f, "  topBorder (full)=%q\n", topBorder)
		fmt.Fprintf(f, "  lipgloss.Width(topBorder)=%d\n", lipgloss.Width(topBorder))
		fmt.Fprintf(f, "  Raw bytes len=%d\n\n", len(firstLine))
		f.Close()
	}

	return result
}

// renderEmpty renders an empty state.
func (m *ChatListModel) renderEmpty() string {
	emptyText := "No chats yet"

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

	// Create centered empty text content
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
	title := " 📋 CHATS "
	// Use lipgloss.Width() for accurate display width
	titleLen := lipgloss.Width(title)

	// Calculate remaining border width after title
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
	return result
}

// renderAllChats renders all chat items for the viewport.
func (m *ChatListModel) renderAllChats() string {
	// Use active chats (filtered or all)
	activeChats := m.getActiveChats()

	if len(activeChats) == 0 {
		if m.searchMode {
			return styles.DimStyle.Render("No chats match your search")
		}
		return styles.DimStyle.Render("No chats yet")
	}

	var chatViews []string
	for i, chat := range activeChats {
		// Show selection highlighting even when chat list is not focused
		// This maintains visual indication of which chat is currently open
		isSelected := i == m.selectedIndex
		isFocused := m.focused
		// Use full viewport width for chat items
		chatItem := components.NewChatItemComponent(chat, isSelected, isFocused, m.viewport.Width)
		chatViews = append(chatViews, chatItem.Render())
	}

	// Compact layout: items are now tighter with consistent margins
	return strings.Join(chatViews, "\n")
}

// updateViewport updates the viewport content and scrolls to selected item.
func (m *ChatListModel) updateViewport() {
	content := m.renderAllChats()
	m.viewport.SetContent(content)

	// Calculate line position to scroll to selected item
	// Each chat item is 4 lines total:
	// - Top border line (1 line)
	// - Title line (1 line)
	// - Preview + metadata combined line (1 line)
	// - Bottom border line (1 line)
	// Note: ALL items have borders (selected = visible, unselected = invisible)
	// This ensures consistent dimensions and prevents visual shifting
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.chats) {
		linePos := m.selectedIndex * 4
		targetOffset := linePos

		// Optimized scrolling: Keep selected item visible with smooth transitions
		// Add padding to keep selected item away from edges for better visibility
		visiblePadding := 1 // Keep 1 item visible above/below when possible

		if targetOffset < m.viewport.YOffset {
			// Scrolling up
			m.viewport.YOffset = targetOffset - (visiblePadding * 4)
			if m.viewport.YOffset < 0 {
				m.viewport.YOffset = 0
			}
		} else if targetOffset >= m.viewport.YOffset+m.viewport.Height {
			// Scrolling down
			m.viewport.YOffset = targetOffset - m.viewport.Height + 4 + (visiblePadding * 4)
			if m.viewport.YOffset < 0 {
				m.viewport.YOffset = 0
			}
		}
	}
}

// moveUp moves the selection up.
func (m *ChatListModel) moveUp() {
	if m.selectedIndex > 0 {
		m.selectedIndex--
		m.updateViewport()
	}
}

// moveDown moves the selection down.
func (m *ChatListModel) moveDown() {
	activeChats := m.getActiveChats()
	if m.selectedIndex < len(activeChats)-1 {
		m.selectedIndex++
		m.updateViewport()
	}
}

// openSelectedChat opens the currently selected chat.
func (m *ChatListModel) openSelectedChat() tea.Cmd {
	activeChats := m.getActiveChats()
	if m.selectedIndex < 0 || m.selectedIndex >= len(activeChats) {
		return nil
	}

	selectedChat := activeChats[m.selectedIndex]
	return func() tea.Msg {
		return chatSelectedMsg{chat: selectedChat}
	}
}

// GetSelectedChat returns the currently selected chat.
func (m *ChatListModel) GetSelectedChat() *types.Chat {
	activeChats := m.getActiveChats()
	if m.selectedIndex < 0 || m.selectedIndex >= len(activeChats) {
		return nil
	}
	return activeChats[m.selectedIndex]
}

// SetChats sets the list of chats.
func (m *ChatListModel) SetChats(chats []*types.Chat) {
	m.chats = chats
	if m.selectedIndex >= len(m.chats) {
		m.selectedIndex = len(m.chats) - 1
	}
	if m.selectedIndex < 0 && len(m.chats) > 0 {
		m.selectedIndex = 0
	}
	// Update viewport to render the chats
	m.updateViewport()
}

// SetSize sets the size of the chat list.
func (m *ChatListModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Reserve minimal space for borders - maximize horizontal usage
	m.viewport.Width = width - 4   // Account for left border (1) + left padding (1) + right padding (1) + right border (1)
	m.viewport.Height = height - 2 // Account for top border (1) + bottom border (1)

	m.updateViewport()
}

// SetFocused sets the focused state.
func (m *ChatListModel) SetFocused(focused bool) {
	m.focused = focused
}

// UpdateChat updates a chat in the list.
func (m *ChatListModel) UpdateChat(chat *types.Chat) {
	for i, c := range m.chats {
		if c.ID == chat.ID {
			m.chats[i] = chat
			// Re-sort chats by last message date
			m.sortChats()
			m.updateViewport()
			return
		}
	}
}

// AddChat adds a new chat to the list.
func (m *ChatListModel) AddChat(chat *types.Chat) {
	m.chats = append(m.chats, chat)
	m.sortChats()
	m.updateViewport()
}

// sortChats sorts chats by last message date (most recent first).
func (m *ChatListModel) sortChats() {
	sort.Slice(m.chats, func(i, j int) bool {
		// Pinned chats always on top
		if m.chats[i].IsPinned != m.chats[j].IsPinned {
			return m.chats[i].IsPinned
		}

		// Then chats with new messages (HasNewMessage flag)
		if m.chats[i].HasNewMessage != m.chats[j].HasNewMessage {
			return m.chats[i].HasNewMessage
		}

		// Then by last message date
		if m.chats[i].LastMessage == nil {
			return false
		}
		if m.chats[j].LastMessage == nil {
			return true
		}
		return m.chats[i].LastMessage.Date.After(m.chats[j].LastMessage.Date)
	})
}

// MoveToTop moves a chat to the top of the list (after pinned chats).
func (m *ChatListModel) MoveToTop(chatID int64) {
	// Find the chat and mark it with HasNewMessage
	for i, chat := range m.chats {
		if chat.ID == chatID {
			m.chats[i].HasNewMessage = true
			break
		}
	}
	// Re-sort to move it to the top
	m.sortChats()
	m.updateViewport()
}

// ClearNewMessageFlag clears the new message flag for a chat.
func (m *ChatListModel) ClearNewMessageFlag(chatID int64) {
	for i, chat := range m.chats {
		if chat.ID == chatID {
			m.chats[i].HasNewMessage = false
			m.cache.SetChat(m.chats[i])
			break
		}
	}
	m.sortChats()
	m.updateViewport()
}

// enterSearchMode activates search/filter mode for quick chat finding.
func (m *ChatListModel) enterSearchMode() {
	m.searchMode = true
	m.searchQuery = ""
	m.filteredChats = m.chats
}

// exitSearchMode deactivates search mode and restores full chat list.
func (m *ChatListModel) exitSearchMode() {
	m.searchMode = false
	m.searchQuery = ""
	m.filteredChats = nil
	m.updateViewport()
}

// filterChats filters the chat list based on the search query.
func (m *ChatListModel) filterChats() {
	if m.searchQuery == "" {
		m.filteredChats = m.chats
		m.selectedIndex = 0
		m.updateViewport()
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.filteredChats = []*types.Chat{}

	for _, chat := range m.chats {
		// Search in chat title
		if strings.Contains(strings.ToLower(chat.Title), query) {
			m.filteredChats = append(m.filteredChats, chat)
			continue
		}
		// Search in chat username
		if chat.Username != "" && strings.Contains(strings.ToLower(chat.Username), query) {
			m.filteredChats = append(m.filteredChats, chat)
			continue
		}
		// Search in last message text
		if chat.LastMessage != nil && strings.Contains(strings.ToLower(chat.LastMessage.Content.Text), query) {
			m.filteredChats = append(m.filteredChats, chat)
		}
	}

	m.selectedIndex = 0
	m.updateViewport()
}

// getActiveChats returns the currently displayed chats (filtered or all).
func (m *ChatListModel) getActiveChats() []*types.Chat {
	if m.searchMode && m.filteredChats != nil {
		return m.filteredChats
	}
	return m.chats
}

// Messages for chat list updates.
type chatsUpdatedMsg struct {
	chats []*types.Chat
}

type chatSelectedMsg struct {
	chat *types.Chat
}
