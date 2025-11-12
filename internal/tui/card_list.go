package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/util"
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
			if rune == 'f' {
				// Open filter panel for search results
				a.showSearchFilter()
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
			if rune == 's' {
				// Cycle through sort options
				a.cycleSortField()
				return nil
			}
			if rune == 'S' {
				// Toggle sort direction
				a.sortAscending = !a.sortAscending
				// Re-sort and refresh
				if allCards, exists := a.allMatchingCards[a.currentQuery]; exists {
					util.SortCards(&allCards, a.sortField, a.sortAscending)
					a.allMatchingCards[a.currentQuery] = allCards
					a.pageCache[a.currentQuery] = make(map[int][]api.Card)
					a.currentPage = 1
					a.showCardList()
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

	// Create footer using Box with custom drawing to ensure text is visible
	footerText := "f: filter | s: sort | S: direction | j/k: navigate | h/l: pages | Enter: details | Esc: back"
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

	// Remove old list page if it exists
	if a.pages.HasPage("list") {
		a.pages.RemovePage("list")
	}

	// Add new table page with footer
	a.pages.AddPage("list", mainFlex, true, true)

	// Set focus
	a.app.SetFocus(a.table)
}

func (a *App) populateList() {
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

func (a *App) setupVimKeybindings() {
	// VIM keybindings are now handled in showCardList's SetInputCapture
	// This method is kept for compatibility but the actual handling is in handleVimKeys
}

func (a *App) handleVimKeys(event *tcell.EventKey) *tcell.EventKey {
	// VIM keys are now handled directly in showCardList's SetInputCapture
	// This method is kept for compatibility but not actively used
	return event
}

// cycleSortField cycles through available sort fields
func (a *App) cycleSortField() {
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
	if allCards, exists := a.allMatchingCards[a.currentQuery]; exists {
		util.SortCards(&allCards, a.sortField, a.sortAscending)
		a.allMatchingCards[a.currentQuery] = allCards
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
		a.currentPage = 1
		a.showCardList()
	}
}
