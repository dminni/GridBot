package indicators

import (
	"math"
	"github.com/dminni/gridbot/models"
)

func CalculateHistoricalVolatility(data []models.OHLCV, period int) float64 {
	n := len(data)
	if n < period+1 {
		return 0
	}

	returns := make([]float64, period)
	sum := 0.0
	for i := 0; i < period; i++ {
		idx := n - period + i
		returns[i] = math.Log(data[idx].Close / data[idx-1].Close)
		sum += returns[i]
	}

	mean := sum / float64(period)
	varianceSum := 0.0
	for _, r := range returns {
		varianceSum += math.Pow(r-mean, 2)
	}

	stdDev := math.Sqrt(varianceSum / float64(period-1))
	// Annualized HV = StdDev * sqrt(365) * 100
	return stdDev * math.Sqrt(365) * 100
}
