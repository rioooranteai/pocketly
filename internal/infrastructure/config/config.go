package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	DBPath     string
	JWTSecret  string
	AIProvider string
	AIAPIKey   string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		AppPort:    os.Getenv("APP_PORT"),
		DBPath:     os.Getenv("DB_PATH"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		AIProvider: os.Getenv("AI_PROVIDER"),
		AIAPIKey:   os.Getenv("AI_API_KEY"),
	}
}
