package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showDecksList displays a list of all user decks
func (a *App) showDecksList() {
	// Remove old decks list page if it exists
	if a.pages.HasPage("decks_list") {
		a.pages.RemovePage("decks_list")
	}

	// Load decks
	decksCollection, err := collections.LoadDecks()
	if err != nil {
		// Show error but continue with empty collection
		decksCollection = &collections.DecksCollection{Decks: []collections.Deck{}}
	}

	// Create list for existing decks
	decksList := tview.NewList()
	decksList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]My Decks[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	decksList.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Add existing decks
	if len(decksCollection.Decks) == 0 {
		decksList.AddItem("No decks yet", "Create your first deck", 0, nil)
	} else {
		for _, deck := range decksCollection.Decks {
			deckCopy := deck // Capture for closure
			deckInfo := fmt.Sprintf("%s • %d cards", deck.Format, len(deck.MainDeck))
			decksList.AddItem(deck.Name, deckInfo, 0, func() {
				a.showDeckDetails(deckCopy)
			})
		}
	}

	// Handle deck selection
	decksList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if len(decksCollection.Decks) > 0 && index < len(decksCollection.Decks) {
			a.showDeckDetails(decksCollection.Decks[index])
		}
	})

	// Store callbacks
	createCallback := func() {
		a.showDeckCreationForm()
	}
	backCallback := func() {
		if a.pages.HasPage("decks_list") {
			a.pages.RemovePage("decks_list")
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

	// Add "Create New Deck" button
	form.AddButton("Create New Deck (n)", createCallback)

	// Add "Back" button
	form.AddButton("Back (b)", backCallback)

	// Style buttons
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	// Handle ESC and navigation for decks list
	decksList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backCallback()
			return nil
		}

		// Handle shortcuts
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'n':
				// Create New Deck shortcut
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
				current := decksList.GetCurrentItem()
				if current < decksList.GetItemCount()-1 {
					decksList.SetCurrentItem(current + 1)
				} else {
					// Move focus to buttons when at bottom
					a.app.SetFocus(form)
				}
				return nil
			case 'k':
				// Move up
				current := decksList.GetCurrentItem()
				if current > 0 {
					decksList.SetCurrentItem(current - 1)
				}
				return nil
			case 'h', 'l':
				// Move focus to buttons
				a.app.SetFocus(form)
				return nil
			}
		}

		// Allow Enter to select deck
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
				// Create New Deck shortcut
				createCallback()
				return nil
			case 'b':
				// Back shortcut
				backCallback()
				return nil
			case 'k':
				// Move focus to decks list
				a.app.SetFocus(decksList)
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

	// Create flex layout: decks list on top, buttons at bottom
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(decksList, 0, 1, true). // Decks list (flexible, takes most space)
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

	a.pages.AddPage("decks_list", verticalFlex, true, true)
	a.pages.SwitchToPage("decks_list")
	a.app.SetFocus(decksList)
}

// showDeckDetails displays details for a specific deck
func (a *App) showDeckDetails(deck collections.Deck) {
	// TODO: Implement deck details view
	// For now, just show a simple modal
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Deck: %s\nFormat: %s\nCards: %d", deck.Name, deck.Format, len(deck.MainDeck))).
		AddButtons([]string{"Back"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if a.pages.HasPage("deck_details") {
				a.pages.RemovePage("deck_details")
			}
			a.showDecksList()
		})
	modal.SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Deck Details[white] ").
		SetTitleColor(tcell.ColorYellow)

	a.pages.AddPage("deck_details", modal, true, true)
	a.pages.SwitchToPage("deck_details")
	a.app.SetFocus(modal)
}
