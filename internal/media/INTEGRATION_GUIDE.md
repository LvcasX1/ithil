# Integration Guide: Adding Thumbnails to ithil

This guide walks through integrating the thumbnail generator into ithil's conversation view.

## Step 1: Update Types

Add thumbnail field to message structure in `pkg/types/types.go`:

```go
type Message struct {
    ID        int64
    // ... existing fields ...
    Content   MessageContent

    // Add this field:
    Thumbnail string // Rendered thumbnail for display (empty if not yet generated)
}
```

## Step 2: Update ConversationModel

Add thumbnail generator to `internal/ui/models/conversation.go`:

```go
import (
    // ... existing imports ...
    "github.com/lvcasx1/ithil/internal/media"
)

type ConversationModel struct {
    // ... existing fields ...

    // Add thumbnail support
    thumbnailGen *media.ThumbnailGenerator
}
```

## Step 3: Initialize in Constructor

Update `NewConversationModel()`:

```go
func NewConversationModel(client *telegram.Client, cache *cache.Cache, keyMap keys.KeyMap) ConversationModel {
    // ... existing initialization ...

    // Initialize thumbnail generator
    detector := media.NewProtocolDetector()
    protocol := detector.DetectProtocol()

    return ConversationModel{
        // ... existing fields ...
        thumbnailGen: media.NewThumbnailGenerator(20, 10, protocol),
    }
}
```

## Step 4: Define Thumbnail Message Types

Add to `internal/ui/models/conversation.go`:

```go
// ThumbnailReadyMsg is sent when a thumbnail finishes generating
type ThumbnailReadyMsg struct {
    MessageID int64
    Thumbnail string
}

// ThumbnailErrorMsg is sent when thumbnail generation fails
type ThumbnailErrorMsg struct {
    MessageID int64
    Error     error
}
```

## Step 5: Generate Thumbnails on Message Receipt

Update the message handling in `Update()`:

```go
func (m ConversationModel) Update(msg tea.Msg) (ConversationModel, tea.Cmd) {
    switch msg := msg.(type) {

    // ... existing cases ...

    case types.UpdateNewMessage:
        // Add message to cache/view
        m.messages = append(m.messages, msg.Message)

        // If message has media, generate thumbnail asynchronously
        if msg.Message.Content.Type == types.MessageTypePhoto {
            if msg.Message.Content.Media.LocalPath != "" {
                // Image already downloaded, generate thumbnail
                return m, m.generateThumbnailCmd(msg.Message.ID, msg.Message.Content.Media.LocalPath)
            } else {
                // Image not downloaded yet, will generate after download completes
                // (See Step 7 for download integration)
            }
        }

        return m, nil

    case ThumbnailReadyMsg:
        // Update message with thumbnail
        for i := range m.messages {
            if m.messages[i].ID == msg.MessageID {
                m.messages[i].Thumbnail = msg.Thumbnail
                break
            }
        }
        return m, nil

    case ThumbnailErrorMsg:
        // Handle thumbnail generation error
        // Could set a flag to show error icon or skip thumbnail
        log.Printf("Thumbnail generation failed for message %d: %v", msg.MessageID, msg.Error)
        return m, nil
    }

    // ... rest of Update ...
}
```

## Step 6: Create Thumbnail Generation Command

Add helper method to `ConversationModel`:

```go
// generateThumbnailCmd creates a command that generates a thumbnail asynchronously
func (m ConversationModel) generateThumbnailCmd(messageID int64, imagePath string) tea.Cmd {
    return func() tea.Msg {
        // Channel to receive result
        result := make(chan tea.Msg, 1)

        // Generate asynchronously
        m.thumbnailGen.GenerateThumbnailAsync(imagePath, func(thumbnail string, err error) {
            if err != nil {
                result <- ThumbnailErrorMsg{
                    MessageID: messageID,
                    Error:     err,
                }
            } else {
                result <- ThumbnailReadyMsg{
                    MessageID: messageID,
                    Thumbnail: thumbnail,
                }
            }
        })

        // Wait for result
        return <-result
    }
}
```

## Step 7: Integrate with Media Download

When media finishes downloading, generate thumbnail. Update media download handler:

```go
case MediaDownloadCompleteMsg:
    // Update message with local path
    for i := range m.messages {
        if m.messages[i].ID == msg.MessageID {
            m.messages[i].Content.Media.LocalPath = msg.LocalPath

            // Now generate thumbnail
            if m.messages[i].Content.Type == types.MessageTypePhoto {
                return m, m.generateThumbnailCmd(msg.MessageID, msg.LocalPath)
            }
            break
        }
    }
    return m, nil
```

## Step 8: Update Message Rendering

Update `renderMessage()` or message component to display thumbnails:

```go
func (m ConversationModel) renderMessage(msg types.Message, selected bool) string {
    var content string

    switch msg.Content.Type {
    case types.MessageTypePhoto:
        if msg.Thumbnail != "" {
            // Thumbnail is ready - display it
            content = msg.Thumbnail

            // Add caption if present
            if msg.Content.Caption != "" {
                content += "\n" + msg.Content.Caption
            }
        } else if msg.Content.Media.LocalPath != "" {
            // Downloaded but thumbnail still generating
            content = m.styles.MediaPlaceholder.Render("[Generating thumbnail...]")
        } else {
            // Not downloaded yet
            content = m.styles.MediaPlaceholder.Render("[Photo - click to download]")
        }

    case types.MessageTypeText:
        content = msg.Content.Text

    // ... other message types ...
    }

    // Apply styling and return
    return m.styles.Message.Render(content)
}
```

## Step 9: Preload Thumbnails for Chat History

When loading a chat, preload thumbnails for better scrolling performance:

```go
func (m *ConversationModel) loadChatHistory(chat *types.Chat) tea.Cmd {
    // ... existing history loading ...

    // Collect image paths for preloading
    var imagePaths []string
    for _, msg := range m.messages {
        if msg.Content.Type == types.MessageTypePhoto && msg.Content.Media.LocalPath != "" {
            imagePaths = append(imagePaths, msg.Content.Media.LocalPath)
        }
    }

    // Preload thumbnails in background
    if len(imagePaths) > 0 {
        go m.thumbnailGen.PreloadThumbnails(imagePaths, func(path, thumbnail string, err error) {
            if err == nil {
                // Find message and update (could send a message to UI)
                // For simplicity, thumbnails will be cached and instant on next render
            }
        })
    }

    return nil
}
```

## Step 10: Add Settings Support (Optional)

Allow users to configure thumbnail settings in `internal/ui/models/settings.go`:

```go
type Settings struct {
    // ... existing settings ...

    // Thumbnail settings
    ThumbnailsEnabled bool
    ThumbnailWidth    int
    ThumbnailHeight   int
    ThumbnailProtocol media.GraphicsProtocol
}

// Apply thumbnail settings
func (m *ConversationModel) applyThumbnailSettings(settings Settings) {
    if !settings.ThumbnailsEnabled {
        // Disable thumbnails
        m.thumbnailGen = nil
        return
    }

    // Update dimensions
    m.thumbnailGen.SetDimensions(settings.ThumbnailWidth, settings.ThumbnailHeight)

    // Update protocol
    m.thumbnailGen.SetProtocol(settings.ThumbnailProtocol)
}
```

## Step 11: Add Keyboard Shortcuts (Optional)

Add keybinding to toggle thumbnails in `internal/ui/keys/keymap.go`:

```go
type KeyMap struct {
    // ... existing bindings ...

    ToggleThumbnails key.Binding
}

// In initialization:
ToggleThumbnails: key.NewBinding(
    key.WithKeys("t"),
    key.WithHelp("t", "toggle thumbnails"),
),
```

Handle in conversation model:

```go
case key.Matches(msg, m.keyMap.ToggleThumbnails):
    if m.thumbnailGen != nil {
        m.thumbnailGen = nil // Disable
    } else {
        // Re-enable
        detector := media.NewProtocolDetector()
        protocol := detector.DetectProtocol()
        m.thumbnailGen = media.NewThumbnailGenerator(20, 10, protocol)
    }
    return m, nil
```

## Step 12: Memory Management

Clear thumbnail cache when switching chats to save memory:

```go
func (m *ConversationModel) switchChat(newChat *types.Chat) tea.Cmd {
    // ... existing chat switching logic ...

    // Clear thumbnail cache to free memory
    if m.thumbnailGen != nil {
        m.thumbnailGen.ClearCache()
    }

    // ... rest of switching logic ...
}
```

## Testing the Integration

### Manual Testing

1. Start ithil: `go run cmd/ithil/main.go`
2. Open a chat with images
3. Images should show thumbnails when downloaded
4. Scroll through history - thumbnails should appear quickly (cached)
5. Switch to another chat - thumbnails should regenerate
6. Try different message types (photos, videos, etc.)

### Test Checklist

- [ ] Thumbnails appear for downloaded images
- [ ] Loading indicator shows while generating
- [ ] Thumbnails are cached (instant on re-render)
- [ ] Cache clears when switching chats
- [ ] No memory leaks with many images
- [ ] UI stays responsive during generation
- [ ] Errors are handled gracefully
- [ ] Works with all graphics protocols
- [ ] Thumbnails maintain aspect ratio
- [ ] Captions display correctly with thumbnails

## Performance Tips

1. **Limit visible preloading**: Only preload thumbnails for visible messages
   ```go
   visibleMessages := m.getVisibleMessages(viewportHeight)
   // Only preload these
   ```

2. **Lazy generation**: Only generate when message scrolls into view
   ```go
   if !m.isVisible(msg) {
       return // Skip thumbnail generation
   }
   ```

3. **Adjust cache size**: Increase for better performance, decrease for lower memory
   ```go
   opts := &media.ThumbnailGeneratorOptions{
       CacheSize: 200, // Adjust based on typical usage
   }
   ```

4. **Protocol selection**: Let users choose protocol for performance/quality trade-off
   - Kitty: Best quality, fastest
   - Sixel: Good quality, fast
   - Unicode Mosaic: Good for true-color terminals
   - ASCII: Most compatible, slowest

## Troubleshooting

### Thumbnails not appearing
- Check that `msg.Content.Media.LocalPath` is set (file downloaded)
- Verify `ThumbnailReadyMsg` is being sent and received
- Check for errors in thumbnail generation

### Slow performance
- Reduce thumbnail size: `SetDimensions(15, 8)`
- Increase cache size to reduce regeneration
- Use async generation (don't block UI)

### High memory usage
- Reduce cache size: `opts.CacheSize = 50`
- Clear cache more frequently
- Only preload visible messages

### Wrong colors/garbled output
- Protocol mismatch with terminal
- Use auto-detection: `opts.AutoDetect = true`
- Or manually select appropriate protocol

## Example: Complete Message Component

```go
// In internal/ui/components/message.go

type MessageComponent struct {
    styles       *styles.Styles
    thumbnailGen *media.ThumbnailGenerator
}

func (mc *MessageComponent) Render(msg types.Message, selected bool) string {
    var parts []string

    // Sender name
    parts = append(parts, mc.renderSender(msg))

    // Message content
    switch msg.Content.Type {
    case types.MessageTypePhoto:
        parts = append(parts, mc.renderPhoto(msg))
    case types.MessageTypeText:
        parts = append(parts, msg.Content.Text)
    // ... other types ...
    }

    // Timestamp
    parts = append(parts, mc.renderTimestamp(msg))

    return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (mc *MessageComponent) renderPhoto(msg types.Message) string {
    if msg.Thumbnail != "" {
        // Thumbnail ready
        result := msg.Thumbnail
        if msg.Content.Caption != "" {
            result += "\n" + mc.styles.Caption.Render(msg.Content.Caption)
        }
        return result
    } else if msg.Content.Media.LocalPath != "" {
        // Downloading or generating
        return mc.styles.Placeholder.Render("⏳ Generating thumbnail...")
    } else {
        // Not downloaded
        size := formatFileSize(msg.Content.Media.Size)
        return mc.styles.Placeholder.Render(fmt.Sprintf("📷 Photo (%s) - Enter to download", size))
    }
}
```

## Summary

This integration guide provides a complete path to adding thumbnail support to ithil. The key points:

1. Add thumbnail field to message type
2. Initialize thumbnail generator in conversation model
3. Generate thumbnails asynchronously when media arrives
4. Update UI when thumbnails are ready
5. Preload for better performance
6. Clear cache when switching chats

The system is designed to be non-blocking, efficient, and user-friendly. All thumbnail operations happen in the background, keeping the UI responsive at all times.
