package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
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

	// Group owned cards by name
	cardGroups := a.groupCardsByName(ownedCards)

	// Store all groups for pagination and pre-loading
	// We'll store them in allMatchingCards with a special key for collection
	if a.allMatchingCards == nil {
		a.allMatchingCards = make(map[string][]api.Card)
	}
	// Convert cardGroups back to cards for caching (we'll need to flatten them)
	allCollectionCards := make([]api.Card, 0)
	for _, group := range cardGroups {
		allCollectionCards = append(allCollectionCards, group.Printings...)
	}
	a.allMatchingCards["collection"] = allCollectionCards

	// Set up for display
	a.currentQuery = "collection"
	a.currentPage = 1
	
	// Use unified pagination
	pageSize := 10
	a.paginateCardGroups(cardGroups, a.currentPage, pageSize)

	// Pre-load next 2 pages in background
	go a.preloadCardGroupPages("collection", cardGroups, a.currentPage, pageSize)

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

	// Handle ESC to go back to menu
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("collection") {
				a.pages.RemovePage("collection")
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
		// Calculate card group index - each card is 4 rows (3 content + 1 blank)
		cardIndex := row / 4
		if cardIndex >= 0 && cardIndex < len(a.cardGroups) {
			a.showCardDetail(a.pages, a.cardGroups[cardIndex])
		}
	})

	// Create footer with keyboard shortcuts
	footerText := "j/k: navigate, Enter: view details, Esc: back"
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

	title := fmt.Sprintf("My Collection - Page %d/%d", a.currentPage, totalPages)
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

	// Group owned cards by name
	cardGroups := a.groupCardsByName(allCollectionCards)

	// Use unified pagination
	pageSize := 10
	a.paginateCardGroups(cardGroups, a.currentPage, pageSize)

	// Pre-load next 2 pages in background
	go a.preloadCardGroupPages("collection", cardGroups, a.currentPage, pageSize)

	a.populateList()
	a.updateCollectionTitle()
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}


