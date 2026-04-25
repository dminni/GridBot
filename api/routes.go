package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *Handlers) {
	api := r.Group("/api")
	{
		api.GET("/assets/universe", h.GetUniverse)
		api.POST("/analysis/run", h.RunAnalysis)
		api.GET("/asset/:symbol", h.GetAssetDetail)
		api.GET("/asset/:symbol/ohlcv", h.GetOHLCV)
	}

	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
}
