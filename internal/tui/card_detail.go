package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/chr-hen/mtg-tui/internal/tui/components"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
)

// showCardDetail creates and displays a modal with card info, showing all printings
func (a *App) showCardDetail(pages *tview.Pages, group util.CardGroup) {
	card := group.Card // Use canonical card for main info

	// Format rarity with proper capitalization
	rarity := card.Rarity
	if rarity != "" {
		rarityLower := strings.ToLower(rarity)
		if len(rarityLower) > 0 {
			rarity = strings.ToUpper(rarityLower[:1]) + rarityLower[1:]
		}
	}

	// Build info line matching results page format: Rarity • Mana Cost
	infoParts := []string{}
	if rarity != "" {
		infoParts = append(infoParts, rarity)
	}
	if card.ManaCost != "" {
		infoParts = append(infoParts, card.ManaCost)
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

	// Create buttons - add "View Printings" if there are multiple printings
	buttons := []string{"Back"}
	if len(group.Printings) > 0 {
		buttons = []string{"View Printings", "Back"}
	}

	// Create modal with proper styling
	modal := tview.NewModal().
		SetText(detailText).
		AddButtons(buttons).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "View Printings" {
				// Show printings modal
				a.showPrintingsModal(pages, group)
			} else {
				// Back button
				pages.RemovePage("detail")
			}
		})
	modal.SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Card Details[white] ").
		SetTitleColor(tcell.ColorYellow)

	pages.AddPage("detail", modal, true, true)
	a.app.SetFocus(modal)
}

// showPrintingsModal displays all printings in a scrollable list with ownership toggles
func (a *App) showPrintingsModal(pages *tview.Pages, group util.CardGroup) {
	// Remove old printings page if it exists
	if a.pages.HasPage("printings") {
		a.pages.RemovePage("printings")
	}
	// Create a list to display printings (scrollable)
	printingsList := tview.NewList().
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Add each printing as a list item with ownership and wants indicators
	for _, printing := range group.Printings {
		// Format printing info
		printingText := fmt.Sprintf("Set: %s", printing.SetCode)
		if printing.CollectorNumber != "" {
			setNumStr := printing.CollectorNumber
			if printing.SetSize > 0 {
				setNumStr = fmt.Sprintf("%s/%d", printing.CollectorNumber, printing.SetSize)
			}
			printingText += fmt.Sprintf(" • SetNum: %s", setNumStr)
		}
		if printing.Set != "" {
			printingText += fmt.Sprintf(" • %s", printing.Set)
		}
		
		// Add collection indicator
		isOwned := a.collection.IsOwned(printing.SetCode, printing.CollectorNumber)
		if isOwned {
			printingText = "[green]✓[white]" + printingText
		} else {
			printingText = "[gray]○[white]" + printingText
		}
		
		// Add wants indicator
		isWanted := a.wants.IsWanted(printing.SetCode, printing.CollectorNumber)
		if isWanted {
			printingText = "[yellow]★[white]" + printingText
		} else {
			printingText = "[gray]☆[white]" + printingText
		}

		// No action on Enter - use 'c' key instead
		printingsList.AddItem(printingText, "", 0, nil)
	}

	// Add border and title to the list
	printingsList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(fmt.Sprintf(" [yellow]Printings for: %s[white] (%d total)[yellow] ", group.Card.Name, len(group.Printings))).
		SetTitleColor(tcell.ColorYellow)

	// Set up input capture for ESC key, shortcuts, and Vim motions
	printingsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.RemovePage("printings")
			// Return to detail modal
			if pages.HasPage("detail") {
				pages.SwitchToPage("detail")
				a.app.SetFocus(pages)
			}
			return nil
		}
		
		// Handle keyboard shortcuts
		if event.Key() == tcell.KeyRune {
			currentIndex := printingsList.GetCurrentItem()
			if currentIndex >= 0 && currentIndex < len(group.Printings) {
				printing := group.Printings[currentIndex]
				
				switch event.Rune() {
				case 'c':
					// Toggle collection
					a.collection.TogglePrinting(printing.SetCode, printing.CollectorNumber)
					if err := collections.SaveCollection(a.collection); err != nil {
						a.showErrorModal(fmt.Sprintf("Error saving collection: %v", err))
						return nil
					}
					// Refresh to show updated checkbox
					a.showPrintingsModal(pages, group)
					return nil
				case 'w':
					// Toggle wants
					a.wants.TogglePrinting(printing.SetCode, printing.CollectorNumber)
					if err := collections.SaveWants(a.wants); err != nil {
						a.showErrorModal(fmt.Sprintf("Error saving wants: %v", err))
						return nil
					}
					// Refresh to show updated checkbox
					a.showPrintingsModal(pages, group)
					return nil
				case 'l':
					// Show list selection
					a.showListSelectionModal(pages, printing, group)
					return nil
				case 'd':
					// Show deck selection
					a.showDeckSelectionModal(pages, printing, group)
					return nil
				case 'j':
					// Move down
					current := printingsList.GetCurrentItem()
					if current < printingsList.GetItemCount()-1 {
						printingsList.SetCurrentItem(current + 1)
					}
					return nil
				case 'k':
					// Move up
					current := printingsList.GetCurrentItem()
					if current > 0 {
						printingsList.SetCurrentItem(current - 1)
					}
					return nil
				}
			}
		}
		
		return event
	})

	// Create footer with controls (split into two lines for readability)
	footerText := "c: collection | l: list | d: deck | w: wants\nEsc: back"
	footer := components.CreateCardListFooter(footerText)

	// Create a centered modal-like view with footer
	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(printingsList, 0, 1, true).
				AddItem(footer, 5, 0, false), 80, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("printings", modalFlex, true, true)
	pages.SwitchToPage("printings")
	a.app.SetFocus(printingsList)
}

// showQuantityInputModal displays a modal to input quantity
func (a *App) showQuantityInputModal(pages *tview.Pages, title string, callback func(int)) {
	// Remove old quantity modal if it exists
	if a.pages.HasPage("quantity_input") {
		a.pages.RemovePage("quantity_input")
	}

	// Create form
	form := tview.NewForm()
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(fmt.Sprintf(" [yellow]%s[white] ", title)).
		SetTitleColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorBlack)

	var quantityStr string

	// Add quantity input field
	quantityField := tview.NewInputField().
		SetLabel("Quantity").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetChangedFunc(func(text string) {
			quantityStr = text
		})
	quantityField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(quantityField)

	// Add buttons
	form.AddButton("Add", func() {
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			a.showErrorModal("Please enter a valid positive quantity")
			return
		}
		// Remove modal
		if a.pages.HasPage("quantity_input") {
			a.pages.RemovePage("quantity_input")
		}
		// Call callback with quantity
		callback(quantity)
	})

	form.AddButton("Cancel", func() {
		if a.pages.HasPage("quantity_input") {
			a.pages.RemovePage("quantity_input")
		}
	})

	// Style buttons
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	// Handle ESC
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if a.pages.HasPage("quantity_input") {
				a.pages.RemovePage("quantity_input")
			}
			return nil
		}
		return event
	})

	// Create centered modal
	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("quantity_input", modalFlex, true, true)
	a.pages.SwitchToPage("quantity_input")
	a.app.SetFocus(form)
	form.SetFocus(0)
}

// showSuccessModal displays a brief success message
func (a *App) showSuccessModal(pages *tview.Pages, message string) {
	// Remove old success modal if it exists
	if a.pages.HasPage("success_modal") {
		a.pages.RemovePage("success_modal")
	}

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if a.pages.HasPage("success_modal") {
				a.pages.RemovePage("success_modal")
			}
		})
	modal.SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorGreen).
		SetTitle(" [green]Success[white] ").
		SetTitleColor(tcell.ColorGreen)

	a.pages.AddPage("success_modal", modal, true, true)
	a.pages.SwitchToPage("success_modal")
	a.app.SetFocus(modal)

	// Auto-dismiss after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		a.app.QueueUpdateDraw(func() {
			if a.pages.HasPage("success_modal") {
				a.pages.RemovePage("success_modal")
			}
		})
	}()
}

// showListSelectionModal displays a list of lists to add the card to
func (a *App) showListSelectionModal(pages *tview.Pages, printing api.Card, group util.CardGroup) {
	// Remove old list selection page if it exists
	if a.pages.HasPage("list_selection") {
		a.pages.RemovePage("list_selection")
	}

	// Load lists
	listsCollection, err := collections.LoadLists()
	if err != nil {
		listsCollection = &collections.ListsCollection{Lists: []collections.List{}}
	}

	// Create list for existing lists
	listsList := tview.NewList()
	listsList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Select List to Add Card[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	listsList.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Add existing lists
	if len(listsCollection.Lists) == 0 {
		listsList.AddItem("No lists yet", "Create your first list", 0, nil)
	} else {
		for _, list := range listsCollection.Lists {
			listCopy := list // Capture for closure
			listInfo := fmt.Sprintf("%d cards", len(list.Cards))
			if list.Description != "" {
				listInfo = list.Description + " • " + listInfo
			}
			listsList.AddItem(list.Name, listInfo, 0, func() {
				// Show quantity input
				a.showQuantityInputModal(pages, fmt.Sprintf("Add to %s", listCopy.Name), func(quantity int) {
					// Add card to list
					listCopy.AddCard(printing.SetCode, printing.CollectorNumber, quantity)
					// Update the list in the collection
					listsCollection.UpdateList(listCopy)
					// Save lists
					if err := collections.SaveLists(listsCollection); err != nil {
						a.showErrorModal(fmt.Sprintf("Error saving list: %v", err))
						return
					}
					// Show success and return to printings
					a.showSuccessModal(pages, fmt.Sprintf("Added %d to %s", quantity, listCopy.Name))
					// Return to printings view
					if a.pages.HasPage("list_selection") {
						a.pages.RemovePage("list_selection")
					}
					a.showPrintingsModal(pages, group)
				})
			})
		}
	}

	// Handle list selection
	listsList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if len(listsCollection.Lists) > 0 && index < len(listsCollection.Lists) {
			list := listsCollection.Lists[index]
			// Show quantity input
			a.showQuantityInputModal(pages, fmt.Sprintf("Add to %s", list.Name), func(quantity int) {
				// Add card to list
				list.AddCard(printing.SetCode, printing.CollectorNumber, quantity)
				// Update the list in the collection
				listsCollection.UpdateList(list)
				// Save lists
				if err := collections.SaveLists(listsCollection); err != nil {
					a.showErrorModal(fmt.Sprintf("Error saving list: %v", err))
					return
				}
				// Show success and return to printings
				a.showSuccessModal(pages, fmt.Sprintf("Added %d to %s", quantity, list.Name))
				// Return to printings view
				if a.pages.HasPage("list_selection") {
					a.pages.RemovePage("list_selection")
				}
				a.showPrintingsModal(pages, group)
			})
		}
	})

	// Store callbacks
	createCallback := func() {
		// Create a temporary list creation that returns to selection
		a.showListCreationFormForSelection(pages, printing, group)
	}
	backCallback := func() {
		if a.pages.HasPage("list_selection") {
			a.pages.RemovePage("list_selection")
		}
		a.showPrintingsModal(pages, group)
	}

	// Create form for buttons
	form := tview.NewForm()
	form.SetBorder(false).
		SetBackgroundColor(tcell.ColorBlack)

	// Add "Create New List" button if no lists exist
	if len(listsCollection.Lists) == 0 {
		form.AddButton("Create New List (n)", createCallback)
	}
	form.AddButton("Back (b)", backCallback)

	// Style buttons
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	// Handle ESC and navigation
	listsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backCallback()
			return nil
		}

		// Handle shortcuts
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'n':
				if len(listsCollection.Lists) == 0 {
					createCallback()
					return nil
				}
			case 'b':
				backCallback()
				return nil
			case 'j':
				current := listsList.GetCurrentItem()
				if current < listsList.GetItemCount()-1 {
					listsList.SetCurrentItem(current + 1)
				}
				return nil
			case 'k':
				current := listsList.GetCurrentItem()
				if current > 0 {
					listsList.SetCurrentItem(current - 1)
				}
				return nil
			}
		}

		return event
	})

	// Create flex layout
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(listsList, 0, 1, true).
		AddItem(form, 3, 0, false)

	// Center the main flex
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, 80, 0, true).
		AddItem(nil, 0, 1, false)

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(horizontalFlex, 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("list_selection", verticalFlex, true, true)
	a.pages.SwitchToPage("list_selection")
	a.app.SetFocus(listsList)
}

// showListCreationFormForSelection creates a list and returns to selection
func (a *App) showListCreationFormForSelection(pages *tview.Pages, printing api.Card, group util.CardGroup) {
	// Similar to showListCreationForm but returns to list selection after creation
	// This is a simplified version that creates and immediately adds the card
	// For now, we'll just call the regular creation form and handle it there
	// Actually, let's create a simpler inline version
	if a.pages.HasPage("list_create_selection") {
		a.pages.RemovePage("list_create_selection")
	}

	form := tview.NewForm()
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Create New List[white] ").
		SetTitleColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorBlack)

	var listName string

	nameField := tview.NewInputField().
		SetLabel("List Name").
		SetFieldWidth(40).
		SetChangedFunc(func(text string) {
			listName = text
		})
	nameField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(nameField)

	form.AddButton("Create", func() {
		if strings.TrimSpace(listName) == "" {
			a.showErrorModal("List name is required")
			return
		}

		// Create list
		list := collections.List{
			ID:          uuid.New().String(),
			Name:        strings.TrimSpace(listName),
			Description: "",
			Notes:       "",
			Cards:       []collections.ListCard{},
		}

		// Load existing lists
		listsCollection, err := collections.LoadLists()
		if err != nil {
			listsCollection = &collections.ListsCollection{Lists: []collections.List{}}
		}

		// Add new list
		listsCollection.AddList(list)

		// Save lists
		if err := collections.SaveLists(listsCollection); err != nil {
			a.showErrorModal(fmt.Sprintf("Failed to save list: %v", err))
			return
		}

		// Remove create page and return to selection
		if a.pages.HasPage("list_create_selection") {
			a.pages.RemovePage("list_create_selection")
		}
		// Show list selection again (will now include the new list)
		a.showListSelectionModal(pages, printing, group)
	})

	form.AddButton("Cancel", func() {
		if a.pages.HasPage("list_create_selection") {
			a.pages.RemovePage("list_create_selection")
		}
		a.showListSelectionModal(pages, printing, group)
	})

	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("list_create_selection", modalFlex, true, true)
	a.pages.SwitchToPage("list_create_selection")
	a.app.SetFocus(form)
	form.SetFocus(0)
}

// showDeckSelectionModal displays a list of decks to add the card to
func (a *App) showDeckSelectionModal(pages *tview.Pages, printing api.Card, group util.CardGroup) {
	// Remove old deck selection page if it exists
	if a.pages.HasPage("deck_selection") {
		a.pages.RemovePage("deck_selection")
	}

	// Load decks
	decksCollection, err := collections.LoadDecks()
	if err != nil {
		decksCollection = &collections.DecksCollection{Decks: []collections.Deck{}}
	}

	// Create list for existing decks
	decksList := tview.NewList()
	decksList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Select Deck to Add Card[white] ").
		SetTitleColor(tcell.ColorYellow)

	// Set selection colors
	decksList.SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedTextColor(tcell.ColorWhite)

	// Add existing decks
	if len(decksCollection.Decks) == 0 {
		decksList.AddItem("No decks yet", "Create your first deck", 0, nil)
	} else {
		for _, deck := range decksCollection.Decks {
			deckCopy := deck // Capture for closure
			deckInfo := fmt.Sprintf("%s • %d cards", deck.Format, len(deck.MainDeck))
			decksList.AddItem(deck.Name, deckInfo, 0, func() {
				// Ask MainDeck or Sideboard
				a.showDeckLocationModal(pages, printing, group, deckCopy)
			})
		}
	}

	// Handle deck selection
	decksList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if len(decksCollection.Decks) > 0 && index < len(decksCollection.Decks) {
			deck := decksCollection.Decks[index]
			// Ask MainDeck or Sideboard
			a.showDeckLocationModal(pages, printing, group, deck)
		}
	})

	// Store callbacks
	createCallback := func() {
		a.showDeckCreationFormForSelection(pages, printing, group)
	}
	backCallback := func() {
		if a.pages.HasPage("deck_selection") {
			a.pages.RemovePage("deck_selection")
		}
		a.showPrintingsModal(pages, group)
	}

	// Create form for buttons
	form := tview.NewForm()
	form.SetBorder(false).
		SetBackgroundColor(tcell.ColorBlack)

	// Add "Create New Deck" button if no decks exist
	if len(decksCollection.Decks) == 0 {
		form.AddButton("Create New Deck (n)", createCallback)
	}
	form.AddButton("Back (b)", backCallback)

	// Style buttons
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	// Handle ESC and navigation
	decksList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			backCallback()
			return nil
		}

		// Handle shortcuts
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'n':
				if len(decksCollection.Decks) == 0 {
					createCallback()
					return nil
				}
			case 'b':
				backCallback()
				return nil
			case 'j':
				current := decksList.GetCurrentItem()
				if current < decksList.GetItemCount()-1 {
					decksList.SetCurrentItem(current + 1)
				}
				return nil
			case 'k':
				current := decksList.GetCurrentItem()
				if current > 0 {
					decksList.SetCurrentItem(current - 1)
				}
				return nil
			}
		}

		return event
	})

	// Create flex layout
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(decksList, 0, 1, true).
		AddItem(form, 3, 0, false)

	// Center the main flex
	horizontalFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, 80, 0, true).
		AddItem(nil, 0, 1, false)

	verticalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(horizontalFlex, 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("deck_selection", verticalFlex, true, true)
	a.pages.SwitchToPage("deck_selection")
	a.app.SetFocus(decksList)
}

// showDeckLocationModal asks whether to add to MainDeck or Sideboard
func (a *App) showDeckLocationModal(pages *tview.Pages, printing api.Card, group util.CardGroup, deck collections.Deck) {
	if a.pages.HasPage("deck_location") {
		a.pages.RemovePage("deck_location")
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Add to %s:\n\nMain Deck or Sideboard?", deck.Name)).
		AddButtons([]string{"Main Deck", "Sideboard", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if a.pages.HasPage("deck_location") {
				a.pages.RemovePage("deck_location")
			}
			if buttonLabel == "Cancel" {
				a.showDeckSelectionModal(pages, printing, group)
				return
			}
			// Show quantity input
			location := "Main Deck"
			if buttonLabel == "Sideboard" {
				location = "Sideboard"
			}
			a.showQuantityInputModal(pages, fmt.Sprintf("Add to %s (%s)", deck.Name, location), func(quantity int) {
				// Load decks
				decksCollection, err := collections.LoadDecks()
				if err != nil {
					a.showErrorModal(fmt.Sprintf("Error loading decks: %v", err))
					return
				}
				// Find and update deck
				deckToUpdate := decksCollection.GetDeck(deck.ID)
				if deckToUpdate == nil {
					a.showErrorModal("Deck not found")
					return
				}
				// Add card to appropriate location
				if buttonLabel == "Sideboard" {
					deckToUpdate.AddCardToSideboard(printing.SetCode, printing.CollectorNumber, quantity)
				} else {
					deckToUpdate.AddCardToMainDeck(printing.SetCode, printing.CollectorNumber, quantity)
				}
				// Update deck
				decksCollection.UpdateDeck(*deckToUpdate)
				// Save decks
				if err := collections.SaveDecks(decksCollection); err != nil {
					a.showErrorModal(fmt.Sprintf("Error saving deck: %v", err))
					return
				}
				// Show success and return to printings
				a.showSuccessModal(pages, fmt.Sprintf("Added %d to %s (%s)", quantity, deck.Name, location))
				// Return to printings view
				if a.pages.HasPage("deck_selection") {
					a.pages.RemovePage("deck_selection")
				}
				a.showPrintingsModal(pages, group)
			})
		})
	modal.SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Select Location[white] ").
		SetTitleColor(tcell.ColorYellow)

	a.pages.AddPage("deck_location", modal, true, true)
	a.pages.SwitchToPage("deck_location")
	a.app.SetFocus(modal)
}

// showDeckCreationFormForSelection creates a deck and returns to selection
func (a *App) showDeckCreationFormForSelection(pages *tview.Pages, printing api.Card, group util.CardGroup) {
	if a.pages.HasPage("deck_create_selection") {
		a.pages.RemovePage("deck_create_selection")
	}

	form := tview.NewForm()
	form.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(" [yellow]Create New Deck[white] ").
		SetTitleColor(tcell.ColorYellow).
		SetBackgroundColor(tcell.ColorBlack)

	var deckName string
	var selectedFormat collections.DeckFormat

	formats := collections.GetAllFormats()
	formatStrings := make([]string, len(formats))
	for i, f := range formats {
		formatStrings[i] = string(f)
	}

	nameField := tview.NewInputField().
		SetLabel("Deck Name").
		SetFieldWidth(40).
		SetChangedFunc(func(text string) {
			deckName = text
		})
	nameField.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(nameField)

	formatDropdown := tview.NewDropDown().
		SetLabel("Format").
		SetOptions(formatStrings, func(option string, optionIndex int) {
			selectedFormat = formats[optionIndex]
		})
	formatDropdown.SetFieldBackgroundColor(tcell.ColorBlue).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorYellow)
	form.AddFormItem(formatDropdown)

	form.AddButton("Create", func() {
		if strings.TrimSpace(deckName) == "" {
			a.showErrorModal("Deck name is required")
			return
		}

		// Create deck
		deck := collections.Deck{
			ID:          uuid.New().String(),
			Name:        strings.TrimSpace(deckName),
			Format:      selectedFormat,
			Description: "",
			MainDeck:    []collections.DeckCard{},
			Sideboard:   []collections.DeckCard{},
		}

		// Load existing decks
		decksCollection, err := collections.LoadDecks()
		if err != nil {
			decksCollection = &collections.DecksCollection{Decks: []collections.Deck{}}
		}

		// Add new deck
		decksCollection.AddDeck(deck)

		// Save decks
		if err := collections.SaveDecks(decksCollection); err != nil {
			a.showErrorModal(fmt.Sprintf("Failed to save deck: %v", err))
			return
		}

		// Remove create page and return to selection
		if a.pages.HasPage("deck_create_selection") {
			a.pages.RemovePage("deck_create_selection")
		}
		// Show deck selection again (will now include the new deck)
		a.showDeckSelectionModal(pages, printing, group)
	})

	form.AddButton("Cancel", func() {
		if a.pages.HasPage("deck_create_selection") {
			a.pages.RemovePage("deck_create_selection")
		}
		a.showDeckSelectionModal(pages, printing, group)
	})

	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorYellow).
		SetButtonTextColor(tcell.ColorBlack)

	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("deck_create_selection", modalFlex, true, true)
	a.pages.SwitchToPage("deck_create_selection")
	a.app.SetFocus(form)
	form.SetFocus(0)
}

