package indicators

import (
	"github.com/dminni/gridbot/models"
)

// CalculateRSI calculates the Relative Strength Index for a given period (usually 14)
func CalculateRSI(data []models.OHLCV, period int) []float64 {
	if len(data) < period+1 {
		return nil
	}

	rsi := make([]float64, len(data))
	gains := make([]float64, len(data))
	losses := make([]float64, len(data))

	for i := 1; i < len(data); i++ {
		diff := data[i].Close - data[i-1].Close
		if diff > 0 {
			gains[i] = diff
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = -diff
		}
	}

	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - (100 / (1 + rs))
	}

	// Wilder's Smoothing
	for i := period + 1; i < len(data); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi
}

func GetLastRSI(data []models.OHLCV, period int) float64 {
	rsi := CalculateRSI(data, period)
	if len(rsi) == 0 {
		return 50 // Neutral default
	}
	return rsi[len(rsi)-1]
}
