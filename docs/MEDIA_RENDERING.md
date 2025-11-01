# Media Rendering Library

This document describes the media rendering system implemented for the Ithil TUI Telegram client.

## Overview

The media rendering library provides comprehensive support for displaying images and audio files directly in the terminal, with both inline previews and full-screen modal viewing capabilities.

## Architecture

### Components

1. **Media Renderers** (`internal/media/`)
   - `ImageRenderer`: Converts images to ASCII/ANSI art
   - `AudioRenderer`: Displays audio metadata and waveforms

2. **UI Components** (`internal/ui/components/`)
   - `ModalComponent`: Base modal overlay system
   - `MediaViewerComponent`: Specialized modal for viewing media
   - `MessageComponent`: Enhanced with inline media previews

3. **Integration** (`internal/ui/models/`)
   - `ConversationModel`: Handles media viewer display and keyboard controls

## Features

### Image Rendering

Images are rendered as ASCII/ANSI art using the `github.com/TheZoraiz/ascii-image-converter` library.

**Capabilities:**
- Full color ANSI art support
- Configurable dimensions (width x height)
- Automatic aspect ratio preservation
- Support for multiple image formats (JPEG, PNG, GIF, BMP, TIFF, WebP)
- Optional braille mode for higher resolution

**Usage:**
```go
renderer := media.NewImageRenderer(width, height, colored)
asciiArt, err := renderer.RenderImageFile("/path/to/image.jpg")
```

### Audio Rendering

Audio files display metadata and visual waveforms.

**Capabilities:**
- File metadata display (name, size, duration, type)
- Visual waveform representation using Unicode block characters
- Support for both audio files and voice messages
- Placeholder for playback controls (visual only)

**Usage:**
```go
renderer := media.NewAudioRenderer(maxWidth)
preview, err := renderer.RenderAudioPreview("/path/to/audio.mp3", mediaInfo)
```

### Modal System

A flexible modal overlay system for displaying content on top of the main UI.

**Features:**
- Floating overlay centered on screen
- Customizable title and content
- Keyboard-based dismissal (ESC or Q)
- Semi-transparent backdrop effect
- Responsive sizing

### Media Viewer Component

Specialized modal for viewing media with type-specific rendering.

**Supported Media Types:**
1. **Images**: Full ASCII art rendering with color
2. **Videos**: Metadata display (video preview not available in terminal)
3. **Audio**: Waveform visualization and metadata
4. **Voice Messages**: Compact waveform and duration
5. **Documents**: File information and path

**Keyboard Controls:**
- `ESC` or `Q`: Close the media viewer
- Automatic download trigger if media not yet cached

## User Workflow

### Inline Preview

When a message contains media that has been downloaded:

1. A small preview is rendered inline in the message bubble
2. Images show a 30x10 ASCII art thumbnail
3. Audio files show a compact waveform
4. A hint displays: "Press Enter to download and view"

### Full Media View

Press `Enter` while focused on the conversation pane:

1. The system finds the most recent message with media
2. If not downloaded, triggers automatic download
3. Opens a modal with full-size media rendering
4. For images: Large ASCII art (viewport size minus margins)
5. For audio: Detailed waveform and metadata

### Dismissing the Viewer

Press `ESC` or `Q`:
1. Modal closes immediately
2. Returns to conversation view
3. Focus remains on conversation pane

## Implementation Details

### Image Rendering Process

1. Check if file exists at local path
2. Pass file path to `aic_package.Convert()`
3. Configure flags for dimensions and color
4. Receive ASCII art string
5. Render in message or modal

### Preview vs Full View

**Preview (Inline):**
- Small dimensions (30x10 for images)
- Simplified waveforms for audio
- Minimal metadata
- Renders in message bubble

**Full View (Modal):**
- Large dimensions (viewport size - 20 width, - 10 height)
- Detailed waveforms for audio
- Complete metadata
- Renders in centered modal overlay

### Media Download Integration

The system integrates with the existing media download infrastructure:

```go
// When media not downloaded
if !message.Content.Media.IsDownloaded {
    m.downloading = true
    return MediaDownloadRequestMsg{Message: message}
}

// When download completes
case MediaDownloadedMsg:
    m.downloading = false
    m.downloadedPath = msg.Path
    m.renderMedia()
```

**Note:** The actual download implementation should be connected to your `MediaManager` in the `downloadMediaForViewer()` method in `internal/ui/models/conversation.go`.

## Configuration

### Image Renderer Settings

```go
renderer := media.NewImageRenderer(width, height, colored)

// Enable braille mode for higher resolution
renderer.SetBraille(true)

// Adjust dimensions dynamically
renderer.SetDimensions(newWidth, newHeight)

// Toggle colored output
renderer.SetColored(false)
```

### Audio Renderer Settings

```go
renderer := media.NewAudioRenderer(maxWidth)

// Adjust width dynamically
renderer.SetMaxWidth(newWidth)
```

### Media Viewer Settings

```go
viewer := components.NewMediaViewerComponent(width, height)

// Resize dynamically
viewer.SetSize(newWidth, newHeight)

// Show media
viewer.ShowMedia(message, mediaPath)

// Hide viewer
viewer.Hide()
```

## Keyboard Shortcuts

In conversation view:

- `Enter`: Open media viewer for most recent media message
- `ESC`/`Q`: Close media viewer (when open)
- `i`/`a`: Focus input to type (closes viewer if open)
- Normal navigation keys work when viewer is closed

## Limitations

1. **Video Playback**: Terminal cannot play videos; only metadata is shown
2. **Audio Playback**: Terminal cannot play audio; waveform is visual only
3. **Animated GIFs**: Rendered as static ASCII art (first frame)
4. **Large Images**: May take time to render; consider size limits
5. **Terminal Support**: Requires terminal with ANSI color support

## Future Enhancements

Potential improvements:

1. **Actual Waveform Generation**: Use `github.com/mdlayher/waveform` to generate real audio waveforms
2. **External Player Integration**: Launch system player for videos/audio
3. **Navigation Between Media**: Arrow keys to browse all media in conversation
4. **Zoom Controls**: +/- keys to adjust ASCII art size
5. **Save ASCII Art**: Export rendered art to file
6. **Braille Rendering**: Higher resolution for compatible terminals
7. **Video Thumbnails**: Extract and display first frame
8. **Progress Indicators**: Show download/rendering progress
9. **Caching**: Cache rendered ASCII art to avoid re-rendering

## Dependencies

### Core Libraries

- `github.com/TheZoraiz/ascii-image-converter/aic_package` v1.13.1: Image to ASCII conversion
- `github.com/mdlayher/waveform` v0.0.0-20200324155202: Audio waveform generation (installed but not yet integrated)

### Bubbletea Ecosystem

- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling and layout

## File Structure

```
internal/
├── media/
│   ├── image_renderer.go    # Image to ASCII conversion
│   └── audio_renderer.go    # Audio waveform rendering
├── ui/
│   ├── components/
│   │   ├── modal.go         # Base modal component
│   │   ├── mediaviewer.go   # Media viewer modal
│   │   ├── message.go       # Enhanced with previews
│   │   └── utils.go         # Shared utilities
│   └── models/
│       ├── conversation.go  # Media viewer integration
│       └── main.go          # Main model (no changes needed)

docs/
└── MEDIA_RENDERING.md       # This documentation
```

## Testing

### Manual Testing Steps

1. **Send an image to a chat**
   - Download it by viewing the message
   - Verify inline preview shows ASCII art thumbnail
   - Press Enter to open full view
   - Verify large ASCII art renders correctly
   - Press ESC to close

2. **Send an audio file**
   - Download it by viewing the message
   - Verify inline preview shows waveform
   - Press Enter to open full view
   - Verify detailed metadata and larger waveform
   - Press ESC to close

3. **Send a video file**
   - Download it
   - Press Enter to open viewer
   - Verify metadata display
   - Verify message about terminal limitations

4. **Test keyboard navigation**
   - Open media viewer
   - Press ESC to close
   - Press Enter again to reopen
   - Press Q to close
   - Verify focus returns to conversation

### Edge Cases

- Empty/corrupt image files
- Very large images (memory usage)
- Very small images (rendering quality)
- Black and white images
- Unsupported file formats
- Network errors during download
- Rapidly opening/closing viewer

## Performance Considerations

1. **ASCII Art Generation**: CPU intensive for large images
2. **Memory Usage**: Large images require substantial memory
3. **Terminal Rendering**: ANSI codes can be slow on some terminals
4. **Caching**: Currently re-renders on every view (consider caching)

## Contributing

When contributing to the media rendering system:

1. Test with various image formats and sizes
2. Ensure ANSI color codes work on common terminals
3. Handle errors gracefully with user-friendly messages
4. Update this documentation for new features
5. Consider accessibility (non-color terminals)

## License

This implementation follows the project's main license.
