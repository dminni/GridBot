package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/dminni/gridbot/models"
)

type CoinGeckoClient struct {
	BaseURL string
	APIKey  string
}

func NewCoinGeckoClient(baseURL, apiKey string) *CoinGeckoClient {
	return &CoinGeckoClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func (c *CoinGeckoClient) GetTopAssets() ([]models.Asset, error) {
	stablecoins := map[string]bool{
		"usdt": true, "usdc": true, "busd": true, "dai": true, "tusd": true,
		"frax": true, "usdd": true, "usdp": true, "gusd": true, "lusd": true,
		"susd": true, "pyusd": true, "fdusd": true, "eurs": true, "xaut": true,
		"paxg": true, "ustc": true, "usde": true, "usdt.e": true, "usdc.e": true,
	}

	var allAssets []models.Asset

	for page := 1; page <= 2; page++ {
		url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=100&page=%d", c.BaseURL, page)
		if c.APIKey != "" {
			url += "&x_cg_demo_api_key=" + c.APIKey
		}

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("coingecko api error: %s", resp.Status)
		}

		var rawAssets []struct {
			ID           string  `json:"id"`
			Symbol       string  `json:"symbol"`
			Name         string  `json:"name"`
			CurrentPrice float64 `json:"current_price"`
			PriceChange  float64 `json:"price_change_percentage_24h"`
			MarketCap    float64 `json:"market_cap"`
			Rank         int     `json:"market_cap_rank"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&rawAssets); err != nil {
			return nil, err
		}

		for _, a := range rawAssets {
			symbol := strings.ToLower(a.Symbol)
			if stablecoins[symbol] {
				continue
			}
			allAssets = append(allAssets, models.Asset{
				ID:           a.ID,
				Symbol:       strings.ToUpper(a.Symbol),
				Name:         a.Name,
				CurrentPrice: a.CurrentPrice,
				PriceChange:  a.PriceChange,
				MarketCap:    a.MarketCap,
				Rank:         a.Rank,
			})
		}
	}

	return allAssets, nil
}
