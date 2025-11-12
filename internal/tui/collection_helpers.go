package tui

import (
	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/collections"
)

// getCollectionCards filters all cards to only those owned in the collection
func (a *App) getCollectionCards() ([]api.Card, error) {
	ownedPrintings := a.collection.OwnedPrintings
	if len(ownedPrintings) == 0 {
		return []api.Card{}, nil
	}

	// Load all cards from cache
	allCards, err := api.LoadCardsFromCache()
	if err != nil {
		return nil, err
	}

	// Create a map of owned printings for quick lookup
	ownedMap := make(map[string]bool)
	for _, owned := range ownedPrintings {
		cardID := collections.GetCardID(owned.SetCode, owned.CollectorNumber)
		ownedMap[cardID] = true
	}

	// Filter cards to only owned printings
	var ownedCards []api.Card
	for _, card := range allCards {
		cardID := collections.GetCardID(card.SetCode, card.CollectorNumber)
		if ownedMap[cardID] {
			ownedCards = append(ownedCards, card)
		}
	}

	return ownedCards, nil
}

// applyCardFilters applies a filter query to a set of cards
// This is a generic function that can be used for collection, search results, etc.
func applyCardFilters(cards []api.Card, filterQuery string) ([]api.Card, error) {
	// If no filter, return all cards
	if filterQuery == "" {
		return cards, nil
	}

	// Parse the filter query
	conditions, err := api.ParseQuery(filterQuery)
	if err != nil {
		return nil, err
	}

	// If no conditions, return all cards
	if len(conditions) == 0 {
		return cards, nil
	}

	// Filter cards based on conditions
	var filteredCards []api.Card
	for _, card := range cards {
		if api.MatchesCard(card, conditions) {
			filteredCards = append(filteredCards, card)
		}
	}

	return filteredCards, nil
}

// applyCollectionFilters applies the current collection filter query to cards
func (a *App) applyCollectionFilters(cards []api.Card) ([]api.Card, error) {
	return applyCardFilters(cards, a.collectionFilterQuery)
}

// applySearchFilters applies the current search filter query to cards
func (a *App) applySearchFilters(cards []api.Card) ([]api.Card, error) {
	return applyCardFilters(cards, a.searchFilterQuery)
}

// preloadCollectionCards pre-loads collection cards in the background
func (a *App) preloadCollectionCards() {
	// Get collection cards
	ownedCards, err := a.getCollectionCards()
	if err != nil {
		// Silently fail - collection will load on demand
		return
	}

	// Cache the collection cards for faster access
	if a.allMatchingCards == nil {
		a.allMatchingCards = make(map[string][]api.Card)
	}
	a.allMatchingCards["collection"] = ownedCards
}

