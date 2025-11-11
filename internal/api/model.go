package api

// Card represents a simplified MTG card from Scryfall
type Card struct {
	Name        string   `json:"name"`
	ManaCost    string   `json:"mana_cost"`
	TypeLine    string   `json:"type_line"`
	OracleText  string   `json:"oracle_text"`
	Set         string   `json:"set_name"`
	SetCode     string   `json:"set"`
	Rarity      string   `json:"rarity"`
	Power       string   `json:"power"`
	Toughness   string   `json:"toughness"`
	Loyalty     string   `json:"loyalty"`
	CMC         float64  `json:"cmc"`
	Colors      []string `json:"colors"`
	ColorIdentity []string `json:"color_identity"`
	Keywords    []string `json:"keywords"`
	FlavorText  string   `json:"flavor_text"`
	Artist      string   `json:"artist"`
	ReleasedAt  string   `json:"released_at"`
}

// PaginationInfo contains pagination metadata from Scryfall API
type PaginationInfo struct {
	HasMore   bool   `json:"has_more"`
	NextPage  string `json:"next_page,omitempty"`
	TotalCards int    `json:"total_cards,omitempty"`
}

// BulkData represents a bulk data file from Scryfall
type BulkData struct {
	ID             string `json:"id"`
	URI            string `json:"uri"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	DownloadURI    string `json:"download_uri"`
	UpdatedAt      string `json:"updated_at"`
	Size           int    `json:"size"`
	ContentType    string `json:"content_type"`
	ContentEncoding string `json:"content_encoding"`
}

// BulkDataResponse represents the response from /bulk-data endpoint
type BulkDataResponse struct {
	Data []BulkData `json:"data"`
}
