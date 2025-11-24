// Package app provides application-level functionality including configuration management.
package app

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	App           AppConfig          `yaml:"app"`
	Telegram      TelegramConfig     `yaml:"telegram"`
	UI            UIConfig           `yaml:"ui"`
	Notifications NotificationConfig `yaml:"notifications"`
	Privacy       PrivacyConfig      `yaml:"privacy"`
	Cache         CacheConfig        `yaml:"cache"`
	Logging       LoggingConfig      `yaml:"logging"`
}

// AppConfig contains general application settings.
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// TelegramConfig contains Telegram API credentials and settings.
type TelegramConfig struct {
	UseDefaultCredentials bool   `yaml:"use_default_credentials"` // Use built-in credentials (default: true)
	APIID                 string `yaml:"api_id"`                  // Custom API ID (optional)
	APIHash               string `yaml:"api_hash"`                // Custom API Hash (optional)
	SessionFile           string `yaml:"session_file"`
	DatabaseDirectory     string `yaml:"database_directory"`
}

// UIConfig contains user interface settings.
type UIConfig struct {
	Theme      string           `yaml:"theme"`
	Layout     LayoutConfig     `yaml:"layout"`
	Appearance AppearanceConfig `yaml:"appearance"`
	Behavior   BehaviorConfig   `yaml:"behavior"`
	Keyboard   KeyboardConfig   `yaml:"keyboard"`
}

// LayoutConfig defines the layout settings.
type LayoutConfig struct {
	ChatListWidth     int  `yaml:"chat_list_width"`
	ConversationWidth int  `yaml:"conversation_width"`
	InfoWidth         int  `yaml:"info_width"`
	ShowInfoPane      bool `yaml:"show_info_pane"`
}

// AppearanceConfig defines appearance settings.
type AppearanceConfig struct {
	ShowAvatars          bool   `yaml:"show_avatars"`
	ShowStatusBar        bool   `yaml:"show_status_bar"`
	DateFormat           string `yaml:"date_format"`
	RelativeTimestamps   bool   `yaml:"relative_timestamps"`
	MessagePreviewLength int    `yaml:"message_preview_length"`
}

// BehaviorConfig defines behavior settings.
type BehaviorConfig struct {
	SendOnEnter       bool   `yaml:"send_on_enter"`
	AutoDownloadLimit int64  `yaml:"auto_download_limit"`
	MarkReadOnScroll  bool   `yaml:"mark_read_on_scroll"`
	EmojiStyle        string `yaml:"emoji_style"`
}

// KeyboardConfig defines keyboard settings.
type KeyboardConfig struct {
	VimMode        bool              `yaml:"vim_mode"`
	CustomBindings map[string]string `yaml:"custom_bindings"`
}

// NotificationConfig contains notification settings.
type NotificationConfig struct {
	Enabled    bool    `yaml:"enabled"`
	Sound      bool    `yaml:"sound"`
	Desktop    bool    `yaml:"desktop"`
	MutedChats []int64 `yaml:"muted_chats"`
}

// PrivacyConfig contains privacy settings.
type PrivacyConfig struct {
	ShowOnlineStatus bool `yaml:"show_online_status"`
	ShowReadReceipts bool `yaml:"show_read_receipts"`
	ShowTyping       bool `yaml:"show_typing"`
	StealthMode      bool `yaml:"stealth_mode"`
}

// CacheConfig contains cache settings.
type CacheConfig struct {
	MaxMessagesPerChat int    `yaml:"max_messages_per_chat"`
	MaxMediaSize       int64  `yaml:"max_media_size"`
	MediaDirectory     string `yaml:"media_directory"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "ithil")
	cacheDir := filepath.Join(homeDir, ".cache", "ithil")

	return &Config{
		App: AppConfig{
			Name:    "Ithil",
			Version: "0.1.3",
		},
		Telegram: TelegramConfig{
			UseDefaultCredentials: true, // Default to built-in credentials
			APIID:                 "",   // Empty = use defaults
			APIHash:               "",   // Empty = use defaults
			SessionFile:           filepath.Join(configDir, "session.json"),
			DatabaseDirectory:     filepath.Join(configDir, "tdlib"),
		},
		UI: UIConfig{
			Theme: "dark",
			Layout: LayoutConfig{
				ChatListWidth:     25,
				ConversationWidth: 50,
				InfoWidth:         25,
				ShowInfoPane:      true,
			},
			Appearance: AppearanceConfig{
				ShowAvatars:          true,
				ShowStatusBar:        true,
				DateFormat:           "12h",
				RelativeTimestamps:   true,
				MessagePreviewLength: 50,
			},
			Behavior: BehaviorConfig{
				SendOnEnter:       true,
				AutoDownloadLimit: 5242880, // 5MB
				MarkReadOnScroll:  true,
				EmojiStyle:        "unicode",
			},
			Keyboard: KeyboardConfig{
				VimMode:        true,
				CustomBindings: make(map[string]string),
			},
		},
		Notifications: NotificationConfig{
			Enabled:    true,
			Sound:      true,
			Desktop:    false,
			MutedChats: []int64{},
		},
		Privacy: PrivacyConfig{
			ShowOnlineStatus: true,
			ShowReadReceipts: true,
			ShowTyping:       true,
			StealthMode:      false,
		},
		Cache: CacheConfig{
			MaxMessagesPerChat: 1000,
			MaxMediaSize:       104857600, // 100MB
			MediaDirectory:     filepath.Join(cacheDir, "media"),
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  filepath.Join(configDir, "ithil.log"),
		},
	}
}

// LoadConfig loads configuration from a YAML file.
// If the file doesn't exist, it returns the default configuration.
func LoadConfig(path string) (*Config, error) {
	// Expand tilde to home directory
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// If no path provided or file doesn't exist, try default locations
	if path == "" || !fileExists(path) {
		path = findConfigFile()
	}

	// If still no config file found, return default config
	if path == "" || !fileExists(path) {
		return DefaultConfig(), nil
	}

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand paths
	config.expandPaths()

	return config, nil
}

// SaveConfig saves the configuration to a YAML file.
func (c *Config) SaveConfig(path string) error {
	// Expand tilde to home directory
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Only validate custom credentials if not using defaults
	if !c.Telegram.UseDefaultCredentials {
		if c.Telegram.APIID == "" || c.Telegram.APIID == "YOUR_API_ID" {
			return fmt.Errorf("telegram API ID is not configured")
		}

		if c.Telegram.APIHash == "" || c.Telegram.APIHash == "YOUR_API_HASH" {
			return fmt.Errorf("telegram API hash is not configured")
		}
	}

	if c.UI.Layout.ChatListWidth+c.UI.Layout.ConversationWidth+c.UI.Layout.InfoWidth != 100 {
		return fmt.Errorf("layout widths must sum to 100%%")
	}

	return nil
}

// expandPaths expands ~ in paths to the home directory.
func (c *Config) expandPaths() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	expandPath := func(path string) string {
		if len(path) > 0 && path[0] == '~' {
			return filepath.Join(homeDir, path[1:])
		}
		return path
	}

	c.Telegram.SessionFile = expandPath(c.Telegram.SessionFile)
	c.Telegram.DatabaseDirectory = expandPath(c.Telegram.DatabaseDirectory)
	c.Cache.MediaDirectory = expandPath(c.Cache.MediaDirectory)
	c.Logging.File = expandPath(c.Logging.File)
}

// findConfigFile searches for a config file in standard locations.
func findConfigFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	locations := []string{
		"config.yaml",
		filepath.Join(".", "config.yaml"),
		filepath.Join(homeDir, ".config", "ithil", "config.yaml"),
		filepath.Join(homeDir, ".ithil.yaml"),
	}

	for _, loc := range locations {
		if fileExists(loc) {
			return loc
		}
	}

	return ""
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// EnsureDirectories creates necessary directories for the application.
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		filepath.Dir(c.Telegram.SessionFile),
		c.Telegram.DatabaseDirectory,
		c.Cache.MediaDirectory,
		filepath.Dir(c.Logging.File),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
