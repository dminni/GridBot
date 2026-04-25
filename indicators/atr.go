package indicators

import (
	"math"
	"github.com/dminni/gridbot/models"
)

// CalculateATR calculates the Average True Range for a given period (usually 14)
func CalculateATR(data []models.OHLCV, period int) []float64 {
	if len(data) < period+1 {
		return nil
	}

	tr := make([]float64, len(data))
	// TR = max(High-Low, |High-PrevClose|, |Low-PrevClose|)
	for i := 1; i < len(data); i++ {
		h_l := data[i].High - data[i].Low
		h_pc := math.Abs(data[i].High - data[i-1].Close)
		l_pc := math.Abs(data[i].Low - data[i-1].Close)
		tr[i] = math.Max(h_l, math.Max(h_pc, l_pc))
	}

	atr := make([]float64, len(data))
	// First ATR is SMA of TR
	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += tr[i]
	}
	atr[period] = sum / float64(period)

	// Wilder's Smoothing (similar to EMA but different smoothing factor)
	// ATR_i = (ATR_{i-1} * (n-1) + TR_i) / n
	for i := period + 1; i < len(data); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[i]) / float64(period)
	}

	return atr
}

// CalculateRelativeATR returns ATR/CurrentPrice * 100
func CalculateRelativeATR(data []models.OHLCV, period int) float64 {
	atr := CalculateATR(data, period)
	if len(atr) == 0 {
		return 0
	}
	currentPrice := data[len(data)-1].Close
	return (atr[len(atr)-1] / currentPrice) * 100
}
