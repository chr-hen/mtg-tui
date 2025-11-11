package api

// FetchCardsResult contains both cards and pagination info
type FetchCardsResult struct {
	Cards []Card
	Pagination PaginationInfo
}

// FetchCards retrieves cards from the local cache for a specific page and query
// Uses the bulk data file downloaded from Scryfall
func FetchCards(query string, page int) (*FetchCardsResult, error) {
	pageSize := 10 // Smaller page size for better UX
	return SearchCardsLocal(query, page, pageSize)
}
