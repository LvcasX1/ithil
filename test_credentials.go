package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gotd/td/telegram"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram struct {
		APIID   string `yaml:"api_id"`
		APIHash string `yaml:"api_hash"`
	} `yaml:"telegram"`
}

func main() {
	// Read config
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		return
	}

	fmt.Printf("Testing credentials:\n")
	fmt.Printf("  API ID: %s\n", config.Telegram.APIID)
	fmt.Printf("  API Hash: %s\n", config.Telegram.APIHash[:8]+"...") // Only show first 8 chars

	// Try to create client
	client := telegram.NewClient(
		// Parse API ID to int
		func() int {
			var id int
			fmt.Sscanf(config.Telegram.APIID, "%d", &id)
			return id
		}(),
		config.Telegram.APIHash,
		telegram.Options{
			Device: telegram.DeviceConfig{
				DeviceModel:    "Desktop",
				SystemVersion:  "Linux",
				AppVersion:     "0.1.0",
				SystemLangCode: "en",
				LangPack:       "tdesktop",
				LangCode:       "en",
			},
		},
	)

	ctx := context.Background()

	fmt.Println("\nAttempting connection test...")

	err = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("✓ Connection successful!")
		fmt.Println("✓ API credentials are valid!")
		return nil
	})

	if err != nil {
		fmt.Printf("✗ Connection failed: %v\n", err)
		fmt.Println("\nThis means your API credentials are invalid or incorrect.")
		fmt.Println("Please:")
		fmt.Println("1. Go to https://my.telegram.org")
		fmt.Println("2. Delete your app")
		fmt.Println("3. Create a NEW app")
		fmt.Println("4. Copy the new credentials to config.yaml")
	}
}
