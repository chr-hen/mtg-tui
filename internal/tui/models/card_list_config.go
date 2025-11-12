package models

import (
	"github.com/chr-hen/mtg-tui/internal/api"
)

// CardListConfig configures the unified card list view
type CardListConfig struct {
	// Identity
	PageName string // "collection", "list", "deck_123", etc.

	// Display
	TitleFunc  func() // Updates the title bar
	FooterText string // Footer controls text

	// Data & Filtering
	FilterFunc func([]api.Card) ([]api.Card, error)

	// Navigation callbacks
	OnEscape func() // ESC key handler
	OnFilter func() // 'f' key handler

	// Optional: for future extensibility
	ExtraKeybinds map[rune]func() // Additional keybindings
}

