package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestFile creates a temporary test file with the given size.
func createTestFile(t *testing.T, dir string, name string, size int64) string {
	path := filepath.Join(dir, name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	// Write data to reach the specified size
	if size > 0 {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		if _, err := file.Write(data); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	return path
}

func TestNewMediaCache(t *testing.T) {
	tempDir := t.TempDir()

	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	if cache.maxSize != 1024*1024 {
		t.Errorf("Expected maxSize 1048576, got %d", cache.maxSize)
	}

	if cache.cacheDir != tempDir {
		t.Errorf("Expected cacheDir %s, got %s", tempDir, cache.cacheDir)
	}

	// Check that directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created")
	}
}

func TestMediaCache_Put_Get(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Create a test file
	testFile := createTestFile(t, tempDir, "test.jpg", 1024)

	// Put the file in cache
	key := "test_key"
	err = cache.Put(key, testFile, 1024)
	if err != nil {
		t.Fatalf("Failed to put file in cache: %v", err)
	}

	// Get the file from cache
	path, found := cache.Get(key)
	if !found {
		t.Fatalf("File not found in cache")
	}

	if path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, path)
	}

	// Verify stats
	stats := cache.GetStats()
	if stats.FileCount != 1 {
		t.Errorf("Expected FileCount 1, got %d", stats.FileCount)
	}
	if stats.TotalSize != 1024 {
		t.Errorf("Expected TotalSize 1024, got %d", stats.TotalSize)
	}
	if stats.HitCount != 1 {
		t.Errorf("Expected HitCount 1, got %d", stats.HitCount)
	}
}

func TestMediaCache_Get_Miss(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Try to get non-existent file
	_, found := cache.Get("nonexistent")
	if found {
		t.Errorf("Expected cache miss, but file was found")
	}

	stats := cache.GetStats()
	if stats.MissCount != 1 {
		t.Errorf("Expected MissCount 1, got %d", stats.MissCount)
	}
}

func TestMediaCache_LRU_Eviction(t *testing.T) {
	tempDir := t.TempDir()
	// Create cache with 2KB limit
	cache, err := NewMediaCache(2048, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Create and add three 1KB files
	file1 := createTestFile(t, tempDir, "file1.jpg", 1024)
	file2 := createTestFile(t, tempDir, "file2.jpg", 1024)
	file3 := createTestFile(t, tempDir, "file3.jpg", 1024)

	// Add file1
	if err := cache.Put("key1", file1, 1024); err != nil {
		t.Fatalf("Failed to put file1: %v", err)
	}

	// Add file2
	if err := cache.Put("key2", file2, 1024); err != nil {
		t.Fatalf("Failed to put file2: %v", err)
	}

	// Cache should now be full (2KB)
	if cache.GetSize() != 2048 {
		t.Errorf("Expected cache size 2048, got %d", cache.GetSize())
	}

	// Add file3 - this should evict file1 (LRU)
	if err := cache.Put("key3", file3, 1024); err != nil {
		t.Fatalf("Failed to put file3: %v", err)
	}

	// file1 should be evicted
	if _, found := cache.Get("key1"); found {
		t.Errorf("Expected key1 to be evicted")
	}

	// file2 and file3 should still be present
	if _, found := cache.Get("key2"); !found {
		t.Errorf("Expected key2 to be present")
	}
	if _, found := cache.Get("key3"); !found {
		t.Errorf("Expected key3 to be present")
	}

	// Check eviction count
	stats := cache.GetStats()
	if stats.EvictionCount != 1 {
		t.Errorf("Expected EvictionCount 1, got %d", stats.EvictionCount)
	}
}

func TestMediaCache_LRU_AccessOrder(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(2048, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Create and add two 1KB files
	file1 := createTestFile(t, tempDir, "file1.jpg", 1024)
	file2 := createTestFile(t, tempDir, "file2.jpg", 1024)
	file3 := createTestFile(t, tempDir, "file3.jpg", 1024)

	cache.Put("key1", file1, 1024)
	cache.Put("key2", file2, 1024)

	// Access key1 to make it more recently used
	cache.Get("key1")

	// Add file3 - this should evict key2 (now LRU), not key1
	cache.Put("key3", file3, 1024)

	// key1 should still be present
	if _, found := cache.Get("key1"); !found {
		t.Errorf("Expected key1 to be present after access")
	}

	// key2 should be evicted
	if _, found := cache.Get("key2"); found {
		t.Errorf("Expected key2 to be evicted")
	}

	// key3 should be present
	if _, found := cache.Get("key3"); !found {
		t.Errorf("Expected key3 to be present")
	}
}

func TestMediaCache_Remove(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	testFile := createTestFile(t, tempDir, "test.jpg", 1024)
	key := "test_key"

	cache.Put(key, testFile, 1024)

	// Verify file exists
	if _, found := cache.Get(key); !found {
		t.Fatalf("File should be in cache before removal")
	}

	// Remove the file
	if err := cache.Remove(key); err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// Verify file is gone from cache
	if _, found := cache.Get(key); found {
		t.Errorf("File should not be in cache after removal")
	}

	// Verify file is deleted from disk
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("File should be deleted from disk")
	}

	// Verify size is updated
	if cache.GetSize() != 0 {
		t.Errorf("Expected cache size 0 after removal, got %d", cache.GetSize())
	}
}

func TestMediaCache_Clear(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Add multiple files
	for i := 0; i < 5; i++ {
		fileName := "test" + string(rune('a'+i)) + ".jpg"
		testFile := createTestFile(t, tempDir, fileName, 1024)
		cache.Put("key"+string(rune('a'+i)), testFile, 1024)
	}

	// Verify files are in cache
	if cache.GetCount() != 5 {
		t.Fatalf("Expected 5 files in cache, got %d", cache.GetCount())
	}

	// Clear the cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify cache is empty
	if cache.GetCount() != 0 {
		t.Errorf("Expected 0 files after clear, got %d", cache.GetCount())
	}
	if cache.GetSize() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.GetSize())
	}
}

func TestMediaCache_SetMaxSize(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(4096, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Add three 1KB files
	for i := 0; i < 3; i++ {
		fileName := "test" + string(rune('a'+i)) + ".jpg"
		testFile := createTestFile(t, tempDir, fileName, 1024)
		cache.Put("key"+string(rune('a'+i)), testFile, 1024)
	}

	// Cache should have 3KB
	if cache.GetSize() != 3072 {
		t.Errorf("Expected cache size 3072, got %d", cache.GetSize())
	}

	// Reduce max size to 2KB - should evict 1 file
	if err := cache.SetMaxSize(2048); err != nil {
		t.Fatalf("Failed to set max size: %v", err)
	}

	// Should have evicted at least one file
	if cache.GetSize() > 2048 {
		t.Errorf("Cache size %d exceeds new max size 2048", cache.GetSize())
	}
}

func TestMediaCache_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Add a file
	testFile := createTestFile(t, tempDir, "test.jpg", 1024)
	cache.Put("key1", testFile, 1024)

	// Perform some operations
	cache.Get("key1")      // hit
	cache.Get("nonexist")  // miss

	stats := cache.GetStats()

	if stats.FileCount != 1 {
		t.Errorf("Expected FileCount 1, got %d", stats.FileCount)
	}
	if stats.TotalSize != 1024 {
		t.Errorf("Expected TotalSize 1024, got %d", stats.TotalSize)
	}
	if stats.HitCount != 1 {
		t.Errorf("Expected HitCount 1, got %d", stats.HitCount)
	}
	if stats.MissCount != 1 {
		t.Errorf("Expected MissCount 1, got %d", stats.MissCount)
	}
	if stats.OldestAccessTime.IsZero() {
		t.Errorf("Expected OldestAccessTime to be set")
	}
	if stats.NewestAccessTime.IsZero() {
		t.Errorf("Expected NewestAccessTime to be set")
	}
}

func TestMediaCache_ValidateIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Add a file
	testFile := createTestFile(t, tempDir, "test.jpg", 1024)
	cache.Put("key1", testFile, 1024)

	// Delete the file externally
	os.Remove(testFile)

	// Validate integrity - should remove the invalid entry
	removed, err := cache.ValidateIntegrity()
	if err != nil {
		t.Fatalf("Failed to validate integrity: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected 1 invalid entry removed, got %d", removed)
	}

	if cache.GetCount() != 0 {
		t.Errorf("Expected cache to be empty after validation, got %d files", cache.GetCount())
	}
}

func TestMediaCache_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(10*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Test concurrent access
	done := make(chan bool, 3)
	iterations := 50
	timeout := time.After(5 * time.Second)

	// Writer goroutine
	go func() {
		defer func() { done <- true }()
		for i := 0; i < iterations; i++ {
			fileName := "write" + string(rune('a'+i%26)) + string(rune('a'+i/26)) + ".jpg"
			testFile := createTestFile(t, tempDir, fileName, 100)
			cache.Put("write_"+string(rune('a'+i%26))+"_"+string(rune('a'+i/26)), testFile, 100)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Reader goroutine
	go func() {
		defer func() { done <- true }()
		for i := 0; i < iterations; i++ {
			cache.Get("write_" + string(rune('a'+i%26)) + "_" + string(rune('a'+i/26)))
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Stats reader goroutine
	go func() {
		defer func() { done <- true }()
		for i := 0; i < iterations; i++ {
			cache.GetStats()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Wait for all goroutines with timeout
	completed := 0
	for completed < 3 {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Test timed out waiting for concurrent operations")
		}
	}

	// If we got here without deadlock or race conditions, test passes
}

func TestMediaCache_DisabledCache(t *testing.T) {
	tempDir := t.TempDir()
	// Create cache with maxSize = 0 (disabled)
	cache, err := NewMediaCache(0, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	testFile := createTestFile(t, tempDir, "test.jpg", 1024)

	// Put should be a no-op
	err = cache.Put("key1", testFile, 1024)
	if err != nil {
		t.Errorf("Put should not return error when cache is disabled: %v", err)
	}

	// Get should return not found
	if _, found := cache.Get("key1"); found {
		t.Errorf("Get should return false when cache is disabled")
	}

	if cache.GetCount() != 0 {
		t.Errorf("Expected count 0 when cache is disabled, got %d", cache.GetCount())
	}
}

func TestGenerateKey(t *testing.T) {
	key := GenerateKey(123, 456, "photo")
	expected := "123_456_photo"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}

func TestGenerateThumbnailKey(t *testing.T) {
	key := GenerateThumbnailKey(123, 456, "photo")
	expected := "123_456_photo_thumb"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}

func TestMediaCache_FileSizeExceedsMax(t *testing.T) {
	tempDir := t.TempDir()
	// Create cache with 1KB limit
	cache, err := NewMediaCache(1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Try to add a 2KB file
	testFile := createTestFile(t, tempDir, "large.jpg", 2048)
	err = cache.Put("key1", testFile, 2048)
	if err == nil {
		t.Errorf("Expected error when file size exceeds max cache size")
	}
}

func TestMediaCache_UpdateExisting(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewMediaCache(1024*1024, tempDir)
	if err != nil {
		t.Fatalf("Failed to create media cache: %v", err)
	}

	// Add a file
	testFile := createTestFile(t, tempDir, "test.jpg", 1024)
	cache.Put("key1", testFile, 1024)

	// Update the same key
	cache.Put("key1", testFile, 512)

	// Should still have only 1 file
	if cache.GetCount() != 1 {
		t.Errorf("Expected 1 file after update, got %d", cache.GetCount())
	}

	// Size should be updated
	if cache.GetSize() != 512 {
		t.Errorf("Expected size 512 after update, got %d", cache.GetSize())
	}
}

func TestGetMediaTypeFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"photo.jpg", "photo"},
		{"image.png", "photo"},
		{"video.mp4", "video"},
		{"audio.mp3", "audio"},
		{"voice.ogg", "voice"},
		{"file.pdf", "document"},
		{"unknown", "document"},
	}

	for _, tt := range tests {
		result := getMediaTypeFromPath(tt.path)
		if result != tt.expected {
			t.Errorf("getMediaTypeFromPath(%s) = %s, expected %s", tt.path, result, tt.expected)
		}
	}
}
