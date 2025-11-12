# Contributing to Ithil

Thank you for your interest in contributing to Ithil! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, constructive, and professional. We're all here to build something great together.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git
- Telegram API credentials from [my.telegram.org](https://my.telegram.org)

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git fork https://github.com/lvcasx1/ithil.git
   cd ithil
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up configuration**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your API credentials
   ```

4. **Install development tools**
   ```bash
   # Hot-reload tool for development
   go install github.com/cosmtrek/air@latest

   # Linter
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

5. **Run the application**
   ```bash
   # With hot-reload
   air

   # Or directly
   go run cmd/ithil/main.go
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   # Run all tests
   go test ./...

   # Run tests with coverage
   go test -cover ./...

   # Run tests with race detection
   go test -race ./...

   # Run linter
   golangci-lint run
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Guidelines

We use conventional commits for clear history:

- `feat: add new feature`
- `fix: fix a bug`
- `docs: update documentation`
- `test: add or update tests`
- `refactor: code refactoring`
- `style: code style changes`
- `chore: maintenance tasks`

## Code Style Guidelines

### Go Code

- Follow standard Go conventions ([Effective Go](https://golang.org/doc/effective_go.html))
- Use `gofmt` for formatting (automatically handled by most editors)
- Write meaningful variable and function names
- Add comments for all exported functions
- Keep functions small and focused
- Use meaningful error messages

### Project-Specific Guidelines

1. **Package Organization**
   - `cmd/` - Application entry points
   - `internal/` - Private application code
   - `pkg/` - Public, reusable packages
   - `internal/telegram/` - Telegram API operations
   - `internal/ui/` - TUI components and models

2. **Bubbletea Models**
   - Follow the Elm Architecture (Model-Update-View)
   - Keep models immutable - return new state from Update()
   - Use tea.Cmd for side effects
   - Handle all message types appropriately

3. **Styling**
   - Use Lipgloss for all terminal styling
   - Follow the Nord color scheme (defined in `styles/styles.go`)
   - Avoid raw ANSI codes
   - Make UI responsive to terminal size

4. **Telegram Integration**
   - Use gotd/td library (not TDLib)
   - Handle all error cases appropriately
   - Add proper logging with slog
   - Cache data when appropriate
   - Use the gaps manager for reliable updates

5. **Testing**
   - Write tests for new features
   - Aim for >70% code coverage
   - Include edge cases and error conditions
   - Use table-driven tests where appropriate
   - Mock external dependencies

## Project Architecture

### Key Components

1. **Telegram Client** (`internal/telegram/`)
   - Wraps gotd/td with high-level operations
   - Handles authentication, messages, chats, media
   - Manages update processing

2. **UI Layer** (`internal/ui/`)
   - Bubbletea models for each pane
   - Reusable components
   - Keyboard bindings
   - Styling definitions

3. **Type System** (`pkg/types/`)
   - Shared type definitions
   - Conversion functions

4. **Cache Layer** (`internal/cache/`)
   - In-memory caching for performance
   - Thread-safe operations

### Adding New Features

#### Adding a New Message Type

1. Add enum to `pkg/types/types.go` (MessageType)
2. Add conversion in `internal/telegram/messages.go` (convertMessageContent)
3. Add rendering in `internal/ui/components/message.go` (renderContent)
4. Add media viewer support if needed

#### Adding a New Keyboard Shortcut

1. Define in `internal/ui/keys/keymap.go`
2. Handle in appropriate model
3. Add to help modal
4. Update README

#### Adding a New Telegram Operation

1. Add method to relevant file in `internal/telegram/`
2. Use gotd/td API correctly
3. Add error handling and logging
4. Update cache if needed
5. Wire up to UI

## Testing Guidelines

### Unit Tests

- Place tests in `*_test.go` files next to the code
- Test public APIs and exported functions
- Mock external dependencies
- Use meaningful test names

### Integration Tests

- Test end-to-end workflows
- Use test fixtures when needed
- Clean up test data

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/cache

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...
```

## Documentation

### Code Documentation

- Add package comments at the top of each package
- Document all exported functions, types, and variables
- Use complete sentences
- Include usage examples for complex functions

### User Documentation

- Update README.md for user-facing changes
- Add entries to CHANGELOG.md
- Update example configuration if needed

## Pull Request Process

1. **Before submitting:**
   - Ensure all tests pass
   - Run the linter
   - Update documentation
   - Add CHANGELOG entry for notable changes
   - Rebase on latest main if needed

2. **PR Description:**
   - Clearly describe what the PR does
   - Reference any related issues
   - Include screenshots for UI changes
   - List any breaking changes

3. **Review Process:**
   - Address review comments promptly
   - Keep discussions focused and professional
   - Update PR based on feedback
   - Maintain a clean commit history

4. **After Merge:**
   - Delete your feature branch
   - Celebrate your contribution! 🎉

## Getting Help

- **Questions:** Open a discussion on GitHub
- **Bugs:** Open an issue with reproduction steps
- **Features:** Open an issue to discuss before implementing
- **Telegram:** Join our community (coming soon!)

## Project Resources

- **Repository:** https://github.com/lvcasx1/ithil
- **Issues:** https://github.com/lvcasx1/ithil/issues
- **Discussions:** https://github.com/lvcasx1/ithil/discussions
- **gotd/td Docs:** https://core.telegram.org/methods
- **Bubbletea:** https://github.com/charmbracelet/bubbletea

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to Ithil! 🌙
