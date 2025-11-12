package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/gdamore/tcell/v2"
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
	pageSize := 10
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	// Create table view for multi-line card display
	a.table = tview.NewTable()
	a.updateCollectionTitle()
	a.table.SetBorder(true)
	a.table.SetBorderColor(tcell.ColorYellow)
	a.table.SetTitleColor(tcell.ColorYellow)
	a.table.SetSelectable(true, false) // Selectable rows, not columns
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

	// Populate table
	a.populateList()

	// Set up VIM keybindings
	a.setupVimKeybindings()

	// Handle ESC to go back to collection menu
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("collection") {
				a.pages.RemovePage("collection")
			}
			a.table = nil
			if a.pages.HasPage("collection_menu") {
				a.pages.SwitchToPage("collection_menu")
			} else {
				a.showCollectionMenu()
			}
			return nil
		}

		// Handle arrow keys - always jump by entire cards to avoid blank rows
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

		// Handle VIM keys
		if event.Key() == tcell.KeyRune {
			rune := event.Rune()
			if rune == 'f' {
				// Open filter panel
				a.showCollectionFilter()
				return nil
			}
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
				prevRow := row - 4
				if prevRow >= 0 {
					a.table.Select(prevRow, 0)
				}
				return nil
			}
			if rune == 'l' {
				// Next page
				if a.pagination.HasMore {
					a.currentPage++
					a.loadCollectionPage()
				}
				return nil
			}
			if rune == 'h' {
				// Previous page
				if a.currentPage > 1 {
					a.currentPage--
					a.loadCollectionPage()
				}
				return nil
			}
			if rune == 's' {
				// Cycle through sort options
				a.cycleSortField()
				return nil
			}
			if rune == 'S' {
				// Toggle sort direction
				a.sortAscending = !a.sortAscending
				// Re-sort and refresh
				if allCards, exists := a.allMatchingCards["collection"]; exists {
					util.SortCards(&allCards, a.sortField, a.sortAscending)
					a.allMatchingCards["collection"] = allCards
					a.currentPage = 1
					a.showCollection()
				}
				return nil
			}
		}

		return event
	})

	// Handle selection changed to skip blank rows
	// Use a flag to prevent recursive calls
	skippingBlankRow := false
	a.table.SetSelectionChangedFunc(func(row, column int) {
		// Prevent recursive calls
		if skippingBlankRow {
			return
		}
		
		// Check if current row is a blank separator row (every 4th row starting from row 3: 3, 7, 11, etc.)
		if row%4 == 3 {
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

	// Handle Enter key to show card detail
	a.table.SetSelectedFunc(func(row, column int) {
		// Skip if we're on a blank row
		if row%4 == 3 {
			return
		}
		// Calculate card group index - each card is 4 rows (3 content + 1 blank)
		cardIndex := row / 4
		if cardIndex >= 0 && cardIndex < len(a.cardGroups) {
			a.showCardDetail(a.pages, a.cardGroups[cardIndex])
		}
	})

	// Create footer with keyboard shortcuts
	footerText := "f: filter | s: sort | S: direction | j/k: navigate | h/l: pages | Enter: view details | Esc: back"
	footerBox := tview.NewBox().
		SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" Controls ").
		SetTitleColor(tcell.ColorYellow)

	// Store footerText in closure for the draw function
	footerTextForDraw := footerText
	footerBox.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// Draw border manually to ensure it's visible
		defStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)

		// Draw horizontal lines
		for i := x + 1; i < x+width-1; i++ {
			screen.SetContent(i, y, '─', nil, defStyle)
			screen.SetContent(i, y+height-1, '─', nil, defStyle)
		}

		// Draw vertical lines
		for i := y + 1; i < y+height-1; i++ {
			screen.SetContent(x, i, '│', nil, defStyle)
			screen.SetContent(x+width-1, i, '│', nil, defStyle)
		}

		// Draw corners
		screen.SetContent(x, y, '┌', nil, defStyle)
		screen.SetContent(x+width-1, y, '┐', nil, defStyle)
		screen.SetContent(x, y+height-1, '└', nil, defStyle)
		screen.SetContent(x+width-1, y+height-1, '┘', nil, defStyle)

		// Draw title on top border
		title := " Controls "
		titleX := x + 2
		if titleX+len(title) < x+width-2 {
			for i, r := range title {
				screen.SetContent(titleX+i, y, r, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack))
			}
		}

		// Get inner rectangle (area inside border)
		innerX := x + 1
		innerY := y + 1
		innerWidth := width - 2
		innerHeight := height - 2

		// Draw text centered in the inner area
		textY := innerY + (innerHeight / 2)
		if textY >= innerY && textY < innerY+innerHeight && innerHeight > 0 && innerWidth > 0 {
			// Use Print to draw the text with white color
			tview.Print(screen, footerTextForDraw, innerX, textY, innerWidth, tview.AlignCenter, tcell.ColorWhite)
		}

		return innerX, innerY, innerWidth, innerHeight
	})

	footer := footerBox

	// Create a flex container with table and footer
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.table, 0, 1, true). // Table takes remaining space
		AddItem(footer, 5, 0, false)  // Footer with fixed height

	// Add collection page
	a.pages.AddPage("collection", mainFlex, true, true)
	a.pages.SwitchToPage("collection")

	// Set focus
	a.app.SetFocus(a.table)
}

// updateCollectionTitle updates the title of the collection view
func (a *App) updateCollectionTitle() {
	pageSize := 10
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
	pageSize := 10
	a.cardGroups, a.pagination = util.PaginateCardGroups(cardGroups, a.currentPage, pageSize)

	a.populateList()
	a.updateCollectionTitle()
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}


