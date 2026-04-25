package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/dminni/gridbot/api"
	"github.com/dminni/gridbot/cache"
	"github.com/dminni/gridbot/services"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bxBase := os.Getenv("BINGX_API_BASE")
	cgBase := os.Getenv("COINGECKO_API_BASE")
	cgKey := os.Getenv("COINGECKO_API_KEY")

	// Initialize services
	cgClient := services.NewCoinGeckoClient(cgBase, cgKey)
	bxClient := services.NewBingXClient(bxBase)
	analyzer := services.NewAnalyzer(bxClient)
	simulator := services.NewSimulator(bxClient)
	memCache := cache.NewCache()

	handlers := api.NewHandlers(cgClient, bxClient, analyzer, simulator, memCache)

	r := gin.Default()
	
	api.SetupRoutes(r, handlers)

	log.Printf("GridBot server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
