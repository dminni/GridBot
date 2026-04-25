package services

import (
	"math"
	"github.com/dminni/gridbot/models"
)

type Simulator struct {
	BingX *BingXClient
}

func NewSimulator(bx *BingXClient) *Simulator {
	return &Simulator{BingX: bx}
}

func (s *Simulator) Simulate(symbol string, periodDays int, config models.GridConfig) (*models.SimulationResult, error) {
	ohlcv, err := s.BingX.GetOHLCV(symbol, "1h", periodDays*24)
	if err != nil || len(ohlcv) == 0 {
		return nil, err
	}

	entryPrice := ohlcv[0].Close
	currentPrice := ohlcv[len(ohlcv)-1].Close
	
	// Simplified simulation logic
	// Count how many times price crosses the average grid level
	gridSize := (config.UpperRange - config.LowerRange) / float64(config.RecommendedGrids)
	
	crosses := 0
	for i := 1; i < len(ohlcv); i++ {
		// Price crossed a grid line (simplified)
		diff := math.Abs(ohlcv[i].Close - ohlcv[i-1].Close)
		if diff > gridSize {
			crosses += int(diff / gridSize)
		}
	}

	initialCapital := 1000.0
	// Profit per cross = (Capital/Grids) * (ProfitPerGridNet/100)
	// Simplified: total ops * profit per grid
	numOps := crosses
	profitNet := float64(numOps) * (initialCapital / float64(config.RecommendedGrids)) * (config.ProfitPerGridNet / 100)
	
	finalCapital := initialCapital + profitNet
	roiNet := (profitNet / initialCapital) * 100
	totalFees := float64(numOps) * (initialCapital / float64(config.RecommendedGrids)) * 0.002 // 0.2% fee

	buyAndHoldROI := (currentPrice - entryPrice) / entryPrice * 100

	return &models.SimulationResult{
		Period:              periodDays,
		EntryPrice:          entryPrice,
		CurrentPrice:        currentPrice,
		FinalCapital:        finalCapital,
		ROINet:              roiNet,
		EstimatedOperations: numOps,
		TotalFees:           totalFees,
		BuyAndHoldROI:       buyAndHoldROI,
		History:             ohlcv,
	}, nil
}
