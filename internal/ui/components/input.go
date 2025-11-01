// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// InputComponent represents the message input field.
type InputComponent struct {
	textInput      textinput.Model
	Width          int
	Height         int
	ReplyToID      int64
	EditMessageID  int64
	Focused        bool
}

// NewInputComponent creates a new input component.
func NewInputComponent(width, height int) *InputComponent {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	// Do NOT focus by default - only focus when explicitly requested
	ti.CharLimit = 4096 // Telegram message limit
	ti.Width = width - 4

	return &InputComponent{
		textInput: ti,
		Width:     width,
		Height:    height,
		Focused:   false, // Start unfocused to prevent input context bleed
	}
}

// Init initializes the input component.
func (i *InputComponent) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles input component updates.
func (i *InputComponent) Update(msg tea.Msg) (*InputComponent, tea.Cmd) {
	var cmd tea.Cmd
	i.textInput, cmd = i.textInput.Update(msg)
	return i, cmd
}

// View renders the input component.
func (i *InputComponent) View() string {
	var sb strings.Builder

	// Calculate available content height
	// InputBoxStyle has: BorderTop (1) + Padding top (1) + Padding bottom (1) = 3 lines overhead
	// Available content lines = Height - 3
	availableLines := i.Height - 3
	if availableLines < 1 {
		availableLines = 1 // Always show at least the input
	}

	linesUsed := 0

	// Reply/Edit indicator (if space available)
	if (i.ReplyToID != 0 || i.EditMessageID != 0) && linesUsed < availableLines {
		indicator := i.renderIndicator()
		sb.WriteString(indicator)
		linesUsed++
		if linesUsed < availableLines {
			sb.WriteString("\n")
		}
	}

	// Input field (always show)
	if linesUsed < availableLines {
		inputView := i.textInput.View()
		sb.WriteString(inputView)
		linesUsed++
	}

	// Help text (if space available)
	help := i.renderHelp()
	if help != "" && linesUsed < availableLines {
		sb.WriteString("\n")
		sb.WriteString(help)
		linesUsed++
	}

	// Apply container style with exact height
	containerStyle := styles.InputBoxStyle.Width(i.Width - 2).Height(i.Height)
	return containerStyle.Render(sb.String())
}

// renderIndicator renders the reply/edit indicator.
func (i *InputComponent) renderIndicator() string {
	if i.EditMessageID != 0 {
		text := "Editing message - Press Esc to cancel"
		return styles.WarningStyle.Render(text)
	}

	if i.ReplyToID != 0 {
		text := "Replying to message - Press Esc to cancel"
		return styles.InfoStyle.Render(text)
	}

	return ""
}

// renderHelp renders help text.
func (i *InputComponent) renderHelp() string {
	if !i.Focused {
		return ""
	}

	helpText := "Enter: Send • Shift+Enter: New line • Esc: Cancel"
	return styles.DimStyle.Render(helpText)
}

// GetValue returns the current input value.
func (i *InputComponent) GetValue() string {
	return i.textInput.Value()
}

// SetValue sets the input value.
func (i *InputComponent) SetValue(value string) {
	i.textInput.SetValue(value)
}

// Clear clears the input value.
func (i *InputComponent) Clear() {
	i.textInput.SetValue("")
	i.ReplyToID = 0
	i.EditMessageID = 0
}

// SetReplyTo sets the message ID to reply to.
func (i *InputComponent) SetReplyTo(messageID int64) {
	i.ReplyToID = messageID
	i.EditMessageID = 0
}

// SetEdit sets the message ID to edit.
func (i *InputComponent) SetEdit(messageID int64, text string) {
	i.EditMessageID = messageID
	i.ReplyToID = 0
	i.SetValue(text)
}

// CancelReplyOrEdit cancels reply or edit mode.
func (i *InputComponent) CancelReplyOrEdit() {
	i.ReplyToID = 0
	i.EditMessageID = 0
	if i.EditMessageID != 0 {
		i.Clear()
	}
}

// Focus focuses the input component.
func (i *InputComponent) Focus() tea.Cmd {
	i.Focused = true
	return i.textInput.Focus()
}

// Blur blurs the input component.
func (i *InputComponent) Blur() {
	i.Focused = false
	i.textInput.Blur()
}

// SetWidth sets the width of the input component.
func (i *InputComponent) SetWidth(width int) {
	i.Width = width
	i.textInput.Width = width - 4
}

// SetHeight sets the height of the input component.
func (i *InputComponent) SetHeight(height int) {
	i.Height = height
}

// IsEmpty returns true if the input is empty.
func (i *InputComponent) IsEmpty() bool {
	return strings.TrimSpace(i.textInput.Value()) == ""
}
