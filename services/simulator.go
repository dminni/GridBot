package services

import (
	"sort"
	"github.com/dminni/gridbot/models"
)

type Simulator struct {
	BingX *BingXClient
}

func NewSimulator(bx *BingXClient) *Simulator {
	return &Simulator{BingX: bx}
}

func (s *Simulator) Simulate(symbol string, periodDays int, config models.GridConfig) (*models.SimulationResult, error) {
	// Fetch daily candles with extra buffer
	ohlcv, err := s.BingX.GetOHLCV(symbol, "1d", periodDays+10)
	if err != nil || len(ohlcv) == 0 {
		return nil, err
	}

	// Sort ascending by timestamp (oldest first)
	sort.Slice(ohlcv, func(i, j int) bool {
		return ohlcv[i].Timestamp < ohlcv[j].Timestamp
	})

	// Trim to exactly periodDays from the end (most recent)
	if len(ohlcv) > periodDays {
		ohlcv = ohlcv[len(ohlcv)-periodDays:]
	}

	entryPrice := ohlcv[0].Close
	currentPrice := ohlcv[len(ohlcv)-1].Close

	// Grid levels
	lower := config.LowerRange
	upper := config.UpperRange
	grids := config.RecommendedGrids
	if grids < 2 {
		grids = 5
	}
	gridSize := (upper - lower) / float64(grids)
	if gridSize <= 0 {
		gridSize = currentPrice * 0.005
	}

	// Simulate trades: detect when price crosses grid levels
	var trades []models.Trade
	for i := 1; i < len(ohlcv); i++ {
		prevClose := ohlcv[i-1].Close
		currLow := ohlcv[i].Low
		currHigh := ohlcv[i].High

		// Check each grid level for crossings within this candle
		for g := 0; g <= grids; g++ {
			level := lower + float64(g)*gridSize

			// Price dropped through a level (BUY signal)
			if prevClose > level && currLow <= level {
				trades = append(trades, models.Trade{
					Timestamp: ohlcv[i].Timestamp,
					Price:     level,
					Type:      "buy",
					GridLevel: g,
				})
			}

			// Price rose through a level (SELL signal)
			if prevClose < level && currHigh >= level {
				trades = append(trades, models.Trade{
					Timestamp: ohlcv[i].Timestamp,
					Price:     level,
					Type:      "sell",
					GridLevel: g,
				})
			}
		}
	}

	initialCapital := 1000.0
	numOps := len(trades)
	profitPerOp := (initialCapital / float64(grids)) * (config.ProfitPerGridNet / 100)
	// Only completed round-trips generate profit (buy+sell pairs)
	completedPairs := numOps / 2
	profitNet := float64(completedPairs) * profitPerOp

	finalCapital := initialCapital + profitNet
	roiNet := (profitNet / initialCapital) * 100
	totalFees := float64(numOps) * (initialCapital / float64(grids)) * 0.002

	buyAndHoldROI := (currentPrice - entryPrice) / entryPrice * 100

	return &models.SimulationResult{
		Period:              periodDays,
		EntryPrice:          entryPrice,
		CurrentPrice:        currentPrice,
		FinalCapital:        finalCapital,
		ROINet:              roiNet,
		EstimatedOperations: completedPairs,
		TotalFees:           totalFees,
		BuyAndHoldROI:       buyAndHoldROI,
		History:             ohlcv,
		Trades:              trades,
	}, nil
}
