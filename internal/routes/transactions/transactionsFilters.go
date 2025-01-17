package transactions

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type TransactionFilters struct {
	PerPage          int
	Page             int
	Currencies       []string
	FromDate         time.Time
	ToDate           time.Time
	AccountIds       []int
	TransactionTypes []string
}

func parseTransactionFilters(c echo.Context) (*TransactionFilters, error) {
	perPage, err := getPerPage(c)
	if err != nil {
		return nil, err
	}
	page, err := getPage(c)
	if err != nil {
		return nil, err
	}
	currencies, err := getCurrencies(c)
	if err != nil {
		return nil, err
	}
	fromDate, err := getFromDate(c)
	if err != nil {
		return nil, err
	}
	toDate, err := getToDate(c)
	if err != nil {
		return nil, err
	}
	accountIds, err := getAccountIds(c)
	if err != nil {
		return nil, err
	}
	transactionTypes, err := getTransactionTypes(c)
	if err != nil {
		return nil, err
	}

	return &TransactionFilters{
		PerPage:          perPage,
		Page:             page,
		Currencies:       currencies,
		FromDate:         fromDate,
		ToDate:           toDate,
		AccountIds:       accountIds,
		TransactionTypes: transactionTypes,
	}, nil
}

func getPerPage(c echo.Context) (int, error) {
	perPage, err := getQueryParamAsInt(c, "per_page", 10)
	if err != nil {
		return 0, err
	}
	if perPage < 1 {
		return 0, errors.New("per_page must be greater than 0")
	}
	return perPage, nil
}

func getPage(c echo.Context) (int, error) {
	return getQueryParamAsInt(c, "page", 1)
}

func getCurrencies(c echo.Context) ([]string, error) {
	return getQueryParamAsStringSlice(c, "currencies")
}

func getFromDate(c echo.Context) (time.Time, error) {
	return getQueryParamAsTime(c, "from_date")
}

func getToDate(c echo.Context) (time.Time, error) {
	return getQueryParamAsTime(c, "to_date")
}

func getAccountIds(c echo.Context) ([]int, error) {
	return getQueryParamAsIntSlice(c, "accounts")
}

func getTransactionTypes(c echo.Context) ([]string, error) {
	return getQueryParamAsStringSlice(c, "types")
}

func getQueryParamAsInt(c echo.Context, paramName string, defaultValue int) (int, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return defaultValue, nil // Return the default value if the parameter is not provided
	}
	return strconv.Atoi(paramStr) // Convert string to integer
}

func getQueryParamAsIntSlice(c echo.Context, paramName string) ([]int, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return []int{}, nil
	}
	strSlice := strings.Split(paramStr, ",")
	var intSlice []int
	for _, str := range strSlice {
		val, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		intSlice = append(intSlice, val)
	}
	return intSlice, nil
}

func getQueryParamAsStringSlice(c echo.Context, paramName string) ([]string, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return []string{}, nil
	}
	return strings.Split(paramStr, ","), nil
}

func getQueryParamAsTime(c echo.Context, paramName string) (time.Time, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.DateOnly, paramStr)
}
