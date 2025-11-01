# Media Support in Ithil

This document describes the media support implementation in Ithil, allowing users to send and receive photos, videos, audio files, and documents through the Telegram TUI client.

## Overview

Ithil now supports sending and receiving media files including:
- **Photos** (JPG, PNG, GIF, WEBP, etc.)
- **Videos** (MP4, MOV, AVI, MKV, WEBM, etc.)
- **Audio** (MP3, WAV, FLAC, OGG, etc.)
- **Voice Messages** (OGG, OPUS)
- **Documents** (any file type)

## Architecture

### Components

1. **Media Manager** (`/home/lvcasx1/Work/personal/ithil/internal/telegram/media.go`)
   - Handles file upload and download operations
   - Manages local media storage
   - Provides type detection for media files

2. **Type System** (`/home/lvcasx1/Work/personal/ithil/pkg/types/types.go`)
   - Already had comprehensive media type definitions
   - Includes Media, Document, Animation, Sticker types
   - Tracks download state and file metadata

3. **UI Components**
   - **Input Component**: Enhanced with file attachment support
   - **File Picker**: Simple directory-based file browser
   - **Message Component**: Enhanced display with media icons and metadata

### Storage Strategy

Media files are stored in the local filesystem:
- **Location**: `~/.ithil/media/`
- **Structure**: `~/.ithil/media/{chat_id}/{filename}`
- **State Tracking**: Download state tracked in Media struct

## Usage

### Sending Media Files

1. **Attach a File**:
   - When composing a message, press `Ctrl+A` to open the file picker
   - Navigate using arrow keys (`↑`/`↓` or `j`/`k`)
   - Press `Enter` to select a file or enter a directory
   - Press `Backspace` to go to parent directory
   - Press `Esc` to cancel

2. **Add Caption** (Optional):
   - After attaching a file, type a caption in the input field
   - The caption will be sent along with the media

3. **Remove Attachment**:
   - Press `Ctrl+X` to remove the attached file before sending

4. **Send**:
   - Press `Enter` to send the media (with optional caption)
   - The file type is automatically detected from the extension

### Receiving Media

When you receive a message with media:
- Media is displayed with an icon (📷 for photos, 🎥 for videos, etc.)
- Metadata is shown: file size, dimensions, duration
- Download status is indicated when available

### Supported Media Types

The system automatically detects media type based on file extension:

**Photos**:
- `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`, `.tiff`

**Videos**:
- `.mp4`, `.mov`, `.avi`, `.mkv`, `.webm`, `.flv`, `.wmv`, `.m4v`

**Audio**:
- `.mp3`, `.wav`, `.flac`, `.m4a`, `.ogg`, `.opus`, `.aac`, `.wma`

**Voice Messages**:
- `.ogg`, `.opus` (detected as voice if in specific format)

**Documents**:
- Any other file type is sent as a document

## Implementation Details

### Media Upload Flow

1. User selects file via file picker (`Ctrl+A`)
2. File path stored in input component
3. On send (`Enter`):
   - File type detected from extension
   - File opened and read
   - Uploaded to Telegram using `uploader.FromReader()`
   - Sent via `MessagesSendMedia` API
   - Message added to conversation

### Media Download Flow

Currently, the infrastructure is in place for downloading media:
- `DownloadPhoto()` and `DownloadDocument()` methods implemented
- Uses `downloader.Download()` from gotd library
- Files saved to `~/.ithil/media/{chat_id}/`

**Note**: Automatic download on receive is not yet implemented. The download methods can be triggered manually when needed.

### File Type Detection

The `DetectMediaType()` function in `media.go` uses file extensions to determine the appropriate Telegram media type:

```go
mediaType := DetectMediaType(filePath)
// Returns: MessageTypePhoto, MessageTypeVideo, MessageTypeAudio,
//          MessageTypeVoice, or MessageTypeDocument
```

### API Integration

The implementation uses the gotd library's upload/download capabilities:

**Upload**:
```go
upload, err := uploader.FromReader(ctx, fileName, fileReader)
inputFile := &tg.InputFile{
    ID:    upload.ID,
    Parts: upload.Parts,
    Name:  fileName,
}
```

**Download**:
```go
location := &tg.InputDocumentFileLocation{...}
_, err := downloader.Download(api, location).Stream(ctx, file)
```

## Keyboard Shortcuts

When input is focused:
- `Ctrl+A`: Open file picker to attach a file
- `Ctrl+X`: Remove current file attachment
- `Enter`: Send message/media
- `Esc`: Cancel and unfocus input

In file picker:
- `↑`/`↓` or `j`/`k`: Navigate files
- `Enter`: Select file or open directory
- `Backspace` or `h`: Go to parent directory
- `Esc` or `q`: Cancel file picker

## Code Structure

### New Files Created

1. **`/home/lvcasx1/Work/personal/ithil/internal/telegram/media.go`**
   - MediaManager struct and methods
   - Upload/download functions for all media types
   - Type detection utilities

2. **`/home/lvcasx1/Work/personal/ithil/internal/ui/components/filepicker.go`**
   - FilePickerComponent implementation
   - Directory navigation
   - File selection messages

### Modified Files

1. **`/home/lvcasx1/Work/personal/ithil/internal/telegram/client.go`**
   - Added MediaManager field
   - Added `SendMediaMessage()` method
   - Initialize media manager with client

2. **`/home/lvcasx1/Work/personal/ithil/internal/ui/components/input.go`**
   - Added AttachedFile field
   - Added attachment management methods
   - Updated help text and indicators

3. **`/home/lvcasx1/Work/personal/ithil/internal/ui/components/message.go`**
   - Enhanced media display with icons
   - Added file size, dimension, and duration formatting
   - Show download status

4. **`/home/lvcasx1/Work/personal/ithil/internal/ui/models/conversation.go`**
   - Handle file attachment requests
   - Updated sendMessage to support media
   - Added keyboard shortcuts for attachments

5. **`/home/lvcasx1/Work/personal/ithil/internal/ui/models/main.go`**
   - Integrated file picker component
   - Handle file selection/cancellation messages
   - Overlay file picker when active

## Future Enhancements

Potential improvements for media support:

1. **Automatic Media Download**
   - Download media automatically when viewing messages
   - Configurable download size limits
   - Background download queue

2. **Media Preview**
   - ASCII art preview for images
   - Thumbnail support
   - External viewer integration

3. **Progress Indicators**
   - Upload/download progress bars
   - File transfer status in UI
   - Cancellable transfers

4. **Media Gallery**
   - View all media in a chat
   - Quick media search
   - Media type filtering

5. **Drag & Drop** (if terminal supports)
   - Drop files to attach
   - Integration with terminal emulator features

6. **Media Compression**
   - Option to compress images before sending
   - Quality selection for photos/videos
   - Optimize file sizes

## Testing

To test media support:

1. **Send a Photo**:
   ```
   - Open a chat
   - Press 'i' or 'a' to focus input
   - Press Ctrl+A to open file picker
   - Navigate to an image file
   - Press Enter to select
   - (Optional) Type a caption
   - Press Enter to send
   ```

2. **Send a Document**:
   - Same process as above
   - Select any non-media file
   - Will be sent as a document

3. **Receive Media**:
   - Open a chat with media messages
   - Media will display with appropriate icons
   - File info shown below media indicator

## Technical Notes

### Dependencies

The media support relies on these gotd packages:
- `github.com/gotd/td/telegram/uploader` - File uploads
- `github.com/gotd/td/telegram/downloader` - File downloads
- `github.com/gotd/td/tg` - Telegram API types

### Performance Considerations

- Large files are streamed to avoid memory issues
- Upload/download operations are asynchronous
- File type detection is based on extension (fast)
- Media metadata is cached in message types

### Security

- File paths are validated before access
- Media directory permissions: 0755
- No automatic execution of downloaded files
- User confirmation required for file selection

## Troubleshooting

**File picker doesn't open**:
- Ensure you're in input mode (press 'i' or 'a' first)
- Check that Ctrl+A is not intercepted by terminal

**Upload fails**:
- Check file permissions
- Verify file exists at the selected path
- Check network connection
- Large files may take time to upload

**Media not displaying**:
- Media display requires proper message type conversion
- Check that message was received correctly
- Verify media metadata in message content

## API Reference

### MediaManager Methods

```go
// Upload photo
UploadPhoto(ctx, chat, filePath, caption, replyToID) (*types.Message, error)

// Upload video
UploadVideo(ctx, chat, filePath, caption, replyToID) (*types.Message, error)

// Upload audio
UploadAudio(ctx, chat, filePath, caption, replyToID) (*types.Message, error)

// Upload generic file
UploadFile(ctx, chat, filePath, caption, replyToID) (*types.Message, error)

// Download photo
DownloadPhoto(ctx, photo, chatID) (string, error)

// Download document
DownloadDocument(ctx, doc, chatID) (string, error)
```

### Client Methods

```go
// Send media (auto-detects type)
SendMediaMessage(chat, filePath, caption, replyToID) (*types.Message, error)

// Get media manager
GetMediaManager() *MediaManager
```

### Input Component Methods

```go
// Attach file
SetAttachment(filePath string)

// Get attached file
GetAttachment() string

// Check if file attached
HasAttachment() bool

// Remove attachment
ClearAttachment()
```

---

**Version**: 1.0
**Last Updated**: 2025-10-31
**Author**: Claude (Anthropic)
