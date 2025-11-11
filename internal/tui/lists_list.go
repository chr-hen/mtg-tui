package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showListsList displays a list of all user lists (cubes)
func (a *App) showListsList() {
	// Remove old lists list page if it exists
	if a.pages.HasPage("lists_list") {
		a.pages.RemovePage("lists_list")
	}

	// TODO: Load lists from file
	// For now, show empty state
	listsList := tview.NewList()
	listsList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]My Lists[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Add placeholder message
	listsList.AddItem("No lists yet", "Create your first list from a card detail view", 0, nil)

	// Handle ESC to go back to collection menu
	listsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("lists_list") {
				a.pages.RemovePage("lists_list")
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
		AddItem(listsList, 80, 0, true).
		AddItem(nil, 0, 1, false)

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(horizontalFlex, 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("lists_list", verticalFlex, true, true)
	a.pages.SwitchToPage("lists_list")
	a.app.SetFocus(listsList)
}

