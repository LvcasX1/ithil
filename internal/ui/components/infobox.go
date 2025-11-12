package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// InfoBoxType represents the type of info box.
type InfoBoxType int

const (
	InfoBoxTypeInfo InfoBoxType = iota
	InfoBoxTypeWarning
	InfoBoxTypeSuccess
	InfoBoxTypeError
)

// InfoBox represents a styled information box.
type InfoBox struct {
	boxType InfoBoxType
	content string
	width   int
}

// NewInfoBox creates a new info box.
func NewInfoBox(boxType InfoBoxType, content string, width int) InfoBox {
	return InfoBox{
		boxType: boxType,
		content: content,
		width:   width,
	}
}

// Render renders the info box.
func (i InfoBox) Render() string {
	var style lipgloss.Style

	switch i.boxType {
	case InfoBoxTypeInfo:
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.AccentCyan)).
			Padding(0, 1).
			Width(i.width - 4)

	case InfoBoxTypeWarning:
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.AccentYellow)).
			Foreground(lipgloss.Color(styles.AccentYellow)).
			Padding(0, 1).
			Width(i.width - 4)

	case InfoBoxTypeSuccess:
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.AccentGreen)).
			Foreground(lipgloss.Color(styles.AccentGreen)).
			Padding(0, 1).
			Width(i.width - 4)

	case InfoBoxTypeError:
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.AccentRed)).
			Foreground(lipgloss.Color(styles.AccentRed)).
			Padding(0, 1).
			Width(i.width - 4)
	}

	// Wrap content to fit width
	lines := strings.Split(i.content, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) <= i.width-8 {
			wrappedLines = append(wrappedLines, line)
		} else {
			// Simple word wrapping
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine)+len(word)+1 <= i.width-8 {
					if currentLine != "" {
						currentLine += " "
					}
					currentLine += word
				} else {
					if currentLine != "" {
						wrappedLines = append(wrappedLines, currentLine)
					}
					currentLine = word
				}
			}
			if currentLine != "" {
				wrappedLines = append(wrappedLines, currentLine)
			}
		}
	}

	content := strings.Join(wrappedLines, "\n")
	return style.Render(content)
}

// InfoMessage renders an info message with an icon.
func InfoMessage(message string, width int) string {
	return NewInfoBox(InfoBoxTypeInfo, "ℹ "+message, width).Render()
}

// WarningMessage renders a warning message with an icon.
func WarningMessage(message string, width int) string {
	return NewInfoBox(InfoBoxTypeWarning, "⚠ "+message, width).Render()
}

// SuccessMessage renders a success message with an icon.
func SuccessMessage(message string, width int) string {
	return NewInfoBox(InfoBoxTypeSuccess, "✓ "+message, width).Render()
}

// ErrorMessage renders an error message with an icon.
func ErrorMessage(message string, width int) string {
	return NewInfoBox(InfoBoxTypeError, "✗ "+message, width).Render()
}
