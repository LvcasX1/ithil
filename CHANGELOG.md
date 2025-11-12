# Changelog

All notable changes to Ithil will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-12

### Added
- **Core Messaging Features**
  - Full Telegram authentication (phone, code, 2FA, registration)
  - Real-time message sending and receiving via MTProto
  - Message history loading with pagination
  - Message editing for sent messages
  - **Message deletion with confirmation modal**
  - **Message forwarding with chat selector**
  - **Message reactions with emoji picker**
  - Reply to messages support

- **Chat Management**
  - Chat list with real-time updates
  - Private chats, groups, supergroups, and channels support
  - Chat pinning, muting, and archiving
  - Unread message counters
  - Real-time search and filtering
  - Quick access with number keys (1-9)

- **Media Support**
  - Image rendering (Kitty/Sixel protocols or Unicode half-blocks)
  - Audio playback with waveform visualization
  - Voice message support
  - Video notes (round videos)
  - Photos, stickers, animations, documents
  - Media download and caching
  - Media viewer component

- **Rich Text Features**
  - Bold, italic, code blocks, inline code
  - Links, mentions, hashtags
  - Text entities and formatting preservation

- **User Interface**
  - Three-pane layout (chat list, conversation, info sidebar)
  - Vim-style navigation with keyboard shortcuts
  - Responsive design adapting to terminal size
  - Nord color scheme
  - Status bar with connection status
  - Help modal with keyboard shortcuts
  - Centered message selection for better visibility

- **Privacy & Status**
  - Stealth mode (disable read receipts and typing indicators)
  - User online/offline status
  - Read receipts
  - Typing indicators
  - Last seen information

- **Performance**
  - Local message caching (configurable limits)
  - Efficient update handling with gaps manager
  - Fast search and filtering
  - Optimized media rendering

- **Testing**
  - Comprehensive test suite (89 tests)
  - Excellent coverage:
    - Types: 100%
    - Cache: 97.8%
    - Utils: 90.7%
    - Config: 47%
  - Race condition testing

### Technical Details
- Built with Go 1.23+
- Uses gotd/td for pure Go MTProto implementation
- Bubbletea framework for TUI (Elm Architecture)
- Lipgloss for terminal styling
- Beep library for audio playback
- ~13,000 lines of Go code (including tests)

### Known Limitations
- No inline bots support yet
- No secret chats (E2E encryption) yet
- No desktop notifications yet
- No advanced search functionality yet
- Limited theme options (Nord only)

### Documentation
- Comprehensive README with installation and usage instructions
- CLAUDE.md with detailed project architecture
- CONTRIBUTING.md for contributors
- Example configuration file

[0.1.0]: https://github.com/lvcasx1/ithil/releases/tag/v0.1.0
