package tui

import (
	"github.com/rivo/tview"
)

// showSearchFilter displays a filter panel for the search results view
func (a *App) showSearchFilter() {
	// Remove old filter page to prevent artifacts
	a.pages.RemovePage("search_filter")

	// Autocomplete data should already be pre-loaded in background
	// Check in background if data needs loading (non-blocking)
	go func() {
		if len(a.uniqueTypes) == 0 && len(a.uniqueSets) == 0 {
			// Data not loaded yet, load it in background (won't block UI)
			a.loadAutocompleteData()
		}
	}()

	// Create filter form using shared logic
	form, fields := a.createFilterForm(FilterFormConfig{
		Title: "Filter Search Results",
	})

	// Store filter callback for Enter key shortcut
	filterCallback := func() {
		query := buildQueryFromFields(fields)

		// Set filter query (empty string means no filter)
		a.searchFilterQuery = query

		// Reset to first page and reload the search view
		a.currentPage = 1

		// Close filter panel
		if a.pages.HasPage("search_filter") {
			a.pages.RemovePage("search_filter")
		}

		// Reload the card list view with new filter
		a.showCardList()
	}

	// Store clear callback for Ctrl+D shortcut
	clearCallback := func() {
		clearFilterFields(fields)
	}

	// Store back callback for ESC key shortcut
	backCallback := func() {
		// Close filter panel and return to card list view
		if a.pages.HasPage("search_filter") {
			a.pages.RemovePage("search_filter")
		}
		if a.pages.HasPage("list") {
			a.pages.SwitchToPage("list")
			if a.table != nil {
				a.app.SetFocus(a.table)
			}
		}
	}

	form.AddButton("Apply Filter (Enter)", filterCallback)
	form.AddButton("Clear (Ctrl+D)", clearCallback)
	form.AddButton("Back (Esc)", backCallback)

	// Set up input handling using shared function
	a.setupFilterFormInputHandling(form, fields, filterCallback, clearCallback, backCallback)

	// Center the form horizontally only (keep vertical alignment consistent)
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 90, 0, true). // Form (fixed width, horizontally centered)
		AddItem(nil, 0, 1, false)   // Right spacer

	a.pages.AddPage("search_filter", horizontalFlex, true, true)
	a.pages.SwitchToPage("search_filter")
	a.app.SetFocus(form)
}

// showCollectionFilter displays a filter panel for the collection view
func (a *App) showCollectionFilter() {
	// Remove old filter page to prevent artifacts
	a.pages.RemovePage("collection_filter")

	// Autocomplete data should already be pre-loaded in background
	// Check in background if data needs loading (non-blocking)
	go func() {
		if len(a.uniqueTypes) == 0 && len(a.uniqueSets) == 0 {
			// Data not loaded yet, load it in background (won't block UI)
			a.loadAutocompleteData()
		}
	}()

	// Create filter form using shared logic
	form, fields := a.createFilterForm(FilterFormConfig{
		Title: "Filter Collection",
	})

	// Store filter callback for Enter key shortcut
	filterCallback := func() {
		query := buildQueryFromFields(fields)

		// Set filter query (empty string means no filter)
		a.collectionFilterQuery = query

		// Reset to first page and reload the collection view
		a.currentPage = 1

		// Close filter panel
		if a.pages.HasPage("collection_filter") {
			a.pages.RemovePage("collection_filter")
		}

		// Reload the collection view with new filter
		a.showCollection()
	}

	// Store clear callback for Ctrl+D shortcut
	clearCallback := func() {
		clearFilterFields(fields)
	}

	// Store back callback for ESC key shortcut
	backCallback := func() {
		// Close filter panel and return to collection view
		if a.pages.HasPage("collection_filter") {
			a.pages.RemovePage("collection_filter")
		}
		if a.pages.HasPage("collection") {
			a.pages.SwitchToPage("collection")
			if a.table != nil {
				a.app.SetFocus(a.table)
			}
		}
	}

	form.AddButton("Apply Filter (Enter)", filterCallback)
	form.AddButton("Clear (Ctrl+D)", clearCallback)
	form.AddButton("Back (Esc)", backCallback)

	// Set up input handling using shared function
	a.setupFilterFormInputHandling(form, fields, filterCallback, clearCallback, backCallback)

	// Center the form horizontally only (keep vertical alignment consistent)
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 90, 0, true). // Form (fixed width, horizontally centered)
		AddItem(nil, 0, 1, false)   // Right spacer

	a.pages.AddPage("collection_filter", horizontalFlex, true, true)
	a.pages.SwitchToPage("collection_filter")
	a.app.SetFocus(form)
}
