// Package media provides media rendering utilities for terminal display.
package media

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/disintegration/imaging"
)

// ThumbnailCache stores generated thumbnails in memory to avoid regenerating them.
type ThumbnailCache struct {
	mu     sync.RWMutex
	cache  map[string]string // filePath -> rendered thumbnail
	maxAge int               // maximum number of cached thumbnails
}

// NewThumbnailCache creates a new thumbnail cache.
func NewThumbnailCache(maxSize int) *ThumbnailCache {
	return &ThumbnailCache{
		cache:  make(map[string]string),
		maxAge: maxSize,
	}
}

// Get retrieves a cached thumbnail.
func (tc *ThumbnailCache) Get(filePath string) (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	thumbnail, exists := tc.cache[filePath]
	return thumbnail, exists
}

// Set stores a thumbnail in the cache.
func (tc *ThumbnailCache) Set(filePath, thumbnail string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Simple cache eviction: if we exceed maxAge, clear the cache
	if len(tc.cache) >= tc.maxAge {
		tc.cache = make(map[string]string)
	}

	tc.cache[filePath] = thumbnail
}

// Clear removes all cached thumbnails.
func (tc *ThumbnailCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.cache = make(map[string]string)
}

// Remove removes a specific thumbnail from the cache.
func (tc *ThumbnailCache) Remove(filePath string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	delete(tc.cache, filePath)
}

// Size returns the number of cached thumbnails.
func (tc *ThumbnailCache) Size() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.cache)
}

// ThumbnailGenerator generates small preview thumbnails for images.
// It is protocol-aware and uses the appropriate renderer based on terminal capabilities.
type ThumbnailGenerator struct {
	mu       sync.RWMutex
	width    int
	height   int
	protocol GraphicsProtocol
	colored  bool
	cache    *ThumbnailCache
}

// NewThumbnailGenerator creates a new thumbnail generator.
// Default dimensions are 20x10 characters, which provides a good balance
// between preview quality and space in the message list.
//
// Parameters:
//   - width: thumbnail width in characters (0 uses default: 20)
//   - height: thumbnail height in characters (0 uses default: 10)
//   - protocol: graphics protocol to use for rendering
//
// Example:
//
//	detector := media.NewProtocolDetector()
//	protocol := detector.DetectProtocol()
//	generator := media.NewThumbnailGenerator(20, 10, protocol)
func NewThumbnailGenerator(width, height int, protocol GraphicsProtocol) *ThumbnailGenerator {
	// Set default dimensions if not provided
	if width <= 0 {
		width = 20
	}
	if height <= 0 {
		height = 10
	}

	return &ThumbnailGenerator{
		width:    width,
		height:   height,
		protocol: protocol,
		colored:  true, // Enable color by default for better previews
		cache:    NewThumbnailCache(100),
	}
}

// GenerateThumbnail generates a thumbnail for the given image file.
// Returns the rendered thumbnail as a string ready for terminal display.
//
// The thumbnail is cached in memory, so subsequent calls with the same
// filePath will return the cached result without re-rendering.
//
// Parameters:
//   - imagePath: absolute path to the image file
//
// Returns:
//   - rendered thumbnail string (with ANSI/escape codes)
//   - error if the image cannot be loaded or rendered
//
// Example:
//
//	thumbnail, err := generator.GenerateThumbnail("/path/to/image.jpg")
//	if err != nil {
//	    log.Printf("Failed to generate thumbnail: %v", err)
//	    return
//	}
//	fmt.Print(thumbnail)
func (tg *ThumbnailGenerator) GenerateThumbnail(imagePath string) (string, error) {
	// Check cache first
	if cached, found := tg.cache.Get(imagePath); found {
		return cached, nil
	}

	// Validate file exists
	if _, err := os.Stat(imagePath); err != nil {
		return "", fmt.Errorf("image file not found: %w", err)
	}

	// Get current settings (thread-safe)
	tg.mu.RLock()
	width := tg.width
	height := tg.height
	protocol := tg.protocol
	colored := tg.colored
	tg.mu.RUnlock()

	// Get the appropriate renderer
	renderer := tg.getRenderer(protocol, width, height, colored)

	// Render the thumbnail
	thumbnail, err := renderer.RenderImageFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to render thumbnail: %w", err)
	}

	// Cache the result
	tg.cache.Set(imagePath, thumbnail)

	return thumbnail, nil
}

// GenerateThumbnailAsync generates a thumbnail asynchronously.
// The callback function is called with the result when generation completes.
// This is useful for non-blocking UI updates.
//
// Parameters:
//   - imagePath: absolute path to the image file
//   - callback: function to call with (thumbnail, error) when complete
//
// Example:
//
//	generator.GenerateThumbnailAsync("/path/to/image.jpg", func(thumbnail string, err error) {
//	    if err != nil {
//	        log.Printf("Thumbnail generation failed: %v", err)
//	        return
//	    }
//	    // Update UI with thumbnail
//	    updateMessageList(thumbnail)
//	})
func (tg *ThumbnailGenerator) GenerateThumbnailAsync(imagePath string, callback func(string, error)) {
	go func() {
		thumbnail, err := tg.GenerateThumbnail(imagePath)
		callback(thumbnail, err)
	}()
}

// GenerateThumbnailFromImage generates a thumbnail from an image.Image.
// This is useful when you already have a decoded image in memory.
//
// Note: Unlike GenerateThumbnail, this method does not use caching since
// there's no file path as a cache key. If you need caching, save the image
// to a temporary file and use GenerateThumbnail instead.
//
// Parameters:
//   - img: decoded image.Image
//
// Returns:
//   - rendered thumbnail string
//   - error if rendering fails
func (tg *ThumbnailGenerator) GenerateThumbnailFromImage(img image.Image) (string, error) {
	// Get current settings (thread-safe)
	tg.mu.RLock()
	width := tg.width
	height := tg.height
	protocol := tg.protocol
	colored := tg.colored
	tg.mu.RUnlock()

	// Resize the image to thumbnail dimensions
	resized := tg.resizeImage(img, width, height, protocol)

	// Render based on protocol
	switch protocol {
	case ProtocolKitty:
		renderer := NewKittyRenderer(width, height)
		return renderer.RenderImage(resized)

	case ProtocolSixel:
		renderer := NewSixelRenderer(width, height)
		return renderer.RenderImage(resized)

	case ProtocolUnicodeMosaic:
		renderer := NewMosaicRenderer(width, height, colored)
		return renderer.RenderImage(resized)

	case ProtocolASCII:
		renderer := NewImageRenderer(width, height, colored)
		return renderer.RenderImage(resized)

	default:
		return "", fmt.Errorf("unsupported graphics protocol: %v", protocol)
	}
}

// SetProtocol changes the graphics protocol used for rendering.
// This clears the thumbnail cache since cached thumbnails were rendered
// with the previous protocol.
//
// Parameters:
//   - protocol: new graphics protocol to use
//
// Example:
//
//	// Switch to ASCII mode for compatibility
//	generator.SetProtocol(media.ProtocolASCII)
func (tg *ThumbnailGenerator) SetProtocol(protocol GraphicsProtocol) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	if tg.protocol != protocol {
		tg.protocol = protocol
		// Clear cache since protocol changed
		tg.cache.Clear()
	}
}

// SetDimensions changes the thumbnail dimensions.
// This clears the thumbnail cache since cached thumbnails have different dimensions.
//
// Parameters:
//   - width: new thumbnail width in characters (must be > 0)
//   - height: new thumbnail height in characters (must be > 0)
//
// Example:
//
//	// Make thumbnails smaller to fit more messages on screen
//	generator.SetDimensions(15, 8)
func (tg *ThumbnailGenerator) SetDimensions(width, height int) {
	if width <= 0 || height <= 0 {
		return // Ignore invalid dimensions
	}

	tg.mu.Lock()
	defer tg.mu.Unlock()

	if tg.width != width || tg.height != height {
		tg.width = width
		tg.height = height
		// Clear cache since dimensions changed
		tg.cache.Clear()
	}
}

// SetColored enables or disables colored output for thumbnails.
// Only affects ASCII and Unicode Mosaic protocols (Kitty and Sixel are always colored).
//
// Parameters:
//   - enabled: true to enable color, false for grayscale
func (tg *ThumbnailGenerator) SetColored(enabled bool) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	if tg.colored != enabled {
		tg.colored = enabled
		// Clear cache if color setting changed
		tg.cache.Clear()
	}
}

// GetDimensions returns the current thumbnail dimensions.
//
// Returns:
//   - width: thumbnail width in characters
//   - height: thumbnail height in characters
func (tg *ThumbnailGenerator) GetDimensions() (width, height int) {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.width, tg.height
}

// GetProtocol returns the current graphics protocol being used.
func (tg *ThumbnailGenerator) GetProtocol() GraphicsProtocol {
	tg.mu.RLock()
	defer tg.mu.RUnlock()
	return tg.protocol
}

// ClearCache removes all cached thumbnails from memory.
// Useful if you need to force regeneration of thumbnails (e.g., after changing settings).
func (tg *ThumbnailGenerator) ClearCache() {
	tg.cache.Clear()
}

// RemoveFromCache removes a specific thumbnail from the cache.
// Useful when a file has been modified and needs to be re-rendered.
//
// Parameters:
//   - imagePath: path to the image whose thumbnail should be removed
func (tg *ThumbnailGenerator) RemoveFromCache(imagePath string) {
	tg.cache.Remove(imagePath)
}

// GetCacheSize returns the number of thumbnails currently cached in memory.
func (tg *ThumbnailGenerator) GetCacheSize() int {
	return tg.cache.Size()
}

// PreloadThumbnails preloads thumbnails for multiple images asynchronously.
// This is useful for preloading thumbnails for messages that are about to be displayed.
// The callback is called for each image with its result.
//
// Parameters:
//   - imagePaths: slice of image file paths to preload
//   - callback: function called for each image with (path, thumbnail, error)
//
// Example:
//
//	paths := []string{"/path/to/img1.jpg", "/path/to/img2.png"}
//	generator.PreloadThumbnails(paths, func(path, thumbnail string, err error) {
//	    if err != nil {
//	        log.Printf("Failed to load %s: %v", path, err)
//	        return
//	    }
//	    log.Printf("Preloaded thumbnail for %s", filepath.Base(path))
//	})
func (tg *ThumbnailGenerator) PreloadThumbnails(imagePaths []string, callback func(string, string, error)) {
	var wg sync.WaitGroup

	for _, path := range imagePaths {
		wg.Add(1)
		go func(imagePath string) {
			defer wg.Done()
			thumbnail, err := tg.GenerateThumbnail(imagePath)
			if callback != nil {
				callback(imagePath, thumbnail, err)
			}
		}(path)
	}

	// Optional: wait in background so this function returns immediately
	go wg.Wait()
}

// ValidateImageFile checks if a file is a valid image that can be rendered.
// Returns true if the file exists and is a supported image format.
//
// Supported formats: PNG, JPEG, GIF
//
// Parameters:
//   - imagePath: path to the image file to validate
//
// Returns:
//   - true if valid, false otherwise
//   - error describing why the file is invalid (nil if valid)
func (tg *ThumbnailGenerator) ValidateImageFile(imagePath string) (bool, error) {
	// Check if file exists
	if _, err := os.Stat(imagePath); err != nil {
		return false, fmt.Errorf("file does not exist: %w", err)
	}

	// Check file extension
	ext := filepath.Ext(imagePath)
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
	}

	if !supportedExts[ext] {
		return false, fmt.Errorf("unsupported image format: %s", ext)
	}

	// Try to decode the image to verify it's valid
	file, err := os.Open(imagePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return false, fmt.Errorf("invalid image file: %w", err)
	}

	return true, nil
}

// getRenderer returns the appropriate renderer for the given protocol.
func (tg *ThumbnailGenerator) getRenderer(protocol GraphicsProtocol, width, height int, colored bool) ImageRendererInterface {
	switch protocol {
	case ProtocolKitty:
		return NewKittyRenderer(width, height)
	case ProtocolSixel:
		return NewSixelRenderer(width, height)
	case ProtocolUnicodeMosaic:
		return NewMosaicRenderer(width, height, colored)
	case ProtocolASCII:
		return NewImageRenderer(width, height, colored)
	default:
		// Fallback to ASCII
		return NewImageRenderer(width, height, colored)
	}
}

// resizeImage resizes an image to fit thumbnail dimensions based on protocol.
// Different protocols have different pixel-to-character ratios.
func (tg *ThumbnailGenerator) resizeImage(img image.Image, width, height int, protocol GraphicsProtocol) image.Image {
	var maxPixelWidth, maxPixelHeight int

	switch protocol {
	case ProtocolKitty, ProtocolSixel:
		// Kitty and Sixel: ~10 pixels per char width, ~20 pixels per char height
		maxPixelWidth = width * 10
		maxPixelHeight = height * 20

	case ProtocolUnicodeMosaic:
		// Unicode mosaic: 1:2 aspect ratio (each char is 2 vertical pixels)
		maxPixelWidth = width
		maxPixelHeight = height * 2

	case ProtocolASCII:
		// ASCII: 1:1 character mapping
		maxPixelWidth = width
		maxPixelHeight = height

	default:
		// Fallback
		maxPixelWidth = width
		maxPixelHeight = height
	}

	// Resize while maintaining aspect ratio
	return imaging.Fit(img, maxPixelWidth, maxPixelHeight, imaging.Lanczos)
}

// ThumbnailGeneratorOptions provides configuration options for creating a thumbnail generator.
type ThumbnailGeneratorOptions struct {
	Width      int              // Thumbnail width in characters
	Height     int              // Thumbnail height in characters
	Protocol   GraphicsProtocol // Graphics protocol to use
	Colored    bool             // Enable colored output
	CacheSize  int              // Maximum number of cached thumbnails
	AutoDetect bool             // Auto-detect protocol if true (ignores Protocol field)
}

// NewThumbnailGeneratorWithOptions creates a thumbnail generator with custom options.
//
// Example:
//
//	opts := &ThumbnailGeneratorOptions{
//	    Width:      25,
//	    Height:     12,
//	    AutoDetect: true,
//	    Colored:    true,
//	    CacheSize:  200,
//	}
//	generator := NewThumbnailGeneratorWithOptions(opts)
func NewThumbnailGeneratorWithOptions(opts *ThumbnailGeneratorOptions) *ThumbnailGenerator {
	if opts == nil {
		opts = &ThumbnailGeneratorOptions{
			Width:      20,
			Height:     10,
			AutoDetect: true,
			Colored:    true,
			CacheSize:  100,
		}
	}

	// Auto-detect protocol if requested
	protocol := opts.Protocol
	if opts.AutoDetect {
		detector := NewProtocolDetector()
		protocol = detector.DetectProtocol()
	}

	// Use defaults for zero values
	width := opts.Width
	if width <= 0 {
		width = 20
	}
	height := opts.Height
	if height <= 0 {
		height = 10
	}
	cacheSize := opts.CacheSize
	if cacheSize <= 0 {
		cacheSize = 100
	}

	return &ThumbnailGenerator{
		width:    width,
		height:   height,
		protocol: protocol,
		colored:  opts.Colored,
		cache:    NewThumbnailCache(cacheSize),
	}
}
