package collections

// Wants represents the user's wanted card printings
type Wants struct {
	WantedPrintings []CardPrintingID `json:"wanted_printings"`
}

// LoadWants loads the wants from the JSON file
func LoadWants() (*Wants, error) {
	wants := &Wants{
		WantedPrintings: []CardPrintingID{},
	}

	filePath := GetWantsFilePath()
	if err := LoadJSONFile(filePath, wants); err != nil {
		return nil, err
	}

	return wants, nil
}

// SaveWants saves the wants to the JSON file
func SaveWants(wants *Wants) error {
	filePath := GetWantsFilePath()
	return SaveJSONFile(filePath, wants)
}

// IsWanted checks if a specific printing is wanted
func (w *Wants) IsWanted(setCode, collectorNumber string) bool {
	cardID := GetCardID(setCode, collectorNumber)
	for _, wanted := range w.WantedPrintings {
		wantedID := GetCardID(wanted.SetCode, wanted.CollectorNumber)
		if cardID == wantedID {
			return true
		}
	}
	return false
}

// AddPrinting adds a printing to the wants (or updates quantity if it exists)
func (w *Wants) AddPrinting(setCode, collectorNumber string, quantity int) {
	cardID := GetCardID(setCode, collectorNumber)

	// Check if already exists
	for i, wanted := range w.WantedPrintings {
		wantedID := GetCardID(wanted.SetCode, wanted.CollectorNumber)
		if cardID == wantedID {
			// Update quantity
			if quantity > 0 {
				w.WantedPrintings[i].Quantity = quantity
			} else {
				w.WantedPrintings[i].Quantity++
			}
			return
		}
	}

	// Add new printing
	newQuantity := quantity
	if newQuantity <= 0 {
		newQuantity = 1
	}
	w.WantedPrintings = append(w.WantedPrintings, CardPrintingID{
		SetCode:         setCode,
		CollectorNumber: collectorNumber,
		Quantity:        newQuantity,
	})
}

// RemovePrinting removes a printing from the wants
func (w *Wants) RemovePrinting(setCode, collectorNumber string) {
	cardID := GetCardID(setCode, collectorNumber)

	for i, wanted := range w.WantedPrintings {
		wantedID := GetCardID(wanted.SetCode, wanted.CollectorNumber)
		if cardID == wantedID {
			// Remove from slice
			w.WantedPrintings = append(w.WantedPrintings[:i], w.WantedPrintings[i+1:]...)
			return
		}
	}
}

// TogglePrinting toggles wanted status of a printing
func (w *Wants) TogglePrinting(setCode, collectorNumber string) {
	if w.IsWanted(setCode, collectorNumber) {
		w.RemovePrinting(setCode, collectorNumber)
	} else {
		w.AddPrinting(setCode, collectorNumber, 1)
	}
}

