// Package models provides Bubbletea models for the Ithil TUI.
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/cache"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/pkg/types"
)

// SidebarModel represents the info/settings sidebar pane.
type SidebarModel struct {
	cache       *cache.Cache
	currentChat *types.Chat
	currentUser *types.User
	width       int
	height      int
	visible     bool
	focused     bool
	viewport    viewport.Model
}

// NewSidebarModel creates a new sidebar model.
func NewSidebarModel(cache *cache.Cache) *SidebarModel {
	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = false

	return &SidebarModel{
		cache:    cache,
		visible:  true,
		viewport: vp,
	}
}

// Init initializes the sidebar model.
func (m *SidebarModel) Init() tea.Cmd {
	return nil
}

// Update handles sidebar model updates.
func (m *SidebarModel) Update(msg tea.Msg) (*SidebarModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

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
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case chatSelectedMsg:
		m.currentChat = msg.chat
		m.loadChatInfo()
		return m, nil
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the sidebar model.
func (m *SidebarModel) View() string {
	if !m.visible {
		return ""
	}

	if m.currentChat == nil {
		return m.renderEmpty()
	}

	// Determine border colors based on focus
	var borderColor lipgloss.Color
	var titleColor lipgloss.Color

	if m.focused {
		borderColor = lipgloss.Color(styles.BorderFocused)
		titleColor = lipgloss.Color(styles.TextBright)
	} else {
		borderColor = lipgloss.Color(styles.BorderNormal)
		titleColor = lipgloss.Color(styles.TextSecondary)
	}

	// Get viewport content
	viewportContent := m.viewport.View()

	// Create custom top border with embedded title
	title := " Info "
	titleLen := len(title)

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

	// Combine all parts
	return topBorder + "\n" + strings.Join(borderedLines, "\n") + "\n" + bottomBorder
}

// renderEmpty renders an empty state.
func (m *SidebarModel) renderEmpty() string {
	emptyText := "No chat selected"

	// Border style
	var borderColor lipgloss.Color
	var titleColor lipgloss.Color

	if m.focused {
		borderColor = lipgloss.Color(styles.BorderFocused)
		titleColor = lipgloss.Color(styles.TextBright)
	} else {
		borderColor = lipgloss.Color(styles.BorderNormal)
		titleColor = lipgloss.Color(styles.TextSecondary)
	}

	// Create centered empty text content
	contentHeight := m.height - 3 // -3 for borders
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
	title := " Info "
	titleLen := len(title)

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
	return topBorder + "\n" + strings.Join(borderedLines, "\n") + "\n" + bottomBorder
}

// renderChatInfoContent renders information about the current chat for viewport.
func (m *SidebarModel) renderChatInfoContent() string {
	var content string

	// Chat title as large heading
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Bold(true).
		Underline(true).
		Align(lipgloss.Center).
		Width(m.viewport.Width)
	content += titleStyle.Render(m.currentChat.Title) + "\n\n"

	// Chat type badge
	chatType := m.getChatTypeName(m.currentChat.Type)
	badgeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(styles.AccentBlue)).
		Foreground(lipgloss.Color(styles.TextBright)).
		Padding(0, 1).
		Bold(true)
	content += lipgloss.NewStyle().Align(lipgloss.Center).Width(m.viewport.Width).Render(badgeStyle.Render(chatType)) + "\n\n"

	// Username (if available)
	if m.currentChat.Username != "" {
		usernameStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentCyan)).
			Italic(true).
			Align(lipgloss.Center).
			Width(m.viewport.Width)
		content += usernameStyle.Render("@"+m.currentChat.Username) + "\n\n"
	}

	// Online status for private chats
	if m.currentChat.Type == types.ChatTypePrivate {
		statusText := m.getUserStatusText(m.currentChat.UserStatus)
		statusColor := m.getUserStatusColor(m.currentChat.UserStatus)
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Align(lipgloss.Center).
			Width(m.viewport.Width)
		content += statusStyle.Render(statusText) + "\n\n"
	}

	// Status badges
	if m.currentChat.IsPinned || m.currentChat.IsMuted || m.currentChat.UnreadCount > 0 {
		var badges []string
		if m.currentChat.IsPinned {
			pinBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.AccentYellow)).
				Render("📌 Pinned")
			badges = append(badges, pinBadge)
		}
		if m.currentChat.IsMuted {
			muteBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.TextSecondary)).
				Render("🔕 Muted")
			badges = append(badges, muteBadge)
		}
		if m.currentChat.UnreadCount > 0 {
			unreadBadge := lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.AccentRed)).
				Bold(true).
				Render(fmt.Sprintf("💬 %d unread", m.currentChat.UnreadCount))
			badges = append(badges, unreadBadge)
		}

		badgeContainer := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(m.viewport.Width)
		for _, badge := range badges {
			content += badgeContainer.Render(badge) + "\n"
		}
		content += "\n"
	}

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Width(m.viewport.Width).
		Align(lipgloss.Center).
		Render("━━━━━━━━━━━━━━━━")
	content += divider + "\n\n"

	// Last message section
	if m.currentChat.LastMessage != nil {
		content += styles.SidebarHeadingStyle.Render("Latest Message") + "\n\n"

		// Message sender
		senderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentYellow)).
			Bold(true)
		if m.currentChat.LastMessage.IsOutgoing {
			content += senderStyle.Render("You:") + "\n"
		} else {
			content += senderStyle.Render(m.currentChat.Title+":") + "\n"
		}

		// Message text (truncated if too long)
		messageText := m.currentChat.LastMessage.Content.Text
		maxLen := 100
		if len(messageText) > maxLen {
			messageText = messageText[:maxLen] + "..."
		}

		messageStyle := lipgloss.NewStyle().
			Width(m.viewport.Width-2).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.BorderNormal))

		content += messageStyle.Render(messageText) + "\n\n"
	}

	// Statistics
	content += divider + "\n\n"
	content += styles.SidebarHeadingStyle.Render("Statistics") + "\n\n"

	messageCount := m.cache.GetMessageCount(m.currentChat.ID)
	statsStyle := lipgloss.NewStyle()
	content += statsStyle.Render(fmt.Sprintf("📊 %d messages in cache", messageCount)) + "\n"

	return content
}

// updateViewport updates the viewport content.
func (m *SidebarModel) updateViewport() {
	if m.currentChat == nil {
		return
	}

	content := m.renderChatInfoContent()
	m.viewport.SetContent(content)
}

// renderField renders a label-value field.
func (m *SidebarModel) renderField(label, value string) string {
	labelText := styles.SidebarLabelStyle.Render(label + ":")
	valueText := styles.SidebarValueStyle.Render(value)
	return labelText + " " + valueText + "\n"
}

// getChatTypeName returns a human-readable name for the chat type.
func (m *SidebarModel) getChatTypeName(chatType types.ChatType) string {
	switch chatType {
	case types.ChatTypePrivate:
		return "Private Chat"
	case types.ChatTypeGroup:
		return "Group"
	case types.ChatTypeSupergroup:
		return "Supergroup"
	case types.ChatTypeChannel:
		return "Channel"
	case types.ChatTypeSecret:
		return "Secret Chat"
	default:
		return "Unknown"
	}
}

// loadChatInfo loads additional information about the current chat.
func (m *SidebarModel) loadChatInfo() {
	// TODO: Load additional chat information from server if needed
	// For now, we just display cached data
	m.updateViewport()
}

// SetCurrentChat sets the current chat.
func (m *SidebarModel) SetCurrentChat(chat *types.Chat) {
	m.currentChat = chat
	m.loadChatInfo()
}

// SetSize sets the size of the sidebar.
func (m *SidebarModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Reserve space for borders only (title is now embedded in border)
	m.viewport.Width = width - 4   // Account for borders and horizontal padding (left border + left padding + right padding + right border)
	m.viewport.Height = height - 2 // Account for top border (1) + bottom border (1)

	m.updateViewport()
}

// SetVisible sets the visibility of the sidebar.
func (m *SidebarModel) SetVisible(visible bool) {
	m.visible = visible
}

// ToggleVisible toggles the visibility of the sidebar.
func (m *SidebarModel) ToggleVisible() {
	m.visible = !m.visible
}

// SetFocused sets the focused state.
func (m *SidebarModel) SetFocused(focused bool) {
	m.focused = focused
}

// IsVisible returns whether the sidebar is visible.
func (m *SidebarModel) IsVisible() bool {
	return m.visible
}

// getUserStatusText returns a human-readable status text.
func (m *SidebarModel) getUserStatusText(status types.UserStatus) string {
	switch status {
	case types.UserStatusOnline:
		return "🟢 Online"
	case types.UserStatusOffline:
		return "⚫ Offline"
	case types.UserStatusRecently:
		return "🟡 Recently active"
	case types.UserStatusLastWeek:
		return "🟠 Active last week"
	case types.UserStatusLastMonth:
		return "🔴 Active last month"
	default:
		return "⚫ Offline"
	}
}

// getUserStatusColor returns the color for a user status.
func (m *SidebarModel) getUserStatusColor(status types.UserStatus) string {
	switch status {
	case types.UserStatusOnline:
		return styles.AccentGreen
	case types.UserStatusRecently:
		return styles.AccentYellow
	case types.UserStatusLastWeek:
		return styles.AccentYellow
	default:
		return styles.TextSecondary
	}
}
