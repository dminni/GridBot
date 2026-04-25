package indicators

import (
	"math"
	"github.com/dminni/gridbot/models"
)

func CalculateCHOP(data []models.OHLCV, period int) float64 {
	n := len(data)
	if n < period {
		return 50.0
	}

	sumATR1 := 0.0
	maxHigh := data[n-period].High
	minLow := data[n-period].Low

	// We need TR(1) for each day in the period
	for i := n - period + 1; i < n; i++ {
		h_l := data[i].High - data[i].Low
		h_pc := math.Abs(data[i].High - data[i-1].Close)
		l_pc := math.Abs(data[i].Low - data[i-1].Close)
		tr := math.Max(h_l, math.Max(h_pc, l_pc))
		sumATR1 += tr

		if data[i].High > maxHigh {
			maxHigh = data[i].High
		}
		if data[i].Low < minLow {
			minLow = data[i].Low
		}
	}

	range_total := maxHigh - minLow
	if range_total == 0 {
		return 100.0
	}

	chop := 100 * math.Log10(sumATR1/range_total) / math.Log10(float64(period))
	return chop
}
