# Media Rendering - Quick Start Guide

## Overview

The Ithil TUI now supports viewing images and audio files directly in your terminal with ASCII art and waveform visualizations.

## Basic Usage

### Viewing Media

1. **Navigate to a message with media** (image, audio, video, or document)
2. **Press `Enter`** to open the media viewer
3. **Press `ESC` or `Q`** to close the viewer

### What You'll See

#### Images
- **Inline**: Small ASCII art preview (30x10 characters)
- **Modal**: Full-size colored ASCII art

#### Audio Files
- **Inline**: Compact waveform visualization
- **Modal**: Large waveform + detailed metadata (duration, size, type)

#### Videos
- **Modal**: File information and metadata (video playback not supported in terminal)

#### Documents
- **Modal**: File information and path

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Open media viewer for most recent media message |
| `ESC` | Close media viewer |
| `Q` | Close media viewer (alternative) |
| `i` or `a` | Return to input mode |

## Features

### Inline Previews

When media is downloaded, you'll see a small preview directly in the message:

```
📷 Photo
2.4 MB • 1920x1080
✓ Downloaded

[ASCII art preview appears here]

Caption text if present
```

### Full Modal Viewer

Press Enter to see the full view in a centered modal:

```
╭────────────────────────────────────────╮
│         📷 Image Viewer                │
│                                        │
│   [Large ASCII art rendering]          │
│                                        │
│                                        │
│   Press ESC or Q to close              │
╰────────────────────────────────────────╯
```

## Download Status

- **Not Downloaded**: Shows "Press Enter to download and view"
- **Downloaded**: Shows "✓ Downloaded" + inline preview
- **Downloading**: Shows "Downloading media..." in modal

## Tips

1. **Best Image Results**: Works best with:
   - Medium-sized images (not too large)
   - Good contrast
   - Simple compositions
   - Terminal with full ANSI color support

2. **Terminal Settings**:
   - Enable 256-color or true-color support
   - Use a monospace font
   - Increase terminal size for better image quality

3. **Performance**:
   - Large images may take a moment to render
   - ASCII art generation is CPU intensive
   - First render may be slower (no caching yet)

## Supported Media Types

| Type | Preview | Full View | Notes |
|------|---------|-----------|-------|
| Images (JPEG, PNG, GIF, etc.) | ✅ ASCII art | ✅ Large ASCII art | All formats supported |
| Audio files | ✅ Waveform | ✅ Detailed waveform + metadata | Visual only, no playback |
| Voice messages | ✅ Compact waveform | ✅ Detailed waveform | Visual only, no playback |
| Videos | ❌ | ✅ Metadata only | No video rendering in terminal |
| Documents | ❌ | ✅ File info | Shows path and size |

## Examples

### Example 1: Viewing a Photo

```
1. Receive an image in chat
2. Message shows: "📷 Photo" with download hint
3. Press Enter to view
4. Modal opens with full ASCII art rendering
5. Press ESC to close
```

### Example 2: Viewing Audio

```
1. Receive an audio file
2. Message shows: "🎵 Audio" with waveform preview
3. Press Enter to view full details
4. Modal shows: filename, duration, size, larger waveform
5. Press Q to close
```

## Troubleshooting

### No Preview Showing
- Check if media is downloaded (look for "✓ Downloaded")
- Try pressing Enter to trigger download
- Check terminal has ANSI color support

### Garbled/Weird Characters
- Your terminal may not support ANSI colors
- Try a different terminal emulator
- Check terminal is set to UTF-8 encoding

### Modal Won't Close
- Press ESC or Q
- If stuck, press Ctrl+C to exit application

### Image Looks Bad
- ASCII art quality depends on:
  - Terminal size (larger is better)
  - Font size (smaller is better for detail)
  - Original image complexity
  - Color support in terminal

### Slow Rendering
- Large images take longer to convert to ASCII
- This is normal for the first render
- Consider the image size and terminal dimensions

## Configuration

Current configuration is hardcoded but can be modified in source:

**Preview Size** (in `message.go`):
```go
previewWidth := 30   // Character width
previewHeight := 10  // Character height
```

**Full View Size** (in `conversation.go`):
```go
viewerWidth := width - 20   // Viewport width - 20
viewerHeight := height - 10 // Viewport height - 10
```

**Image Renderer Settings** (in `image_renderer.go`):
```go
colored := true      // Enable/disable colors
braille := false     // Enable braille for higher resolution
```

## Limitations

1. **No Video Playback**: Terminal cannot display video frames
2. **No Audio Playback**: Terminal cannot play audio (visual only)
3. **Static GIFs**: Animated GIFs show as static image
4. **ASCII Quality**: Limited by terminal resolution
5. **File Size**: Very large images may be slow or memory-intensive

## Future Features

Coming soon:
- [ ] Navigation between multiple media with arrow keys
- [ ] Zoom controls (+/- keys)
- [ ] Save ASCII art to file
- [ ] Open in external player
- [ ] Real audio waveform generation
- [ ] Cached rendering for faster display
- [ ] User-configurable preview sizes

## Getting Help

For more information:
- Full documentation: `docs/MEDIA_RENDERING.md`
- Implementation details: `IMPLEMENTATION_SUMMARY.md`
- Source code: `internal/media/` and `internal/ui/components/`

Report issues or request features in your project's issue tracker.
