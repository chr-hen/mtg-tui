package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
)

func (a *App) showCardList() {
	// Capture currentQuery for the closure
	query := a.currentQuery
	
	// Use generic helper to handle all the loading/caching/sorting/filtering/display logic
	a.showCardsFromSource(
		query,
		func() ([]api.Card, error) {
			return api.GetAllMatchingCards(query)
		},
		models.CardListConfig{
			PageName:   "list",
			TitleFunc:  a.updateListTitle,
			FooterText: "f: filter | s: sort | S: direction | j/k: navigate | h/l: pages | Enter: details | Esc: back",
			FilterFunc: a.applySearchFilters,
			OnEscape: func() {
				if a.pages.HasPage("menu") {
					a.pages.SwitchToPage("menu")
					if a.menu != nil {
						a.app.SetFocus(a.menu)
					}
				} else {
					a.showMainMenu()
				}
			},
			OnFilter: func() {
				a.showSearchFilter()
			},
		},
	)
}

// populateList now delegates to populateListUnified
func (a *App) populateList() {
	a.populateListUnified()
}

func (a *App) updateListTitle() {
	pageSize := a.settings.PageSize
	totalPages := 1
	if a.pagination.TotalCards > 0 {
		// Calculate total pages: ceiling division
		totalPages = (a.pagination.TotalCards + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}
	}

	// Get sort field display name
	sortFieldNames := map[string]string{
		"name":      "Name",
		"released":  "Release Date",
		"set":       "Set/Number",
		"rarity":    "Rarity",
		"color":     "Color",
		"cmc":       "Mana Value",
		"power":     "Power",
		"toughness": "Toughness",
	}
	sortFieldDisplay := sortFieldNames[a.sortField]
	if sortFieldDisplay == "" {
		sortFieldDisplay = "Name"
	}
	sortDir := "↑"
	if !a.sortAscending {
		sortDir = "↓"
	}

	title := fmt.Sprintf("MTG Cards - Page %d/%d", a.currentPage, totalPages)
	if a.currentQuery != "" {
		// Truncate long queries
		queryDisplay := a.currentQuery
		if len(queryDisplay) > 25 {
			queryDisplay = queryDisplay[:22] + "..."
		}
		title += fmt.Sprintf(" | Query: %s", queryDisplay)
	}
	if a.searchFilterQuery != "" {
		// Truncate long filter queries
		filterDisplay := a.searchFilterQuery
		if len(filterDisplay) > 25 {
			filterDisplay = filterDisplay[:22] + "..."
		}
		title += fmt.Sprintf(" | Filter: %s", filterDisplay)
	}
	title += fmt.Sprintf(" | Sort: %s %s", sortFieldDisplay, sortDir)
	if a.pagination.TotalCards > 0 {
		title += fmt.Sprintf(" | Total: %d", a.pagination.TotalCards)
	}
	if a.table != nil {
		a.table.SetTitle(title)
	}
}

