package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"ypeskov/budget-go/internal/config"
)

type CurrencyBeaconService struct {
	config *config.Config
}

type CurrencyBeaconResponse struct {
	Date  string                 `json:"date"`
	Base  string                 `json:"base"`
	Rates map[string]interface{} `json:"rates"`
}

func NewCurrencyBeaconService(cfg *config.Config) *CurrencyBeaconService {
	return &CurrencyBeaconService{config: cfg}
}

func (c *CurrencyBeaconService) GetCurrencyRates(date string) (map[string]interface{}, error) {
	if c.config.CurrencyBeaconAPIKey == "" {
		return nil, fmt.Errorf("CurrencyBeacon API key is not configured")
	}

	url := fmt.Sprintf("%s/%s/historical?api_key=%s&date=%s",
		c.config.CurrencyBeaconAPIURL,
		c.config.CurrencyBeaconAPIVersion,
		c.config.CurrencyBeaconAPIKey,
		date)

	log.Infof("Fetching exchange rates from CurrencyBeacon for date: %s", date)

	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Failed to make request to CurrencyBeacon: %v", err)
		return nil, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("CurrencyBeacon API error: status %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response CurrencyBeaconResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Errorf("Failed to decode CurrencyBeacon response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := map[string]interface{}{
		"rates":                response.Rates,
		"actual_date":          response.Date,
		"base_currency_code":   response.Base,
		"service_name":         "CurrencyBeacon",
	}

	log.Infof("Successfully fetched exchange rates for %s (base: %s, %d rates)",
		response.Date, response.Base, len(response.Rates))

	return result, nil
}