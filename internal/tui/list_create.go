package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
)

// showListCreationForm displays the form for creating a new list
func (a *App) showListCreationForm() {
	// Remove old form page if it exists
	if a.pages.HasPage("list_create") {
		a.pages.RemovePage("list_create")
	}

	// Create form
	form := tview.NewForm()
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Create New List - Press 'i' to edit[white] ").
		SetTitleColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorBlack)

	// Store form values
	var listName string
	var listDescription string
	var listNotes string

	// Vim insert mode tracking
	insertMode := false
	currentFieldIndex := 0 // Track which field/button is focused (0 = List Name, 1 = Description, 2 = Notes, 3+ = buttons)

	// Add List Name field
	nameField := tview.NewInputField().
		SetLabel("List Name").
		SetFieldWidth(40).
		SetChangedFunc(func(text string) {
			listName = text
		})
	nameField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(nameField)

	// Add Description field
	descField := tview.NewInputField().
		SetLabel("List Description").
		SetFieldWidth(60).
		SetChangedFunc(func(text string) {
			listDescription = text
		})
	descField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(descField)

	// Add Notes field
	notesField := tview.NewInputField().
		SetLabel("Other List Notes").
		SetFieldWidth(60).
		SetChangedFunc(func(text string) {
			listNotes = text
		})
	notesField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(notesField)

	// Add buttons
	form.AddButton("Create", func() {
		// Validate
		if strings.TrimSpace(listName) == "" {
			a.showErrorModal("List name is required")
			return
		}

		// Create list
		list := collections.List{
			ID:          uuid.New().String(),
			Name:        strings.TrimSpace(listName),
			Description: strings.TrimSpace(listDescription),
			Notes:       strings.TrimSpace(listNotes),
			Cards:       []collections.ListCard{},
		}

		// Load existing lists
		listsCollection, err := collections.LoadLists()
		if err != nil {
			listsCollection = &collections.ListsCollection{Lists: []collections.List{}}
		}

		// Add new list
		listsCollection.AddList(list)

		// Save lists
		if err := collections.SaveLists(listsCollection); err != nil {
			a.showErrorModal(fmt.Sprintf("Failed to save list: %v", err))
			return
		}

		// Show success and go back to lists list
		if a.pages.HasPage("list_create") {
			a.pages.RemovePage("list_create")
		}
		a.showListsList()
	})

	form.AddButton("Cancel", func() {
		if a.pages.HasPage("list_create") {
			a.pages.RemovePage("list_create")
		}
		a.showListsList()
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
				form.SetTitle(" [yellow]Create New List - Press 'i' to edit[white] ")
				form.SetFocus(currentFieldIndex)
				return nil
			} else {
				// Cancel and go back
				if a.pages.HasPage("list_create") {
					a.pages.RemovePage("list_create")
				}
				a.showListsList()
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
				form.SetTitle(" [yellow]Create New List - INSERT MODE (ESC to exit)[white] ")
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

	a.pages.AddPage("list_create", verticalFlex, true, true)
	a.pages.SwitchToPage("list_create")
	// Set initial focus to first field
	form.SetFocus(0)
	a.app.SetFocus(form)
}

