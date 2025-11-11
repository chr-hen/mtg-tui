package tui

import (
	"github.com/chr-hen/mtg-tui/internal/api"
)

// paginateCardGroups handles the common pagination logic for CardGroups
func (a *App) paginateCardGroups(cardGroups []CardGroup, page int, pageSize int) {
	totalGroups := len(cardGroups)
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalGroups {
		a.cardGroups = []CardGroup{}
		a.pagination = api.PaginationInfo{
			HasMore:    false,
			TotalCards: totalGroups,
		}
	} else {
		if endIdx > totalGroups {
			endIdx = totalGroups
		}
		a.cardGroups = cardGroups[startIdx:endIdx]
		hasMore := endIdx < totalGroups
		a.pagination = api.PaginationInfo{
			HasMore:    hasMore,
			TotalCards: totalGroups,
		}
	}
}

// preloadCardGroupPages pre-loads the next few pages of CardGroups in the background
func (a *App) preloadCardGroupPages(query string, allGroups []CardGroup, currentPage int, pageSize int) {
	// Pre-load next 2 pages
	for i := 1; i <= 2; i++ {
		page := currentPage + i
		startIdx := (page - 1) * pageSize
		endIdx := startIdx + pageSize

		if startIdx >= len(allGroups) {
			break
		}

		if endIdx > len(allGroups) {
			endIdx = len(allGroups)
		}

		// Check if already cached
		if a.pageCache == nil {
			a.pageCache = make(map[string]map[int][]api.Card)
		}
		if a.pageCache[query] == nil {
			a.pageCache[query] = make(map[int][]api.Card)
		}
		if _, exists := a.pageCache[query][page]; !exists {
			// Cache this page (convert groups to cards for consistency with existing cache structure)
			pageGroups := allGroups[startIdx:endIdx]
			pageCards := make([]api.Card, 0)
			for _, group := range pageGroups {
				pageCards = append(pageCards, group.Printings...)
			}
			a.pageCache[query][page] = pageCards
		}
	}
}

