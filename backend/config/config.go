package config

import (
	"os"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	Environment string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost/sistem_akuntansi?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-here"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
