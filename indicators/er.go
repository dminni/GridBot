package indicators

import (
	"math"
	"github.com/dminni/gridbot/models"
)

func CalculateER(data []models.OHLCV, period int) float64 {
	n := len(data)
	if n < period+1 {
		return 0.5
	}

	change := math.Abs(data[n-1].Close - data[n-period-1].Close)
	
	volatility := 0.0
	for i := n - period; i < n; i++ {
		volatility += math.Abs(data[i].Close - data[i-1].Close)
	}

	if volatility == 0 {
		return 0
	}

	return change / volatility
}
