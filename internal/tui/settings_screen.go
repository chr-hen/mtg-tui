package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showSettingsScreen displays the settings screen
func (a *App) showSettingsScreen() {
	// Remove old settings page if it exists
	if a.pages.HasPage("settings") {
		a.pages.RemovePage("settings")
	}

	// Create form for settings
	form := tview.NewForm()

	// Track checkbox state using a pointer so the callback can update it
	arenaCardsValue := a.settings.ShowArenaCards
	form.AddCheckbox("Show Arena Cards", arenaCardsValue, func(checked bool) {
		arenaCardsValue = checked
	})
	
	// Track checkbox state for type "card" cards
	typeCardValue := a.settings.ShowTypeCard
	form.AddCheckbox("Include cards with type \"card\"", typeCardValue, func(checked bool) {
		typeCardValue = checked
	})
	
	// Track checkbox state for art cards
	artCardsValue := a.settings.ShowArtCards
	form.AddCheckbox("Show art cards", artCardsValue, func(checked bool) {
		artCardsValue = checked
	})
	
	// Track page size input
	pageSizeValue := fmt.Sprintf("%d", a.settings.PageSize)
	pageSizeInput := tview.NewInputField().
		SetLabel("Cards per page: ").
		SetText(pageSizeValue).
		SetFieldWidth(5).
		SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
			// Only allow digits
			if lastChar < '0' || lastChar > '9' {
				return false
			}
			// Limit to reasonable values (1-100)
			if len(textToCheck) > 3 {
				return false
			}
			return true
		})
	form.AddFormItem(pageSizeInput)

	// Save callback function
	saveCallback := func() {
		// Update settings with the current checkbox values
		// The closures capture the values, so they will have the latest values
		a.settings.ShowArenaCards = arenaCardsValue
		a.settings.ShowTypeCard = typeCardValue
		a.settings.ShowArtCards = artCardsValue
		
		// Parse and update page size
		pageSizeText := pageSizeInput.GetText()
		var pageSize int
		if n, err := fmt.Sscanf(pageSizeText, "%d", &pageSize); n == 1 && err == nil && pageSize > 0 {
			a.settings.PageSize = pageSize
		} else {
			// Invalid input, show error and don't save
			errorModal := tview.NewModal().
				SetText("Invalid page size. Please enter a number between 1 and 100.").
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("settings_error")
				})
			a.pages.AddPage("settings_error", errorModal, true, true)
			return
		}
		
		// Save to file
		if err := models.SaveSettings(a.settings); err != nil {
			// Show error modal
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error saving settings: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("settings_error")
				})
			a.pages.AddPage("settings_error", errorModal, true, true)
			return
		}
		
		// Clear caches since filtering may have changed
		a.allMatchingCards = make(map[string][]api.Card)
		a.pageCache = make(map[string]map[int][]api.Card)
		
		// Go back to menu
		if a.pages.HasPage("settings") {
			a.pages.RemovePage("settings")
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

	// Add buttons
	form.AddButton("Save", saveCallback)

	form.AddButton("Cancel", func() {
		// Go back to menu without saving
		if a.pages.HasPage("settings") {
			a.pages.RemovePage("settings")
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

	// Track current field index and insert mode for vim motions
	currentFieldIndex := 0
	insertMode := false

	// Helper function to update title based on insert mode
	updateTitle := func() {
		if insertMode {
			form.SetTitle(" [yellow]Settings - INSERT MODE (ESC to exit)[white] ")
		} else {
			form.SetTitle(" [yellow]Settings[white] ")
		}
	}

	// Style the form
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Settings[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Handle vim motions and ESC key
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// In insert mode, allow normal input except ESC to exit
		if insertMode {
			if event.Key() == tcell.KeyEscape {
				// ESC exits insert mode
				insertMode = false
				updateTitle()
				// Return focus to form
				a.app.SetFocus(form)
				return nil
			}
			// Allow all other input in insert mode
			return event
		}

		// Normal mode - handle navigation and mode switching
		if event.Key() == tcell.KeyEscape {
			// Go back to menu
			if a.pages.HasPage("settings") {
				a.pages.RemovePage("settings")
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

		// Handle Enter key
		if event.Key() == tcell.KeyEnter {
			// If on Save button (index 4), trigger the save callback
			if currentFieldIndex == 4 {
				saveCallback()
				return nil
			}
			// For other items, let Enter work normally (toggles checkboxes, activates buttons)
			return event
		}

		// Handle Space key for toggling checkboxes
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			// If on a checkbox (indices 0-2), let space through to toggle
			if currentFieldIndex >= 0 && currentFieldIndex <= 2 {
				return event
			}
			// Block space for other items
			return nil
		}

		// Handle vim motions
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'i':
				// Enter insert mode on text input field (index 3)
				if currentFieldIndex == 3 {
					insertMode = true
					updateTitle()
					// Focus the input field for immediate editing
					a.app.SetFocus(pageSizeInput)
					return nil
				}
				// Block i on other fields
				return nil
			case 'j':
				// Move down to next field/button
				// Form has 4 items (3 checkboxes + 1 input) and 2 buttons
				// Total indices: 0-5 (items 0-3, buttons 4-5)
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
				// Block other text input in normal mode
				return nil
			}
		}

		// Handle Tab/Shift+Tab for navigation
		if event.Key() == tcell.KeyTab {
			maxIndex := form.GetFormItemCount() + 1
			if currentFieldIndex < maxIndex {
				currentFieldIndex++
				form.SetFocus(currentFieldIndex)
			} else {
				// Wrap to first field
				currentFieldIndex = 0
				form.SetFocus(currentFieldIndex)
			}
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			if currentFieldIndex > 0 {
				currentFieldIndex--
				form.SetFocus(currentFieldIndex)
			} else {
				// Wrap to last field/button
				maxIndex := form.GetFormItemCount() + 1
				currentFieldIndex = maxIndex
				form.SetFocus(currentFieldIndex)
			}
			return nil
		}

		return event
	})

	// Center the form horizontally
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 70, 0, true).  // Form (fixed width, horizontally centered)
		AddItem(nil, 0, 1, false)    // Right spacer

	a.pages.AddPage("settings", horizontalFlex, true, true)
	a.pages.SwitchToPage("settings")
	a.app.SetFocus(form)
}

