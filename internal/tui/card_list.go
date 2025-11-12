package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/rivo/tview"
)

func (a *App) showCardList() {
	// Check if we have cached matching cards for this query
	allCards, hasCache := a.allMatchingCards[a.currentQuery]
	if !hasCache {
		// Load all matching cards and cache them
		var err error
		allCards, err = api.GetAllMatchingCards(a.currentQuery)
		if err != nil {
			// Show error modal
			modal := tview.NewModal().
				SetText(fmt.Sprintf("Error fetching cards: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					// Force a full redraw by recreating the menu
					a.showMainMenu()
				})
			a.pages.AddPage("error", modal, true, true)
			return
		}
		// Sort the cards according to current sort settings
		util.SortCards(&allCards, a.sortField, a.sortAscending)
		// Cache all matching cards
		a.allMatchingCards[a.currentQuery] = allCards
		// Initialize page cache for this query
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
	} else {
		// Re-sort cached cards if sort settings changed
		util.SortCards(&allCards, a.sortField, a.sortAscending)
		a.allMatchingCards[a.currentQuery] = allCards
		// Clear page cache when sort changes
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
	}

	// Apply search filters to the cards
	filteredCards, err := a.applySearchFilters(allCards)
	if err != nil {
		// Show error modal
		errorModal := tview.NewModal().
			SetText(fmt.Sprintf("Error applying filters: %v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.RemovePage("search_filter_error")
				if a.pages.HasPage("list") {
					a.pages.SwitchToPage("list")
					if a.table != nil {
						a.app.SetFocus(a.table)
					}
				}
			})
		a.pages.AddPage("search_filter_error", errorModal, true, true)
		return
	}

	// Group filtered cards by name
	cardGroups := util.GroupCardsByName(filteredCards, a.settings.ShowArenaCards, a.settings.ShowTypeCard, a.settings.ShowArtCards)

	// Use unified pagination
	pageSize := a.settings.PageSize
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	// Show unified card list with search-specific config
	a.showUnifiedCardList(models.CardListConfig{
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
	})
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

