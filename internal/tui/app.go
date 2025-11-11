package tui

import (
	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app          *tview.Application
	currentPage  int
	currentQuery string
	cards        []api.Card
	pagination   api.PaginationInfo
	list         *tview.List
	table        *tview.Table
	menu         *tview.List
	pages        *tview.Pages
	// Caching for faster page navigation
	pageCache        map[string]map[int][]api.Card // query -> page -> cards
	allMatchingCards map[string][]api.Card         // query -> all matching cards
	// Autocomplete data
	uniqueTypes    []string
	uniqueSets     []string
	uniqueRarities []string
	uniqueKeywords []string
	// Sorting
	sortField    string // "name", "released", "set", "rarity", "color", "cmc", "power", "toughness"
	sortAscending bool
}

func NewApp() *App {
	app := tview.NewApplication()
	// Enable mouse support and configure input properly
	app.EnableMouse(false)
	a := &App{
		app:              app,
		currentPage:      1,
		currentQuery:     "",
		pageCache:        make(map[string]map[int][]api.Card),
		allMatchingCards: make(map[string][]api.Card),
		uniqueTypes:      []string{},
		uniqueSets:       []string{},
		uniqueRarities:   []string{},
		uniqueKeywords:   []string{},
		sortField:        "name", // Default to name sort
		sortAscending:    true,
	}

	// Load autocomplete data in background
	go a.loadAutocompleteData()

	return a
}

func (a *App) Run() error {
	// Pages container to switch between menu, list, and detail popup
	a.pages = tview.NewPages()

	// Set input capture at pages level to prevent artifacts
	a.pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Only handle VIM keys when we're on the list page
		currentPage, _ := a.pages.GetFrontPage()
		if currentPage == "list" && a.table != nil {
			// Let the list handle it
			return event
		}
		return event
	})

	// Create and show main menu first
	a.showMainMenu()

	// Set root
	a.app.SetRoot(a.pages, true)

	// Check if cache exists, if not download bulk data in background
	if !api.CacheExists() {
		go a.downloadBulkData()
	}

	return a.app.Run()
}
