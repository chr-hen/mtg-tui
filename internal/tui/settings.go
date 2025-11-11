package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Settings represents user preferences
type Settings struct {
	ShowArenaCards bool `json:"show_arena_cards"` // Show cards with names starting with "A-"
	ShowTypeCard   bool `json:"show_type_card"`   // Show cards with type "card"
}

// GetSettingsFilePath returns the path to the settings file
func GetSettingsFilePath() string {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".mtg-tui")
	return filepath.Join(cacheDir, "config.json")
}

// LoadSettings loads settings from the config file, creating default settings if the file doesn't exist
func LoadSettings() (*Settings, error) {
	settingsPath := GetSettingsFilePath()
	
	// Check if settings file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// Return default settings
		return &Settings{
			ShowArenaCards: false, // Default to hiding Arena cards
			ShowTypeCard:   false, // Default to hiding cards with type "card"
		}, nil
	}
	
	// Read settings file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}
	
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}
	
	return &settings, nil
}

// SaveSettings saves settings to the config file
func SaveSettings(settings *Settings) error {
	settingsPath := GetSettingsFilePath()
	
	// Ensure the directory exists
	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}
	
	// Marshal settings to JSON with indentation
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	
	return nil
}

// IsArenaCard checks if a card name indicates it's an Arena card (starts with "A-")
func IsArenaCard(cardName string) bool {
	return len(cardName) >= 2 && cardName[0] == 'A' && cardName[1] == '-'
}

// IsTypeCard checks if a card's type line is exactly "card"
func IsTypeCard(typeLine string) bool {
	return strings.ToLower(strings.TrimSpace(typeLine)) == "card"
}

