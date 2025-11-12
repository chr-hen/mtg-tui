package tui

import (
	"strings"

	"github.com/rivo/tview"
)

// FilterPanelConfig holds configuration for displaying a filter panel
type FilterPanelConfig struct {
	PageName       string        // Name of the filter page (e.g., "search_filter")
	Title          string        // Title for the form (e.g., "Filter Search Results")
	ReturnPageName string        // Name of the page to return to (e.g., "list")
	InitialQuery   string        // Initial query to populate the form with
	SetFilterQuery func(string)  // Function to set the filter query
	ReloadView     func()        // Function to reload the view with the filter
}

// showFilterPanel displays a generic filter panel that can be used for any view
func (a *App) showFilterPanel(config FilterPanelConfig) {
	// Remove old filter page to prevent artifacts
	a.pages.RemovePage(config.PageName)

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
		Title:        config.Title,
		InitialQuery: config.InitialQuery,
	})

	// Store filter callback for Enter key shortcut
	filterCallback := func() {
		query := buildQueryFromFields(fields)

		// Set filter query (empty string means no filter)
		config.SetFilterQuery(query)

		// Reset to first page and reload the view
		a.currentPage = 1

		// Close filter panel
		if a.pages.HasPage(config.PageName) {
			a.pages.RemovePage(config.PageName)
		}

		// Reload the view with new filter
		config.ReloadView()
	}

	// Store clear callback for Ctrl+D shortcut
	clearCallback := func() {
		clearFilterFields(fields)
	}

	// Store back callback for ESC key shortcut
	backCallback := func() {
		// Close filter panel and return to original view
		if a.pages.HasPage(config.PageName) {
			a.pages.RemovePage(config.PageName)
		}
		if a.pages.HasPage(config.ReturnPageName) {
			a.pages.SwitchToPage(config.ReturnPageName)
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

	a.pages.AddPage(config.PageName, horizontalFlex, true, true)
	a.pages.SwitchToPage(config.PageName)
	a.app.SetFocus(form)
}

// showSearchFilter displays a filter panel for the search results view
func (a *App) showSearchFilter() {
	// Combine currentQuery and searchFilterQuery to show the complete current filter state
	combinedQuery := a.currentQuery
	if a.searchFilterQuery != "" {
		if combinedQuery != "" {
			combinedQuery = combinedQuery + " " + a.searchFilterQuery
		} else {
			combinedQuery = a.searchFilterQuery
		}
	}

	a.showFilterPanel(FilterPanelConfig{
		PageName:       "search_filter",
		Title:          "Filter Search Results",
		ReturnPageName: "list",
		InitialQuery:   combinedQuery,
		SetFilterQuery: func(query string) {
			// When applying filter, we need to separate the base search from additional filters
			query = strings.TrimSpace(query)
			if query == "" {
				a.searchFilterQuery = ""
				return
			}

			// If currentQuery is empty, treat the entire query as a new search
			if a.currentQuery == "" {
				a.currentQuery = query
				a.searchFilterQuery = ""
				// Need to reload the search results with new query
				a.currentPage = 1
				a.showCardList()
				return
			}

			// Check if query starts with currentQuery (with space separator)
			// This handles the common case where user keeps the base search and adds filters
			if strings.HasPrefix(query, a.currentQuery+" ") {
				// Extract the additional filter part
				additional := strings.TrimPrefix(query, a.currentQuery+" ")
				a.searchFilterQuery = strings.TrimSpace(additional)
			} else if query == a.currentQuery {
				// Query matches currentQuery exactly, clear the filter
				a.searchFilterQuery = ""
			} else if strings.HasPrefix(query, a.currentQuery) {
				// Query starts with currentQuery but no space (edge case)
				additional := strings.TrimPrefix(query, a.currentQuery)
				a.searchFilterQuery = strings.TrimSpace(additional)
			} else {
				// Query doesn't start with currentQuery - user changed the base search
				// For now, treat as new search (this might need refinement)
				a.currentQuery = query
				a.searchFilterQuery = ""
				// Need to reload the search results with new query
				a.currentPage = 1
				a.showCardList()
				return
			}
		},
		ReloadView: a.showCardList,
	})
}

// showCollectionFilter displays a filter panel for the collection view
func (a *App) showCollectionFilter() {
	a.showFilterPanel(FilterPanelConfig{
		PageName:       "collection_filter",
		Title:          "Filter Collection",
		ReturnPageName: "collection",
		InitialQuery:   a.collectionFilterQuery,
		SetFilterQuery: func(query string) { a.collectionFilterQuery = query },
		ReloadView:     a.showCollection,
	})
}

// showWantsFilter displays a filter panel for the wants view
func (a *App) showWantsFilter() {
	a.showFilterPanel(FilterPanelConfig{
		PageName:       "wants_filter",
		Title:          "Filter Wants",
		ReturnPageName: "wants",
		InitialQuery:   a.wantsFilterQuery,
		SetFilterQuery: func(query string) { a.wantsFilterQuery = query },
		ReloadView:     a.showWants,
	})
}
