package tui

import (
	"sort"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/api"
)

// loadAutocompleteData extracts unique values from card data for autocomplete
func (a *App) loadAutocompleteData() {
	cards, err := api.LoadCardsFromCache()
	if err != nil {
		// Cache not available yet, will be loaded later
		return
	}

	typeMap := make(map[string]bool)
	setMap := make(map[string]bool)
	rarityMap := make(map[string]bool)
	keywordMap := make(map[string]bool)

	for _, card := range cards {
		// Extract types (split TypeLine by spaces and dashes)
		typeParts := strings.FieldsFunc(card.TypeLine, func(r rune) bool {
			return r == ' ' || r == '-' || r == '—'
		})
		for _, part := range typeParts {
			part = strings.TrimSpace(part)
			if part != "" && len(part) > 1 {
				typeMap[strings.ToLower(part)] = true
			}
		}

		// Extract sets (use both SetCode and Set name)
		if card.SetCode != "" {
			setMap[strings.ToLower(card.SetCode)] = true
		}
		if card.Set != "" {
			setMap[strings.ToLower(card.Set)] = true
		}

		// Extract rarities
		if card.Rarity != "" {
			rarityMap[strings.ToLower(card.Rarity)] = true
		}

		// Extract keywords
		for _, kw := range card.Keywords {
			if kw != "" {
				keywordMap[strings.ToLower(kw)] = true
			}
		}
	}

	// Convert maps to sorted slices
	a.uniqueTypes = make([]string, 0, len(typeMap))
	for t := range typeMap {
		a.uniqueTypes = append(a.uniqueTypes, t)
	}
	sort.Strings(a.uniqueTypes)

	a.uniqueSets = make([]string, 0, len(setMap))
	for s := range setMap {
		a.uniqueSets = append(a.uniqueSets, s)
	}
	sort.Strings(a.uniqueSets)

	a.uniqueRarities = make([]string, 0, len(rarityMap))
	for r := range rarityMap {
		a.uniqueRarities = append(a.uniqueRarities, r)
	}
	sort.Strings(a.uniqueRarities)

	a.uniqueKeywords = make([]string, 0, len(keywordMap))
	for k := range keywordMap {
		a.uniqueKeywords = append(a.uniqueKeywords, k)
	}
	sort.Strings(a.uniqueKeywords)
}

// createAutocompleteFunc creates an autocomplete function for a given list of options
func (a *App) createAutocompleteFunc(options []string) func(currentText string) []string {
	return func(currentText string) []string {
		currentText = strings.TrimSpace(currentText)
		if currentText == "" {
			return []string{}
		}

		// Remove common query prefixes for matching
		searchText := strings.ToLower(currentText)
		prefixes := []string{"t:", "type:", "s:", "set:", "r:", "rarity:", "kw:", "keyword:"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(searchText, prefix) {
				searchText = strings.TrimPrefix(searchText, prefix)
				break
			}
		}

		if searchText == "" {
			return []string{}
		}

		// Find all matching options (limit to 10 for performance)
		var matches []string
		for _, option := range options {
			if strings.HasPrefix(strings.ToLower(option), searchText) {
				matches = append(matches, option)
				if len(matches) >= 10 {
					break
				}
			}
		}

		return matches
	}
}
