// Package media provides media rendering utilities for terminal display.
package media

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/disintegration/imaging"
	"github.com/mattn/go-sixel"
)

// SixelRenderer renders images using the Sixel graphics protocol.
// Sixel is widely supported by terminals like XTerm, WezTerm, Alacritty, and more.
// It provides good quality image rendering with 256-color palette.
type SixelRenderer struct {
	maxWidth  int
	maxHeight int
}

// NewSixelRenderer creates a new Sixel protocol renderer.
func NewSixelRenderer(maxWidth, maxHeight int) *SixelRenderer {
	return &SixelRenderer{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

// RenderImageFile renders an image file using the Sixel graphics protocol.
func (r *SixelRenderer) RenderImageFile(filePath string) (string, error) {
	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}

	// Open and decode the image
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Render the image
	return r.RenderImage(img)
}

// RenderImage renders an image.Image using the Sixel graphics protocol.
func (r *SixelRenderer) RenderImage(img interface{}) (string, error) {
	// Type assertion to image.Image
	imgData, ok := img.(image.Image)
	if !ok {
		return "", fmt.Errorf("invalid image type, expected image.Image")
	}

	// Resize image to fit terminal dimensions
	// Sixel uses pixel dimensions, estimate character cell size
	// Typical terminal: 1 char = 8-10 pixels wide, 16-20 pixels tall
	maxPixelWidth := r.maxWidth * 10
	maxPixelHeight := r.maxHeight * 20

	// Get original image dimensions
	bounds := imgData.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Only scale DOWN if image is larger than available space, never scale UP
	// This preserves original quality for small images like thumbnails
	if originalWidth > maxPixelWidth || originalHeight > maxPixelHeight {
		imgData = imaging.Fit(imgData, maxPixelWidth, maxPixelHeight, imaging.Lanczos)
	}

	// Encode image to Sixel format
	var buf bytes.Buffer

	// Create Sixel encoder with options
	encoder := sixel.NewEncoder(&buf)

	// Set dithering for better quality (options: None, Atkinson, FloydSteinberg, etc.)
	encoder.Dither = true

	// Encode the image
	if err := encoder.Encode(imgData); err != nil {
		return "", fmt.Errorf("failed to encode image as Sixel: %w", err)
	}

	return buf.String(), nil
}

// RenderImageWithDithering renders an image with specific dithering algorithm.
func (r *SixelRenderer) RenderImageWithDithering(filePath string, dither bool) (string, error) {
	// Open and decode the image
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image - only scale DOWN, never UP
	maxPixelWidth := r.maxWidth * 10
	maxPixelHeight := r.maxHeight * 20

	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	if originalWidth > maxPixelWidth || originalHeight > maxPixelHeight {
		img = imaging.Fit(img, maxPixelWidth, maxPixelHeight, imaging.Lanczos)
	}

	// Encode with specified dithering
	var buf bytes.Buffer
	encoder := sixel.NewEncoder(&buf)
	encoder.Dither = dither

	if err := encoder.Encode(img); err != nil {
		return "", fmt.Errorf("failed to encode image as Sixel: %w", err)
	}

	return buf.String(), nil
}

// RenderThumbnail renders a small thumbnail using Sixel protocol.
func (r *SixelRenderer) RenderThumbnail(filePath string, width, height int) (string, error) {
	// Open and decode the image
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to thumbnail dimensions
	maxPixelWidth := width * 10
	maxPixelHeight := height * 20
	img = imaging.Fit(img, maxPixelWidth, maxPixelHeight, imaging.Lanczos)

	// Encode as Sixel
	var buf bytes.Buffer
	encoder := sixel.NewEncoder(&buf)
	encoder.Dither = true // Use dithering for better quality thumbnails

	if err := encoder.Encode(img); err != nil {
		return "", fmt.Errorf("failed to encode thumbnail as Sixel: %w", err)
	}

	return buf.String(), nil
}

// SetColored is a no-op for Sixel renderer as it always renders in color.
func (r *SixelRenderer) SetColored(enabled bool) {
	// Sixel always renders in color (256-color palette), this is a no-op for interface compatibility
}

// GetDimensions returns the current max dimensions.
func (r *SixelRenderer) GetDimensions() (width, height int) {
	return r.maxWidth, r.maxHeight
}

// SetDimensions sets the max dimensions for rendering.
func (r *SixelRenderer) SetDimensions(width, height int) {
	r.maxWidth = width
	r.maxHeight = height
}

// SupportsSixelProtocol checks if the terminal supports Sixel graphics protocol.
// This is a utility function for testing and capability detection.
func SupportsSixelProtocol() bool {
	detector := NewProtocolDetector()
	return detector.supportsSixel()
}

// GetOptimalSixelSettings returns optimal Sixel encoder settings based on image characteristics.
func GetOptimalSixelSettings(img image.Image) (dither bool, colors int) {
	// Analyze image to determine optimal settings
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixelCount := width * height

	// For large images, use dithering for better quality
	if pixelCount > 100000 {
		dither = true
	} else {
		dither = false
	}

	// Color count (Sixel supports up to 256 colors)
	colors = 256

	return dither, colors
}
