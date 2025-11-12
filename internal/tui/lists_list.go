package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showListsList displays a list of all user lists
func (a *App) showListsList() {
	// Remove old lists list page if it exists
	if a.pages.HasPage("lists_list") {
		a.pages.RemovePage("lists_list")
	}

	// Load lists
	listsCollection, err := collections.LoadLists()
	if err != nil {
		// Show error but continue with empty collection
		listsCollection = &collections.ListsCollection{Lists: []collections.List{}}
	}

	// Create list for existing lists
	listsList := tview.NewList()
	listsList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]My Lists[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	listsList.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Add existing lists
	if len(listsCollection.Lists) == 0 {
		listsList.AddItem("No lists yet", "Create your first list", 0, nil)
		} else {
		for _, list := range listsCollection.Lists {
			listInfo := fmt.Sprintf("%d cards", len(list.Cards))
			if list.Description != "" {
				listInfo = list.Description + " • " + listInfo
			}
			listsList.AddItem(list.Name, listInfo, 0, nil)
		}
	}

	// Handle list selection
	listsList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if len(listsCollection.Lists) > 0 && index < len(listsCollection.Lists) {
			a.showListView(listsCollection.Lists[index])
		}
	})

	// Store callbacks
	createCallback := func() {
		a.showListCreationForm()
	}
	backCallback := func() {
		if a.pages.HasPage("lists_list") {
			a.pages.RemovePage("lists_list")
		}
		if a.pages.HasPage("collection_menu") {
			a.pages.SwitchToPage("collection_menu")
		} else {
			a.showCollectionMenu()
		}
	}

	// Create form for buttons
	form := tview.NewForm()
	form.SetBorder(false).
		SetBackgroundColor(tcell.ColorBlack)

	// Add "Create New List" button
	form.AddButton("Create New List (n)", createCallback)

	// Add "Back" button
	form.AddButton("Back (b)", backCallback)

	// Style buttons
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	// Handle ESC and navigation for lists list
	listsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backCallback()
			return nil
		}

		// Handle shortcuts
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'n':
				// Create New List shortcut
				createCallback()
				return nil
			case 'b':
				// Back shortcut
				backCallback()
				return nil
			}
		}

		// Handle Vim motions
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'j':
				// Move down
				current := listsList.GetCurrentItem()
				if current < listsList.GetItemCount()-1 {
					listsList.SetCurrentItem(current + 1)
				} else {
					// Move focus to buttons when at bottom
					a.app.SetFocus(form)
				}
				return nil
			case 'k':
				// Move up
				current := listsList.GetCurrentItem()
				if current > 0 {
					listsList.SetCurrentItem(current - 1)
				}
				return nil
			case 'h', 'l':
				// Move focus to buttons
				a.app.SetFocus(form)
				return nil
			}
		}

		// Allow Enter to select list
		if event.Key() == tcell.KeyEnter {
			return event
		}

		return event
	})

	// Handle button navigation
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backCallback()
			return nil
		}

		// Handle shortcuts
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'n':
				// Create New List shortcut
				createCallback()
				return nil
			case 'b':
				// Back shortcut
				backCallback()
				return nil
			case 'k':
				// Move focus to lists list
				a.app.SetFocus(listsList)
				return nil
			case 'j':
				// Already at bottom (buttons)
				return nil
			case 'h', 'l':
				// Tab between buttons
				return event
			}
		}

		// Allow Tab/Enter for buttons
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyEnter {
			return event
		}

		return event
	})

	// Create flex layout: lists list on top, buttons at bottom
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(listsList, 0, 1, true). // Lists list (flexible, takes most space)
		AddItem(form, 3, 0, false)      // Buttons form (fixed height)

	// Center the main flex
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, 80, 0, true).
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

