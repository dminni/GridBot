package services

import (
	"fmt"
	"math"
	"time"
	"github.com/dminni/gridbot/indicators"
	"github.com/dminni/gridbot/models"
)

type Analyzer struct {
	BingX *BingXClient
}

func NewAnalyzer(bx *BingXClient) *Analyzer {
	return &Analyzer{BingX: bx}
}

func (az *Analyzer) AnalyzeAsset(asset models.Asset, selectedIndicators []string) models.Asset {
	ohlcv, err := az.BingX.GetOHLCV(asset.Symbol, "1d", 200)
	if err != nil || len(ohlcv) < 50 {
		return asset
	}

	analysis := &models.Analysis{
		Indicators: make(models.MapIndicators),
		LastUpdate: time.Now(),
	}

	var score float64
	var totalWeight float64

	// Weights mapping (parametric)
	weights := map[string]float64{
		"ATR": 15, "ADX": 25, "RSI": 15, "BB": 10, "SR": 15, "HV": 10, "RVOL": 5, "CHOP": 10, "ER": 10,
	}

	selectedMap := make(map[string]bool)
	for _, id := range selectedIndicators {
		selectedMap[id] = true
	}

	// Calculate indicators and partial scores
	if selectedMap["ATR"] {
		relATR := indicators.CalculateRelativeATR(ohlcv, 14)
		analysis.Indicators["ATR"] = relATR
		// Ideal 3-15%
		var s float64
		if relATR >= 3 && relATR <= 15 {
			s = 100
		} else if relATR < 3 {
			s = (relATR / 3) * 100
		} else {
			s = math.Max(0, 100-(relATR-15)*10)
		}
		score += s * weights["ATR"]
		totalWeight += weights["ATR"]
	}

	if selectedMap["ADX"] {
		adxRes := indicators.CalculateADX(ohlcv, 14)
		if adxRes != nil {
			analysis.Indicators["ADX"] = adxRes.ADX
			var s float64
			if adxRes.ADX < 20 {
				s = 100
			} else if adxRes.ADX < 25 {
				s = 70
			} else {
				s = math.Max(0, 100-(adxRes.ADX-20)*4)
			}
			score += s * weights["ADX"]
			totalWeight += weights["ADX"]
		}
	}

	if selectedMap["RSI"] {
		rsi := indicators.GetLastRSI(ohlcv, 14)
		analysis.Indicators["RSI"] = rsi
		var s float64
		if rsi >= 40 && rsi <= 60 {
			s = 100
		} else if (rsi >= 30 && rsi < 40) || (rsi > 60 && rsi <= 70) {
			s = 60
		} else {
			s = 30
		}
		score += s * weights["RSI"]
		totalWeight += weights["RSI"]
	}

	if selectedMap["BB"] {
		bb := indicators.CalculateBollingerBands(ohlcv, 20, 2)
		if bb != nil {
			analysis.Indicators["BB_BW"] = bb.Bandwidth
			analysis.Indicators["BB_PctB"] = bb.PercentB
			var s float64
			if bb.PercentB >= 0.2 && bb.PercentB <= 0.8 {
				s += 50
			}
			if bb.Bandwidth < 30 { // Example compression
				s += 50
			}
			score += s * weights["BB"]
			totalWeight += weights["BB"]
		}
	}

	sr := indicators.CalculateSupportResistance(ohlcv)
	if selectedMap["SR"] && sr != nil {
		analysis.Indicators["SR_S"] = sr.CurrentS
		analysis.Indicators["SR_R"] = sr.CurrentR
		// Score based on range amplitude
		amp := (sr.CurrentR - sr.CurrentS) / sr.CurrentS * 100
		var s float64
		if amp > 5 && amp < 25 {
			s = 100
		} else {
			s = 50
		}
		score += s * weights["SR"]
		totalWeight += weights["SR"]
	}

	if selectedMap["HV"] {
		hv := indicators.CalculateHistoricalVolatility(ohlcv, 30)
		analysis.Indicators["HV"] = hv
		var s float64
		if hv >= 50 && hv <= 100 {
			s = 100
		} else if hv < 50 {
			s = 60
		} else {
			s = 40
		}
		score += s * weights["HV"]
		totalWeight += weights["HV"]
	}

	if selectedMap["RVOL"] {
		rvol := indicators.CalculateRelativeVolume(ohlcv)
		analysis.Indicators["RVOL"] = rvol
		var s float64
		if rvol >= 0.5 && rvol <= 1.5 {
			s = 100
		} else {
			s = 40
		}
		score += s * weights["RVOL"]
		totalWeight += weights["RVOL"]
	}

	if selectedMap["CHOP"] {
		chop := indicators.CalculateCHOP(ohlcv, 14)
		analysis.Indicators["CHOP"] = chop
		var s float64
		if chop > 61.8 {
			s = 100
		} else if chop < 38.2 {
			s = 20
		} else {
			s = 60
		}
		score += s * weights["CHOP"]
		totalWeight += weights["CHOP"]
	}

	if selectedMap["ER"] {
		er := indicators.CalculateER(ohlcv, 14)
		analysis.Indicators["ER"] = er
		var s float64
		if er < 0.3 {
			s = 100
		} else if er > 0.7 {
			s = 20
		} else {
			s = 60
		}
		score += s * weights["ER"]
		totalWeight += weights["ER"]
	}

	if totalWeight > 0 {
		analysis.Score = int(score / totalWeight)
	}

	// Status classification
	if analysis.Score >= 80 {
		analysis.Status = "Excellent"
	} else if analysis.Score >= 65 {
		analysis.Status = "Good"
	} else if analysis.Score >= 50 {
		analysis.Status = "Regular"
	} else {
		analysis.Status = "Not Recommended"
	}

	// Grid configuration
	if analysis.Score >= 50 {
		az.calculateGridConfig(analysis, asset.CurrentPrice, ohlcv, sr)
	}

	// Justification
	analysis.Justification = az.generateJustification(asset.Symbol, analysis)

	asset.Analysis = analysis
	return asset
}

func (az *Analyzer) calculateGridConfig(analysis *models.Analysis, currentPrice float64, ohlcv []models.OHLCV, sr *indicators.SRResult) {
	n := len(ohlcv)
	lookback := 30
	if n < lookback {
		lookback = n
	}

	var maxHigh float64 = 0
	var minLow float64 = math.MaxFloat64
	for i := n - lookback; i < n; i++ {
		if ohlcv[i].High > maxHigh {
			maxHigh = ohlcv[i].High
		}
		if ohlcv[i].Low < minLow {
			minLow = ohlcv[i].Low
		}
	}

	lower := minLow
	upper := maxHigh

	// SAFETY: Ensure a minimum range width if it's too narrow
	if upper <= lower*1.01 {
		lower = currentPrice * 0.98
		upper = currentPrice * 1.02
	}

	amplitude := (upper - lower) / lower * 100
	
	// BingX fees 0.1% maker + 0.1% taker = 0.2%
	// Min profit > 0.2%, Recommended > 0.6%
	recGrids := int(amplitude / 0.6)
	if recGrids < 5 { recGrids = 5 }
	if recGrids > 150 { recGrids = 150 }

	profitGross := amplitude / float64(recGrids)
	profitNet := profitGross - 0.2

	analysis.GridConfig = models.GridConfig{
		LowerRange:         lower,
		UpperRange:         upper,
		Amplitude:          amplitude,
		RecommendedGrids:   recGrids,
		ProfitPerGridGross: profitGross,
		ProfitPerGridNet:   profitNet,
	}

	if profitNet < 0.1 {
		analysis.GridConfig.Warning = "ADVERTENCIA: el grid no cubre las comisiones con esta configuración"
	}

	// Alternatives
	analysis.GridConfig.AlternativeConfigs = []models.AlternativeConfig{
		{Type: "Conservadora", Grids: recGrids / 2, ProfitPerGrid: (amplitude / float64(recGrids/2)) - 0.2, Observation: "Mayor rentabilidad/grid"},
		{Type: "Recomendada", Grids: recGrids, ProfitPerGrid: profitNet, Observation: "Balance óptimo"},
		{Type: "Agresiva", Grids: int(float64(recGrids) * 1.5), ProfitPerGrid: (amplitude / (float64(recGrids) * 1.5)) - 0.2, Observation: "Más operaciones"},
	}
}

func (az *Analyzer) generateJustification(symbol string, analysis *models.Analysis) string {
	text := fmt.Sprintf("%s ocupa una posición con un score de %d/100. ", symbol, analysis.Score)
	
	if val, ok := analysis.Indicators["ADX"]; ok {
		if val < 20 {
			text += "A favor: el ADX de %.1f confirma un mercado en rango sin tendencia definida. "
			text = fmt.Sprintf(text, val)
		} else {
			text += "En contra: el ADX de %.1f sugiere una tendencia en formación que puede romper el grid. "
			text = fmt.Sprintf(text, val)
		}
	}

	if val, ok := analysis.Indicators["RSI"]; ok {
		if val >= 40 && val <= 60 {
			text += "El RSI de %.1f indica neutralidad del momentum ideal para grillas. "
			text = fmt.Sprintf(text, val)
		}
	}

	if val, ok := analysis.Indicators["ATR"]; ok {
		text += "El ATR relativo del %.1f%% ofrece volatilidad para generar rendimiento. "
		text = fmt.Sprintf(text, val)
	}

	return text
}
