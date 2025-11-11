package tui

import (
	"fmt"

	"github.com/chr-hen/mtg-tui/internal/api"
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

	// Add buttons
	form.AddButton("Save", func() {
		// Update settings with the current checkbox value
		// The closure captures arenaCardsValue, so it will have the latest value
		a.settings.ShowArenaCards = arenaCardsValue
		
		// Save to file
		if err := SaveSettings(a.settings); err != nil {
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

