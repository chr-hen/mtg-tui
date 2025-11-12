package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FilterFormFields holds all the input fields for the filter form
type FilterFormFields struct {
	Name      *tview.InputField
	Type      *tview.InputField
	Color     *tview.InputField
	Oracle    *tview.InputField
	Mana      *tview.InputField
	Power     *tview.InputField
	Toughness *tview.InputField
	Set       *tview.InputField
	Rarity    *tview.InputField
	Year      *tview.InputField
	Artist    *tview.InputField
	Keyword   *tview.InputField
	Is        *tview.InputField
}

// FilterFormConfig holds configuration for creating a filter form
type FilterFormConfig struct {
	Title         string
	ApplyCallback func(query string)
	ClearCallback func()
	BackCallback  func()
	InitialQuery  string
}

// createFilterForm creates a reusable filter form with all common fields
func (a *App) createFilterForm(config FilterFormConfig) (*tview.Form, *FilterFormFields) {
	form := tview.NewForm()
	form.SetTitle(fmt.Sprintf(" [yellow]%s[white] ", config.Title))
	form.SetBorder(true)
	form.SetBorderColor(tcell.ColorYellow)
	form.SetTitleColor(tcell.ColorYellow)
	form.SetBackgroundColor(tcell.ColorBlack)
	form.SetButtonTextColor(tcell.ColorBlack)
	form.SetButtonBackgroundColor(tcell.ColorYellow)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	fields := &FilterFormFields{}

	// Name
	fields.Name = tview.NewInputField()
	fields.Name.SetLabel("Name: ")
	fields.Name.SetFieldWidth(40)
	fields.Name.SetPlaceholder("Card name (e.g., Lightning Bolt)")
	fields.Name.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Name.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Type
	fields.Type = tview.NewInputField()
	fields.Type.SetLabel("Type: ")
	fields.Type.SetFieldWidth(40)
	fields.Type.SetPlaceholder("t:creature (autocomplete available)")
	fields.Type.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Type.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	if len(a.uniqueTypes) > 0 {
		fields.Type.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueTypes))
	}

	// Color
	fields.Color = tview.NewInputField()
	fields.Color.SetLabel("Color: ")
	fields.Color.SetFieldWidth(40)
	fields.Color.SetPlaceholder("c:r or c:uw (w/u/b/r/g)")
	fields.Color.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Color.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Oracle Text
	fields.Oracle = tview.NewInputField()
	fields.Oracle.SetLabel("Oracle Text: ")
	fields.Oracle.SetFieldWidth(40)
	fields.Oracle.SetPlaceholder("o:\"draw a card\"")
	fields.Oracle.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Oracle.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Mana Cost
	fields.Mana = tview.NewInputField()
	fields.Mana.SetLabel("Mana Cost: ")
	fields.Mana.SetFieldWidth(40)
	fields.Mana.SetPlaceholder("m:{G}{U} or mv<=3")
	fields.Mana.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Mana.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Power
	fields.Power = tview.NewInputField()
	fields.Power.SetLabel("Power: ")
	fields.Power.SetFieldWidth(40)
	fields.Power.SetPlaceholder("pow>=4 or pow>tou")
	fields.Power.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Power.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Toughness
	fields.Toughness = tview.NewInputField()
	fields.Toughness.SetLabel("Toughness: ")
	fields.Toughness.SetFieldWidth(40)
	fields.Toughness.SetPlaceholder("tou>=4")
	fields.Toughness.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Toughness.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Set
	fields.Set = tview.NewInputField()
	fields.Set.SetLabel("Set: ")
	fields.Set.SetFieldWidth(40)
	fields.Set.SetPlaceholder("s:khm (autocomplete available)")
	fields.Set.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Set.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	if len(a.uniqueSets) > 0 {
		fields.Set.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueSets))
	}

	// Rarity
	fields.Rarity = tview.NewInputField()
	fields.Rarity.SetLabel("Rarity: ")
	fields.Rarity.SetFieldWidth(40)
	fields.Rarity.SetPlaceholder("r:rare (autocomplete available)")
	fields.Rarity.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Rarity.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	if len(a.uniqueRarities) > 0 {
		fields.Rarity.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueRarities))
	}

	// Year
	fields.Year = tview.NewInputField()
	fields.Year.SetLabel("Year: ")
	fields.Year.SetFieldWidth(40)
	fields.Year.SetPlaceholder("year>=2020")
	fields.Year.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Year.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Artist
	fields.Artist = tview.NewInputField()
	fields.Artist.SetLabel("Artist: ")
	fields.Artist.SetFieldWidth(40)
	fields.Artist.SetPlaceholder("a:avon")
	fields.Artist.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Artist.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Keyword
	fields.Keyword = tview.NewInputField()
	fields.Keyword.SetLabel("Keyword: ")
	fields.Keyword.SetFieldWidth(40)
	fields.Keyword.SetPlaceholder("kw:flying (autocomplete available)")
	fields.Keyword.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Keyword.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)
	if len(a.uniqueKeywords) > 0 {
		fields.Keyword.SetAutocompleteFunc(a.createAutocompleteFunc(a.uniqueKeywords))
	}

	// Special flags
	fields.Is = tview.NewInputField()
	fields.Is.SetLabel("Is: ")
	fields.Is.SetFieldWidth(40)
	fields.Is.SetPlaceholder("is:multicolor, is:spell, is:permanent")
	fields.Is.SetPlaceholderTextColor(tcell.ColorGray)
	fields.Is.SetFormAttributes(10, tcell.ColorWhite, tcell.ColorBlack, tcell.ColorWhite, tcell.ColorDarkGray)

	// Add fields to form based on visibility settings
	if a.settings.FieldVisibility.Name {
		form.AddFormItem(fields.Name)
	}
	if a.settings.FieldVisibility.Type {
		form.AddFormItem(fields.Type)
	}
	if a.settings.FieldVisibility.Color {
		form.AddFormItem(fields.Color)
	}
	if a.settings.FieldVisibility.Oracle {
		form.AddFormItem(fields.Oracle)
	}
	if a.settings.FieldVisibility.Mana {
		form.AddFormItem(fields.Mana)
	}
	if a.settings.FieldVisibility.Power {
		form.AddFormItem(fields.Power)
	}
	if a.settings.FieldVisibility.Toughness {
		form.AddFormItem(fields.Toughness)
	}
	if a.settings.FieldVisibility.Set {
		form.AddFormItem(fields.Set)
	}
	if a.settings.FieldVisibility.Rarity {
		form.AddFormItem(fields.Rarity)
	}
	if a.settings.FieldVisibility.Year {
		form.AddFormItem(fields.Year)
	}
	if a.settings.FieldVisibility.Artist {
		form.AddFormItem(fields.Artist)
	}
	if a.settings.FieldVisibility.Keyword {
		form.AddFormItem(fields.Keyword)
	}
	if a.settings.FieldVisibility.Is {
		form.AddFormItem(fields.Is)
	}

	// Force form to apply field colors after adding all items
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorDarkGray)

	// Populate fields from initial query if provided
	if config.InitialQuery != "" {
		populateFieldsFromQuery(fields, config.InitialQuery)
	}

	return form, fields
}

// buildQueryFromFields builds a query string from filter form fields
func buildQueryFromFields(fields *FilterFormFields) string {
	var queryParts []string

	if name := strings.TrimSpace(fields.Name.GetText()); name != "" {
		queryParts = append(queryParts, name)
	}
	if typ := strings.TrimSpace(fields.Type.GetText()); typ != "" {
		if !strings.HasPrefix(typ, "t:") && !strings.HasPrefix(typ, "type:") {
			queryParts = append(queryParts, "t:"+typ)
		} else {
			queryParts = append(queryParts, typ)
		}
	}
	if color := strings.TrimSpace(fields.Color.GetText()); color != "" {
		if !strings.HasPrefix(color, "c:") && !strings.HasPrefix(color, "color:") {
			queryParts = append(queryParts, "c:"+color)
		} else {
			queryParts = append(queryParts, color)
		}
	}
	if oracle := strings.TrimSpace(fields.Oracle.GetText()); oracle != "" {
		if !strings.HasPrefix(oracle, "o:") && !strings.HasPrefix(oracle, "oracle:") {
			queryParts = append(queryParts, "o:"+oracle)
		} else {
			queryParts = append(queryParts, oracle)
		}
	}
	if mana := strings.TrimSpace(fields.Mana.GetText()); mana != "" {
		if !strings.HasPrefix(mana, "m:") && !strings.HasPrefix(mana, "mana:") && !strings.HasPrefix(mana, "mv") && !strings.HasPrefix(mana, "cmc") {
			queryParts = append(queryParts, "m:"+mana)
		} else {
			queryParts = append(queryParts, mana)
		}
	}
	if power := strings.TrimSpace(fields.Power.GetText()); power != "" {
		if !strings.HasPrefix(power, "pow:") && !strings.HasPrefix(power, "power:") {
			queryParts = append(queryParts, "pow:"+power)
		} else {
			queryParts = append(queryParts, power)
		}
	}
	if toughness := strings.TrimSpace(fields.Toughness.GetText()); toughness != "" {
		if !strings.HasPrefix(toughness, "tou:") && !strings.HasPrefix(toughness, "toughness:") {
			queryParts = append(queryParts, "tou:"+toughness)
		} else {
			queryParts = append(queryParts, toughness)
		}
	}
	if set := strings.TrimSpace(fields.Set.GetText()); set != "" {
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
	if rarity := strings.TrimSpace(fields.Rarity.GetText()); rarity != "" {
		if !strings.HasPrefix(rarity, "r:") && !strings.HasPrefix(rarity, "rarity:") {
			queryParts = append(queryParts, "r:"+rarity)
		} else {
			queryParts = append(queryParts, rarity)
		}
	}
	if year := strings.TrimSpace(fields.Year.GetText()); year != "" {
		if !strings.HasPrefix(year, "year:") {
			queryParts = append(queryParts, "year:"+year)
		} else {
			queryParts = append(queryParts, year)
		}
	}
	if artist := strings.TrimSpace(fields.Artist.GetText()); artist != "" {
		if !strings.HasPrefix(artist, "a:") && !strings.HasPrefix(artist, "artist:") {
			queryParts = append(queryParts, "a:"+artist)
		} else {
			queryParts = append(queryParts, artist)
		}
	}
	if keyword := strings.TrimSpace(fields.Keyword.GetText()); keyword != "" {
		if !strings.HasPrefix(keyword, "kw:") && !strings.HasPrefix(keyword, "keyword:") {
			queryParts = append(queryParts, "kw:"+keyword)
		} else {
			queryParts = append(queryParts, keyword)
		}
	}
	if is := strings.TrimSpace(fields.Is.GetText()); is != "" {
		if !strings.HasPrefix(is, "is:") {
			queryParts = append(queryParts, "is:"+is)
		} else {
			queryParts = append(queryParts, is)
		}
	}

	return strings.Join(queryParts, " ")
}

// populateFieldsFromQuery populates form fields from a query string
func populateFieldsFromQuery(fields *FilterFormFields, query string) {
	if strings.TrimSpace(query) == "" {
		return
	}

	conditions, err := api.ParseQuery(query)
	if err != nil {
		// If parsing fails, just return without populating
		return
	}

	// Track if we've set the name field (for plain text queries without prefixes)
	nameSet := false

	for _, cond := range conditions {
		// Build the value with operator if present
		value := cond.Value
		if cond.Operator != "" && cond.Operator != "equals" && cond.Operator != "contains" {
			value = cond.Operator + value
		}

		// Handle negation
		if cond.Negate {
			value = "-" + value
		}

		switch cond.Field {
		case "name":
			// Plain text queries go to name field
			if !nameSet {
				fields.Name.SetText(value)
				nameSet = true
			} else {
				// If name already set, append with space
				current := fields.Name.GetText()
				fields.Name.SetText(current + " " + value)
			}
		case "t", "type":
			// Remove prefix if present, but keep the value
			cleanValue := strings.TrimPrefix(value, "t:")
			cleanValue = strings.TrimPrefix(cleanValue, "type:")
			fields.Type.SetText(cleanValue)
		case "c", "color":
			cleanValue := strings.TrimPrefix(value, "c:")
			cleanValue = strings.TrimPrefix(cleanValue, "color:")
			fields.Color.SetText(cleanValue)
		case "o", "oracle":
			cleanValue := strings.TrimPrefix(value, "o:")
			cleanValue = strings.TrimPrefix(cleanValue, "oracle:")
			fields.Oracle.SetText(cleanValue)
		case "m", "mana":
			// Mana cost - could be m: or mv/cmc with operators
			if strings.HasPrefix(value, "mv") || strings.HasPrefix(value, "cmc") {
				fields.Mana.SetText(value)
			} else {
				cleanValue := strings.TrimPrefix(value, "m:")
				cleanValue = strings.TrimPrefix(cleanValue, "mana:")
				fields.Mana.SetText(cleanValue)
			}
		case "mv", "manavalue", "cmc":
			fields.Mana.SetText(value)
		case "pow", "power":
			cleanValue := strings.TrimPrefix(value, "pow:")
			cleanValue = strings.TrimPrefix(cleanValue, "power:")
			fields.Power.SetText(cleanValue)
		case "tou", "toughness":
			cleanValue := strings.TrimPrefix(value, "tou:")
			cleanValue = strings.TrimPrefix(cleanValue, "toughness:")
			fields.Toughness.SetText(cleanValue)
		case "s", "set", "e":
			cleanValue := strings.TrimPrefix(value, "s:")
			cleanValue = strings.TrimPrefix(cleanValue, "set:")
			cleanValue = strings.TrimPrefix(cleanValue, "e:")
			// Remove quotes if present
			cleanValue = strings.Trim(cleanValue, `"`)
			fields.Set.SetText(cleanValue)
		case "r", "rarity":
			cleanValue := strings.TrimPrefix(value, "r:")
			cleanValue = strings.TrimPrefix(cleanValue, "rarity:")
			fields.Rarity.SetText(cleanValue)
		case "year":
			cleanValue := strings.TrimPrefix(value, "year:")
			fields.Year.SetText(cleanValue)
		case "a", "artist":
			cleanValue := strings.TrimPrefix(value, "a:")
			cleanValue = strings.TrimPrefix(cleanValue, "artist:")
			fields.Artist.SetText(cleanValue)
		case "kw", "keyword":
			cleanValue := strings.TrimPrefix(value, "kw:")
			cleanValue = strings.TrimPrefix(cleanValue, "keyword:")
			fields.Keyword.SetText(cleanValue)
		case "is":
			cleanValue := strings.TrimPrefix(value, "is:")
			fields.Is.SetText(cleanValue)
		}
	}
}

// clearFilterFields clears all fields in the filter form
func clearFilterFields(fields *FilterFormFields) {
	fields.Name.SetText("")
	fields.Type.SetText("")
	fields.Color.SetText("")
	fields.Oracle.SetText("")
	fields.Mana.SetText("")
	fields.Power.SetText("")
	fields.Toughness.SetText("")
	fields.Set.SetText("")
	fields.Rarity.SetText("")
	fields.Year.SetText("")
	fields.Artist.SetText("")
	fields.Keyword.SetText("")
	fields.Is.SetText("")
}

// setupFilterFormInputHandling sets up vim-style input handling for the filter form
func (a *App) setupFilterFormInputHandling(
	form *tview.Form,
	fields *FilterFormFields,
	applyCallback func(),
	clearCallback func(),
	backCallback func(),
) {
	// Vim-style insert mode state
	insertMode := false
	// Store visible input fields for navigation (only fields that are actually in the form)
	inputFields := []*tview.InputField{}
	fieldOrder := []struct {
		enabled bool
		field   *tview.InputField
	}{
		{a.settings.FieldVisibility.Name, fields.Name},
		{a.settings.FieldVisibility.Type, fields.Type},
		{a.settings.FieldVisibility.Color, fields.Color},
		{a.settings.FieldVisibility.Oracle, fields.Oracle},
		{a.settings.FieldVisibility.Mana, fields.Mana},
		{a.settings.FieldVisibility.Power, fields.Power},
		{a.settings.FieldVisibility.Toughness, fields.Toughness},
		{a.settings.FieldVisibility.Set, fields.Set},
		{a.settings.FieldVisibility.Rarity, fields.Rarity},
		{a.settings.FieldVisibility.Year, fields.Year},
		{a.settings.FieldVisibility.Artist, fields.Artist},
		{a.settings.FieldVisibility.Keyword, fields.Keyword},
		{a.settings.FieldVisibility.Is, fields.Is},
	}
	for _, item := range fieldOrder {
		if item.enabled {
			inputFields = append(inputFields, item.field)
		}
	}
	currentFieldIndex := 0

	// Helper function to update title based on mode
	updateTitle := func() {
		title := form.GetTitle()
		// Remove any existing mode indicator
		title = strings.TrimSuffix(title, " [gray](INSERT)[white] ")
		if insertMode {
			form.SetTitle(title + " [gray](INSERT)[white] ")
		} else {
			form.SetTitle(title)
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
			// ESC goes back
			backCallback()
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			// Enter triggers apply callback
			applyCallback()
			return nil
		}
		if event.Key() == tcell.KeyCtrlD {
			// Ctrl+D triggers clear callback
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
		// Handle Tab/Shift+Tab for navigation
		if event.Key() == tcell.KeyTab {
			// Tab navigation - move to next field/button
			// All forms using this function have 3 buttons, so maxIndex = formItems + 2 (for indices N, N+1, N+2)
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
			// Shift+Tab navigation - move to previous field/button
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
}
