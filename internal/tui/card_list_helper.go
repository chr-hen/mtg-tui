package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/rivo/tview"
)

// showCardsFromSource is a generic helper that handles the complete flow of:
// loading → caching → sorting → filtering → grouping → paginating → displaying cards
//
// This eliminates duplication across collection_view, list_view, search results, etc.
func (a *App) showCardsFromSource(
	cacheKey string,
	loadCards func() ([]api.Card, error),
	config models.CardListConfig,
) {
	// Remove old page if it exists
	if a.pages.HasPage(config.PageName) {
		a.pages.RemovePage(config.PageName)
	}

	// 1. Load cards from source (if not cached)
	allCards, exists := a.allMatchingCards[cacheKey]
	if !exists {
		var err error
		allCards, err = loadCards()
		if err != nil {
			// Show error modal
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error loading cards: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("card_load_error")
					config.OnEscape()
				})
			a.pages.AddPage("card_load_error", errorModal, true, true)
			return
		}
		
		// Cache the cards
		if a.allMatchingCards == nil {
			a.allMatchingCards = make(map[string][]api.Card)
		}
		a.allMatchingCards[cacheKey] = allCards
	}

	// 2. Sort the cards according to current sort settings
	util.SortCards(&allCards, a.sortField, a.sortAscending)
	a.allMatchingCards[cacheKey] = allCards

	// 3. Apply filters
	filteredCards, err := config.FilterFunc(allCards)
	if err != nil {
		// Show error modal
		errorModal := tview.NewModal().
			SetText(fmt.Sprintf("Error applying filters: %v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				a.pages.RemovePage("card_filter_error")
				if a.pages.HasPage(config.PageName) {
					a.pages.SwitchToPage(config.PageName)
					if a.table != nil {
						a.app.SetFocus(a.table)
					}
				}
			})
		a.pages.AddPage("card_filter_error", errorModal, true, true)
		return
	}

	// 4. Group filtered cards by name
	cardGroups := util.GroupCardsByName(filteredCards, a.settings.ShowArenaCards, a.settings.ShowTypeCard, a.settings.ShowArtCards)

	// 5. Set up for display
	a.currentQuery = cacheKey
	a.currentPage = 1

	// 6. Use unified pagination
	pageSize := a.settings.PageSize
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	// 7. Show unified card list
	a.showUnifiedCardList(config)
}

