package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/rivo/tview"
)

// showCollection displays the user's collection of owned cards
func (a *App) showCollection() {
	// Remove old collection page if it exists
	if a.pages.HasPage("collection") {
		a.pages.RemovePage("collection")
	}

	// Get all owned printings
	ownedPrintings := a.collection.OwnedPrintings
	if len(ownedPrintings) == 0 {
		// Show empty collection message
		modal := tview.NewModal().
			SetText("Your collection is empty.\n\nMark cards as owned from the card detail view.").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.RemovePage("collection_empty")
				if a.pages.HasPage("menu") {
					a.pages.SwitchToPage("menu")
					if a.menu != nil {
						a.app.SetFocus(a.menu)
					}
				} else {
					a.showMainMenu()
				}
			})
		a.pages.AddPage("collection_empty", modal, true, true)
		return
	}

	// Get collection cards using helper function
	ownedCards, err := a.getCollectionCards()
	if err != nil {
		// Show error modal
		errorModal := tview.NewModal().
			SetText(fmt.Sprintf("Error loading collection cards: %v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.RemovePage("collection_error")
				if a.pages.HasPage("menu") {
					a.pages.SwitchToPage("menu")
					if a.menu != nil {
						a.app.SetFocus(a.menu)
					}
				} else {
					a.showMainMenu()
				}
			})
		a.pages.AddPage("collection_error", errorModal, true, true)
		return
	}

	// Check if collection cards are already cached (pre-loaded)
	allCollectionCards, exists := a.allMatchingCards["collection"]
	if !exists {
		// Not pre-loaded, cache them now
		if a.allMatchingCards == nil {
			a.allMatchingCards = make(map[string][]api.Card)
		}
		a.allMatchingCards["collection"] = ownedCards
		allCollectionCards = ownedCards
	} else {
		// Use pre-loaded cards
		ownedCards = allCollectionCards
	}

	// Sort the cards according to current sort settings
	util.SortCards(&allCollectionCards, a.sortField, a.sortAscending)
	a.allMatchingCards["collection"] = allCollectionCards

	// Apply filters to collection cards
	filteredCards, err := a.applyCollectionFilters(allCollectionCards)
	if err != nil {
		// Show error modal
		errorModal := tview.NewModal().
			SetText(fmt.Sprintf("Error applying filters: %v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.RemovePage("collection_filter_error")
				if a.pages.HasPage("collection") {
					a.pages.SwitchToPage("collection")
					if a.table != nil {
						a.app.SetFocus(a.table)
					}
				}
			})
		a.pages.AddPage("collection_filter_error", errorModal, true, true)
		return
	}

	// Group filtered cards by name
	cardGroups := util.GroupCardsByName(filteredCards, a.settings.ShowArenaCards, a.settings.ShowTypeCard, a.settings.ShowArtCards)

	// Set up for display
	a.currentQuery = "collection"
	a.currentPage = 1
	
	// Use unified pagination
	pageSize := a.settings.PageSize
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	// Show unified card list with collection-specific config
	a.showUnifiedCardList(models.CardListConfig{
		PageName:   "collection",
		TitleFunc:  a.updateCollectionTitle,
		FooterText: "f: filter | s: sort | S: direction | j/k: navigate | h/l: pages | Enter: view details | Esc: back",
		FilterFunc: a.applyCollectionFilters,
		OnEscape: func() {
			if a.pages.HasPage("collection_menu") {
				a.pages.SwitchToPage("collection_menu")
			} else {
				a.showCollectionMenu()
			}
		},
		OnFilter: func() {
			a.showCollectionFilter()
		},
	})
}

// updateCollectionTitle updates the title of the collection view
func (a *App) updateCollectionTitle() {
	pageSize := a.settings.PageSize
	// Use TotalCards from pagination info (total unique cards across all pages)
	totalUniqueCards := a.pagination.TotalCards
	totalPages := 1
	if totalUniqueCards > 0 {
		// Calculate total pages: ceiling division
		totalPages = (totalUniqueCards + pageSize - 1) / pageSize
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

	title := fmt.Sprintf("My Collection - Page %d/%d", a.currentPage, totalPages)
	if a.collectionFilterQuery != "" {
		// Truncate long filter queries
		filterDisplay := a.collectionFilterQuery
		if len(filterDisplay) > 25 {
			filterDisplay = filterDisplay[:22] + "..."
		}
		title += fmt.Sprintf(" | Filter: %s", filterDisplay)
	}
	title += fmt.Sprintf(" | Sort: %s %s", sortFieldDisplay, sortDir)
	if totalUniqueCards > 0 {
		title += fmt.Sprintf(" | Total: %d unique cards", totalUniqueCards)
		// Count total printings across all owned printings in collection
		totalPrintings := len(a.collection.OwnedPrintings)
		title += fmt.Sprintf(" (%d printings)", totalPrintings)
	} else {
		title += " | Empty"
	}

	if a.table != nil {
		a.table.SetTitle(title)
	}
}

// loadCollectionPage loads the current page of the collection
// This now delegates to the unified reload logic
func (a *App) loadCollectionPage() {
	// Check if we have cached collection cards
	allCollectionCards, exists := a.allMatchingCards["collection"]
	if !exists {
		// Fallback: reload from collection using helper
		ownedCards, err := a.getCollectionCards()
		if err != nil {
			return
		}
		allCollectionCards = ownedCards
		a.allMatchingCards["collection"] = allCollectionCards
	}

	// Sort the cards according to current sort settings
	util.SortCards(&allCollectionCards, a.sortField, a.sortAscending)
	a.allMatchingCards["collection"] = allCollectionCards

	// Apply filters to collection cards
	filteredCards, err := a.applyCollectionFilters(allCollectionCards)
	if err != nil {
		// Silently fail - show unfiltered cards
		filteredCards = allCollectionCards
	}

	// Group filtered cards by name
	cardGroups := util.GroupCardsByName(filteredCards, a.settings.ShowArenaCards, a.settings.ShowTypeCard, a.settings.ShowArtCards)

	// Use unified pagination
	pageSize := a.settings.PageSize
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	a.populateList()
	a.updateCollectionTitle()
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}


