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
		// Use cached data
		pageSize := 10
		totalCards := len(allCards)
		startIdx := (a.currentPage - 1) * pageSize
		endIdx := startIdx + pageSize

		if startIdx >= totalCards {
			a.cards = []api.Card{}
			a.pagination = api.PaginationInfo{
				HasMore:    false,
				TotalCards: totalCards,
			}
		} else {
			if endIdx > totalCards {
				endIdx = totalCards
			}

			// Check cache first
			if cachedPage, exists := a.pageCache[a.currentQuery][a.currentPage]; exists {
				a.cards = cachedPage
			} else {
				// Calculate and cache this page
				a.cards = allCards[startIdx:endIdx]
				a.pageCache[a.currentQuery][a.currentPage] = a.cards
			}

			hasMore := endIdx < totalCards
			a.pagination = api.PaginationInfo{
				HasMore:    hasMore,
				TotalCards: totalCards,
			}
		}

		// Pre-load next pages in background
		go a.preloadPages(a.currentQuery, a.currentPage, allCards, pageSize)
	}

	a.populateList()
	a.updateListTitle()
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}

// preloadPages pre-loads the next few pages in the background for faster navigation
func (a *App) preloadPages(query string, currentPage int, allCards []api.Card, pageSize int) {
	// Pre-load next 2 pages
	for i := 1; i <= 2; i++ {
		page := currentPage + i
		startIdx := (page - 1) * pageSize
		endIdx := startIdx + pageSize

		if startIdx >= len(allCards) {
			break
		}

		if endIdx > len(allCards) {
			endIdx = len(allCards)
		}

		// Check if already cached
		if _, exists := a.pageCache[query][page]; !exists {
			// Cache this page
			pageCards := allCards[startIdx:endIdx]
			a.pageCache[query][page] = pageCards
		}
	}
}

