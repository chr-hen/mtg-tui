package collections

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GetCollectionsDir returns the path to the collections directory
func GetCollectionsDir() string {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".mtg-tui")
	return cacheDir
}

// GetCollectionFilePath returns the path to the collection file
func GetCollectionFilePath() string {
	return filepath.Join(GetCollectionsDir(), "collection.json")
}

// GetDecksFilePath returns the path to the decks file
func GetDecksFilePath() string {
	return filepath.Join(GetCollectionsDir(), "decks.json")
}

// GetCubesFilePath returns the path to the cubes file
func GetCubesFilePath() string {
	return filepath.Join(GetCollectionsDir(), "cubes.json")
}

// GetListsFilePath returns the path to the lists file
func GetListsFilePath() string {
	return filepath.Join(GetCollectionsDir(), "lists.json")
}

// GetWantsFilePath returns the path to the wants file
func GetWantsFilePath() string {
	return filepath.Join(GetCollectionsDir(), "wants.json")
}

// EnsureCollectionsDir creates the collections directory if it doesn't exist
func EnsureCollectionsDir() error {
	dir := GetCollectionsDir()
	return os.MkdirAll(dir, 0755)
}

// LoadJSONFile loads a JSON file into the provided struct
func LoadJSONFile(filePath string, target interface{}) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return empty/default struct
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// SaveJSONFile saves a struct to a JSON file
func SaveJSONFile(filePath string, data interface{}) error {
	// Ensure directory exists
	if err := EnsureCollectionsDir(); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

