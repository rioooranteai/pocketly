package config

import (
	"os"

	"github.com/joho/godotenv"
)

/*
Config holds all environment-derived settings the application needs
at startup: server port, database location, JWT signing secret, and
AI provider credentials.
*/
type Config struct {
	AppPort    string
	DBPath     string
	JWTSecret  string
	AIProvider string
	AIAPIKey   string
}

/*
Load reads environment variables into a Config. It attempts to load a
local .env file first, but does not fail if one is absent — in
production, environment variables are typically injected directly by
the deployment platform (Docker, Kubernetes, etc.) instead.
*/
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppPort:    os.Getenv("APP_PORT"),
		DBPath:     os.Getenv("DB_PATH"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		AIProvider: os.Getenv("AI_PROVIDER"),
		AIAPIKey:   os.Getenv("AI_API_KEY"),
	}
}
