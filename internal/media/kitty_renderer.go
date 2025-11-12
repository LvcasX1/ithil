// Package media provides media rendering utilities for terminal display.
package media

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"

	"github.com/disintegration/imaging"
)

// KittyRenderer renders images using the Kitty graphics protocol.
// Kitty protocol provides high-fidelity image rendering directly in the terminal.
// Documentation: https://sw.kovidgoyal.net/kitty/graphics-protocol/
type KittyRenderer struct {
	maxWidth  int
	maxHeight int
}

// NewKittyRenderer creates a new Kitty protocol renderer.
func NewKittyRenderer(maxWidth, maxHeight int) *KittyRenderer {
	return &KittyRenderer{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

// RenderImageFile renders an image file using the Kitty graphics protocol.
func (r *KittyRenderer) RenderImageFile(filePath string) (string, error) {
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

// RenderImage renders an image.Image using the Kitty graphics protocol.
func (r *KittyRenderer) RenderImage(img interface{}) (string, error) {
	// Type assertion to image.Image
	imgData, ok := img.(image.Image)
	if !ok {
		return "", fmt.Errorf("invalid image type, expected image.Image")
	}

	// Resize image to fit terminal dimensions
	// Kitty uses pixel dimensions, we need to estimate character cell size
	// Typical terminal: 1 char = 8-10 pixels wide, 16-20 pixels tall
	// We'll use conservative estimates: 10px per char width, 20px per char height
	maxPixelWidth := r.maxWidth * 10
	maxPixelHeight := r.maxHeight * 20

	// Resize while maintaining aspect ratio
	imgData = imaging.Fit(imgData, maxPixelWidth, maxPixelHeight, imaging.Lanczos)

	// Encode image to PNG format (Kitty supports PNG directly)
	var buf bytes.Buffer
	if err := png.Encode(&buf, imgData); err != nil {
		return "", fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	// Base64 encode the PNG data
	encodedData := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Split into chunks (Kitty has a 4096 byte limit per transmission)
	const chunkSize = 4096
	chunks := splitIntoChunks(encodedData, chunkSize)

	// Build Kitty escape sequence
	var result string

	for i, chunk := range chunks {
		if i == 0 {
			// First chunk: transmit with metadata
			// Format: \x1b_Ga=T,f=100,t=d,<data>;\x1b\
			// a=T : transmit action
			// f=100 : PNG format
			// t=d : direct transmission (data in same escape code)
			// m=1 : more data to come (if multi-chunk)
			if len(chunks) > 1 {
				result += fmt.Sprintf("\x1b_Ga=T,f=100,t=d,m=1;%s\x1b\\", chunk)
			} else {
				result += fmt.Sprintf("\x1b_Ga=T,f=100,t=d;%s\x1b\\", chunk)
			}
		} else if i == len(chunks)-1 {
			// Last chunk: mark as final
			result += fmt.Sprintf("\x1b_Gm=0;%s\x1b\\", chunk)
		} else {
			// Middle chunks: continue transmission
			result += fmt.Sprintf("\x1b_Gm=1;%s\x1b\\", chunk)
		}
	}

	// Note: Kitty protocol handles image positioning automatically.
	// The Bubbletea TUI framework manages layout, so no manual newlines are needed.
	// Adding newlines here would break the viewport and shift the UI.

	return result, nil
}

// RenderImageWithID renders an image with a specific ID for later reference.
// This allows for image reuse and manipulation.
func (r *KittyRenderer) RenderImageWithID(filePath string, imageID uint32) (string, error) {
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

	// Resize image
	maxPixelWidth := r.maxWidth * 10
	maxPixelHeight := r.maxHeight * 20
	img = imaging.Fit(img, maxPixelWidth, maxPixelHeight, imaging.Lanczos)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image as PNG: %w", err)
	}

	// Base64 encode
	encodedData := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Transmit with ID
	// Format: \x1b_Ga=T,f=100,t=d,i=<id>,<data>;\x1b\
	result := fmt.Sprintf("\x1b_Ga=T,f=100,t=d,i=%d;%s\x1b\\", imageID, encodedData)

	// Note: Kitty protocol handles image positioning automatically.
	// The Bubbletea TUI framework manages layout, so no manual newlines are needed.

	return result, nil
}

// DeleteImage deletes an image by ID from the Kitty terminal.
func (r *KittyRenderer) DeleteImage(imageID uint32) string {
	// Format: \x1b_Ga=d,d=I,i=<id>;\x1b\
	// a=d : delete action
	// d=I : delete by ID
	// i=<id> : image ID
	return fmt.Sprintf("\x1b_Ga=d,d=I,i=%d;\x1b\\", imageID)
}

// SetColored is a no-op for Kitty renderer as it always renders in full color.
func (r *KittyRenderer) SetColored(enabled bool) {
	// Kitty always renders in full color, this is a no-op for interface compatibility
}

// GetDimensions returns the current max dimensions.
func (r *KittyRenderer) GetDimensions() (width, height int) {
	return r.maxWidth, r.maxHeight
}

// SetDimensions sets the max dimensions for rendering.
func (r *KittyRenderer) SetDimensions(width, height int) {
	r.maxWidth = width
	r.maxHeight = height
}

// splitIntoChunks splits a string into chunks of specified size.
func splitIntoChunks(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}

	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}

	return chunks
}

// SupportsKittyProtocol checks if the terminal supports Kitty graphics protocol.
// This is a utility function for testing and capability detection.
func SupportsKittyProtocol() bool {
	detector := NewProtocolDetector()
	return detector.isKittyTerminal()
}
