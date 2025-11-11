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

