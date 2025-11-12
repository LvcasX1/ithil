# Ithil

![Version](https://img.shields.io/badge/version-0.1.2-blue)
![Go Version](https://img.shields.io/badge/go-1.23%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)

**Ithil** (Sindarin for "moon") is a feature-rich Terminal User Interface (TUI) Telegram client built with Go and Bubbletea. It brings the full Telegram experience to your terminal with a beautiful, keyboard-driven interface.

## 🚀 Current Status

Ithil is **production-ready** for daily use! All core features are implemented and stable.

✅ **Complete:** Authentication, real-time messaging, chat management, message history, read receipts, typing indicators, message editing, message deletion, message forwarding, message reactions, rich text formatting, stealth mode, chat pinning/muting/archiving, multimedia support (images with Kitty/Sixel/Unicode rendering, audio playback with controls)

⚠️ **Partially Complete:** Video support (placeholder with metadata, external player integration planned)

🔜 **Planned:** Notifications, advanced search, thumbnails, media caching with LRU eviction, video thumbnails, inline bots, secret chats, multiple themes

The application uses the official Telegram MTProto protocol via gotd/td and implements a sophisticated update handling system for reliable real-time messaging. With ~8,500 lines of well-structured Go code, Ithil demonstrates modern TUI development practices with the Elm Architecture pattern.

## Features

### 🎯 Core Functionality
- **Full Telegram Authentication**: Phone number, verification code, 2FA, and registration support
- **Real-time Messaging**: Send and receive messages instantly via MTProto
- **Chat Management**: Access private chats, groups, supergroups, and channels
- **Message History**: Load and browse complete message history
- **Live Updates**: Real-time message delivery, read receipts, and typing indicators

### 🎨 User Interface
- **Beautiful TUI**: Built with Bubbletea and Lipgloss for smooth terminal rendering
- **Three-Pane Layout**: Chat list, conversation view, and info sidebar
- **Keyboard-Driven**: Vim-style navigation with extensive keyboard shortcuts
- **Responsive Design**: Adapts to terminal size with configurable pane widths
- **Status Bar**: Shows connection status, unread count, and current chat

### ✨ Rich Messaging
- **Message Formatting**: Bold, italic, code blocks, links, mentions, and more
- **Media Support**: Full support for photos, videos, documents, stickers, animations
- **Voice & Video Notes**: Audio playback with waveform visualization for voice messages and video notes
- **Image Rendering**: Display images in terminal using Kitty/Sixel protocols or Unicode half-blocks
- **Special Content**: Polls, contacts, locations, and forwarded messages
- **Message Editing**: Edit your sent messages
- **Reply Support**: Reply to specific messages in conversations

### 🔐 Privacy & Control
- **Stealth Mode**: Disable read receipts and typing indicators (press `S`)
- **Session Management**: Secure session storage with automatic recovery
- **User Status**: See when users are online, offline, or recently active
- **Read Receipts**: Track which messages have been read

### ⚙️ Customization
- **Configurable Layout**: Adjust pane widths and visibility
- **Vim Mode**: Optional vim-style navigation keybindings
- **Flexible Settings**: Control timestamps, avatars, auto-download limits
- **Theme Support**: Dark mode with Nord color scheme (more themes planned)

### 🚀 Performance
- **Fast and Lightweight**: Native Go implementation using gotd/td
- **Local Caching**: Message and user caching for instant access
- **Efficient Updates**: Gaps-aware update handler for reliable message delivery
- **Low Resource Usage**: Minimal dependencies and memory footprint
- **Optimized Navigation**: Centered message selection with smooth scrolling
- **Smart Search**: Real-time chat filtering for instant access to any conversation
- **Advanced Media Rendering**: Automatic graphics protocol detection with fallback chain (Kitty → Sixel → Unicode Mosaic → ASCII)
- **High-Fidelity Images**: Support for Kitty graphics protocol (pixel-perfect) and Sixel protocol (256-color)
- **Audio Playback**: Built-in audio player with waveform visualization and playback controls

## Screenshots

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CHATS              │  CONVERSATION                │  INFO                 │
│                     │                              │                       │
│  📌 Alice           │  Alice: Hey! How are you?    │  Chat Info            │
│     Just now        │  12:34                       │  ─────────────        │
│                     │                              │  Title: Alice         │
│  📌 Development     │  You: I'm good, thanks!      │  Type: Private        │
│     2m ago          │  12:35                       │  Username: @alice     │
│                     │                              │                       │
│  Friends Group   3  │  Alice: Great! Working on    │  Statistics           │
│     5m ago          │  anything interesting?       │  ─────────────        │
│                     │  12:36                       │  Messages: 142        │
│  Mom                │                              │                       │
│     1h ago          │  You: Yes! Building a TUI    │                       │
│                     │  Telegram client 🚀          │                       │
│  🔇 Notifications   │  12:37                       │                       │
│     Yesterday       │                              │                       │
│                     │  ┌─────────────────────────┐ │                       │
│                     │  │ Type a message...       │ │                       │
│                     │  └─────────────────────────┘ │                       │
├─────────────────────┴──────────────────────────────┴───────────────────────┤
│  ITHIL  Connected                                    0 unread  ? for help  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Installation

### Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap lvcasx1/tap
brew install ithil
```

#### Scoop (Windows)
```bash
scoop bucket add lvcasx1 https://github.com/lvcasx1/scoop-bucket
scoop install ithil
```

#### Chocolatey (Windows)
```bash
choco install ithil
```

#### AUR (Arch Linux)
```bash
yay -S ithil-bin
# or
paru -S ithil-bin
```

#### DEB (Debian/Ubuntu)
```bash
# Download the .deb file from releases
wget https://github.com/lvcasx1/ithil/releases/latest/download/ithil_*_Linux_x86_64.deb
sudo dpkg -i ithil_*_Linux_x86_64.deb
```

#### RPM (Fedora/RHEL/CentOS)
```bash
# Download the .rpm file from releases
wget https://github.com/lvcasx1/ithil/releases/latest/download/ithil_*_Linux_x86_64.rpm
sudo rpm -i ithil_*_Linux_x86_64.rpm
```

#### APK (Alpine Linux)
```bash
# Download the .apk file from releases
wget https://github.com/lvcasx1/ithil/releases/latest/download/ithil_*_Linux_x86_64.apk
sudo apk add --allow-untrusted ithil_*_Linux_x86_64.apk
```

### Direct Download

Download pre-built binaries from the [releases page](https://github.com/lvcasx1/ithil/releases).

### From Source

**Prerequisites**: Go 1.23 or later

```bash
# Clone the repository
git clone https://github.com/lvcasx1/ithil.git
cd ithil

# Install dependencies
go mod download

# Build the application
go build -o bin/ithil ./cmd/ithil

# Run Ithil
./bin/ithil
```

### Using Go Install

```bash
go install github.com/lvcasx1/ithil/cmd/ithil@latest
```

## API Credentials

### Zero-Setup Experience (Default)

**Ithil works out of the box with no configuration required!**

By default, Ithil uses built-in API credentials, allowing you to download and run the application immediately without any setup. Just install and start chatting!

```bash
# Download and run - that's it!
ithil
```

### Custom Credentials (Enhanced Privacy)

For users who want enhanced privacy, Ithil supports custom API credentials. With custom credentials:

- You have your own Telegram app identity
- You get your own rate limits
- You have complete control over your API usage

**To use custom credentials:**

#### Option 1: Settings Menu (Recommended)

1. Run Ithil with default credentials
2. Log in to your account
3. Press `Ctrl+,` to open Settings
4. Navigate to the Account tab
5. Toggle to "Use custom credentials"
6. Enter your API ID and API Hash
7. Press `Ctrl+S` to save
8. Restart Ithil and log in again

#### Option 2: Configuration File

1. Get credentials from https://my.telegram.org:
   - Log in with your phone number
   - Go to "API Development Tools"
   - Create a new application
   - Copy your `api_id` and `api_hash`

2. Create or edit `~/.config/ithil/config.yaml`:
```yaml
telegram:
  use_default_credentials: false
  api_id: "12345678"
  api_hash: "abcdef1234567890abcdef1234567890"
```

3. Restart Ithil

**Note:** Switching between default and custom credentials will clear your session for privacy. You'll need to log in again.

## Configuration

Ithil looks for configuration files in the following order:

1. `./config.yaml`
2. `~/.config/ithil/config.yaml`
3. `~/.ithil.yaml`

### Initial Setup (Optional)

Configuration is **completely optional**. Ithil works with default settings right away.

If you want to customize:

```bash
# Copy the example configuration
cp config.example.yaml ~/.config/ithil/config.yaml

# Edit the configuration
nano ~/.config/ithil/config.yaml  # or your preferred editor
```

### Configuration Options

```yaml
telegram:
  api_id: "YOUR_API_ID"              # Required: From my.telegram.org
  api_hash: "YOUR_API_HASH"          # Required: From my.telegram.org
  session_file: "~/.config/ithil/session.json"
  database_directory: "~/.config/ithil/tdlib"

ui:
  theme: "dark"                      # dark, light, nord

  layout:
    chat_list_width: 25              # Percentage
    conversation_width: 50
    info_width: 25
    show_info_pane: true

  appearance:
    show_avatars: true
    show_status_bar: true
    date_format: "12h"               # 12h or 24h
    relative_timestamps: true
    message_preview_length: 50

  behavior:
    send_on_enter: true              # false for Ctrl+Enter
    auto_download_limit: 5242880     # 5MB in bytes
    mark_read_on_scroll: true
    emoji_style: "unicode"           # unicode or ascii

  keyboard:
    vim_mode: true                   # Enable vim-style navigation
    custom_bindings: {}

privacy:
  stealth_mode: false                # Toggle with 'S' - disables read receipts/typing
  show_online_status: true
  show_read_receipts: true
  show_typing: true

cache:
  max_messages_per_chat: 1000
  max_media_size: 104857600          # 100MB
  media_directory: "~/.cache/ithil/media"

logging:
  level: "info"                      # debug, info, warn, error
  file: "~/.config/ithil/ithil.log"
```

## Usage

### Basic Commands

```bash
# Run Ithil
ithil

# Specify a custom config file
ithil -config /path/to/config.yaml

# Show version
ithil -version

# Show help
ithil -help
```

### Keyboard Shortcuts

#### Global

| Key | Action |
|-----|--------|
| `Ctrl+C`, `Ctrl+Q` | Quit application |
| `?` | Toggle help |
| `Tab` | Next pane |
| `Shift+Tab` | Previous pane |
| `Ctrl+1` | Focus chat list |
| `Ctrl+2` | Focus conversation |
| `Ctrl+3` | Focus sidebar |
| `Ctrl+S` | Toggle sidebar |
| `Ctrl+,` | Open settings |
| `S` | Toggle stealth mode |
| `Ctrl+R` | Refresh |
| `/`, `Ctrl+F` | Search |

#### Chat List Navigation

| Key | Action |
|-----|--------|
| `j`, `↓` | Move down |
| `k`, `↑` | Move up |
| `g`, `Home` | Go to top |
| `G`, `End` | Go to bottom |
| `PgUp`, `Ctrl+B` | Page up |
| `PgDown`, `Ctrl+F` | Page down |
| `Ctrl+U` | Jump up 5 chats (fast navigation) |
| `Ctrl+D` | Jump down 5 chats (fast navigation) |
| `1-9` | Quick jump to chat 1-9 and open |
| `Enter`, `l`, `→` | Open selected chat |
| `/` | Enter search mode |

#### Chat List Actions

| Key | Action |
|-----|--------|
| `p` | Pin/unpin chat |
| `m` | Mute/unmute chat |
| `a` | Archive chat |
| `r` | Mark as read |
| `d` | Delete chat |

#### Conversation Navigation

**Note:** The conversation view uses a **centered selection cursor** (▶) that keeps the selected message in the middle of the viewport for maximum visibility.

| Key | Action |
|-----|--------|
| `j`, `↓` | Select next message (moves ▶ cursor, keeps centered) |
| `k`, `↑` | Select previous message (moves ▶ cursor, keeps centered) |
| `Ctrl+U` | Scroll up half page |
| `Ctrl+D` | Scroll down half page |
| `Ctrl+B`, `PgUp` | Scroll up full page |
| `Ctrl+F`, `PgDn` | Scroll down full page |
| `g`, `Home` | Go to top |
| `G`, `End` | Go to bottom |
| `i`, `a` | Focus input field |
| `Enter` | View/play media for selected message |

#### Message Actions

**Note:** All actions apply to the selected message (marked with ▶ cursor)

| Key | Action |
|-----|--------|
| `r` | Reply to selected message |
| `e` | Edit selected message (if outgoing) |
| `d` | Delete message |
| `f` | Forward message |
| `y` | Copy message |
| `x` | React to message |
| `p` | Pin message |
| `s` | Save/download |
| `v` | View media |
| `o` | Open link |

#### Message Input

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+Enter` | Send message (alternative) |
| `Shift+Enter` | New line |
| `Ctrl+A` | Attach file |
| `Ctrl+E` | Insert emoji |
| `Esc` | Cancel reply/edit |

## Development

### Project Structure

```
ithil/
├── cmd/
│   └── ithil/              # Application entry point
│       └── main.go         # Main program with initialization
├── internal/
│   ├── app/                # Application core
│   │   ├── app.go          # Application lifecycle management
│   │   └── config.go       # Configuration loading and validation
│   ├── telegram/           # Telegram client (gotd/td wrapper)
│   │   ├── client.go       # Client initialization and lifecycle
│   │   ├── auth.go         # Authentication flow implementation
│   │   ├── messages.go     # Message operations (send, edit, get)
│   │   ├── chats.go        # Chat operations (list, get, search)
│   │   ├── updates.go      # Real-time update handler
│   │   ├── session.go      # Session storage and management
│   │   └── media.go        # Media download/upload manager
│   ├── media/              # Media rendering and playback
│   │   ├── audio_player.go    # Beep-based audio playback
│   │   ├── audio_renderer.go  # Waveform visualization and playback UI
│   │   ├── image_renderer.go  # Kitty/Sixel image rendering
│   │   └── mosaic_renderer.go # Unicode half-block image rendering
│   ├── ui/                 # User interface (Bubbletea)
│   │   ├── models/         # Bubbletea models
│   │   │   ├── main.go     # Root model with update routing
│   │   │   ├── auth.go     # Authentication screen
│   │   │   ├── chatlist.go # Chat list pane
│   │   │   ├── conversation.go # Message view and input
│   │   │   └── sidebar.go  # Info sidebar
│   │   ├── components/     # Reusable UI components
│   │   │   ├── statusbar.go  # Status bar component
│   │   │   ├── input.go      # Text input component
│   │   │   ├── chatitem.go   # Chat list item
│   │   │   ├── message.go    # Message bubble
│   │   │   ├── helpmodal.go  # Help modal with keyboard shortcuts
│   │   │   ├── mediaviewer.go # Media viewer component
│   │   │   ├── modal.go      # Generic modal component
│   │   │   ├── filepicker.go # File picker component
│   │   │   └── utils.go      # UI utility functions
│   │   ├── styles/         # Lipgloss styles
│   │   │   └── styles.go   # Color schemes and styling
│   │   └── keys/           # Keyboard shortcuts
│   │       └── keymap.go   # Key binding definitions
│   ├── cache/              # Local caching layer
│   │   └── cache.go        # Message, chat, and user cache
│   └── utils/              # Utility functions
│       ├── time.go         # Time formatting helpers
│       └── formatting.go   # Text formatting utilities
├── pkg/
│   └── types/              # Shared type definitions
│       └── types.go        # Core types (Message, Chat, User, etc.)
├── .air.toml               # Hot-reload configuration for development
├── config.example.yaml     # Example configuration file
├── CLAUDE.md               # Claude Code project instructions
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── README.md               # This file
```

**Key Directories:**
- `cmd/ithil/` - Entry point and CLI setup
- `internal/telegram/` - All Telegram protocol operations (~2000 LOC)
- `internal/ui/` - Complete TUI implementation (~4000 LOC)
- `internal/media/` - Media rendering and audio playback (~800 LOC)
- `internal/cache/` - Performance-critical caching layer
- `pkg/types/` - Shared data structures (370 LOC)

**Total:** ~13,000 lines of Go code (including tests)

### Development Setup

```bash
# Install Air for hot-reloading
go install github.com/cosmtrek/air@latest

# Run with hot-reload
air

# Or run directly
go run cmd/ithil/main.go
```

### Building

```bash
# Build for current platform
go build -o bin/ithil ./cmd/ithil

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/ithil-linux-amd64 ./cmd/ithil

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o bin/ithil-darwin-amd64 ./cmd/ithil
GOOS=darwin GOARCH=arm64 go build -o bin/ithil-darwin-arm64 ./cmd/ithil

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/ithil-windows-amd64.exe ./cmd/ithil
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

## Architecture

Ithil follows the Elm Architecture pattern (Model-Update-View) using Bubbletea:

- **Models**: Represent application state
- **Update**: Handle messages and update state
- **View**: Render the current state to the terminal

### Key Components

1. **MainModel** (`internal/ui/models/main.go`): Root model coordinating all sub-models and handling Telegram updates
2. **AuthModel** (`internal/ui/models/auth.go`): Interactive authentication flow (phone, code, 2FA, registration)
3. **ChatListModel** (`internal/ui/models/chatlist.go`): Chat list management with real-time updates
4. **ConversationModel** (`internal/ui/models/conversation.go`): Message display, input, and editing with centered selection
5. **SidebarModel** (`internal/ui/models/sidebar.go`): Chat/user information and statistics
6. **TelegramClient** (`internal/telegram/client.go`): Wrapper around gotd/td with MTProto implementation
7. **UpdateHandler** (`internal/telegram/updates.go`): Real-time update processing and gap recovery
8. **MediaManager** (`internal/telegram/media.go`): Media download/upload operations
9. **AudioPlayer** (`internal/media/audio_player.go`): Beep-based audio playback for voice messages
10. **ImageRenderer** (`internal/media/image_renderer.go`): Multi-protocol image rendering (Kitty/Sixel/Unicode)
11. **Cache** (`internal/cache/cache.go`): Local caching layer for messages, chats, and users
12. **SessionStorage** (`internal/telegram/session.go`): Secure session and auth data persistence

### Data Flow

```
Telegram Server (MTProto)
        ↓
gotd/td Client
        ↓
UpdateHandler → Cache → UI Models → Bubbletea Renderer → Terminal
        ↑                    ↓
        └─── User Input ─────┘
```

### Update Processing

Ithil uses a sophisticated update handling system:

1. **Gaps Manager**: Ensures no updates are lost using sequence tracking
2. **UpdateHandler**: Converts Telegram updates to internal types
3. **Cache Layer**: Stores messages, users, and chats for instant access
4. **UI Updates**: Reactive updates trigger re-renders only when needed

## Recent Enhancements

### Media Support ✅ (Fully Implemented - v0.3.0)
- **Automatic Protocol Detection**: Detects and uses the best graphics protocol available (Kitty → Sixel → Unicode → ASCII)
- **Kitty Graphics Protocol**: Pixel-perfect image rendering in Kitty terminal
- **Sixel Graphics Protocol**: High-quality 256-color rendering in XTerm, WezTerm, Alacritty, and more
- **Unicode Mosaic**: True-color half-block rendering for terminals without graphics support
- **ASCII Art Fallback**: Universal compatibility with all terminals
- **Audio Playback**: Full-featured audio player with waveform visualization, playback controls (play/pause, seek, volume)
- **Voice Messages**: Specialized UI for voice notes with duration and progress tracking
- **Video Notes**: Support for round video messages with preview and playback
- **Media Viewer**: Dedicated media viewer component with keyboard navigation
- **File Management**: Download and cache media files with configurable limits

**Note:** Media support is functional but not fully complete. Some edge cases and advanced features are still in development.

### Navigation Improvements ✅ (v0.2.0)
- **Centered Selection**: Message selection cursor (▶) stays centered in viewport for better context
- **Fast Chat Navigation**: `Ctrl+U`/`Ctrl+D` to jump 5 chats at a time
- **Quick Access**: Number keys `1-9` for instant chat access
- **Smart Search**: Press `/` to filter chats in real-time by name or content
- **Enhanced Vim Support**: Additional vim-style bindings for power users

## Roadmap

### Phase 1: Foundation ✅ (Completed)
- [x] Project setup and structure
- [x] Configuration management
- [x] Basic UI layout (three panes)
- [x] Keyboard navigation
- [x] Telegram authentication (phone, code, 2FA, registration)

### Phase 2: Core Features ✅ (Completed)
- [x] gotd/td client integration (MTProto)
- [x] Message sending and receiving
- [x] Chat list with real data
- [x] Message history loading
- [x] Real-time updates via update handler
- [x] Local message caching
- [x] Session management

### Phase 3: Rich Features ✅ (Completed)
- [x] Message editing
- [x] Rich text formatting (bold, italic, code, links, etc.)
- [x] Media support (photos, videos, documents, stickers, animations)
- [x] Polls
- [x] Contact messages
- [x] Location messages
- [x] Media download and display
- [x] Image rendering (Kitty/Sixel/Unicode)
- [x] Audio playback with waveform visualization
- [x] Video notes (round videos)

### Phase 4: Advanced Features ✅ (Completed)
- [x] Read receipts
- [x] Typing indicators
- [x] User online/offline status
- [x] Stealth mode (disable read receipts/typing)
- [x] Voice messages
- [x] Video messages
- [x] Chat pinning/muting/archiving
- [x] Message reactions with emoji picker
- [x] Message forwarding with chat selector
- [x] Message deletion with confirmation
- [ ] Inline bots
- [ ] Secret chats

### Phase 5: Polish 🔜 (Planned)
- [ ] Notifications
- [ ] Search functionality
- [ ] Multiple themes
- [ ] Performance optimizations
- [x] Comprehensive testing (89 tests, excellent coverage)
- [ ] Media caching
- [ ] File upload support

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines

1. Follow Go best practices and idioms
2. Use meaningful variable and function names
3. Add comments for exported functions
4. Write tests for new features
5. Update documentation as needed
6. Use the Nord color scheme for styling

## Dependencies

### Core Libraries
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework for the Elm Architecture
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling and layout
- [Bubbles](https://github.com/charmbracelet/bubbles) - Pre-built UI components (textarea, viewport, etc.)
- [gotd/td](https://github.com/gotd/td) - Pure Go MTProto implementation for Telegram
- [Beep](https://github.com/faiface/beep) - Audio playback library for voice messages

### Why gotd/td?

Ithil uses **gotd/td** instead of TDLib for several advantages:

- **Pure Go**: No CGo dependencies, easier cross-compilation
- **Type-safe**: Fully typed Telegram API with generated client
- **Modern**: Built for Go with context support and proper error handling
- **Lightweight**: No heavy C++ runtime, smaller binary size
- **Direct MTProto**: Direct protocol implementation without abstractions
- **Update System**: Built-in gaps manager for reliable update delivery

### Development Tools
- [Air](https://github.com/cosmtrek/air) - Hot-reload for development
- Go 1.23+ - Modern Go features and performance

## Technical Highlights

### Authentication System
- **Multi-step Flow**: Handles phone number → code → 2FA → registration seamlessly
- **Session Recovery**: Automatic detection and recovery from invalid sessions
- **Error Handling**: User-friendly error messages for common auth issues (AUTH_RESTART, PHONE_CODE_INVALID, etc.)
- **Secure Storage**: Session data persisted with proper file permissions

### Real-time Updates
- **Gaps Manager Integration**: Zero message loss with sequence number tracking
- **Update Type Coverage**: 10+ update types handled (messages, edits, deletions, read receipts, typing, user status)
- **Efficient Processing**: Updates processed in background goroutines without blocking UI
- **Smart Caching**: Users and messages cached during update processing for instant display

### Message Handling
- **Rich Content Support**: 13+ message entity types (bold, italic, links, mentions, code, etc.)
- **Media Type Support**: Full support for photos, videos, documents, stickers, animations, voice, video notes, audio
- **Media Rendering**: In-terminal image display using Kitty/Sixel protocols or Unicode half-blocks
- **Audio Playback**: Voice messages and audio files with waveform visualization using Beep library
- **Special Messages**: Full support for polls, contacts, locations, forwarded messages
- **Optimistic UI**: Sent messages appear instantly, confirmed by server asynchronously
- **Centered Selection**: Message selection cursor stays centered for better context and visibility

### Performance Optimizations
- **Lazy Loading**: Messages loaded on-demand as chats are opened
- **Viewport Rendering**: Only visible messages rendered to terminal
- **Efficient Re-renders**: Granular update messages prevent full UI redraws
- **Local Cache**: In-memory cache with configurable limits for instant access
- **Centered Selection**: Smart scrolling keeps selected message centered for optimal visibility
- **Media Caching**: Downloaded media files cached locally for instant re-access
- **Real-time Search**: O(n) filtering across titles, usernames, and message content
- **Async Media Loading**: Media downloads and renders in background without blocking UI

### Code Quality
- **Clear Architecture**: Separation of concerns (UI, Business Logic, Protocol Layer)
- **Type Safety**: Strongly typed throughout with custom type definitions
- **Error Handling**: Comprehensive error handling with logging at each layer
- **Documentation**: Well-commented code with package documentation
- **Maintainability**: ~8,500 LOC organized into logical packages

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Telegram](https://telegram.org) for the amazing messaging platform
- [Charm](https://charm.sh) for the excellent TUI libraries
- [TDLib](https://core.telegram.org/tdlib) for the Telegram client library
- [Nord](https://www.nordtheme.com/) for the beautiful color scheme

## Support

- **Issues**: https://github.com/lvcasx1/ithil/issues
- **Discussions**: https://github.com/lvcasx1/ithil/discussions
- **Telegram**: Coming soon!

## Disclaimer

This project is not affiliated with Telegram or its parent company. It is an independent client built using the official Telegram API.

---

**Built with ❤️ using Go and Bubbletea**
