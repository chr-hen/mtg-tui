package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	searchInput.SetFieldWidth(35)
	// Use smaller label width to center the field within the form
	searchInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	searchInput.SetPlaceholder("Enter card name or search query (e.g., 'Lightning Bolt' or 'type:creature')")
	searchInput.SetPlaceholderTextColor(tcell.ColorGray)

	form.AddFormItem(searchInput)

	// Vim-style insert mode state
	insertMode := false
	currentFieldIndex := 0 // 0 = input field, 1+ = buttons

	// Force form to apply field colors after adding item
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	// Store search callback for Enter key shortcut
	searchCallback := func() {
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
	}

	// Store back callback for ESC key shortcut
	backCallback := func() {
		// Remove the search page first
		if a.pages.HasPage("search") {
			a.pages.RemovePage("search")
		}
		// Force a full redraw by recreating the menu
		a.showMainMenu()
	}

	form.AddButton("Search (Enter)", searchCallback)
	form.AddButton("Back (Esc)", backCallback)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// In insert mode, allow normal input except ESC to exit
		if insertMode {
			if event.Key() == tcell.KeyEscape {
				// ESC exits insert mode
				insertMode = false
				// Update title to show normal mode
				form.SetTitle(" [yellow]Search for Card[white] ")
				return nil
			}
			// Allow all other input in insert mode
			return event
		}

		// Normal mode - handle navigation and mode switching
		if event.Key() == tcell.KeyEscape {
			// ESC goes back to menu
			if a.pages.HasPage("search") {
				a.pages.RemovePage("search")
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
		if event.Key() == tcell.KeyEnter {
			// Enter triggers Search button
			searchCallback()
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'i':
				// Enter insert mode
				insertMode = true
				form.SetTitle(" [yellow]Search for Card[white] [gray](INSERT)[white] ")
				// Set focus to the input field for immediate editing
				a.app.SetFocus(searchInput)
				return nil
			case 'j':
				// Move to next field (or button)
				if currentFieldIndex == 0 {
					currentFieldIndex = 1
					form.SetFocus(1) // Move to Search button
				}
				return nil
			case 'k':
				// Move to previous field (or button)
				if currentFieldIndex > 0 {
					currentFieldIndex = 0
					form.SetFocus(0) // Move to input field
				}
				return nil
			}
			// Block all other text input in normal mode
			return nil
		}
		// Allow Tab/Shift+Tab for navigation
		return event
	})

	// Center the form horizontally only (keep vertical alignment consistent)
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 70, 0, true). // Form (fixed width, horizontally centered)
		AddItem(nil, 0, 1, false)   // Right spacer

	a.pages.AddPage("search", horizontalFlex, true, true)
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
	nameInput.SetFieldWidth(30)
	nameInput.SetPlaceholder("Card name (e.g., Lightning Bolt)")
	nameInput.SetPlaceholderTextColor(tcell.ColorGray)
	// Use smaller label width to center the field within the form
	nameInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Type
	typeInput := tview.NewInputField()
	typeInput.SetLabel("Type: ")
	typeInput.SetFieldWidth(40)
	typeInput.SetPlaceholder("t:creature (autocomplete available)")
	typeInput.SetPlaceholderTextColor(tcell.ColorGray)
	typeInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for types
	if len(a.uniqueTypes) > 0 {
		typeInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueTypes))
	}

	// Color
	colorInput := tview.NewInputField()
	colorInput.SetLabel("Color: ")
	colorInput.SetFieldWidth(30)
	colorInput.SetPlaceholder("c:r or c:uw (w/u/b/r/g)")
	colorInput.SetPlaceholderTextColor(tcell.ColorGray)
	colorInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Oracle Text
	oracleInput := tview.NewInputField()
	oracleInput.SetLabel("Oracle Text: ")
	oracleInput.SetFieldWidth(30)
	oracleInput.SetPlaceholder("o:\"draw a card\"")
	oracleInput.SetPlaceholderTextColor(tcell.ColorGray)
	oracleInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Mana Cost
	manaInput := tview.NewInputField()
	manaInput.SetLabel("Mana Cost: ")
	manaInput.SetFieldWidth(30)
	manaInput.SetPlaceholder("m:{G}{U} or mv<=3")
	manaInput.SetPlaceholderTextColor(tcell.ColorGray)
	manaInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Power/Toughness
	powerInput := tview.NewInputField()
	powerInput.SetLabel("Power: ")
	powerInput.SetFieldWidth(30)
	powerInput.SetPlaceholder("pow>=4 or pow>tou")
	powerInput.SetPlaceholderTextColor(tcell.ColorGray)
	powerInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	toughnessInput := tview.NewInputField()
	toughnessInput.SetLabel("Toughness: ")
	toughnessInput.SetFieldWidth(30)
	toughnessInput.SetPlaceholder("tou>=4")
	toughnessInput.SetPlaceholderTextColor(tcell.ColorGray)
	toughnessInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Set
	setInput := tview.NewInputField()
	setInput.SetLabel("Set: ")
	setInput.SetFieldWidth(30)
	setInput.SetPlaceholder("s:khm (autocomplete available)")
	setInput.SetPlaceholderTextColor(tcell.ColorGray)
	setInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for sets
	if len(a.uniqueSets) > 0 {
		setInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueSets))
	}

	// Rarity
	rarityInput := tview.NewInputField()
	rarityInput.SetLabel("Rarity: ")
	rarityInput.SetFieldWidth(30)
	rarityInput.SetPlaceholder("r:rare (autocomplete available)")
	rarityInput.SetPlaceholderTextColor(tcell.ColorGray)
	rarityInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for rarities
	if len(a.uniqueRarities) > 0 {
		rarityInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueRarities))
	}

	// Year
	yearInput := tview.NewInputField()
	yearInput.SetLabel("Year: ")
	yearInput.SetFieldWidth(30)
	yearInput.SetPlaceholder("year>=2020")
	yearInput.SetPlaceholderTextColor(tcell.ColorGray)
	yearInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Artist
	artistInput := tview.NewInputField()
	artistInput.SetLabel("Artist: ")
	artistInput.SetFieldWidth(30)
	artistInput.SetPlaceholder("a:avon")
	artistInput.SetPlaceholderTextColor(tcell.ColorGray)
	artistInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Keyword
	keywordInput := tview.NewInputField()
	keywordInput.SetLabel("Keyword: ")
	keywordInput.SetFieldWidth(30)
	keywordInput.SetPlaceholder("kw:flying (autocomplete available)")
	keywordInput.SetPlaceholderTextColor(tcell.ColorGray)
	keywordInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	// Add autocomplete for keywords
	if len(a.uniqueKeywords) > 0 {
		keywordInput.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueKeywords))
	}

	// Special flags
	isInput := tview.NewInputField()
	isInput.SetLabel("Is: ")
	isInput.SetFieldWidth(30)
	isInput.SetPlaceholder("is:multicolor, is:spell, is:permanent")
	isInput.SetPlaceholderTextColor(tcell.ColorGray)
	isInput.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

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

	// Vim-style insert mode state
	insertMode := false
	// Store all input fields for navigation
	inputFields := []*tview.InputField{
		nameInput, typeInput, colorInput, oracleInput, manaInput,
		powerInput, toughnessInput, setInput, rarityInput,
		yearInput, artistInput, keywordInput, isInput,
	}
	currentFieldIndex := 0

	// Store search callback for Enter key shortcut
	searchCallback := func() {
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
			// If set value contains spaces, quote it to preserve as single token
			if strings.Contains(set, " ") && !strings.HasPrefix(set, `"`) && !strings.HasSuffix(set, `"`) {
				set = `"` + set + `"`
			}
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
	}

	// Store clear callback for Ctrl+D shortcut
	clearCallback := func() {
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
	}

	// Store back callback for ESC key shortcut
	backCallback := func() {
		// Force a full redraw by recreating the menu
		a.showMainMenu()
	}

	form.AddButton("Search (Enter)", searchCallback)
	form.AddButton("Clear (Ctrl+D)", clearCallback)
	form.AddButton("Back (Esc)", backCallback)

	// Helper function to update title based on mode
	updateTitle := func() {
		if insertMode {
			form.SetTitle(" [yellow]Advanced Search (Scryfall Syntax)[white] [gray](INSERT)[white] ")
		} else {
			form.SetTitle(" [yellow]Advanced Search (Scryfall Syntax)[white] ")
		}
	}

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
			// ESC goes back to menu
			if a.pages.HasPage("advanced") {
				a.pages.RemovePage("advanced")
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
		if event.Key() == tcell.KeyEnter {
			// Enter triggers Search button
			searchCallback()
			return nil
		}
		if event.Key() == tcell.KeyCtrlD {
			// Ctrl+D triggers Clear button
			clearCallback()
			return nil
		}
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'i':
				// Enter insert mode on current field
				insertMode = true
				updateTitle()
				// Focus the current field for immediate editing
				if currentFieldIndex < len(inputFields) {
					a.app.SetFocus(inputFields[currentFieldIndex])
				}
				return nil
			case 'j':
				// Move to next field
				if currentFieldIndex < len(inputFields)-1 {
					currentFieldIndex++
					form.SetFocus(currentFieldIndex)
				} else {
					// Move to first button
					form.SetFocus(len(inputFields))
				}
				return nil
			case 'k':
				// Move to previous field
				if currentFieldIndex > 0 {
					currentFieldIndex--
					form.SetFocus(currentFieldIndex)
				} else {
					// Wrap to last field
					currentFieldIndex = len(inputFields) - 1
					form.SetFocus(currentFieldIndex)
				}
				return nil
			}
			// Block all other text input in normal mode
			return nil
		}
		// Allow Tab/Shift+Tab for navigation
		if event.Key() == tcell.KeyTab {
			// Tab navigation - update currentFieldIndex after navigation
			// We'll track this by checking which item has focus after Tab
			return event
		}
		return event
	})

	// Center the form horizontally only (keep vertical alignment consistent)
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).  // Left spacer
		AddItem(form, 90, 0, true). // Form (fixed width, wider for advanced, horizontally centered)
		AddItem(nil, 0, 1, false)   // Right spacer

	a.pages.AddPage("advanced", horizontalFlex, true, true)
	a.pages.SwitchToPage("advanced")
	a.app.SetFocus(form)
}
