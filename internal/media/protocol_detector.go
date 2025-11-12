// Package media provides media rendering utilities for terminal display.
package media

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// GraphicsProtocol represents the type of graphics protocol supported by the terminal.
type GraphicsProtocol int

const (
	// ProtocolKitty - Kitty graphics protocol (highest quality)
	ProtocolKitty GraphicsProtocol = iota
	// ProtocolSixel - Sixel graphics protocol (widely supported)
	ProtocolSixel
	// ProtocolUnicodeMosaic - Unicode half-block rendering with true color
	ProtocolUnicodeMosaic
	// ProtocolASCII - Fallback ASCII art rendering
	ProtocolASCII
)

// String returns the string representation of the protocol.
func (p GraphicsProtocol) String() string {
	switch p {
	case ProtocolKitty:
		return "Kitty"
	case ProtocolSixel:
		return "Sixel"
	case ProtocolUnicodeMosaic:
		return "Unicode Mosaic"
	case ProtocolASCII:
		return "ASCII"
	default:
		return "Unknown"
	}
}

// ProtocolDetector handles detection of terminal graphics capabilities.
type ProtocolDetector struct {
	detectedProtocol *GraphicsProtocol
	termProgram      string
	termVar          string
	colorTerm        string
}

// NewProtocolDetector creates a new protocol detector.
func NewProtocolDetector() *ProtocolDetector {
	return &ProtocolDetector{
		termProgram: os.Getenv("TERM_PROGRAM"),
		termVar:     os.Getenv("TERM"),
		colorTerm:   os.Getenv("COLORTERM"),
	}
}

// DetectProtocol detects the best available graphics protocol for the current terminal.
// Returns the highest quality protocol supported by the terminal.
func (d *ProtocolDetector) DetectProtocol() GraphicsProtocol {
	// Return cached result if already detected
	if d.detectedProtocol != nil {
		return *d.detectedProtocol
	}

	protocol := d.detectProtocolInternal()
	d.detectedProtocol = &protocol
	return protocol
}

// detectProtocolInternal performs the actual protocol detection.
func (d *ProtocolDetector) detectProtocolInternal() GraphicsProtocol {
	// 1. Check for Kitty terminal
	if d.isKittyTerminal() {
		return ProtocolKitty
	}

	// 2. Check for Sixel support
	if d.supportsSixel() {
		return ProtocolSixel
	}

	// 3. Check for true color support (24-bit color)
	if d.supportsTrueColor() {
		return ProtocolUnicodeMosaic
	}

	// 4. Fallback to ASCII
	return ProtocolASCII
}

// isKittyTerminal checks if the terminal is Kitty or supports Kitty graphics protocol.
func (d *ProtocolDetector) isKittyTerminal() bool {
	// Check TERM_PROGRAM environment variable
	if strings.Contains(strings.ToLower(d.termProgram), "kitty") {
		return true
	}

	// Check TERM environment variable
	if strings.Contains(d.termVar, "kitty") {
		return true
	}

	// Check for KITTY_WINDOW_ID (set by Kitty terminal)
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}

	return false
}

// supportsSixel checks if the terminal supports Sixel graphics protocol.
func (d *ProtocolDetector) supportsSixel() bool {
	// Known terminals with Sixel support
	sixelTerminals := []string{
		"xterm",
		"mlterm",
		"wezterm",
		"foot",
		"contour",
		"yaft",
	}

	termLower := strings.ToLower(d.termVar)
	programLower := strings.ToLower(d.termProgram)

	for _, terminal := range sixelTerminals {
		if strings.Contains(termLower, terminal) || strings.Contains(programLower, terminal) {
			return true
		}
	}

	// Check for Sixel support via terminal query
	// Note: This is a best-effort check and may not work in all environments
	// In a TUI application, querying the terminal for capabilities can be tricky
	// We rely on environment variables as the primary detection method

	return false
}

// supportsTrueColor checks if the terminal supports 24-bit true color.
func (d *ProtocolDetector) supportsTrueColor() bool {
	// Check COLORTERM environment variable
	colorTermLower := strings.ToLower(d.colorTerm)
	if colorTermLower == "truecolor" || colorTermLower == "24bit" {
		return true
	}

	// Check TERM for true color indicators
	termLower := strings.ToLower(d.termVar)
	if strings.Contains(termLower, "24bit") || strings.Contains(termLower, "truecolor") {
		return true
	}

	// Modern terminals that typically support true color
	trueColorTerminals := []string{
		"iterm",
		"vte",
		"konsole",
		"gnome",
		"terminator",
		"alacritty",
		"wezterm",
		"kitty",
	}

	programLower := strings.ToLower(d.termProgram)
	for _, terminal := range trueColorTerminals {
		if strings.Contains(programLower, terminal) || strings.Contains(termLower, terminal) {
			return true
		}
	}

	return false
}

// GetProtocolInfo returns information about the detected protocol and terminal environment.
func (d *ProtocolDetector) GetProtocolInfo() string {
	protocol := d.DetectProtocol()

	info := fmt.Sprintf("Graphics Protocol: %s\n", protocol)
	info += fmt.Sprintf("TERM: %s\n", d.termVar)
	info += fmt.Sprintf("TERM_PROGRAM: %s\n", d.termProgram)
	info += fmt.Sprintf("COLORTERM: %s\n", d.colorTerm)

	return info
}

// ForceProtocol forces the use of a specific protocol (for testing or user preference).
func (d *ProtocolDetector) ForceProtocol(protocol GraphicsProtocol) {
	d.detectedProtocol = &protocol
}

// QueryTerminalCapabilities attempts to query the terminal for its capabilities.
// This is experimental and may not work reliably in all environments.
// Returns a channel that will receive the query response or timeout.
func (d *ProtocolDetector) QueryTerminalCapabilities() <-chan string {
	responseChan := make(chan string, 1)

	go func() {
		// Query terminal using Device Attributes (DA1)
		// Send: \x1b[c
		// Response format: \x1b[?<attrs>c
		// If attrs contains '4', Sixel is supported

		fmt.Print("\x1b[c")

		// Wait for response with timeout
		timeout := time.After(100 * time.Millisecond)
		select {
		case <-timeout:
			responseChan <- ""
		}
	}()

	return responseChan
}

// GetRenderer returns the appropriate renderer for the detected protocol.
func (d *ProtocolDetector) GetRenderer(maxWidth, maxHeight int, colored bool) ImageRendererInterface {
	protocol := d.DetectProtocol()

	switch protocol {
	case ProtocolKitty:
		return NewKittyRenderer(maxWidth, maxHeight)
	case ProtocolSixel:
		return NewSixelRenderer(maxWidth, maxHeight)
	case ProtocolUnicodeMosaic:
		return NewMosaicRenderer(maxWidth, maxHeight, colored)
	case ProtocolASCII:
		return NewImageRenderer(maxWidth, maxHeight, colored)
	default:
		// Fallback to ASCII
		return NewImageRenderer(maxWidth, maxHeight, colored)
	}
}

// ImageRendererInterface defines the interface for image renderers.
// Note: RenderImage accepts image.Image for type safety
type ImageRendererInterface interface {
	RenderImageFile(filePath string) (string, error)
	SetColored(enabled bool)
	GetDimensions() (width, height int)
	SetDimensions(width, height int)
}
