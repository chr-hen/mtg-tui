package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var (
	cacheDir      = filepath.Join(os.Getenv("HOME"), ".mtg-tui")
	cacheFilePath = filepath.Join(cacheDir, "default-cards.json")
)

// EnsureCacheDir creates the cache directory if it doesn't exist
func EnsureCacheDir() error {
	return os.MkdirAll(cacheDir, 0755)
}

// GetBulkDataMetadata fetches the list of bulk data files from Scryfall
func GetBulkDataMetadata() ([]BulkData, error) {
	resp, err := http.Get("https://api.scryfall.com/bulk-data")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bulk data metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result BulkDataResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result.Data, nil
}

// GetDefaultCardsBulkData finds the default_cards bulk data entry
func GetDefaultCardsBulkData(bulkDataList []BulkData) (*BulkData, error) {
	for _, data := range bulkDataList {
		if data.Type == "default_cards" {
			return &data, nil
		}
	}
	return nil, fmt.Errorf("default_cards bulk data not found")
}

// DownloadBulkData downloads the bulk data file from the given URL
func DownloadBulkData(downloadURL string, progress func(bytesRead, totalBytes int64)) error {
	if err := EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download bulk data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status: %s", resp.Status)
	}

	// Create a temporary file first
	tmpFile := cacheFilePath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer out.Close()

	// Copy with progress tracking if provided
	if progress != nil {
		totalBytes := resp.ContentLength
		buf := make([]byte, 64*1024) // 64KB buffer for better performance
		var bytesRead int64
		lastUpdate := int64(0)
		updateInterval := int64(1024 * 1024) // Update every 1MB

		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				written, writeErr := out.Write(buf[:n])
				if writeErr != nil {
					return fmt.Errorf("failed to write to cache file: %w", writeErr)
				}
				bytesRead += int64(written)

				// Update progress more frequently (every MB or if totalBytes is unknown)
				if totalBytes == -1 || bytesRead-lastUpdate >= updateInterval || err == io.EOF {
					progress(bytesRead, totalBytes)
					lastUpdate = bytesRead
				}
			}
			if err == io.EOF {
				// Final progress update
				progress(bytesRead, totalBytes)
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read download: %w", err)
			}
		}
	} else {
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to copy download: %w", err)
		}
	}

	// Atomically replace the old file
	if err := os.Rename(tmpFile, cacheFilePath); err != nil {
		return fmt.Errorf("failed to replace cache file: %w", err)
	}

	return nil
}

// LoadCardsFromCache loads all cards from the cached JSON file
func LoadCardsFromCache() ([]Card, error) {
	file, err := os.Open(cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cards []Card

	// The bulk data file is a JSON array, so we decode it as an array
	if err := decoder.Decode(&cards); err != nil {
		return nil, fmt.Errorf("failed to decode cache file: %w", err)
	}

	return cards, nil
}

// CacheExists checks if the cache file exists
func CacheExists() bool {
	_, err := os.Stat(cacheFilePath)
	return err == nil
}

// GetCacheFilePath returns the path where the cache file is stored
func GetCacheFilePath() string {
	return cacheFilePath
}

// GetAllMatchingCards returns all cards matching a query (no pagination)
func GetAllMatchingCards(query string) ([]Card, error) {
	allCards, err := LoadCardsFromCache()
	if err != nil {
		return nil, err
	}

	// Parse query into conditions
	conditions, err := ParseQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Simple search implementation - matches card name or type line
	var matchingCards []Card

	for _, card := range allCards {
		// Check if card matches all conditions
		if MatchesCard(card, conditions) {
			matchingCards = append(matchingCards, card)
		}
	}

	return matchingCards, nil
}

// SearchCardsLocal searches through the local card cache using a simple query
func SearchCardsLocal(query string, page int, pageSize int) (*FetchCardsResult, error) {
	matchingCards, err := GetAllMatchingCards(query)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalCards := len(matchingCards)
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalCards {
		// Return empty result for pages beyond available data
		return &FetchCardsResult{
			Cards: []Card{},
			Pagination: PaginationInfo{
				HasMore:    false,
				TotalCards: totalCards,
			},
		}, nil
	}

	if endIdx > totalCards {
		endIdx = totalCards
	}

	paginatedCards := matchingCards[startIdx:endIdx]
	hasMore := endIdx < totalCards

	return &FetchCardsResult{
		Cards: paginatedCards,
		Pagination: PaginationInfo{
			HasMore:    hasMore,
			TotalCards: totalCards,
		},
	}, nil
}

// matchesQuery is kept for backward compatibility but now uses the new parser
func matchesQuery(card Card, query string) bool {
	conditions, err := ParseQuery(query)
	if err != nil {
		return false
	}
	return MatchesCard(card, conditions)
}
