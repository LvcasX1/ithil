# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Ithil** is a Terminal User Interface (TUI) Telegram client built with Go, Bubbletea, and gotd/td. It implements the full Telegram MTProto protocol for real-time messaging in the terminal with ~8,500 lines of well-structured Go code.

## Build & Development Commands

### Building
```bash
# Standard build
go build -o bin/ithil ./cmd/ithil

# Run directly
go run cmd/ithil/main.go

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o bin/ithil-linux-amd64 ./cmd/ithil
GOOS=darwin GOARCH=amd64 go build -o bin/ithil-darwin-amd64 ./cmd/ithil
GOOS=darwin GOARCH=arm64 go build -o bin/ithil-darwin-arm64 ./cmd/ithil
GOOS=windows GOARCH=amd64 go build -o bin/ithil-windows-amd64.exe ./cmd/ithil
```

### Development with Hot-Reload
```bash
# Install Air (if not already installed)
go install github.com/cosmtrek/air@latest

# Run with hot-reload (uses .air.toml config)
air
```

Air is configured to:
- Build to `./tmp/main`
- Watch `.go`, `.tpl`, `.tmpl`, `.html` files
- Exclude `_test.go`, `tmp/`, `vendor/`, `testdata/` directories
- Log build errors to `build-errors.log`

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

### Configuration Setup
```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit with your Telegram API credentials from https://my.telegram.org
# Configuration is searched in order:
# 1. ./config.yaml
# 2. ~/.config/ithil/config.yaml
# 3. ~/.ithil.yaml
```

## Architecture Overview

### High-Level Pattern: Elm Architecture (Model-Update-View)

Ithil uses the **Elm Architecture** pattern via Bubbletea:
- **Models**: Represent application state (immutable)
- **Update**: Handle messages and return new state
- **View**: Render the current state to the terminal

### Critical Data Flow

```
Telegram Server (MTProto)
        ↓
gotd/td Client (internal/telegram/client.go)
        ↓
UpdateHandler (internal/telegram/updates.go) → Cache (internal/cache/cache.go)
        ↓                                           ↓
   UI Models (internal/ui/models/) → Bubbletea Renderer → Terminal
        ↑
   User Input (keyboard)
```

### Key Architectural Components

#### 1. Telegram Client Layer (`internal/telegram/`)
- **`client.go`**: Main wrapper around gotd/td with MTProto implementation
- **`updates.go`**: Real-time update handler with gaps manager for zero message loss
- **`auth.go`**: Multi-step authentication flow (phone → code → 2FA → registration)
- **`messages.go`**: Message operations (send, edit, get history)
- **`chats.go`**: Chat operations (list, get, pin, mute, archive)
- **`session.go`**: Session storage with automatic recovery
- **`media.go`**: Media download/upload manager

**Critical**: The `UpdateHandler` uses a `gaps.Manager` to ensure no updates are lost. Updates flow through:
1. gotd/td receives MTProto updates
2. `gaps.Manager` handles sequence tracking and gap recovery
3. `UpdateHandler.Handle()` processes updates and converts to internal types
4. Updates are cached and sent to UI via channels

#### 2. UI Layer (`internal/ui/`)
- **`models/main.go`**: Root Bubbletea model that coordinates all sub-models
- **`models/auth.go`**: Authentication screen (phone, code, 2FA)
- **`models/chatlist.go`**: Chat list pane with search/filter
- **`models/conversation.go`**: Message view, input, and editing
- **`models/sidebar.go`**: Chat/user info display
- **`components/`**: Reusable UI components (statusbar, help modal, media viewer)
- **`keys/keymap.go`**: Keyboard binding definitions (supports vim mode)
- **`styles/styles.go`**: Lipgloss styling with Nord color scheme

**Critical**: Focus management is controlled by `FocusPane` enum in `main.go`. Each pane only responds to input when focused.

#### 3. Type System (`pkg/types/`)
- **`types.go`**: All shared type definitions
- **Key types**: `User`, `Chat`, `Message`, `MessageContent`, `Media`, `Update`
- **Message types**: Text, Photo, Video, Voice, VideoNote, Audio, Document, Sticker, Animation, Location, Contact, Poll
- **Entity types**: Bold, Italic, Code, Pre, TextURL, Mention, Hashtag, etc.

**Critical**: All types use Go's type safety. `MessageType` and `EntityType` are enums (iota). `Chat.AccessHash` is required for all API calls to users and channels.

#### 4. Cache Layer (`internal/cache/`)
- In-memory cache for messages, chats, and users
- Thread-safe with `sync.RWMutex`
- Configurable limits per chat
- Automatically trims old messages when limit exceeded

**Critical**: The cache must be set on the client via `client.SetCache()` before starting, so the update handler can cache users during update processing.

#### 5. Media System (`internal/media/`)
- **`protocol_detector.go`**: Automatic graphics protocol detection (Kitty, Sixel, Unicode, ASCII)
- **`kitty_renderer.go`**: Kitty graphics protocol implementation (pixel-perfect rendering)
- **`sixel_renderer.go`**: Sixel graphics protocol implementation (256-color rendering)
- **`mosaic_renderer.go`**: Unicode half-block image rendering (true color)
- **`image_renderer.go`**: ASCII art image rendering (fallback)
- **`audio_player_cgo.go`**: Beep-based audio playback for voice/audio messages
- **`audio_renderer.go`**: Waveform visualization and playback UI

**Critical**: Audio playback uses Beep library with goroutines. Must properly clean up resources when stopping playback to avoid memory leaks.

### Message Flow Example

**Sending a message:**
1. User types in `ConversationModel.input` (textarea)
2. User presses Enter → `conversation.go` calls `client.SendMessage()`
3. `messages.go` converts to gotd/td types and sends via MTProto
4. Message appears immediately (optimistic UI)
5. Server confirms → update arrives → cache updated → UI re-rendered with confirmed message

**Receiving a message:**
1. Telegram server sends update via MTProto
2. `gaps.Manager` receives and tracks sequence
3. `UpdateHandler.Handle()` processes `tg.UpdateNewMessage`
4. Message converted to `types.Message` and cached
5. Update sent via channel to `MainModel`
6. `MainModel` routes to appropriate sub-model
7. Sub-model updates state and triggers re-render

### Authentication State Machine

Authentication follows this state flow (defined in `types.AuthState`):
1. `AuthStateWaitPhoneNumber` → user enters phone
2. `AuthStateWaitCode` → user enters verification code
3. If 2FA enabled: `AuthStateWait2FA` → user enters password
4. If new user: `AuthStateWaitRegistration` → user enters name
5. `AuthStateReady` → authenticated, load chat list

**Critical**: Session data is persisted in `SessionStorage`. Invalid sessions trigger `AUTH_RESTART` which resets to `AuthStateWaitPhoneNumber`.

## gotd/td Specifics

### Why gotd/td instead of TDLib?
- Pure Go (no CGo dependencies)
- Type-safe generated API
- Built-in gaps manager for reliable updates
- Smaller binary size
- Direct MTProto implementation

### Common gotd Patterns

**Creating InputPeer (required for most API calls):**
```go
// Use client.chatToInputPeer(chat) which handles all chat types
inputPeer, err := c.chatToInputPeer(chat)
// Returns:
// - *tg.InputPeerUser for private chats (needs AccessHash)
// - *tg.InputPeerChat for groups
// - *tg.InputPeerChannel for supergroups/channels (needs AccessHash)
```

**Handling API responses:**
Most gotd API calls return interface types that need type switching:
```go
result, err := c.api.MessagesGetHistory(ctx, req)
switch h := result.(type) {
case *tg.MessagesMessages:
    // Full message list
case *tg.MessagesMessagesSlice:
    // Partial message list (with total count)
case *tg.MessagesChannelMessages:
    // Channel-specific messages
}
```

**Update types to handle:**
- `tg.UpdateNewMessage` - New incoming message
- `tg.UpdateEditMessage` - Message was edited
- `tg.UpdateDeleteMessages` - Messages deleted
- `tg.UpdateReadHistoryInbox` - Messages marked as read
- `tg.UpdateUserTyping` - User is typing
- `tg.UpdateUserStatus` - User online/offline status
- `tg.UpdateChatUserTyping` - Group typing indicator

### Common Issues

1. **AccessHash is zero**: For private chats and channels, `AccessHash` must be obtained from initial chat retrieval and stored. Without it, API calls fail with `PEER_ID_INVALID`.

2. **Updates not arriving**: Ensure `gaps.Manager` is set as both the `UpdateHandler` AND in the middleware:
   ```go
   opts := telegram.Options{
       UpdateHandler: client.gaps,
       Middlewares: []telegram.Middleware{
           updhook.UpdateHook(client.gaps.Handle),
       },
   }
   ```

3. **Media files**: Use `MediaManager.DownloadMedia()` which handles all document types. File paths are stored in `Media.LocalPath` after download.

### Navigation System

**Centered Message Selection** (conversation pane):
- `j`/`k` (or `↓`/`↑`) moves a **selection cursor** (shown as `▶`) through messages
- **The selected message stays centered** in the viewport (like modern text editors)
- This provides maximum visibility - you always see context above and below the selection
- All message actions (`r` for reply, `e` for edit, `Enter` for media) operate on the selected message
- Selection is initialized to the last message when opening a chat
- `Ctrl+U`/`Ctrl+D` and other scroll keys still provide manual viewport control

**Key implementation details:**
- `selectedMsgIdx` in `ConversationModel` tracks the currently selected message
- `messageLinesMap` tracks actual rendered line positions (not estimates)
- `scrollToSelectedMessage()` centers the selected message: `targetOffset = messageStartLine - (viewportHeight / 2)`
- `NewMessageComponentWithSelection()` renders the `▶` indicator for selected messages
- Selection wraps around (top → bottom, bottom → top) for continuous navigation
- Bounds checking prevents scrolling past content start/end

## Development Workflow

### Adding a New Message Type

1. Add enum to `pkg/types/types.go`:
   ```go
   const (
       MessageTypeText MessageType = iota
       // ... existing types
       MessageTypeNewType
   )
   ```

2. Add conversion in `internal/telegram/messages.go` `convertMessageContent()`:
   ```go
   case *tg.MessageMediaNewType:
       content.Type = types.MessageTypeNewType
       // Convert fields
   ```

3. Add rendering in `internal/ui/components/message.go` `renderContent()`:
   ```go
   case types.MessageTypeNewType:
       return m.renderMediaContent("New Type", content.Caption, content.Entities)
   ```

4. Add media viewer support in `internal/ui/components/mediaviewer.go` if needed

### Adding a New Keyboard Shortcut

1. Define key binding in `internal/ui/keys/keymap.go`:
   ```go
   type KeyMap struct {
       // ... existing
       NewAction key.Binding
   }

   NewAction: key.NewBinding(
       key.WithKeys("x"),
       key.WithHelp("x", "new action"),
   ),
   ```

2. Handle in appropriate model (e.g., `conversation.go`):
   ```go
   case key.Matches(msg, m.keyMap.NewAction):
       // Handle action
       return m, nil
   ```

3. Add to help modal in `internal/ui/components/helpmodal.go`

### Adding a New Chat Operation

1. Add method to `internal/telegram/chats.go`:
   ```go
   func (c *Client) NewOperation(chat *types.Chat) error {
       inputPeer, err := c.chatToInputPeer(chat)
       if err != nil {
           return err
       }
       // Call gotd API
       _, err = c.api.SomeMethod(c.ctx, &tg.SomeRequest{
           Peer: inputPeer,
       })
       return err
   }
   ```

2. Wire up in UI model (e.g., `chatlist.go`)

## Code Style

- Follow Go best practices and idioms
- Use meaningful variable and function names
- Add comments for all exported functions
- Use Nord color scheme for styling (defined in `styles/styles.go`)
- Prefer Lipgloss styling over ANSI codes
- Keep models immutable - return new state from Update()
- Use structured logging with `slog`

## Testing Notes

- Telegram API credentials required for integration tests
- Mock the `telegram.Client` interface for unit tests
- Use `cache.Cache` directly in tests without real Telegram client
- Bubbletea models can be tested by sending messages and checking state

## Current Development Phase

**Phase 3 (Rich Features)** - Completed:
- ✅ Media download/display with automatic protocol detection
- ✅ Kitty graphics protocol support (pixel-perfect rendering)
- ✅ Sixel graphics protocol support (256-color rendering)
- ✅ Unicode mosaic rendering with true color
- ✅ ASCII art fallback for maximum compatibility
- ❌ Message reactions (not started)
- ❌ Message forwarding (stub only)
- ❌ Message deletion (stub only)

**Phase 4 (Advanced Features)** - Completed:
- ✅ Voice messages with audio playback (Beep library)
- ✅ Video messages (round video notes)
- ✅ Chat pinning/muting/archiving

**Phase 4 (Advanced Features)** - Remaining:
- Inline bots (complex - requires inline query UI)
- Secret chats (very complex - requires E2E encryption)

**Phase 5 (Polish)** - Planned:
- Notifications, search, multiple themes, performance optimizations

## Important Files to Read First

When working on a new feature, start by reading these in order:
1. `pkg/types/types.go` - Understand the type system
2. `internal/telegram/client.go` - See how Telegram client works
3. `internal/ui/models/main.go` - Understand UI architecture and focus management
4. Relevant model in `internal/ui/models/` - Find the specific pane you're modifying
5. `internal/telegram/updates.go` - Understand real-time update flow

## Dependencies

- **Bubbletea**: TUI framework (Elm Architecture)
- **Lipgloss**: Terminal styling and layout
- **Bubbles**: Pre-built UI components (textarea, viewport, list)
- **gotd/td**: Pure Go Telegram MTProto implementation
- **Beep**: Audio playback library (MP3, OGG, WAV support)
