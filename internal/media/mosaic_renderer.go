// Package media provides media rendering utilities for terminal display.
package media

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/disintegration/imaging"
)

// MosaicRenderer renders images using Unicode half-block characters for better quality.
type MosaicRenderer struct {
	maxWidth  int
	maxHeight int
	colored   bool
}

// NewMosaicRenderer creates a new mosaic renderer.
func NewMosaicRenderer(maxWidth, maxHeight int, colored bool) *MosaicRenderer {
	return &MosaicRenderer{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		colored:   colored,
	}
}

// RenderImageFile renders an image file using Unicode half-blocks.
// This provides 2x vertical resolution compared to standard ASCII art.
func (r *MosaicRenderer) RenderImageFile(filePath string) (string, error) {
	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file, no such file or directory: %w", err)
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

// RenderImage renders an image.Image using Unicode half-blocks.
func (r *MosaicRenderer) RenderImage(img image.Image) (string, error) {
	// Each character represents 2 vertical pixels (top half, bottom half)
	// So we need height * 2 pixels for the target height
	targetWidth := r.maxWidth
	targetHeight := r.maxHeight * 2

	// Resize the image to fit the target dimensions while maintaining aspect ratio
	img = imaging.Fit(img, targetWidth, targetHeight, imaging.Lanczos)

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Build the output string
	var result string

	// Process two rows at a time (for top and bottom half of each character)
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x++ {
			// Get the top pixel color
			topPixel := img.At(x, y)
			topR, topG, topB, _ := topPixel.RGBA()
			topR >>= 8
			topG >>= 8
			topB >>= 8

			// Get the bottom pixel color (or use top if we're at the last row)
			var botR, botG, botB uint32
			if y+1 < height {
				botPixel := img.At(x, y+1)
				botR, botG, botB, _ = botPixel.RGBA()
				botR >>= 8
				botG >>= 8
				botB >>= 8
			} else {
				// If we're at the last row and it's odd, use the top pixel for both
				botR, botG, botB = topR, topG, topB
			}

			if r.colored {
				// Use ANSI escape codes for colored output
				// Upper half block (▀) shows the top pixel in foreground and bottom pixel in background
				result += fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m",
					topR, topG, topB, botR, botG, botB)
			} else {
				// Grayscale version
				topGray := (topR + topG + topB) / 3
				botGray := (botR + botG + botB) / 3

				// Choose the best character based on the pixel intensities
				var char string
				if topGray > 200 && botGray > 200 {
					char = " " // Both bright
				} else if topGray < 50 && botGray < 50 {
					char = "█" // Both dark
				} else if topGray > botGray {
					char = "▀" // Top bright, bottom dark
				} else {
					char = "▄" // Top dark, bottom bright
				}
				result += char
			}
		}
		result += "\n"
	}

	return result, nil
}

// SetColored enables or disables colored output.
func (r *MosaicRenderer) SetColored(enabled bool) {
	r.colored = enabled
}

// GetDimensions returns the current max dimensions.
func (r *MosaicRenderer) GetDimensions() (width, height int) {
	return r.maxWidth, r.maxHeight
}

// SetDimensions sets the max dimensions for rendering.
func (r *MosaicRenderer) SetDimensions(width, height int) {
	r.maxWidth = width
	r.maxHeight = height
}
