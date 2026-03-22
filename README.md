# Ithil

![Version](https://img.shields.io/badge/version-0.2.1-blue)
![Rust](https://img.shields.io/badge/rust-1.75%2B-orange)
![License](https://img.shields.io/badge/license-MIT-green)

**Ithil** (Sindarin for "moon") is a feature-rich Terminal User Interface (TUI) Telegram client built with Rust and Ratatui. It brings the full Telegram experience to your terminal with a beautiful, keyboard-driven interface.

## Features

### Core Functionality
- **Full Telegram Authentication**: Phone number, verification code, 2FA support
- **Real-time Messaging**: Send and receive messages instantly via MTProto
- **Chat Management**: Access private chats, groups, supergroups, and channels
- **Message History**: Load and browse complete message history
- **Live Updates**: Real-time message delivery, read receipts, and typing indicators

### User Interface
- **Beautiful TUI**: Built with Ratatui and Crossterm for smooth terminal rendering
- **Three-Pane Layout**: Chat list, conversation view, and info sidebar
- **Keyboard-Driven**: Vim-style navigation with extensive keyboard shortcuts
- **Responsive Design**: Adapts to terminal size with configurable pane widths
- **Nord Theme**: Consistent styling with the Nord color scheme
- **Status Bar**: Shows connection status, unread count, and current chat

### Rich Messaging
- **Message Formatting**: Bold, italic, code blocks, links, mentions, and more
- **Media Support**: Photos with download and viewing capabilities
- **Special Content**: Polls, contacts, locations, and forwarded messages
- **Message Editing**: Edit your sent messages
- **Reply Support**: Reply to specific messages in conversations

### Privacy & Control
- **Stealth Mode**: Disable read receipts and typing indicators (press `S`)
- **Session Management**: Secure session storage with automatic recovery
- **User Status**: See when users are online, offline, or recently active
- **Read Receipts**: Track which messages have been read

### Customization
- **Configurable Layout**: Adjust pane widths and visibility
- **Vim Mode**: Optional vim-style navigation keybindings
- **Flexible Settings**: Control timestamps, avatars, auto-download limits

### Performance
- **Fast and Lightweight**: Native Rust implementation with async Tokio runtime
- **Local Caching**: In-memory message and user caching for instant access
- **Efficient Updates**: Real-time update streaming without blocking the UI
- **Low Resource Usage**: Minimal memory footprint with optimized rendering
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

### Homebrew (macOS/Linux)
```bash
brew tap lvcasx1/tap
brew install ithil
```

### AUR (Arch Linux)
```bash
yay -S ithil-bin
# or
paru -S ithil-bin
```

### Cargo
```bash
cargo install ithil
```

### Direct Download

Download pre-built binaries from the [releases page](https://github.com/lvcasx1/ithil/releases).

Available platforms: Linux (x86_64, ARM64), macOS (x86_64, ARM64), Windows (x86_64).

### From Source

**Prerequisites**: Rust 1.75 or later

```bash
git clone https://github.com/lvcasx1/ithil.git
cd ithil
cargo build --release
./target/release/ithil
```

## API Credentials

### Zero-Setup Experience (Default)

**Ithil works out of the box with no configuration required!**

By default, Ithil uses built-in API credentials, allowing you to download and run the application immediately without any setup. Just install and start chatting!

```bash
ithil
```

### Custom Credentials (Enhanced Privacy)

For users who want enhanced privacy, Ithil supports custom API credentials:

- You have your own Telegram app identity
- You get your own rate limits
- You have complete control over your API usage

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

Configuration is **completely optional**. Ithil works with default settings right away.

```yaml
telegram:
  api_id: "YOUR_API_ID"
  api_hash: "YOUR_API_HASH"
  session_file: "~/.config/ithil/ithil.session"

ui:
  layout:
    chat_list_width: 25
    conversation_width: 50
    info_width: 25
    show_info_pane: true

  appearance:
    show_avatars: true
    show_status_bar: true
    date_format: "12h"
    relative_timestamps: true
    message_preview_length: 50

  behavior:
    send_on_enter: true
    auto_download_limit: 5242880
    mark_read_on_scroll: true

  keyboard:
    vim_mode: true

privacy:
  stealth_mode: false
  show_online_status: true
  show_read_receipts: true
  show_typing: true

cache:
  max_messages_per_chat: 1000
  max_media_size: 104857600
  media_directory: "~/.cache/ithil/media"

logging:
  level: "info"
  file: "~/.config/ithil/ithil.log"
```

## Usage

### Basic Commands

```bash
# Run Ithil
ithil

# Specify a custom config file
ithil --config /path/to/config.yaml

# Enable debug logging
ithil --debug

# Show version
ithil --version

# Show help
ithil --help
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
| `Ctrl+U` | Jump up 5 chats |
| `Ctrl+D` | Jump down 5 chats |
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
| `j`, `↓` | Select next message |
| `k`, `↑` | Select previous message |
| `Ctrl+U` | Scroll up half page |
| `Ctrl+D` | Scroll down half page |
| `Ctrl+B`, `PgUp` | Scroll up full page |
| `Ctrl+F`, `PgDn` | Scroll down full page |
| `g`, `Home` | Go to top |
| `G`, `End` | Go to bottom |
| `i`, `a` | Focus input field |
| `Enter` | View/play media for selected message |

#### Message Actions

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
| `Esc` | Cancel reply/edit |

## Development

### Project Structure

```
ithil/
├── src/
│   ├── main.rs              # Entry point, CLI parsing, event loop
│   ├── lib.rs               # Library root, module re-exports
│   ├── app/                 # Configuration and credentials
│   │   ├── config.rs        # YAML configuration loading
│   │   └── credentials.rs   # Telegram API credentials
│   ├── telegram/            # Telegram client (grammers wrapper)
│   │   ├── client.rs        # Client lifecycle and connection
│   │   ├── auth.rs          # Authentication flow
│   │   ├── messages.rs      # Message operations
│   │   ├── chats.rs         # Chat/dialog operations
│   │   ├── media.rs         # Media handling
│   │   ├── updates.rs       # Real-time update streaming
│   │   └── error.rs         # Error types
│   ├── ui/                  # Ratatui UI layer
│   │   ├── app.rs           # Main state machine and event loop
│   │   ├── keys.rs          # Key binding system
│   │   ├── styles.rs        # Nord color theme
│   │   └── components/      # Reusable UI components
│   ├── cache/               # In-memory caching
│   ├── types/               # Shared domain types
│   └── utils/               # Time and text formatting helpers
├── Cargo.toml
├── config.example.yaml
└── README.md
```

### Building

```bash
cargo build            # Debug build
cargo build --release  # Release build (optimized, stripped)
cargo run              # Run directly
```

### Testing

```bash
cargo test                    # Run all tests
cargo test -- --nocapture     # With output
cargo clippy --all-targets    # Lint
cargo fmt --check             # Format check
```

## Architecture

Ithil follows an event-driven architecture with async operations handled by Tokio:

- **Event Loop**: Crossterm events and Telegram updates processed in a unified async loop
- **State Machine**: `App` struct manages UI state transitions (auth, main view, settings)
- **Telegram Client**: Grammers-based MTProto client with async message/update handling
- **Cache Layer**: Thread-safe in-memory cache (`Arc<RwLock>`) for messages, chats, and users

### Data Flow

```
Telegram Server (MTProto)
        ↓
grammers-client
        ↓
Update Stream → Cache → App State → Ratatui Renderer → Terminal
        ↑                    ↓
        └─── User Input ─────┘
```

## Dependencies

### Core Libraries
- [Ratatui](https://ratatui.rs) - Terminal UI framework
- [Crossterm](https://github.com/crossterm-rs/crossterm) - Cross-platform terminal manipulation
- [Tokio](https://tokio.rs) - Async runtime
- [grammers-client](https://github.com/nicegram/nicegram-grammers) - Telegram MTProto client
- [serde](https://serde.rs) / serde_yaml - Configuration serialization
- [clap](https://clap.rs) - CLI argument parsing
- [tracing](https://tracing.rs) - Logging and diagnostics

### Why grammers?

- **Pure Rust**: No C/C++ dependencies, easy cross-compilation
- **Type-safe**: Strongly typed Telegram API
- **Async**: Built for Tokio with native async/await
- **Lightweight**: No heavy runtime, small binary size
- **Direct MTProto**: Direct protocol implementation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines

1. Follow Rust best practices and idioms
2. Run `cargo clippy` and `cargo fmt` before committing
3. Write tests for new features
4. Use the Nord color scheme for styling

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Telegram](https://telegram.org) for the messaging platform
- [Ratatui](https://ratatui.rs) for the TUI framework
- [grammers](https://github.com/nicegram/nicegram-grammers) for the Telegram client library
- [Nord](https://www.nordtheme.com/) for the color scheme

## Support

- **Issues**: https://github.com/lvcasx1/ithil/issues
- **Discussions**: https://github.com/lvcasx1/ithil/discussions

## Disclaimer

This project is not affiliated with Telegram or its parent company. It is an independent client built using the official Telegram API.

---

**Built with Rust and Ratatui**
