package config

import "os"

type Config struct {
	ServiceName string
	Port        string
	DatabaseURL string
}

func Load() Config {
	return Config{
		ServiceName: envOrDefault("SERVICE_NAME", "delivery"),
		Port:        envOrDefault("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
