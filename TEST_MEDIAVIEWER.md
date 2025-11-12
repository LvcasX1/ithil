# Media Viewer Fix - Testing Guide

## Quick Testing Steps

### 1. Build the Fixed Version
```bash
go build -o bin/ithil ./cmd/ithil
```

### 2. Run Ithil
```bash
./bin/ithil
```

### 3. Test Image Viewing
1. Navigate to a chat with image messages
2. Use `j`/`k` to select an image message
3. Press `Enter` to open the image in the media viewer

### 4. Verify the Fix
**BEFORE FIX (buggy behavior):**
- ❌ Entire TUI shifts to top ~40% of terminal
- ❌ ~60% black space at bottom
- ❌ Modal appears but UI is broken

**AFTER FIX (correct behavior):**
- ✅ TUI remains in normal position
- ✅ Modal opens centered
- ✅ Image renders using Unicode Mosaic (colored blocks)
- ✅ No black space or UI shift
- ✅ Can close with ESC/Q and return to normal view

### 5. Visual Inspection
The image should appear as:
```
┌─────────────────────────────────────┐
│                                     │
│        📷 Image Viewer              │
│                                     │
│  ▀▀▄▄██████▀▀  (colored blocks)    │
│  ██▀▀▄▄▀▀██                         │
│  ▀▀██▄▄▀▀▄▄    (true color)        │
│                                     │
│    Press ESC or Q to close          │
└─────────────────────────────────────┘
```

**Note**: The actual image will show colored Unicode half-blocks (▀ ▄ █) with true color (24-bit RGB).

## Debug Logging

If you need to verify what's happening internally:

### Check Protocol Detection
```bash
# Check your terminal's detected protocol
echo $TERM_PROGRAM
echo $TERM
echo $COLORTERM
```

### View Debug Logs
The application logs updates to `/tmp/ithil-updates.log`. You can monitor it:
```bash
tail -f /tmp/ithil-updates.log
```

## Common Scenarios

### Scenario 1: Kitty Terminal
- **Detected protocol**: Kitty
- **Rendering mode**: Unicode Mosaic (forced)
- **Expected**: No UI shift, colored block rendering

### Scenario 2: WezTerm
- **Detected protocol**: Sixel
- **Rendering mode**: Unicode Mosaic (forced)
- **Expected**: No UI shift, colored block rendering

### Scenario 3: Alacritty/iTerm2
- **Detected protocol**: Unicode Mosaic
- **Rendering mode**: Unicode Mosaic (native)
- **Expected**: No UI shift, colored block rendering

## Troubleshooting

### Issue: Image doesn't appear
**Possible causes:**
1. Image not downloaded yet - wait for "Downloading media..." to complete
2. Unsupported image format - check file type
3. Corrupted image file - error message should appear

**Solution**: Check the modal shows "Downloading media..." first, then the image

### Issue: Image appears blurry or low quality
**Expected behavior:**
- Unicode Mosaic provides ~2x vertical resolution
- Colors are true color (24-bit RGB)
- Some loss of detail compared to pixel protocols is normal for TUI

**Not a bug**: This is the trade-off for UI stability

### Issue: Modal doesn't close
**Solution**: Press ESC or Q (case sensitive)

## Performance Testing

### Large Images
Test with high-resolution images (4K, 8K):
- Should resize to fit terminal
- Rendering should complete in < 1 second
- No memory leaks

### Rapid Open/Close
Open and close the media viewer repeatedly:
- No UI flickering
- No position drift
- No memory leaks

## Regression Testing

Ensure other features still work:

- [ ] Chat list navigation
- [ ] Sending messages
- [ ] Receiving messages
- [ ] Audio/voice message playback
- [ ] Video/document info display
- [ ] Help modal (press `?`)
- [ ] Settings modal (press `Ctrl+,`)

## Success Criteria

The fix is verified if:
1. ✅ No UI shift when opening image viewer
2. ✅ Modal appears centered
3. ✅ Image renders correctly (colored blocks)
4. ✅ Can close modal and return to normal UI
5. ✅ No regression in other features
6. ✅ Works across different terminal emulators

## Report Issues

If you find any issues:
1. Note your terminal emulator (name and version)
2. Note your OS (Darwin, Linux, Windows)
3. Capture the bug behavior (screenshot if possible)
4. Check `/tmp/ithil-updates.log` for errors
5. Report with steps to reproduce
