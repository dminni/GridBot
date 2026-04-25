package models

import "time"

type Asset struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	CurrentPrice float64   `json:"current_price"`
	PriceChange  float64   `json:"price_change_24h"`
	MarketCap    float64   `json:"market_cap"`
	Rank         int       `json:"rank"`
	Analysis     *Analysis `json:"analysis,omitempty"`
}

type Analysis struct {
	Score             int                `json:"score"`
	Indicators        MapIndicators      `json:"indicators"`
	Justification     string             `json:"justification"`
	GridConfig        GridConfig         `json:"grid_config"`
	Status            string             `json:"status"` // "Excellent", "Good", "Regular", "Not Recommended"
	LastUpdate        time.Time          `json:"last_update"`
}

type MapIndicators map[string]float64

type GridConfig struct {
	LowerRange          float64             `json:"lower_range"`
	UpperRange          float64             `json:"upper_range"`
	Amplitude           float64             `json:"amplitude_pct"`
	RecommendedGrids    int                 `json:"recommended_grids"`
	ProfitPerGridNet    float64             `json:"profit_per_grid_net"`
	ProfitPerGridGross  float64             `json:"profit_per_grid_gross"`
	AlternativeConfigs  []AlternativeConfig `json:"alternative_configs"`
	Warning             string              `json:"warning,omitempty"`
}

type AlternativeConfig struct {
	Type          string  `json:"type"` // "Conservative", "Recommended", "Aggressive"
	Grids         int     `json:"grids"`
	ProfitPerGrid float64 `json:"profit_per_grid"`
	Observation   string  `json:"observation"`
}

type OHLCV struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}
