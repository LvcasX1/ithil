package cache_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lvcasx1/ithil/internal/cache"
)

// Example demonstrates basic usage of the MediaCache.
func Example() {
	// Create a new media cache with 100MB limit
	mediaCache, err := cache.NewMediaCache(100*1024*1024, "/tmp/ithil_cache")
	if err != nil {
		log.Fatalf("Failed to create media cache: %v", err)
	}

	// Generate a cache key for a photo
	chatID := int64(123456)
	messageID := int64(789012)
	key := cache.GenerateKey(chatID, messageID, "photo")

	// Example: After downloading a file, add it to cache
	filePath := "/path/to/downloaded/photo.jpg"
	fileSize := int64(2048576) // 2MB

	// Put the file in cache
	if err := mediaCache.Put(key, filePath, fileSize); err != nil {
		log.Printf("Failed to cache file: %v", err)
	}

	// Later, retrieve the file from cache
	if cachedPath, found := mediaCache.Get(key); found {
		fmt.Printf("File found in cache: %s\n", cachedPath)
		// Use the cached file...
	} else {
		fmt.Println("File not in cache, need to download")
	}

	// Get cache statistics
	stats := mediaCache.GetStats()
	fmt.Printf("Cache stats: %d files, %d bytes, %d hits, %d misses\n",
		stats.FileCount, stats.TotalSize, stats.HitCount, stats.MissCount)
}

// Example_withThumbnails demonstrates caching both full images and thumbnails.
func Example_withThumbnails() {
	mediaCache, err := cache.NewMediaCache(100*1024*1024, "/tmp/ithil_cache")
	if err != nil {
		log.Fatalf("Failed to create media cache: %v", err)
	}

	chatID := int64(123456)
	messageID := int64(789012)

	// Cache the full image
	fullImageKey := cache.GenerateKey(chatID, messageID, "photo")
	mediaCache.Put(fullImageKey, "/path/to/photo.jpg", 2048576)

	// Cache the thumbnail
	thumbKey := cache.GenerateThumbnailKey(chatID, messageID, "photo")
	mediaCache.Put(thumbKey, "/path/to/photo_thumb.jpg", 10240)

	// Retrieve both
	if fullPath, found := mediaCache.Get(fullImageKey); found {
		fmt.Printf("Full image: %s\n", fullPath)
	}
	if thumbPath, found := mediaCache.Get(thumbKey); found {
		fmt.Printf("Thumbnail: %s\n", thumbPath)
	}
}

// Example_managingCacheSize demonstrates cache size management and eviction.
func Example_managingCacheSize() {
	// Create cache with 5MB limit
	mediaCache, err := cache.NewMediaCache(5*1024*1024, "/tmp/ithil_cache")
	if err != nil {
		log.Fatalf("Failed to create media cache: %v", err)
	}

	// Add files until cache is full
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("file_%d", i)
		path := fmt.Sprintf("/tmp/file_%d.jpg", i)
		// Simulate 1MB file
		mediaCache.Put(key, path, 1024*1024)
	}

	// Check current size
	fmt.Printf("Current cache size: %d bytes\n", mediaCache.GetSize())

	// Get statistics to see evictions
	stats := mediaCache.GetStats()
	fmt.Printf("Files evicted due to size limit: %d\n", stats.EvictionCount)

	// Manually adjust cache size
	if err := mediaCache.SetMaxSize(2 * 1024 * 1024); err != nil {
		log.Printf("Failed to set max size: %v", err)
	}

	// Check new size (should have evicted more files)
	fmt.Printf("New cache size: %d bytes\n", mediaCache.GetSize())
}

// Example_validation demonstrates cache integrity validation.
func Example_validation() {
	mediaCache, err := cache.NewMediaCache(100*1024*1024, "/tmp/ithil_cache")
	if err != nil {
		log.Fatalf("Failed to create media cache: %v", err)
	}

	// Periodically validate cache integrity
	// This removes entries for files that no longer exist on disk
	removed, err := mediaCache.ValidateIntegrity()
	if err != nil {
		log.Printf("Validation error: %v", err)
	}
	fmt.Printf("Removed %d invalid cache entries\n", removed)

	// Clean up orphaned files (files on disk not tracked in cache)
	if err := mediaCache.CleanupOrphans(); err != nil {
		log.Printf("Cleanup error: %v", err)
	}
}

// Example_integration demonstrates how to integrate MediaCache with the media download system.
func Example_integration() {
	// Initialize media cache
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "ithil", "media")
	mediaCache, err := cache.NewMediaCache(500*1024*1024, cacheDir)
	if err != nil {
		log.Fatalf("Failed to create media cache: %v", err)
	}

	// Example download function that uses cache
	downloadPhoto := func(chatID, messageID int64, photoURL string) (string, error) {
		// Generate cache key
		key := cache.GenerateKey(chatID, messageID, "photo")

		// Check if already cached
		if cachedPath, found := mediaCache.Get(key); found {
			fmt.Println("Using cached photo")
			return cachedPath, nil
		}

		// Not in cache, download it
		fmt.Println("Downloading photo...")
		downloadedPath := "/tmp/downloaded_photo.jpg"

		// Simulate download
		// In real code: download from photoURL to downloadedPath

		// Get file size
		fileInfo, err := os.Stat(downloadedPath)
		if err != nil {
			return "", err
		}

		// Add to cache
		if err := mediaCache.Put(key, downloadedPath, fileInfo.Size()); err != nil {
			log.Printf("Warning: failed to cache file: %v", err)
			// Continue anyway, return the downloaded file
		}

		return downloadedPath, nil
	}

	// Use the download function
	path, err := downloadPhoto(123, 456, "https://example.com/photo.jpg")
	if err != nil {
		log.Printf("Download failed: %v", err)
	} else {
		fmt.Printf("Photo available at: %s\n", path)
	}

	// Print cache statistics
	stats := mediaCache.GetStats()
	fmt.Printf("Cache: %d files, %.2f MB, hit rate: %.1f%%\n",
		stats.FileCount,
		float64(stats.TotalSize)/(1024*1024),
		float64(stats.HitCount)/float64(stats.HitCount+stats.MissCount)*100,
	)
}

// Example_disabledCache demonstrates how caching can be disabled.
func Example_disabledCache() {
	// Create cache with maxSize=0 to disable caching
	mediaCache, err := cache.NewMediaCache(0, "/tmp/ithil_cache")
	if err != nil {
		log.Fatalf("Failed to create media cache: %v", err)
	}

	// All operations become no-ops when cache is disabled
	mediaCache.Put("key1", "/path/to/file", 1024) // No-op
	_, found := mediaCache.Get("key1")             // Always returns false

	fmt.Printf("Cache disabled, found: %v\n", found)
	// Output: Cache disabled, found: false
}
