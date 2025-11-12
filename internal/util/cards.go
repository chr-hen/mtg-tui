package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chr-hen/mtg-tui/internal/api"
)

// CardGroup represents a card with all its printings across different sets
type CardGroup struct {
	Card      api.Card   // The "canonical" card (we'll use the first one or one with most common data)
	Printings []api.Card // All printings of this card
}

// GroupCardsByName groups cards by their name, collecting all printings
// Note: This function preserves the order of cards as they appear in the input slice
func GroupCardsByName(cards []api.Card, showArenaCards bool, showTypeCard bool, showArtCards bool) []CardGroup {
	groupsMap := make(map[string]*CardGroup)

	for _, card := range cards {
		// Filter out Arena cards if setting is disabled
		if !showArenaCards && IsArenaCard(card.Name) {
			continue
		}

		// Filter out cards with type "card" if setting is disabled
		if !showTypeCard && IsTypeCard(card.TypeLine) {
			continue
		}

		// Filter out art cards if setting is disabled
		if !showArtCards && IsArtCard(card.TypeLine) {
			continue
		}

		cardName := strings.ToLower(card.Name)
		if group, exists := groupsMap[cardName]; exists {
			// Add this printing to the existing group
			group.Printings = append(group.Printings, card)
		} else {
			// Create a new group
			groupsMap[cardName] = &CardGroup{
				Card:      card, // Use first occurrence as canonical
				Printings: []api.Card{card},
			}
		}
	}

	// Convert map to slice, preserving order from input
	// Create a mapping to track first occurrence order
	groupOrder := make([]string, 0, len(groupsMap))
	seenNames := make(map[string]bool)
	
	for _, card := range cards {
		// Skip filtered cards
		if !showArenaCards && IsArenaCard(card.Name) {
			continue
		}
		if !showTypeCard && IsTypeCard(card.TypeLine) {
			continue
		}
		if !showArtCards && IsArtCard(card.TypeLine) {
			continue
		}
		
		cardName := strings.ToLower(card.Name)
		if !seenNames[cardName] {
			groupOrder = append(groupOrder, cardName)
			seenNames[cardName] = true
		}
	}
	
	// Build groups in the order they appeared in the sorted input
	groups := make([]CardGroup, 0, len(groupsMap))
	for _, cardName := range groupOrder {
		if group, exists := groupsMap[cardName]; exists {
			// Sort printings by set code, then collector number
			sort.Slice(group.Printings, func(i, j int) bool {
				if group.Printings[i].SetCode != group.Printings[j].SetCode {
					return group.Printings[i].SetCode < group.Printings[j].SetCode
				}
				return CompareCollectorNumbers(group.Printings[i].CollectorNumber, group.Printings[j].CollectorNumber)
			})
			groups = append(groups, *group)
		}
	}

	return groups
}

// CompareCollectorNumbers compares two collector numbers
// Numeric collector numbers are sorted numerically; alphanumeric are sorted lexicographically
func CompareCollectorNumbers(a, b string) bool {
	// Try to parse as integers first
	numA, errA := strconv.Atoi(a)
	numB, errB := strconv.Atoi(b)

	// If both are numeric, compare numerically
	if errA == nil && errB == nil {
		return numA < numB
	}

	// If one is numeric and one isn't, numeric comes first
	if errA == nil && errB != nil {
		return true
	}
	if errA != nil && errB == nil {
		return false
	}

	// Both are alphanumeric, compare as strings
	return strings.ToLower(a) < strings.ToLower(b)
}

// ParsePowerToughness parses power/toughness values, handling * and other special values
func ParsePowerToughness(value string) (float64, error) {
	if value == "" || value == "*" {
		return 0, fmt.Errorf("cannot parse special value")
	}
	return strconv.ParseFloat(value, 64)
}

// IsArenaCard checks if a card name indicates it's an Arena card (starts with "A-")
func IsArenaCard(cardName string) bool {
	return len(cardName) >= 2 && cardName[0] == 'A' && cardName[1] == '-'
}

// IsTypeCard checks if a card's type line is exactly "card"
func IsTypeCard(typeLine string) bool {
	return strings.ToLower(strings.TrimSpace(typeLine)) == "card"
}

// IsArtCard checks if a card's type line is "Card // Card" (art cards)
func IsArtCard(typeLine string) bool {
	return strings.TrimSpace(typeLine) == "Card // Card"
}

// SortCards sorts cards according to the specified sort field and direction
func SortCards(cards *[]api.Card, sortField string, sortAscending bool) {
	sort.Slice(*cards, func(i, j int) bool {
		less := CompareCards((*cards)[i], (*cards)[j], sortField)
		if !sortAscending {
			return !less
		}
		return less
	})
}

// CompareCards compares two cards based on the specified sort field
func CompareCards(card1, card2 api.Card, sortField string) bool {
	switch sortField {
	case "name":
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "released":
		// Parse release dates and compare
		date1, err1 := time.Parse("2006-01-02", card1.ReleasedAt)
		date2, err2 := time.Parse("2006-01-02", card2.ReleasedAt)
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		return date1.Before(date2)
	case "set":
		// Sort by set code, then collector number
		if card1.SetCode != card2.SetCode {
			return card1.SetCode < card2.SetCode
		}
		return CompareCollectorNumbers(card1.CollectorNumber, card2.CollectorNumber)
	case "rarity":
		// Rarity order: Common < Uncommon < Rare < Mythic
		rarityOrder := map[string]int{
			"common":   1,
			"uncommon": 2,
			"rare":     3,
			"mythic":   4,
		}
		rarity1 := rarityOrder[strings.ToLower(card1.Rarity)]
		rarity2 := rarityOrder[strings.ToLower(card2.Rarity)]
		if rarity1 == 0 {
			rarity1 = 99
		}
		if rarity2 == 0 {
			rarity2 = 99
		}
		if rarity1 != rarity2 {
			return rarity1 < rarity2
		}
		// If same rarity, sort by name
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "color":
		// Sort by number of colors, then by color identity
		if len(card1.Colors) != len(card2.Colors) {
			return len(card1.Colors) < len(card2.Colors)
		}
		// If same number of colors, compare color identity strings
		colors1 := strings.Join(card1.Colors, "")
		colors2 := strings.Join(card2.Colors, "")
		if colors1 != colors2 {
			return colors1 < colors2
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "cmc":
		if card1.CMC != card2.CMC {
			return card1.CMC < card2.CMC
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "power":
		power1, err1 := ParsePowerToughness(card1.Power)
		power2, err2 := ParsePowerToughness(card2.Power)
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		if power1 != power2 {
			return power1 < power2
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "toughness":
		tough1, err1 := ParsePowerToughness(card1.Toughness)
		tough2, err2 := ParsePowerToughness(card2.Toughness)
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		if tough1 != tough2 {
			return tough1 < tough2
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	default:
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	}
}

