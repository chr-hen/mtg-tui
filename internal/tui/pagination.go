package tui

import (
	"fmt"
	"log"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/rivo/tview"
)

func (a *App) loadNextPage() {
	a.currentPage++
	a.loadPageFromCache()
}

func (a *App) loadPreviousPage() {
	a.currentPage--
	a.loadPageFromCache()
}

func (a *App) loadPageFromCache() {
	// Get all matching cards for this query
	allCards, exists := a.allMatchingCards[a.currentQuery]
	if !exists {
		// Fallback to API call if cache doesn't exist
		result, err := api.FetchCards(a.currentQuery, a.currentPage)
		if err != nil {
			log.Printf("Error fetching page: %v", err)
			modal := tview.NewModal().
				SetText(fmt.Sprintf("Error fetching page: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
				})
			a.pages.AddPage("error", modal, true, true)
			return
		}
		a.cards = result.Cards
		a.pagination = result.Pagination
	} else {
		// Apply appropriate filters based on current view
		var filteredCards []api.Card
		var err error
		if a.currentQuery == "collection" {
			// Apply collection filters
			filteredCards, err = a.applyCollectionFilters(allCards)
		} else {
			// Apply search filters
			filteredCards, err = a.applySearchFilters(allCards)
		}
		if err != nil {
			// Silently fail - show unfiltered cards
			filteredCards = allCards
		}

		// Group filtered cards by name
		cardGroups := util.GroupCardsByName(filteredCards, a.settings.ShowArenaCards, a.settings.ShowTypeCard)

		// Use cached data with unified pagination
		pageSize := 10
		a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)
	}

	a.populateList()
	// Use appropriate title update based on query
	if a.currentQuery == "collection" {
		a.updateCollectionTitle()
	} else {
		a.updateListTitle()
	}
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}

