package models

type AnalysisRequest struct {
	Indicators []string `json:"indicators"`
}

type AnalysisResult struct {
	RankedAssets   []Asset `json:"ranked_assets"`
	DiscardedAssets []Asset `json:"discarded_assets"`
	Summary        string  `json:"summary"`
}

type SimulationResult struct {
	Period              int     `json:"period"`
	EntryPrice          float64 `json:"entry_price"`
	CurrentPrice        float64 `json:"current_price"`
	FinalCapital        float64 `json:"final_capital"`
	ROINet              float64 `json:"roi_net"`
	EstimatedOperations int     `json:"estimated_operations"`
	TotalFees           float64 `json:"total_fees"`
	BuyAndHoldROI       float64 `json:"buy_and_hold_roi"`
	History             []OHLCV `json:"history"`
	Trades              []Trade `json:"trades"`
}

type Trade struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
	Type      string  `json:"type"` // "buy" or "sell"
	GridLevel int     `json:"grid_level"`
}
