package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
)

// showDeckCreationForm displays the form for creating a new deck
func (a *App) showDeckCreationForm() {
	// Remove old form page if it exists
	if a.pages.HasPage("deck_create") {
		a.pages.RemovePage("deck_create")
	}

	// Create form
	form := tview.NewForm()
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Create New Deck - Press 'i' to edit[white] ").
		SetTitleColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorBlack)

	// Store form values
	var deckName string
	var deckDescription string
	var selectedFormat collections.DeckFormat

	// Format options
	formats := collections.GetAllFormats()
	formatStrings := make([]string, len(formats))
	for i, f := range formats {
		formatStrings[i] = string(f)
	}

	// Vim insert mode tracking
	insertMode := false
	currentFieldIndex := 0 // Track which field/button is focused (0 = Deck Name, 1 = Format, 2 = Description, 3+ = buttons)

	// Add Deck Name field
	nameField := tview.NewInputField().
		SetLabel("Deck Name").
		SetFieldWidth(40).
		SetChangedFunc(func(text string) {
			deckName = text
		})
	nameField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(nameField)

	// Add Format dropdown
	formatDropdown := tview.NewDropDown().
		SetLabel("Format").
		SetOptions(formatStrings, func(option string, optionIndex int) {
			selectedFormat = formats[optionIndex]
		})
	formatDropdown.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(formatDropdown)

	// Add Description field
	descField := tview.NewInputField().
		SetLabel("Description").
		SetFieldWidth(60).
		SetChangedFunc(func(text string) {
			deckDescription = text
		})
	descField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(descField)

	// Add buttons
	form.AddButton("Create", func() {
		// Validate
		if strings.TrimSpace(deckName) == "" {
			a.showErrorModal("Deck name is required")
			return
		}

		// Create deck
		deck := collections.Deck{
			ID:          uuid.New().String(),
			Name:        strings.TrimSpace(deckName),
			Format:      selectedFormat,
			Description: strings.TrimSpace(deckDescription),
			MainDeck:    []collections.DeckCard{},
			Sideboard:   []collections.DeckCard{},
		}

		// Load existing decks
		decksCollection, err := collections.LoadDecks()
		if err != nil {
			decksCollection = &collections.DecksCollection{Decks: []collections.Deck{}}
		}

		// Add new deck
		decksCollection.AddDeck(deck)

		// Save decks
		if err := collections.SaveDecks(decksCollection); err != nil {
			a.showErrorModal(fmt.Sprintf("Failed to save deck: %v", err))
			return
		}

		// Show success and go back to decks list
		if a.pages.HasPage("deck_create") {
			a.pages.RemovePage("deck_create")
		}
		a.showDecksList()
	})

	form.AddButton("Cancel", func() {
		if a.pages.HasPage("deck_create") {
			a.pages.RemovePage("deck_create")
		}
		a.showDecksList()
	})

	// Style form
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	// Handle Vim motions and insert mode
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// ESC handling
		if event.Key() == tcell.KeyEscape {
			if insertMode {
				// Exit insert mode
				insertMode = false
				form.SetTitle(" [yellow]Create New Deck - Press 'i' to edit[white] ")
				form.SetFocus(currentFieldIndex)
				return nil
			} else {
				// Cancel and go back
				if a.pages.HasPage("deck_create") {
					a.pages.RemovePage("deck_create")
				}
				a.showDecksList()
				return nil
			}
		}

		// If in insert mode, allow all input
		if insertMode {
			return event
		}

		// Normal mode - handle vim motions
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'i':
				// Enter insert mode
				insertMode = true
				form.SetTitle(" [yellow]Create New Deck - INSERT MODE (ESC to exit)[white] ")
				// Focus on the current field to allow editing
				form.SetFocus(currentFieldIndex)
				return nil
			case 'j':
				// Move down to next field/button
				// Form items are indexed 0-N, buttons start after that
				// We have 3 form items (0,1,2) and 2 buttons (3,4)
				maxIndex := form.GetFormItemCount() + 1 // Last button index
				if currentFieldIndex < maxIndex {
					currentFieldIndex++
					form.SetFocus(currentFieldIndex)
				}
				return nil
			case 'k':
				// Move up to previous field/button
				if currentFieldIndex > 0 {
					currentFieldIndex--
					form.SetFocus(currentFieldIndex)
				}
				return nil
			default:
				// Block other keys in normal mode
				return nil
			}
		}

		// Allow Tab and Backtab for navigation in normal mode
		if event.Key() == tcell.KeyTab {
			maxIndex := form.GetFormItemCount() + 1
			if currentFieldIndex < maxIndex {
				currentFieldIndex++
				form.SetFocus(currentFieldIndex)
			}
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			if currentFieldIndex > 0 {
				currentFieldIndex--
				form.SetFocus(currentFieldIndex)
			}
			return nil
		}

		// Allow Enter for buttons
		if event.Key() == tcell.KeyEnter {
			return event
		}

		// Block other keys in normal mode
		return nil
	})

	// Center the form
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(form, 80, 0, true).
		AddItem(nil, 0, 1, false)

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(horizontalFlex, 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("deck_create", verticalFlex, true, true)
	a.pages.SwitchToPage("deck_create")
	// Set initial focus to first field
	form.SetFocus(0)
	a.app.SetFocus(form)
}

// showErrorModal displays an error message
func (a *App) showErrorModal(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if a.pages.HasPage("error_modal") {
				a.pages.RemovePage("error_modal")
			}
		})
	modal.SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorRed).
		SetTitle(" [red]Error[white] ").
		SetTitleColor(tcell.ColorRed)

	a.pages.AddPage("error_modal", modal, true, true)
	a.pages.SwitchToPage("error_modal")
	a.app.SetFocus(modal)
}

