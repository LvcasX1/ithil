# Download Progress Tracking

This document explains how to use the download progress tracking feature in ithil.

## Overview

The MediaManager now supports real-time progress tracking for media downloads. This allows UI components to display download status, percentage, speed, and estimated time remaining.

## Architecture

### Key Components

1. **DownloadProgress** (`pkg/types/types.go`): Contains progress information
   - `Status`: Current status (NotDownloaded, Downloading, Downloaded, Failed)
   - `BytesTotal`: Total file size in bytes
   - `BytesLoaded`: Bytes downloaded so far
   - `StartTime`: When the download started
   - `LastUpdate`: Last progress update time
   - Helper methods: `GetPercentage()`, `GetSpeed()`, `GetETA()`

2. **MediaManager** (`internal/telegram/media.go`): Manages downloads and progress
   - `SubscribeProgress(key)`: Subscribe to progress updates for a download
   - `UnsubscribeProgress(key)`: Unsubscribe and clean up
   - Internal `progressWriter`: Tracks bytes written during download

3. **Client Methods** (`internal/telegram/client.go`):
   - `DownloadMedia(message)`: Download without progress (backward compatible)
   - `DownloadMediaWithProgress(message, progressKey)`: Download with progress tracking

## Usage Examples

### Basic Usage (No Progress Tracking)

```go
// Existing code continues to work without changes
localPath, err := client.DownloadMedia(message)
if err != nil {
    log.Printf("Download failed: %v", err)
    return
}
log.Printf("Downloaded to: %s", localPath)
```

### Download with Progress Tracking

```go
import (
    "fmt"
    "time"
    "github.com/lvcasx1/ithil/pkg/types"
)

// Use message ID as the progress key (must be unique per download)
progressKey := fmt.Sprintf("msg_%d", message.ID)

// Subscribe to progress updates BEFORE starting the download
progressChan := mediaManager.SubscribeProgress(progressKey)

// Start a goroutine to receive progress updates
go func() {
    for progress := range progressChan {
        switch progress.Status {
        case types.DownloadStatusDownloading:
            percentage := progress.GetPercentage()
            speed := progress.GetSpeed()
            eta := progress.GetETA()

            fmt.Printf("Downloading: %.1f%% (%.2f KB/s, ETA: %v)\n",
                percentage,
                speed/1024,
                eta.Round(time.Second))

        case types.DownloadStatusDownloaded:
            fmt.Printf("Download complete! 100%%\n")

        case types.DownloadStatusFailed:
            fmt.Printf("Download failed: %v\n", progress.Error)
        }
    }
}()

// Start the download with progress tracking
localPath, err := client.DownloadMediaWithProgress(message, progressKey)

// Clean up when done
defer mediaManager.UnsubscribeProgress(progressKey)

if err != nil {
    log.Printf("Download failed: %v", err)
    return
}

log.Printf("Downloaded to: %s", localPath)
```

### Bubbletea Integration Example

```go
// In your Bubbletea model
type ConversationModel struct {
    // ... existing fields ...
    downloadingMessages map[int64]types.DownloadProgress
}

// Message type for progress updates
type ProgressUpdateMsg struct {
    MessageID int64
    Progress  types.DownloadProgress
}

// Command to subscribe to progress
func subscribeToProgress(client *telegram.Client, msg *types.Message) tea.Cmd {
    progressKey := fmt.Sprintf("msg_%d", msg.ID)
    progressChan := client.MediaManager().SubscribeProgress(progressKey)

    return func() tea.Msg {
        for progress := range progressChan {
            return ProgressUpdateMsg{
                MessageID: msg.ID,
                Progress:  progress,
            }
        }
        return nil
    }
}

// Command to download with progress
func downloadMediaCmd(client *telegram.Client, msg *types.Message) tea.Cmd {
    return func() tea.Msg {
        progressKey := fmt.Sprintf("msg_%d", msg.ID)
        _, err := client.DownloadMediaWithProgress(msg, progressKey)
        if err != nil {
            return ErrorMsg{err}
        }
        return DownloadCompleteMsg{MessageID: msg.ID}
    }
}

// Update method
func (m ConversationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ProgressUpdateMsg:
        // Update download progress in model
        if m.downloadingMessages == nil {
            m.downloadingMessages = make(map[int64]types.DownloadProgress)
        }
        m.downloadingMessages[msg.MessageID] = msg.Progress

        // Clean up if download complete or failed
        if msg.Progress.Status == types.DownloadStatusDownloaded ||
           msg.Progress.Status == types.DownloadStatusFailed {
            delete(m.downloadingMessages, msg.MessageID)
        }

        return m, nil

    // ... other cases ...
    }
    return m, nil
}

// View method - render progress
func (m ConversationModel) renderDownloadProgress(msgID int64) string {
    progress, exists := m.downloadingMessages[msgID]
    if !exists {
        return "[Download]"
    }

    switch progress.Status {
    case types.DownloadStatusDownloading:
        pct := progress.GetPercentage()
        speed := progress.GetSpeed() / 1024 // KB/s
        return fmt.Sprintf("[Downloading: %.0f%% (%.1f KB/s)]", pct, speed)

    case types.DownloadStatusFailed:
        return "[Download Failed]"

    default:
        return "[Download]"
    }
}
```

## Progress Update Frequency

Progress updates are sent based on these thresholds (whichever comes first):
- **100 KB** downloaded since last update
- **100 ms** elapsed since last update

This balances between responsiveness and performance.

## Progress Key Guidelines

The progress key should be:
1. **Unique** per download
2. **Consistent** between subscribe and download calls
3. **Easy to generate** (e.g., message ID, file ID)

Examples of good progress keys:
```go
// For message downloads
progressKey := fmt.Sprintf("msg_%d", message.ID)

// For direct media downloads
progressKey := fmt.Sprintf("photo_%d", photo.ID)
progressKey := fmt.Sprintf("doc_%d", doc.ID)

// For batch downloads
progressKey := fmt.Sprintf("batch_%s_%d", batchID, index)
```

## Thread Safety

All progress tracking components are thread-safe:
- Progress channels use buffered channels (size: 100)
- Channel map protected by `sync.RWMutex`
- Non-blocking sends prevent deadlocks
- Safe for concurrent downloads

## Memory Management

### Channel Cleanup

Always call `UnsubscribeProgress()` when done to prevent memory leaks:

```go
progressKey := fmt.Sprintf("msg_%d", msg.ID)
progressChan := mediaManager.SubscribeProgress(progressKey)

// Ensure cleanup happens
defer mediaManager.UnsubscribeProgress(progressKey)

// ... download code ...
```

### Automatic Cleanup

The progressWriter sends a final update (success or failure) before finishing,
allowing consumers to detect completion and clean up.

## Error Handling

### Download Errors

When a download fails, the final progress update will have:
- `Status`: `DownloadStatusFailed`
- `Error`: The actual error that occurred

Example:
```go
for progress := range progressChan {
    if progress.Status == types.DownloadStatusFailed {
        log.Printf("Download failed: %v", progress.Error)
        // Handle error appropriately
    }
}
```

### Network Interruptions

If the network connection is interrupted:
1. The download fails with an error
2. A failed status is sent with the error
3. Partial files are cleaned up automatically
4. The caller can retry the download

### Concurrent Downloads

Multiple downloads can run concurrently. Each needs a unique progress key:

```go
for i, msg := range messages {
    progressKey := fmt.Sprintf("batch_%d", i)
    go func(m *types.Message, key string) {
        progressChan := mediaManager.SubscribeProgress(key)
        defer mediaManager.UnsubscribeProgress(key)

        // Handle progress in goroutine
        go func() {
            for progress := range progressChan {
                // Update UI for this download
            }
        }()

        // Start download
        _, err := client.DownloadMediaWithProgress(m, key)
        if err != nil {
            log.Printf("Download %s failed: %v", key, err)
        }
    }(msg, progressKey)
}
```

## Helper Methods

### GetPercentage()
```go
percentage := progress.GetPercentage() // 0.0 to 100.0
```
Returns the download completion percentage. Safe to call at any time.

### GetSpeed()
```go
speed := progress.GetSpeed() // bytes per second
speedKB := speed / 1024      // KB/s
speedMB := speed / (1024 * 1024) // MB/s
```
Returns average download speed in bytes per second since the download started.

### GetETA()
```go
eta := progress.GetETA() // time.Duration
etaSeconds := eta.Round(time.Second)
fmt.Printf("ETA: %v\n", etaSeconds)
```
Returns estimated time remaining based on current average speed.

## Performance Considerations

### Minimal Overhead

Progress tracking adds minimal overhead:
- Only active for downloads with a progress key
- Non-blocking channel sends (skips update if channel full)
- Throttled updates (100KB or 100ms threshold)

### No Progress Tracking

For batch operations where progress isn't needed, use `DownloadMedia()`:
```go
// No progress tracking overhead
localPath, err := client.DownloadMedia(message)
```

## Testing

### Manual Testing
```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/lvcasx1/ithil/internal/telegram"
    "github.com/lvcasx1/ithil/pkg/types"
)

func testDownloadProgress(client *telegram.Client, message *types.Message) {
    progressKey := fmt.Sprintf("test_%d", message.ID)

    // Subscribe to progress
    progressChan := client.MediaManager().SubscribeProgress(progressKey)
    defer client.MediaManager().UnsubscribeProgress(progressKey)

    // Monitor progress
    done := make(chan bool)
    go func() {
        for progress := range progressChan {
            fmt.Printf("\rProgress: %.1f%% (%d/%d bytes, %.2f KB/s)",
                progress.GetPercentage(),
                progress.BytesLoaded,
                progress.BytesTotal,
                progress.GetSpeed()/1024)

            if progress.Status == types.DownloadStatusDownloaded {
                fmt.Println("\n✓ Download complete!")
                done <- true
                return
            } else if progress.Status == types.DownloadStatusFailed {
                fmt.Printf("\n✗ Download failed: %v\n", progress.Error)
                done <- true
                return
            }
        }
    }()

    // Start download
    localPath, err := client.DownloadMediaWithProgress(message, progressKey)
    if err != nil {
        log.Printf("Download failed: %v", err)
        return
    }

    // Wait for completion
    <-done
    log.Printf("File saved to: %s", localPath)
}
```

## Future Enhancements

Potential future improvements:
1. Resumable downloads (save partial state)
2. Download queuing system
3. Bandwidth throttling
4. Multiple parallel chunks for large files
5. Download history/statistics
6. Configurable progress update thresholds
