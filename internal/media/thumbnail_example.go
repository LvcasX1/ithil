package media

// This file provides usage examples for the ThumbnailGenerator.
// These examples demonstrate how to integrate thumbnail generation
// into the ithil TUI application.

import (
	"fmt"
	"log"
)

// Example1_BasicUsage demonstrates basic thumbnail generation.
func Example1_BasicUsage() {
	// Detect the best graphics protocol for the terminal
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	fmt.Printf("Detected protocol: %s\n", protocol)

	// Create a thumbnail generator with default dimensions (20x10)
	generator := NewThumbnailGenerator(0, 0, protocol)

	// Generate a thumbnail
	imagePath := "/path/to/image.jpg"
	thumbnail, err := generator.GenerateThumbnail(imagePath)
	if err != nil {
		log.Printf("Failed to generate thumbnail: %v", err)
		return
	}

	// Display the thumbnail in the terminal
	fmt.Print(thumbnail)
}

// Example2_CustomDimensions demonstrates using custom thumbnail dimensions.
func Example2_CustomDimensions() {
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	// Create a larger thumbnail (30x15 characters)
	generator := NewThumbnailGenerator(30, 15, protocol)

	imagePath := "/path/to/image.png"
	thumbnail, err := generator.GenerateThumbnail(imagePath)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Print(thumbnail)
}

// Example3_AsyncGeneration demonstrates non-blocking thumbnail generation.
// This is useful in a TUI to avoid freezing the UI while generating thumbnails.
func Example3_AsyncGeneration() {
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	generator := NewThumbnailGenerator(20, 10, protocol)

	imagePath := "/path/to/large-image.jpg"

	// Generate thumbnail asynchronously
	generator.GenerateThumbnailAsync(imagePath, func(thumbnail string, err error) {
		if err != nil {
			log.Printf("Async generation failed: %v", err)
			return
		}

		// In a real TUI, you would send a message to update the UI
		// For example, with Bubbletea:
		// return m, func() tea.Msg {
		//     return ThumbnailReadyMsg{
		//         MessageID: messageID,
		//         Thumbnail: thumbnail,
		//     }
		// }

		fmt.Println("Thumbnail ready!")
		fmt.Print(thumbnail)
	})

	// Continue with other work while thumbnail generates in background
	fmt.Println("Thumbnail generation started in background...")
}

// Example4_PreloadThumbnails demonstrates preloading thumbnails for multiple images.
// Useful when loading a chat with many media messages.
func Example4_PreloadThumbnails() {
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	generator := NewThumbnailGenerator(20, 10, protocol)

	// List of images to preload (e.g., from recent messages)
	imagePaths := []string{
		"/path/to/image1.jpg",
		"/path/to/image2.png",
		"/path/to/image3.gif",
	}

	// Preload all thumbnails
	generator.PreloadThumbnails(imagePaths, func(path, thumbnail string, err error) {
		if err != nil {
			log.Printf("Failed to preload %s: %v", path, err)
			return
		}

		// In a real TUI, update the UI with the loaded thumbnail
		log.Printf("Preloaded thumbnail for %s", path)

		// The thumbnail is now cached and will be instant on next access
	})

	// Later, when rendering messages, thumbnails will be instant from cache
	for _, path := range imagePaths {
		thumbnail, err := generator.GenerateThumbnail(path)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Print(thumbnail)
	}
}

// Example5_DynamicProtocolSwitching demonstrates changing protocol at runtime.
// Useful if user changes terminal or settings.
func Example5_DynamicProtocolSwitching() {
	// Start with auto-detected protocol
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	generator := NewThumbnailGenerator(20, 10, protocol)

	imagePath := "/path/to/image.jpg"

	// Generate with detected protocol
	thumbnail1, _ := generator.GenerateThumbnail(imagePath)
	fmt.Println("Thumbnail with auto-detected protocol:")
	fmt.Print(thumbnail1)

	// User requests ASCII-only mode (for compatibility or piping output)
	generator.SetProtocol(ProtocolASCII)
	generator.SetColored(false)

	// Generate again with ASCII protocol
	// Note: cache was cleared when protocol changed
	thumbnail2, _ := generator.GenerateThumbnail(imagePath)
	fmt.Println("\nThumbnail with ASCII protocol:")
	fmt.Print(thumbnail2)
}

// Example6_ValidationBeforeGeneration demonstrates validating images before generation.
func Example6_ValidationBeforeGeneration() {
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	generator := NewThumbnailGenerator(20, 10, protocol)

	imagePath := "/path/to/image.jpg"

	// Validate the image file first
	valid, err := generator.ValidateImageFile(imagePath)
	if !valid {
		log.Printf("Invalid image file: %v", err)
		return
	}

	// Only generate if valid
	thumbnail, err := generator.GenerateThumbnail(imagePath)
	if err != nil {
		log.Printf("Generation failed: %v", err)
		return
	}

	fmt.Print(thumbnail)
}

// Example7_UsingOptions demonstrates creating generator with options struct.
func Example7_UsingOptions() {
	// Create generator with custom configuration
	opts := &ThumbnailGeneratorOptions{
		Width:      25,           // Slightly larger thumbnails
		Height:     12,           // Slightly taller
		AutoDetect: true,         // Auto-detect protocol
		Colored:    true,         // Enable color
		CacheSize:  200,          // Cache up to 200 thumbnails
	}

	generator := NewThumbnailGeneratorWithOptions(opts)

	fmt.Printf("Generator created with %dx%d dimensions\n",
		opts.Width, opts.Height)
	fmt.Printf("Using protocol: %s\n", generator.GetProtocol())
	fmt.Printf("Cache size limit: %d\n", opts.CacheSize)

	imagePath := "/path/to/image.jpg"
	thumbnail, err := generator.GenerateThumbnail(imagePath)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Print(thumbnail)
}

// Example8_IntegrationWithBubbletea demonstrates how to integrate with a Bubbletea model.
// This shows a pattern for ithil's conversation view.
//
// Note: This is pseudocode showing the integration pattern.
func Example8_IntegrationWithBubbletea() {
	// In your model initialization:
	/*
		type ConversationModel struct {
			// ... other fields
			thumbnailGen *media.ThumbnailGenerator
		}

		func NewConversationModel() ConversationModel {
			detector := media.NewProtocolDetector()
			protocol := detector.DetectProtocol()

			return ConversationModel{
				thumbnailGen: media.NewThumbnailGenerator(20, 10, protocol),
			}
		}
	*/

	// When receiving a message with media:
	/*
		type MediaMessage struct {
			ID        int64
			FilePath  string
			Thumbnail string // Store the rendered thumbnail
		}

		// In your Update() function when a new media message arrives:
		case NewMessageMsg:
			if msg.HasMedia {
				// Generate thumbnail asynchronously
				m.thumbnailGen.GenerateThumbnailAsync(msg.FilePath,
					func(thumbnail string, err error) {
						if err != nil {
							// Handle error
							return
						}

						// Send message to update UI with thumbnail
						// (You'd need to capture necessary IDs in closure)
						p.Send(ThumbnailReadyMsg{
							MessageID: msg.ID,
							Thumbnail: thumbnail,
						})
					})
			}
			return m, nil

		case ThumbnailReadyMsg:
			// Update the message with the thumbnail
			for i := range m.messages {
				if m.messages[i].ID == msg.MessageID {
					m.messages[i].Thumbnail = msg.Thumbnail
					break
				}
			}
			return m, nil
	*/

	// When rendering messages in View():
	/*
		func (m ConversationModel) View() string {
			var b strings.Builder

			for _, msg := range m.messages {
				if msg.HasMedia {
					if msg.Thumbnail != "" {
						// Thumbnail is ready, display it
						b.WriteString(msg.Thumbnail)
					} else {
						// Thumbnail still loading, show placeholder
						b.WriteString("[Loading thumbnail...]")
					}
				} else {
					// Regular text message
					b.WriteString(msg.Text)
				}
				b.WriteString("\n")
			}

			return b.String()
		}
	*/
}

// Example9_CacheManagement demonstrates cache management best practices.
func Example9_CacheManagement() {
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	generator := NewThumbnailGenerator(20, 10, protocol)

	// Generate some thumbnails
	images := []string{
		"/path/to/image1.jpg",
		"/path/to/image2.png",
		"/path/to/image3.gif",
	}

	for _, path := range images {
		_, err := generator.GenerateThumbnail(path)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}

	fmt.Printf("Cache size: %d thumbnails\n", generator.GetCacheSize())

	// If a file was modified, remove it from cache to force regeneration
	modifiedFile := "/path/to/image1.jpg"
	generator.RemoveFromCache(modifiedFile)
	fmt.Printf("Removed %s from cache\n", modifiedFile)

	// Regenerate (won't use cache)
	_, err := generator.GenerateThumbnail(modifiedFile)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Clear entire cache (e.g., when changing chats)
	generator.ClearCache()
	fmt.Println("Cache cleared")
}

// Example10_ErrorHandling demonstrates proper error handling patterns.
func Example10_ErrorHandling() {
	detector := NewProtocolDetector()
	protocol := detector.DetectProtocol()

	generator := NewThumbnailGenerator(20, 10, protocol)

	imagePath := "/path/to/image.jpg"

	// Validate before generating
	if valid, err := generator.ValidateImageFile(imagePath); !valid {
		switch {
		case err.Error() == "file does not exist":
			log.Println("Image file not downloaded yet")
			// Trigger download
		case err.Error() == "unsupported image format":
			log.Println("Cannot display this image format")
			// Show unsupported format message
		default:
			log.Printf("Image validation failed: %v", err)
		}
		return
	}

	// Generate thumbnail with error handling
	thumbnail, err := generator.GenerateThumbnail(imagePath)
	if err != nil {
		log.Printf("Thumbnail generation failed: %v", err)

		// Fallback: show a simple text placeholder
		fmt.Println("[Image]")
		return
	}

	// Success - display thumbnail
	fmt.Print(thumbnail)
}
