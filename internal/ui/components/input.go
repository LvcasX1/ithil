// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// InputComponent represents the message input field.
type InputComponent struct {
	textInput     textinput.Model
	Width         int
	Height        int
	ReplyToID     int64
	EditMessageID int64
	Focused       bool
	AttachedFile  string // Path to attached file for sending media
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

	// Reply/Edit/Attachment indicator (if space available)
	if (i.ReplyToID != 0 || i.EditMessageID != 0 || i.AttachedFile != "") && linesUsed < availableLines {
		indicator := i.renderIndicator()
		sb.WriteString(indicator)
		linesUsed++
		if linesUsed < availableLines {
			sb.WriteString("\n")
		}
	}

	// Input field with prefix icon (always show)
	if linesUsed < availableLines {
		inputLine := i.renderInputLine()
		sb.WriteString(inputLine)
		linesUsed++
	}

	// Help text (if space available)
	help := i.renderHelp()
	if help != "" && linesUsed < availableLines {
		sb.WriteString("\n")
		sb.WriteString(help)
		linesUsed++
	}

	// Apply container style with exact height and enhanced border
	containerStyle := i.getContainerStyle()
	return containerStyle.Render(sb.String())
}

// renderIndicator renders the reply/edit/attachment indicator with enhanced styling.
func (i *InputComponent) renderIndicator() string {
	var indicators []string

	if i.EditMessageID != 0 {
		icon := "✏️"
		text := "Editing message"
		hint := "Esc to cancel"

		styledText := styles.WarningStyle.Render(icon + " " + text)
		styledHint := styles.DimStyle.Render("(" + hint + ")")
		indicators = append(indicators, styledText+" "+styledHint)
	}

	if i.ReplyToID != 0 {
		icon := "↩️"
		text := "Replying to message"
		hint := "Esc to cancel"

		styledText := styles.InfoStyle.Render(icon + " " + text)
		styledHint := styles.DimStyle.Render("(" + hint + ")")
		indicators = append(indicators, styledText+" "+styledHint)
	}

	if i.AttachedFile != "" {
		icon := "📎"
		// Truncate long file paths
		filePath := i.AttachedFile
		if len(filePath) > 40 {
			filePath = "..." + filePath[len(filePath)-37:]
		}
		text := "Attachment: " + filePath
		hint := "Ctrl+X to remove"

		styledText := styles.SuccessStyle.Render(icon + " " + text)
		styledHint := styles.DimStyle.Render("(" + hint + ")")
		indicators = append(indicators, styledText+" "+styledHint)
	}

	if len(indicators) == 0 {
		return ""
	}

	// Use visual separators
	separator := styles.DimStyle.Render(" │ ")
	return strings.Join(indicators, separator)
}

// renderHelp renders help text with visual separators.
func (i *InputComponent) renderHelp() string {
	if !i.Focused {
		return ""
	}

	// Create help items with icons
	helpItems := []string{
		"⏎ Send",
		"Ctrl+A Attach",
		"Ctrl+X Remove",
		"Esc Cancel",
	}

	separator := styles.DimStyle.Render(" │ ")
	helpText := strings.Join(helpItems, separator)
	return styles.DimStyle.Render(helpText)
}

// renderInputLine renders the input field with a visual prefix.
func (i *InputComponent) renderInputLine() string {
	// Add input prefix based on state
	var prefix string
	var prefixStyle lipgloss.Style

	if i.Focused {
		prefix = "✎"
		prefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.AccentCyan)).Bold(true)
	} else {
		prefix = "○"
		prefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.TextSecondary))
	}

	// Get character count info
	charInfo := i.getCharacterInfo()

	// Combine prefix + input + char info
	styledPrefix := prefixStyle.Render(prefix + " ")
	inputView := i.textInput.View()

	return styledPrefix + inputView + charInfo
}

// getCharacterInfo returns character count information when focused.
func (i *InputComponent) getCharacterInfo() string {
	if !i.Focused {
		return ""
	}

	charCount := len(i.textInput.Value())
	charLimit := i.textInput.CharLimit

	// Only show when approaching limit or when there's content
	if charCount == 0 {
		return ""
	}

	// Different styling based on how close to limit
	var charStyle lipgloss.Style
	percentage := float64(charCount) / float64(charLimit) * 100

	if percentage >= 90 {
		charStyle = styles.ErrorStyle
	} else if percentage >= 75 {
		charStyle = styles.WarningStyle
	} else {
		charStyle = styles.DimStyle
	}

	info := fmt.Sprintf(" %d/%d", charCount, charLimit)
	return charStyle.Render(info)
}

// getContainerStyle returns the appropriate container style based on focus state.
func (i *InputComponent) getContainerStyle() lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Padding(1).
		BorderTop(true).
		Width(i.Width - 2).
		Height(i.Height)

	if i.Focused {
		// Enhanced border when focused
		return baseStyle.
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color(styles.BorderFocused))
	}

	// Subtle border when not focused
	return baseStyle.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(styles.BorderNormal))
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
	i.AttachedFile = ""
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

// SetAttachment sets a file attachment for sending media.
func (i *InputComponent) SetAttachment(filePath string) {
	i.AttachedFile = filePath
}

// GetAttachment returns the attached file path.
func (i *InputComponent) GetAttachment() string {
	return i.AttachedFile
}

// HasAttachment returns true if a file is attached.
func (i *InputComponent) HasAttachment() bool {
	return i.AttachedFile != ""
}

// ClearAttachment clears the file attachment.
func (i *InputComponent) ClearAttachment() {
	i.AttachedFile = ""
}
