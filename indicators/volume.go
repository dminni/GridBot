package indicators

import (
	"github.com/dminni/gridbot/models"
)

func CalculateRelativeVolume(data []models.OHLCV) float64 {
	n := len(data)
	if n < 20 {
		return 1.0
	}

	sum20 := 0.0
	for i := n - 20; i < n; i++ {
		sum20 += data[i].Volume
	}
	avg20 := sum20 / 20.0

	sum5 := 0.0
	for i := n - 5; i < n; i++ {
		sum5 += data[i].Volume
	}
	avg5 := sum5 / 5.0

	if avg20 == 0 {
		return 1.0
	}

	return avg5 / avg20
}
