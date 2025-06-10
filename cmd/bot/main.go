package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/dfaile/backstage-nobl9/internal/bot"
	"github.com/dfaile/backstage-nobl9/internal/nobl9"
)

func main() {
	// Command line flags
	clientID := flag.String("client-id", "", "Nobl9 client ID")
	clientSecret := flag.String("client-secret", "", "Nobl9 client secret")
	org := flag.String("organization", "", "Nobl9 organization")
	baseURL := flag.String("url", "", "Nobl9 base URL")
	flag.Parse()

	// Set environment variables if command line flags are provided
	// This allows the Nobl9 SDK to pick them up automatically
	if *clientID != "" {
		os.Setenv("NOBL9_SDK_CLIENT_ID", *clientID)
	}
	if *clientSecret != "" {
		os.Setenv("NOBL9_SDK_CLIENT_SECRET", *clientSecret)
	}
	if *org != "" {
		os.Setenv("NOBL9_SDK_ORGANIZATION", *org)
	}
	if *baseURL != "" {
		os.Setenv("NOBL9_SDK_URL", *baseURL)
	}

	// Create Nobl9 client using the SDK's built-in configuration system
	// This will automatically read from:
	// 1. Environment variables (highest priority)
	// 2. ~/.nobl9/config.toml (if exists)
	// 3. Default values
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