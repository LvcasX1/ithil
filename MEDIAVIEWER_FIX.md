# Media Viewer UI Shift Bug - Root Cause Analysis & Fix

## Problem Statement
When opening the media viewer modal to view images, the entire TUI would shift upward to the top ~40% of the terminal, leaving ~60% black space at the bottom.

## Root Cause Analysis

### Technical Investigation
The bug was caused by a **fundamental incompatibility between pixel-based graphics protocols (Sixel/Kitty) and Bubbletea's Lipgloss layout system**.

#### How Pixel Graphics Work in Terminals
1. **Sixel/Kitty protocols** render images by emitting escape sequences that tell the terminal to draw pixels
2. These escape sequences **physically move the terminal cursor DOWN** as they render the pixel data
3. The cursor movement is a **side effect** of the graphics protocol, not controlled by the application

#### How Lipgloss Works
1. **Lipgloss `Place()` function** positions content by calculating text dimensions
2. It assumes all content is **static text** that doesn't move the cursor
3. It uses ANSI escape codes for positioning, which work perfectly for text

#### The Conflict
When Sixel/Kitty content is rendered inside a Lipgloss modal:
1. Lipgloss calculates position: "Center this modal at row X, column Y"
2. Lipgloss moves cursor to position and starts rendering
3. **Sixel escape sequences execute**, drawing pixels AND moving cursor down
4. The cursor is now MUCH lower than Lipgloss expected
5. Remaining content renders from the new cursor position
6. **Result**: The entire UI appears shifted upward because the cursor moved down unexpectedly

### Previous Failed Attempts
1. **Removing newlines from Kitty renderer** - Didn't address the core issue (cursor movement from escape sequences)
2. **Making MaxHeight conditional** - Didn't prevent the cursor movement, just changed layout constraints

## The Solution

### Approach
**Force text-based rendering (Unicode Mosaic or ASCII) for ALL images displayed in modals**, regardless of terminal capabilities.

### Why This Works
1. **Unicode Mosaic rendering** uses colored Unicode characters (▀ ▄ █) with ANSI color codes
2. **No cursor movement side effects** - it's just text with color
3. **Fully compatible with Lipgloss** layout system
4. **Still provides good quality** - True color (24-bit) rendering with 2x vertical resolution

### Implementation Changes

#### File: `internal/ui/components/mediaviewer.go`

**Function: `renderImageWithBestProtocol()`**
```go
// BEFORE (buggy):
switch m.detectedProtocol {
case media.ProtocolKitty:
    return m.kittyRenderer.RenderImageFile(filePath)  // ❌ Causes UI shift
case media.ProtocolSixel:
    return m.sixelRenderer.RenderImageFile(filePath)  // ❌ Causes UI shift
case media.ProtocolUnicodeMosaic:
    return m.mosaicRenderer.RenderImageFile(filePath)
// ...
}

// AFTER (fixed):
switch m.detectedProtocol {
case media.ProtocolKitty, media.ProtocolSixel:
    // FORCE Unicode Mosaic to prevent cursor movement
    return m.mosaicRenderer.RenderImageFile(filePath)  // ✅ Text-based, no shift
case media.ProtocolUnicodeMosaic:
    return m.mosaicRenderer.RenderImageFile(filePath)
// ...
}
```

**Function: `View()`**
```go
// BEFORE (buggy):
usePixelProtocol := m.detectedProtocol == media.ProtocolKitty || m.detectedProtocol == media.ProtocolSixel
modalStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color(styles.AccentCyan)).
    Padding(1, 2).
    Width(m.width)
if !usePixelProtocol {
    modalStyle = modalStyle.MaxHeight(m.height)  // Conditional height
}

// AFTER (fixed):
modalStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color(styles.AccentCyan)).
    Padding(1, 2).
    Width(m.width).
    MaxHeight(m.height)  // ✅ Always safe with text-based rendering
```

## Verification & Testing

### Build Verification
```bash
go build -o bin/ithil ./cmd/ithil
```
Expected: Clean build with no errors ✅

### Manual Testing Checklist
1. **Basic Image Display**
   - [ ] Open a chat with image messages
   - [ ] Press Enter on an image
   - [ ] Verify modal opens WITHOUT UI shift
   - [ ] Verify image renders using Unicode Mosaic (colored blocks)

2. **Terminal Resize**
   - [ ] With image modal open, resize terminal window
   - [ ] Verify layout recalculates correctly
   - [ ] Verify no UI shift occurs

3. **Different Image Formats**
   - [ ] Test with JPEG images
   - [ ] Test with PNG images
   - [ ] Test with GIF images (static)

4. **Different Terminal Emulators**
   - [ ] Test in Kitty terminal (protocol detection: Kitty → renders as Mosaic)
   - [ ] Test in WezTerm (protocol detection: Sixel → renders as Mosaic)
   - [ ] Test in Alacritty (protocol detection: Mosaic → renders as Mosaic)
   - [ ] Test in iTerm2 (protocol detection: Mosaic → renders as Mosaic)

5. **Edge Cases**
   - [ ] Very large images (should resize to fit)
   - [ ] Very small images (should display correctly)
   - [ ] Corrupted/invalid images (should show error message)

### Expected Behavior
- ✅ Media viewer modal opens centered in the terminal
- ✅ Main TUI remains in its normal position (not shifted)
- ✅ Image displays correctly using Unicode Mosaic rendering
- ✅ Colors appear vibrant (true color support)
- ✅ Modal can be closed with ESC/Q
- ✅ UI remains stable after closing modal

## Trade-offs & Future Considerations

### Trade-offs
**Downside**: Users with Kitty/Sixel-capable terminals don't get pixel-perfect image rendering in modals
**Upside**: UI stability and consistent experience across all terminals

### Quality Comparison
- **Pixel protocols (Sixel/Kitty)**: ~16-24 pixels per character cell, smooth gradients
- **Unicode Mosaic**: 2x vertical resolution, 24-bit true color, block-based rendering
- **Quality difference**: Noticeable but acceptable for a TUI application

### Alternative Solutions (Not Implemented)
1. **Full-screen image viewer outside Lipgloss** - Would require complete rewrite of rendering
2. **Save cursor position before Sixel** - Doesn't work reliably across terminals
3. **Use external image viewer** - Poor UX, requires additional dependencies

### Future Enhancements
If pixel-perfect rendering is desired in the future:
1. Implement a **full-screen image viewer mode** that bypasses Lipgloss entirely
2. Use **raw terminal output** with manual cursor control
3. Detect terminal dimensions in pixels (not characters) for precise placement
4. Handle terminal resize events manually

## References
- Lipgloss documentation: https://github.com/charmbracelet/lipgloss
- Sixel graphics protocol: https://en.wikipedia.org/wiki/Sixel
- Kitty graphics protocol: https://sw.kovidgoyal.net/kitty/graphics-protocol/
- Unicode half-block rendering: https://en.wikipedia.org/wiki/Block_Elements

## Conclusion
The fix ensures **UI stability** by using text-based image rendering (Unicode Mosaic) for all images displayed in modals. This eliminates cursor movement side effects from pixel graphics protocols while maintaining good visual quality through true color Unicode characters.
