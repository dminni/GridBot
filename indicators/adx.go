package indicators

import (
	"math"
	"github.com/dminni/gridbot/models"
)

type ADXResult struct {
	ADX float64
	PDI float64 // +DI
	MDI float64 // -DI
}

func CalculateADX(data []models.OHLCV, period int) *ADXResult {
	if len(data) < 2*period {
		return nil
	}

	n := len(data)
	tr := make([]float64, n)
	pDM := make([]float64, n)
	mDM := make([]float64, n)

	for i := 1; i < n; i++ {
		h_l := data[i].High - data[i].Low
		h_pc := math.Abs(data[i].High - data[i-1].Close)
		l_pc := math.Abs(data[i].Low - data[i-1].Close)
		tr[i] = math.Max(h_l, math.Max(h_pc, l_pc))

		upMove := data[i].High - data[i-1].High
		downMove := data[i-1].Low - data[i].Low

		if upMove > downMove && upMove > 0 {
			pDM[i] = upMove
		} else {
			pDM[i] = 0
		}

		if downMove > upMove && downMove > 0 {
			mDM[i] = downMove
		} else {
			mDM[i] = 0
		}
	}

	atr := make([]float64, n)
	pDI_smooth := make([]float64, n)
	mDI_smooth := make([]float64, n)

	var trSum, pDMSum, mDMSum float64
	for i := 1; i <= period; i++ {
		trSum += tr[i]
		pDMSum += pDM[i]
		mDMSum += mDM[i]
	}

	atr[period] = trSum
	pDI_smooth[period] = pDMSum
	mDI_smooth[period] = mDMSum

	for i := period + 1; i < n; i++ {
		atr[i] = atr[i-1] - (atr[i-1] / float64(period)) + tr[i]
		pDI_smooth[i] = pDI_smooth[i-1] - (pDI_smooth[i-1] / float64(period)) + pDM[i]
		mDI_smooth[i] = mDI_smooth[i-1] - (mDI_smooth[i-1] / float64(period)) + mDM[i]
	}

	dx := make([]float64, n)
	for i := period; i < n; i++ {
		plusDI := 100 * (pDI_smooth[i] / atr[i])
		minusDI := 100 * (mDI_smooth[i] / atr[i])
		
		div := math.Abs(plusDI + minusDI)
		if div == 0 {
			dx[i] = 0
		} else {
			dx[i] = 100 * math.Abs(plusDI-minusDI) / div
		}
	}

	var dxSum float64
	for i := period; i < 2*period; i++ {
		dxSum += dx[i]
	}
	
	adx := make([]float64, n)
	adx[2*period-1] = dxSum / float64(period)

	for i := 2 * period; i < n; i++ {
		adx[i] = (adx[i-1]*float64(period-1) + dx[i]) / float64(period)
	}

	lastIdx := n - 1
	return &ADXResult{
		ADX: adx[lastIdx],
		PDI: 100 * (pDI_smooth[lastIdx] / atr[lastIdx]),
		MDI: 100 * (mDI_smooth[lastIdx] / atr[lastIdx]),
	}
}
