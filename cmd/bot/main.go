package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Daniel/backstage-nobl9/internal/bot"
	"github.com/Daniel/backstage-nobl9/internal/config"
	"github.com/Daniel/backstage-nobl9/internal/nobl9"
)

func main() {
	// Parse command line flags
	clientID := flag.String("client-id", "", "Nobl9 client ID")
	clientSecret := flag.String("client-secret", "", "Nobl9 client secret")
	org := flag.String("org", "", "Nobl9 organization")
	baseURL := flag.String("base-url", "", "Nobl9 base URL (default: https://app.nobl9.com)")
	configPath := flag.String("config", "", "Path to config file (default: ~/.nobl9/config.json)")
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
		cfg.BaseURL = *baseURL
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

	// Create Nobl9 client
	nobl9Client, err := nobl9.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.Organization, cfg.BaseURL)
	if err != nil {
		log.Fatalf("Failed to create Nobl9 client: %v", err)
	}

	// Create bot instance
	b, err := bot.New(nobl9Client)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start the bot
	if err := b.Start(ctx); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
} 