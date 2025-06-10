package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/dfaile/backstage-nobl9/internal/bot"
	"github.com/dfaile/backstage-nobl9/internal/config"
	"github.com/dfaile/backstage-nobl9/internal/nobl9"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	clientID := flag.String("client-id", "", "Nobl9 client ID")
	clientSecret := flag.String("client-secret", "", "Nobl9 client secret")
	org := flag.String("organization", "", "Nobl9 organization")
	baseURL := flag.String("url", "", "Nobl9 base URL")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override config with command line flags if provided
	if *clientID != "" {
		cfg.ClientID = *clientID
	}
	if *clientSecret != "" {
		cfg.ClientSecret = *clientSecret
	}
	if *org != "" {
		cfg.Organization = *org
	}
	if *baseURL != "" {
		cfg.URL = *baseURL
	}

	// Validate required configuration
	if cfg.ClientID == "" {
		log.Fatal("Client ID is required")
	}
	if cfg.ClientSecret == "" {
		log.Fatal("Client secret is required")
	}
	if cfg.Organization == "" {
		log.Fatal("Organization is required")
	}

	// Save updated configuration
	if err := cfg.Save(); err != nil {
		log.Printf("Warning: Failed to save config: %v", err)
	}

	// Set environment variables for the Nobl9 SDK
	// The SDK uses these standard environment variable names
	os.Setenv("NOBL9_SDK_CLIENT_ID", cfg.ClientID)
	os.Setenv("NOBL9_SDK_CLIENT_SECRET", cfg.ClientSecret)
	os.Setenv("NOBL9_SDK_ORGANIZATION", cfg.Organization)
	if cfg.URL != "" {
		os.Setenv("NOBL9_SDK_URL", cfg.URL)
	}

	// Create Nobl9 client
	// Pass empty values since the SDK will read from environment variables
	nobl9Client, err := nobl9.NewClient("", "", "", "")
	if err != nil {
		log.Fatalf("Failed to create Nobl9 client: %v", err)
	}

	// Create and start bot
	slackBot, err := bot.New(nobl9Client)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create context for the bot
	ctx := context.Background()

	// Start the bot (this is a CLI bot, so it runs interactively)
	if err := slackBot.Start(ctx); err != nil {
		log.Fatalf("Bot failed: %v", err)
	}
} 