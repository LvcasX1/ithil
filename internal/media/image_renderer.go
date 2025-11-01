// Package media provides media rendering utilities for terminal display.
package media

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
)

// ImageRenderer handles rendering images as ASCII/ANSI art in the terminal.
type ImageRenderer struct {
	maxWidth  int
	maxHeight int
	colored   bool
	braille   bool
}

// NewImageRenderer creates a new image renderer.
func NewImageRenderer(maxWidth, maxHeight int, colored bool) *ImageRenderer {
	return &ImageRenderer{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		colored:   colored,
		braille:   false, // Braille can be enabled for higher resolution
	}
}

// RenderImageFile renders an image file to ASCII/ANSI art.
func (r *ImageRenderer) RenderImageFile(filePath string) (string, error) {
	// Create flags for the ascii-image-converter
	flags := aic_package.DefaultFlags()

	// Set dimensions
	flags.Dimensions = []int{r.maxWidth, r.maxHeight}
	flags.Width = r.maxWidth
	flags.Height = r.maxHeight

	// Enable/disable color
	flags.Colored = r.colored

	// Use braille for higher resolution if enabled
	flags.Braille = r.braille

	// Don't save to file, just return the string
	flags.SaveTxtPath = ""
	flags.SaveImagePath = ""

	// Disable complex features for better terminal compatibility
	flags.Complex = false
	flags.Full = false

	// Convert image to ASCII (library expects file path)
	asciiArt, err := aic_package.Convert(filePath, flags)
	if err != nil {
		return "", fmt.Errorf("failed to convert image to ASCII: %w", err)
	}

	return asciiArt, nil
}

// RenderImage renders an image.Image to ASCII/ANSI art.
// Note: This is a convenience method that saves the image to a temp file first.
func (r *ImageRenderer) RenderImage(img image.Image) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "ithil-image-*.png")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Encode the image as PNG
	if err := png.Encode(tmpFile, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Now convert the file
	return r.RenderImageFile(tmpFile.Name())
}

// RenderThumbnail renders a small thumbnail preview of an image.
func (r *ImageRenderer) RenderThumbnail(filePath string, width, height int) (string, error) {
	// Create a temporary renderer with thumbnail dimensions
	thumbRenderer := &ImageRenderer{
		maxWidth:  width,
		maxHeight: height,
		colored:   r.colored,
		braille:   r.braille,
	}

	return thumbRenderer.RenderImageFile(filePath)
}

// SetBraille enables or disables braille mode for higher resolution.
func (r *ImageRenderer) SetBraille(enabled bool) {
	r.braille = enabled
}

// SetColored enables or disables colored output.
func (r *ImageRenderer) SetColored(enabled bool) {
	r.colored = enabled
}

// GetDimensions returns the current max dimensions.
func (r *ImageRenderer) GetDimensions() (width, height int) {
	return r.maxWidth, r.maxHeight
}

// SetDimensions sets the max dimensions for rendering.
func (r *ImageRenderer) SetDimensions(width, height int) {
	r.maxWidth = width
	r.maxHeight = height
}
