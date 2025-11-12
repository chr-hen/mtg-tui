package collections

import (
	"time"
)

// DeckFormat represents the format of a deck
type DeckFormat string

const (
	FormatCommander          DeckFormat = "Commander"
	FormatStandard           DeckFormat = "Standard"
	FormatModern             DeckFormat = "Modern"
	FormatLegacy             DeckFormat = "Legacy"
	FormatCustom             DeckFormat = "Custom"
	Format1v1Commander       DeckFormat = "1v1 Commander"
	FormatAlchemy            DeckFormat = "Alchemy"
	FormatStandardBrawl      DeckFormat = "Standard Brawl"
	FormatPauper             DeckFormat = "Pauper"
	FormatCanadianHighlander DeckFormat = "Canadian Highlander"
	FormatDuelCommander      DeckFormat = "Duel Commander"
	FormatFrontier           DeckFormat = "Frontier"
	FormatFutureStandard     DeckFormat = "Future Standard"
	FormatGladiator          DeckFormat = "Gladiator"
	FormatHistoric           DeckFormat = "Historic"
	FormatPennyDreadful      DeckFormat = "Penny Dreadful"
	FormatOathbreaker        DeckFormat = "Oathbreaker"
	FormatBrawl              DeckFormat = "Brawl"
	FormatPauperEDH          DeckFormat = "Pauper EDH"
	FormatPioneer            DeckFormat = "Pioneer"
	FormatPreDH              DeckFormat = "PreDH"
	FormatPremodern          DeckFormat = "Premodern"
	FormatVintage            DeckFormat = "Vintage"
	FormatTimeless           DeckFormat = "Timeless"
)

// Deck represents a Magic: The Gathering deck
type Deck struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Format       DeckFormat `json:"format"`
	Description  string     `json:"description,omitempty"`
	Commander    string     `json:"commander,omitempty"` // Card ID for Commander format
	DateCreated  time.Time  `json:"date_created"`
	DateModified time.Time  `json:"date_modified"`
	MainDeck     []DeckCard `json:"main_deck"`
	Sideboard    []DeckCard `json:"sideboard,omitempty"`
}

// DeckCard represents a card in a deck with quantity
type DeckCard struct {
	SetCode         string `json:"set_code"`
	CollectorNumber string `json:"collector_number"`
	Quantity        int    `json:"quantity"`
}

// DecksCollection represents all user decks
type DecksCollection struct {
	Decks []Deck `json:"decks"`
}

// LoadDecks loads all decks from the JSON file
func LoadDecks() (*DecksCollection, error) {
	decks := &DecksCollection{
		Decks: []Deck{},
	}

	filePath := GetDecksFilePath()
	if err := LoadJSONFile(filePath, decks); err != nil {
		return nil, err
	}

	return decks, nil
}

// SaveDecks saves all decks to the JSON file
func SaveDecks(decks *DecksCollection) error {
	filePath := GetDecksFilePath()
	return SaveJSONFile(filePath, decks)
}

// GetDeck finds a deck by ID
func (dc *DecksCollection) GetDeck(id string) *Deck {
	for i := range dc.Decks {
		if dc.Decks[i].ID == id {
			return &dc.Decks[i]
		}
	}
	return nil
}

// AddDeck adds a new deck to the collection
func (dc *DecksCollection) AddDeck(deck Deck) {
	deck.DateCreated = time.Now()
	deck.DateModified = time.Now()
	dc.Decks = append(dc.Decks, deck)
}

// UpdateDeck updates an existing deck
func (dc *DecksCollection) UpdateDeck(deck Deck) {
	for i := range dc.Decks {
		if dc.Decks[i].ID == deck.ID {
			deck.DateModified = time.Now()
			dc.Decks[i] = deck
			return
		}
	}
}

// DeleteDeck removes a deck from the collection
func (dc *DecksCollection) DeleteDeck(id string) {
	for i := range dc.Decks {
		if dc.Decks[i].ID == id {
			dc.Decks = append(dc.Decks[:i], dc.Decks[i+1:]...)
			return
		}
	}
}

// GetAllFormats returns all available deck formats
func GetAllFormats() []DeckFormat {
	return []DeckFormat{
		FormatCommander,
		FormatStandard,
		FormatModern,
		FormatLegacy,
		FormatCustom,
		Format1v1Commander,
		FormatAlchemy,
		FormatStandardBrawl,
		FormatPauper,
		FormatCanadianHighlander,
		FormatDuelCommander,
		FormatFrontier,
		FormatFutureStandard,
		FormatGladiator,
		FormatHistoric,
		FormatPennyDreadful,
		FormatOathbreaker,
		FormatBrawl,
		FormatPauperEDH,
		FormatPioneer,
		FormatPreDH,
		FormatPremodern,
		FormatVintage,
		FormatTimeless,
	}
}

// GetFormatRules returns format-specific rules
func GetFormatRules(format DeckFormat) FormatRules {
	switch format {
	case FormatCommander, FormatDuelCommander:
		return FormatRules{
			MinDeckSize:      100,
			MaxDeckSize:      100,
			HasCommander:     true,
			HasSideboard:     false,
			MaxCopies:        1, // Singleton format
			MaxSideboardSize: 0,
		}
	case Format1v1Commander:
		return FormatRules{
			MinDeckSize:      100,
			MaxDeckSize:      100,
			HasCommander:     true,
			HasSideboard:     true,
			MaxCopies:        1, // Singleton format
			MaxSideboardSize: 10,
		}
	case FormatBrawl, FormatStandardBrawl:
		return FormatRules{
			MinDeckSize:      60,
			MaxDeckSize:      60,
			HasCommander:     true,
			HasSideboard:     false,
			MaxCopies:        1, // Singleton format
			MaxSideboardSize: 0,
		}
	case FormatOathbreaker:
		return FormatRules{
			MinDeckSize:      60,
			MaxDeckSize:      60,
			HasCommander:     true, // Oathbreaker + Signature Spell
			HasSideboard:     false,
			MaxCopies:        1, // Singleton format
			MaxSideboardSize: 0,
		}
	case FormatPauperEDH, FormatPreDH:
		return FormatRules{
			MinDeckSize:      100,
			MaxDeckSize:      100,
			HasCommander:     true,
			HasSideboard:     false,
			MaxCopies:        1, // Singleton format
			MaxSideboardSize: 0,
		}
	case FormatCanadianHighlander:
		return FormatRules{
			MinDeckSize:      100,
			MaxDeckSize:      100,
			HasCommander:     false,
			HasSideboard:     true,
			MaxCopies:        1, // Singleton format
			MaxSideboardSize: 10,
		}
	case FormatPauper:
		return FormatRules{
			MinDeckSize:      60,
			MaxDeckSize:      0, // No max
			HasCommander:     false,
			HasSideboard:     true,
			MaxCopies:        4,
			MaxSideboardSize: 15,
		}
	case FormatCustom:
		return FormatRules{
			MinDeckSize:      0, // No minimum
			MaxDeckSize:      0, // No max
			HasCommander:     false,
			HasSideboard:     false,
			MaxCopies:        0, // No limit
			MaxSideboardSize: 0,
		}
	default:
		// Standard, Modern, Pioneer, Legacy, Vintage, Historic, Timeless,
		// Alchemy, Frontier, Future Standard, Gladiator, Penny Dreadful, Premodern
		return FormatRules{
			MinDeckSize:      60,
			MaxDeckSize:      0, // No max
			HasCommander:     false,
			HasSideboard:     true,
			MaxCopies:        4,
			MaxSideboardSize: 15,
		}
	}
}

// FormatRules defines the rules for a specific format
type FormatRules struct {
	MinDeckSize      int
	MaxDeckSize      int // 0 means no max
	HasCommander     bool
	HasSideboard     bool
	MaxCopies        int
	MaxSideboardSize int
}

// AddCardToMainDeck adds a card to the main deck (or increments quantity if it exists)
func (d *Deck) AddCardToMainDeck(setCode, collectorNumber string, quantity int) {
	cardID := GetCardID(setCode, collectorNumber)

	// Check if already exists
	for i, card := range d.MainDeck {
		existingID := GetCardID(card.SetCode, card.CollectorNumber)
		if cardID == existingID {
			// Increment quantity
			d.MainDeck[i].Quantity += quantity
			if d.MainDeck[i].Quantity <= 0 {
				d.MainDeck[i].Quantity = quantity
			}
			return
		}
	}

	// Add new card
	newQuantity := quantity
	if newQuantity <= 0 {
		newQuantity = 1
	}
	d.MainDeck = append(d.MainDeck, DeckCard{
		SetCode:         setCode,
		CollectorNumber: collectorNumber,
		Quantity:        newQuantity,
	})
}

// AddCardToSideboard adds a card to the sideboard (or increments quantity if it exists)
func (d *Deck) AddCardToSideboard(setCode, collectorNumber string, quantity int) {
	cardID := GetCardID(setCode, collectorNumber)

	// Check if already exists
	for i, card := range d.Sideboard {
		existingID := GetCardID(card.SetCode, card.CollectorNumber)
		if cardID == existingID {
			// Increment quantity
			d.Sideboard[i].Quantity += quantity
			if d.Sideboard[i].Quantity <= 0 {
				d.Sideboard[i].Quantity = quantity
			}
			return
		}
	}

	// Add new card
	newQuantity := quantity
	if newQuantity <= 0 {
		newQuantity = 1
	}
	d.Sideboard = append(d.Sideboard, DeckCard{
		SetCode:         setCode,
		CollectorNumber: collectorNumber,
		Quantity:        newQuantity,
	})
}
