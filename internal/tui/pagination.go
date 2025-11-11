package tui

import (
	"fmt"
	"log"

	"github.com/chr-hen/mtg-tui/internal/api"
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
		// Group cards by name
		cardGroups := a.groupCardsByName(allCards)

		// Use cached data with unified pagination
		pageSize := 10
		a.paginateCardGroups(cardGroups, a.currentPage, pageSize)

		// Pre-load next pages in background (using CardGroups)
		go a.preloadCardGroupPages(a.currentQuery, cardGroups, a.currentPage, pageSize)
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

// preloadPages pre-loads the next few pages in the background for faster navigation
// This is a legacy function that converts cards to groups and uses the unified preloader
func (a *App) preloadPages(query string, currentPage int, allCards []api.Card, pageSize int) {
	// Group cards first, then use unified preloader
	cardGroups := a.groupCardsByName(allCards)
	a.preloadCardGroupPages(query, cardGroups, currentPage, pageSize)
}

