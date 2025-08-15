package services

import (
	"fmt"
	"strconv"
	"strings"
)

// ConvertCategoryIDsToString converts a slice of category IDs to a comma-separated string
func ConvertCategoryIDsToString(categoryIDs []int) string {
	if len(categoryIDs) == 0 {
		return ""
	}

	strIDs := make([]string, len(categoryIDs))
	for i, id := range categoryIDs {
		strIDs[i] = strconv.Itoa(id)
	}

	return strings.Join(strIDs, ",")
}

// ParseCategoryIDsFromString parses a comma-separated string of category IDs into a slice of integers
func ParseCategoryIDsFromString(categoriesStr string) ([]int, error) {
	if categoriesStr == "" {
		return []int{}, nil
	}

	strIDs := strings.Split(categoriesStr, ",")
	categoryIDs := make([]int, len(strIDs))

	for i, strID := range strIDs {
		id, err := strconv.Atoi(strings.TrimSpace(strID))
		if err != nil {
			return nil, fmt.Errorf("invalid category ID: %s", strID)
		}
		categoryIDs[i] = id
	}

	return categoryIDs, nil
}
