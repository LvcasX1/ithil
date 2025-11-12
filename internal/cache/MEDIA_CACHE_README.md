# Media Cache System

A thread-safe, LRU (Least Recently Used) disk-based cache for media files in the Ithil Telegram TUI client.

## Overview

The media cache system provides efficient caching of downloaded media files (photos, videos, audio, documents) with automatic eviction based on the LRU policy when size limits are reached. It persists across application restarts by scanning the cache directory on initialization.

## Features

- **LRU Eviction Policy**: Automatically removes least recently used files when cache is full
- **Thread-Safe**: All operations use RWMutex for safe concurrent access
- **Persistent Storage**: Files remain on disk between application restarts
- **Size Management**: Configurable maximum cache size with automatic eviction
- **Statistics Tracking**: Hit/miss rates, eviction counts, access times
- **Cache Validation**: Detect and remove stale cache entries
- **Orphan Cleanup**: Remove files not tracked by the cache
- **Flexible Configuration**: Customizable cache directory and size limits
- **Disable Support**: Can be completely disabled by setting size to 0

## Architecture

### Core Components

#### MediaCache
The main cache manager that tracks files, enforces size limits, and manages LRU eviction.

```go
type MediaCache struct {
    mu           sync.RWMutex
    cacheDir     string                   // Cache directory path
    maxSize      int64                    // Maximum cache size in bytes
    currentSize  int64                    // Current total size
    files        map[string]*CachedMedia  // Key -> cached file metadata
    lruList      *list.List               // Doubly-linked list for LRU
    hitCount     int64                    // Cache hits
    missCount    int64                    // Cache misses
    evictCount   int64                    // Files evicted
}
```

#### CachedMedia
Metadata for a single cached file including file path, size, type, and access times.

```go
type CachedMedia struct {
    FilePath     string    // Absolute path to cached file
    FileSize     int64     // Size in bytes
    MediaType    string    // photo, video, audio, voice, document
    AccessTime   time.Time // Last access time
    DownloadTime time.Time // Download timestamp
    element      *list.Element // LRU list reference
}
```

#### CacheStats
Statistics about cache performance and state.

```go
type CacheStats struct {
    TotalSize         int64     // Total cached bytes
    FileCount         int       // Number of cached files
    HitCount          int64     // Cache hits
    MissCount         int64     // Cache misses
    EvictionCount     int64     // Files evicted
    OldestAccessTime  time.Time // LRU file timestamp
    NewestAccessTime  time.Time // MRU file timestamp
}
```

## Usage

### Basic Usage

```go
// Create cache with 500MB limit in default directory
cache, err := cache.NewDefaultMediaCache()
if err != nil {
    log.Fatal(err)
}

// Or with custom settings
cache, err := cache.NewMediaCache(100*1024*1024, "/custom/cache/dir")
if err != nil {
    log.Fatal(err)
}

// Generate cache key
key := cache.GenerateKey(chatID, messageID, "photo")

// Add file to cache
err = cache.Put(key, "/path/to/downloaded/file.jpg", fileSize)

// Retrieve from cache
if filePath, found := cache.Get(key); found {
    // Use cached file
    fmt.Printf("Using cached file: %s\n", filePath)
} else {
    // Download file
    fmt.Println("Cache miss, downloading...")
}
```

### Integration with Media Manager

The media cache integrates seamlessly with Ithil's media download system:

```go
// In MediaManager
type MediaManager struct {
    client     *Client
    downloader *downloader.Downloader
    uploader   *uploader.Uploader
    mediaDir   string
    cache      *cache.MediaCache // Add this
}

// Initialize cache
func (m *MediaManager) InitCache(maxSize int64) error {
    var err error
    m.cache, err = cache.NewMediaCache(maxSize, m.mediaDir)
    return err
}

// Modified download method with caching
func (m *MediaManager) DownloadPhoto(ctx context.Context, photo *tg.Photo, chatID int64) (string, error) {
    // Generate cache key
    key := cache.GenerateKey(chatID, photo.ID, "photo")

    // Check cache first
    if cachedPath, found := m.cache.Get(key); found {
        m.client.logger.Info("Using cached photo", "photoID", photo.ID)
        return cachedPath, nil
    }

    // Not cached, download normally
    localPath, err := m.downloadPhotoFromTelegram(ctx, photo, chatID)
    if err != nil {
        return "", err
    }

    // Add to cache
    fileInfo, _ := os.Stat(localPath)
    if err := m.cache.Put(key, localPath, fileInfo.Size()); err != nil {
        m.client.logger.Warn("Failed to cache photo", "error", err)
    }

    return localPath, nil
}
```

### Thumbnail Support

The cache supports separate entries for full images and thumbnails:

```go
// Cache full image
fullKey := cache.GenerateKey(chatID, msgID, "photo")
cache.Put(fullKey, fullImagePath, fullSize)

// Cache thumbnail
thumbKey := cache.GenerateThumbnailKey(chatID, msgID, "photo")
cache.Put(thumbKey, thumbPath, thumbSize)
```

### Statistics and Monitoring

```go
stats := cache.GetStats()
fmt.Printf("Cache Statistics:\n")
fmt.Printf("  Files: %d\n", stats.FileCount)
fmt.Printf("  Size: %.2f MB\n", float64(stats.TotalSize)/(1024*1024))
fmt.Printf("  Hit Rate: %.1f%%\n",
    float64(stats.HitCount)/float64(stats.HitCount+stats.MissCount)*100)
fmt.Printf("  Evictions: %d\n", stats.EvictionCount)
```

### Cache Maintenance

```go
// Validate integrity (remove entries for missing files)
removed, err := cache.ValidateIntegrity()
fmt.Printf("Removed %d stale entries\n", removed)

// Clean up orphaned files
err = cache.CleanupOrphans()

// Adjust cache size
err = cache.SetMaxSize(200 * 1024 * 1024) // 200MB

// Manually trigger eviction
err = cache.Evict()

// Clear entire cache
err = cache.Clear()
```

## Configuration

### Default Settings

```go
const (
    DefaultMaxCacheSize = 500 * 1024 * 1024  // 500 MB
    DefaultCacheDir     = "~/.cache/ithil/media"
)
```

### Custom Configuration

Users can configure the cache via config file:

```yaml
# config.yaml
media:
  cache_dir: "~/.cache/ithil/media"
  cache_size_mb: 500
  cache_enabled: true
```

### Disabling Cache

Set max size to 0 to disable caching:

```go
cache, _ := cache.NewMediaCache(0, cacheDir)
// All Put/Get operations become no-ops
```

## Cache Key Format

Cache keys are generated using a consistent format:

```go
// Full media: "{chatID}_{messageID}_{mediaType}"
key := cache.GenerateKey(123456, 789012, "photo")
// Result: "123456_789012_photo"

// Thumbnails: "{chatID}_{messageID}_{mediaType}_thumb"
thumbKey := cache.GenerateThumbnailKey(123456, 789012, "photo")
// Result: "123456_789012_photo_thumb"
```

Media types: `photo`, `video`, `audio`, `voice`, `document`

## LRU Algorithm

The cache uses a doubly-linked list to track access order:

1. **Put**: New files are added to the front (most recently used)
2. **Get**: Accessed files move to the front
3. **Evict**: Files at the back (least recently used) are removed first
4. **Automatic**: Eviction happens automatically when size limit is reached

### Example LRU Flow

```
Initial state (empty cache, 2KB limit):
  []

After Put("A", 1KB):
  [A] (1KB)

After Put("B", 1KB):
  [B, A] (2KB, cache full)

After Get("A"):
  [A, B] (A moved to front)

After Put("C", 1KB):
  [C, A] (3KB would exceed limit, B evicted)
  Evicted: B (was LRU)
```

## Thread Safety

All public methods are thread-safe using `sync.RWMutex`:

- Read operations (`Get`, `GetStats`, `GetSize`, `GetCount`) use read locks
- Write operations (`Put`, `Remove`, `Clear`, `Evict`) use write locks
- Internal unsafe methods (`removeUnsafe`, `evictLRU`) must be called with lock held

## Performance Considerations

### Memory Usage

- **Map overhead**: ~48 bytes per entry
- **List overhead**: ~32 bytes per element
- **CachedMedia**: ~120 bytes per entry
- **Total**: ~200 bytes metadata per cached file

For 1000 cached files: ~200KB memory overhead

### Disk I/O

- **Reads**: Only on Get miss validation (stat call)
- **Writes**: None (cache tracks files already on disk)
- **Deletes**: On eviction and explicit removal

### Lock Contention

- Use `RWMutex` to allow concurrent reads
- Write operations (Put/Remove) acquire exclusive lock
- Statistics methods use read lock only

## Error Handling

The cache handles several error scenarios:

### File Not Found
```go
// If cached file is deleted externally
path, found := cache.Get("key")
// Returns: "", false (removes stale entry)
```

### Disk Full
```go
err := cache.Put("key", path, size)
// Returns error if eviction fails due to disk issues
```

### File Too Large
```go
err := cache.Put("key", path, 1GB)
// Returns error if file exceeds max cache size
```

### Directory Creation
```go
cache, err := cache.NewMediaCache(size, "/invalid/path")
// Returns error if directory cannot be created
```

## Testing

Comprehensive test suite covers:

- Basic Put/Get operations
- LRU eviction policy
- Access order tracking
- Cache statistics
- Concurrent access
- Size management
- Integrity validation
- Edge cases (disabled cache, file too large, etc.)

Run tests:

```bash
go test -v ./internal/cache -run TestMediaCache
```

Run with race detection:

```bash
go test -race ./internal/cache
```

## Future Enhancements

Potential improvements for future versions:

1. **TTL Support**: Expire entries after time duration
2. **Compression**: Compress cached files to save space
3. **Metrics Export**: Prometheus metrics for monitoring
4. **Background Cleanup**: Periodic validation and cleanup goroutine
5. **Smart Prefetch**: Prefetch likely-to-be-accessed media
6. **Tiered Storage**: Hot/cold storage separation
7. **Deduplication**: Detect duplicate files across chats
8. **Encryption**: Encrypt cached files at rest

## Performance Benchmarks

Expected performance characteristics:

- **Get (hit)**: O(1) - map lookup + list move
- **Get (miss)**: O(1) - map lookup + stat syscall
- **Put (new)**: O(1) - map insert + list prepend
- **Put (evict)**: O(n) where n = files to evict (usually 1)
- **Memory**: O(n) where n = cached file count

Typical operations per second (on modern SSD):
- Get: 1M+ ops/sec
- Put: 100K+ ops/sec
- Evict: 10K+ ops/sec

## License

Part of the Ithil project. See main LICENSE file.
