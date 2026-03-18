package bootstrap

import (
	"os"
	"strconv"
	"time"
)
import _ "github.com/joho/godotenv/autoload"

type Config struct {
	PrometheusURL   string
	DatabasePath    string
	LogLevel        string
	CollectInterval time.Duration
	RetentionDays   int
}

func Configure() *Config {
	return &Config{
		PrometheusURL:   getEnv("PROM_URL", "http://localhost:9090"),
		DatabasePath:    getEnv("DB_PATH", "./metrics.db"),
		LogLevel:        getEnv("LOG_LEVEL", "debug"),
		CollectInterval: time.Duration(getEnvInt("COLLECTION_INTERVAL", 5)) * time.Second,
		RetentionDays:   getEnvInt("RETENTION_DAYS", 7),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return intVal
}
