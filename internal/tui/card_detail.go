package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/tui/collections"
	"github.com/chr-hen/mtg-tui/internal/util"
	"github.com/gdamore/tcell/v2"
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

	// Add each printing as a list item with ownership indicator
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
		
		// Add ownership indicator
		isOwned := a.collection.IsOwned(printing.SetCode, printing.CollectorNumber)
		if isOwned {
			printingText = "[green]✓[white] " + printingText
		} else {
			printingText = "[gray]○[white] " + printingText
		}

		// Store printing for callback (capture for closure)
		printing := printing
		printingsList.AddItem(printingText, "", 0, func() {
			// Toggle ownership on Enter
			a.collection.TogglePrinting(printing.SetCode, printing.CollectorNumber)
			// Save collection
			if err := collections.SaveCollection(a.collection); err != nil {
				// Show error modal
				errorModal := tview.NewModal().
					SetText(fmt.Sprintf("Error saving collection: %v", err)).
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						pages.RemovePage("collection_error")
					})
				pages.AddPage("collection_error", errorModal, true, true)
				return
			}
			// Refresh the printings list to show updated ownership
			a.showPrintingsModal(pages, group)
		})
	}

	// Add border and title to the list
	printingsList.SetBorder(true).
		SetBorderColor(tcell.ColorYellow).
		SetTitle(fmt.Sprintf(" [yellow]Printings for: %s[white] (%d total)[yellow] ", group.Card.Name, len(group.Printings))).
		SetTitleColor(tcell.ColorYellow)

	// Set up input capture for ESC key and Vim motions
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
		
		// Handle Vim motions
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
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
		
		return event
	})

	// Create a centered modal-like view
	modalFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(printingsList, 80, 0, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("printings", modalFlex, true, true)
	pages.SwitchToPage("printings")
	a.app.SetFocus(printingsList)
}

