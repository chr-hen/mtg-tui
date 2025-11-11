package tui

import (
	"fmt"
	"log"
	"sort"
	"strings"

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
	}

	// Load autocomplete data in background
	go a.loadAutocompleteData()

	return a
}

// loadAutocompleteData extracts unique values from card data for autocomplete
func (a *App) loadAutocompleteData() {
	cards, err := api.LoadCardsFromCache()
	if err != nil {
		// Cache not available yet, will be loaded later
		return
	}

	typeMap := make(map[string]bool)
	setMap := make(map[string]bool)
	rarityMap := make(map[string]bool)
	keywordMap := make(map[string]bool)

	for _, card := range cards {
		// Extract types (split TypeLine by spaces and dashes)
		typeParts := strings.FieldsFunc(card.TypeLine, func(r rune) bool {
			return r == ' ' || r == '-' || r == '—'
		})
		for _, part := range typeParts {
			part = strings.TrimSpace(part)
			if part != "" && len(part) > 1 {
				typeMap[strings.ToLower(part)] = true
			}
		}

		// Extract sets (use both SetCode and Set name)
		if card.SetCode != "" {
			setMap[strings.ToLower(card.SetCode)] = true
		}
		if card.Set != "" {
			setMap[strings.ToLower(card.Set)] = true
		}

		// Extract rarities
		if card.Rarity != "" {
			rarityMap[strings.ToLower(card.Rarity)] = true
		}

		// Extract keywords
		for _, kw := range card.Keywords {
			if kw != "" {
				keywordMap[strings.ToLower(kw)] = true
			}
		}
	}

	// Convert maps to sorted slices
	a.uniqueTypes = make([]string, 0, len(typeMap))
	for t := range typeMap {
		a.uniqueTypes = append(a.uniqueTypes, t)
	}
	sort.Strings(a.uniqueTypes)

	a.uniqueSets = make([]string, 0, len(setMap))
	for s := range setMap {
		a.uniqueSets = append(a.uniqueSets, s)
	}
	sort.Strings(a.uniqueSets)

	a.uniqueRarities = make([]string, 0, len(rarityMap))
	for r := range rarityMap {
		a.uniqueRarities = append(a.uniqueRarities, r)
	}
	sort.Strings(a.uniqueRarities)

	a.uniqueKeywords = make([]string, 0, len(keywordMap))
	for k := range keywordMap {
		a.uniqueKeywords = append(a.uniqueKeywords, k)
	}
	sort.Strings(a.uniqueKeywords)
}

// createAutocompleteFunc creates an autocomplete function for a given list of options
func (a *App) createAutocompleteFunc(options []string) func(currentText string) []string {
	return func(currentText string) []string {
		currentText = strings.TrimSpace(currentText)
		if currentText == "" {
			return []string{}
		}

		// Remove common query prefixes for matching
		searchText := strings.ToLower(currentText)
		prefixes := []string{"t:", "type:", "s:", "set:", "r:", "rarity:", "kw:", "keyword:"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(searchText, prefix) {
				searchText = strings.TrimPrefix(searchText, prefix)
				break
			}
		}

		if searchText == "" {
			return []string{}
		}

		// Find all matching options (limit to 10 for performance)
		var matches []string
		for _, option := range options {
			if strings.HasPrefix(strings.ToLower(option), searchText) {
				matches = append(matches, option)
				if len(matches) >= 10 {
					break
				}
			}
		}

		return matches
	}
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

func (a *App) downloadBulkData() {
	// Show download modal - make it visible immediately
	modal := tview.NewModal().
		SetText("Downloading card database from Scryfall...\nThis may take a few minutes.\n\nPlease wait...").
		AddButtons([]string{})

	a.app.QueueUpdate(func() {
		a.pages.AddPage("download", modal, true, true)
		a.pages.SwitchToPage("download")
	})

	// Fetch bulk data metadata
	bulkDataList, err := api.GetBulkDataMetadata()
	if err != nil {
		a.app.QueueUpdate(func() {
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					a.app.Stop()
				})
			a.pages.AddPage("error", errorModal, true, true)
		})
		return
	}

	// Find default_cards
	defaultCards, err := api.GetDefaultCardsBulkData(bulkDataList)
	if err != nil {
		a.app.QueueUpdate(func() {
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					a.app.Stop()
				})
			a.pages.AddPage("error", errorModal, true, true)
		})
		return
	}

	// Update modal with file size info and download location
	sizeMB := float64(defaultCards.Size) / (1024 * 1024)
	cachePath := api.GetCacheFilePath()
	a.app.QueueUpdate(func() {
		modal.SetText(fmt.Sprintf("Downloading card database...\n\nFile size: %.1f MB\nSaving to: %s\n\nPlease wait...", sizeMB, cachePath))
	})

	// Download with progress updates
	err = api.DownloadBulkData(defaultCards.DownloadURI, func(bytesRead, totalBytes int64) {
		downloadedMB := float64(bytesRead) / (1024 * 1024)
		var progressText string

		if totalBytes > 0 {
			percent := float64(bytesRead) / float64(totalBytes) * 100
			totalMB := float64(totalBytes) / (1024 * 1024)
			progressText = fmt.Sprintf("Downloading card database...\n\nProgress: %.1f%%\n%.1f MB / %.1f MB", percent, downloadedMB, totalMB)
		} else {
			// Content-Length not available, just show bytes downloaded
			progressText = fmt.Sprintf("Downloading card database...\n\nDownloaded: %.1f MB\n\nPlease wait...", downloadedMB)
		}

		a.app.QueueUpdate(func() {
			modal.SetText(progressText)
		})
	})

	if err != nil {
		a.app.QueueUpdate(func() {
			errorModal := tview.NewModal().
				SetText(fmt.Sprintf("Error downloading: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					a.app.Stop()
				})
			a.pages.AddPage("error", errorModal, true, true)
		})
		return
	}

	// Remove download modal
	a.app.QueueUpdate(func() {
		a.pages.RemovePage("download")
	})
}

func (a *App) showMainMenu() {
	// Remove old menu page to prevent artifacts (only if it exists)
	if a.pages.HasPage("menu") {
		a.pages.RemovePage("menu")
	}

	// Create ASCII art for MTG-TUI with color
	asciiArt := `[yellow]███╗   ███╗[white]████████╗ ██████╗     [yellow]███████╗[white]██╗   ██╗██╗
[yellow]████╗ ████║[white]╚══██╔══╝██╔════╝     [yellow]╚══██╔══╝[white]██║   ██║██║
[yellow]██╔████╔██║[white]   ██║   ██║            [yellow]   ██║   [white]██║   ██║██║
[yellow]██║╚██╔╝██║[white]   ██║   ██║            [yellow]   ██║   [white]██║   ██║██║
[yellow]██║ ╚═╝ ██║[white]   ██║   ╚██████╗       [yellow]   ██║   [white]╚██████╔╝██║
[yellow]╚═╝     ╚═╝[white]   ╚═╝    ╚═════╝       [yellow]   ╚═╝   [white]╚═════╝ ╚═╝`

	// Create text view for ASCII art
	artView := tview.NewTextView().
		SetText(asciiArt).
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Create subtitle with styling
	subtitle := tview.NewTextView().
		SetText("[gray]Magic: The Gathering Card Browser[white]").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Create menu list with better styling and separators
	menu := tview.NewList()
	a.menu = menu // Store reference for later use

	// Store callbacks for later use
	callbacks := []func(){
		func() { a.showSearchInput() },
		nil, // separator
		func() { a.showAdvancedSearch() },
		nil, // separator
		func() { a.app.Stop() },
	}

	menu.AddItem("Search for Card", "Enter a card name or search query", 's', callbacks[0]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("Advanced Search", "Use Scryfall syntax for detailed filtering", 'a', callbacks[2]).
		AddItem("[gray]─────────────────────────────[white]", "", 0, nil).
		AddItem("Quit", "Exit the application", 'q', callbacks[4])

	menu.SetBorder(true).
		SetTitle(" [yellow]Navigation[white] ").
		SetBorderColor(tcell.ColorYellow).
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	menu.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Override selected func to skip separators
	menu.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Skip separator items - move to next item
		if strings.Contains(mainText, "─") {
			// Try to move to next item
			if index < menu.GetItemCount()-1 {
				menu.SetCurrentItem(index + 1)
			} else if index > 0 {
				menu.SetCurrentItem(index - 1)
			}
			return
		}
		// Call the original callback if it exists
		if index < len(callbacks) && callbacks[index] != nil {
			callbacks[index]()
		} else {
			// Handle direct index mapping for menu items
			if index == 0 {
				a.showSearchInput()
			} else if index == 2 {
				a.showAdvancedSearch()
			} else if index == 4 {
				a.app.Stop()
			}
		}
	})

	menu.SetDoneFunc(func() {
		a.app.Stop()
	})

	// Handle navigation to skip separators
	menu.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// If we land on a separator, move to the next item
		if strings.Contains(mainText, "─") {
			if index < menu.GetItemCount()-1 {
				menu.SetCurrentItem(index + 1)
			} else if index > 0 {
				menu.SetCurrentItem(index - 1)
			}
		}
	})

	// Add VIM keybindings for menu navigation
	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'j':
				// Move down
				current := menu.GetCurrentItem()
				next := current + 1
				// Skip separators
				for next < menu.GetItemCount() {
					mainText, _ := menu.GetItemText(next)
					if !strings.Contains(mainText, "─") {
						menu.SetCurrentItem(next)
						return nil
					}
					next++
				}
				return nil
			case 'k':
				// Move up
				current := menu.GetCurrentItem()
				prev := current - 1
				// Skip separators
				for prev >= 0 {
					mainText, _ := menu.GetItemText(prev)
					if !strings.Contains(mainText, "─") {
						menu.SetCurrentItem(prev)
						return nil
					}
					prev--
				}
				return nil
			}
		}
		return event
	})

	// Create a flex container to center everything vertically
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false). // Top spacer
		AddItem(artView, 6, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(nil, 1, 0, false).  // Spacer between title and menu
		AddItem(menu, 12, 0, true). // Menu with fixed height to show all items (5 items + separators)
		AddItem(nil, 0, 1, false)   // Bottom spacer

	// Create another flex to center horizontally
	centeredFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(flex, 70, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("menu", centeredFlex, true, true)
	a.pages.SwitchToPage("menu")

	// Only queue updates if the app is already running
	// Otherwise just set focus directly (for initial menu creation)
	if a.pages.HasFocus() {
		a.app.QueueUpdateDraw(func() {
			a.app.SetFocus(menu)
		})
	} else {
		a.app.SetFocus(menu)
	}
}

func (a *App) showSearchInput() {
	// Remove old search page to prevent artifacts
	a.pages.RemovePage("search")

	form := tview.NewForm()
	form.SetTitle(" [yellow]Search for Card[white] ")
	form.SetBorder(true)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetTitleColor(tcell.ColorYellow)
	form.SetBackgroundColor(tcell.ColorBlack)
	form.SetButtonTextColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorYellow)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	searchInput := tview.NewInputField()
	searchInput.SetLabel("Search: ")
	searchInput.SetFieldWidth(50)
	searchInput.SetPlaceholder("Enter card name or search query (e.g., 'Lightning Bolt' or 'type:creature')")
	searchInput.SetPlaceholderTextColor(tcell.ColorGray)
	// Use SetFormAttributes to set all colors - this should control all states
	// Parameters: labelWidth, labelColor, bgColor, fieldTextColor, fieldBgColor
	searchInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	form.AddFormItem(searchInput)

	// Force form to apply field colors after adding item
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	form.AddButton("Search", func() {
		query := strings.TrimSpace(searchInput.GetText())
		if query == "" {
			// Show warning modal instead of crashing
			modal := tview.NewModal().
				SetText("Please enter a search query.\n\nA card name or search query is required.").
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("warning")
					// Return focus to the search form
					a.app.SetFocus(form)
				})
			a.pages.AddPage("warning", modal, true, true)
			return
		}
		a.currentQuery = query
		a.currentPage = 1
		a.showCardList()
	})

	form.AddButton("Back", func() {
		// Force a full redraw by recreating the menu
		a.showMainMenu()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// Force a full redraw by recreating the menu
			a.showMainMenu()
			return nil
		}
		return event
	})

	a.pages.AddPage("search", form, true, true)
	a.pages.SwitchToPage("search")
	a.app.SetFocus(form)
}

func (a *App) showAdvancedSearch() {
	// Remove old advanced page to prevent artifacts
	a.pages.RemovePage("advanced")

	// Ensure autocomplete data is loaded (in case cache wasn't ready at startup)
	if len(a.uniqueTypes) == 0 && len(a.uniqueSets) == 0 {
		// Try to load synchronously if not already loaded
		a.loadAutocompleteData()
	}

	form := tview.NewForm()
	form.SetTitle(" [yellow]Advanced Search (Scryfall Syntax)[white] ")
	form.SetBorder(true)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetTitleColor(tcell.ColorYellow)
	form.SetBackgroundColor(tcell.ColorBlack)
	form.SetButtonTextColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorYellow)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	// Name
	nameInput := tview.NewInputField()
	nameInput.SetLabel("Name: ")
	nameInput.SetFieldWidth(40)
	nameInput.SetPlaceholder("Card name (e.g., Lightning Bolt)")
	nameInput.SetPlaceholderTextColor(tcell.ColorGray)
	// Use SetFormAttributes to set all colors consistently
	nameInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Type
	typeInput := tview.NewInputField()
	typeInput.SetLabel("Type: ")
	typeInput.SetFieldWidth(40)
	typeInput.SetPlaceholder("t:creature (autocomplete available)")
	typeInput.SetPlaceholderTextColor(tcell.ColorGray)
	typeInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for types
	if len(a.uniqueTypes) > 0 {
		typeInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueTypes))
	}

	// Color
	colorInput := tview.NewInputField()
	colorInput.SetLabel("Color: ")
	colorInput.SetFieldWidth(40)
	colorInput.SetPlaceholder("c:r or c:uw (w/u/b/r/g)")
	colorInput.SetPlaceholderTextColor(tcell.ColorGray)
	colorInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Oracle Text
	oracleInput := tview.NewInputField()
	oracleInput.SetLabel("Oracle Text: ")
	oracleInput.SetFieldWidth(40)
	oracleInput.SetPlaceholder("o:\"draw a card\"")
	oracleInput.SetPlaceholderTextColor(tcell.ColorGray)
	oracleInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Mana Cost
	manaInput := tview.NewInputField()
	manaInput.SetLabel("Mana Cost: ")
	manaInput.SetFieldWidth(40)
	manaInput.SetPlaceholder("m:{G}{U} or mv<=3")
	manaInput.SetPlaceholderTextColor(tcell.ColorGray)
	manaInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Power/Toughness
	powerInput := tview.NewInputField()
	powerInput.SetLabel("Power: ")
	powerInput.SetFieldWidth(40)
	powerInput.SetPlaceholder("pow>=4 or pow>tou")
	powerInput.SetPlaceholderTextColor(tcell.ColorGray)
	powerInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	toughnessInput := tview.NewInputField()
	toughnessInput.SetLabel("Toughness: ")
	toughnessInput.SetFieldWidth(40)
	toughnessInput.SetPlaceholder("tou>=4")
	toughnessInput.SetPlaceholderTextColor(tcell.ColorGray)
	toughnessInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Set
	setInput := tview.NewInputField()
	setInput.SetLabel("Set: ")
	setInput.SetFieldWidth(40)
	setInput.SetPlaceholder("s:khm (autocomplete available)")
	setInput.SetPlaceholderTextColor(tcell.ColorGray)
	setInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for sets
	if len(a.uniqueSets) > 0 {
		setInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueSets))
	}

	// Rarity
	rarityInput := tview.NewInputField()
	rarityInput.SetLabel("Rarity: ")
	rarityInput.SetFieldWidth(40)
	rarityInput.SetPlaceholder("r:rare (autocomplete available)")
	rarityInput.SetPlaceholderTextColor(tcell.ColorGray)
	rarityInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for rarities
	if len(a.uniqueRarities) > 0 {
		rarityInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueRarities))
	}

	// Year
	yearInput := tview.NewInputField()
	yearInput.SetLabel("Year: ")
	yearInput.SetFieldWidth(40)
	yearInput.SetPlaceholder("year>=2020")
	yearInput.SetPlaceholderTextColor(tcell.ColorGray)
	yearInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Artist
	artistInput := tview.NewInputField()
	artistInput.SetLabel("Artist: ")
	artistInput.SetFieldWidth(40)
	artistInput.SetPlaceholder("a:avon")
	artistInput.SetPlaceholderTextColor(tcell.ColorGray)
	artistInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Keyword
	keywordInput := tview.NewInputField()
	keywordInput.SetLabel("Keyword: ")
	keywordInput.SetFieldWidth(40)
	keywordInput.SetPlaceholder("kw:flying (autocomplete available)")
	keywordInput.SetPlaceholderTextColor(tcell.ColorGray)
	keywordInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for keywords
	if len(a.uniqueKeywords) > 0 {
		keywordInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueKeywords))
	}

	// Special flags
	isInput := tview.NewInputField()
	isInput.SetLabel("Is: ")
	isInput.SetFieldWidth(40)
	isInput.SetPlaceholder("is:multicolor, is:spell, is:permanent")
	isInput.SetPlaceholderTextColor(tcell.ColorGray)
	isInput.SetFormAttributes(0, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	form.AddFormItem(nameInput)
	form.AddFormItem(typeInput)
	form.AddFormItem(colorInput)
	form.AddFormItem(oracleInput)
	form.AddFormItem(manaInput)
	form.AddFormItem(powerInput)
	form.AddFormItem(toughnessInput)
	form.AddFormItem(setInput)
	form.AddFormItem(rarityInput)
	form.AddFormItem(yearInput)
	form.AddFormItem(artistInput)
	form.AddFormItem(keywordInput)
	form.AddFormItem(isInput)

	// Force form to apply field colors after adding all items
	// This ensures colors persist even when fields are focused
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	form.AddButton("Search", func() {
		// Build query from all fields
		var queryParts []string

		if name := strings.TrimSpace(nameInput.GetText()); name != "" {
			queryParts = append(queryParts, name)
		}
		if typ := strings.TrimSpace(typeInput.GetText()); typ != "" {
			if !strings.HasPrefix(typ, "t:") && !strings.HasPrefix(typ, "type:") {
				queryParts = append(queryParts, "t:"+typ)
			} else {
				queryParts = append(queryParts, typ)
			}
		}
		if color := strings.TrimSpace(colorInput.GetText()); color != "" {
			if !strings.HasPrefix(color, "c:") && !strings.HasPrefix(color, "color:") {
				queryParts = append(queryParts, "c:"+color)
			} else {
				queryParts = append(queryParts, color)
			}
		}
		if oracle := strings.TrimSpace(oracleInput.GetText()); oracle != "" {
			if !strings.HasPrefix(oracle, "o:") && !strings.HasPrefix(oracle, "oracle:") {
				queryParts = append(queryParts, "o:"+oracle)
			} else {
				queryParts = append(queryParts, oracle)
			}
		}
		if mana := strings.TrimSpace(manaInput.GetText()); mana != "" {
			if !strings.HasPrefix(mana, "m:") && !strings.HasPrefix(mana, "mana:") && !strings.HasPrefix(mana, "mv") && !strings.HasPrefix(mana, "cmc") {
				queryParts = append(queryParts, "m:"+mana)
			} else {
				queryParts = append(queryParts, mana)
			}
		}
		if power := strings.TrimSpace(powerInput.GetText()); power != "" {
			if !strings.HasPrefix(power, "pow:") && !strings.HasPrefix(power, "power:") {
				queryParts = append(queryParts, "pow:"+power)
			} else {
				queryParts = append(queryParts, power)
			}
		}
		if toughness := strings.TrimSpace(toughnessInput.GetText()); toughness != "" {
			if !strings.HasPrefix(toughness, "tou:") && !strings.HasPrefix(toughness, "toughness:") {
				queryParts = append(queryParts, "tou:"+toughness)
			} else {
				queryParts = append(queryParts, toughness)
			}
		}
		if set := strings.TrimSpace(setInput.GetText()); set != "" {
			if !strings.HasPrefix(set, "s:") && !strings.HasPrefix(set, "set:") && !strings.HasPrefix(set, "e:") {
				queryParts = append(queryParts, "s:"+set)
			} else {
				queryParts = append(queryParts, set)
			}
		}
		if rarity := strings.TrimSpace(rarityInput.GetText()); rarity != "" {
			if !strings.HasPrefix(rarity, "r:") && !strings.HasPrefix(rarity, "rarity:") {
				queryParts = append(queryParts, "r:"+rarity)
			} else {
				queryParts = append(queryParts, rarity)
			}
		}
		if year := strings.TrimSpace(yearInput.GetText()); year != "" {
			if !strings.HasPrefix(year, "year:") {
				queryParts = append(queryParts, "year:"+year)
			} else {
				queryParts = append(queryParts, year)
			}
		}
		if artist := strings.TrimSpace(artistInput.GetText()); artist != "" {
			if !strings.HasPrefix(artist, "a:") && !strings.HasPrefix(artist, "artist:") {
				queryParts = append(queryParts, "a:"+artist)
			} else {
				queryParts = append(queryParts, artist)
			}
		}
		if keyword := strings.TrimSpace(keywordInput.GetText()); keyword != "" {
			if !strings.HasPrefix(keyword, "kw:") && !strings.HasPrefix(keyword, "keyword:") {
				queryParts = append(queryParts, "kw:"+keyword)
			} else {
				queryParts = append(queryParts, keyword)
			}
		}
		if is := strings.TrimSpace(isInput.GetText()); is != "" {
			if !strings.HasPrefix(is, "is:") {
				queryParts = append(queryParts, "is:"+is)
			} else {
				queryParts = append(queryParts, is)
			}
		}

		if len(queryParts) == 0 {
			return
		}

		query := strings.Join(queryParts, " ")
		a.currentQuery = query
		a.currentPage = 1
		a.showCardList()
	})

	form.AddButton("Clear", func() {
		nameInput.SetText("")
		typeInput.SetText("")
		colorInput.SetText("")
		oracleInput.SetText("")
		manaInput.SetText("")
		powerInput.SetText("")
		toughnessInput.SetText("")
		setInput.SetText("")
		rarityInput.SetText("")
		yearInput.SetText("")
		artistInput.SetText("")
		keywordInput.SetText("")
		isInput.SetText("")
	})

	form.AddButton("Back", func() {
		// Force a full redraw by recreating the menu
		a.showMainMenu()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// Force a full redraw by recreating the menu
			a.showMainMenu()
			return nil
		}
		return event
	})

	a.pages.AddPage("advanced", form, true, true)
	a.pages.SwitchToPage("advanced")
	a.app.SetFocus(form)
}

func (a *App) showCardList() {
	// Check if we have cached matching cards for this query
	allCards, hasCache := a.allMatchingCards[a.currentQuery]
	if !hasCache {
		// Load all matching cards and cache them
		var err error
		allCards, err = api.GetAllMatchingCards(a.currentQuery)
		if err != nil {
			// Show error modal
			modal := tview.NewModal().
				SetText(fmt.Sprintf("Error fetching cards: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
					// Force a full redraw by recreating the menu
					a.showMainMenu()
				})
			a.pages.AddPage("error", modal, true, true)
			return
		}
		// Cache all matching cards
		a.allMatchingCards[a.currentQuery] = allCards
		// Initialize page cache for this query
		a.pageCache[a.currentQuery] = make(map[int][]api.Card)
	}

	// Get current page from cache or calculate it
	pageSize := 10
	totalCards := len(allCards)
	startIdx := (a.currentPage - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalCards {
		// Empty page
		a.cards = []api.Card{}
		a.pagination = api.PaginationInfo{
			HasMore:    false,
			TotalCards: totalCards,
		}
	} else {
		if endIdx > totalCards {
			endIdx = totalCards
		}

		// Check cache first
		if cachedPage, exists := a.pageCache[a.currentQuery][a.currentPage]; exists {
			a.cards = cachedPage
		} else {
			// Calculate and cache this page
			a.cards = allCards[startIdx:endIdx]
			a.pageCache[a.currentQuery][a.currentPage] = a.cards
		}

		hasMore := endIdx < totalCards
		a.pagination = api.PaginationInfo{
			HasMore:    hasMore,
			TotalCards: totalCards,
		}
	}

	// Pre-load next 2 pages in background
	go a.preloadPages(a.currentQuery, a.currentPage, allCards, pageSize)

	// Create table view for multi-line card display
	a.table = tview.NewTable()
	a.updateListTitle()
	a.table.SetBorder(true)
	a.table.SetBorderColor(tcell.ColorYellow)
	a.table.SetTitleColor(tcell.ColorYellow)
	a.table.SetSelectable(true, false) // Selectable rows, not columns
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite))

	// Populate table
	a.populateList()

	// Set up VIM keybindings
	a.setupVimKeybindings()

	// Handle ESC to go back to menu
	a.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("list") {
				a.pages.RemovePage("list")
			}
			a.table = nil
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

		// Handle VIM keys
		if event.Key() == tcell.KeyRune {
			rune := event.Rune()
			if rune == 'j' {
				// Move down - skip to next card (4 rows per card: 3 content + 1 blank)
				row, _ := a.table.GetSelection()
				nextRow := row + 4
				if nextRow < a.table.GetRowCount() {
					a.table.Select(nextRow, 0)
				}
				return nil
			}
			if rune == 'k' {
				// Move up - skip to previous card
				row, _ := a.table.GetSelection()
				nextRow := row - 4
				if nextRow >= 0 {
					a.table.Select(nextRow, 0)
				}
				return nil
			}
			if rune == 'l' {
				// Next page
				if a.pagination.HasMore {
					a.loadNextPage()
				}
				return nil
			}
			if rune == 'h' {
				// Previous page
				if a.currentPage > 1 {
					a.loadPreviousPage()
				}
				return nil
			}
		}

		return event
	})

	// Handle selection changed to skip blank rows
	a.table.SetSelectionChangedFunc(func(row, column int) {
		// If we're on a blank row (row % 4 == 3), move to next card
		if row%4 == 3 {
			nextRow := row + 1
			if nextRow < a.table.GetRowCount() {
				a.table.Select(nextRow, 0)
			} else {
				// If we're at the end, go to previous card
				prevRow := row - 1
				if prevRow >= 0 {
					a.table.Select(prevRow, 0)
				}
			}
		}
	})

	// Handle Enter key to show card detail
	a.table.SetSelectedFunc(func(row, column int) {
		// Calculate card index - each card is 4 rows (3 content + 1 blank)
		cardIndex := row / 4
		if cardIndex >= 0 && cardIndex < len(a.cards) {
			a.showCardDetail(a.pages, a.cards[cardIndex])
		}
	})

	// Remove old list page if it exists
	if a.pages.HasPage("list") {
		a.pages.RemovePage("list")
	}

	// Add new table page
	a.pages.AddPage("list", a.table, true, true)
	a.pages.SwitchToPage("list")

	// Set focus
	a.app.SetFocus(a.table)
}

func (a *App) populateList() {
	if a.table == nil {
		return
	}
	a.table.Clear()

	row := 0
	for i, card := range a.cards {
		// Row 1: Card Name (white)
		cell := tview.NewTableCell(card.Name).
			SetTextColor(tcell.ColorWhite).
			SetSelectable(true)
		a.table.SetCell(row, 0, cell)
		row++

		// Row 2: Type (gray)
		typeLine := card.TypeLine
		if typeLine == "" {
			typeLine = " "
		}
		cell = tview.NewTableCell(typeLine).
			SetTextColor(tcell.ColorGray).
			SetSelectable(true)
		a.table.SetCell(row, 0, cell)
		row++

		// Row 3: Rarity • Mana Cost • Set: SetCode (gray)
		infoParts := []string{}
		if card.Rarity != "" {
			rarity := strings.ToLower(card.Rarity)
			if len(rarity) > 0 {
				rarity = strings.ToUpper(rarity[:1]) + rarity[1:]
			}
			infoParts = append(infoParts, rarity)
		}
		if card.ManaCost != "" {
			infoParts = append(infoParts, card.ManaCost)
		}
		if card.SetCode != "" {
			infoParts = append(infoParts, fmt.Sprintf("Set: %s", card.SetCode))
		}

		infoLine := strings.Join(infoParts, " • ")
		if infoLine == "" {
			infoLine = " "
		}
		cell = tview.NewTableCell(infoLine).
			SetTextColor(tcell.ColorGray).
			SetSelectable(true)
		a.table.SetCell(row, 0, cell)
		row++

		// Add blank row separator (except after last card)
		if i < len(a.cards)-1 {
			cell = tview.NewTableCell("").
				SetSelectable(false)
			a.table.SetCell(row, 0, cell)
			row++
		}
	}
}

func (a *App) updateListTitle() {
	pageSize := 10
	totalPages := 1
	if a.pagination.TotalCards > 0 {
		// Calculate total pages: ceiling division
		totalPages = (a.pagination.TotalCards + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}
	}

	title := fmt.Sprintf("MTG Cards - Page %d/%d", a.currentPage, totalPages)
	if a.currentQuery != "" {
		// Truncate long queries
		queryDisplay := a.currentQuery
		if len(queryDisplay) > 40 {
			queryDisplay = queryDisplay[:37] + "..."
		}
		title += fmt.Sprintf(" | Query: %s", queryDisplay)
	}
	if a.pagination.TotalCards > 0 {
		title += fmt.Sprintf(" | Total: %d", a.pagination.TotalCards)
	}
	if a.table != nil {
		a.table.SetTitle(title)
	}
}

func (a *App) setupVimKeybindings() {
	// VIM keybindings are now handled in showCardList's SetInputCapture
	// This method is kept for compatibility but the actual handling is in handleVimKeys
}

func (a *App) handleVimKeys(event *tcell.EventKey) *tcell.EventKey {
	// VIM keys are now handled directly in showCardList's SetInputCapture
	// This method is kept for compatibility but not actively used
	return event
}

func (a *App) loadNextPage() {
	a.currentPage++
	a.loadPageFromCache()
}

func (a *App) loadPreviousPage() {
	a.currentPage--
	a.loadPageFromCache()
}

func (a *App) loadPageFromCache() {
	// Get all matching cards for this query
	allCards, exists := a.allMatchingCards[a.currentQuery]
	if !exists {
		// Fallback to API call if cache doesn't exist
		result, err := api.FetchCards(a.currentQuery, a.currentPage)
		if err != nil {
			log.Printf("Error fetching page: %v", err)
			modal := tview.NewModal().
				SetText(fmt.Sprintf("Error fetching page: %v", err)).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.pages.RemovePage("error")
				})
			a.pages.AddPage("error", modal, true, true)
			return
		}
		a.cards = result.Cards
		a.pagination = result.Pagination
	} else {
		// Use cached data
		pageSize := 10
		totalCards := len(allCards)
		startIdx := (a.currentPage - 1) * pageSize
		endIdx := startIdx + pageSize

		if startIdx >= totalCards {
			a.cards = []api.Card{}
			a.pagination = api.PaginationInfo{
				HasMore:    false,
				TotalCards: totalCards,
			}
		} else {
			if endIdx > totalCards {
				endIdx = totalCards
			}

			// Check cache first
			if cachedPage, exists := a.pageCache[a.currentQuery][a.currentPage]; exists {
				a.cards = cachedPage
			} else {
				// Calculate and cache this page
				a.cards = allCards[startIdx:endIdx]
				a.pageCache[a.currentQuery][a.currentPage] = a.cards
			}

			hasMore := endIdx < totalCards
			a.pagination = api.PaginationInfo{
				HasMore:    hasMore,
				TotalCards: totalCards,
			}
		}

		// Pre-load next pages in background
		go a.preloadPages(a.currentQuery, a.currentPage, allCards, pageSize)
	}

	a.populateList()
	a.updateListTitle()
	if a.table != nil {
		a.table.Select(0, 0) // Reset to top of table
		a.app.SetFocus(a.table)
	}
}

// preloadPages pre-loads the next few pages in the background for faster navigation
func (a *App) preloadPages(query string, currentPage int, allCards []api.Card, pageSize int) {
	// Pre-load next 2 pages
	for i := 1; i <= 2; i++ {
		page := currentPage + i
		startIdx := (page - 1) * pageSize
		endIdx := startIdx + pageSize

		if startIdx >= len(allCards) {
			break
		}

		if endIdx > len(allCards) {
			endIdx = len(allCards)
		}

		// Check if already cached
		if _, exists := a.pageCache[query][page]; !exists {
			// Cache this page
			pageCards := allCards[startIdx:endIdx]
			a.pageCache[query][page] = pageCards
		}
	}
}

// showCardDetail creates and displays a modal with card info
func (a *App) showCardDetail(pages *tview.Pages, card api.Card) {
	// Reorder to match results page: Name, Type, Rarity • Mana Cost • Set: SetCode
	// Then add Oracle Text, Set, and Rarity on separate lines

	// Format rarity with proper capitalization
	rarity := card.Rarity
	if rarity != "" {
		rarityLower := strings.ToLower(rarity)
		if len(rarityLower) > 0 {
			rarity = strings.ToUpper(rarityLower[:1]) + rarityLower[1:]
		}
	}

	// Build info line matching results page format: Rarity • Mana Cost • Set: SetCode
	infoParts := []string{}
	if rarity != "" {
		infoParts = append(infoParts, rarity)
	}
	if card.ManaCost != "" {
		infoParts = append(infoParts, card.ManaCost)
	}
	if card.SetCode != "" {
		infoParts = append(infoParts, fmt.Sprintf("Set: %s", card.SetCode))
	}
	infoLine := strings.Join(infoParts, " • ")

	// Build detail text in the same order as results page
	detailText := fmt.Sprintf(
		"[yellow]%s[white]\n\n%s",
		card.Name,
		card.TypeLine,
	)

	// Add Power/Toughness for creatures
	if strings.Contains(strings.ToLower(card.TypeLine), "creature") {
		if card.Power != "" && card.Toughness != "" {
			detailText += fmt.Sprintf("\n%s/%s", card.Power, card.Toughness)
		} else if card.Power != "" {
			detailText += fmt.Sprintf("\n%s/*", card.Power)
		} else if card.Toughness != "" {
			detailText += fmt.Sprintf("\n*/%s", card.Toughness)
		}
	}

	// Add Loyalty for planeswalkers
	if strings.Contains(strings.ToLower(card.TypeLine), "planeswalker") && card.Loyalty != "" {
		detailText += fmt.Sprintf("\nLoyalty: %s", card.Loyalty)
	}

	detailText += fmt.Sprintf("\n\n%s", infoLine)

	// Add Oracle Text if available
	if card.OracleText != "" {
		detailText += fmt.Sprintf("\n\n%s", card.OracleText)
	}

	// Create modal with proper styling
	modal := tview.NewModal().
		SetText(detailText).
		AddButtons([]string{"Back"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("detail")
		})
	modal.SetBackgroundColor(tcell.ColorBlack)

	pages.AddPage("detail", modal, true, true)
	a.app.SetFocus(modal)
}
