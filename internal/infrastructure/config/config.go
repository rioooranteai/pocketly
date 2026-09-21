package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

/*
Config holds all environment-derived settings the application needs
at startup: server port, database location, JWT signing secret, and
AI provider credentials.
*/
type Config struct {
	AppPort        string
	DBPath         string
	JWTSecret      string
	AIProvider     string
	AIAPIKey       string
	VisionAPIKey   string
	MaxImageSizeMB string

	VisionConfig VisionConfig
}

type VisionConfig struct {
	APIKey      string
	MaxFileSize int64
}

/*
Load reads environment variables into a Config. It attempts to load a
local .env file first, but does not fail if one is absent — in
production, environment variables are typically injected directly by
the deployment platform (Docker, Kubernetes, etc.) instead.
*/
func Load() *Config {
	_ = godotenv.Load()

	maxFileSize, err := strconv.ParseInt(os.Getenv("MAX_IMAGE_SIZE_MB"), 10, 64)
	if err != nil || maxFileSize <= 0 {
		maxFileSize = 5 * 1024 * 1024 // default 5MB kalau env kosong/invalid
	}

	return &Config{
		AppPort:    os.Getenv("APP_PORT"),
		DBPath:     os.Getenv("DB_PATH"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		AIProvider: os.Getenv("AI_PROVIDER"),
		AIAPIKey:   os.Getenv("AI_API_KEY"),

		VisionConfig: VisionConfig{
			APIKey:      os.Getenv("VISION_API_KEY"),
			MaxFileSize: maxFileSize,
		},
	}
}
