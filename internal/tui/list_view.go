package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showListView displays the cards in a specific list
func (a *App) showListView(list collections.List) {
	pageName := fmt.Sprintf("list_%s", list.ID)

	// Check if list is empty and show message if so
	if len(list.Cards) == 0 {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("List \"%s\" is empty.\n\nAdd cards to this list from the card detail view.", list.Name)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.RemovePage("list_empty")
				a.showListsList()
			})
		a.pages.AddPage("list_empty", modal, true, true)
		return
	}

	// Use generic helper to handle all the loading/caching/sorting/filtering/display logic
	a.showCardsFromSource(
		pageName,
		func() ([]api.Card, error) {
			return a.getListCards(list)
		},
		models.CardListConfig{
			PageName: pageName,
			TitleFunc: func() {
				a.updateListViewTitle(list)
			},
			FooterText: "f: filter | s: sort | S: direction | j/k: navigate | h/l: pages | Enter: view details | Esc: back",
			FilterFunc: func(cards []api.Card) ([]api.Card, error) {
				return a.applyListFilters(cards, list.ID)
			},
			OnEscape: func() {
				a.showListsList()
			},
			OnFilter: func() {
				a.showListFilter(list.ID)
			},
		},
	)
}

// updateListViewTitle updates the title of the list view
func (a *App) updateListViewTitle(list collections.List) {
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

	title := fmt.Sprintf("%s - Page %d/%d", list.Name, a.currentPage, totalPages)

	// Get filter query for this list (stored in a map)
	listFilterQuery := a.getListFilterQuery(list.ID)
	if listFilterQuery != "" {
		// Truncate long filter queries
		filterDisplay := listFilterQuery
		if len(filterDisplay) > 25 {
			filterDisplay = filterDisplay[:22] + "..."
		}
		title += fmt.Sprintf(" | Filter: %s", filterDisplay)
	}
	title += fmt.Sprintf(" | Sort: %s %s", sortFieldDisplay, sortDir)
	if totalUniqueCards > 0 {
		title += fmt.Sprintf(" | Total: %d unique cards", totalUniqueCards)
		// Count total printings in list
		totalPrintings := len(list.Cards)
		title += fmt.Sprintf(" (%d printings)", totalPrintings)
	} else {
		title += " | Empty"
	}

	if a.table != nil {
		a.table.SetTitle(title)
	}
}

// applyListFilters applies the current list filter query to cards
func (a *App) applyListFilters(cards []api.Card, listID string) ([]api.Card, error) {
	filterQuery := a.getListFilterQuery(listID)
	return applyCardFilters(cards, filterQuery)
}

// getListFilterQuery gets the filter query for a specific list
func (a *App) getListFilterQuery(listID string) string {
	// For now, return empty string (no filtering)
	// In the future, we could store filter queries per list in a map
	return ""
}

// showListFilter shows the filter form for a list
func (a *App) showListFilter(listID string) {
	// TODO: Implement list filter form
	// For now, just show a message that filtering is not yet implemented
	modal := tview.NewModal().
		SetText("List filtering is not yet implemented.").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("list_filter_not_implemented")
		})
	modal.SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Filter List[white] ").
		SetTitleColor(tcell.ColorYellow)

	a.pages.AddPage("list_filter_not_implemented", modal, true, true)
	a.pages.SwitchToPage("list_filter_not_implemented")
	a.app.SetFocus(modal)
}
