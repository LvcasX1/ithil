// Package app provides the main application logic and state management.
package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/lvcasx1/ithil/internal/telegram"
)

// Application represents the main application state.
type Application struct {
	Config         *Config
	TelegramClient *telegram.Client
	Logger         *slog.Logger
	Context        context.Context
	Cancel         context.CancelFunc
}

// New creates a new Application instance.
func New(configPath string) (*Application, error) {
	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Ensure required directories exist
	if err := config.EnsureDirectories(); err != nil {
		return nil, err
	}

	// Setup logger
	logger := setupLogger(config.Logging)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		Config:  config,
		Logger:  logger,
		Context: ctx,
		Cancel:  cancel,
	}

	return app, nil
}

// InitTelegram initializes the Telegram client.
// TODO: This will be implemented when the telegram package is complete.
func (a *Application) InitTelegram() error {
	// This will be implemented when we create the telegram client
	// For now, it's just a placeholder
	return nil
}

// Shutdown performs graceful shutdown of the application.
func (a *Application) Shutdown() error {
	a.Logger.Info("Shutting down application...")
	a.Cancel()

	// Close Telegram client if initialized
	if a.TelegramClient != nil {
		// TODO: Implement proper shutdown when telegram client is ready
		a.Logger.Info("Closing Telegram client...")
	}

	a.Logger.Info("Application shutdown complete")
	return nil
}

// setupLogger creates and configures a logger based on the logging configuration.
func setupLogger(config LoggingConfig) *slog.Logger {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if config.File != "" {
		// Try to create log file
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			handler = slog.NewTextHandler(file, opts)
		} else {
			// Fall back to stdout if file creation fails
			handler = slog.NewTextHandler(os.Stdout, opts)
		}
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
