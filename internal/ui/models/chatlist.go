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

		switch msg.String() {
		case "up", "k":
			m.moveUp()
			return m, nil
		case "down", "j":
			m.moveDown()
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
		case "enter":
			return m, m.openSelectedChat()
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

	// Wrap viewport content with side borders
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
	if len(m.chats) == 0 {
		return styles.DimStyle.Render("No chats yet")
	}

	var chatViews []string
	for i, chat := range m.chats {
		isSelected := i == m.selectedIndex && m.focused
		// Use full viewport width for chat items
		chatItem := components.NewChatItemComponent(chat, isSelected, m.viewport.Width)
		chatViews = append(chatViews, chatItem.Render())
	}

	return strings.Join(chatViews, "\n") // Compact spacing between chats
}

// updateViewport updates the viewport content and scrolls to selected item.
func (m *ChatListModel) updateViewport() {
	content := m.renderAllChats()
	m.viewport.SetContent(content)

	// Calculate line position to scroll to selected item
	// Each chat item is now approximately 6 lines:
	// - Top border (1 line)
	// - Title line with padding (1 line)
	// - Preview line with padding (1 line)
	// - Metadata line with padding (1 line)
	// - Bottom border (1 line)
	// - Single newline spacing (1 line)
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.chats) {
		linePos := m.selectedIndex * 6
		targetOffset := linePos

		// Keep selected item in view
		if targetOffset < m.viewport.YOffset {
			m.viewport.YOffset = targetOffset
		} else if targetOffset >= m.viewport.YOffset+m.viewport.Height {
			m.viewport.YOffset = targetOffset - m.viewport.Height + 6
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
	if m.selectedIndex < len(m.chats)-1 {
		m.selectedIndex++
		m.updateViewport()
	}
}

// openSelectedChat opens the currently selected chat.
func (m *ChatListModel) openSelectedChat() tea.Cmd {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.chats) {
		return nil
	}

	selectedChat := m.chats[m.selectedIndex]
	return func() tea.Msg {
		return chatSelectedMsg{chat: selectedChat}
	}
}

// GetSelectedChat returns the currently selected chat.
func (m *ChatListModel) GetSelectedChat() *types.Chat {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.chats) {
		return nil
	}
	return m.chats[m.selectedIndex]
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

	// Reserve space for borders only (title is now embedded in border)
	m.viewport.Width = width - 4   // Account for borders and horizontal padding (left border + left padding + right padding + right border)
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

// Messages for chat list updates.
type chatsUpdatedMsg struct {
	chats []*types.Chat
}

type chatSelectedMsg struct {
	chat *types.Chat
}
