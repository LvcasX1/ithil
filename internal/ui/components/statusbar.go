// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/media"
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

	// Audio playback state
	AudioPlaying  bool
	AudioFilename string
	AudioPosition time.Duration
	AudioDuration time.Duration
	AudioState    media.PlaybackState
	AudioSpeed    float64 // Playback speed (0.5x, 0.75x, 1x, 1.25x, 1.5x, 2x)
}

// NewStatusBarComponent creates a new status bar component.
func NewStatusBarComponent(width int) *StatusBarComponent {
	return &StatusBarComponent{
		Width:            width,
		ConnectionStatus: StatusDisconnected,
		AppVersion:       "0.1.0",
	}
}

// Render renders the status bar component with enhanced styling.
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
	middleWidth := s.Width - leftWidth - rightWidth - 4 // Add padding

	if middleWidth < 0 {
		middleWidth = 0
	}

	// Center the middle section
	middleStyle := lipgloss.NewStyle().
		Width(middleWidth).
		Align(lipgloss.Center)
	middleRendered := middleStyle.Render(middleSection)

	// Combine sections with visual separators
	separator := s.getSeparator()
	statusBar := leftSection + separator + middleRendered + separator + rightSection

	// Apply enhanced container style
	containerStyle := s.getContainerStyle()
	return containerStyle.Render(statusBar)
}

// renderLeftSection renders the left section of the status bar with enhanced visuals.
func (s *StatusBarComponent) renderLeftSection() string {
	var parts []string

	// App name/logo with icon
	appIcon := "⚡"
	appName := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Bold(true).
		Render(appIcon + " ITHIL")
	parts = append(parts, appName)

	// Connection status with visual indicator
	var statusIcon string
	var statusText string
	var statusStyle lipgloss.Style

	switch s.ConnectionStatus {
	case StatusConnected:
		statusIcon = "●"
		statusText = "Connected"
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentGreen)).
			Bold(true)
	case StatusConnecting:
		statusIcon = "◐"
		statusText = "Connecting"
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentYellow))
	case StatusDisconnected:
		statusIcon = "○"
		statusText = "Disconnected"
		statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentRed))
	}

	statusDisplay := statusStyle.Render(statusIcon + " " + statusText)
	parts = append(parts, statusDisplay)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render("│")

	return strings.Join(parts, " "+separator+" ")
}

// renderMiddleSection renders the middle section of the status bar with enhanced styling.
func (s *StatusBarComponent) renderMiddleSection() string {
	// Audio playback takes priority
	if s.AudioPlaying {
		return s.renderAudioPlayback()
	}

	if s.Message != "" {
		icon := "ℹ️"
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentYellow)).
			Bold(true).
			Render(icon + " " + s.Message)
	}

	if s.CurrentChat != "" {
		icon := "💬"
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentCyan)).
			Render(icon + " " + s.CurrentChat)
	}

	if s.FilterName != "" {
		icon := "📁"
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentMagenta)).
			Render(icon + " Filter: " + s.FilterName)
	}

	// Default message when nothing is selected
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Italic(true).
		Render("Ready")
}

// renderAudioPlayback renders the audio playback status.
func (s *StatusBarComponent) renderAudioPlayback() string {
	// State icon
	var stateIcon string
	switch s.AudioState {
	case media.StatePlaying:
		stateIcon = "▶"
	case media.StatePaused:
		stateIcon = "⏸"
	default:
		stateIcon = "⏹"
	}

	// Format time
	positionStr := formatTime(s.AudioPosition)
	durationStr := formatTime(s.AudioDuration)

	// Truncate filename if too long
	filename := s.AudioFilename
	if len(filename) > 30 {
		filename = filename[:27] + "..."
	}

	// Build status string with speed indicator
	speedStr := ""
	if s.AudioSpeed != 1.0 {
		speedStr = fmt.Sprintf(" %.2fx", s.AudioSpeed)
	}
	status := fmt.Sprintf("%s %s [%s/%s]%s", stateIcon, filename, positionStr, durationStr, speedStr)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Bold(true).
		Render(status)
}

// formatTime formats a duration as MM:SS or HH:MM:SS
func formatTime(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// renderRightSection renders the right section of the status bar with enhanced visuals.
func (s *StatusBarComponent) renderRightSection() string {
	var parts []string

	// Unread count with badge styling
	if s.UnreadCount > 0 {
		icon := "📬"
		unreadText := fmt.Sprintf("%d", s.UnreadCount)
		if s.UnreadCount > 99 {
			unreadText = "99+"
		}

		unreadBadge := lipgloss.NewStyle().
			Background(lipgloss.Color(styles.AccentRed)).
			Foreground(lipgloss.Color(styles.ColorBlack)).
			Padding(0, 1).
			Bold(true).
			Render(unreadText)

		parts = append(parts, icon+" "+unreadBadge)
	}

	// Version info
	versionText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render("v" + s.AppVersion)
	parts = append(parts, versionText)

	// Help indicator with icon
	helpIcon := "❓"
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Render(helpIcon + " Press ? for help")
	parts = append(parts, helpText)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render("│")

	return strings.Join(parts, " "+separator+" ")
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

// SetAudioPlayback sets the audio playback status.
func (s *StatusBarComponent) SetAudioPlayback(filename string, position, duration time.Duration, state media.PlaybackState, speed float64) {
	s.AudioPlaying = true
	s.AudioFilename = filename
	s.AudioPosition = position
	s.AudioDuration = duration
	s.AudioState = state
	s.AudioSpeed = speed
}

// ClearAudioPlayback clears the audio playback status.
func (s *StatusBarComponent) ClearAudioPlayback() {
	s.AudioPlaying = false
	s.AudioFilename = ""
	s.AudioPosition = 0
	s.AudioDuration = 0
	s.AudioState = media.StateStopped
	s.AudioSpeed = 1.0
}

// getSeparator returns a styled separator for the status bar.
func (s *StatusBarComponent) getSeparator() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render(" ")
}

// getContainerStyle returns an enhanced container style for the status bar.
func (s *StatusBarComponent) getContainerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(styles.AccentCyan)).
		Width(s.Width)
}
