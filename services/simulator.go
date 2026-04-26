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

	// Simulate trades: stateful tracking of inventory per grid level
	// state[g] == true means we hold the base asset (ready to sell if price rises)
	// state[g] == false means we hold USDT (ready to buy if price drops)
	state := make([]bool, grids+1)
	levels := make([]float64, grids+1)

	for g := 0; g <= grids; g++ {
		levels[g] = lower + float64(g)*gridSize
		// If the level is above our entry price, we assume we bought market on day 0
		// and placed a sell limit order there.
		if levels[g] >= entryPrice {
			state[g] = true
		} else {
			// If below, we just placed a buy limit order.
			state[g] = false
		}
	}

	var trades []models.Trade
	for i := 1; i < len(ohlcv); i++ {
		currLow := ohlcv[i].Low
		currHigh := ohlcv[i].High

		for g := 0; g <= grids; g++ {
			level := levels[g]

			// If we are ready to buy, and price drops to/below level
			if !state[g] && currLow <= level {
				trades = append(trades, models.Trade{
					Timestamp: ohlcv[i].Timestamp,
					Price:     level,
					Type:      "buy",
					GridLevel: g,
				})
				state[g] = true // we now hold it, ready to sell
			} else if state[g] && currHigh >= level {
				// If we hold it, and price rises to/above level
				trades = append(trades, models.Trade{
					Timestamp: ohlcv[i].Timestamp,
					Price:     level,
					Type:      "sell",
					GridLevel: g,
				})
				state[g] = false // we sold it, ready to buy again
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
