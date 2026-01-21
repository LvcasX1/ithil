// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// InputComponent represents the message input field.
type InputComponent struct {
	textArea      textarea.Model
	Width         int
	Height        int
	ReplyToID     int64
	EditMessageID int64
	Focused       bool
	AttachedFile  string // Path to attached file for sending media
}

// NewInputComponent creates a new input component.
func NewInputComponent(width, height int) *InputComponent {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	// Do NOT focus by default - only focus when explicitly requested
	ta.CharLimit = 4096 // Telegram message limit
	ta.SetWidth(width - 2)
	ta.SetHeight(1) // Start with minimal height, will expand dynamically
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false) // Enter sends message, not newline

	return &InputComponent{
		textArea: ta,
		Width:    width,
		Height:   2, // Minimal initial height (border + 1 line)
		Focused:  false, // Start unfocused to prevent input context bleed
	}
}

// Init initializes the input component.
func (i *InputComponent) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles input component updates.
func (i *InputComponent) Update(msg tea.Msg) (*InputComponent, tea.Cmd) {
	var cmd tea.Cmd
	i.textArea, cmd = i.textArea.Update(msg)
	return i, cmd
}

// View renders the input component.
func (i *InputComponent) View() string {
	var sb strings.Builder

	// Reply/Edit/Attachment indicator
	if i.ReplyToID != 0 || i.EditMessageID != 0 || i.AttachedFile != "" {
		indicator := i.renderIndicator()
		sb.WriteString(indicator)
		sb.WriteString("\n")
	}

	// Input field (textarea)
	sb.WriteString(i.textArea.View())

	// Apply container style (border + width, no fixed height)
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

// getContainerStyle returns the appropriate container style based on focus state.
// Note: We intentionally do NOT set Height() here - let content determine height naturally.
// Setting a fixed height causes clipping/padding issues with the textarea.
func (i *InputComponent) getContainerStyle() lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		BorderTop(true).
		Width(i.Width - 2)

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
	return i.textArea.Value()
}

// SetValue sets the input value.
func (i *InputComponent) SetValue(value string) {
	i.textArea.SetValue(value)
}

// Clear clears the input value.
func (i *InputComponent) Clear() {
	i.textArea.Reset()
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
	return i.textArea.Focus()
}

// Blur blurs the input component.
func (i *InputComponent) Blur() {
	i.Focused = false
	i.textArea.Blur()
}

// SetWidth sets the width of the input component.
func (i *InputComponent) SetWidth(width int) {
	i.Width = width
	i.textArea.SetWidth(width - 2)
}

// SetHeight sets the height of the input component.
func (i *InputComponent) SetHeight(height int) {
	i.Height = height

	// Calculate overhead: border (1) + indicator (0-1)
	// Char count is inline with border, not extra line
	overhead := 1 // border

	hasIndicator := i.ReplyToID != 0 || i.EditMessageID != 0 || i.AttachedFile != ""
	if hasIndicator {
		overhead++
	}

	taHeight := height - overhead
	if taHeight < 1 {
		taHeight = 1
	}
	i.textArea.SetHeight(taHeight)
}

// IsEmpty returns true if the input is empty.
func (i *InputComponent) IsEmpty() bool {
	return strings.TrimSpace(i.textArea.Value()) == ""
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

// CalculateRequiredHeight calculates the height needed to display all content.
// Returns the number of lines needed (minimum 2, maximum maxHeight).
// This includes: border (1) + indicator (0-1) + text lines
func (i *InputComponent) CalculateRequiredHeight(maxHeight int) int {
	totalHeight := 1 // border

	hasIndicator := i.ReplyToID != 0 || i.EditMessageID != 0 || i.AttachedFile != ""
	if hasIndicator {
		totalHeight++
	}

	text := i.textArea.Value()
	if text == "" {
		totalHeight++ // placeholder line
	} else {
		// Get the textarea's actual width for accurate wrapping calculation
		textareaWidth := i.textArea.Width()
		if textareaWidth < 1 {
			textareaWidth = i.Width - 2 // fallback
		}

		// Split by hard line breaks first
		hardLines := strings.Split(text, "\n")
		for _, line := range hardLines {
			if line == "" {
				totalHeight++
				continue
			}
			// Calculate soft-wrapped lines for this hard line
			lineWidth := lipgloss.Width(line)
			wrappedLines := (lineWidth + textareaWidth - 1) / textareaWidth
			if wrappedLines < 1 {
				wrappedLines = 1
			}
			totalHeight += wrappedLines
		}
	}

	// Clamp
	if totalHeight < 2 {
		totalHeight = 2
	}
	if totalHeight > maxHeight {
		totalHeight = maxHeight
	}

	return totalHeight
}

