package tui

import (
	"fmt"
	"strings"

	"github.com/chr-hen/mtg-tui/internal/api"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

