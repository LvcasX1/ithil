# Thumbnail Generator for ithil

This document describes the thumbnail generation system for ithil's TUI Telegram client.

## Overview

The `ThumbnailGenerator` provides an efficient, protocol-aware system for generating small preview thumbnails of images in the message list. It supports all graphics protocols (Kitty, Sixel, Unicode Mosaic, ASCII) and includes intelligent caching to avoid regenerating thumbnails.

## Features

- **Protocol-Aware**: Automatically uses the best rendering method for your terminal
- **Intelligent Caching**: Thumbnails are cached in memory to avoid regeneration
- **Thread-Safe**: Safe for concurrent use in multi-threaded applications
- **Async Support**: Non-blocking thumbnail generation for smooth UI
- **Batch Preloading**: Preload thumbnails for multiple images at once
- **Configurable**: Adjustable dimensions, colors, and protocols
- **Memory Efficient**: LRU-style cache with configurable size limits

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/lvcasx1/ithil/internal/media"
    "log"
)

func main() {
    // Auto-detect the best graphics protocol
    detector := media.NewProtocolDetector()
    protocol := detector.DetectProtocol()

    // Create a thumbnail generator (20x10 characters)
    generator := media.NewThumbnailGenerator(20, 10, protocol)

    // Generate a thumbnail
    thumbnail, err := generator.GenerateThumbnail("/path/to/image.jpg")
    if err != nil {
        log.Fatalf("Failed to generate thumbnail: %v", err)
    }

    // Display in terminal
    fmt.Print(thumbnail)
}
```

### Async Generation (Recommended for TUI)

```go
// Generate thumbnail without blocking the UI
generator.GenerateThumbnailAsync("/path/to/image.jpg",
    func(thumbnail string, err error) {
        if err != nil {
            log.Printf("Error: %v", err)
            return
        }

        // In Bubbletea, send a message to update the UI
        // p.Send(ThumbnailReadyMsg{Thumbnail: thumbnail})
    })
```

### Using Options

```go
opts := &media.ThumbnailGeneratorOptions{
    Width:      25,           // 25 characters wide
    Height:     12,           // 12 characters tall
    AutoDetect: true,         // Auto-detect protocol
    Colored:    true,         // Enable color
    CacheSize:  200,          // Cache up to 200 thumbnails
}

generator := media.NewThumbnailGeneratorWithOptions(opts)
```

## Integration with ithil

### In Conversation Model

```go
type ConversationModel struct {
    // ... existing fields
    thumbnailGen *media.ThumbnailGenerator
}

func NewConversationModel() ConversationModel {
    detector := media.NewProtocolDetector()
    protocol := detector.DetectProtocol()

    return ConversationModel{
        thumbnailGen: media.NewThumbnailGenerator(20, 10, protocol),
    }
}
```

### When New Media Message Arrives

```go
case NewMediaMessageMsg:
    // Start async thumbnail generation
    m.thumbnailGen.GenerateThumbnailAsync(msg.LocalPath,
        func(thumbnail string, err error) {
            if err != nil {
                return
            }

            // Send update to UI
            p.Send(ThumbnailReadyMsg{
                MessageID: msg.ID,
                Thumbnail: thumbnail,
            })
        })

    return m, nil

case ThumbnailReadyMsg:
    // Update the message with thumbnail
    for i := range m.messages {
        if m.messages[i].ID == msg.MessageID {
            m.messages[i].Thumbnail = msg.Thumbnail
            break
        }
    }
    return m, nil
```

### When Rendering Messages

```go
func (m ConversationModel) renderMessage(msg types.Message) string {
    if msg.Content.Type == types.MessageTypePhoto {
        if msg.Thumbnail != "" {
            // Thumbnail ready - display it
            return msg.Thumbnail
        } else {
            // Still loading - show placeholder
            return "[Loading thumbnail...]"
        }
    }

    // Regular text message
    return msg.Content.Text
}
```

### Preloading for Chat History

```go
// When loading chat history, preload thumbnails
func (m *ConversationModel) loadChatHistory(messages []types.Message) {
    var imagePaths []string

    // Collect all image paths
    for _, msg := range messages {
        if msg.Content.Type == types.MessageTypePhoto && msg.Content.Media.LocalPath != "" {
            imagePaths = append(imagePaths, msg.Content.Media.LocalPath)
        }
    }

    // Preload all thumbnails
    m.thumbnailGen.PreloadThumbnails(imagePaths, func(path, thumbnail string, err error) {
        if err == nil {
            // Thumbnail is now cached and ready
            // Send UI update if needed
        }
    })
}
```

## API Reference

### ThumbnailGenerator

#### Constructors

- `NewThumbnailGenerator(width, height int, protocol GraphicsProtocol) *ThumbnailGenerator`
  - Creates generator with specified dimensions (0 = use defaults: 20x10)
  - Requires a graphics protocol (use `ProtocolDetector.DetectProtocol()`)

- `NewThumbnailGeneratorWithOptions(opts *ThumbnailGeneratorOptions) *ThumbnailGenerator`
  - Creates generator with full configuration options
  - Supports auto-detection, custom cache size, etc.

#### Generation Methods

- `GenerateThumbnail(imagePath string) (string, error)`
  - Generates thumbnail from file (uses cache if available)
  - Returns rendered string ready for terminal display
  - Thread-safe

- `GenerateThumbnailAsync(imagePath string, callback func(string, error))`
  - Async generation with callback
  - Non-blocking, perfect for UI
  - Callback receives (thumbnail, error)

- `GenerateThumbnailFromImage(img image.Image) (string, error)`
  - Generates thumbnail from in-memory image
  - No caching (no file path key)
  - Useful for temporary images

- `PreloadThumbnails(imagePaths []string, callback func(string, string, error))`
  - Preloads multiple thumbnails in parallel
  - Callback receives (path, thumbnail, error) for each image
  - Returns immediately, processes in background

#### Configuration Methods

- `SetProtocol(protocol GraphicsProtocol)`
  - Changes graphics protocol
  - Clears cache (different protocol = different rendering)

- `SetDimensions(width, height int)`
  - Changes thumbnail size
  - Clears cache (different size = different rendering)
  - Ignores invalid dimensions (≤ 0)

- `SetColored(enabled bool)`
  - Enable/disable color output
  - Only affects ASCII and Unicode Mosaic protocols
  - Clears cache if changed

- `GetDimensions() (width, height int)`
  - Returns current dimensions

- `GetProtocol() GraphicsProtocol`
  - Returns current graphics protocol

#### Cache Management Methods

- `ClearCache()`
  - Removes all cached thumbnails
  - Frees memory
  - Next generation will be from scratch

- `RemoveFromCache(imagePath string)`
  - Removes specific thumbnail from cache
  - Use when file has been modified

- `GetCacheSize() int`
  - Returns number of cached thumbnails

#### Validation Methods

- `ValidateImageFile(imagePath string) (bool, error)`
  - Checks if file exists and is a valid image
  - Supported: PNG, JPEG, GIF, WebP, BMP, TIFF
  - Returns (true, nil) if valid
  - Returns (false, error) with reason if invalid

### ThumbnailGeneratorOptions

Configuration structure for `NewThumbnailGeneratorWithOptions`:

```go
type ThumbnailGeneratorOptions struct {
    Width      int              // Thumbnail width in characters (0 = default: 20)
    Height     int              // Thumbnail height in characters (0 = default: 10)
    Protocol   GraphicsProtocol // Protocol to use (ignored if AutoDetect = true)
    Colored    bool             // Enable colored output
    CacheSize  int              // Max cached thumbnails (0 = default: 100)
    AutoDetect bool             // Auto-detect protocol if true
}
```

### Graphics Protocols

```go
const (
    ProtocolKitty          // Highest quality (Kitty terminal)
    ProtocolSixel          // High quality (XTerm, WezTerm, etc.)
    ProtocolUnicodeMosaic  // Good quality (any true-color terminal)
    ProtocolASCII          // Fallback (any terminal)
)
```

## Performance Considerations

### Caching Strategy

The thumbnail generator uses an in-memory LRU-style cache:

- **Cache key**: Absolute file path
- **Cache invalidation**: Automatic when settings change (protocol, dimensions, color)
- **Cache size**: Configurable (default: 100 thumbnails)
- **Memory usage**: ~1-5 KB per cached thumbnail (protocol-dependent)

### Async vs Sync

**Use Async When**:
- In UI thread (to avoid blocking)
- Loading chat history with many images
- User is actively scrolling/browsing

**Use Sync When**:
- Background processing
- CLI tools
- Tests

### Preloading Best Practices

```go
// Good: Preload visible messages only
visibleMessages := m.getVisibleMessages()
var paths []string
for _, msg := range visibleMessages {
    if msg.HasMedia {
        paths = append(paths, msg.LocalPath)
    }
}
m.thumbnailGen.PreloadThumbnails(paths, nil)

// Bad: Preload entire chat history
// This wastes memory and processing for offscreen messages
```

### Protocol Selection

**Performance Ranking** (fastest to slowest):
1. **Kitty**: Native graphics, very fast
2. **Sixel**: Fast, widely supported
3. **Unicode Mosaic**: Medium speed, pixel-by-pixel rendering
4. **ASCII**: Slowest, but most compatible

**Quality Ranking** (best to worst):
1. **Kitty**: True image quality
2. **Sixel**: 256-color palette
3. **Unicode Mosaic**: True color but lower resolution
4. **ASCII**: Grayscale, very low resolution

## Testing

Run tests:
```bash
go test ./internal/media/...
```

Run with coverage:
```bash
go test ./internal/media/... -cover
```

Run benchmarks:
```bash
go test ./internal/media/... -bench=.
```

## Troubleshooting

### "Failed to generate thumbnail: unable to open file"
- File doesn't exist at specified path
- Ensure media was downloaded before generating thumbnail
- Use `ValidateImageFile()` to check first

### "Failed to generate thumbnail: failed to decode image"
- Corrupted image file
- Unsupported image format
- Try re-downloading the image

### Thumbnails look wrong/garbled
- Wrong protocol for terminal
- Use `ProtocolDetector.DetectProtocol()` to auto-detect
- Or manually select appropriate protocol for your terminal

### High memory usage
- Reduce cache size: `opts.CacheSize = 50`
- Clear cache when switching chats: `generator.ClearCache()`
- Don't preload too many thumbnails at once

### Slow thumbnail generation
- Use async generation: `GenerateThumbnailAsync()`
- Preload thumbnails before displaying
- Consider smaller dimensions for faster generation

## Examples

See `thumbnail_example.go` for comprehensive usage examples:
- Basic usage
- Custom dimensions
- Async generation
- Preloading
- Protocol switching
- Validation
- Bubbletea integration
- Cache management
- Error handling

## Future Improvements

Potential enhancements:
- [ ] Disk-based cache for persistence across sessions
- [ ] Thumbnail size optimization based on available terminal width
- [ ] Support for video thumbnails (extract first frame)
- [ ] Progressive loading (low-res → high-res)
- [ ] Lazy loading with scroll position awareness
- [ ] Memory pressure monitoring and automatic cache eviction

## Related Files

- `thumbnail_generator.go` - Main implementation
- `thumbnail_generator_test.go` - Comprehensive tests
- `thumbnail_example.go` - Usage examples
- `protocol_detector.go` - Graphics protocol detection
- `kitty_renderer.go` - Kitty graphics renderer
- `sixel_renderer.go` - Sixel graphics renderer
- `mosaic_renderer.go` - Unicode mosaic renderer
- `image_renderer.go` - ASCII art renderer
