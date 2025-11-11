package collections

import (
	"fmt"
	"strings"
)

// CardPrintingID uniquely identifies a card printing
type CardPrintingID struct {
	SetCode         string `json:"set_code"`
	CollectorNumber string `json:"collector_number"`
	Quantity        int    `json:"quantity,omitempty"` // Optional: quantity if tracking multiples
}

// Collection represents the user's owned card printings
type Collection struct {
	OwnedPrintings []CardPrintingID `json:"owned_printings"`
}

// GetCardID creates a unique identifier string for a card printing
func GetCardID(setCode, collectorNumber string) string {
	return fmt.Sprintf("%s:%s", setCode, collectorNumber)
}

// ParseCardID parses a card ID string into set code and collector number
func ParseCardID(id string) (setCode, collectorNumber string) {
	parts := strings.Split(id, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

// LoadCollection loads the collection from the JSON file
func LoadCollection() (*Collection, error) {
	collection := &Collection{
		OwnedPrintings: []CardPrintingID{},
	}

	filePath := GetCollectionFilePath()
	if err := LoadJSONFile(filePath, collection); err != nil {
		return nil, err
	}

	return collection, nil
}

// SaveCollection saves the collection to the JSON file
func SaveCollection(collection *Collection) error {
	filePath := GetCollectionFilePath()
	return SaveJSONFile(filePath, collection)
}

// IsOwned checks if a specific printing is owned
func (c *Collection) IsOwned(setCode, collectorNumber string) bool {
	cardID := GetCardID(setCode, collectorNumber)
	for _, owned := range c.OwnedPrintings {
		ownedID := GetCardID(owned.SetCode, owned.CollectorNumber)
		if cardID == ownedID {
			return true
		}
	}
	return false
}

// AddPrinting adds a printing to the collection (or updates quantity if it exists)
func (c *Collection) AddPrinting(setCode, collectorNumber string, quantity int) {
	cardID := GetCardID(setCode, collectorNumber)
	
	// Check if already exists
	for i, owned := range c.OwnedPrintings {
		ownedID := GetCardID(owned.SetCode, owned.CollectorNumber)
		if cardID == ownedID {
			// Update quantity
			if quantity > 0 {
				c.OwnedPrintings[i].Quantity = quantity
			} else {
				c.OwnedPrintings[i].Quantity++
			}
			return
		}
	}

	// Add new printing
	newQuantity := quantity
	if newQuantity <= 0 {
		newQuantity = 1
	}
	c.OwnedPrintings = append(c.OwnedPrintings, CardPrintingID{
		SetCode:         setCode,
		CollectorNumber: collectorNumber,
		Quantity:        newQuantity,
	})
}

// RemovePrinting removes a printing from the collection
func (c *Collection) RemovePrinting(setCode, collectorNumber string) {
	cardID := GetCardID(setCode, collectorNumber)
	
	for i, owned := range c.OwnedPrintings {
		ownedID := GetCardID(owned.SetCode, owned.CollectorNumber)
		if cardID == ownedID {
			// Remove from slice
			c.OwnedPrintings = append(c.OwnedPrintings[:i], c.OwnedPrintings[i+1:]...)
			return
		}
	}
}

// TogglePrinting toggles ownership of a printing
func (c *Collection) TogglePrinting(setCode, collectorNumber string) {
	if c.IsOwned(setCode, collectorNumber) {
		c.RemovePrinting(setCode, collectorNumber)
	} else {
		c.AddPrinting(setCode, collectorNumber, 1)
	}
}

