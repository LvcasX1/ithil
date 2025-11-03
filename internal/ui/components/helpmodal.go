// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// HelpModalComponent represents a scrollable help modal.
type HelpModalComponent struct {
	viewport viewport.Model
	width    int
	height   int
	visible  bool
	content  string
	ready    bool
}

// NewHelpModalComponent creates a new help modal component.
func NewHelpModalComponent() *HelpModalComponent {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle()

	return &HelpModalComponent{
		viewport: vp,
		visible:  false,
		ready:    false,
	}
}

// Init initializes the help modal component.
func (h *HelpModalComponent) Init() tea.Cmd {
	return nil
}

// Update handles help modal updates.
func (h *HelpModalComponent) Update(msg tea.Msg) (*HelpModalComponent, tea.Cmd) {
	if !h.visible {
		return h, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "?":
			// Close modal
			h.visible = false
			return h, func() tea.Msg {
				return HelpModalDismissedMsg{}
			}
		case "up", "k":
			h.viewport.LineUp(1)
		case "down", "j":
			h.viewport.LineDown(1)
		case "pgup", "ctrl+b":
			h.viewport.ViewUp()
		case "pgdown", "ctrl+f":
			h.viewport.ViewDown()
		case "ctrl+u":
			h.viewport.HalfViewUp()
		case "ctrl+d":
			h.viewport.HalfViewDown()
		case "g", "home":
			h.viewport.GotoTop()
		case "G", "end":
			h.viewport.GotoBottom()
		}
	}

	h.viewport, cmd = h.viewport.Update(msg)
	return h, cmd
}

// View renders the help modal.
func (h *HelpModalComponent) View() string {
	if !h.visible {
		return ""
	}

	// Build help content if not already done
	if !h.ready {
		h.buildContent()
		h.ready = true
	}

	// Create the viewport view
	viewportView := h.viewport.View()

	// Create border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.AccentCyan)).
		Padding(0, 1)

	// Add title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Bold(true).
		Width(h.width - 4).
		Align(lipgloss.Center)

	title := titleStyle.Render("Ithil - Keyboard Shortcuts")

	// Add footer with navigation hints
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Width(h.width - 4).
		Align(lipgloss.Center)

	scrollPercentage := int(h.viewport.ScrollPercent() * 100)
	footer := footerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextSecondary)).Render(
				"↑/↓ j/k: Line • PgUp/PgDn: Page • Ctrl+U/D: Half Page • g/G: Top/Bottom",
			),
			lipgloss.NewStyle().Foreground(lipgloss.Color(styles.AccentYellow)).Render(
				"ESC, q, or ? to close",
			),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextSecondary)).Italic(true).Render(
				lipgloss.JoinHorizontal(lipgloss.Left, "Scroll: ", lipgloss.NewStyle().Foreground(lipgloss.Color(styles.AccentCyan)).Render(
					lipgloss.JoinHorizontal(lipgloss.Left,
						strings.Repeat("█", scrollPercentage/5),
						strings.Repeat("░", 20-scrollPercentage/5),
					),
				), " ", lipgloss.NewStyle().Render(lipgloss.JoinHorizontal(lipgloss.Left, string(rune(scrollPercentage/10+'0')), string(rune((scrollPercentage%10)+'0')), "%"))),
			),
		),
	)

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		viewportView,
		"",
		footer,
	)

	return borderStyle.Render(content)
}

// Show shows the help modal.
func (h *HelpModalComponent) Show() {
	h.visible = true
	if !h.ready {
		h.buildContent()
		h.ready = true
	}
}

// Hide hides the help modal.
func (h *HelpModalComponent) Hide() {
	h.visible = false
}

// IsVisible returns whether the help modal is visible.
func (h *HelpModalComponent) IsVisible() bool {
	return h.visible
}

// SetSize sets the help modal size.
func (h *HelpModalComponent) SetSize(width, height int) {
	h.width = width
	h.height = height

	// Update viewport size (account for border, padding, title, footer)
	// Border: 2 lines (top+bottom), Padding: 2 lines, Title: 2 lines, Footer: 7 lines
	viewportHeight := height - 13
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	viewportWidth := width - 4 // Account for border and padding
	if viewportWidth < 40 {
		viewportWidth = 40
	}

	h.viewport.Width = viewportWidth
	h.viewport.Height = viewportHeight

	// Rebuild content with new width
	if h.ready {
		h.buildContent()
	}
}

// buildContent builds the help content with all keyboard shortcuts.
func (h *HelpModalComponent) buildContent() {
	contentWidth := h.viewport.Width

	// Helper function to create section headers
	sectionHeader := func(title string) string {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentYellow)).
			Bold(true).
			Underline(true).
			Width(contentWidth)
		return style.Render(title)
	}

	// Helper function to create shortcut entries
	shortcut := func(keys, description string) string {
		keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentCyan)).
			Bold(true).
			Width(20).
			Align(lipgloss.Left)

		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextPrimary)).
			Width(contentWidth - 20)

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			keyStyle.Render(keys),
			descStyle.Render(description),
		)
	}

	// Build the complete help content
	var sections []string

	// Introduction
	introStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Italic(true).
		Width(contentWidth).
		Align(lipgloss.Center)
	sections = append(sections, introStyle.Render("Welcome to Ithil - A Telegram TUI Client"))
	sections = append(sections, "")

	// Global Navigation
	sections = append(sections, sectionHeader("Global Shortcuts"))
	sections = append(sections, shortcut("?", "Toggle this help modal"))
	sections = append(sections, shortcut("Ctrl+C, Ctrl+Q", "Quit application"))
	sections = append(sections, shortcut("Esc", "Back / Cancel current action"))
	sections = append(sections, shortcut("Ctrl+R", "Refresh current view"))
	sections = append(sections, shortcut("/, Ctrl+F", "Search"))
	sections = append(sections, "")

	// Pane Navigation
	sections = append(sections, sectionHeader("Pane Navigation"))
	sections = append(sections, shortcut("Tab", "Switch to next pane"))
	sections = append(sections, shortcut("Shift+Tab", "Switch to previous pane"))
	sections = append(sections, shortcut("Ctrl+1", "Focus chat list pane"))
	sections = append(sections, shortcut("Ctrl+2", "Focus conversation pane"))
	sections = append(sections, shortcut("Ctrl+3", "Focus sidebar pane"))
	sections = append(sections, shortcut("Ctrl+S", "Toggle sidebar visibility"))
	sections = append(sections, "")

	// List Navigation
	sections = append(sections, sectionHeader("List Navigation (Chat List & Messages)"))
	sections = append(sections, shortcut("↑, k", "Move up one item"))
	sections = append(sections, shortcut("↓, j", "Move down one item"))
	sections = append(sections, shortcut("PgUp, Ctrl+B", "Scroll up one page"))
	sections = append(sections, shortcut("PgDn, Ctrl+F", "Scroll down one page"))
	sections = append(sections, shortcut("Ctrl+U", "Scroll up half page"))
	sections = append(sections, shortcut("Ctrl+D", "Scroll down half page"))
	sections = append(sections, shortcut("g, Home", "Go to top"))
	sections = append(sections, shortcut("G, End", "Go to bottom"))
	sections = append(sections, "")

	// Chat List Actions
	sections = append(sections, sectionHeader("Chat List Actions"))
	sections = append(sections, shortcut("Enter", "Open selected chat"))
	sections = append(sections, shortcut("p", "Pin/unpin chat"))
	sections = append(sections, shortcut("m", "Mute/unmute chat"))
	sections = append(sections, shortcut("a", "Archive chat"))
	sections = append(sections, shortcut("r", "Mark chat as read"))
	sections = append(sections, shortcut("d", "Delete chat"))
	sections = append(sections, "")

	// Message Actions
	sections = append(sections, sectionHeader("Message Actions (in Conversation)"))
	sections = append(sections, shortcut("r", "Reply to selected message"))
	sections = append(sections, shortcut("e", "Edit selected message (your messages only)"))
	sections = append(sections, shortcut("d", "Delete selected message (your messages only)"))
	sections = append(sections, shortcut("f", "Forward selected message"))
	sections = append(sections, shortcut("y", "Copy message text to clipboard"))
	sections = append(sections, shortcut("x", "React to message with emoji"))
	sections = append(sections, shortcut("p", "Pin/unpin message in chat"))
	sections = append(sections, shortcut("s", "Save/download message media"))
	sections = append(sections, "")

	// Message Input
	sections = append(sections, sectionHeader("Message Input"))
	sections = append(sections, shortcut("Enter", "Send message"))
	sections = append(sections, shortcut("Ctrl+Enter", "Send message (alternative)"))
	sections = append(sections, shortcut("Shift+Enter", "Insert new line in message"))
	sections = append(sections, shortcut("Ctrl+A", "Attach file to message"))
	sections = append(sections, shortcut("Ctrl+E", "Insert emoji picker"))
	sections = append(sections, shortcut("Esc", "Cancel reply or edit mode"))
	sections = append(sections, "")

	// Media Actions
	sections = append(sections, sectionHeader("Media & Links"))
	sections = append(sections, shortcut("v", "View media (images, videos, etc.)"))
	sections = append(sections, shortcut("o", "Open link in default browser"))
	sections = append(sections, "")

	// Sidebar
	sections = append(sections, sectionHeader("Sidebar"))
	sections = append(sections, shortcut("i", "View user/chat info"))
	sections = append(sections, shortcut("Ctrl+S", "Toggle sidebar visibility"))
	sections = append(sections, "")

	// Settings
	sections = append(sections, sectionHeader("Settings & Preferences"))
	sections = append(sections, shortcut("Ctrl+,", "Open settings"))
	sections = append(sections, shortcut("Ctrl+T", "Toggle theme (if available)"))
	sections = append(sections, shortcut("S", "Toggle stealth mode (no read receipts)"))
	sections = append(sections, "")

	// Filters/Folders
	sections = append(sections, sectionHeader("Filters & Folders"))
	sections = append(sections, shortcut("Ctrl+L", "Show chat filters/folders"))
	sections = append(sections, shortcut("Ctrl+]", "Switch to next filter"))
	sections = append(sections, shortcut("Ctrl+[", "Switch to previous filter"))
	sections = append(sections, "")

	// Tips section
	sections = append(sections, sectionHeader("Tips & Tricks"))
	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Width(contentWidth).
		PaddingLeft(2)

	tips := []string{
		"• Use stealth mode (S) to prevent sending read receipts and 'typing' indicators",
		"• Pin important chats to keep them at the top of your list",
		"• Use quick jump numbers (1-9) in chat list to open chats quickly",
		"• Press '/' to quickly search for chats or messages",
		"• Use Ctrl+1/2/3 to quickly switch between panes",
		"• Vim-style navigation (j/k, g/G) works throughout the app",
	}

	for _, tip := range tips {
		sections = append(sections, tipStyle.Render(tip))
	}
	sections = append(sections, "")

	// Footer note
	noteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Italic(true).
		Width(contentWidth).
		Align(lipgloss.Center)
	sections = append(sections, noteStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	sections = append(sections, noteStyle.Render("Ithil - Fast, keyboard-driven Telegram experience"))
	sections = append(sections, noteStyle.Render("For more information, visit the documentation"))

	// Join all sections
	h.content = strings.Join(sections, "\n")
	h.viewport.SetContent(h.content)
}

// HelpModalDismissedMsg is sent when the help modal is dismissed.
type HelpModalDismissedMsg struct{}
