// Package telegram provides media upload and download functionality.
package telegram

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/lvcasx1/ithil/pkg/types"
)

// MediaManager handles media upload and download operations.
type MediaManager struct {
	client     *Client
	downloader *downloader.Downloader
	uploader   *uploader.Uploader
	mediaDir   string
}

// NewMediaManager creates a new media manager.
func (c *Client) NewMediaManager(mediaDir string) (*MediaManager, error) {
	// Ensure media directory exists
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create media directory: %w", err)
	}

	return &MediaManager{
		client:     c,
		downloader: downloader.NewDownloader(),
		uploader:   uploader.NewUploader(c.api),
		mediaDir:   mediaDir,
	}, nil
}

// GetMediaDirectory returns the path to the media directory.
func (m *MediaManager) GetMediaDirectory() string {
	return m.mediaDir
}

// getPhotoSizePriority returns the priority for a given photo size type.
// Higher priority means better quality/resolution.
// Telegram photo size types:
// - "w" (2560x2560) - highest quality
// - "y" (1280x1280) - high quality
// - "x" (800x800)   - medium-high quality
// - "m" (320x320)   - medium quality
// - "s" (100x100)   - thumbnail quality
func getPhotoSizePriority(sizeType string) int {
	priorities := map[string]int{
		"w": 5, // 2560x2560 - highest
		"y": 4, // 1280x1280
		"x": 3, // 800x800
		"m": 2, // 320x320
		"s": 1, // 100x100 - thumbnail
	}

	if priority, ok := priorities[sizeType]; ok {
		return priority
	}
	return 0 // Unknown types get lowest priority
}

// selectBestPhotoSize selects the best photo size based on priority and byte size.
// Priority order: w > y > x > m > s
// Within the same priority level, larger byte size is preferred.
func selectBestPhotoSize(sizes []tg.PhotoSizeClass) *tg.PhotoSize {
	var bestSize *tg.PhotoSize
	bestPriority := -1

	for _, size := range sizes {
		photoSize, ok := size.(*tg.PhotoSize)
		if !ok {
			continue
		}

		priority := getPhotoSizePriority(photoSize.Type)

		// Select if:
		// 1. We don't have a best size yet, OR
		// 2. This size has higher priority, OR
		// 3. Same priority but larger byte size
		if bestSize == nil ||
			priority > bestPriority ||
			(priority == bestPriority && photoSize.Size > bestSize.Size) {
			bestSize = photoSize
			bestPriority = priority
		}
	}

	return bestSize
}

// DownloadPhoto downloads a photo from a message.
func (m *MediaManager) DownloadPhoto(ctx context.Context, photo *tg.Photo, chatID int64) (string, error) {
	if photo == nil {
		return "", fmt.Errorf("photo is nil")
	}

	m.client.logger.Debug("DownloadPhoto called",
		"photoID", photo.ID,
		"sizeCount", len(photo.Sizes))

	// Log all available sizes for debugging
	for _, size := range photo.Sizes {
		if photoSize, ok := size.(*tg.PhotoSize); ok {
			m.client.logger.Debug("Available photo size",
				"type", photoSize.Type,
				"priority", getPhotoSizePriority(photoSize.Type),
				"width", photoSize.W,
				"height", photoSize.H,
				"bytes", photoSize.Size)
		}
	}

	// Select the best photo size based on priority
	bestSize := selectBestPhotoSize(photo.Sizes)

	if bestSize == nil {
		m.client.logger.Error("No valid photo size found",
			"photoID", photo.ID,
			"totalSizes", len(photo.Sizes))
		return "", fmt.Errorf("no valid photo size found")
	}

	m.client.logger.Info("Selected best photo size for download",
		"type", bestSize.Type,
		"priority", getPhotoSizePriority(bestSize.Type),
		"width", bestSize.W,
		"height", bestSize.H,
		"bytes", bestSize.Size)

	// Create file location from photo
	location := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		ThumbSize:     bestSize.Type,
	}

	// Generate local file path
	fileName := fmt.Sprintf("photo_%d_%d.jpg", chatID, photo.ID)
	localPath := filepath.Join(m.mediaDir, fmt.Sprintf("%d", chatID), fileName)

	// Ensure chat directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create chat directory: %w", err)
	}

	// Download the file
	if err := m.downloadFile(ctx, location, localPath, int64(bestSize.Size)); err != nil {
		return "", fmt.Errorf("failed to download photo: %w", err)
	}

	fileInfo, err := os.Stat(localPath)
	if err == nil {
		m.client.logger.Info("Downloaded photo to file",
			"path", localPath,
			"actualFileSize", fileInfo.Size(),
			"expectedSize", bestSize.Size)
	} else {
		m.client.logger.Error("Failed to stat downloaded file",
			"path", localPath,
			"error", err)
	}

	return localPath, nil
}

// DownloadDocument downloads a document (video, audio, file, etc.) from a message.
func (m *MediaManager) DownloadDocument(ctx context.Context, doc *tg.Document, chatID int64) (string, error) {
	if doc == nil {
		return "", fmt.Errorf("document is nil")
	}

	// Get file extension from mime type or filename
	ext := ".bin"
	fileName := fmt.Sprintf("document_%d_%d", chatID, doc.ID)

	// Try to get filename from attributes
	for _, attr := range doc.Attributes {
		if fileAttr, ok := attr.(*tg.DocumentAttributeFilename); ok {
			fileName = fileAttr.FileName
			break
		}
	}

	// If no filename attribute, try to get extension from mime type
	if !strings.Contains(fileName, ".") {
		exts, err := mime.ExtensionsByType(doc.MimeType)
		if err == nil && len(exts) > 0 {
			ext = exts[0]
			fileName = fmt.Sprintf("%s%s", fileName, ext)
		}
	}

	// Create file location from document
	location := &tg.InputDocumentFileLocation{
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: doc.FileReference,
		ThumbSize:     "", // Empty for full document
	}

	// Generate local file path
	localPath := filepath.Join(m.mediaDir, fmt.Sprintf("%d", chatID), fileName)

	// Ensure chat directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create chat directory: %w", err)
	}

	// Download the file
	if err := m.downloadFile(ctx, location, localPath, doc.Size); err != nil {
		return "", fmt.Errorf("failed to download document: %w", err)
	}

	return localPath, nil
}

// downloadFile downloads a file from Telegram to local storage.
func (m *MediaManager) downloadFile(ctx context.Context, location tg.InputFileLocationClass, localPath string, size int64) error {
	// Create the local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Download the file
	_, err = m.downloader.Download(m.client.api, location).Stream(ctx, file)
	if err != nil {
		// Clean up partial download on error
		os.Remove(localPath)
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// UploadPhoto uploads a photo file to Telegram and sends it as a message.
func (m *MediaManager) UploadPhoto(ctx context.Context, chat *types.Chat, filePath string, caption string, replyToMessageID int64) (*types.Message, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Upload the file
	inputFile, err := m.uploader.FromReader(ctx, filepath.Base(filePath), file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create InputMedia for photo
	inputMedia := &tg.InputMediaUploadedPhoto{
		File: inputFile,
	}

	// Convert chat to InputPeer
	inputPeer, err := m.client.chatToInputPeer(chat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Build send media request
	request := &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    inputMedia,
		Message:  caption,
		RandomID: m.client.generateRandomID(),
	}

	// Add reply-to information if provided
	if replyToMessageID > 0 {
		request.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: int(replyToMessageID),
		}
	}

	// Send the media
	updates, err := m.client.api.MessagesSendMedia(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send media: %w", err)
	}

	// Extract the sent message from updates
	message, err := m.client.extractMessageFromUpdates(updates, chat.ID, caption)
	if err != nil {
		m.client.logger.Error("Failed to extract message from updates", "error", err)
		// Return a basic message even if extraction fails
		return &types.Message{
			ID:     0,
			ChatID: chat.ID,
			Content: types.MessageContent{
				Type:    types.MessageTypePhoto,
				Caption: caption,
				Media: &types.Media{
					Size: fileInfo.Size(),
				},
			},
			IsOutgoing: true,
		}, nil
	}

	m.client.logger.Info("Photo sent successfully", "messageID", message.ID)
	return message, nil
}

// UploadDocument uploads a document (video, audio, file, etc.) to Telegram and sends it as a message.
func (m *MediaManager) UploadDocument(ctx context.Context, chat *types.Chat, filePath string, caption string, replyToMessageID int64, mediaType types.MessageType) (*types.Message, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Upload the file
	inputFile, err := m.uploader.FromReader(ctx, filepath.Base(filePath), file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Detect mime type
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Create attributes based on media type
	var attributes []tg.DocumentAttributeClass
	attributes = append(attributes, &tg.DocumentAttributeFilename{
		FileName: filepath.Base(filePath),
	})

	// Add specific attributes based on type
	switch mediaType {
	case types.MessageTypeVideo:
		// For videos, we could add video attributes if we had dimension/duration info
		// attributes = append(attributes, &tg.DocumentAttributeVideo{...})
	case types.MessageTypeAudio, types.MessageTypeVoice:
		// For audio, we could add audio attributes if we had duration info
		// attributes = append(attributes, &tg.DocumentAttributeAudio{...})
	}

	// Create InputMedia for document
	inputMedia := &tg.InputMediaUploadedDocument{
		File:       inputFile,
		MimeType:   mimeType,
		Attributes: attributes,
	}

	// Convert chat to InputPeer
	inputPeer, err := m.client.chatToInputPeer(chat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Build send media request
	request := &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    inputMedia,
		Message:  caption,
		RandomID: m.client.generateRandomID(),
	}

	// Add reply-to information if provided
	if replyToMessageID > 0 {
		request.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: int(replyToMessageID),
		}
	}

	// Send the media
	updates, err := m.client.api.MessagesSendMedia(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send media: %w", err)
	}

	// Extract the sent message from updates
	message, err := m.client.extractMessageFromUpdates(updates, chat.ID, caption)
	if err != nil {
		m.client.logger.Error("Failed to extract message from updates", "error", err)
		// Return a basic message even if extraction fails
		return &types.Message{
			ID:     0,
			ChatID: chat.ID,
			Content: types.MessageContent{
				Type:    mediaType,
				Caption: caption,
				Document: &types.Document{
					FileName: filepath.Base(filePath),
					MimeType: mimeType,
					File: &types.Media{
						Size:     fileInfo.Size(),
						MimeType: mimeType,
					},
				},
			},
			IsOutgoing: true,
		}, nil
	}

	m.client.logger.Info("Document sent successfully", "messageID", message.ID, "type", mediaType)
	return message, nil
}

// UploadAudio uploads an audio file to Telegram and sends it as a message.
func (m *MediaManager) UploadAudio(ctx context.Context, chat *types.Chat, filePath string, caption string, replyToMessageID int64) (*types.Message, error) {
	return m.UploadDocument(ctx, chat, filePath, caption, replyToMessageID, types.MessageTypeAudio)
}

// UploadVideo uploads a video file to Telegram and sends it as a message.
func (m *MediaManager) UploadVideo(ctx context.Context, chat *types.Chat, filePath string, caption string, replyToMessageID int64) (*types.Message, error) {
	return m.UploadDocument(ctx, chat, filePath, caption, replyToMessageID, types.MessageTypeVideo)
}

// UploadFile uploads a generic file to Telegram and sends it as a message.
func (m *MediaManager) UploadFile(ctx context.Context, chat *types.Chat, filePath string, caption string, replyToMessageID int64) (*types.Message, error) {
	return m.UploadDocument(ctx, chat, filePath, caption, replyToMessageID, types.MessageTypeDocument)
}

// DetectMediaType detects the type of media based on file extension and mime type.
func DetectMediaType(filePath string) types.MessageType {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Image extensions
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".tiff": true,
	}

	// Video extensions
	videoExts := map[string]bool{
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true,
		".webm": true, ".flv": true, ".wmv": true, ".m4v": true,
	}

	// Audio extensions
	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".m4a": true,
		".ogg": true, ".opus": true, ".aac": true, ".wma": true,
	}

	// Voice extensions (typically used for voice messages)
	voiceExts := map[string]bool{
		".ogg": true, ".opus": true,
	}

	if imageExts[ext] {
		return types.MessageTypePhoto
	}
	if videoExts[ext] {
		return types.MessageTypeVideo
	}
	if voiceExts[ext] {
		return types.MessageTypeVoice
	}
	if audioExts[ext] {
		return types.MessageTypeAudio
	}

	return types.MessageTypeDocument
}

// GetMediaFromMessage extracts media information from a message for downloading.
func GetMediaFromMessage(msg *tg.Message) (location tg.InputFileLocationClass, size int64, fileName string, err error) {
	if msg.Media == nil {
		return nil, 0, "", fmt.Errorf("message has no media")
	}

	switch media := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := media.Photo.(*tg.Photo)
		if !ok {
			return nil, 0, "", fmt.Errorf("invalid photo type")
		}

		// Select the best photo size based on priority
		bestSize := selectBestPhotoSize(photo.Sizes)

		if bestSize == nil {
			return nil, 0, "", fmt.Errorf("no valid photo size found")
		}

		location = &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     bestSize.Type,
		}
		return location, int64(bestSize.Size), fmt.Sprintf("photo_%d.jpg", photo.ID), nil

	case *tg.MessageMediaDocument:
		doc, ok := media.Document.(*tg.Document)
		if !ok {
			return nil, 0, "", fmt.Errorf("invalid document type")
		}

		// Get filename from attributes
		fileName = fmt.Sprintf("document_%d", doc.ID)
		for _, attr := range doc.Attributes {
			if fileAttr, ok := attr.(*tg.DocumentAttributeFilename); ok {
				fileName = fileAttr.FileName
				break
			}
		}

		// If no extension, add one based on mime type
		if !strings.Contains(fileName, ".") {
			exts, err := mime.ExtensionsByType(doc.MimeType)
			if err == nil && len(exts) > 0 {
				fileName = fileName + exts[0]
			}
		}

		location = &tg.InputDocumentFileLocation{
			ID:            doc.ID,
			AccessHash:    doc.AccessHash,
			FileReference: doc.FileReference,
			ThumbSize:     "",
		}
		return location, doc.Size, fileName, nil

	default:
		return nil, 0, "", fmt.Errorf("unsupported media type: %T", media)
	}
}

// SendMediaFromReader sends media from an io.Reader.
// This is useful for sending media that's already in memory or from a stream.
func (m *MediaManager) SendMediaFromReader(ctx context.Context, chat *types.Chat, reader io.Reader, fileName string, mediaType types.MessageType, caption string, replyToMessageID int64) (*types.Message, error) {
	// Upload the file
	inputFile, err := m.uploader.FromReader(ctx, fileName, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload from reader: %w", err)
	}

	// Detect mime type
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var inputMedia tg.InputMediaClass

	// Create appropriate InputMedia based on type
	if mediaType == types.MessageTypePhoto {
		inputMedia = &tg.InputMediaUploadedPhoto{
			File: inputFile,
		}
	} else {
		// For documents, videos, audio, etc.
		attributes := []tg.DocumentAttributeClass{
			&tg.DocumentAttributeFilename{
				FileName: fileName,
			},
		}

		inputMedia = &tg.InputMediaUploadedDocument{
			File:       inputFile,
			MimeType:   mimeType,
			Attributes: attributes,
		}
	}

	// Convert chat to InputPeer
	inputPeer, err := m.client.chatToInputPeer(chat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Build send media request
	request := &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    inputMedia,
		Message:  caption,
		RandomID: m.client.generateRandomID(),
	}

	// Add reply-to information if provided
	if replyToMessageID > 0 {
		request.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: int(replyToMessageID),
		}
	}

	// Send the media
	updates, err := m.client.api.MessagesSendMedia(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send media: %w", err)
	}

	// Extract the sent message from updates
	message, err := m.client.extractMessageFromUpdates(updates, chat.ID, caption)
	if err != nil {
		m.client.logger.Error("Failed to extract message from updates", "error", err)
		// Return a basic message even if extraction fails
		return &types.Message{
			ID:     0,
			ChatID: chat.ID,
			Content: types.MessageContent{
				Type:    mediaType,
				Caption: caption,
			},
			IsOutgoing: true,
		}, nil
	}

	m.client.logger.Info("Media sent successfully from reader", "messageID", message.ID, "type", mediaType)
	return message, nil
}
