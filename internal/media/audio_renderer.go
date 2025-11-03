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
	sb.WriteString(strings.Repeat("─", min(r.maxWidth, 60)))
	sb.WriteString("\n")

	// File name
	fileName := filepath.Base(filePath)
	sb.WriteString(fmt.Sprintf("📄 %s\n", truncateString(fileName, r.maxWidth-8)))

	// Duration if available
	if media != nil && media.Duration > 0 {
		duration := formatDuration(media.Duration)
		sb.WriteString(fmt.Sprintf("⏱  %s", duration))

		// File size on same line if available
		if media.Size > 0 {
			size := formatFileSize(media.Size)
			sb.WriteString(fmt.Sprintf("  •  💾 %s", size))
		}
		sb.WriteString("\n")
	} else if media != nil && media.Size > 0 {
		// Only size if no duration
		size := formatFileSize(media.Size)
		sb.WriteString(fmt.Sprintf("💾 %s\n", size))
	}

	// MIME type if available
	if media != nil && media.MimeType != "" {
		sb.WriteString(fmt.Sprintf("🔧 %s\n", media.MimeType))
	}

	// Enhanced waveform visualization
	sb.WriteString("\n")
	sb.WriteString(r.renderEnhancedWaveform(media))
	sb.WriteString("\n")

	// Progress bar (simulated)
	sb.WriteString(r.renderProgressBar(0, media))

	return sb.String(), nil
}

// RenderVoicePreview renders a preview of a voice message.
func (r *AudioRenderer) RenderVoicePreview(filePath string, media *types.Media) (string, error) {
	var sb strings.Builder

	// Voice icon and basic info
	sb.WriteString("🎤 Voice Message\n")
	sb.WriteString(strings.Repeat("─", min(r.maxWidth, 60)))
	sb.WriteString("\n")

	// Duration if available
	if media != nil && media.Duration > 0 {
		duration := formatDuration(media.Duration)
		sb.WriteString(fmt.Sprintf("⏱  %s\n", duration))
	}

	// Enhanced waveform visualization for voice
	sb.WriteString("\n")
	sb.WriteString(r.renderEnhancedWaveform(media))
	sb.WriteString("\n")

	// Progress bar (simulated)
	sb.WriteString(r.renderProgressBar(0, media))

	return sb.String(), nil
}

// RenderFullAudioView renders a full audio view for the modal.
func (r *AudioRenderer) RenderFullAudioView(filePath string, media *types.Media) (string, error) {
	var sb strings.Builder

	// Header with visual appeal
	sb.WriteString("🎵 ════════════════════════════════════════\n")
	sb.WriteString("          AUDIO PLAYER\n")
	sb.WriteString("════════════════════════════════════════\n\n")

	// File info section
	fileName := filepath.Base(filePath)
	sb.WriteString(fmt.Sprintf("📄 File: %s\n", fileName))

	if media != nil {
		var details []string
		if media.Duration > 0 {
			duration := formatDuration(media.Duration)
			details = append(details, fmt.Sprintf("⏱  %s", duration))
		}
		if media.Size > 0 {
			size := formatFileSize(media.Size)
			details = append(details, fmt.Sprintf("💾 %s", size))
		}
		if media.MimeType != "" {
			details = append(details, fmt.Sprintf("🔧 %s", media.MimeType))
		}

		if len(details) > 0 {
			sb.WriteString("\n")
			for _, detail := range details {
				sb.WriteString(fmt.Sprintf("   %s\n", detail))
			}
		}
	}

	// File path
	sb.WriteString(fmt.Sprintf("\n📁 Path: %s\n", filePath))

	// Waveform visualization section
	sb.WriteString("\n")
	sb.WriteString("┌" + strings.Repeat("─", min(r.maxWidth-2, 60)) + "┐\n")
	sb.WriteString("│ WAVEFORM" + strings.Repeat(" ", min(r.maxWidth-11, 51)) + "│\n")
	sb.WriteString("├" + strings.Repeat("─", min(r.maxWidth-2, 60)) + "┤\n")

	// Render multiple lines of waveform for detail
	waveformLines := r.renderLargeWaveform()
	for _, line := range strings.Split(waveformLines, "\n") {
		sb.WriteString("│ " + line)
		// Pad to box width
		padding := min(r.maxWidth-4, 60) - len(line)
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString("│\n")
	}

	sb.WriteString("└" + strings.Repeat("─", min(r.maxWidth-2, 60)) + "┘\n")

	// Progress bar
	sb.WriteString("\n")
	sb.WriteString(r.renderProgressBar(0, media))
	sb.WriteString("\n")

	// Playback controls section
	sb.WriteString("\n")
	sb.WriteString("┌" + strings.Repeat("─", min(r.maxWidth-2, 60)) + "┐\n")
	sb.WriteString("│ PLAYBACK CONTROLS" + strings.Repeat(" ", min(r.maxWidth-20, 42)) + "│\n")
	sb.WriteString("├" + strings.Repeat("─", min(r.maxWidth-2, 60)) + "┤\n")
	sb.WriteString("│                                                           │\n")
	sb.WriteString("│   ⏯  Space      Play/Pause                                │\n")
	sb.WriteString("│   ⏮  ←          Skip Back 5s                              │\n")
	sb.WriteString("│   ⏭  →          Skip Forward 5s                           │\n")
	sb.WriteString("│   🔊  ↑          Volume Up                                 │\n")
	sb.WriteString("│   🔉  ↓          Volume Down                               │\n")
	sb.WriteString("│   ⏹  Q          Stop & Close                               │\n")
	sb.WriteString("│                                                           │\n")
	sb.WriteString("└" + strings.Repeat("─", min(r.maxWidth-2, 60)) + "┘\n")

	// External player instructions
	sb.WriteString("\n")
	sb.WriteString("╔" + strings.Repeat("═", min(r.maxWidth-2, 60)) + "╗\n")
	sb.WriteString("║ ⚠  NOTE: Terminal audio playback not yet supported       ║\n")
	sb.WriteString("╠" + strings.Repeat("═", min(r.maxWidth-2, 60)) + "╣\n")
	sb.WriteString("║                                                           ║\n")
	sb.WriteString("║ To listen to this audio file, use:                       ║\n")
	sb.WriteString("║                                                           ║\n")
	sb.WriteString("║   • macOS:    open \"<path>\"                             ║\n")
	sb.WriteString("║   • Linux:    xdg-open \"<path>\" or mpv \"<path>\"      ║\n")
	sb.WriteString("║   • Windows:  start \"<path>\"                           ║\n")
	sb.WriteString("║                                                           ║\n")
	sb.WriteString("║ Or copy the path above and open with your preferred      ║\n")
	sb.WriteString("║ audio player application.                                ║\n")
	sb.WriteString("║                                                           ║\n")
	sb.WriteString("╚" + strings.Repeat("═", min(r.maxWidth-2, 60)) + "╝\n")

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

// renderEnhancedWaveform renders an enhanced multi-line waveform visualization.
func (r *AudioRenderer) renderEnhancedWaveform(media *types.Media) string {
	var sb strings.Builder

	// Waveform characters from lowest to highest
	waveChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	// Calculate width for waveform
	width := min(r.maxWidth-4, 60)
	if width < 20 {
		width = 20
	}

	// Generate dynamic waveform based on duration or use default pattern
	duration := 60 // default
	if media != nil && media.Duration > 0 {
		duration = media.Duration
	}

	// Create waveform line
	sb.WriteString("  ")
	for i := 0; i < width; i++ {
		// Create more interesting pattern based on position and duration
		// Use sine-like pattern for more realistic audio visualization
		progress := float64(i) / float64(width)
		pattern := (i*7 + duration*3) % len(waveChars)

		// Add variation based on position (simulate audio dynamics)
		if i%5 == 0 || i%7 == 0 {
			pattern = min(pattern+2, len(waveChars)-1)
		}
		if i%11 == 0 {
			pattern = min(pattern+3, len(waveChars)-1)
		}

		// Lower amplitude at edges (fade in/out effect)
		if progress < 0.1 || progress > 0.9 {
			pattern = pattern / 2
		}

		sb.WriteString(waveChars[pattern])
	}

	return sb.String()
}

// renderProgressBar renders a progress bar for audio playback.
func (r *AudioRenderer) renderProgressBar(currentTime int, media *types.Media) string {
	var sb strings.Builder

	// Calculate bar width
	barWidth := min(r.maxWidth-4, 60)
	if barWidth < 20 {
		barWidth = 20
	}

	duration := 100 // default
	if media != nil && media.Duration > 0 {
		duration = media.Duration
	}

	// Calculate progress (for now, always at start)
	progress := float64(currentTime) / float64(duration)
	if progress > 1.0 {
		progress = 1.0
	}
	filledWidth := int(progress * float64(barWidth))

	// Build progress bar
	sb.WriteString("  ")
	sb.WriteString(formatDuration(currentTime))
	sb.WriteString(" ")

	// Progress bar
	sb.WriteString("▕")
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			sb.WriteString("━")
		} else if i == filledWidth {
			sb.WriteString("●") // playhead
		} else {
			sb.WriteString("─")
		}
	}
	sb.WriteString("▏")

	sb.WriteString(" ")
	sb.WriteString(formatDuration(duration))

	return sb.String()
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
