package utils

import (
	"fmt"
	"strconv"

	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

// ParseCategoryIdFromInterface parses CategoryID from any (can be string or int)
func ParseCategoryIdFromInterface(value any) (*int, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Error parsing categoryId:", err)
			return nil, err
		}
		return &id, nil
	case int:
		id := int(v)
		return &id, nil
	case float64:
		id := int(v)
		return &id, nil
	default:
		return nil, fmt.Errorf("invalid categoryId type: %T", v)
	}
}

// ParseAmountFromInterface parses Amount from any (can be string or float64)
func ParseAmountFromInterface(value any) (decimal.Decimal, error) {
	if value == nil {
		return decimal.Zero, nil
	}

	switch v := value.(type) {
	case string:
		amount, err := decimal.NewFromString(v)
		if err != nil {
			log.Error("Error parsing amount:", err)
			return decimal.Zero, err
		}
		return amount, nil
	case float64:
		return decimal.NewFromFloat(v), nil
	case int:
		return decimal.NewFromInt(int64(v)), nil
	default:
		return decimal.Zero, fmt.Errorf("invalid amount type: %T", v)
	}
}

// ParseIDFromInterface parses ID from interface{} (can be string or int)
func ParseIDFromInterface(value interface{}) (int, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return 0, nil
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			log.Error("Error parsing id:", err)
			return 0, err
		}
		return id, nil
	case int:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid id type: %T", v)
	}
}
