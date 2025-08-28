package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"ypeskov/budget-go/internal/config"
	"ypeskov/budget-go/internal/logger"
)

type CurrencyBeaconService interface {
	GetCurrencyRates(date string) (map[string]interface{}, error)
}

type CurrencyBeaconServiceInstance struct {
	config *config.Config
}

type CurrencyBeaconResponse struct {
	Date  string                 `json:"date"`
	Base  string                 `json:"base"`
	Rates map[string]interface{} `json:"rates"`
}

var (
	currencyBeaconInstance *CurrencyBeaconServiceInstance
	currencyBeaconOnce     sync.Once
)

func NewCurrencyBeaconService(cfg *config.Config) CurrencyBeaconService {
	currencyBeaconOnce.Do(func() {
		logger.Debug("Creating CurrencyBeaconService instance")
		currencyBeaconInstance = &CurrencyBeaconServiceInstance{config: cfg}
	})

	return currencyBeaconInstance
}

func (c *CurrencyBeaconServiceInstance) GetCurrencyRates(date string) (map[string]interface{}, error) {
	if c.config.CurrencyBeaconAPIKey == "" {
		return nil, fmt.Errorf("CurrencyBeacon API key is not configured")
	}

	url := fmt.Sprintf("%s/%s/historical?api_key=%s&date=%s",
		c.config.CurrencyBeaconAPIURL,
		c.config.CurrencyBeaconAPIVersion,
		c.config.CurrencyBeaconAPIKey,
		date)

	logger.Info("Fetching exchange rates from CurrencyBeacon for date", "date", date)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to make request to CurrencyBeacon", "error", err)
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("CurrencyBeacon API error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response CurrencyBeaconResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error("Failed to decode CurrencyBeacon response", "error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := map[string]interface{}{
		"rates":              response.Rates,
		"actual_date":        response.Date,
		"base_currency_code": response.Base,
		"service_name":       "CurrencyBeacon",
	}

	logger.Info("Successfully fetched exchange rates", "date", date, "base", response.Base, "rateCount", len(response.Rates))

	return result, nil
}
