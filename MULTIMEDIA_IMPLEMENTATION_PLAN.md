# Multimedia Implementation Plan for ithil

## Executive Summary

After comprehensive research of **nchat** (C++/ncurses) and **tgt** (Rust/Ratatui) multimedia implementations, **ithil is already more advanced** than both competitors. This document provides a roadmap to complete and enhance ithil's multimedia capabilities.

## Current State Analysis

### ✅ Strengths (Already Implemented)

| Feature | Status | Implementation |
|---------|--------|----------------|
| **Image Rendering** | ✅ Partial | ASCII art via `ascii-image-converter` |
| **Unicode Mosaic** | ✅ Complete | Half-block rendering with true color |
| **Audio Playback** | ✅ Complete | Beep library (MP3, OGG, WAV) |
| **Waveform Visualization** | ✅ Complete | Dynamic waveform with progress bar |
| **Media Downloads** | ✅ Complete | gotd/td downloader with caching |
| **Media Uploads** | ✅ Complete | Photo, document, audio, video |
| **Playback Controls** | ✅ Complete | Play/pause, seek, volume control |

### ⚠️ Gaps (Missing or Incomplete)

| Feature | Status | Priority |
|---------|--------|----------|
| Kitty/Sixel Protocol | ❌ Not implemented | HIGH |
| Protocol Auto-detection | ❌ Not implemented | HIGH |
| Image Thumbnails | ❌ Not implemented | HIGH |
| Download Progress UI | ❌ Not implemented | HIGH |
| Video Thumbnails | ❌ Not implemented | MEDIUM |
| External Player Integration | ❌ Not implemented | MEDIUM |
| Stickers | ❌ Not implemented | MEDIUM |
| Animations (GIFs) | ❌ Not implemented | MEDIUM |
| Polls | ❌ Not implemented | LOW |
| Locations | ❌ Not implemented | LOW |

---

## Competitor Analysis

### nchat Approach (Minimalist)
- **Strategy**: Text-only placeholders with external player delegation
- **Pros**: Clean, minimal dependencies, simple
- **Cons**: No in-terminal media viewing at all
- **Learnings**: Good file status tracking system (not downloaded, downloading, downloaded, failed)

### tgt Approach (Framework-Limited)
- **Strategy**: Emoji indicators, planned image support blocked
- **Pros**: Recognizes all media types correctly
- **Cons**: Ratatui framework limitations prevent implementation
- **Learnings**: Shows the importance of custom implementations over framework dependencies

### ithil Approach (Best-in-Class)
- **Strategy**: Custom rendering with multiple fallback options
- **Advantage**: Pure Go + gotd/td gives complete control
- **Result**: Already more capable than both competitors

---

## Implementation Roadmap

## Phase 1: Complete Image Support (HIGH PRIORITY)

### 1.1 Kitty Graphics Protocol Support

**Goal**: Implement native Kitty terminal graphics protocol for high-fidelity image rendering.

**Location**: `internal/media/kitty_renderer.go` (new file)

**Implementation**:
```go
type KittyRenderer struct {
    maxWidth  int
    maxHeight int
}

func (r *KittyRenderer) RenderImageFile(filePath string) (string, error) {
    // 1. Read image file
    // 2. Encode to PNG if not already
    // 3. Base64 encode image data
    // 4. Generate Kitty escape sequence
    // Format: \033_Ga=T,f=100,t=d,<data>;\033\
    // Where: a=T (transmit), f=100 (PNG), t=d (direct)
}
```

**Kitty Protocol Documentation**: https://sw.kovidgoyal.net/kitty/graphics-protocol/

**Key Features**:
- Direct image transmission via escape sequences
- Supports PNG, JPEG formats
- True pixel-perfect rendering
- Z-index layering support

### 1.2 Sixel Graphics Protocol Support

**Goal**: Implement Sixel protocol for terminals like XTerm, WezTerm, and Alacritty.

**Location**: `internal/media/sixel_renderer.go` (new file)

**Implementation**:
```go
type SixelRenderer struct {
    maxWidth  int
    maxHeight int
}

func (r *SixelRenderer) RenderImageFile(filePath string) (string, error) {
    // 1. Read and decode image
    // 2. Resize to fit terminal
    // 3. Convert to 256-color palette
    // 4. Generate Sixel escape sequence
    // Format: \033Pq<sixel-data>\033\
}
```

**Library Options**:
- Use `github.com/mattn/go-sixel` for encoding
- Or implement minimal custom encoder

### 1.3 Protocol Auto-Detection

**Goal**: Automatically detect and use the best available graphics protocol.

**Location**: `internal/media/protocol_detector.go` (new file)

**Detection Strategy**:
```go
type GraphicsProtocol int

const (
    ProtocolKitty GraphicsProtocol = iota
    ProtocolSixel
    ProtocolUnicodeMosaic
    ProtocolASCII
)

func DetectGraphicsProtocol() GraphicsProtocol {
    // 1. Check $TERM_PROGRAM == "kitty" → Kitty
    // 2. Query terminal with \033[c and check response → Sixel
    // 3. Check $COLORTERM == "truecolor" → Unicode Mosaic
    // 4. Fallback → ASCII
}
```

**Environment Variables to Check**:
- `$TERM_PROGRAM` (kitty, iTerm.app, WezTerm)
- `$COLORTERM` (truecolor, 24bit)
- `$TERM` (xterm-256color, xterm-kitty)

**Terminal Query Method**:
```
Send: \033[c
Response: \033[?<attrs>c where attrs contains '4' for Sixel support
```

### 1.4 Image Caching and Lazy Loading

**Goal**: Implement memory-efficient image caching with LRU eviction.

**Location**: `internal/cache/media_cache.go` (new file)

**Implementation**:
```go
type MediaCache struct {
    mu         sync.RWMutex
    images     map[string]*CachedImage
    maxSize    int64  // Max cache size in bytes
    currentSize int64
    lru        *list.List // For LRU eviction
}

type CachedImage struct {
    Path       string
    Data       []byte
    Size       int64
    AccessTime time.Time
    lruElement *list.Element
}

func (c *MediaCache) Get(path string) (*CachedImage, bool)
func (c *MediaCache) Put(path string, data []byte)
func (c *MediaCache) Evict() // Remove least recently used
```

**Cache Configuration** (`config.yaml`):
```yaml
media:
  cache_size_mb: 100  # Max 100MB of images in memory
  auto_download: true  # Auto-download on scroll
```

### 1.5 Thumbnail Support

**Goal**: Show small preview thumbnails in message list, full size on selection.

**Location**: `internal/media/thumbnail_generator.go` (new file)

**Implementation**:
```go
func GenerateThumbnail(imagePath string, maxWidth, maxHeight int) (string, error) {
    // 1. Load image using imaging library
    // 2. Resize to thumbnail dimensions (e.g., 20x10 chars)
    // 3. Cache thumbnail separately from full image
    // 4. Return thumbnail path
}
```

**UI Integration** (`internal/ui/components/message.go`):
```go
// In message list: Show 20x10 char thumbnail
// On Enter key: Show full-size image in media viewer
```

---

## Phase 2: Enhanced Audio Support (HIGH PRIORITY)

### 2.1 Background Playback

**Goal**: Allow audio playback to continue while navigating other chats.

**Current Issue**: Audio stops when switching chats (media viewer closes).

**Solution**:
1. Move `AudioPlayer` to global state in `MainModel`
2. Add persistent playback indicator in status bar
3. Implement global keyboard shortcuts (Ctrl+P for pause/play)

**Location**: `internal/ui/models/main.go`

```go
type MainModel struct {
    // ... existing fields
    globalAudioPlayer *media.AudioPlayer
    audioPlayingFile  string
}

// Status bar shows: "▶ voice_message.ogg [2:34/5:12]"
```

### 2.2 Enhanced Format Support Verification

**Goal**: Ensure all common audio formats work correctly.

**Formats to Test**:
- ✅ MP3 (already supported via Beep)
- ✅ OGG/Opus (already supported via Beep)
- ✅ WAV (already supported via Beep)
- ❓ M4A (test if Beep supports via external library)
- ❓ FLAC (test if Beep supports via external library)

**Action**: Add tests for all formats or document limitations.

### 2.3 Speed Control

**Goal**: Add playback speed control (0.5x, 1x, 1.5x, 2x) for voice messages.

**Location**: `internal/media/audio_player_cgo.go`

**Implementation**:
```go
type AudioPlayer struct {
    // ... existing fields
    speed *effects.Resampler
}

func (p *AudioPlayer) SetSpeed(speed float64) error {
    // Use beep.Resample for speed adjustment
    // speed: 0.5 = half speed, 2.0 = double speed
}
```

**UI Keyboard Shortcuts**:
- `[` - Decrease speed (0.75x → 0.5x)
- `]` - Increase speed (1x → 1.25x → 1.5x → 2x)

---

## Phase 3: Video Support (MEDIUM PRIORITY)

### 3.1 Video Thumbnail Extraction

**Goal**: Extract and display first frame of video as static image.

**Approach**: Hybrid - show thumbnail in-app, play externally.

**Library Options**:
1. **Pure Go**: `github.com/3d0c/gmf` (FFmpeg bindings, requires CGo)
2. **External Tool**: Call `ffmpeg` binary if available
3. **Fallback**: Show document icon with duration/size

**Location**: `internal/media/video_renderer.go` (new file)

**Implementation**:
```go
func ExtractVideoThumbnail(videoPath string) (string, error) {
    // Option 1: Use ffmpeg binary
    // ffmpeg -i video.mp4 -ss 00:00:01 -vframes 1 thumb.jpg

    // Option 2: Use gmf library (if CGo enabled)
    // Extract first frame programmatically

    // Option 3: Return placeholder icon
}
```

### 3.2 External Player Integration

**Goal**: Open videos in system default player (like nchat).

**Location**: `internal/media/external_player.go` (new file)

**Implementation**:
```go
func OpenInExternalPlayer(filePath string) error {
    switch runtime.GOOS {
    case "darwin":
        return exec.Command("open", filePath).Run()
    case "linux":
        return exec.Command("xdg-open", filePath).Run()
    case "windows":
        return exec.Command("cmd", "/c", "start", filePath).Run()
    default:
        return fmt.Errorf("unsupported platform")
    }
}
```

**UI Integration** (`internal/ui/components/mediaviewer.go`):
```go
// On video message:
// - Show: [Thumbnail] + "Press 'o' to open in external player"
// - 'o' key → OpenInExternalPlayer()
// - 'Enter' key → Same as 'o' (for consistency)
```

### 3.3 ASCII Video Playback (EXPERIMENTAL)

**Goal**: Render short video clips (video notes) as ASCII animation.

**Feasibility**: Low - requires significant CPU and complex frame timing.

**Decision**: **SKIP for now** - external player is more practical.

---

## Phase 4: File Management Enhancements (MEDIUM PRIORITY)

### 4.1 Download Progress Indicator

**Goal**: Show download progress for media files.

**Inspired By**: nchat's status tracking (not downloaded, downloading, downloaded, failed).

**Location**: `internal/telegram/media.go` (enhance existing)

**Implementation**:
```go
type DownloadStatus int

const (
    StatusNotDownloaded DownloadStatus = iota
    StatusDownloading
    StatusDownloaded
    StatusFailed
)

type DownloadProgress struct {
    Status      DownloadStatus
    BytesTotal  int64
    BytesLoaded int64
    Error       error
}

// In media manager:
func (m *MediaManager) DownloadWithProgress(
    ctx context.Context,
    location tg.InputFileLocationClass,
    progressChan chan<- DownloadProgress,
) error
```

**UI Integration** (`internal/ui/components/message.go`):
```go
// Message display:
// [📷 Photo] → Not downloaded
// [📥 Downloading... 45%] → Downloading
// [✅ Photo] → Downloaded
// [❌ Download failed] → Failed
```

### 4.2 Manual Save to Custom Location

**Goal**: Add `Ctrl+S` to save media file to user-specified location.

**Location**: `internal/ui/components/mediaviewer.go`

**Implementation**:
```go
// On Ctrl+S:
// 1. Show file picker dialog (or prompt for path)
// 2. Copy file from cache to user location
// 3. Show confirmation message
```

**Alternative (TUI-friendly)**:
```go
// On Ctrl+S:
// 1. Show text input: "Save to: ~/Downloads/filename.jpg"
// 2. User can edit path
// 3. Press Enter to save
```

### 4.3 Cache Management Settings

**Goal**: User control over cache size and cleanup.

**Location**: `config.yaml`

**Configuration**:
```yaml
media:
  cache_dir: ~/.cache/ithil/media
  cache_size_mb: 500  # Max cache size
  auto_cleanup: true  # Auto-delete old files
  max_age_days: 30    # Delete files older than 30 days
```

**UI**: Add "Clear Cache" option in settings menu.

---

## Phase 5: Missing Media Types (LOW PRIORITY)

### 5.1 Stickers

**Goal**: Render stickers as images with emoji fallback.

**Current**: Stickers are likely treated as documents.

**Location**: `internal/telegram/messages.go` (enhance message conversion)

**Implementation**:
```go
// In convertMessageContent():
case *tg.MessageMediaDocument:
    // Check for sticker attribute
    for _, attr := range doc.Attributes {
        if _, ok := attr.(*tg.DocumentAttributeSticker); ok {
            content.Type = types.MessageTypeSticker
            // If has emoji alt → use as fallback
            // Otherwise download and render as image
        }
    }
```

**Rendering**:
- **If image rendering available**: Show sticker as image
- **Fallback**: Show emoji alternative (if available)
- **Ultimate fallback**: "[Sticker]"

### 5.2 Animations (GIF/MPEG4)

**Goal**: Treat animations as videos or static images.

**Options**:
1. **Static approach**: Show first frame (like video thumbnails)
2. **External player**: Open in external viewer
3. **ASCII animation**: Too complex, skip

**Recommendation**: Use static first frame + option to open externally.

### 5.3 Polls

**Goal**: Display poll questions and results with progress bars.

**Location**: `internal/ui/components/poll_renderer.go` (new file)

**UI Design**:
```
📊 Which feature do you want next?

  ◉ Image rendering        45% ▰▰▰▰▰▰▰▰▰▱▱▱▱▱▱ (9 votes)
  ◯ Video playback         33% ▰▰▰▰▰▰▱▱▱▱▱▱▱▱▱ (7 votes)
  ◯ Secret chats           22% ▰▰▰▰▱▱▱▱▱▱▱▱▱▱▱ (4 votes)

  20 total votes • Poll closed
```

**Implementation**:
```go
type PollRenderer struct{}

func (r *PollRenderer) RenderPoll(poll *types.Poll) string {
    // 1. Show question
    // 2. For each option:
    //    - Show checkbox (◉ if voted, ◯ otherwise)
    //    - Show option text
    //    - Show percentage bar
    //    - Show vote count
    // 3. Show total votes and status (open/closed)
}
```

### 5.4 Locations

**Goal**: Display coordinates with optional ASCII map.

**Location**: `internal/ui/components/location_renderer.go` (new file)

**Basic Implementation**:
```go
func RenderLocation(loc *types.Location) string {
    return fmt.Sprintf(
        "📍 Location\n" +
        "   Latitude:  %.6f\n" +
        "   Longitude: %.6f\n" +
        "   🌐 View on map: https://maps.google.com/?q=%.6f,%.6f",
        loc.Latitude, loc.Longitude,
        loc.Latitude, loc.Longitude,
    )
}
```

**Enhanced (Optional)**:
- Fetch ASCII map from external API (e.g., `https://wttr.in/`)
- Show small map preview in terminal
- Low priority - basic coordinates are sufficient

---

## Implementation Priority Matrix

| Phase | Feature | Priority | Effort | Impact | Value Score |
|-------|---------|----------|--------|--------|-------------|
| 1.3 | Protocol Auto-Detection | HIGH | Medium | High | 9/10 |
| 1.1 | Kitty Protocol | HIGH | High | High | 9/10 |
| 1.2 | Sixel Protocol | HIGH | High | High | 9/10 |
| 1.5 | Thumbnails | HIGH | Medium | High | 8/10 |
| 4.1 | Download Progress | HIGH | Low | Medium | 7/10 |
| 2.1 | Background Playback | HIGH | Medium | Medium | 7/10 |
| 1.4 | Image Caching | MEDIUM | Medium | Medium | 6/10 |
| 3.1 | Video Thumbnails | MEDIUM | High | Medium | 6/10 |
| 3.2 | External Player | MEDIUM | Low | Medium | 6/10 |
| 4.2 | Manual Save | MEDIUM | Low | Low | 5/10 |
| 2.3 | Speed Control | MEDIUM | Low | Low | 5/10 |
| 5.1 | Stickers | LOW | Medium | Low | 4/10 |
| 5.2 | Animations | LOW | Medium | Low | 4/10 |
| 5.3 | Polls | LOW | Medium | Low | 4/10 |
| 5.4 | Locations | LOW | Low | Low | 3/10 |

---

## Suggested Implementation Order

### Sprint 1 (Week 1-2): Graphics Protocols
1. Protocol auto-detection system
2. Kitty graphics protocol
3. Sixel graphics protocol
4. Integration with existing media viewer

### Sprint 2 (Week 3): Image Enhancements
1. Thumbnail generation system
2. Image cache with LRU eviction
3. Download progress indicators
4. Lazy loading on scroll

### Sprint 3 (Week 4): Audio & Video
1. Background audio playback
2. Speed control for voice messages
3. Video thumbnail extraction
4. External player integration

### Sprint 4 (Week 5): Polish & Missing Types
1. Manual save feature
2. Cache management settings
3. Sticker support
4. Poll rendering

---

## Testing Strategy

### Unit Tests
- `media/*_test.go` - Test each renderer independently
- `cache/media_cache_test.go` - Test LRU eviction logic
- `telegram/media_test.go` - Test download progress tracking

### Integration Tests
- Test protocol detection on different terminals (Kitty, iTerm2, XTerm, Alacritty)
- Test image rendering with large files
- Test audio playback with all supported formats
- Test cache eviction with memory limits

### Manual Testing Checklist
```
□ Images render correctly in Kitty terminal
□ Images render correctly in Sixel-capable terminal
□ Images fallback to Unicode mosaic in basic terminal
□ Thumbnails appear in message list
□ Download progress shows correctly
□ Audio plays in background while switching chats
□ Speed control works for voice messages (0.5x - 2x)
□ Video thumbnails extract correctly
□ External player opens videos
□ Manual save works with custom paths
□ Cache eviction removes old files
□ Stickers render or show emoji fallback
□ Polls display with progress bars
□ Locations show coordinates
```

---

## Configuration Schema

### Updated `config.yaml`

```yaml
app:
  api_id: 12345
  api_hash: "your_api_hash"

media:
  # Graphics protocol (auto, kitty, sixel, unicode, ascii)
  graphics_protocol: "auto"

  # Cache settings
  cache_dir: "~/.cache/ithil/media"
  cache_size_mb: 500
  auto_cleanup: true
  max_age_days: 30

  # Download settings
  auto_download_images: true
  auto_download_videos: false
  max_auto_download_size_mb: 10

  # Image rendering
  image_max_width: 80
  image_max_height: 40
  image_colored: true

  # Audio settings
  audio_default_volume: 70  # 0-100%
  audio_background_playback: true

  # Video settings
  video_extract_thumbnails: true
  video_external_player: "default"  # or specify path

ui:
  # Existing UI settings...
```

---

## Dependencies to Add

### New Go Packages

```bash
# For Kitty/Sixel support
go get github.com/mattn/go-sixel

# For video thumbnail extraction (optional, requires CGo)
go get github.com/3d0c/gmf

# Already in use (verify versions):
# - github.com/disintegration/imaging (image manipulation)
# - github.com/faiface/beep (audio playback)
# - github.com/gotd/td (Telegram MTProto)
```

### System Dependencies (Optional)
- FFmpeg binary (for video thumbnails if not using CGo)
- libjpeg, libpng (usually already present)

---

## Comparison: ithil vs Competitors After Implementation

| Feature | nchat | tgt | **ithil (After)** |
|---------|-------|-----|-------------------|
| Image Rendering | ❌ Text only | ❌ None | ✅✅ Kitty/Sixel/Unicode/ASCII |
| Audio Playback | ❌ External | ❌ None | ✅✅ In-app with controls |
| Video Support | ❌ External | ❌ None | ✅ Thumbnail + external |
| Stickers | ❌ None | ❌ Emoji only | ✅ Image rendering |
| Background Audio | N/A | N/A | ✅ Yes |
| Download Progress | ✅ Yes | ❌ No | ✅ Yes |
| Protocol Detection | N/A | ❌ No | ✅ Auto-detect |
| Cache Management | ✅ Basic | ❌ TDLib only | ✅✅ Advanced LRU |

**Result**: ithil will be the **most feature-complete terminal Telegram client** with superior multimedia support.

---

## Key Architectural Decisions

### 1. **Why Custom Rendering Instead of Framework Solutions?**
- **Lesson from tgt**: Ratatui framework limitations blocked image support
- **ithil advantage**: Direct terminal control via crossterm/bubbletea
- **Decision**: Implement custom renderers for each protocol

### 2. **Why Hybrid Video Approach?**
- **Technical**: TUI video playback is extremely CPU-intensive and impractical
- **User Experience**: External players provide better controls and quality
- **Decision**: Show thumbnails in-app, delegate playback to system

### 3. **Why Beep Library for Audio?**
- **Pure Go**: No external dependencies (when CGo disabled)
- **Format Support**: MP3, OGG, WAV cover 99% of use cases
- **Performance**: Efficient, low resource usage
- **Decision**: Keep Beep, add optional CGo-enabled enhancements

### 4. **Why LRU Cache Instead of Unlimited?**
- **Memory Management**: Prevent unbounded memory growth
- **Performance**: Faster lookups with bounded cache
- **User Control**: Configurable limits
- **Decision**: Implement LRU with configurable size limits

---

## Next Steps

1. **Review and Prioritize**: Discuss priorities with team/community
2. **Start with Sprint 1**: Graphics protocol support is highest value
3. **Incremental Releases**: Ship features as they complete (don't wait for all)
4. **Community Feedback**: Get terminal protocol testing from users
5. **Documentation**: Update README and CLAUDE.md as features land

---

## Conclusion

**ithil's multimedia capabilities are already ahead of both nchat and tgt.** This plan completes the remaining gaps to create the **best-in-class terminal Telegram client** with:

- ✅ **Best image rendering**: Kitty, Sixel, Unicode mosaic, ASCII
- ✅ **Best audio support**: In-app playback with full controls
- ✅ **Practical video support**: Thumbnails + external player
- ✅ **Comprehensive file management**: Progress tracking, caching, manual save
- ✅ **Complete media type support**: All Telegram media types covered

**Estimated Timeline**: 4-5 weeks for all high-priority features, 2-3 additional weeks for polish and low-priority items.

**Competitive Advantage**: Pure Go implementation with gotd/td provides flexibility that nchat (C++) and tgt (Rust/Ratatui) cannot match.
