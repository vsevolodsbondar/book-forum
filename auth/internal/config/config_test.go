package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("DATABASE", "./test.db")
	t.Setenv("SESSION_LIFETIME_SECONDS", "3600")
	t.Setenv("SESSION_IDLE_TIMEOUT_SECONDS", "600")
	t.Setenv("ARGON2_MAX_CONCURRENCY", "4")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	want := Config{
		Port:                 "9000",
		Database:             "./test.db",
		SessionLifetime:      time.Hour,
		SessionIdleTimeout:   10 * time.Minute,
		Argon2MaxConcurrency: 4,
	}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{
		"PORT",
		"DATABASE",
		"SESSION_LIFETIME_SECONDS",
		"SESSION_IDLE_TIMEOUT_SECONDS",
		"ARGON2_MAX_CONCURRENCY",
	} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	want := Config{
		Port:                 "8081",
		Database:             "./data/auth.db",
		SessionLifetime:      30 * 24 * time.Hour,
		SessionIdleTimeout:   7 * 24 * time.Hour,
		Argon2MaxConcurrency: 2,
	}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadInvalidValues(t *testing.T) {
	for _, key := range []string{
		"SESSION_LIFETIME_SECONDS",
		"SESSION_IDLE_TIMEOUT_SECONDS",
		"ARGON2_MAX_CONCURRENCY",
	} {
		for _, value := range []string{
			"",
			"invalid",
			"0",
			"-1",
			"9223372036854775808",
		} {
			t.Run(key+"/"+value, func(t *testing.T) {
				t.Setenv("SESSION_LIFETIME_SECONDS", "3600")
				t.Setenv("SESSION_IDLE_TIMEOUT_SECONDS", "600")
				t.Setenv("ARGON2_MAX_CONCURRENCY", "2")
				t.Setenv(key, value)

				if _, err := Load(); err == nil {
					t.Fatalf("expected error for %s=%q", key, value)
				}
			})
		}
	}
}

func TestLoadDurationOverflow(t *testing.T) {
	for _, key := range []string{
		"SESSION_LIFETIME_SECONDS",
		"SESSION_IDLE_TIMEOUT_SECONDS",
	} {
		t.Run(key, func(t *testing.T) {
			t.Setenv("SESSION_LIFETIME_SECONDS", "3600")
			t.Setenv("SESSION_IDLE_TIMEOUT_SECONDS", "600")
			t.Setenv("ARGON2_MAX_CONCURRENCY", "2")
			t.Setenv(key, "9223372037")

			if _, err := Load(); err == nil {
				t.Fatalf("expected duration overflow error for %s", key)
			}
		})
	}
}
