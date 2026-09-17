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

	t.Setenv("ARGON2_MEMORY_KIB", "32768")
	t.Setenv("ARGON2_ITERATIONS", "2")
	t.Setenv("ARGON2_PARALLELISM", "2")
	t.Setenv("ARGON2_MAX_CONCURRENCY", "4")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}

	want := Config{
		Port:     "9000",
		Database: "./test.db",

		SessionLifetime:    time.Hour,
		SessionIdleTimeout: 10 * time.Minute,

		Argon2MemoryKiB:      32768,
		Argon2Iterations:     2,
		Argon2Parallelism:    2,
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

		"ARGON2_MEMORY_KIB",
		"ARGON2_ITERATIONS",
		"ARGON2_PARALLELISM",
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
		Port:     "8081",
		Database: "./data/auth.db",

		SessionLifetime:    30 * 24 * time.Hour,
		SessionIdleTimeout: 7 * 24 * time.Hour,

		Argon2MemoryKiB:      65536,
		Argon2Iterations:     3,
		Argon2Parallelism:    1,
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

		"ARGON2_MEMORY_KIB",
		"ARGON2_ITERATIONS",
		"ARGON2_PARALLELISM",
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

				t.Setenv("ARGON2_MEMORY_KIB", "65536")
				t.Setenv("ARGON2_ITERATIONS", "3")
				t.Setenv("ARGON2_PARALLELISM", "1")
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

			t.Setenv("ARGON2_MEMORY_KIB", "65536")
			t.Setenv("ARGON2_ITERATIONS", "3")
			t.Setenv("ARGON2_PARALLELISM", "1")
			t.Setenv("ARGON2_MAX_CONCURRENCY", "2")

			t.Setenv(key, "9223372037")

			if _, err := Load(); err == nil {
				t.Fatalf("expected duration overflow error for %s", key)
			}
		})
	}
}

func TestLoadArgon2Bounds(t *testing.T) {
	for _, tc := range []struct {
		key   string
		value string
		valid bool
	}{
		{"ARGON2_MEMORY_KIB", "19455", false},
		{"ARGON2_MEMORY_KIB", "19456", true},
		{"ARGON2_MEMORY_KIB", "262144", true},
		{"ARGON2_MEMORY_KIB", "262145", false},
		{"ARGON2_ITERATIONS", "1", false},
		{"ARGON2_ITERATIONS", "2", true},
		{"ARGON2_ITERATIONS", "10", true},
		{"ARGON2_ITERATIONS", "11", false},
		{"ARGON2_PARALLELISM", "1", true},
		{"ARGON2_PARALLELISM", "4", true},
		{"ARGON2_PARALLELISM", "5", false},
	} {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			t.Setenv("SESSION_LIFETIME_SECONDS", "3600")
			t.Setenv("SESSION_IDLE_TIMEOUT_SECONDS", "600")

			t.Setenv("ARGON2_MEMORY_KIB", "65536")
			t.Setenv("ARGON2_ITERATIONS", "3")
			t.Setenv("ARGON2_PARALLELISM", "1")
			t.Setenv("ARGON2_MAX_CONCURRENCY", "2")

			t.Setenv(tc.key, tc.value)

			_, err := Load()
			if tc.valid && err != nil {
				t.Fatalf("expected valid configuration: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected error for %s=%q", tc.key, tc.value)
			}
		})
	}
}
