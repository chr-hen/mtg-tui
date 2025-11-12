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

	// Add buttons
	form.AddButton("Save", func() {
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
	})

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

	// Style the form
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Settings[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Handle ESC key to go back
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

