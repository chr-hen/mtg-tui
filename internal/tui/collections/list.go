package collections

import (
	"time"
)

// List represents a Magic: The Gathering list (cube, wishlist, etc.)
type List struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	DateCreated time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
	Cards       []ListCard `json:"cards,omitempty"`
}

// ListCard represents a card in a list with quantity
type ListCard struct {
	SetCode         string `json:"set_code"`
	CollectorNumber string `json:"collector_number"`
	Quantity        int    `json:"quantity"`
}

// ListsCollection represents all user lists
type ListsCollection struct {
	Lists []List `json:"lists"`
}

// LoadLists loads all lists from the JSON file
func LoadLists() (*ListsCollection, error) {
	lists := &ListsCollection{
		Lists: []List{},
	}

	filePath := GetListsFilePath()
	if err := LoadJSONFile(filePath, lists); err != nil {
		return nil, err
	}

	return lists, nil
}

// SaveLists saves all lists to the JSON file
func SaveLists(lists *ListsCollection) error {
	filePath := GetListsFilePath()
	return SaveJSONFile(filePath, lists)
}

// GetList finds a list by ID
func (lc *ListsCollection) GetList(id string) *List {
	for i := range lc.Lists {
		if lc.Lists[i].ID == id {
			return &lc.Lists[i]
		}
	}
	return nil
}

// AddList adds a new list to the collection
func (lc *ListsCollection) AddList(list List) {
	list.DateCreated = time.Now()
	list.DateModified = time.Now()
	lc.Lists = append(lc.Lists, list)
}

// UpdateList updates an existing list
func (lc *ListsCollection) UpdateList(list List) {
	for i := range lc.Lists {
		if lc.Lists[i].ID == list.ID {
			list.DateModified = time.Now()
			lc.Lists[i] = list
			return
		}
	}
}

// DeleteList removes a list from the collection
func (lc *ListsCollection) DeleteList(id string) {
	for i := range lc.Lists {
		if lc.Lists[i].ID == id {
			lc.Lists = append(lc.Lists[:i], lc.Lists[i+1:]...)
			return
		}
	}
}

// AddCard adds a card to the list (or increments quantity if it exists)
func (l *List) AddCard(setCode, collectorNumber string, quantity int) {
	cardID := GetCardID(setCode, collectorNumber)

	// Check if already exists
	for i, card := range l.Cards {
		existingID := GetCardID(card.SetCode, card.CollectorNumber)
		if cardID == existingID {
			// Increment quantity
			l.Cards[i].Quantity += quantity
			if l.Cards[i].Quantity <= 0 {
				l.Cards[i].Quantity = quantity
			}
			return
		}
	}

	// Add new card
	newQuantity := quantity
	if newQuantity <= 0 {
		newQuantity = 1
	}
	l.Cards = append(l.Cards, ListCard{
		SetCode:         setCode,
		CollectorNumber: collectorNumber,
		Quantity:        newQuantity,
	})
}

