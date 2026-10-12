package service

import (
	"context"
	"log/slog"

	"auth/internal/repository"
)

// CheckResult contains the outcome of one readiness check.
type CheckResult struct {
	Status string
}

// HealthResult contains overall readiness and individual check results.
type HealthResult struct {
	Ready  bool
	Checks map[string]CheckResult
}

// HealthService checks whether required dependencies are available.
type HealthService struct {
	repo *repository.HealthRepository
}

// NewHealthService creates a health service.
func NewHealthService(repo *repository.HealthRepository) *HealthService {
	return &HealthService{repo: repo}
}

// Check runs readiness checks and returns their results.
func (s *HealthService) Check(ctx context.Context) HealthResult {
	result := HealthResult{Ready: true, Checks: make(map[string]CheckResult)}

	err := s.repo.CheckDatabase(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "readiness check failed", "check", "database", "err", err)
	}
	result.addCheck("database", err)

	return result
}

func (r *HealthResult) addCheck(name string, err error) {
	if err != nil {
		r.Checks[name] = CheckResult{Status: "unavailable"}
		r.Ready = false
		return
	}

	r.Checks[name] = CheckResult{Status: "ok"}
}
