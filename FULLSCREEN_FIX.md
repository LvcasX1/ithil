# Fullscreen Image Viewer - Final Fix

## Problem Summary

The initial fullscreen implementation failed because Bubbletea renders ALL sub-models before outputting the final view. This meant:
1. MainModel rendered everything (chat list, conversation, sidebar)
2. ConversationModel wrapped media viewer with `lipgloss.Place()`
3. The screen clear escape sequence (`\x1b[2J`) was buried deep in the output
4. Result: Normal TUI remained visible at top, fullscreen content appeared below

## Root Cause

**Bubbletea's rendering flow:**
- MainModel.View() → calls all sub-model View() methods
- All views are combined into a single output string
- Lipgloss applies layout/styling throughout the tree
- Final string is sent to terminal

**Why raw escape sequences didn't work:**
- The `\x1b[2J` (clear screen) only works if it's at the VERY START of the output
- When wrapped in Lipgloss styling and rendered after other content, it has no effect
- The cursor was already positioned after rendering the normal TUI

## Solution: Short-Circuit Rendering

The fix detects fullscreen mode at the ROOT LEVEL (MainModel) and bypasses all normal rendering.

### Implementation Changes

#### 1. MediaViewerComponent - Added Fullscreen Detection
**File:** `internal/ui/components/mediaviewer.go`

```go
// IsFullscreenMode returns true if the media viewer should use fullscreen mode.
func (m *MediaViewerComponent) IsFullscreenMode() bool {
    if !m.visible {
        return false
    }

    // Fullscreen mode only for pixel protocols with photos
    return (m.detectedProtocol == media.ProtocolKitty ||
            m.detectedProtocol == media.ProtocolSixel) &&
           m.message != nil &&
           m.message.Content.Type == types.MessageTypePhoto &&
           !m.downloading &&
           m.renderError == nil
}
```

**Logic:** Fullscreen mode is ONLY used for:
- Kitty or Sixel protocols (pixel-based, high quality)
- Photo messages (not audio/video/docs)
- Successfully rendered content (no download/error states)

#### 2. ConversationModel - Added Helper Methods
**File:** `internal/ui/models/conversation.go`

```go
// IsMediaViewerFullscreen returns true if the media viewer is visible and in fullscreen mode.
func (m *ConversationModel) IsMediaViewerFullscreen() bool {
    return m.mediaViewer.IsVisible() && m.mediaViewer.IsFullscreenMode()
}

// GetMediaViewerFullscreenView returns the raw fullscreen view from the media viewer.
func (m *ConversationModel) GetMediaViewerFullscreenView() string {
    return m.mediaViewer.ViewFullscreen()
}
```

Also updated `View()` to avoid wrapping fullscreen content:
```go
if m.mediaViewer.IsVisible() {
    // For fullscreen mode, return raw view (though MainModel should handle this)
    if m.mediaViewer.IsFullscreenMode() {
        return m.mediaViewer.ViewFullscreen()
    }

    // For modal mode, center it with Lipgloss
    viewerView := m.mediaViewer.View()
    overlay := lipgloss.Place(...)
}
```

#### 3. MainModel - Short-Circuit Normal Rendering
**File:** `internal/ui/models/main.go`

```go
func (m *MainModel) View() string {
    // ... auth and settings checks ...

    // If media viewer is in fullscreen mode, short-circuit normal rendering
    // and return ONLY the fullscreen view (bypassing all Lipgloss layout)
    if m.conversation.IsMediaViewerFullscreen() {
        return m.conversation.GetMediaViewerFullscreenView()
    }

    // Show main UI with three panes
    return m.renderMainUI()
}
```

**Critical change:** When fullscreen mode is active:
- Skip rendering chat list, conversation pane, sidebar, status bar
- Return ONLY the fullscreen image view
- The `\x1b[2J` is now at the START of the output string
- Terminal properly interprets the clear screen command

### How It Works Now

**Normal flow (modal mode - audio/video/docs/low-res images):**
1. MainModel.View() → renderMainUI()
2. Renders all three panes (chat list, conversation, sidebar)
3. Conversation.View() wraps media viewer with lipgloss.Place()
4. Modal appears centered over the TUI
5. ESC closes modal, returns to normal view

**Fullscreen flow (Kitty/Sixel photos):**
1. MainModel.View() detects fullscreen mode
2. Short-circuits normal rendering
3. Returns ONLY: `\x1b[2J\x1b[H` + image data + text overlay
4. Screen clears completely
5. High-quality image renders using full terminal dimensions
6. ESC triggers re-render → fullscreen mode false → normal UI returns

### Visual Comparison

**Before (broken):**
```
┌─────────────────────────────────┐
│ Chat List | Conversation | Side │  ← Normal TUI rendered
├─────────────────────────────────┤
│                                 │
│  Press ESC to return            │  ← Fullscreen content below
│                                 │
│  (mostly black space)           │
└─────────────────────────────────┘
```

**After (fixed):**
```
┌─────────────────────────────────┐
│                                 │
│                                 │
│    [HIGH QUALITY IMAGE]         │  ← Full screen, no TUI
│                                 │
│                                 │
│  Press ESC or Q to return       │
└─────────────────────────────────┘
```

### Expected Behavior

**For Kitty/Sixel terminals viewing photos:**
1. User presses Enter on an image message
2. Screen clears completely (no TUI visible)
3. High-quality pixel-based image renders using entire screen
4. Simple text overlay at bottom: "Press ESC or Q to return to chat"
5. Press ESC → screen clears → normal TUI reappears

**For all other scenarios (UNCHANGED):**
- Audio messages → Lipgloss modal with playback controls
- Video messages → Lipgloss modal with metadata
- Documents → Lipgloss modal with file info
- Unicode Mosaic/ASCII images → Lipgloss modal
- Terminals without Kitty/Sixel → Lipgloss modal with Mosaic rendering

## Benefits

1. **High Quality Images**: Kitty/Sixel protocols render at full quality with entire terminal
2. **Stable UI**: No cursor movement issues (fullscreen bypasses Lipgloss completely)
3. **Clean Transitions**: Clear screen ensures smooth switching between modes
4. **Preserved Functionality**: Audio, video, docs unchanged
5. **Correct Architecture**: Follows Bubbletea best practices for fullscreen overlays

## Testing

Build and test:
```bash
go build -o bin/ithil ./cmd/ithil
./bin/ithil
```

Test cases:
1. ✅ Open image in Kitty/Sixel terminal → fullscreen high quality
2. ✅ Open image in other terminal → modal with Mosaic rendering
3. ✅ Open audio message → modal with playback controls
4. ✅ Open video message → modal with metadata
5. ✅ Press ESC from fullscreen → returns to normal TUI cleanly

## Why This Approach Works

The key insight is that **Bubbletea's View() method is called from the root down**:
- If MainModel.View() returns early, no sub-models are rendered
- The returned string is sent directly to the terminal
- Escape sequences at the start of the string execute properly
- This is the same pattern used by other fullscreen tools (vim, less, etc.)

By detecting fullscreen mode at the root and short-circuiting the render tree, we ensure:
- No Lipgloss interference
- No cursor positioning side effects
- Clean, simple terminal output
- Proper screen clearing

## References
- Bubbletea documentation: https://github.com/charmbracelet/bubbletea
- ANSI escape codes: https://en.wikipedia.org/wiki/ANSI_escape_code
- Previous fix attempt: MEDIAVIEWER_FIX.md (documented the Lipgloss cursor issue)
