// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/media"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/pkg/types"
)

// MediaViewerComponent represents a modal for viewing media files.
type MediaViewerComponent struct {
	message        *types.Message
	mediaPath      string
	width          int
	height         int
	visible        bool
	content        string
	imageRenderer  *media.ImageRenderer
	mosaicRenderer *media.MosaicRenderer
	audioRenderer  *media.AudioRenderer
	renderError    error
	downloading    bool
	downloadedPath string
}

// NewMediaViewerComponent creates a new media viewer component.
func NewMediaViewerComponent(width, height int) *MediaViewerComponent {
	return &MediaViewerComponent{
		width:          width,
		height:         height,
		visible:        false,
		imageRenderer:  media.NewImageRenderer(width-6, height-6, true),
		mosaicRenderer: media.NewMosaicRenderer(width-6, height-6, true),
		audioRenderer:  media.NewAudioRenderer(width - 6),
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
		switch msg.String() {
		case "esc", "q":
			m.visible = false
			return m, func() tea.Msg {
				return MediaViewerDismissedMsg{}
			}
		}
	case MediaDownloadedMsg:
		// Media has been downloaded, render it
		m.downloading = false
		m.downloadedPath = msg.Path
		m.renderMedia()
		return m, nil
	}

	return m, nil
}

// View renders the media viewer component.
func (m *MediaViewerComponent) View() string {
	if !m.visible {
		return ""
	}

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
	return nil
}

// Hide hides the media viewer.
func (m *MediaViewerComponent) Hide() {
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
	m.audioRenderer.SetMaxWidth(width - 6)
}

// renderMedia renders the media based on its type.
func (m *MediaViewerComponent) renderMedia() {
	if m.message == nil {
		return
	}

	var err error
	// Maximize content area - account for border (4 chars), padding (4 chars), title/footer (6 lines)
	contentWidth := m.width - 8
	contentHeight := m.height - 12

	switch m.message.Content.Type {
	case types.MessageTypePhoto:
		// Render image using mosaic (Unicode half-blocks) for better quality
		m.mosaicRenderer.SetDimensions(contentWidth, contentHeight)
		m.content, err = m.mosaicRenderer.RenderImageFile(m.downloadedPath)
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
		// Render voice message
		m.content, err = m.audioRenderer.RenderFullAudioView(m.downloadedPath, m.message.Content.Media)
		if err != nil {
			m.renderError = err
		}

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
