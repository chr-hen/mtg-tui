package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) showMainMenu() {
	// Remove old menu page and search page to prevent artifacts
	if a.pages.HasPage("menu") {
		a.pages.RemovePage("menu")
	}
	if a.pages.HasPage("search") {
		a.pages.RemovePage("search")
	}
	if a.pages.HasPage("advanced") {
		a.pages.RemovePage("advanced")
	}

	asciiArt := `
  __  __ _____ ____   _____ _   _ ___ 
 |  \/  |_   _/ ___| |_   _| | | |_ _|
 | |\/| | | || |  _    | | | | | || | 
 | |  | | | || |_| |   | | | |_| || | 
 |_|  |_| |_| \____|   |_|  \___/|___|
`

	// Create text view for ASCII art
	artView := tview.NewTextView().
		SetText(asciiArt).
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Create subtitle with styling
	subtitle := tview.NewTextView().
		SetText("[gray]Magic: The Gathering Card - Text User Interface[white]").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Create menu list with better styling and separators
	menu := tview.NewList()
	a.menu = menu // Store reference for later use

	// Store callbacks for later use
	callbacks := []func(){
		func() { a.showSearchInput() },
		nil, // separator
		func() { a.showAdvancedSearch() },
		nil, // separator
		func() { a.showSettingsScreen() },
		nil, // separator
		func() { a.app.Stop() },
	}

	menu.AddItem("Search for Card", "Enter a card name or search query", 's', callbacks[0]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("Advanced Search", "Use Scryfall syntax for detailed filtering", 'a', callbacks[2]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("Settings", "Configure application preferences", 'c', callbacks[4]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("Quit", "Exit the application", 'q', callbacks[6])

	menu.SetBorder(true).
		SetTitle(" [yellow]Navigation[white] ").
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	menu.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Override selected func to skip separators
	menu.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Skip separator items - move to next item
		if strings.Contains(mainText, "─") {
			// Try to move to next item
			if index < menu.GetItemCount()-1 {
				menu.SetCurrentItem(index + 1)
			} else if index > 0 {
				menu.SetCurrentItem(index - 1)
			}
			return
		}
		// Call the original callback if it exists
		if index < len(callbacks) && callbacks[index] != nil {
			callbacks[index]()
		} else {
			// Handle direct index mapping for menu items
			if index == 0 {
				a.showSearchInput()
			} else if index == 2 {
				a.showAdvancedSearch()
			} else if index == 4 {
				a.app.Stop()
			}
		}
	})

	// Remove SetDoneFunc so ESC doesn't quit - only Ctrl+C can quit now
	menu.SetDoneFunc(func() {
		// Do nothing - prevent accidental quitting with ESC
	})

	// Handle navigation to skip separators
	menu.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// If we land on a separator, move to the next item
		if strings.Contains(mainText, "─") {
			if index < menu.GetItemCount()-1 {
				menu.SetCurrentItem(index + 1)
			} else if index > 0 {
				menu.SetCurrentItem(index - 1)
			}
		}
	})

	// Add VIM keybindings for menu navigation
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'j':
				// Move down
				current := menu.GetCurrentItem()
				next := current + 1
				// Skip separators
				for next < menu.GetItemCount() {
					mainText, _ := menu.GetItemText(next)
					if !strings.Contains(mainText, "─") {
						menu.SetCurrentItem(next)
						return nil
					}
					next++
				}
				return nil
			case 'k':
				// Move up
				current := menu.GetCurrentItem()
				prev := current - 1
				// Skip separators
				for prev >= 0 {
					mainText, _ := menu.GetItemText(prev)
					if !strings.Contains(mainText, "─") {
						menu.SetCurrentItem(prev)
						return nil
					}
					prev--
				}
				return nil
			}
		}
		return event
	})

	// Create a flex container to center everything vertically
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false). // Top spacer
		AddItem(artView, 6, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(nil, 1, 0, false).  // Spacer between title and menu
		AddItem(menu, 12, 0, true). // Menu with fixed height to show all items (5 items + separators)
		AddItem(nil, 0, 1, false)   // Bottom spacer

	// Create another flex to center horizontally
	centeredFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(flex, 70, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("menu", centeredFlex, true, true)
	a.pages.SwitchToPage("menu")

	// Only queue updates if the app is already running
	// Otherwise just set focus directly (for initial menu creation)
	if a.pages.HasFocus() {
		a.app.QueueUpdateDraw(func() {
			a.app.SetFocus(menu)
		})
	} else {
		a.app.SetFocus(menu)
	}
}

