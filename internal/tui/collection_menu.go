package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCollectionMenu displays the collection menu with options for Cards, Decks, and Lists
func (a *App) showCollectionMenu() {
	// Remove old collection menu page if it exists
	if a.pages.HasPage("collection_menu") {
		a.pages.RemovePage("collection_menu")
	}

	// Create menu list
	menu := tview.NewList()

	// Store callbacks
	callbacks := []func(){
		func() { a.showCollection() }, // My Cards - existing collection view
		func() { a.showWants() },      // My Wants - wanted cards view
		func() { a.showDecksList() },  // My Decks - list of decks
		func() { a.showListsList() },  // My Lists - list of cubes/lists
	}

	menu.AddItem("My Cards", "View your owned cards", 'c', callbacks[0]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("My Wants", "View your wanted cards", 'w', callbacks[1]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("My Decks", "View and manage your decks", 'd', callbacks[2]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("My Lists", "View and manage your lists", 'l', callbacks[3]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("Back", "Return to main menu", 'b', func() {
			if a.pages.HasPage("collection_menu") {
				a.pages.RemovePage("collection_menu")
			}
			if a.pages.HasPage("menu") {
				a.pages.SwitchToPage("menu")
				if a.menu != nil {
					a.app.SetFocus(a.menu)
				}
			} else {
				a.showMainMenu()
			}
		})

	menu.SetBorder(true).
		SetTitle(" [yellow]My Collection[white] ").
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	menu.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

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

	// Override selected func to skip separators
	menu.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Skip separator items
		if strings.Contains(mainText, "─") {
			return
		}
		// Handle based on index
		if index == 0 {
			a.showCollection()
		} else if index == 2 {
			a.showWants()
		} else if index == 4 {
			a.showDecksList()
		} else if index == 6 {
			a.showListsList()
		} else if index == 8 {
			// Back button
			if a.pages.HasPage("collection_menu") {
				a.pages.RemovePage("collection_menu")
			}
			if a.pages.HasPage("menu") {
				a.pages.SwitchToPage("menu")
				if a.menu != nil {
					a.app.SetFocus(a.menu)
				}
			} else {
				a.showMainMenu()
			}
		}
	})

	// Handle ESC key and Vim motions
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("collection_menu") {
				a.pages.RemovePage("collection_menu")
			}
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

		// Handle Vim motions
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

	// Center the menu horizontally and vertically
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(menu, 50, 0, true).  // Menu (fixed width, horizontally centered)
		AddItem(nil, 0, 1, false)    // Right spacer

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).  // Top spacer
		AddItem(horizontalFlex, 0, 1, true).  // Menu (flexible height)
		AddItem(nil, 0, 1, false)   // Bottom spacer

	a.pages.AddPage("collection_menu", verticalFlex, true, true)
	a.pages.SwitchToPage("collection_menu")
	a.app.SetFocus(menu)
}

