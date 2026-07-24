// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// Pinger wraps the Ping method for dependency health checks.
// Both postgres.DB and redis.Client satisfy this interface.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles readiness probe endpoints (GET /ready).
type HealthHandler struct {
	postgres Pinger
	redis    Pinger // Optional — may be nil if Redis is not configured.
	logger   *slog.Logger
}

// NewHealthHandler creates a new HealthHandler.
// redis may be nil if Redis is not configured.
func NewHealthHandler(postgres, redis Pinger, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		postgres: postgres,
		redis:    redis,
		logger:   logger,
	}
}

// ReadyResponse is the JSON payload returned by GET /ready.
type ReadyResponse struct {
	Status    string    `json:"status"`    // "ok" or "degraded"
	Postgres  string    `json:"postgres"`  // "ok" or "unhealthy"
	Redis     string    `json:"redis"`     // "ok", "unhealthy", or "not_configured"
	Timestamp time.Time `json:"timestamp"`
}

// Ready handles GET /ready — a Kubernetes-style readiness probe that pings
// PostgreSQL (required) and Redis (optional). Returns 200 with per-service
// status in the body when all dependencies are healthy. Returns 503 if any
// dependency is unhealthy.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := ReadyResponse{
		Status:    "ok",
		Postgres:  "ok",
		Redis:     "not_configured",
		Timestamp: time.Now().UTC(),
	}

	// Check PostgreSQL (required).
	if err := h.postgres.Ping(ctx); err != nil {
		h.logger.Warn("readiness probe: postgres ping failed",
			slog.String("error", err.Error()),
		)
		resp.Status = "degraded"
		resp.Postgres = "unhealthy"
	}

	// Check Redis (optional — only if configured).
	if h.redis != nil {
		if err := h.redis.Ping(ctx); err != nil {
			h.logger.Warn("readiness probe: redis ping failed",
				slog.String("error", err.Error()),
			)
			resp.Status = "degraded"
			resp.Redis = "unhealthy"
		} else {
			resp.Redis = "ok"
		}
	}

	statusCode := http.StatusOK
	if resp.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
