package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test app defaults
	if config.App.Name != "Ithil" {
		t.Errorf("Expected app name 'Ithil', got '%s'", config.App.Name)
	}
	if config.App.Version != "0.1.0" {
		t.Errorf("Expected version '0.1.0', got '%s'", config.App.Version)
	}

	// Test UI layout defaults sum to 100
	totalWidth := config.UI.Layout.ChatListWidth +
		config.UI.Layout.ConversationWidth +
		config.UI.Layout.InfoWidth
	if totalWidth != 100 {
		t.Errorf("Expected layout widths to sum to 100, got %d", totalWidth)
	}

	// Test behavior defaults
	if config.UI.Behavior.AutoDownloadLimit != 5242880 {
		t.Errorf("Expected auto download limit 5242880, got %d", config.UI.Behavior.AutoDownloadLimit)
	}

	// Test cache defaults
	if config.Cache.MaxMessagesPerChat != 1000 {
		t.Errorf("Expected max messages 1000, got %d", config.Cache.MaxMessagesPerChat)
	}
	if config.Cache.MaxMediaSize != 104857600 {
		t.Errorf("Expected max media size 104857600, got %d", config.Cache.MaxMediaSize)
	}
}

func TestValidate_Success(t *testing.T) {
	config := DefaultConfig()
	config.Telegram.APIID = "12345"
	config.Telegram.APIHash = "abcdef123456"

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
}

func TestValidate_MissingAPIID(t *testing.T) {
	config := DefaultConfig()
	config.Telegram.APIID = ""
	config.Telegram.APIHash = "abcdef123456"

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for missing API ID")
	}
}

func TestValidate_MissingAPIHash(t *testing.T) {
	config := DefaultConfig()
	config.Telegram.APIID = "12345"
	config.Telegram.APIHash = ""

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for missing API hash")
	}
}

func TestValidate_PlaceholderAPIID(t *testing.T) {
	config := DefaultConfig()
	config.Telegram.APIID = "YOUR_API_ID"
	config.Telegram.APIHash = "abcdef123456"

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for placeholder API ID")
	}
}

func TestValidate_InvalidLayoutWidths(t *testing.T) {
	config := DefaultConfig()
	config.Telegram.APIID = "12345"
	config.Telegram.APIHash = "abcdef123456"
	config.UI.Layout.ChatListWidth = 30
	config.UI.Layout.ConversationWidth = 30
	config.UI.Layout.InfoWidth = 30 // Sum is 90, not 100

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for invalid layout widths")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Create a config with known values
	originalConfig := DefaultConfig()
	originalConfig.Telegram.APIID = "test-id"
	originalConfig.Telegram.APIHash = "test-hash"
	originalConfig.UI.Theme = "nord"
	originalConfig.Cache.MaxMessagesPerChat = 500

	// Save the config
	err := originalConfig.SaveConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check if file was created
	if !fileExists(configPath) {
		t.Fatal("Config file was not created")
	}

	// Load the config back
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values match
	if loadedConfig.Telegram.APIID != "test-id" {
		t.Errorf("Expected API ID 'test-id', got '%s'", loadedConfig.Telegram.APIID)
	}
	if loadedConfig.Telegram.APIHash != "test-hash" {
		t.Errorf("Expected API hash 'test-hash', got '%s'", loadedConfig.Telegram.APIHash)
	}
	if loadedConfig.UI.Theme != "nord" {
		t.Errorf("Expected theme 'nord', got '%s'", loadedConfig.UI.Theme)
	}
	if loadedConfig.Cache.MaxMessagesPerChat != 500 {
		t.Errorf("Expected max messages 500, got %d", loadedConfig.Cache.MaxMessagesPerChat)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	config, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Expected default config when file doesn't exist, got error: %v", err)
	}
	if config == nil {
		t.Fatal("Expected default config, got nil")
	}
	// Should return default config
	if config.App.Name != "Ithil" {
		t.Errorf("Expected default config, got different app name: %s", config.App.Name)
	}
}

func TestExpandPaths(t *testing.T) {
	config := DefaultConfig()
	homeDir, _ := os.UserHomeDir()

	// Set paths with tilde
	config.Telegram.SessionFile = "~/test/session.json"
	config.Cache.MediaDirectory = "~/test/media"

	config.expandPaths()

	// Check if tilde was expanded
	expectedSessionFile := filepath.Join(homeDir, "test/session.json")
	if config.Telegram.SessionFile != expectedSessionFile {
		t.Errorf("Expected session file '%s', got '%s'", expectedSessionFile, config.Telegram.SessionFile)
	}

	expectedMediaDir := filepath.Join(homeDir, "test/media")
	if config.Cache.MediaDirectory != expectedMediaDir {
		t.Errorf("Expected media directory '%s', got '%s'", expectedMediaDir, config.Cache.MediaDirectory)
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")

	if fileExists(tempFile) {
		t.Error("File should not exist yet")
	}

	// Create the file
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !fileExists(tempFile) {
		t.Error("File should exist now")
	}
}
