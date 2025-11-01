// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// ModalComponent represents a floating modal overlay.
type ModalComponent struct {
	title       string
	content     string
	width       int
	height      int
	visible     bool
	centered    bool
	dismissible bool
}

// NewModalComponent creates a new modal component.
func NewModalComponent(title string, width, height int) *ModalComponent {
	return &ModalComponent{
		title:       title,
		width:       width,
		height:      height,
		visible:     false,
		centered:    true,
		dismissible: true,
	}
}

// Init initializes the modal component.
func (m *ModalComponent) Init() tea.Cmd {
	return nil
}

// Update handles modal updates.
func (m *ModalComponent) Update(msg tea.Msg) (*ModalComponent, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.dismissible {
			switch msg.String() {
			case "esc", "q":
				m.visible = false
				return m, func() tea.Msg {
					return ModalDismissedMsg{}
				}
			}
		}
	}

	return m, nil
}

// View renders the modal component.
func (m *ModalComponent) View() string {
	if !m.visible {
		return ""
	}

	// Create modal style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.AccentCyan)).
		Padding(1, 2).
		Width(m.width).
		Height(m.height)

	// Build modal content
	var contentBuilder strings.Builder

	// Title
	if m.title != "" {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(styles.TextBright)).
			Width(m.width - 6). // Account for padding and border
			Align(lipgloss.Center)
		contentBuilder.WriteString(titleStyle.Render(m.title))
		contentBuilder.WriteString("\n\n")
	}

	// Content
	contentBuilder.WriteString(m.content)

	// Footer hint
	if m.dismissible {
		contentBuilder.WriteString("\n\n")
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextSecondary)).
			Width(m.width - 6).
			Align(lipgloss.Center)
		contentBuilder.WriteString(hintStyle.Render("Press ESC or Q to close"))
	}

	return modalStyle.Render(contentBuilder.String())
}

// Show shows the modal.
func (m *ModalComponent) Show() {
	m.visible = true
}

// Hide hides the modal.
func (m *ModalComponent) Hide() {
	m.visible = false
}

// IsVisible returns whether the modal is visible.
func (m *ModalComponent) IsVisible() bool {
	return m.visible
}

// SetTitle sets the modal title.
func (m *ModalComponent) SetTitle(title string) {
	m.title = title
}

// SetContent sets the modal content.
func (m *ModalComponent) SetContent(content string) {
	m.content = content
}

// SetSize sets the modal size.
func (m *ModalComponent) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetDismissible sets whether the modal can be dismissed with ESC/Q.
func (m *ModalComponent) SetDismissible(dismissible bool) {
	m.dismissible = dismissible
}

// ModalDismissedMsg is sent when the modal is dismissed.
type ModalDismissedMsg struct{}
