package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Env                 string
	CoinGeckoAPIKey     string
	CMCAPIKey           string
	BingXAPIKey         string
	BingXSecretKey      string
	DefaultCapital      float64
	BingXCommission     float64
	UpdateIntervalHours int
	AnalysisLimit       int
	DBPath              string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:                getEnv("PORT", "8080"),
		Env:                 getEnv("ENV", "development"),
		CoinGeckoAPIKey:     getEnv("COINGECKO_API_KEY", ""),
		CMCAPIKey:           getEnv("CMC_API_KEY", ""),
		BingXAPIKey:         getEnv("BINGX_API_KEY", ""),
		BingXSecretKey:      getEnv("BINGX_SECRET_KEY", ""),
		DefaultCapital:      getEnvAsFloat("DEFAULT_CAPITAL", 1000.0),
		BingXCommission:     getEnvAsFloat("DEFAULT_BINGX_COMMISSION", 0.001),
		UpdateIntervalHours: getEnvAsInt("UPDATE_INTERVAL_HOURS", 6),
		AnalysisLimit:       getEnvAsInt("ANALYSIS_LIMIT", 100),
		DBPath:              getEnv("DB_PATH", "gridbot.db"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsFloat(key string, fallback float64) float64 {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseFloat(strValue, 64); err == nil {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
