package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showDecksList displays a list of all user decks
func (a *App) showDecksList() {
	// Remove old decks list page if it exists
	if a.pages.HasPage("decks_list") {
		a.pages.RemovePage("decks_list")
	}

	// TODO: Load decks from file
	// For now, show empty state
	decksList := tview.NewList()
	decksList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]My Decks[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Add placeholder message
	decksList.AddItem("No decks yet", "Create your first deck from a card detail view", 0, nil)

	// Handle ESC to go back to collection menu
	decksList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("decks_list") {
				a.pages.RemovePage("decks_list")
			}
			if a.pages.HasPage("collection_menu") {
				a.pages.SwitchToPage("collection_menu")
			} else {
				a.showCollectionMenu()
			}
			return nil
		}
		return event
	})

	// Center the list
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(decksList, 80, 0, true).
		AddItem(nil, 0, 1, false)

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(horizontalFlex, 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("decks_list", verticalFlex, true, true)
	a.pages.SwitchToPage("decks_list")
	a.app.SetFocus(decksList)
}

