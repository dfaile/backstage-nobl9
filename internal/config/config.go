package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the Nobl9 configuration following SDK conventions
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Organization string `json:"organization"`
	URL          string `json:"url"` // Changed from BaseURL to URL to match SDK
}

// DefaultConfigPath returns the default path for the config file
// Following Nobl9 SDK convention of ~/.nobl9/config.toml, but using JSON for simplicity
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".nobl9", "config.json"), nil
}

// LoadConfig loads the configuration from the specified path or default location
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			URL: "https://app.nobl9.com",
		}, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default URL if not specified
	if config.URL == "" {
		config.URL = "https://app.nobl9.com"
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified path
func SaveConfig(config *Config, path string) error {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal and write config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	return SaveConfig(c, "")
} 