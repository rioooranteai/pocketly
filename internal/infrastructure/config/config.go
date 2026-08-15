package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort   string
	DBPath    string
	JWTSecret string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		AppPort: os.Getenv("APP_PORT"),
	}
}
