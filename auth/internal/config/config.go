package config

import "os"

// Holds application's runtime settings.
type Config struct {
	Port     string
	Database string
}

// Load reads settings from environment variables, using defaults for unset values.
// It does not load a .env file itself.
func Load() Config {
	return Config{
		Port:     getEnv("PORT", "8081"),
		Database: getEnv("DATABASE", "./data/auth.db"),
	}
}

// getEnv returns the environment variable's value, or fallback if it is unset.
// An explicitly set empty value is returned as-is.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
