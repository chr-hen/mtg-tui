package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FieldVisibility represents which fields are visible in search/filter forms
type FieldVisibility struct {
	Name      bool `json:"name"`      // Name field
	Type      bool `json:"type"`      // Type field
	Color     bool `json:"color"`     // Color field
	Oracle    bool `json:"oracle"`    // Oracle Text field
	Mana      bool `json:"mana"`      // Mana Cost field
	Power     bool `json:"power"`     // Power field
	Toughness bool `json:"toughness"` // Toughness field
	Set       bool `json:"set"`       // Set field
	Rarity    bool `json:"rarity"`   // Rarity field
	Year      bool `json:"year"`      // Year field
	Artist    bool `json:"artist"`   // Artist field
	Keyword   bool `json:"keyword"`  // Keyword field
	Is        bool `json:"is"`       // Is field
}

// GetDefaultFieldVisibility returns field visibility with all fields enabled by default
func GetDefaultFieldVisibility() FieldVisibility {
	return FieldVisibility{
		Name:      true,
		Type:      true,
		Color:     true,
		Oracle:    true,
		Mana:      true,
		Power:     true,
		Toughness: true,
		Set:       true,
		Rarity:    true,
		Year:      true,
		Artist:    true,
		Keyword:   true,
		Is:        true,
	}
}

// Settings represents user preferences
type Settings struct {
	ShowArenaCards   bool            `json:"show_arena_cards"`   // Show cards with names starting with "A-"
	ShowTypeCard     bool            `json:"show_type_card"`     // Show cards with type "card"
	ShowArtCards     bool            `json:"show_art_cards"`     // Show art cards with type "Card // Card"
	PageSize         int             `json:"page_size"`          // Number of cards to show per page
	FieldVisibility  FieldVisibility `json:"field_visibility"`   // Which fields are visible in search/filter forms
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
			ShowArenaCards:  false,                    // Default to hiding Arena cards
			ShowTypeCard:    false,                  // Default to hiding cards with type "card"
			ShowArtCards:    false,                  // Default to hiding art cards
			PageSize:        10,                     // Default to 10 cards per page
			FieldVisibility: GetDefaultFieldVisibility(), // All fields enabled by default
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
	
	// Ensure field visibility is initialized (for old config files that don't have it)
	if settings.FieldVisibility == (FieldVisibility{}) {
		settings.FieldVisibility = GetDefaultFieldVisibility()
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

