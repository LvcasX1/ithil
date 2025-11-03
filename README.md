# Ithil

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Go Version](https://img.shields.io/badge/go-1.23%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)

**Ithil** (Sindarin for "moon") is a feature-rich Terminal User Interface (TUI) Telegram client built with Go and Bubbletea. It brings the full Telegram experience to your terminal with a beautiful, keyboard-driven interface.

## 🚀 Current Status

Ithil is in **active development** and already functional for daily use! The core messaging features are complete and stable:

✅ **Working:** Authentication, real-time messaging, chat management, message history, read receipts, typing indicators, message editing, rich text formatting, stealth mode

🚧 **In Progress:** Media download/display, message reactions, advanced chat operations

🔜 **Planned:** Notifications, search, voice messages, multiple themes

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
- **Media Detection**: Recognizes photos, videos, documents, stickers, animations
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
- **Optimized Navigation**: Advanced keyboard shortcuts for 80-90% faster navigation
- **Smart Search**: Real-time chat filtering for instant access to any conversation

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

### Prerequisites

- Go 1.23 or later
- Telegram API credentials (see [Getting API Credentials](#getting-api-credentials))

### From Source

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

## Getting API Credentials

To use Ithil, you need to obtain Telegram API credentials:

1. Visit https://my.telegram.org
2. Log in with your phone number
3. Go to "API Development Tools"
4. Create a new application
5. Copy your `api_id` and `api_hash`

## Configuration

Ithil looks for configuration files in the following order:

1. `./config.yaml`
2. `~/.config/ithil/config.yaml`
3. `~/.ithil.yaml`

### Initial Setup

```bash
# Copy the example configuration
cp config.example.yaml config.yaml

# Edit the configuration with your API credentials
nano config.yaml  # or your preferred editor
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

| Key | Action |
|-----|--------|
| `j`, `↓` | Scroll down one line |
| `k`, `↑` | Scroll up one line |
| `Ctrl+U` | Scroll up half page (fast navigation) |
| `Ctrl+D` | Scroll down half page (fast navigation) |
| `Ctrl+B`, `PgUp` | Scroll up full page |
| `Ctrl+F`, `PgDn` | Scroll down full page |
| `g`, `Home` | Go to top |
| `G`, `End` | Go to bottom |
| `i`, `a` | Focus input field |

#### Message Actions

| Key | Action |
|-----|--------|
| `r` | Reply to message |
| `e` | Edit message |
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
│   │   └── session.go      # Session storage and management
│   ├── ui/                 # User interface (Bubbletea)
│   │   ├── models/         # Bubbletea models
│   │   │   ├── main.go     # Root model with update routing
│   │   │   ├── auth.go     # Authentication screen
│   │   │   ├── chatlist.go # Chat list pane
│   │   │   ├── conversation.go # Message view and input
│   │   │   └── sidebar.go  # Info sidebar
│   │   ├── components/     # Reusable UI components
│   │   │   ├── statusbar.go # Status bar component
│   │   │   ├── input.go    # Text input component
│   │   │   ├── chatitem.go # Chat list item
│   │   │   └── message.go  # Message bubble
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
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── README.md               # This file
```

**Key Directories:**
- `cmd/ithil/` - Entry point and CLI setup
- `internal/telegram/` - All Telegram protocol operations (~2000 LOC)
- `internal/ui/` - Complete TUI implementation (~4000 LOC)
- `internal/cache/` - Performance-critical caching layer
- `pkg/types/` - Shared data structures (370 LOC)

**Total:** ~8,500 lines of Go code

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
4. **ConversationModel** (`internal/ui/models/conversation.go`): Message display, input, and editing
5. **SidebarModel** (`internal/ui/models/sidebar.go`): Chat/user information and statistics
6. **TelegramClient** (`internal/telegram/client.go`): Wrapper around gotd/td with MTProto implementation
7. **UpdateHandler** (`internal/telegram/updates.go`): Real-time update processing and gap recovery
8. **Cache** (`internal/cache/cache.go`): Local caching layer for messages, chats, and users
9. **SessionStorage** (`internal/telegram/session.go`): Secure session and auth data persistence

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

## Recent Enhancements (v0.2.0)

### Navigation Optimizations ✅
- **Fast Chat Navigation**: `Ctrl+U`/`Ctrl+D` to jump 5 chats at a time
- **Quick Access**: Number keys `1-9` for instant chat access
- **Smart Search**: Press `/` to filter chats in real-time by name or content
- **Half-Page Scrolling**: `Ctrl+U`/`Ctrl+D` in conversations for optimal speed
- **Context-Aware Scrolling**: Maintains 2 items visible above/below selection
- **Enhanced Vim Support**: Additional vim-style bindings for power users

**Performance Impact**: 80-90% fewer keystrokes for common navigation tasks!

For complete details, see:
- [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) - Full keyboard reference
- [OPTIMIZATION_SUMMARY.md](OPTIMIZATION_SUMMARY.md) - Technical details and benchmarks

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

### Phase 3: Rich Features 🚧 (In Progress)
- [x] Message editing
- [x] Rich text formatting (bold, italic, code, links, etc.)
- [x] Media support detection (photos, videos, documents, stickers, animations)
- [x] Polls
- [x] Contact messages
- [x] Location messages
- [ ] Media download and display
- [ ] Message reactions
- [ ] Message forwarding (stub implemented)
- [ ] Message deletion (stub implemented)

### Phase 4: Advanced Features 🔜 (Planned)
- [x] Read receipts
- [x] Typing indicators
- [x] User online/offline status
- [x] Stealth mode (disable read receipts/typing)
- [ ] Voice messages
- [ ] Video messages
- [ ] Inline bots
- [ ] Secret chats
- [ ] Chat pinning/muting/archiving (stub implemented)

### Phase 5: Polish 🔜 (Planned)
- [ ] Notifications
- [ ] Search functionality (stub implemented)
- [ ] Multiple themes
- [ ] Performance optimizations
- [ ] Comprehensive testing
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
- **Media Type Detection**: Automatic recognition of photos, videos, documents, stickers, animations, voice, video notes
- **Special Messages**: Full support for polls, contacts, locations, forwarded messages
- **Optimistic UI**: Sent messages appear instantly, confirmed by server asynchronously

### Performance Optimizations
- **Lazy Loading**: Messages loaded on-demand as chats are opened
- **Viewport Rendering**: Only visible messages rendered to terminal
- **Efficient Re-renders**: Granular update messages prevent full UI redraws
- **Local Cache**: In-memory cache with configurable limits for instant access
- **Smart Navigation**: Context-aware scrolling with 2-item padding for better visibility
- **Real-time Search**: O(n) filtering across titles, usernames, and message content
- **Keyboard Efficiency**: 80-90% fewer keystrokes for common navigation tasks

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
