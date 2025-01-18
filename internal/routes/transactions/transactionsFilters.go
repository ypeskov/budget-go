package transactions

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"ypeskov/budget-go/internal/routes/routeErrors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type TransactionFilters struct {
	PerPage          int
	Page             int
	Currencies       []string
	FromDate         time.Time
	ToDate           time.Time
	AccountIds       []int
	TransactionTypes []string
	CategoryIds      []int
}

func parseTransactionFilters(c echo.Context) (*TransactionFilters, error) {
	perPage, err := getPerPage(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	page, err := getPage(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	currencies, err := getCurrencies(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	fromDate, err := getFromDate(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	toDate, err := getToDate(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	accountIds, err := getAccountIds(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	transactionTypes, err := getTransactionTypes(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}
	categoryIds, err := getCategoryIds(c)
	if err != nil {
		return nil, &routeErrors.InvalidRequestError{Message: err.Error()}
	}

	return &TransactionFilters{
		PerPage:          perPage,
		Page:             page,
		Currencies:       currencies,
		FromDate:         fromDate,
		ToDate:           toDate,
		AccountIds:       accountIds,
		TransactionTypes: transactionTypes,
		CategoryIds:      categoryIds,
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

func getCategoryIds(c echo.Context) ([]int, error) {
	return getQueryParamAsIntSlice(c, "categories")
}

func getTransactionTypes(c echo.Context) ([]string, error) {
	return getQueryParamAsStringSlice(c, "types")
}

func getQueryParamAsInt(c echo.Context, paramName string, defaultValue int) (int, error) {
	paramStr := c.QueryParam(paramName)
	if paramStr == "" {
		return defaultValue, nil // Return the default value if the parameter is not provided
	}
	val, err := strconv.Atoi(paramStr)
	if err != nil {
		return 0, errors.New("invalid value for " + paramName)
	}
	return val, nil
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
			log.Warn("invalid value for " + paramName + ": [" + str + "] skipping")
			continue
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
	parsedTime, err := time.Parse(time.DateOnly, paramStr)
	if err != nil {
		return time.Time{}, errors.New("invalid value for " + paramName)
	}
	return parsedTime, nil
}
