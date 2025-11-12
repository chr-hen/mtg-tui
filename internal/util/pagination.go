package util

import (
	"github.com/chr-hen/mtg-tui/internal/api"
)

// PaginateCardGroups handles pagination logic for CardGroups
// Returns the paginated groups and pagination info
func PaginateCardGroups(cardGroups []CardGroup, page int, pageSize int) ([]CardGroup, api.PaginationInfo) {
	totalGroups := len(cardGroups)
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= totalGroups {
		return []CardGroup{}, api.PaginationInfo{
			HasMore:    false,
			TotalCards: totalGroups,
		}
	}

	if endIdx > totalGroups {
		endIdx = totalGroups
	}

	paginatedGroups := cardGroups[startIdx:endIdx]
	hasMore := endIdx < totalGroups

	return paginatedGroups, api.PaginationInfo{
		HasMore:    hasMore,
		TotalCards: totalGroups,
	}
}

