package service

import (
	"errors"
	"testing"
)

func TestHealthResultAddCheck(t *testing.T) {
	result := HealthResult{
		Ready:  true,
		Checks: make(map[string]CheckResult),
	}

	result.addCheck("database", nil)

	if !result.Ready {
		t.Fatal("healthy check marked result unavailable")
	}
	if result.Checks["database"].Status != "ok" {
		t.Fatalf("got status %q, want ok",
			result.Checks["database"].Status)
	}

	result.addCheck("cache", errors.New("cache unavailable"))

	if result.Ready {
		t.Fatal("failed check left result ready")
	}
	if result.Checks["cache"].Status != "unavailable" {
		t.Fatalf("got status %q, want unavailable",
			result.Checks["cache"].Status)
	}

	result.addCheck("database", nil)

	if result.Ready {
		t.Fatal("successful later check restored readiness")
	}
}
