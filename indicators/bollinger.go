package indicators

import (
	"math"
	"github.com/dminni/gridbot/models"
)

type BBResult struct {
	Upper     float64
	Middle    float64
	Lower     float64
	PercentB  float64
	Bandwidth float64
}

func CalculateBollingerBands(data []models.OHLCV, period int, stdDev float64) *BBResult {
	n := len(data)
	if n < period {
		return nil
	}

	sum := 0.0
	for i := n - period; i < n; i++ {
		sum += data[i].Close
	}
	sma := sum / float64(period)

	variance := 0.0
	for i := n - period; i < n; i++ {
		variance += math.Pow(data[i].Close-sma, 2)
	}
	std := math.Sqrt(variance / float64(period))

	upper := sma + (stdDev * std)
	lower := sma - (stdDev * std)
	
	currentPrice := data[n-1].Close
	percentB := (currentPrice - lower) / (upper - lower)
	bandwidth := (upper - lower) / sma * 100

	return &BBResult{
		Upper:     upper,
		Middle:    sma,
		Lower:     lower,
		PercentB:  percentB,
		Bandwidth: bandwidth,
	}
}
