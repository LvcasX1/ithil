package media

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// createTestImage creates a simple test image for testing purposes.
func createTestImage(t *testing.T) (string, func()) {
	t.Helper()

	// Create a simple 100x100 gradient image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			// Create a gradient from black to white
			intensity := uint8((x + y) * 255 / 200)
			img.Set(x, y, color.RGBA{intensity, intensity, intensity, 255})
		}
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "test-image-*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Encode image
	if err := png.Encode(tmpFile, img); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to encode test image: %v", err)
	}
	tmpFile.Close()

	// Return path and cleanup function
	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}

// createColoredTestImage creates a colored test image.
func createColoredTestImage(t *testing.T) (string, func()) {
	t.Helper()

	// Create a 100x100 image with colored quarters
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			var c color.RGBA
			if x < 50 && y < 50 {
				c = color.RGBA{255, 0, 0, 255} // Red
			} else if x >= 50 && y < 50 {
				c = color.RGBA{0, 255, 0, 255} // Green
			} else if x < 50 && y >= 50 {
				c = color.RGBA{0, 0, 255, 255} // Blue
			} else {
				c = color.RGBA{255, 255, 0, 255} // Yellow
			}
			img.Set(x, y, c)
		}
	}

	tmpFile, err := os.CreateTemp("", "test-colored-*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if err := png.Encode(tmpFile, img); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to encode test image: %v", err)
	}
	tmpFile.Close()

	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}

func TestNewThumbnailGenerator(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		height   int
		protocol GraphicsProtocol
		wantW    int
		wantH    int
	}{
		{
			name:     "default dimensions",
			width:    0,
			height:   0,
			protocol: ProtocolASCII,
			wantW:    20,
			wantH:    10,
		},
		{
			name:     "custom dimensions",
			width:    30,
			height:   15,
			protocol: ProtocolKitty,
			wantW:    30,
			wantH:    15,
		},
		{
			name:     "partial defaults",
			width:    25,
			height:   0,
			protocol: ProtocolSixel,
			wantW:    25,
			wantH:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewThumbnailGenerator(tt.width, tt.height, tt.protocol)
			if gen == nil {
				t.Fatal("NewThumbnailGenerator returned nil")
			}

			w, h := gen.GetDimensions()
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("GetDimensions() = (%d, %d), want (%d, %d)", w, h, tt.wantW, tt.wantH)
			}

			if gen.GetProtocol() != tt.protocol {
				t.Errorf("GetProtocol() = %v, want %v", gen.GetProtocol(), tt.protocol)
			}
		})
	}
}

func TestThumbnailGenerator_GenerateThumbnail(t *testing.T) {
	testImage, cleanup := createTestImage(t)
	defer cleanup()

	tests := []struct {
		name        string
		protocol    GraphicsProtocol
		wantErr     bool
		skipInCI    bool // Skip tests that require terminal color detection
	}{
		{
			name:     "ASCII protocol",
			protocol: ProtocolASCII,
			wantErr:  false,
			skipInCI: true, // ASCII requires terminal color detection
		},
		{
			name:     "Unicode Mosaic protocol",
			protocol: ProtocolUnicodeMosaic,
			wantErr:  false,
		},
		{
			name:     "Kitty protocol",
			protocol: ProtocolKitty,
			wantErr:  false,
		},
		{
			name:     "Sixel protocol",
			protocol: ProtocolSixel,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewThumbnailGenerator(20, 10, tt.protocol)
			thumbnail, err := gen.GenerateThumbnail(testImage)

			// Skip ASCII tests in CI (no terminal color support)
			if tt.skipInCI && err != nil {
				t.Skipf("Skipping %s test (no terminal color support): %v", tt.name, err)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateThumbnail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && thumbnail == "" {
				t.Error("GenerateThumbnail() returned empty string")
			}
		})
	}
}

func TestThumbnailGenerator_GenerateThumbnail_InvalidFile(t *testing.T) {
	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "nonexistent file",
			path:    "/path/to/nonexistent/file.png",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.GenerateThumbnail(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateThumbnail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThumbnailGenerator_Cache(t *testing.T) {
	testImage, cleanup := createTestImage(t)
	defer cleanup()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	// First generation
	thumbnail1, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		t.Fatalf("GenerateThumbnail() error = %v", err)
	}

	// Second generation (should come from cache)
	thumbnail2, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		t.Fatalf("GenerateThumbnail() error = %v", err)
	}

	// Should be identical (from cache)
	if thumbnail1 != thumbnail2 {
		t.Error("Cached thumbnail differs from original")
	}

	// Check cache size
	if gen.GetCacheSize() != 1 {
		t.Errorf("GetCacheSize() = %d, want 1", gen.GetCacheSize())
	}

	// Clear cache
	gen.ClearCache()
	if gen.GetCacheSize() != 0 {
		t.Errorf("GetCacheSize() after clear = %d, want 0", gen.GetCacheSize())
	}
}

func TestThumbnailGenerator_RemoveFromCache(t *testing.T) {
	testImage, cleanup := createTestImage(t)
	defer cleanup()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	// Generate thumbnail (will be cached)
	_, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		t.Fatalf("GenerateThumbnail() error = %v", err)
	}

	if gen.GetCacheSize() != 1 {
		t.Errorf("GetCacheSize() = %d, want 1", gen.GetCacheSize())
	}

	// Remove from cache
	gen.RemoveFromCache(testImage)

	if gen.GetCacheSize() != 0 {
		t.Errorf("GetCacheSize() after remove = %d, want 0", gen.GetCacheSize())
	}
}

func TestThumbnailGenerator_SetDimensions(t *testing.T) {
	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	// Add something to cache
	testImage, cleanup := createTestImage(t)
	defer cleanup()
	_, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		t.Fatalf("GenerateThumbnail() error = %v", err)
	}

	// Change dimensions (should clear cache)
	gen.SetDimensions(30, 15)

	w, h := gen.GetDimensions()
	if w != 30 || h != 15 {
		t.Errorf("GetDimensions() = (%d, %d), want (30, 15)", w, h)
	}

	// Cache should be cleared
	if gen.GetCacheSize() != 0 {
		t.Errorf("Cache was not cleared after SetDimensions()")
	}

	// Invalid dimensions should be ignored
	gen.SetDimensions(0, 0)
	w, h = gen.GetDimensions()
	if w != 30 || h != 15 {
		t.Errorf("Invalid dimensions changed the generator state")
	}
}

func TestThumbnailGenerator_SetProtocol(t *testing.T) {
	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	// Add something to cache
	testImage, cleanup := createTestImage(t)
	defer cleanup()
	_, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		t.Fatalf("GenerateThumbnail() error = %v", err)
	}

	// Change protocol (should clear cache)
	gen.SetProtocol(ProtocolKitty)

	if gen.GetProtocol() != ProtocolKitty {
		t.Errorf("GetProtocol() = %v, want %v", gen.GetProtocol(), ProtocolKitty)
	}

	// Cache should be cleared
	if gen.GetCacheSize() != 0 {
		t.Errorf("Cache was not cleared after SetProtocol()")
	}
}

func TestThumbnailGenerator_SetColored(t *testing.T) {
	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	// Add something to cache
	testImage, cleanup := createTestImage(t)
	defer cleanup()
	_, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		t.Fatalf("GenerateThumbnail() error = %v", err)
	}

	// Change colored setting (should clear cache)
	gen.SetColored(false)

	// Cache should be cleared
	if gen.GetCacheSize() != 0 {
		t.Errorf("Cache was not cleared after SetColored()")
	}
}

func TestThumbnailGenerator_GenerateThumbnailAsync(t *testing.T) {
	testImage, cleanup := createTestImage(t)
	defer cleanup()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	var wg sync.WaitGroup
	wg.Add(1)

	var thumbnail string
	var err error

	gen.GenerateThumbnailAsync(testImage, func(thumb string, e error) {
		thumbnail = thumb
		err = e
		wg.Done()
	})

	// Wait for async operation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Async thumbnail generation timed out")
	}

	if err != nil {
		t.Errorf("GenerateThumbnailAsync() error = %v", err)
	}

	if thumbnail == "" {
		t.Error("GenerateThumbnailAsync() returned empty thumbnail")
	}
}

func TestThumbnailGenerator_PreloadThumbnails(t *testing.T) {
	// Create multiple test images
	img1, cleanup1 := createTestImage(t)
	defer cleanup1()
	img2, cleanup2 := createColoredTestImage(t)
	defer cleanup2()
	img3, cleanup3 := createTestImage(t)
	defer cleanup3()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	imagePaths := []string{img1, img2, img3}

	var mu sync.Mutex
	results := make(map[string]bool)

	var wg sync.WaitGroup
	wg.Add(len(imagePaths))

	gen.PreloadThumbnails(imagePaths, func(path, thumbnail string, err error) {
		mu.Lock()
		defer mu.Unlock()
		results[path] = err == nil && thumbnail != ""
		wg.Done()
	})

	// Wait for all preloads
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("PreloadThumbnails timed out")
	}

	// Check all images were processed
	mu.Lock()
	defer mu.Unlock()
	for _, path := range imagePaths {
		if !results[path] {
			t.Errorf("Failed to preload thumbnail for %s", filepath.Base(path))
		}
	}

	// All should be in cache
	if gen.GetCacheSize() != len(imagePaths) {
		t.Errorf("GetCacheSize() = %d, want %d", gen.GetCacheSize(), len(imagePaths))
	}
}

func TestThumbnailGenerator_ValidateImageFile(t *testing.T) {
	testImage, cleanup := createTestImage(t)
	defer cleanup()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	tests := []struct {
		name    string
		path    string
		wantOk  bool
		wantErr bool
	}{
		{
			name:    "valid PNG",
			path:    testImage,
			wantOk:  true,
			wantErr: false,
		},
		{
			name:    "nonexistent file",
			path:    "/nonexistent/file.png",
			wantOk:  false,
			wantErr: true,
		},
		{
			name:    "unsupported extension",
			path:    "/tmp/test.xyz",
			wantOk:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := gen.ValidateImageFile(tt.path)
			if ok != tt.wantOk {
				t.Errorf("ValidateImageFile() ok = %v, want %v", ok, tt.wantOk)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewThumbnailGeneratorWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *ThumbnailGeneratorOptions
		want struct {
			width  int
			height int
		}
	}{
		{
			name: "nil options uses defaults",
			opts: nil,
			want: struct {
				width  int
				height int
			}{20, 10},
		},
		{
			name: "custom options",
			opts: &ThumbnailGeneratorOptions{
				Width:     25,
				Height:    12,
				Protocol:  ProtocolASCII,
				Colored:   false,
				CacheSize: 200,
			},
			want: struct {
				width  int
				height int
			}{25, 12},
		},
		{
			name: "auto-detect protocol",
			opts: &ThumbnailGeneratorOptions{
				Width:      30,
				Height:     15,
				AutoDetect: true,
				Colored:    true,
			},
			want: struct {
				width  int
				height int
			}{30, 15},
		},
		{
			name: "zero values use defaults",
			opts: &ThumbnailGeneratorOptions{
				Width:     0,
				Height:    0,
				CacheSize: 0,
			},
			want: struct {
				width  int
				height int
			}{20, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewThumbnailGeneratorWithOptions(tt.opts)
			if gen == nil {
				t.Fatal("NewThumbnailGeneratorWithOptions returned nil")
			}

			w, h := gen.GetDimensions()
			if w != tt.want.width || h != tt.want.height {
				t.Errorf("GetDimensions() = (%d, %d), want (%d, %d)",
					w, h, tt.want.width, tt.want.height)
			}
		})
	}
}

func TestThumbnailGenerator_GenerateThumbnailFromImage(t *testing.T) {
	// Create a simple test image in memory
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	tests := []struct {
		name     string
		protocol GraphicsProtocol
		wantErr  bool
		skipInCI bool // Skip tests that require terminal color detection
	}{
		{
			name:     "ASCII from memory",
			protocol: ProtocolASCII,
			wantErr:  false,
			skipInCI: true, // ASCII requires terminal color detection
		},
		{
			name:     "Mosaic from memory",
			protocol: ProtocolUnicodeMosaic,
			wantErr:  false,
		},
		{
			name:     "Kitty from memory",
			protocol: ProtocolKitty,
			wantErr:  false,
		},
		{
			name:     "Sixel from memory",
			protocol: ProtocolSixel,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewThumbnailGenerator(20, 10, tt.protocol)
			thumbnail, err := gen.GenerateThumbnailFromImage(img)

			// Skip ASCII tests in CI (no terminal color support)
			if tt.skipInCI && err != nil {
				t.Skipf("Skipping %s test (no terminal color support): %v", tt.name, err)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateThumbnailFromImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && thumbnail == "" {
				t.Error("GenerateThumbnailFromImage() returned empty string")
			}
		})
	}
}

func TestThumbnailCache_Concurrent(t *testing.T) {
	cache := NewThumbnailCache(100)

	// Test concurrent access
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := filepath.Join("/tmp", "test", "image", string(rune(id))+".png")
			value := "thumbnail_" + string(rune(id))

			// Set
			cache.Set(key, value)

			// Get
			if got, ok := cache.Get(key); ok {
				if got != value {
					t.Errorf("Cache returned wrong value: got %s, want %s", got, value)
				}
			}

			// Remove
			cache.Remove(key)
		}(i)
	}

	wg.Wait()
}

// Benchmark thumbnail generation
func BenchmarkThumbnailGenerator_Mosaic(b *testing.B) {
	testImage, cleanup := createTestImage(&testing.T{})
	defer cleanup()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)
	gen.ClearCache() // Ensure no caching for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.RemoveFromCache(testImage) // Force regeneration
		_, err := gen.GenerateThumbnail(testImage)
		if err != nil {
			b.Fatalf("GenerateThumbnail() error = %v", err)
		}
	}
}

func BenchmarkThumbnailGenerator_Cached(b *testing.B) {
	testImage, cleanup := createTestImage(&testing.T{})
	defer cleanup()

	gen := NewThumbnailGenerator(20, 10, ProtocolUnicodeMosaic)

	// Prime the cache
	_, err := gen.GenerateThumbnail(testImage)
	if err != nil {
		b.Fatalf("GenerateThumbnail() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateThumbnail(testImage)
		if err != nil {
			b.Fatalf("GenerateThumbnail() error = %v", err)
		}
	}
}
