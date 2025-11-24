// Ithil - A Terminal User Interface for Telegram
//
// Ithil (Sindarin for "moon") is a feature-rich TUI Telegram client
// built with Go and Bubbletea, designed to bring the full Telegram
// experience to the terminal.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvcasx1/ithil/internal/app"
	"github.com/lvcasx1/ithil/internal/telegram"
	"github.com/lvcasx1/ithil/internal/ui/models"
)

var (
	version     = "0.1.4"
	configPath  = flag.String("config", "", "Path to config file")
	showVersion = flag.Bool("version", false, "Show version information")
	showHelp    = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("Ithil version %s\n", version)
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Create application
	application, err := app.New(*configPath)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Check for credential change marker and clear session if needed
	checkAndClearSessionOnCredentialChange(application)

	// Get API credentials (default or custom)
	apiID, apiHash := app.GetAPICredentials(application.Config)

	// Log credential type being used
	if app.IsUsingDefaultCredentials(application.Config) {
		log.Println("Using default API credentials")
		log.Println("For enhanced privacy, configure custom credentials in Settings (Ctrl+,)")
	} else {
		log.Println("Using custom API credentials")
	}

	// Validate configuration
	if err := application.Config.Validate(); err != nil {
		log.Printf("Warning: Configuration validation failed: %v", err)
	}

	// Initialize Telegram client with appropriate credentials
	telegramConfig := &telegram.Config{
		APIID:             strconv.Itoa(apiID),
		APIHash:           apiHash,
		SessionFile:       application.Config.Telegram.SessionFile,
		DatabaseDirectory: application.Config.Telegram.DatabaseDirectory,
	}
	client, err := telegram.New(
		telegramConfig,
		application.Logger,
		application.Context,
	)
	if err != nil {
		log.Fatalf("Failed to create Telegram client: %v", err)
	}

	// Start Telegram client
	if err := client.Start(); err != nil {
		application.Logger.Error("Failed to start Telegram client", "error", err)
	}

	// Create main TUI model
	model := models.NewMainModel(application.Config, client)

	// Create Bubbletea program with alternate screen
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Setup graceful shutdown
	go func() {
		<-application.Context.Done()
		p.Quit()
	}()

	// Run the program
	application.Logger.Info("Starting Ithil TUI", "version", version)
	if _, err := p.Run(); err != nil {
		application.Logger.Error("Error running TUI", "error", err)
		os.Exit(1)
	}

	// Shutdown
	if err := application.Shutdown(); err != nil {
		application.Logger.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}

	application.Logger.Info("Ithil shutdown complete")
}

// printHelp prints the help message.
func printHelp() {
	fmt.Printf(`Ithil - A Terminal User Interface for Telegram (version %s)

USAGE:
    ithil [OPTIONS]

OPTIONS:
    -config <path>    Path to configuration file (default: auto-detect)
    -version          Show version information
    -help             Show this help message

CONFIGURATION:
    Ithil looks for configuration files in the following locations:
    1. ./config.yaml
    2. ~/.config/ithil/config.yaml
    3. ~/.ithil.yaml

    Copy config.example.yaml to config.yaml and edit it with your settings.
    You will need to obtain Telegram API credentials from https://my.telegram.org

KEYBOARD SHORTCUTS:
    Global:
        Ctrl+C, Ctrl+Q  - Quit application
        ?               - Toggle help
        Tab             - Next pane
        Shift+Tab       - Previous pane
        Ctrl+1/2/3      - Focus chat list/conversation/sidebar
        Ctrl+S          - Toggle sidebar

    Chat List:
        j/k, ↑/↓        - Navigate chats
        Enter           - Open chat
        p               - Pin/unpin chat
        m               - Mute/unmute chat
        r               - Mark as read
        /               - Search

    Conversation:
        j/k, ↑/↓        - Scroll messages
        i               - Focus input field
        r               - Reply to message
        e               - Edit message
        d               - Delete message

    Message Input:
        Enter           - Send message
        Shift+Enter     - New line
        Esc             - Cancel reply/edit

DOCUMENTATION:
    For more information, visit: https://github.com/lvcasx1/ithil

`, version)
}

// checkAndClearSessionOnCredentialChange checks for credential change marker
// and clears the session if credentials have been changed.
func checkAndClearSessionOnCredentialChange(application *app.Application) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	markerPath := filepath.Join(homeDir, ".config", "ithil", ".credential-change")
	if _, err := os.Stat(markerPath); err == nil {
		// Marker file exists, clear session
		log.Println("Credential change detected, clearing session for privacy...")

		// Clear session file
		sessionPath := application.Config.Telegram.SessionFile
		if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Could not remove session file: %v\n", err)
		}

		// Clear auth data file
		authPath := sessionPath + ".auth"
		if err := os.Remove(authPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Could not remove auth data file: %v\n", err)
		}

		// Remove marker file
		if err := os.Remove(markerPath); err != nil {
			log.Printf("Warning: Could not remove credential change marker: %v\n", err)
		}

		log.Println("Session cleared successfully. You will need to log in again.")
	}
}
