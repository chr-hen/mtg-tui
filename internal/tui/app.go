package tui

import (
	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/chr-hen/mtg-tui/internal/tui/models"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app          *tview.Application
	currentPage  int
	currentQuery string
	cards        []api.Card
	cardGroups   []util.CardGroup // Grouped cards for display
	pagination   api.PaginationInfo
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
	sortField     string // "name", "released", "set", "rarity", "color", "cmc", "power", "toughness"
	sortAscending bool
	// Settings
	settings *models.Settings
	// Collections
	collection            *collections.Collection
	wants                 *collections.Wants
	collectionFilterQuery string // Filter query for collection view
	wantsFilterQuery      string // Filter query for wants view
	searchFilterQuery     string // Filter query for search results view
}

func NewApp() *App {
	app := tview.NewApplication()
	// Enable mouse support and configure input properly
	app.EnableMouse(false)

	// Load settings
	settings, err := models.LoadSettings()
	if err != nil {
		// If loading fails, use default settings
		settings = &models.Settings{
			ShowArenaCards: false,
			ShowTypeCard:   false,
			ShowArtCards:   false,
			PageSize:       10,
		}
	}
	// Ensure PageSize has a valid default if not set (for old config files)
	if settings.PageSize <= 0 {
		settings.PageSize = 10
	}

	// Load collection
	collection, err := collections.LoadCollection()
	if err != nil {
		// If loading fails, create empty collection
		collection = &collections.Collection{
			OwnedPrintings: []collections.CardPrintingID{},
		}
	}

	// Load wants
	wants, err := collections.LoadWants()
	if err != nil {
		// If loading fails, create empty wants
		wants = &collections.Wants{
			WantedPrintings: []collections.CardPrintingID{},
		}
	}

	a := &App{
		app:                   app,
		currentPage:           1,
		currentQuery:          "",
		pageCache:             make(map[string]map[int][]api.Card),
		allMatchingCards:      make(map[string][]api.Card),
		uniqueTypes:           []string{},
		uniqueSets:            []string{},
		uniqueRarities:        []string{},
		uniqueKeywords:        []string{},
		sortField:             "name", // Default to name sort
		sortAscending:         true,
		settings:              settings,
		collection:            collection,
		wants:                 wants,
		collectionFilterQuery: "", // No filter by default
		wantsFilterQuery:      "", // No filter by default
		searchFilterQuery:     "", // No filter by default
	}

	// Load autocomplete data in background
	go a.loadAutocompleteData()

	// Pre-load collection data in background
	go a.preloadCollectionCards()

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
