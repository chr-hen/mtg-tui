package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/gdamore/tcell/v2"
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
		// Cache all matching cards
		a.allMatchingCards[a.currentQuery] = allCards
		// Initialize page cache for this query
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
	}

	// Get current page from cache or calculate it
	pageSize := 10
	totalCards := len(allCards)
	startIdx := (a.currentPage - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalCards {
		// Empty page
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

	// Pre-load next 2 pages in background
	go a.preloadPages(a.currentQuery, a.currentPage, allCards, pageSize)

	// Create table view for multi-line card display
	a.table = tview.NewTable()
	a.updateListTitle()
	a.table.SetBorder(true)
	a.table.SetBorderColor(tcell.ColorYellow)
	a.table.SetTitleColor(tcell.ColorYellow)
	a.table.SetSelectable(true, false) // Selectable rows, not columns
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

	// Populate table
	a.populateList()

	// Set up VIM keybindings
	a.setupVimKeybindings()

	// Handle ESC to go back to menu
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("list") {
				a.pages.RemovePage("list")
			}
			a.table = nil
			if a.pages.HasPage("menu") {
				a.pages.SwitchToPage("menu")
				if a.menu != nil {
					a.app.SetFocus(a.menu)
				}
			} else {
				a.showMainMenu()
			}
			return nil
		}

		// Handle VIM keys
		if event.Key() == tcell.KeyRune {
			rune := event.Rune()
			if rune == 'j' {
				// Move down - skip to next card (4 rows per card: 3 content + 1 blank)
				row, _ := a.table.GetSelection()
				nextRow := row + 4
				if nextRow < a.table.GetRowCount() {
					a.table.Select(nextRow, 0)
				}
				return nil
			}
			if rune == 'k' {
				// Move up - skip to previous card
				row, _ := a.table.GetSelection()
				nextRow := row - 4
				if nextRow >= 0 {
					a.table.Select(nextRow, 0)
				}
				return nil
			}
			if rune == 'l' {
				// Next page
				if a.pagination.HasMore {
					a.loadNextPage()
				}
				return nil
			}
			if rune == 'h' {
				// Previous page
				if a.currentPage > 1 {
					a.loadPreviousPage()
				}
				return nil
			}
		}

		return event
	})

	// Handle selection changed to skip blank rows
	a.table.SetSelectionChangedFunc(func(row, column int) {
		// If we're on a blank row (row % 4 == 3), move to next card
		if row%4 == 3 {
			nextRow := row + 1
			if nextRow < a.table.GetRowCount() {
				a.table.Select(nextRow, 0)
			} else {
				// If we're at the end, go to previous card
				prevRow := row - 1
				if prevRow >= 0 {
					a.table.Select(prevRow, 0)
				}
			}
		}
	})

	// Handle Enter key to show card detail
	a.table.SetSelectedFunc(func(row, column int) {
		// Calculate card index - each card is 4 rows (3 content + 1 blank)
		cardIndex := row / 4
		if cardIndex >= 0 && cardIndex < len(a.cards) {
			a.showCardDetail(a.pages, a.cards[cardIndex])
		}
	})

	// Remove old list page if it exists
	if a.pages.HasPage("list") {
		a.pages.RemovePage("list")
	}

	// Add new table page
	a.pages.AddPage("list", a.table, true, true)
	a.pages.SwitchToPage("list")

	// Set focus
	a.app.SetFocus(a.table)
}

func (a *App) populateList() {
	if a.table == nil {
		return
	}
	a.table.Clear()

	row := 0
	for i, card := range a.cards {
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

		// Row 3: Rarity • Mana Cost • Set: SetCode (gray)
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
		if card.SetCode != "" {
			infoParts = append(infoParts, fmt.Sprintf("Set: %s", card.SetCode))
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
		if i < len(a.cards)-1 {
			cell = tview.NewTableCell("").
				SetSelectable(false)
			a.table.SetCell(row, 0, cell)
			row++
		}
	}
}

func (a *App) updateListTitle() {
	pageSize := 10
	totalPages := 1
	if a.pagination.TotalCards > 0 {
		// Calculate total pages: ceiling division
		totalPages = (a.pagination.TotalCards + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}
	}

	title := fmt.Sprintf("MTG Cards - Page %d/%d", a.currentPage, totalPages)
	if a.currentQuery != "" {
		// Truncate long queries
		queryDisplay := a.currentQuery
		if len(queryDisplay) > 40 {
			queryDisplay = queryDisplay[:37] + "..."
		}
		title += fmt.Sprintf(" | Query: %s", queryDisplay)
	}
	if a.pagination.TotalCards > 0 {
		title += fmt.Sprintf(" | Total: %d", a.pagination.TotalCards)
	}
	if a.table != nil {
		a.table.SetTitle(title)
	}
}

func (a *App) setupVimKeybindings() {
	// VIM keybindings are now handled in showCardList's SetInputCapture
	// This method is kept for compatibility but the actual handling is in handleVimKeys
}

func (a *App) handleVimKeys(event *tcell.EventKey) *tcell.EventKey {
	// VIM keys are now handled directly in showCardList's SetInputCapture
	// This method is kept for compatibility but not actively used
	return event
}
