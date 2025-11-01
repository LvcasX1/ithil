// Package styles provides lipgloss style definitions for the Ithil TUI.
// This package uses terminal default colors to adapt to the user's theme.
package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Terminal ANSI Color Constants
// These use standard ANSI color codes that adapt to the terminal's theme
const (
	// Standard colors (0-7)
	ColorBlack   = "0"
	ColorRed     = "1"
	ColorGreen   = "2"
	ColorYellow  = "3"
	ColorBlue    = "4"
	ColorMagenta = "5"
	ColorCyan    = "6"
	ColorWhite   = "7"

	// Bright colors (8-15)
	ColorBrightBlack   = "8" // Gray
	ColorBrightRed     = "9"
	ColorBrightGreen   = "10"
	ColorBrightYellow  = "11"
	ColorBrightBlue    = "12"
	ColorBrightMagenta = "13"
	ColorBrightCyan    = "14"
	ColorBrightWhite   = "15"
)

// Semantic color mappings for better readability
// These map to ANSI codes that will adapt to the user's terminal theme
const (
	// Primary text - uses terminal default
	TextPrimary   = ""               // Empty string uses terminal default foreground
	TextSecondary = ColorBrightBlack // Gray/dim text
	TextBright    = ColorBrightWhite // Emphasized text

	// Backgrounds - mostly omitted to use terminal default
	BgPrimary   = ""         // Terminal default background
	BgSecondary = ColorBlack // Subtle background (if needed)

	// Accents and highlights
	AccentBlue    = ColorBrightBlue
	AccentCyan    = ColorBrightCyan
	AccentGreen   = ColorBrightGreen
	AccentYellow  = ColorBrightYellow
	AccentRed     = ColorBrightRed
	AccentMagenta = ColorBrightMagenta

	// Borders
	BorderNormal  = ColorBrightBlack // Subtle border
	BorderFocused = ColorBrightBlue  // Highlighted border
)

// Global base styles
var (
	// Base text color - uses terminal default
	BaseStyle = lipgloss.NewStyle()

	// Dimmed/secondary text
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSecondary))

	// Highlighted/focused text
	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentCyan)).
			Bold(true)

	// Selected item style
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextBright)).
			Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentRed)).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentGreen)).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentYellow)).
			Bold(true)

	// Info style
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentCyan))
)

// Chat List Styles
var (
	// Chat list container
	ChatListStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Individual chat item
	ChatItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Selected chat item
	ChatItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextBright)).
				Bold(true).
				Padding(0, 1)

	// Chat title
	ChatTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextBright)).
			Bold(true)

	// Chat message preview
	ChatPreviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary))

	// Unread badge
	UnreadBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(AccentRed)).
				Foreground(lipgloss.Color(ColorBlack)).
				Bold(true).
				Padding(0, 1).
				MarginLeft(1)

	// Pinned indicator
	PinnedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentYellow))

	// Muted indicator
	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSecondary))
)

// Conversation/Message Styles
var (
	// Conversation container
	ConversationStyle = lipgloss.NewStyle().
				Padding(1)

	// Message bubble for incoming messages (left-aligned, cyan accent)
	MessageIncomingStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(AccentCyan))

	// Message bubble for outgoing messages (right-aligned, blue accent)
	MessageOutgoingStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(AccentBlue))

	// Message sender name (bright and bold)
	SenderNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentYellow)).
			Bold(true)

	// Message timestamp (subtle and small)
	TimestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSecondary))

	// Message edited indicator
	EditedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentYellow)).
			Italic(true)

	// Reply preview
	ReplyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSecondary)).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(AccentCyan)).
			PaddingLeft(1).
			Italic(true)

	// System message (user joined, etc.)
	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary)).
				Italic(true).
				Align(lipgloss.Center)

	// Reply indicator
	ReplyIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AccentCyan)).
				Padding(0, 1).
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color(AccentCyan))

	// Edit indicator
	EditIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AccentYellow)).
				Padding(0, 1).
				BorderLeft(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color(AccentYellow))
)

// Input Styles
var (
	// Input box container
	InputBoxStyle = lipgloss.NewStyle().
			Padding(1).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(BorderNormal))

	// Input text
	InputTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextBright))

	// Input placeholder
	InputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary)).
				Italic(true)

	// Input cursor
	InputCursorStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(AccentCyan))
)

// Status Bar Styles
var (
	// Status bar container
	StatusBarStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Status bar active section
	StatusActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(AccentCyan)).
				Foreground(lipgloss.Color(ColorBlack)).
				Bold(true).
				Padding(0, 1)

	// Status bar info section
	StatusInfoStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Connection status - connected
	ConnectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentGreen)).
			Bold(true)

	// Connection status - connecting
	ConnectingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentYellow))

	// Connection status - disconnected
	DisconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AccentRed))
)

// Sidebar Styles
var (
	// Sidebar container
	SidebarStyle = lipgloss.NewStyle().
			Padding(1).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(BorderNormal))

	// Sidebar heading
	SidebarHeadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AccentCyan)).
				Bold(true).
				Underline(true).
				MarginBottom(1)

	// Sidebar label
	SidebarLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary)).
				Bold(true)

	// Sidebar value
	SidebarValueStyle = lipgloss.NewStyle()
)

// Authentication Screen Styles
var (
	// Auth container
	AuthContainerStyle = lipgloss.NewStyle().
				Padding(2).
				Align(lipgloss.Center, lipgloss.Center)

	// Auth title
	AuthTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentCyan)).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(2)

	// Auth prompt
	AuthPromptStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			MarginBottom(1)

	// Auth input field
	AuthInputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextBright)).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(AccentCyan))

	// Auth error message
	AuthErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentRed)).
			Bold(true).
			Align(lipgloss.Center).
			MarginTop(1)
)

// Border styles
var (
	// Standard border (unfocused pane)
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderNormal)).
			Padding(0, 1)

	// Focused border (active pane - bright blue with thick border)
	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color(BorderFocused)).
				Padding(0, 1)
)

// Title styles for pane headers
var (
	// Standard title (unfocused pane)
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextBright)).
			Padding(0, 1).
			Bold(true)

	// Focused title (active pane - bright blue)
	FocusedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(AccentBlue)).
				Padding(0, 1).
				Bold(true)
)

// Text Entity Styles (for formatted text in messages)
var (
	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	ItalicStyle = lipgloss.NewStyle().
			Italic(true)

	CodeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(BgSecondary)).
			Foreground(lipgloss.Color(AccentYellow)).
			Padding(0, 1)

	PreStyle = lipgloss.NewStyle().
			Padding(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(BorderNormal))

	LinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentCyan)).
			Underline(true)

	StrikethroughStyle = lipgloss.NewStyle().
				Strikethrough(true)

	UnderlineStyle = lipgloss.NewStyle().
			Underline(true)

	SpoilerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(TextSecondary)).
			Foreground(lipgloss.Color(TextSecondary))
)

// Help/Shortcuts Styles
var (
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(AccentCyan)).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle()

	HelpSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TextSecondary))
)

// Logo/Title Style
var (
	LogoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(AccentCyan)).
		Bold(true).
		Align(lipgloss.Center)
)

// Typing Indicator Style
var (
	TypingIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextSecondary)).
		Italic(true).
		Padding(0, 1)
)

// Width sets the width for a style and returns it.
func Width(style lipgloss.Style, width int) lipgloss.Style {
	return style.Width(width)
}

// Height sets the height for a style and returns it.
func Height(style lipgloss.Style, height int) lipgloss.Style {
	return style.Height(height)
}

// MaxWidth sets the max width for a style and returns it.
func MaxWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.MaxWidth(width)
}

// MaxHeight sets the max height for a style and returns it.
func MaxHeight(style lipgloss.Style, height int) lipgloss.Style {
	return style.MaxHeight(height)
}
