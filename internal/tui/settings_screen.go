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

	// Add button to configure search/filter fields
	form.AddButton("Configure search/filter fields", func() {
		a.showFieldVisibilityModal()
	})

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
			// If on Save button (index 5), trigger the save callback
			// If on Configure fields button (index 4), trigger that callback
			if currentFieldIndex == 5 {
				saveCallback()
				return nil
			}
			if currentFieldIndex == 4 {
				// Trigger the configure fields button
				a.showFieldVisibilityModal()
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
				// Form has 4 items (3 checkboxes + 1 input) and 3 buttons
				// Total indices: 0-6 (items 0-3, buttons 4-6)
				maxIndex := form.GetFormItemCount() + 2 // Last button index
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
			maxIndex := form.GetFormItemCount() + 2
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
				maxIndex := form.GetFormItemCount() + 2
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

// showFieldVisibilityModal displays a modal for configuring which fields appear in search/filter forms
func (a *App) showFieldVisibilityModal() {
	// Remove old modal if it exists
	if a.pages.HasPage("field_visibility_modal") {
		a.pages.RemovePage("field_visibility_modal")
	}

	// Create form for field visibility
	form := tview.NewForm()

	// Track checkbox states with local copies (only update settings on Save)
	fieldValues := map[string]bool{
		"Name":      a.settings.FieldVisibility.Name,
		"Type":      a.settings.FieldVisibility.Type,
		"Color":     a.settings.FieldVisibility.Color,
		"Oracle":    a.settings.FieldVisibility.Oracle,
		"Mana":      a.settings.FieldVisibility.Mana,
		"Power":     a.settings.FieldVisibility.Power,
		"Toughness": a.settings.FieldVisibility.Toughness,
		"Set":       a.settings.FieldVisibility.Set,
		"Rarity":    a.settings.FieldVisibility.Rarity,
		"Year":      a.settings.FieldVisibility.Year,
		"Artist":    a.settings.FieldVisibility.Artist,
		"Keyword":   a.settings.FieldVisibility.Keyword,
		"Is":        a.settings.FieldVisibility.Is,
	}

	// Create checkboxes for each field
	fieldOrder := []string{"Name", "Type", "Color", "Oracle", "Mana", "Power", "Toughness", "Set", "Rarity", "Year", "Artist", "Keyword", "Is"}
	for _, fieldName := range fieldOrder {
		fieldValue := fieldValues[fieldName]
		// Create a closure to capture the field name
		fieldNameCopy := fieldName
		form.AddCheckbox(fieldName, fieldValue, func(checked bool) {
			fieldValues[fieldNameCopy] = checked
		})
	}

	// Save callback
	saveCallback := func() {
		// Update settings with current checkbox values
		a.settings.FieldVisibility.Name = fieldValues["Name"]
		a.settings.FieldVisibility.Type = fieldValues["Type"]
		a.settings.FieldVisibility.Color = fieldValues["Color"]
		a.settings.FieldVisibility.Oracle = fieldValues["Oracle"]
		a.settings.FieldVisibility.Mana = fieldValues["Mana"]
		a.settings.FieldVisibility.Power = fieldValues["Power"]
		a.settings.FieldVisibility.Toughness = fieldValues["Toughness"]
		a.settings.FieldVisibility.Set = fieldValues["Set"]
		a.settings.FieldVisibility.Rarity = fieldValues["Rarity"]
		a.settings.FieldVisibility.Year = fieldValues["Year"]
		a.settings.FieldVisibility.Artist = fieldValues["Artist"]
		a.settings.FieldVisibility.Keyword = fieldValues["Keyword"]
		a.settings.FieldVisibility.Is = fieldValues["Is"]

		// Save to file
		if err := models.SaveSettings(a.settings); err != nil {
			// Show error modal
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error saving settings: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("field_visibility_error")
				})
			a.pages.AddPage("field_visibility_error", errorModal, true, true)
			return
		}

		// Close modal and return to settings
		if a.pages.HasPage("field_visibility_modal") {
			a.pages.RemovePage("field_visibility_modal")
		}
		if a.pages.HasPage("settings") {
			a.pages.SwitchToPage("settings")
		}
	}

	// Add buttons
	form.AddButton("Save", saveCallback)
	form.AddButton("Cancel", func() {
		// Close modal without saving
		if a.pages.HasPage("field_visibility_modal") {
			a.pages.RemovePage("field_visibility_modal")
		}
		if a.pages.HasPage("settings") {
			a.pages.SwitchToPage("settings")
		}
	})

	// Style the form
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Configure Search/Filter Fields[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Track current field index for navigation
	currentFieldIndex := 0

	// Handle input
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// ESC closes modal
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("field_visibility_modal") {
				a.pages.RemovePage("field_visibility_modal")
			}
			if a.pages.HasPage("settings") {
				a.pages.SwitchToPage("settings")
			}
			return nil
		}

		// Handle Enter key
		if event.Key() == tcell.KeyEnter {
			// If on Save button (last item), trigger save
			maxIndex := form.GetFormItemCount() + 1
			if currentFieldIndex == maxIndex-1 {
				saveCallback()
				return nil
			}
			// If on Cancel button (last item), trigger cancel
			if currentFieldIndex == maxIndex {
				if a.pages.HasPage("field_visibility_modal") {
					a.pages.RemovePage("field_visibility_modal")
				}
				if a.pages.HasPage("settings") {
					a.pages.SwitchToPage("settings")
				}
				return nil
			}
			// For checkboxes, let Enter work normally (toggles)
			return event
		}

		// Handle Space key for toggling checkboxes
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			// If on a checkbox, let space through to toggle
			maxIndex := form.GetFormItemCount()
			if currentFieldIndex < maxIndex {
				return event
			}
			// Block space for buttons
			return nil
		}

		// Handle vim motions
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'j':
				// Move down
				maxIndex := form.GetFormItemCount() + 1
				if currentFieldIndex < maxIndex {
					currentFieldIndex++
					form.SetFocus(currentFieldIndex)
				}
				return nil
			case 'k':
				// Move up
				if currentFieldIndex > 0 {
					currentFieldIndex--
					form.SetFocus(currentFieldIndex)
				}
				return nil
			default:
				// Block other text input
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

	// Center the form horizontally and vertically
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 60, 0, true). // Form (fixed width, horizontally centered)
		AddItem(nil, 0, 1, false)   // Right spacer

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).      // Top spacer
		AddItem(horizontalFlex, 0, 1, true). // Form (centered vertically)
		AddItem(nil, 0, 1, false)      // Bottom spacer

	a.pages.AddPage("field_visibility_modal", verticalFlex, true, true)
	a.pages.SwitchToPage("field_visibility_modal")
	a.app.SetFocus(form)
}

