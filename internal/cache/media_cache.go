// Package cache provides caching functionality for messages and media.
package cache

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// DefaultMaxCacheSize is the default maximum size of the media cache (500 MB).
	DefaultMaxCacheSize = 500 * 1024 * 1024

	// DefaultCacheDir is the default directory for cached media files.
	DefaultCacheDir = "~/.cache/ithil/media"
)

// CachedMedia represents a single cached media file with metadata.
type CachedMedia struct {
	FilePath     string    // Absolute path to the cached file
	FileSize     int64     // Size of the file in bytes
	MediaType    string    // Type of media (photo, video, audio, etc.)
	AccessTime   time.Time // Last time the file was accessed
	DownloadTime time.Time // When the file was downloaded
	element      *list.Element // Reference to the LRU list element
}

// CacheStats contains statistics about the media cache.
type CacheStats struct {
	TotalSize         int64     // Total size of all cached files in bytes
	FileCount         int       // Number of cached files
	HitCount          int64     // Number of cache hits
	MissCount         int64     // Number of cache misses
	EvictionCount     int64     // Number of files evicted
	OldestAccessTime  time.Time // Timestamp of the least recently used file
	NewestAccessTime  time.Time // Timestamp of the most recently used file
}

// MediaCache manages a disk-based LRU cache for media files.
// It is thread-safe and automatically evicts least recently used files when the size limit is exceeded.
type MediaCache struct {
	mu           sync.RWMutex
	cacheDir     string                   // Directory where cached files are stored
	maxSize      int64                    // Maximum cache size in bytes
	currentSize  int64                    // Current cache size in bytes
	files        map[string]*CachedMedia  // Map of cache key to cached media
	lruList      *list.List               // Doubly-linked list for LRU tracking
	hitCount     int64                    // Number of cache hits
	missCount    int64                    // Number of cache misses
	evictCount   int64                    // Number of evictions
}

// NewMediaCache creates a new media cache with the specified maximum size and cache directory.
// If maxSizeBytes is 0, caching is disabled and all operations become no-ops.
// The cache directory will be created if it doesn't exist.
func NewMediaCache(maxSizeBytes int64, cacheDir string) (*MediaCache, error) {
	// Expand home directory if needed
	if len(cacheDir) > 0 && cacheDir[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, cacheDir[1:])
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	mc := &MediaCache{
		cacheDir:    cacheDir,
		maxSize:     maxSizeBytes,
		files:       make(map[string]*CachedMedia),
		lruList:     list.New(),
		currentSize: 0,
	}

	// Load existing cache entries from disk
	if err := mc.loadExistingFiles(); err != nil {
		// Log the error but don't fail - we'll start with an empty cache
		fmt.Fprintf(os.Stderr, "warning: failed to load existing cache: %v\n", err)
	}

	return mc, nil
}

// NewDefaultMediaCache creates a new media cache with default settings.
func NewDefaultMediaCache() (*MediaCache, error) {
	return NewMediaCache(DefaultMaxCacheSize, DefaultCacheDir)
}

// loadExistingFiles scans the cache directory and loads existing files into the cache.
// This allows the cache to persist across application restarts.
func (mc *MediaCache) loadExistingFiles() error {
	// Walk through the cache directory
	return filepath.Walk(mc.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Generate key from relative path
		relPath, err := filepath.Rel(mc.cacheDir, path)
		if err != nil {
			return err
		}

		// Create cache entry
		cached := &CachedMedia{
			FilePath:     path,
			FileSize:     info.Size(),
			MediaType:    "unknown",
			AccessTime:   info.ModTime(),
			DownloadTime: info.ModTime(),
		}

		// Add to LRU list (oldest first, will be pushed to back on access)
		cached.element = mc.lruList.PushBack(relPath)

		// Add to map
		mc.files[relPath] = cached
		mc.currentSize += info.Size()

		return nil
	})
}

// Put adds a file to the cache. If the cache is full, it will evict least recently used files
// to make space. The file at filePath should already exist on disk.
func (mc *MediaCache) Put(key string, filePath string, size int64) error {
	if mc.maxSize == 0 {
		return nil // Caching disabled
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	actualSize := fileInfo.Size()
	if size == 0 {
		size = actualSize
	}

	// If file already exists in cache, update it
	if existing, exists := mc.files[key]; exists {
		// Remove old size from current size
		mc.currentSize -= existing.FileSize

		// Update access time and move to front of LRU list
		existing.AccessTime = time.Now()
		existing.FileSize = size
		mc.lruList.MoveToFront(existing.element)

		// Add new size to current size
		mc.currentSize += size
		return nil
	}

	// Evict files if necessary to make space
	for mc.currentSize+size > mc.maxSize && mc.lruList.Len() > 0 {
		if err := mc.evictLRU(); err != nil {
			return fmt.Errorf("failed to evict LRU file: %w", err)
		}
	}

	// If file is still too large after eviction, return error
	if size > mc.maxSize {
		return fmt.Errorf("file size (%d bytes) exceeds maximum cache size (%d bytes)", size, mc.maxSize)
	}

	// Determine media type from file extension
	mediaType := getMediaTypeFromPath(filePath)

	// Create new cache entry
	cached := &CachedMedia{
		FilePath:     filePath,
		FileSize:     size,
		MediaType:    mediaType,
		AccessTime:   time.Now(),
		DownloadTime: time.Now(),
	}

	// Add to front of LRU list (most recently used)
	cached.element = mc.lruList.PushFront(key)

	// Add to map
	mc.files[key] = cached
	mc.currentSize += size

	return nil
}

// Get retrieves a file path from the cache and updates its access time.
// Returns the file path and true if found, empty string and false otherwise.
func (mc *MediaCache) Get(key string) (string, bool) {
	if mc.maxSize == 0 {
		return "", false // Caching disabled
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	cached, exists := mc.files[key]
	if !exists {
		mc.missCount++
		return "", false
	}

	// Verify file still exists on disk
	if _, err := os.Stat(cached.FilePath); os.IsNotExist(err) {
		// File was deleted externally, remove from cache
		mc.removeUnsafe(key)
		mc.missCount++
		return "", false
	}

	// Update access time and move to front of LRU list
	cached.AccessTime = time.Now()
	mc.lruList.MoveToFront(cached.element)
	mc.hitCount++

	return cached.FilePath, true
}

// Remove removes a specific file from the cache and deletes it from disk.
func (mc *MediaCache) Remove(key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	return mc.removeUnsafe(key)
}

// removeUnsafe removes a file from the cache without locking.
// Must be called with the lock held.
func (mc *MediaCache) removeUnsafe(key string) error {
	cached, exists := mc.files[key]
	if !exists {
		return nil // Already removed
	}

	// Remove from LRU list
	mc.lruList.Remove(cached.element)

	// Remove from map
	delete(mc.files, key)

	// Update current size
	mc.currentSize -= cached.FileSize

	// Delete file from disk
	if err := os.Remove(cached.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Clear removes all files from the cache and deletes them from disk.
func (mc *MediaCache) Clear() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var errs []error

	// Remove all files
	for key := range mc.files {
		if err := mc.removeUnsafe(key); err != nil {
			errs = append(errs, err)
		}
	}

	// Reset cache
	mc.files = make(map[string]*CachedMedia)
	mc.lruList = list.New()
	mc.currentSize = 0

	if len(errs) > 0 {
		return fmt.Errorf("failed to clear cache: %v", errs)
	}

	return nil
}

// GetStats returns current cache statistics.
func (mc *MediaCache) GetStats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := CacheStats{
		TotalSize:     mc.currentSize,
		FileCount:     len(mc.files),
		HitCount:      mc.hitCount,
		MissCount:     mc.missCount,
		EvictionCount: mc.evictCount,
	}

	// Find oldest and newest access times
	if mc.lruList.Len() > 0 {
		// Back of list is LRU (oldest)
		if back := mc.lruList.Back(); back != nil {
			if key, ok := back.Value.(string); ok {
				if cached, exists := mc.files[key]; exists {
					stats.OldestAccessTime = cached.AccessTime
				}
			}
		}

		// Front of list is MRU (newest)
		if front := mc.lruList.Front(); front != nil {
			if key, ok := front.Value.(string); ok {
				if cached, exists := mc.files[key]; exists {
					stats.NewestAccessTime = cached.AccessTime
				}
			}
		}
	}

	return stats
}

// Evict manually triggers eviction of the least recently used file.
// Returns an error if the cache is empty or if eviction fails.
func (mc *MediaCache) Evict() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	return mc.evictLRU()
}

// evictLRU evicts the least recently used file from the cache.
// Must be called with the lock held.
func (mc *MediaCache) evictLRU() error {
	if mc.lruList.Len() == 0 {
		return fmt.Errorf("cache is empty, nothing to evict")
	}

	// Get the least recently used item (back of list)
	back := mc.lruList.Back()
	if back == nil {
		return fmt.Errorf("failed to get LRU item")
	}

	key, ok := back.Value.(string)
	if !ok {
		return fmt.Errorf("invalid LRU list element")
	}

	// Remove the file
	if err := mc.removeUnsafe(key); err != nil {
		return err
	}

	mc.evictCount++
	return nil
}

// SetMaxSize adjusts the maximum cache size. If the new size is smaller than the current size,
// files will be evicted until the cache fits within the new limit.
func (mc *MediaCache) SetMaxSize(maxSizeBytes int64) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.maxSize = maxSizeBytes

	// If caching is disabled, clear everything
	if maxSizeBytes == 0 {
		for key := range mc.files {
			mc.removeUnsafe(key)
		}
		return nil
	}

	// Evict files until we're under the new limit
	for mc.currentSize > mc.maxSize && mc.lruList.Len() > 0 {
		if err := mc.evictLRU(); err != nil {
			return fmt.Errorf("failed to evict during resize: %w", err)
		}
	}

	return nil
}

// GetSize returns the current total size of cached files in bytes.
func (mc *MediaCache) GetSize() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.currentSize
}

// GetCount returns the number of files currently in the cache.
func (mc *MediaCache) GetCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return len(mc.files)
}

// CleanupOrphans removes files from disk that are in the cache directory but not tracked in the cache.
// This is useful for cleaning up after crashes or manual file additions.
func (mc *MediaCache) CleanupOrphans() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var orphans []string

	// Walk through cache directory to find orphaned files
	err := filepath.Walk(mc.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the root cache directory
		if info.IsDir() {
			return nil
		}

		// Generate key from relative path
		relPath, err := filepath.Rel(mc.cacheDir, path)
		if err != nil {
			return err
		}

		// Check if file is tracked in cache
		if _, exists := mc.files[relPath]; !exists {
			orphans = append(orphans, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan for orphans: %w", err)
	}

	// Delete orphaned files
	var deleteErrs []error
	for _, orphan := range orphans {
		if err := os.Remove(orphan); err != nil {
			deleteErrs = append(deleteErrs, err)
		}
	}

	if len(deleteErrs) > 0 {
		return fmt.Errorf("failed to delete some orphaned files: %v", deleteErrs)
	}

	return nil
}

// ValidateIntegrity checks that all cached files still exist on disk and removes any that don't.
// Returns the number of invalid entries removed.
func (mc *MediaCache) ValidateIntegrity() (int, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var invalidKeys []string

	// Check each cached file
	for key, cached := range mc.files {
		if _, err := os.Stat(cached.FilePath); os.IsNotExist(err) {
			invalidKeys = append(invalidKeys, key)
		}
	}

	// Remove invalid entries
	for _, key := range invalidKeys {
		mc.removeUnsafe(key)
	}

	return len(invalidKeys), nil
}

// getMediaTypeFromPath determines the media type based on file path/extension.
func getMediaTypeFromPath(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "photo"
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return "video"
	case ".mp3", ".wav", ".flac", ".m4a", ".aac":
		return "audio"
	case ".ogg", ".opus":
		return "voice"
	default:
		return "document"
	}
}

// GenerateKey generates a cache key from chat ID, message ID, and media type.
// This is a helper function for consistent key generation across the application.
func GenerateKey(chatID, messageID int64, mediaType string) string {
	return fmt.Sprintf("%d_%d_%s", chatID, messageID, mediaType)
}

// GenerateThumbnailKey generates a cache key for a thumbnail.
func GenerateThumbnailKey(chatID, messageID int64, mediaType string) string {
	return fmt.Sprintf("%d_%d_%s_thumb", chatID, messageID, mediaType)
}
