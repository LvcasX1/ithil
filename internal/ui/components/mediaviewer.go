// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/media"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/pkg/types"
)

// MediaViewerComponent represents a modal for viewing media files.
type MediaViewerComponent struct {
	message          *types.Message
	mediaPath        string
	width            int
	height           int
	visible          bool
	content          string
	imageRenderer    *media.ImageRenderer
	mosaicRenderer   *media.MosaicRenderer
	kittyRenderer    *media.KittyRenderer
	sixelRenderer    *media.SixelRenderer
	audioRenderer    *media.AudioRenderer
	externalAudioPlayer *media.AudioPlayer // External audio player for background playback
	protocolDetector *media.ProtocolDetector
	detectedProtocol media.GraphicsProtocol
	renderError      error
	downloading      bool
	downloadedPath   string
	refreshTicker    *time.Ticker
	stopRefresh      chan bool
}

// NewMediaViewerComponent creates a new media viewer component.
func NewMediaViewerComponent(width, height int) *MediaViewerComponent {
	detector := media.NewProtocolDetector()
	protocol := detector.DetectProtocol()

	return &MediaViewerComponent{
		width:            width,
		height:           height,
		visible:          false,
		imageRenderer:    media.NewImageRenderer(width-6, height-6, true),
		mosaicRenderer:   media.NewMosaicRenderer(width-6, height-6, true),
		kittyRenderer:    media.NewKittyRenderer(width-6, height-6),
		sixelRenderer:    media.NewSixelRenderer(width-6, height-6),
		audioRenderer:    media.NewAudioRenderer(width - 6),
		protocolDetector: detector,
		detectedProtocol: protocol,
		stopRefresh:      make(chan bool),
	}
}

// Init initializes the media viewer component.
func (m *MediaViewerComponent) Init() tea.Cmd {
	return nil
}

// Update handles media viewer updates.
func (m *MediaViewerComponent) Update(msg tea.Msg) (*MediaViewerComponent, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if this is an audio/voice message to handle playback controls
		isAudioMessage := m.message != nil &&
			(m.message.Content.Type == types.MessageTypeAudio ||
				m.message.Content.Type == types.MessageTypeVoice)

		switch msg.String() {
		case "esc", "q":
			// Stop playback if playing audio and using external player
			if isAudioMessage && m.externalAudioPlayer != nil {
				m.externalAudioPlayer.Stop()
			} else if isAudioMessage {
				m.audioRenderer.GetAudioPlayer().Stop()
			}
			m.stopUIRefresh()
			m.visible = false
			return m, func() tea.Msg {
				return MediaViewerDismissedMsg{}
			}

		// Audio playback controls
		case " ":
			if isAudioMessage {
				// Use external player if available, otherwise use local renderer's player
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.TogglePlayPause()
				} else {
					m.audioRenderer.GetAudioPlayer().TogglePlayPause()
				}
				m.renderMedia()
			}

		case "left":
			if isAudioMessage {
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.SkipBackward(5 * time.Second)
				} else {
					m.audioRenderer.GetAudioPlayer().SkipBackward(5 * time.Second)
				}
				m.renderMedia()
			}

		case "right":
			if isAudioMessage {
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.SkipForward(5 * time.Second)
				} else {
					m.audioRenderer.GetAudioPlayer().SkipForward(5 * time.Second)
				}
				m.renderMedia()
			}

		case "up":
			if isAudioMessage {
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.VolumeUp()
				} else {
					m.audioRenderer.GetAudioPlayer().VolumeUp()
				}
				m.renderMedia()
			}

		case "down":
			if isAudioMessage {
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.VolumeDown()
				} else {
					m.audioRenderer.GetAudioPlayer().VolumeDown()
				}
				m.renderMedia()
			}

		// Speed control
		case "[":
			if isAudioMessage {
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.SpeedDown()
				} else {
					m.audioRenderer.GetAudioPlayer().SpeedDown()
				}
				m.renderMedia()
			}

		case "]":
			if isAudioMessage {
				if m.externalAudioPlayer != nil {
					m.externalAudioPlayer.SpeedUp()
				} else {
					m.audioRenderer.GetAudioPlayer().SpeedUp()
				}
				m.renderMedia()
			}
		}

	case MediaDownloadedMsg:
		// Media has been downloaded, render it
		m.downloading = false
		m.downloadedPath = msg.Path
		m.renderMedia()
		return m, nil

	case RefreshUIMsg:
		// Periodic refresh for audio playback UI
		if m.message != nil &&
			(m.message.Content.Type == types.MessageTypeAudio ||
				m.message.Content.Type == types.MessageTypeVoice) {
			m.renderMedia()
			return m, m.scheduleRefresh()
		}
	}

	return m, nil
}

// View renders the media viewer component.
func (m *MediaViewerComponent) View() string {
	if !m.visible {
		return ""
	}

	// For pixel protocols displaying images, use fullscreen mode
	// This bypasses Lipgloss entirely to avoid cursor movement issues
	useFullscreen := (m.detectedProtocol == media.ProtocolKitty || m.detectedProtocol == media.ProtocolSixel) &&
		m.message != nil &&
		m.message.Content.Type == types.MessageTypePhoto &&
		!m.downloading &&
		m.renderError == nil

	if useFullscreen {
		return m.ViewFullscreen()
	}

	// For all other content, use normal Lipgloss modal
	return m.ViewModal()
}

// ViewModal renders the media viewer as a Lipgloss modal (for audio, video, docs, and low-res images).
func (m *MediaViewerComponent) ViewModal() string {
	// Create modal style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.AccentCyan)).
		Padding(1, 2).
		Width(m.width).
		MaxHeight(m.height)

	// Build modal content
	var contentBuilder strings.Builder

	// Title
	title := m.getTitle()
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(styles.TextBright)).
		Width(m.width - 6).
		Align(lipgloss.Center)
	contentBuilder.WriteString(titleStyle.Render(title))
	contentBuilder.WriteString("\n\n")

	// Content
	if m.downloading {
		// Show loading message
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextSecondary)).
			Width(m.width - 6).
			Align(lipgloss.Center)
		contentBuilder.WriteString(loadingStyle.Render("Downloading media..."))
	} else if m.renderError != nil {
		// Show error message
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.AccentRed)).
			Width(m.width - 6).
			Align(lipgloss.Center)
		contentBuilder.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.renderError.Error())))
	} else if m.content != "" {
		// Show rendered content
		contentBuilder.WriteString(m.content)
	} else {
		// Show placeholder
		contentBuilder.WriteString("No content available")
	}

	// Footer hint
	contentBuilder.WriteString("\n\n")
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Width(m.width - 6).
		Align(lipgloss.Center)
	contentBuilder.WriteString(hintStyle.Render("Press ESC or Q to close"))

	return modalStyle.Render(contentBuilder.String())
}

// ViewFullscreen renders the image in fullscreen mode.
// This bypasses Lipgloss entirely to avoid cursor movement issues with pixel protocols.
// Renders image with comfortable margins from the terminal edges.
func (m *MediaViewerComponent) ViewFullscreen() string {
	var result strings.Builder

	// Clear screen and move cursor to top-left
	result.WriteString("\x1b[2J\x1b[H")

	// Add top margin (2 blank lines)
	result.WriteString("\n\n")

	// Add left margin (5 spaces) before the image
	result.WriteString("     ")

	// Render the image content
	if m.content != "" {
		result.WriteString(m.content)
	}

	// Add text overlay at bottom
	result.WriteString("\n\n")
	overlayText := "Press ESC or Q to return to chat"
	// Center the text
	padding := (m.width - len(overlayText)) / 2
	if padding > 0 {
		result.WriteString(strings.Repeat(" ", padding))
	}
	result.WriteString(overlayText)

	return result.String()
}

// ShowMedia shows the media viewer with a specific message.
func (m *MediaViewerComponent) ShowMedia(message *types.Message, mediaPath string) tea.Cmd {
	m.message = message
	m.mediaPath = mediaPath
	m.visible = true
	m.content = ""
	m.renderError = nil
	m.downloading = false
	m.downloadedPath = mediaPath

	// DEFENSIVE CHECK: Don't trust IsDownloaded blindly
	// Verify that the file actually exists before trying to render
	needsDownload := false
	if message.Content.Media != nil {
		if !message.Content.Media.IsDownloaded {
			// Case 1: Marked as not downloaded
			needsDownload = true
		} else if message.Content.Media.LocalPath == "" {
			// Case 2: Marked as downloaded but no path
			needsDownload = true
			message.Content.Media.IsDownloaded = false // Fix stale state
		} else {
			// Case 3: Marked as downloaded with a path - verify file exists
			if _, err := os.Stat(message.Content.Media.LocalPath); os.IsNotExist(err) {
				needsDownload = true
				message.Content.Media.IsDownloaded = false // Fix stale state
			} else if err != nil {
				// Other error (permission, etc)
				needsDownload = true
				message.Content.Media.IsDownloaded = false
			}
		}
	}

	// Download if needed
	if needsDownload {
		m.downloading = true
		return func() tea.Msg {
			return MediaDownloadRequestMsg{
				Message: message,
			}
		}
	}

	// Render immediately if file exists
	m.renderMedia()

	// For audio/voice messages, load the file and start UI refresh
	if message.Content.Type == types.MessageTypeAudio ||
		message.Content.Type == types.MessageTypeVoice {
		// Load audio file into the appropriate player
		var audioPlayer *media.AudioPlayer
		if m.externalAudioPlayer != nil {
			// Use external player for background playback
			audioPlayer = m.externalAudioPlayer
		} else {
			// Fallback to local renderer's player
			audioPlayer = m.audioRenderer.GetAudioPlayer()
		}

		if err := audioPlayer.LoadFile(m.downloadedPath); err != nil {
			m.renderError = err
		}

		// Start UI refresh for playback updates
		m.startUIRefresh()
		return m.scheduleRefresh()
	}

	return nil
}

// Hide hides the media viewer.
func (m *MediaViewerComponent) Hide() {
	// DON'T stop audio playback when using external player (for background playback)
	// Only stop if using local renderer's player
	if m.message != nil &&
		(m.message.Content.Type == types.MessageTypeAudio ||
			m.message.Content.Type == types.MessageTypeVoice) {
		// Only stop if NOT using external player (local playback only)
		if m.externalAudioPlayer == nil {
			m.audioRenderer.GetAudioPlayer().Stop()
		}
		// If using external player, audio continues in background
	}
	m.stopUIRefresh()
	m.visible = false
}

// IsVisible returns whether the media viewer is visible.
func (m *MediaViewerComponent) IsVisible() bool {
	return m.visible
}

// SetSize sets the media viewer size.
func (m *MediaViewerComponent) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Update renderer dimensions - maximize available space while leaving room for borders
	m.imageRenderer.SetDimensions(width-6, height-6)
	m.mosaicRenderer.SetDimensions(width-6, height-6)
	m.kittyRenderer.SetDimensions(width-6, height-6)
	m.sixelRenderer.SetDimensions(width-6, height-6)
	m.audioRenderer.SetMaxWidth(width - 6)
}

// SetExternalAudioPlayer sets an external audio player for background playback.
func (m *MediaViewerComponent) SetExternalAudioPlayer(player *media.AudioPlayer) {
	m.externalAudioPlayer = player
}

// renderMedia renders the media based on its type.
func (m *MediaViewerComponent) renderMedia() {
	if m.message == nil {
		return
	}

	var err error
	var contentWidth, contentHeight int

	// Determine if we'll use fullscreen mode for this media
	useFullscreen := (m.detectedProtocol == media.ProtocolKitty || m.detectedProtocol == media.ProtocolSixel) &&
		m.message.Content.Type == types.MessageTypePhoto

	if useFullscreen {
		// FULLSCREEN MODE: Account for margins
		// Left margin: 5 spaces
		// Top margin: 2 lines
		// Bottom overlay: 3 lines
		contentWidth = m.width - 5      // Reserve 5 chars for left margin
		contentHeight = m.height - 5    // Reserve 2 lines for top margin + 3 lines for overlay text
	} else {
		// MODAL MODE: Account for border and padding
		// Border takes: 2 chars left + 2 chars right = 4 chars
		// Padding takes: 2 chars left + 2 chars right = 4 chars
		// Title takes: 1 line + 2 blank lines = 3 lines
		// Footer takes: 2 blank lines + 1 line = 3 lines
		// Border takes: 2 lines (top + bottom)
		contentWidth = m.width - 4  // Minimal border/padding
		contentHeight = m.height - 6
	}

	switch m.message.Content.Type {
	case types.MessageTypePhoto:
		// Render image using the best available protocol
		m.content, err = m.renderImageWithBestProtocol(m.downloadedPath, contentWidth, contentHeight)
		if err != nil {
			m.renderError = err
		}

	case types.MessageTypeVideo:
		// For videos, show a placeholder with metadata
		m.content = m.renderVideoPlaceholder()

	case types.MessageTypeAudio:
		// Render audio metadata and waveform
		m.content, err = m.audioRenderer.RenderFullAudioView(m.downloadedPath, m.message.Content.Media)
		if err != nil {
			m.renderError = err
		}

	case types.MessageTypeVoice:
		// Render voice message with dedicated voice UI
		m.content, err = m.audioRenderer.RenderFullVoiceView(m.downloadedPath, m.message.Content.Media)
		if err != nil {
			m.renderError = err
		}

	case types.MessageTypeVideoNote:
		// For video notes (round messages), show a placeholder with metadata
		m.content = m.renderVideoNotePlaceholder()

	case types.MessageTypeDocument:
		// Show document info
		m.content = m.renderDocumentInfo()

	default:
		m.content = "Unsupported media type"
	}
}

// renderVideoPlaceholder renders a placeholder for video files.
func (m *MediaViewerComponent) renderVideoPlaceholder() string {
	var sb strings.Builder

	sb.WriteString("🎥 Video File\n")
	sb.WriteString(strings.Repeat("═", min(m.width-6, 60)))
	sb.WriteString("\n\n")

	if m.message.Content.Document != nil {
		sb.WriteString(fmt.Sprintf("File: %s\n", m.message.Content.Document.FileName))
	}

	if m.message.Content.Media != nil {
		if m.message.Content.Media.Width > 0 && m.message.Content.Media.Height > 0 {
			sb.WriteString(fmt.Sprintf("Resolution: %dx%d\n", m.message.Content.Media.Width, m.message.Content.Media.Height))
		}
		if m.message.Content.Media.Duration > 0 {
			sb.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(m.message.Content.Media.Duration)))
		}
		if m.message.Content.Media.Size > 0 {
			sb.WriteString(fmt.Sprintf("Size: %s\n", formatFileSize(m.message.Content.Media.Size)))
		}
	}

	sb.WriteString(fmt.Sprintf("\nPath: %s\n", m.downloadedPath))

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width-6, 60)))
	sb.WriteString("\n\n")
	sb.WriteString("Video preview is not available in terminal.\n")
	sb.WriteString("Use an external player to view this file.\n")

	return sb.String()
}

// renderVideoNotePlaceholder renders a placeholder for video note (round video message) files.
func (m *MediaViewerComponent) renderVideoNotePlaceholder() string {
	var sb strings.Builder

	sb.WriteString("🎥 Video Message (Round)\n")
	sb.WriteString(strings.Repeat("═", min(m.width-6, 60)))
	sb.WriteString("\n\n")

	if m.message.Content.Media != nil {
		if m.message.Content.Media.Width > 0 && m.message.Content.Media.Height > 0 {
			sb.WriteString(fmt.Sprintf("Resolution: %dx%d (circular)\n", m.message.Content.Media.Width, m.message.Content.Media.Height))
		}
		if m.message.Content.Media.Duration > 0 {
			sb.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(m.message.Content.Media.Duration)))
		}
		if m.message.Content.Media.Size > 0 {
			sb.WriteString(fmt.Sprintf("Size: %s\n", formatFileSize(m.message.Content.Media.Size)))
		}
	}

	sb.WriteString(fmt.Sprintf("\nPath: %s\n", m.downloadedPath))

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width-6, 60)))
	sb.WriteString("\n\n")
	sb.WriteString("Video note preview is not available in terminal.\n")
	sb.WriteString("Use an external player to view this file.\n")

	return sb.String()
}

// renderDocumentInfo renders document information.
func (m *MediaViewerComponent) renderDocumentInfo() string {
	var sb strings.Builder

	sb.WriteString("📄 Document\n")
	sb.WriteString(strings.Repeat("═", min(m.width-6, 60)))
	sb.WriteString("\n\n")

	if m.message.Content.Document != nil {
		sb.WriteString(fmt.Sprintf("File: %s\n", m.message.Content.Document.FileName))
		if m.message.Content.Document.MimeType != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", m.message.Content.Document.MimeType))
		}
	}

	if m.message.Content.Media != nil && m.message.Content.Media.Size > 0 {
		sb.WriteString(fmt.Sprintf("Size: %s\n", formatFileSize(m.message.Content.Media.Size)))
	}

	sb.WriteString(fmt.Sprintf("\nPath: %s\n", m.downloadedPath))

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width-6, 60)))
	sb.WriteString("\n\n")
	sb.WriteString("Open this file with an appropriate application.\n")

	return sb.String()
}

// getTitle returns the title for the media viewer.
func (m *MediaViewerComponent) getTitle() string {
	if m.message == nil {
		return "Media Viewer"
	}

	switch m.message.Content.Type {
	case types.MessageTypePhoto:
		return "📷 Image Viewer"
	case types.MessageTypeVideo:
		return "🎥 Video Info"
	case types.MessageTypeVideoNote:
		return "🎥 Video Message"
	case types.MessageTypeAudio:
		return "🎵 Audio Player"
	case types.MessageTypeVoice:
		return "🎤 Voice Message"
	case types.MessageTypeDocument:
		return "📄 Document Info"
	default:
		return "Media Viewer"
	}
}

// MediaViewerDismissedMsg is sent when the media viewer is dismissed.
type MediaViewerDismissedMsg struct{}

// MediaDownloadRequestMsg requests a media download.
type MediaDownloadRequestMsg struct {
	Message *types.Message
}

// MediaDownloadedMsg indicates media has been downloaded.
type MediaDownloadedMsg struct {
	Path string
}

// RefreshUIMsg triggers a UI refresh for real-time updates.
type RefreshUIMsg struct{}

// scheduleRefresh schedules the next UI refresh.
func (m *MediaViewerComponent) scheduleRefresh() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return RefreshUIMsg{}
	})
}

// startUIRefresh starts periodic UI refresh.
func (m *MediaViewerComponent) startUIRefresh() {
	// No-op: refresh is handled via scheduleRefresh command
}

// stopUIRefresh stops periodic UI refresh.
func (m *MediaViewerComponent) stopUIRefresh() {
	// No-op: ticker is managed by Bubble Tea commands
}

// IsFullscreenMode returns true if the media viewer should use fullscreen mode.
// Fullscreen mode is used for Kitty/Sixel images to provide high-quality rendering
// without Lipgloss layout interference.
func (m *MediaViewerComponent) IsFullscreenMode() bool {
	if !m.visible {
		return false
	}

	// Fullscreen mode only for pixel protocols with photos
	return (m.detectedProtocol == media.ProtocolKitty || m.detectedProtocol == media.ProtocolSixel) &&
		m.message != nil &&
		m.message.Content.Type == types.MessageTypePhoto &&
		!m.downloading &&
		m.renderError == nil
}

// renderImageWithBestProtocol renders an image using the best available graphics protocol.
// It automatically selects between Kitty, Sixel, Unicode Mosaic, or ASCII based on terminal capabilities.
//
// For Kitty/Sixel protocols, this will render at full quality for fullscreen mode.
// The fullscreen mode bypasses Lipgloss entirely, avoiding cursor movement issues.
func (m *MediaViewerComponent) renderImageWithBestProtocol(filePath string, width, height int) (string, error) {
	// Update dimensions for all renderers
	m.kittyRenderer.SetDimensions(width, height)
	m.sixelRenderer.SetDimensions(width, height)
	m.mosaicRenderer.SetDimensions(width, height)
	m.imageRenderer.SetDimensions(width, height)

	// Use the actual detected protocol to render at highest quality
	switch m.detectedProtocol {
	case media.ProtocolKitty:
		// Use Kitty renderer for full quality pixel-based rendering
		return m.kittyRenderer.RenderImageFile(filePath)

	case media.ProtocolSixel:
		// Use Sixel renderer for full quality pixel-based rendering
		return m.sixelRenderer.RenderImageFile(filePath)

	case media.ProtocolUnicodeMosaic:
		// Use Unicode Mosaic for text-based rendering
		return m.mosaicRenderer.RenderImageFile(filePath)

	case media.ProtocolASCII:
		// Use ASCII for basic text-based rendering
		return m.imageRenderer.RenderImageFile(filePath)

	default:
		// Fallback to Unicode Mosaic as safe default
		return m.mosaicRenderer.RenderImageFile(filePath)
	}
}

// GetDetectedProtocol returns the currently detected graphics protocol.
func (m *MediaViewerComponent) GetDetectedProtocol() media.GraphicsProtocol {
	return m.detectedProtocol
}

// GetProtocolInfo returns information about the detected protocol.
func (m *MediaViewerComponent) GetProtocolInfo() string {
	return m.protocolDetector.GetProtocolInfo()
}
