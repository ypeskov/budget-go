package transactions

import (
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func getPerPage(c echo.Context) (int, error) {
	perPageStr := c.QueryParam("per_page")
	if perPageStr == "" {
		return 10, nil
	}
	return strconv.Atoi(perPageStr)
}

func getPage(c echo.Context) (int, error) {
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		return 1, nil
	}
	return strconv.Atoi(pageStr)
}

func getCurrencies(c echo.Context) ([]string, error) {
	currenciesStr := c.QueryParam("currencies")
	if currenciesStr == "" {
		return []string{}, nil
	}
	return strings.Split(currenciesStr, ","), nil
}

func getFromDate(c echo.Context) (time.Time, error) {
	fromDateStr := c.QueryParam("from_date")
	if fromDateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.DateOnly, fromDateStr)
}

func getToDate(c echo.Context) (time.Time, error) {
	toDateStr := c.QueryParam("to_date")
	if toDateStr == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.DateOnly, toDateStr)
}

func getAccountIds(c echo.Context) ([]int, error) {
	accountIdsStr := c.QueryParam("accounts")
	if accountIdsStr == "" {
		return []int{}, nil
	}
	accountIds := strings.Split(accountIdsStr, ",")
	var ids []int
	for _, idStr := range accountIds {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return []int{}, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func getTypes(c echo.Context) ([]string, error) {
	typesStr := c.QueryParam("types")

	if typesStr == "" {
		return []string{}, nil
	}
	return strings.Split(typesStr, ","), nil
}
