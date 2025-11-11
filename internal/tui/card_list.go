package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

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
		// Sort the cards according to current sort settings
		a.sortCards(&allCards)
		// Cache all matching cards
		a.allMatchingCards[a.currentQuery] = allCards
		// Initialize page cache for this query
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
	} else {
		// Re-sort cached cards if sort settings changed
		a.sortCards(&allCards)
		a.allMatchingCards[a.currentQuery] = allCards
		// Clear page cache when sort changes
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
	}

	// Group cards by name
	cardGroups := a.groupCardsByName(allCards)

	// Get current page from cache or calculate it
	pageSize := 10
	totalGroups := len(cardGroups)
	startIdx := (a.currentPage - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalGroups {
		// Empty page
		a.cardGroups = []CardGroup{}
		a.pagination = api.PaginationInfo{
			HasMore:    false,
			TotalCards: totalGroups,
		}
	} else {
		if endIdx > totalGroups {
			endIdx = totalGroups
		}

		// Get the page of grouped cards
		a.cardGroups = cardGroups[startIdx:endIdx]

		hasMore := endIdx < totalGroups
		a.pagination = api.PaginationInfo{
			HasMore:    hasMore,
			TotalCards: totalGroups,
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
					a.sortCards(&allCards)
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
	footerText := "s: sort | S: direction | j/k: navigate | h/l: pages | Enter: details | Esc: back"
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
		if i < len(a.cardGroups)-1 {
			cell = tview.NewTableCell("").
				SetSelectable(false)
			a.table.SetCell(row, 0, cell)
			row++
		}
	}
}

// groupCardsByName groups cards by their name, collecting all printings
func (a *App) groupCardsByName(cards []api.Card) []CardGroup {
	groupsMap := make(map[string]*CardGroup)

	for _, card := range cards {
		cardName := strings.ToLower(card.Name)
		if group, exists := groupsMap[cardName]; exists {
			// Add this printing to the existing group
			group.Printings = append(group.Printings, card)
		} else {
			// Create a new group
			groupsMap[cardName] = &CardGroup{
				Card:      card, // Use first occurrence as canonical
				Printings: []api.Card{card},
			}
		}
	}

	// Convert map to slice and sort by canonical card name
	groups := make([]CardGroup, 0, len(groupsMap))
	for _, group := range groupsMap {
		// Sort printings by set code, then collector number
		sort.Slice(group.Printings, func(i, j int) bool {
			if group.Printings[i].SetCode != group.Printings[j].SetCode {
				return group.Printings[i].SetCode < group.Printings[j].SetCode
			}
			return compareCollectorNumbers(group.Printings[i].CollectorNumber, group.Printings[j].CollectorNumber)
		})
		groups = append(groups, *group)
	}

	// Sort groups by canonical card name
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Card.Name) < strings.ToLower(groups[j].Card.Name)
	})

	return groups
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
		a.sortCards(&allCards)
		a.allMatchingCards[a.currentQuery] = allCards
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
		a.currentPage = 1
		a.showCardList()
	}
}

// sortCards sorts cards according to the current sort field and direction
func (a *App) sortCards(cards *[]api.Card) {
	sort.Slice(*cards, func(i, j int) bool {
		less := a.compareCards((*cards)[i], (*cards)[j])
		if !a.sortAscending {
			return !less
		}
		return less
	})
}

// compareCards compares two cards based on the current sort field
func (a *App) compareCards(card1, card2 api.Card) bool {
	switch a.sortField {
	case "name":
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "released":
		// Parse release dates and compare
		date1, err1 := time.Parse("2006-01-02", card1.ReleasedAt)
		date2, err2 := time.Parse("2006-01-02", card2.ReleasedAt)
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		return date1.Before(date2)
	case "set":
		// Sort by set code, then collector number
		if card1.SetCode != card2.SetCode {
			return card1.SetCode < card2.SetCode
		}
		return compareCollectorNumbers(card1.CollectorNumber, card2.CollectorNumber)
	case "rarity":
		// Rarity order: Common < Uncommon < Rare < Mythic
		rarityOrder := map[string]int{
			"common":   1,
			"uncommon": 2,
			"rare":     3,
			"mythic":   4,
		}
		rarity1 := rarityOrder[strings.ToLower(card1.Rarity)]
		rarity2 := rarityOrder[strings.ToLower(card2.Rarity)]
		if rarity1 == 0 {
			rarity1 = 99
		}
		if rarity2 == 0 {
			rarity2 = 99
		}
		if rarity1 != rarity2 {
			return rarity1 < rarity2
		}
		// If same rarity, sort by name
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "color":
		// Sort by number of colors, then by color identity
		if len(card1.Colors) != len(card2.Colors) {
			return len(card1.Colors) < len(card2.Colors)
		}
		// If same number of colors, compare color identity strings
		colors1 := strings.Join(card1.Colors, "")
		colors2 := strings.Join(card2.Colors, "")
		if colors1 != colors2 {
			return colors1 < colors2
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "cmc":
		if card1.CMC != card2.CMC {
			return card1.CMC < card2.CMC
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "power":
		power1, err1 := parsePowerToughness(card1.Power)
		power2, err2 := parsePowerToughness(card2.Power)
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		if power1 != power2 {
			return power1 < power2
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	case "toughness":
		tough1, err1 := parsePowerToughness(card1.Toughness)
		tough2, err2 := parsePowerToughness(card2.Toughness)
		if err1 != nil && err2 != nil {
			return false
		}
		if err1 != nil {
			return false
		}
		if err2 != nil {
			return true
		}
		if tough1 != tough2 {
			return tough1 < tough2
		}
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	default:
		return strings.ToLower(card1.Name) < strings.ToLower(card2.Name)
	}
}

// compareCollectorNumbers compares two collector numbers (helper function)
func compareCollectorNumbers(a, b string) bool {
	// Try to parse as integers first
	numA, errA := strconv.Atoi(a)
	numB, errB := strconv.Atoi(b)

	// If both are numeric, compare numerically
	if errA == nil && errB == nil {
		return numA < numB
	}

	// If one is numeric and one isn't, numeric comes first
	if errA == nil && errB != nil {
		return true
	}
	if errA != nil && errB == nil {
		return false
	}

	// Both are alphanumeric, compare as strings
	return strings.ToLower(a) < strings.ToLower(b)
}

// parsePowerToughness parses power/toughness values, handling * and other special values
func parsePowerToughness(value string) (float64, error) {
	if value == "" || value == "*" {
		return 0, fmt.Errorf("cannot parse special value")
	}
	return strconv.ParseFloat(value, 64)
}
