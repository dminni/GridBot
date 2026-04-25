package api

import (
	"net/http"
	"strconv"
	"sync"
	"github.com/gin-gonic/gin"
	"github.com/dminni/gridbot/cache"
	"github.com/dminni/gridbot/models"
	"github.com/dminni/gridbot/services"
)

type Handlers struct {
	CoinGecko *services.CoinGeckoClient
	BingX     *services.BingXClient
	Analyzer  *services.Analyzer
	Simulator *services.Simulator
	Cache     *cache.Cache
}

func NewHandlers(cg *services.CoinGeckoClient, bx *services.BingXClient, az *services.Analyzer, sim *services.Simulator, c *cache.Cache) *Handlers {
	return &Handlers{cg, bx, az, sim, c}
}

func (h *Handlers) GetUniverse(c *gin.Context) {
	if cached, ok := h.Cache.Get("universe"); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	assets, err := h.CoinGecko.GetTop100Assets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter by BingX availability
	bxSymbols, err := h.BingX.GetSymbols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var filtered []models.Asset
	for _, a := range assets {
		if bxSymbols[a.Symbol+"-USDT"] {
			filtered = append(filtered, a)
		}
	}

	h.Cache.Set("universe", filtered, 30*60*1e9) // 30 mins
	c.JSON(http.StatusOK, filtered)
}

func (h *Handlers) RunAnalysis(c *gin.Context) {
	var req models.AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.Indicators) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Seleccioná al menos un indicador para ejecutar el análisis"})
		return
	}

	var assets []models.Asset
	if cached, ok := h.Cache.Get("universe"); ok {
		assets = cached.([]models.Asset)
	} else {
		// If not in cache, we need to fetch it (or return error)
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Universe not loaded. Please refresh assets first."})
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	results := make([]models.Asset, len(assets))

	for i, a := range assets {
		wg.Add(1)
		go func(idx int, asset models.Asset) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			
			results[idx] = h.Analyzer.AnalyzeAsset(asset, req.Indicators)
		}(i, a)
	}

	wg.Wait()

	var ranked []models.Asset
	var discarded []models.Asset

	for _, a := range results {
		if a.Analysis != nil && a.Analysis.Score >= 50 {
			ranked = append(ranked, a)
		} else {
			discarded = append(discarded, a)
		}
	}

	res := models.AnalysisResult{
		RankedAssets:    ranked,
		DiscardedAssets: discarded,
		Summary:         "Análisis completado",
	}

	h.Cache.Set("last_analysis", res, 5*60*1e9)
	c.JSON(http.StatusOK, res)
}

func (h *Handlers) GetAssetDetail(c *gin.Context) {
	symbol := c.Param("symbol")
	periodStr := c.DefaultQuery("period", "30")
	period, _ := strconv.Atoi(periodStr)

	// In a real app we'd fetch fresh data. For this, we'll use cached analysis if available.
	var lastAnalysis models.AnalysisResult
	if cached, ok := h.Cache.Get("last_analysis"); ok {
		lastAnalysis = cached.(models.AnalysisResult)
		for _, a := range lastAnalysis.RankedAssets {
			if a.Symbol == symbol {
				// Simulating history too
				sim, _ := h.Simulator.Simulate(symbol, period, a.Analysis.GridConfig)
				c.JSON(http.StatusOK, gin.H{
					"asset":      a,
					"simulation": sim,
				})
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found in last analysis"})
}
func (h *Handlers) GetOHLCV(c *gin.Context) {
	symbol := c.Param("symbol")
	limitStr := c.DefaultQuery("limit", "200")
	interval := c.DefaultQuery("interval", "1d")
	limit, _ := strconv.Atoi(limitStr)

	data, err := h.BingX.GetOHLCV(symbol, interval, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
