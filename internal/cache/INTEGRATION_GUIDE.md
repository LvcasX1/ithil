# Media Cache Integration Guide

This guide shows how to integrate the MediaCache into Ithil's existing media download system.

## Step 1: Add Cache to MediaManager

Modify `/Users/lvcasx1/Work/personal/ithil/internal/telegram/media.go`:

```go
type MediaManager struct {
    client     *Client
    downloader *downloader.Downloader
    uploader   *uploader.Uploader
    mediaDir   string
    cache      *cache.MediaCache // ADD THIS LINE
}
```

## Step 2: Initialize Cache in NewMediaManager

```go
func (c *Client) NewMediaManager(mediaDir string) (*MediaManager, error) {
    // Ensure media directory exists
    if err := os.MkdirAll(mediaDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create media directory: %w", err)
    }

    // Initialize cache with 500MB limit
    mediaCache, err := cache.NewMediaCache(500*1024*1024, mediaDir)
    if err != nil {
        c.logger.Warn("Failed to initialize media cache", "error", err)
        // Continue without cache
    }

    return &MediaManager{
        client:     c,
        downloader: downloader.NewDownloader(),
        uploader:   uploader.NewUploader(c.api),
        mediaDir:   mediaDir,
        cache:      mediaCache, // ADD THIS LINE
    }, nil
}
```

## Step 3: Update DownloadPhoto with Caching

```go
func (m *MediaManager) DownloadPhoto(ctx context.Context, photo *tg.Photo, chatID int64) (string, error) {
    if photo == nil {
        return "", fmt.Errorf("photo is nil")
    }

    // Generate cache key
    cacheKey := cache.GenerateKey(chatID, photo.ID, "photo")

    // Check cache first
    if m.cache != nil {
        if cachedPath, found := m.cache.Get(cacheKey); found {
            m.client.logger.Debug("Using cached photo",
                "photoID", photo.ID,
                "path", cachedPath)
            return cachedPath, nil
        }
    }

    m.client.logger.Debug("DownloadPhoto called",
        "photoID", photo.ID,
        "sizeCount", len(photo.Sizes))

    // Log all available sizes for debugging
    for _, size := range photo.Sizes {
        if photoSize, ok := size.(*tg.PhotoSize); ok {
            m.client.logger.Debug("Available photo size",
                "type", photoSize.Type,
                "priority", getPhotoSizePriority(photoSize.Type),
                "width", photoSize.W,
                "height", photoSize.H,
                "bytes", photoSize.Size)
        }
    }

    // Select the best photo size based on priority
    bestSize := selectBestPhotoSize(photo.Sizes)

    if bestSize == nil {
        m.client.logger.Error("No valid photo size found",
            "photoID", photo.ID,
            "totalSizes", len(photo.Sizes))
        return "", fmt.Errorf("no valid photo size found")
    }

    m.client.logger.Info("Selected best photo size for download",
        "type", bestSize.Type,
        "priority", getPhotoSizePriority(bestSize.Type),
        "width", bestSize.W,
        "height", bestSize.H,
        "bytes", bestSize.Size)

    // Create file location from photo
    location := &tg.InputPhotoFileLocation{
        ID:            photo.ID,
        AccessHash:    photo.AccessHash,
        FileReference: photo.FileReference,
        ThumbSize:     bestSize.Type,
    }

    // Generate local file path
    fileName := fmt.Sprintf("photo_%d_%d.jpg", chatID, photo.ID)
    localPath := filepath.Join(m.mediaDir, fmt.Sprintf("%d", chatID), fileName)

    // Ensure chat directory exists
    if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
        return "", fmt.Errorf("failed to create chat directory: %w", err)
    }

    // Download the file
    if err := m.downloadFile(ctx, location, localPath, int64(bestSize.Size)); err != nil {
        return "", fmt.Errorf("failed to download photo: %w", err)
    }

    fileInfo, err := os.Stat(localPath)
    if err == nil {
        m.client.logger.Info("Downloaded photo to file",
            "path", localPath,
            "actualFileSize", fileInfo.Size(),
            "expectedSize", bestSize.Size)

        // Add to cache
        if m.cache != nil {
            if err := m.cache.Put(cacheKey, localPath, fileInfo.Size()); err != nil {
                m.client.logger.Warn("Failed to cache photo", "error", err)
                // Continue anyway, file is downloaded
            }
        }
    } else {
        m.client.logger.Error("Failed to stat downloaded file",
            "path", localPath,
            "error", err)
    }

    return localPath, nil
}
```

## Step 4: Update DownloadDocument with Caching

```go
func (m *MediaManager) DownloadDocument(ctx context.Context, doc *tg.Document, chatID int64) (string, error) {
    if doc == nil {
        return "", fmt.Errorf("document is nil")
    }

    // Determine media type for cache key
    mediaType := "document"
    for _, attr := range doc.Attributes {
        switch attr.(type) {
        case *tg.DocumentAttributeVideo:
            mediaType = "video"
        case *tg.DocumentAttributeAudio:
            mediaType = "audio"
        case *tg.DocumentAttributeAnimated:
            mediaType = "animation"
        }
    }

    // Generate cache key
    cacheKey := cache.GenerateKey(chatID, doc.ID, mediaType)

    // Check cache first
    if m.cache != nil {
        if cachedPath, found := m.cache.Get(cacheKey); found {
            m.client.logger.Debug("Using cached document",
                "docID", doc.ID,
                "path", cachedPath)
            return cachedPath, nil
        }
    }

    // Get file extension from mime type or filename
    ext := ".bin"
    fileName := fmt.Sprintf("document_%d_%d", chatID, doc.ID)

    // Try to get filename from attributes
    for _, attr := range doc.Attributes {
        if fileAttr, ok := attr.(*tg.DocumentAttributeFilename); ok {
            fileName = fileAttr.FileName
            break
        }
    }

    // If no filename attribute, try to get extension from mime type
    if !strings.Contains(fileName, ".") {
        exts, err := mime.ExtensionsByType(doc.MimeType)
        if err == nil && len(exts) > 0 {
            ext = exts[0]
            fileName = fmt.Sprintf("%s%s", fileName, ext)
        }
    }

    // Create file location from document
    location := &tg.InputDocumentFileLocation{
        ID:            doc.ID,
        AccessHash:    doc.AccessHash,
        FileReference: doc.FileReference,
        ThumbSize:     "", // Empty for full document
    }

    // Generate local file path
    localPath := filepath.Join(m.mediaDir, fmt.Sprintf("%d", chatID), fileName)

    // Ensure chat directory exists
    if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
        return "", fmt.Errorf("failed to create chat directory: %w", err)
    }

    // Download the file
    if err := m.downloadFile(ctx, location, localPath, doc.Size); err != nil {
        return "", fmt.Errorf("failed to download document: %w", err)
    }

    // Add to cache
    if m.cache != nil {
        if err := m.cache.Put(cacheKey, localPath, doc.Size); err != nil {
            m.client.logger.Warn("Failed to cache document", "error", err)
        }
    }

    return localPath, nil
}
```

## Step 5: Add Cache Management Methods

Add these methods to MediaManager for cache management:

```go
// GetCacheStats returns statistics about the media cache.
func (m *MediaManager) GetCacheStats() cache.CacheStats {
    if m.cache == nil {
        return cache.CacheStats{}
    }
    return m.cache.GetStats()
}

// ClearCache clears all cached media files.
func (m *MediaManager) ClearCache() error {
    if m.cache == nil {
        return nil
    }
    return m.cache.Clear()
}

// ValidateCacheIntegrity checks cache integrity and removes stale entries.
func (m *MediaManager) ValidateCacheIntegrity() (int, error) {
    if m.cache == nil {
        return 0, nil
    }
    return m.cache.ValidateIntegrity()
}

// SetCacheSize adjusts the maximum cache size.
func (m *MediaManager) SetCacheSize(sizeBytes int64) error {
    if m.cache == nil {
        return nil
    }
    return m.cache.SetMaxSize(sizeBytes)
}
```

## Step 6: Add Cache Settings to Config

Update `/Users/lvcasx1/Work/personal/ithil/internal/config/config.go`:

```go
type Config struct {
    Telegram TelegramConfig `yaml:"telegram"`
    UI       UIConfig       `yaml:"ui"`
    Media    MediaConfig    `yaml:"media"` // ADD THIS LINE
}

// ADD THIS STRUCT
type MediaConfig struct {
    CacheEnabled bool   `yaml:"cache_enabled"`
    CacheSizeMB  int    `yaml:"cache_size_mb"`
    CacheDir     string `yaml:"cache_dir"`
}
```

Update `config.example.yaml`:

```yaml
telegram:
  api_id: YOUR_API_ID
  api_hash: "YOUR_API_HASH"
  phone_number: "+1234567890"
  session_file: "session.json"

ui:
  theme: "nord"
  vim_mode: false

media:
  cache_enabled: true
  cache_size_mb: 500
  cache_dir: "~/.cache/ithil/media"
```

## Step 7: Use Config in NewMediaManager

```go
func (c *Client) NewMediaManager(mediaDir string, config MediaConfig) (*MediaManager, error) {
    // Ensure media directory exists
    if err := os.MkdirAll(mediaDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create media directory: %w", err)
    }

    var mediaCache *cache.MediaCache
    if config.CacheEnabled {
        cacheSize := int64(config.CacheSizeMB) * 1024 * 1024
        cacheDir := config.CacheDir
        if cacheDir == "" {
            cacheDir = mediaDir
        }

        var err error
        mediaCache, err = cache.NewMediaCache(cacheSize, cacheDir)
        if err != nil {
            c.logger.Warn("Failed to initialize media cache", "error", err)
            // Continue without cache
        } else {
            c.logger.Info("Media cache initialized",
                "maxSize", cacheSize,
                "dir", cacheDir)
        }
    }

    return &MediaManager{
        client:     c,
        downloader: downloader.NewDownloader(),
        uploader:   uploader.NewUploader(c.api),
        mediaDir:   mediaDir,
        cache:      mediaCache,
    }, nil
}
```

## Step 8: Add UI for Cache Management (Optional)

You can add cache statistics to the status bar or create a cache management screen.

### Status Bar Stats

In `/Users/lvcasx1/Work/personal/ithil/internal/ui/components/statusbar.go`:

```go
func (s *StatusBar) renderCacheInfo() string {
    if s.mediaManager == nil {
        return ""
    }

    stats := s.mediaManager.GetCacheStats()
    if stats.FileCount == 0 {
        return ""
    }

    cacheSizeMB := float64(stats.TotalSize) / (1024 * 1024)
    hitRate := float64(stats.HitCount) / float64(stats.HitCount+stats.MissCount) * 100

    return fmt.Sprintf("Cache: %d files, %.1fMB, %.0f%% hits",
        stats.FileCount, cacheSizeMB, hitRate)
}
```

### Cache Management Menu

Add keyboard shortcuts to the conversation view:

- `Shift+C`: Show cache statistics
- `Ctrl+Shift+C`: Clear cache

```go
// In conversation.go Update() method
case key.Matches(msg, key.NewBinding(key.WithKeys("C"))): // Shift+C
    stats := m.mediaManager.GetCacheStats()
    // Show stats in a modal or status message

case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+C"))): // Ctrl+Shift+C
    if err := m.mediaManager.ClearCache(); err != nil {
        // Show error
    } else {
        // Show success message
    }
```

## Performance Benefits

With the cache integrated, you'll see:

1. **Faster Image Loading**: Cached images load instantly
2. **Reduced Bandwidth**: No re-downloading of previously viewed media
3. **Better Scrolling**: Smooth scrolling through message history with cached media
4. **Offline Access**: View cached media even when offline

## Monitoring

Log cache statistics periodically:

```go
// In main.go or a monitoring goroutine
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        if mediaManager.cache != nil {
            stats := mediaManager.GetCacheStats()
            logger.Info("Media cache stats",
                "files", stats.FileCount,
                "size_mb", float64(stats.TotalSize)/(1024*1024),
                "hits", stats.HitCount,
                "misses", stats.MissCount,
                "evictions", stats.EvictionCount,
            )
        }
    }
}()
```

## Testing Integration

Test the integration:

```go
func TestMediaManagerWithCache(t *testing.T) {
    // Create test client and media manager
    client := NewTestClient()
    mediaManager, err := client.NewMediaManager("/tmp/test_media", MediaConfig{
        CacheEnabled: true,
        CacheSizeMB:  100,
    })
    if err != nil {
        t.Fatal(err)
    }

    // Download photo first time (cache miss)
    path1, err := mediaManager.DownloadPhoto(ctx, photo, chatID)
    if err != nil {
        t.Fatal(err)
    }

    // Download same photo again (cache hit)
    path2, err := mediaManager.DownloadPhoto(ctx, photo, chatID)
    if err != nil {
        t.Fatal(err)
    }

    // Should return same path
    if path1 != path2 {
        t.Errorf("Expected same path, got %s and %s", path1, path2)
    }

    // Verify cache hit
    stats := mediaManager.GetCacheStats()
    if stats.HitCount != 1 {
        t.Errorf("Expected 1 cache hit, got %d", stats.HitCount)
    }
}
```

## Migration Notes

When rolling out the cache:

1. **Backward Compatible**: Existing code works without cache (nil checks)
2. **Opt-In**: Cache is only enabled if configured
3. **No Breaking Changes**: No changes to public API signatures
4. **Graceful Degradation**: If cache fails to initialize, app continues without it

## Troubleshooting

### Cache Not Working

Check:
- `cache_enabled: true` in config
- Cache directory is writable
- Sufficient disk space
- Logs for cache initialization errors

### High Eviction Rate

If `stats.EvictionCount` is high:
- Increase `cache_size_mb` in config
- Check if media files are very large
- Consider cleaning up old chats

### Stale Entries

Run periodic maintenance:
```go
// Every hour
ticker := time.NewTicker(1 * time.Hour)
go func() {
    for range ticker.C {
        removed, _ := mediaManager.ValidateCacheIntegrity()
        if removed > 0 {
            logger.Info("Cleaned up stale cache entries", "count", removed)
        }
    }
}()
```

## Complete Example

See `/Users/lvcasx1/Work/personal/ithil/internal/cache/media_cache_example_test.go` for complete usage examples.
