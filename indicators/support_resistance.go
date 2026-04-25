package indicators

import (
	"sort"
	"github.com/dminni/gridbot/models"
)

type SRResult struct {
	Supports    []float64
	Resistances []float64
	CurrentS    float64
	CurrentR    float64
}

func CalculateSupportResistance(data []models.OHLCV) *SRResult {
	n := len(data)
	if n < 5 {
		return nil
	}

	var highs []float64
	var lows []float64

	// Identify local peaks/troughs
	for i := 2; i < n-2; i++ {
		if data[i].High > data[i-1].High && data[i].High > data[i-2].High &&
			data[i].High > data[i+1].High && data[i].High > data[i+2].High {
			highs = append(highs, data[i].High)
		}
		if data[i].Low < data[i-1].Low && data[i].Low < data[i-2].Low &&
			data[i].Low < data[i+1].Low && data[i].Low < data[i+2].Low {
			lows = append(lows, data[i].Low)
		}
	}

	sort.Float64s(highs)
	sort.Float64s(lows)

	currentPrice := data[n-1].Close
	
	var currentS, currentR float64
	
	// Find closest support below price
	for i := len(lows) - 1; i >= 0; i-- {
		if lows[i] < currentPrice {
			currentS = lows[i]
			break
		}
	}
	
	// Find closest resistance above price
	for i := 0; i < len(highs); i++ {
		if highs[i] > currentPrice {
			currentR = highs[i]
			break
		}
	}

	// Fallback if none found
	if currentS == 0 && n > 0 {
		currentS = data[n-1].Low * 0.95
	}
	if currentR == 0 && n > 0 {
		currentR = data[n-1].High * 1.05
	}

	return &SRResult{
		Supports:    lows,
		Resistances: highs,
		CurrentS:    currentS,
		CurrentR:    currentR,
	}
}
