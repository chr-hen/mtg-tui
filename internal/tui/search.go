package tui

import (
	"github.com/rivo/tview"
)

func (a *App) showSearch() {
	// Remove old advanced page to prevent artifacts
	a.pages.RemovePage("search")

	// Autocomplete data should already be pre-loaded in background
	// If not ready yet, form will work without autocomplete and it will be available on next open
	// Check in background if data needs loading (non-blocking)
	go func() {
		if len(a.uniqueTypes) == 0 && len(a.uniqueSets) == 0 {
			// Data not loaded yet, load it in background (won't block UI)
			a.loadAutocompleteData()
		}
	}()

	// Create filter form using shared logic
	form, fields := a.createFilterForm(FilterFormConfig{
		Title: "Search",
	})

	// Store search callback for Enter key shortcut
	searchCallback := func() {
		query := buildQueryFromFields(fields)
		if query == "" {
			return
		}
		a.currentQuery = query
		a.currentPage = 1
		a.showCardList()
	}

	// Store clear callback for Ctrl+D shortcut
	clearCallback := func() {
		clearFilterFields(fields)
	}

	// Store back callback for ESC key shortcut
	backCallback := func() {
		// Force a full redraw by recreating the menu
		a.showMainMenu()
	}

	form.AddButton("Search (Enter)", searchCallback)
	form.AddButton("Clear (Ctrl+D)", clearCallback)
	form.AddButton("Back (Esc)", backCallback)

	// Set up input handling using shared function
	a.setupFilterFormInputHandling(form, fields, searchCallback, clearCallback, backCallback)

	// Center the form horizontally only (keep vertical alignment consistent)
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 90, 0, true). // Form (fixed width, wider for advanced, horizontally centered)
		AddItem(nil, 0, 1, false)   // Right spacer

	a.pages.AddPage("search", horizontalFlex, true, true)
	a.pages.SwitchToPage("search")
	a.app.SetFocus(form)
}
