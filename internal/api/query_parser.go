package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// QueryCondition represents a single search condition
type QueryCondition struct {
	Operator string
	Field    string
	Value    string
	Negate   bool
}

// ParseQuery parses a Scryfall query string into conditions
func ParseQuery(query string) ([]QueryCondition, error) {
	if strings.TrimSpace(query) == "" || query == "*" {
		return []QueryCondition{}, nil // Empty query matches all
	}

	// Split by spaces, but preserve quoted strings
	parts := tokenizeQuery(query)
	conditions := []QueryCondition{}

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle negation
		negate := false
		if strings.HasPrefix(part, "-") {
			negate = true
			part = part[1:]
		}

		// Handle operators with colons (type:, c:, etc.)
		if strings.Contains(part, ":") {
			cond := parseCondition(part, negate)
			if cond != nil {
				conditions = append(conditions, *cond)
			}
		} else {
			// Simple text search in name
			conditions = append(conditions, QueryCondition{
				Field:    "name",
				Value:    part,
				Negate:   negate,
				Operator: "contains",
			})
		}
	}

	return conditions, nil
}

// tokenizeQuery splits a query string while preserving quoted strings
func tokenizeQuery(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, r := range query {
		if r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
		} else if r == ' ' && !inQuotes {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// parseCondition parses a single condition like "type:creature" or "pow>=4"
func parseCondition(part string, negate bool) *QueryCondition {
	// Handle comparison operators
	comparisonRegex := regexp.MustCompile(`^([a-z]+)(>=|<=|>|<|=|!=)(.+)$`)
	if matches := comparisonRegex.FindStringSubmatch(part); matches != nil {
		return &QueryCondition{
			Field:    matches[1],
			Operator: matches[2],
			Value:    strings.Trim(matches[3], `"`),
			Negate:   negate,
		}
	}

	// Handle colon-separated conditions
	if strings.Contains(part, ":") {
		parts := strings.SplitN(part, ":", 2)
		if len(parts) == 2 {
			return &QueryCondition{
				Field:    parts[0],
				Value:    strings.Trim(parts[1], `"`),
				Negate:   negate,
				Operator: "equals",
			}
		}
	}

	return nil
}

// MatchesCard checks if a card matches all the given conditions
func MatchesCard(card Card, conditions []QueryCondition) bool {
	if len(conditions) == 0 {
		return true // No conditions means match all
	}

	for _, cond := range conditions {
		if !matchesCondition(card, cond) {
			return false
		}
	}

	return true
}

// matchesCondition checks if a card matches a single condition
func matchesCondition(card Card, cond QueryCondition) bool {
	result := false

	switch cond.Field {
	case "t", "type":
		result = matchesType(card, cond.Value)
	case "c", "color":
		result = matchesColor(card, cond.Value)
	case "id", "identity":
		result = matchesColorIdentity(card, cond.Value)
	case "o", "oracle":
		result = strings.Contains(strings.ToLower(card.OracleText), strings.ToLower(cond.Value))
	case "fo", "fulloracle":
		result = strings.Contains(strings.ToLower(card.OracleText), strings.ToLower(cond.Value))
	case "kw", "keyword":
		result = matchesKeyword(card, cond.Value)
	case "m", "mana":
		result = matchesMana(card, cond.Value)
	case "mv", "manavalue", "cmc":
		result = matchesManaValue(card, cond.Value, cond.Operator)
	case "pow", "power":
		result = matchesPower(card, cond.Value, cond.Operator)
	case "tou", "toughness":
		result = matchesToughness(card, cond.Value, cond.Operator)
	case "pt", "powtou":
		result = matchesPowerToughness(card, cond.Value, cond.Operator)
	case "loy", "loyalty":
		result = matchesLoyalty(card, cond.Value, cond.Operator)
	case "r", "rarity":
		result = matchesRarity(card, cond.Value)
	case "s", "set", "e":
		result = matchesSet(card, cond.Value)
	case "a", "artist":
		result = strings.Contains(strings.ToLower(card.Artist), strings.ToLower(cond.Value))
	case "ft", "flavor":
		result = strings.Contains(strings.ToLower(card.FlavorText), strings.ToLower(cond.Value))
	case "year":
		result = matchesYear(card, cond.Value, cond.Operator)
	case "name":
		result = strings.Contains(strings.ToLower(card.Name), strings.ToLower(cond.Value))
	case "is":
		result = matchesIs(card, cond.Value)
	case "not":
		result = !matchesIs(card, cond.Value)
	default:
		// Default to name search
		result = strings.Contains(strings.ToLower(card.Name), strings.ToLower(cond.Value))
	}

	if cond.Negate {
		return !result
	}
	return result
}

func matchesType(card Card, value string) bool {
	// Match whole words only, not substrings
	// This prevents "cow" from matching "coward"
	value = strings.ToLower(value)
	typeLine := strings.ToLower(card.TypeLine)
	
	// Use word boundaries to match whole words
	// Split by common separators (space, dash, em dash) and check each word
	words := strings.FieldsFunc(typeLine, func(r rune) bool {
		return r == ' ' || r == '-' || r == '—' || r == '/' || r == ','
	})
	
	for _, word := range words {
		if word == value {
			return true
		}
	}
	
	return false
}

func matchesColor(card Card, value string) bool {
	value = strings.ToLower(value)
	colorMap := map[string]string{
		"w": "W", "white": "W",
		"u": "U", "blue": "U",
		"b": "B", "black": "B",
		"r": "R", "red": "R",
		"g": "G", "green": "G",
	}

	// Handle multicolor and colorless
	if value == "m" || value == "multicolor" {
		return len(card.Colors) > 1
	}
	if value == "c" || value == "colorless" {
		return len(card.Colors) == 0
	}

	// Check if card has the specified color
	if symbol, ok := colorMap[value]; ok {
		for _, c := range card.Colors {
			if c == symbol {
				return true
			}
		}
		// Also check mana cost
		return strings.Contains(card.ManaCost, symbol)
	}

	return false
}

func matchesColorIdentity(card Card, value string) bool {
	value = strings.ToLower(value)
	colorMap := map[string]string{
		"w": "W", "white": "W",
		"u": "U", "blue": "U",
		"b": "B", "black": "B",
		"r": "R", "red": "R",
		"g": "G", "green": "G",
	}

	if symbol, ok := colorMap[value]; ok {
		for _, c := range card.ColorIdentity {
			if c == symbol {
				return true
			}
		}
	}
	return false
}

func matchesKeyword(card Card, value string) bool {
	value = strings.ToLower(value)
	for _, kw := range card.Keywords {
		if strings.ToLower(kw) == value {
			return true
		}
	}
	// Also check oracle text
	return strings.Contains(strings.ToLower(card.OracleText), strings.ToLower(value))
}

func matchesMana(card Card, value string) bool {
	// Simple check - see if mana symbols appear in mana cost
	value = strings.ToUpper(value)
	// Remove braces for comparison
	value = strings.ReplaceAll(value, "{", "")
	value = strings.ReplaceAll(value, "}", "")
	return strings.Contains(card.ManaCost, value)
}

func matchesManaValue(card Card, value string, op string) bool {
	cardValue := card.CMC
	queryValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	switch op {
	case "=", "equals":
		return cardValue == queryValue
	case ">":
		return cardValue > queryValue
	case "<":
		return cardValue < queryValue
	case ">=":
		return cardValue >= queryValue
	case "<=":
		return cardValue <= queryValue
	case "!=":
		return cardValue != queryValue
	default:
		return cardValue == queryValue
	}
}

func matchesPower(card Card, value string, op string) bool {
	if card.Power == "" {
		return false
	}
	cardValue, err := parseNumericValue(card.Power)
	if err != nil {
		return false
	}
	queryValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	return compareValues(cardValue, queryValue, op)
}

func matchesToughness(card Card, value string, op string) bool {
	if card.Toughness == "" {
		return false
	}
	cardValue, err := parseNumericValue(card.Toughness)
	if err != nil {
		return false
	}
	queryValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	return compareValues(cardValue, queryValue, op)
}

func matchesPowerToughness(card Card, value string, op string) bool {
	pow, _ := parseNumericValue(card.Power)
	tou, _ := parseNumericValue(card.Toughness)
	cardValue := pow + tou
	queryValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	return compareValues(cardValue, queryValue, op)
}

func matchesLoyalty(card Card, value string, op string) bool {
	if card.Loyalty == "" {
		return false
	}
	cardValue, err := parseNumericValue(card.Loyalty)
	if err != nil {
		return false
	}
	queryValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	return compareValues(cardValue, queryValue, op)
}

func matchesRarity(card Card, value string) bool {
	return strings.ToLower(card.Rarity) == strings.ToLower(value)
}

func matchesSet(card Card, value string) bool {
	value = strings.ToUpper(value)
	return strings.ToUpper(card.SetCode) == value || strings.Contains(strings.ToUpper(card.Set), value)
}

func matchesYear(card Card, value string, op string) bool {
	queryValue, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	
	// Parse year from released_at date string (format: "2020-01-01")
	if card.ReleasedAt == "" {
		return false
	}
	
	// Extract year from date string (first 4 characters)
	if len(card.ReleasedAt) < 4 {
		return false
	}
	
	cardYear, err := strconv.Atoi(card.ReleasedAt[:4])
	if err != nil {
		return false
	}

	return compareValues(float64(cardYear), float64(queryValue), op)
}

func matchesIs(card Card, value string) bool {
	value = strings.ToLower(value)
	switch value {
	case "multicolor", "m":
		return len(card.Colors) > 1
	case "colorless", "c":
		return len(card.Colors) == 0
	case "spell":
		return !strings.Contains(strings.ToLower(card.TypeLine), "land") &&
			!strings.Contains(strings.ToLower(card.TypeLine), "creature") &&
			!strings.Contains(strings.ToLower(card.TypeLine), "enchantment") &&
			!strings.Contains(strings.ToLower(card.TypeLine), "artifact")
	case "permanent":
		return strings.Contains(strings.ToLower(card.TypeLine), "creature") ||
			strings.Contains(strings.ToLower(card.TypeLine), "land") ||
			strings.Contains(strings.ToLower(card.TypeLine), "enchantment") ||
			strings.Contains(strings.ToLower(card.TypeLine), "artifact") ||
			strings.Contains(strings.ToLower(card.TypeLine), "planeswalker")
	case "vanilla":
		return strings.Contains(strings.ToLower(card.TypeLine), "creature") && card.OracleText == ""
	case "reprint":
		// This would require tracking printings, simplified for now
		return false
	default:
		return false
	}
}

func parseNumericValue(s string) (float64, error) {
	// Handle * and other special values
	if s == "*" || s == "" {
		return 0, fmt.Errorf("cannot parse special value")
	}
	return strconv.ParseFloat(s, 64)
}

func compareValues(cardValue, queryValue float64, op string) bool {
	switch op {
	case "=", "equals":
		return cardValue == queryValue
	case ">":
		return cardValue > queryValue
	case "<":
		return cardValue < queryValue
	case ">=":
		return cardValue >= queryValue
	case "<=":
		return cardValue <= queryValue
	case "!=":
		return cardValue != queryValue
	default:
		return cardValue == queryValue
	}
}

