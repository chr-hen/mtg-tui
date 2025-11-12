package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/rivo/tview"
)

// showCollection displays the user's collection of owned cards
func (a *App) showCollection() {
	// Check if collection is empty
	if len(a.collection.OwnedPrintings) == 0 {
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

	// Use generic helper to handle all the loading/caching/sorting/filtering/display logic
	a.showCardsFromSource(
		"collection",
		a.getCollectionCards,
		models.CardListConfig{
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
		},
	)
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

