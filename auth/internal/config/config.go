package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application's runtime settings.
type Config struct {
	Port     string
	Database string

	SessionLifetime    time.Duration
	SessionIdleTimeout time.Duration

	Argon2MemoryKiB      uint32
	Argon2Iterations     uint32
	Argon2Parallelism    uint8
	Argon2MaxConcurrency int
}

// Load reads settings from environment variables, using defaults for unset values.
// It does not load a .env file itself.
func Load() (cfg Config, err error) {
	cfg.Port = getEnv("PORT", "8081")
	cfg.Database = getEnv("DATABASE", "./data/auth.db")

	cfg.SessionLifetime, err = getDuration("SESSION_LIFETIME_SECONDS", "2592000")
	if err != nil {
		return Config{}, err
	}

	cfg.SessionIdleTimeout, err = getDuration("SESSION_IDLE_TIMEOUT_SECONDS", "604800")
	if err != nil {
		return Config{}, err
	}

	memory, err := getUint("ARGON2_MEMORY_KIB", "65536", 32, 19456, 262144)
	if err != nil {
		return Config{}, err
	}
	cfg.Argon2MemoryKiB = uint32(memory)

	iterations, err := getUint("ARGON2_ITERATIONS", "3", 32, 2, 10)
	if err != nil {
		return Config{}, err
	}
	cfg.Argon2Iterations = uint32(iterations)

	parallelism, err := getUint("ARGON2_PARALLELISM", "1", 8, 1, 4)
	if err != nil {
		return Config{}, err
	}
	cfg.Argon2Parallelism = uint8(parallelism)

	cfg.Argon2MaxConcurrency, err = getPositiveInt("ARGON2_MAX_CONCURRENCY", "2")
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// getEnv returns the environment variable's value, or fallback if it is unset.
// An explicitly set empty value is returned as-is.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getDuration reads a positive duration in seconds from an environment variable.
func getDuration(key, fallback string) (time.Duration, error) {
	seconds, err := strconv.ParseInt(getEnv(key, fallback), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	if seconds <= 0 || seconds > int64((1<<63-1)/time.Second) {
		return 0, fmt.Errorf("%s: seconds must be positive and fit in time.Duration", key)
	}

	return time.Duration(seconds) * time.Second, nil
}

// getPositiveInt reads a positive integer from an environment variable.
func getPositiveInt(key, fallback string) (int, error) {
	value, err := strconv.Atoi(getEnv(key, fallback))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s: must be positive", key)
	}

	return value, nil
}

// getUint reads an unsigned integer within the given bounds from an environment variable.
func getUint(key, fallback string, bits int, min, max uint64) (uint64, error) {
	value, err := strconv.ParseUint(getEnv(key, fallback), 10, bits)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	if value < min || value > max {
		return 0, fmt.Errorf("%s: must be between %d and %d", key, min, max)
	}

	return value, nil
}
