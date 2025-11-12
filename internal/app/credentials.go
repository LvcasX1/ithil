package app

import (
	"strconv"
)

const (
	// DefaultAPIID is the built-in API ID for Ithil
	// This allows users to run the application without obtaining their own credentials
	// For enhanced privacy, users can configure custom credentials in Settings (Ctrl+,)
	DefaultAPIID = "23883389"

	// DefaultAPIHash is the built-in API Hash for Ithil
	// This allows users to run the application without obtaining their own credentials
	// For enhanced privacy, users can configure custom credentials in Settings (Ctrl+,)
	DefaultAPIHash = "d817d412503c7e65b7e3250fcac047cc"
)

// GetAPICredentials returns the appropriate API credentials based on configuration
// If custom credentials are configured and valid, they are used
// Otherwise, default built-in credentials are returned
func GetAPICredentials(cfg *Config) (apiID int, apiHash string) {
	// Check if user wants to use default credentials explicitly
	if cfg.Telegram.UseDefaultCredentials {
		return parseAPIID(DefaultAPIID), DefaultAPIHash
	}

	// Check if custom credentials are provided and valid
	if cfg.Telegram.APIID != "" && cfg.Telegram.APIHash != "" &&
		cfg.Telegram.APIID != "YOUR_API_ID" && cfg.Telegram.APIHash != "YOUR_API_HASH" {
		return parseAPIID(cfg.Telegram.APIID), cfg.Telegram.APIHash
	}

	// Fall back to default credentials
	return parseAPIID(DefaultAPIID), DefaultAPIHash
}

// IsUsingDefaultCredentials checks if the application is using default credentials
func IsUsingDefaultCredentials(cfg *Config) bool {
	// Explicitly set to use defaults
	if cfg.Telegram.UseDefaultCredentials {
		return true
	}

	// Empty or placeholder credentials = use defaults
	if cfg.Telegram.APIID == "" || cfg.Telegram.APIHash == "" {
		return true
	}

	if cfg.Telegram.APIID == "YOUR_API_ID" || cfg.Telegram.APIHash == "YOUR_API_HASH" {
		return true
	}

	// Check if configured credentials match defaults
	apiID, apiHash := cfg.Telegram.APIID, cfg.Telegram.APIHash
	if apiID == DefaultAPIID && apiHash == DefaultAPIHash {
		return true
	}

	return false
}

// parseAPIID converts API ID string to int
// Returns 0 if parsing fails (will cause auth error, which is intentional)
func parseAPIID(apiID string) int {
	id, err := strconv.Atoi(apiID)
	if err != nil {
		return 0
	}
	return id
}
