// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// ConnectionStatus represents the connection status.
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
)

// StatusBarComponent represents the status bar at the bottom of the screen.
type StatusBarComponent struct {
	Width            int
	ConnectionStatus ConnectionStatus
	CurrentChat      string
	UnreadCount      int
	FilterName       string
	AppVersion       string
	Message          string
}

// NewStatusBarComponent creates a new status bar component.
func NewStatusBarComponent(width int) *StatusBarComponent {
	return &StatusBarComponent{
		Width:            width,
		ConnectionStatus: StatusDisconnected,
		AppVersion:       "0.1.0",
	}
}

// Render renders the status bar component.
func (s *StatusBarComponent) Render() string {
	// Left section: App name and connection status
	leftSection := s.renderLeftSection()

	// Middle section: Current chat or filter
	middleSection := s.renderMiddleSection()

	// Right section: Unread count and help
	rightSection := s.renderRightSection()

	// Calculate widths
	leftWidth := lipgloss.Width(leftSection)
	rightWidth := lipgloss.Width(rightSection)
	middleWidth := s.Width - leftWidth - rightWidth - 2

	if middleWidth < 0 {
		middleWidth = 0
	}

	// Center the middle section
	middleStyle := lipgloss.NewStyle().
		Width(middleWidth).
		Align(lipgloss.Center)
	middleRendered := middleStyle.Render(middleSection)

	// Combine sections
	statusBar := leftSection + middleRendered + rightSection

	// Apply container style
	return styles.StatusBarStyle.Width(s.Width).Render(statusBar)
}

// renderLeftSection renders the left section of the status bar.
func (s *StatusBarComponent) renderLeftSection() string {
	var parts []string

	// App name/logo
	appName := styles.StatusActiveStyle.Render("ITHIL")
	parts = append(parts, appName)

	// Connection status
	var statusText string
	var statusStyle lipgloss.Style

	switch s.ConnectionStatus {
	case StatusConnected:
		statusText = "Connected"
		statusStyle = styles.ConnectedStyle
	case StatusConnecting:
		statusText = "Connecting..."
		statusStyle = styles.ConnectingStyle
	case StatusDisconnected:
		statusText = "Disconnected"
		statusStyle = styles.DisconnectedStyle
	}

	parts = append(parts, statusStyle.Render(statusText))

	return strings.Join(parts, " ")
}

// renderMiddleSection renders the middle section of the status bar.
func (s *StatusBarComponent) renderMiddleSection() string {
	if s.Message != "" {
		return styles.StatusActiveStyle.Render(s.Message)
	}

	if s.CurrentChat != "" {
		return styles.StatusInfoStyle.Render(s.CurrentChat)
	}

	if s.FilterName != "" {
		return styles.StatusInfoStyle.Render(fmt.Sprintf("Filter: %s", s.FilterName))
	}

	return ""
}

// renderRightSection renders the right section of the status bar.
func (s *StatusBarComponent) renderRightSection() string {
	var parts []string

	// Unread count
	if s.UnreadCount > 0 {
		unreadText := fmt.Sprintf("%d unread", s.UnreadCount)
		parts = append(parts, styles.InfoStyle.Render(unreadText))
	}

	// Help indicator
	helpText := "? for help"
	parts = append(parts, styles.DimStyle.Render(helpText))

	return strings.Join(parts, " • ")
}

// SetConnectionStatus sets the connection status.
func (s *StatusBarComponent) SetConnectionStatus(status ConnectionStatus) {
	s.ConnectionStatus = status
}

// SetCurrentChat sets the current chat name.
func (s *StatusBarComponent) SetCurrentChat(chatName string) {
	s.CurrentChat = chatName
}

// SetUnreadCount sets the total unread count.
func (s *StatusBarComponent) SetUnreadCount(count int) {
	s.UnreadCount = count
}

// SetFilterName sets the current filter/folder name.
func (s *StatusBarComponent) SetFilterName(name string) {
	s.FilterName = name
}

// SetWidth sets the width of the status bar.
func (s *StatusBarComponent) SetWidth(width int) {
	s.Width = width
}

// SetMessage sets a temporary message to display in the status bar.
func (s *StatusBarComponent) SetMessage(message string) {
	s.Message = message
}

// ClearMessage clears the temporary message.
func (s *StatusBarComponent) ClearMessage() {
	s.Message = ""
}
