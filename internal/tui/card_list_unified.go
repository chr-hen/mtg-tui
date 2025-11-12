package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/components"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showUnifiedCardList renders a paginated card list with common functionality
func (a *App) showUnifiedCardList(config models.CardListConfig) {
	// 1. Create table with common styling
	a.table = tview.NewTable()
	config.TitleFunc() // Update title
	a.table.SetBorder(true)
	a.table.SetBorderColor(tcell.ColorYellow)
	a.table.SetTitleColor(tcell.ColorYellow)
	a.table.SetSelectable(true, false)
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

	// 2. Populate table
	a.populateListUnified()

	// 3. Setup keybindings (unified)
	a.setupCardListKeybindings(config)

	// 4. Setup blank row handling
	a.setupBlankRowHandling()

	// 5. Setup selection handler
	a.setupCardSelection()

	// 6. Create footer using extracted component
	footer := components.CreateCardListFooter(config.FooterText)

	// 7. Create layout
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.table, 0, 1, true).
		AddItem(footer, 5, 0, false)

	// 8. Add and show page
	if a.pages.HasPage(config.PageName) {
		a.pages.RemovePage(config.PageName)
	}
	a.pages.AddPage(config.PageName, mainFlex, true, true)
	a.pages.SwitchToPage(config.PageName)
	a.app.SetFocus(a.table)
}

// setupCardListKeybindings handles all keyboard input for card lists
func (a *App) setupCardListKeybindings(config models.CardListConfig) {
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// ESC key
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage(config.PageName) {
				a.pages.RemovePage(config.PageName)
			}
			a.table = nil
			config.OnEscape()
			return nil
		}

		// Arrow keys (unified logic for all views)
		if event.Key() == tcell.KeyDown {
			row, _ := a.table.GetSelection()
			// If we're on a blank row, move to next card
			if row%4 == 3 {
				nextRow := ((row / 4) + 1) * 4
				if nextRow < a.table.GetRowCount() {
					a.table.Select(nextRow, 0)
				}
				return nil
			}
			// Otherwise, move to next card's first row
			currentCard := row / 4
			nextCard := currentCard + 1
			nextRow := nextCard * 4
			if nextRow < a.table.GetRowCount() {
				a.table.Select(nextRow, 0)
			}
			return nil
		}
		if event.Key() == tcell.KeyUp {
			row, _ := a.table.GetSelection()
			// If we're on a blank row, move to previous card
			if row%4 == 3 {
				prevRow := ((row / 4) - 1) * 4 + 2 // Last row of previous card
				if prevRow >= 0 {
					a.table.Select(prevRow, 0)
				}
				return nil
			}
			// Otherwise, move to previous card's first row
			currentCard := row / 4
			prevCard := currentCard - 1
			if prevCard >= 0 {
				prevRow := prevCard * 4
				a.table.Select(prevRow, 0)
			}
			return nil
		}

		// Rune-based keys
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'f':
				config.OnFilter()
				return nil
			case 'j':
				// Move down - skip to next card (4 rows per card: 3 content + 1 blank)
				row, _ := a.table.GetSelection()
				nextRow := row + 4
				if nextRow < a.table.GetRowCount() {
					a.table.Select(nextRow, 0)
				}
				return nil
			case 'k':
				// Move up - skip to previous card
				row, _ := a.table.GetSelection()
				prevRow := row - 4
				if prevRow >= 0 {
					a.table.Select(prevRow, 0)
				}
				return nil
			case 'l':
				// Next page
				if a.pagination.HasMore {
					a.currentPage++
					a.reloadCurrentList(config)
				}
				return nil
			case 'h':
				// Previous page
				if a.currentPage > 1 {
					a.currentPage--
					a.reloadCurrentList(config)
				}
				return nil
			case 's':
				a.cycleSortFieldUnified(config)
				return nil
			case 'S':
				a.sortAscending = !a.sortAscending
				a.resortAndRefresh(config)
				return nil
			default:
				// Check for extra keybinds
				if config.ExtraKeybinds != nil {
					if handler, exists := config.ExtraKeybinds[event.Rune()]; exists {
						handler()
						return nil
					}
				}
			}
		}
		return event
	})
}

// setupBlankRowHandling prevents selection of separator rows
func (a *App) setupBlankRowHandling() {
	skippingBlankRow := false
	a.table.SetSelectionChangedFunc(func(row, column int) {
		if skippingBlankRow {
			return
		}
		if row%4 == 3 { // Blank row
			skippingBlankRow = true
			// We're on a blank row - immediately move to a valid row
			// Try to move to the first row of the next card
			nextRow := ((row / 4) + 1) * 4
			if nextRow < a.table.GetRowCount() {
				a.table.Select(nextRow, 0)
			} else {
				// If we're at the end, go to the last row of the previous card
				prevRow := ((row / 4) - 1) * 4 + 2
				if prevRow >= 0 {
					a.table.Select(prevRow, 0)
				} else {
					// Fallback: go to row 0
					a.table.Select(0, 0)
				}
			}
			skippingBlankRow = false
		}
	})
}

// setupCardSelection handles Enter key to show card details
func (a *App) setupCardSelection() {
	a.table.SetSelectedFunc(func(row, column int) {
		if row%4 == 3 {
			return
		}
		cardIndex := row / 4
		if cardIndex >= 0 && cardIndex < len(a.cardGroups) {
			a.showCardDetail(a.pages, a.cardGroups[cardIndex])
		}
	})
}

// reloadCurrentList reloads the current page with filters applied
func (a *App) reloadCurrentList(config models.CardListConfig) {
	// Get all matching cards for this query
	allCards, exists := a.allMatchingCards[a.currentQuery]
	if !exists {
		// Cannot reload without cached data
		return
	}

	// Apply appropriate filters
	filteredCards, err := config.FilterFunc(allCards)
	if err != nil {
		// Silently fail - show unfiltered cards
		filteredCards = allCards
	}

	// Group filtered cards by name
	cardGroups := util.GroupCardsByName(filteredCards, a.settings.ShowArenaCards, a.settings.ShowTypeCard, a.settings.ShowArtCards)

	// Use cached data with unified pagination
	pageSize := a.settings.PageSize
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	a.populateListUnified()
	config.TitleFunc()
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}

// populateListUnified populates the table with card groups
// This replaces the existing populateList() function from card_list.go
func (a *App) populateListUnified() {
	if a.table == nil {
		return
	}
	a.table.Clear()

	row := 0
	for i, group := range a.cardGroups {
		card := group.Card // Use the canonical card for display

		// Row 1: Card Name (white)
		cell := tview.NewTableCell(card.Name).
			SetTextColor(tcell.ColorWhite).
			SetSelectable(true)
		a.table.SetCell(row, 0, cell)
		row++

		// Row 2: Type (gray)
		typeLine := card.TypeLine
		if typeLine == "" {
			typeLine = " "
		}
		cell = tview.NewTableCell(typeLine).
			SetTextColor(tcell.ColorGray).
			SetSelectable(true)
		a.table.SetCell(row, 0, cell)
		row++

		// Row 3: Rarity • Mana Cost (gray) - removed set info since we have multiple printings
		infoParts := []string{}
		if card.Rarity != "" {
			rarity := strings.ToLower(card.Rarity)
			if len(rarity) > 0 {
				rarity = strings.ToUpper(rarity[:1]) + rarity[1:]
			}
			infoParts = append(infoParts, rarity)
		}
		if card.ManaCost != "" {
			infoParts = append(infoParts, card.ManaCost)
		}
		// Show number of printings if more than one
		if len(group.Printings) > 1 {
			infoParts = append(infoParts, fmt.Sprintf("(%d printings)", len(group.Printings)))
		}

		infoLine := strings.Join(infoParts, " • ")
		if infoLine == "" {
			infoLine = " "
		}
		cell = tview.NewTableCell(infoLine).
			SetTextColor(tcell.ColorGray).
			SetSelectable(true)
		a.table.SetCell(row, 0, cell)
		row++

		// Add blank row separator (except after last card)
		// Use a space character instead of empty string to ensure it's visible but non-interactive
		if i < len(a.cardGroups)-1 {
			cell = tview.NewTableCell(" ").
				SetSelectable(false).
				SetExpansion(1)
			a.table.SetCell(row, 0, cell)
			row++
		}
	}
}

// cycleSortFieldUnified cycles through available sort fields (unified version)
func (a *App) cycleSortFieldUnified(config models.CardListConfig) {
	sortFields := []string{"name", "released", "set", "rarity", "color", "cmc", "power", "toughness"}
	currentIdx := -1
	for i, field := range sortFields {
		if field == a.sortField {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 {
		a.sortField = "name"
	} else {
		nextIdx := (currentIdx + 1) % len(sortFields)
		a.sortField = sortFields[nextIdx]
	}
	// Re-sort and refresh
	a.resortAndRefresh(config)
}

// resortAndRefresh re-sorts cards and refreshes the view
func (a *App) resortAndRefresh(config models.CardListConfig) {
	if allCards, exists := a.allMatchingCards[a.currentQuery]; exists {
		util.SortCards(&allCards, a.sortField, a.sortAscending)
		a.allMatchingCards[a.currentQuery] = allCards
		// Clear page cache
		if a.pageCache[a.currentQuery] != nil {
			a.pageCache[a.currentQuery] = make(map[int][]api.Card)
		}
		a.currentPage = 1
		a.reloadCurrentList(config)
	}
}

