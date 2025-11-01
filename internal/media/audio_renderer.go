// Package media provides media rendering utilities for terminal display.
package media

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/lvcasx1/ithil/pkg/types"
)

// AudioRenderer handles rendering audio metadata and waveforms in the terminal.
type AudioRenderer struct {
	maxWidth int
}

// NewAudioRenderer creates a new audio renderer.
func NewAudioRenderer(maxWidth int) *AudioRenderer {
	return &AudioRenderer{
		maxWidth: maxWidth,
	}
}

// RenderAudioPreview renders a preview of an audio file with metadata.
func (r *AudioRenderer) RenderAudioPreview(filePath string, media *types.Media) (string, error) {
	var sb strings.Builder

	// Audio icon and basic info
	sb.WriteString("🎵 Audio File\n")
	sb.WriteString(strings.Repeat("─", min(r.maxWidth, 40)))
	sb.WriteString("\n")

	// File name
	fileName := filepath.Base(filePath)
	sb.WriteString(fmt.Sprintf("File: %s\n", truncateString(fileName, r.maxWidth-6)))

	// Duration if available
	if media != nil && media.Duration > 0 {
		duration := formatDuration(media.Duration)
		sb.WriteString(fmt.Sprintf("Duration: %s\n", duration))
	}

	// File size if available
	if media != nil && media.Size > 0 {
		size := formatFileSize(media.Size)
		sb.WriteString(fmt.Sprintf("Size: %s\n", size))
	}

	// MIME type if available
	if media != nil && media.MimeType != "" {
		sb.WriteString(fmt.Sprintf("Type: %s\n", media.MimeType))
	}

	// Simple waveform visualization (placeholder)
	sb.WriteString("\n")
	sb.WriteString(r.renderSimpleWaveform())

	return sb.String(), nil
}

// RenderVoicePreview renders a preview of a voice message.
func (r *AudioRenderer) RenderVoicePreview(filePath string, media *types.Media) (string, error) {
	var sb strings.Builder

	// Voice icon and basic info
	sb.WriteString("🎤 Voice Message\n")
	sb.WriteString(strings.Repeat("─", min(r.maxWidth, 40)))
	sb.WriteString("\n")

	// Duration if available
	if media != nil && media.Duration > 0 {
		duration := formatDuration(media.Duration)
		sb.WriteString(fmt.Sprintf("Duration: %s\n", duration))
	}

	// Simple waveform visualization
	sb.WriteString("\n")
	sb.WriteString(r.renderSimpleWaveform())

	return sb.String(), nil
}

// RenderFullAudioView renders a full audio view for the modal.
func (r *AudioRenderer) RenderFullAudioView(filePath string, media *types.Media) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("🎵 Audio Player\n")
	sb.WriteString(strings.Repeat("═", min(r.maxWidth, 60)))
	sb.WriteString("\n\n")

	// File info
	fileName := filepath.Base(filePath)
	sb.WriteString(fmt.Sprintf("File: %s\n", fileName))

	if media != nil {
		if media.Duration > 0 {
			duration := formatDuration(media.Duration)
			sb.WriteString(fmt.Sprintf("Duration: %s\n", duration))
		}
		if media.Size > 0 {
			size := formatFileSize(media.Size)
			sb.WriteString(fmt.Sprintf("Size: %s\n", size))
		}
		if media.MimeType != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", media.MimeType))
		}
	}

	// File path
	sb.WriteString(fmt.Sprintf("Path: %s\n", filePath))

	// Waveform visualization (larger for modal)
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(r.maxWidth, 60)))
	sb.WriteString("\n")
	sb.WriteString(r.renderLargeWaveform())
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(r.maxWidth, 60)))

	// Playback controls (visual only - actual playback not implemented)
	sb.WriteString("\n\n")
	sb.WriteString("Playback Controls:\n")
	sb.WriteString("  [Space] Play/Pause  [←][→] Seek  [↑][↓] Volume\n")
	sb.WriteString("\n")
	sb.WriteString("Note: Audio playback in terminal is not yet supported.\n")
	sb.WriteString("      Use an external player to listen to this file.\n")

	return sb.String(), nil
}

// renderSimpleWaveform renders a simple waveform visualization.
func (r *AudioRenderer) renderSimpleWaveform() string {
	// Generate a simple static waveform for preview
	waveform := "▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▂▃▄▅▆▇█▇▆▅▄▃▂▁"
	width := min(r.maxWidth-4, len(waveform))
	if width < len(waveform) {
		waveform = waveform[:width]
	}
	return fmt.Sprintf("  %s", waveform)
}

// renderLargeWaveform renders a larger waveform visualization for the modal.
func (r *AudioRenderer) renderLargeWaveform() string {
	var sb strings.Builder

	// Generate multiple lines of waveform for a more detailed view
	waveChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	width := min(r.maxWidth-4, 60)

	// Create 3 lines of waveform
	for line := 0; line < 3; line++ {
		sb.WriteString("  ")
		for i := 0; i < width; i++ {
			// Simple pattern-based waveform
			height := (i*3 + line*7) % len(waveChars)
			sb.WriteString(waveChars[height])
		}
		if line < 2 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// SetMaxWidth sets the maximum width for rendering.
func (r *AudioRenderer) SetMaxWidth(width int) {
	r.maxWidth = width
}

// Helper functions

func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
