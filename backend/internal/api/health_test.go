package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// mockPinger is a test double for the api.Pinger interface.
type mockPinger struct {
	pingFn func(ctx context.Context) error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestReady_AllHealthy(t *testing.T) {
	db := &mockPinger{pingFn: func(ctx context.Context) error { return nil }}
	redis := &mockPinger{pingFn: func(ctx context.Context) error { return nil }}
	handler := NewHealthHandler(db, redis, testLogger())

	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handler.Ready(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ReadyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Status)
	}
	if resp.Postgres != "ok" {
		t.Errorf("postgres = %q, want ok", resp.Postgres)
	}
	if resp.Redis != "ok" {
		t.Errorf("redis = %q, want ok", resp.Redis)
	}
	if resp.Timestamp.IsZero() {
		t.Error("timestamp is zero — expected a non-zero time")
	}
}

func TestReady_PostgresUnhealthy(t *testing.T) {
	db := &mockPinger{pingFn: func(ctx context.Context) error {
		return errors.New("connection refused")
	}}
	redis := &mockPinger{pingFn: func(ctx context.Context) error { return nil }}
	handler := NewHealthHandler(db, redis, testLogger())

	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handler.Ready(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d: %s", w.Code, w.Body.String())
	}

	var resp ReadyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "degraded" {
		t.Errorf("status = %q, want degraded", resp.Status)
	}
	if resp.Postgres != "unhealthy" {
		t.Errorf("postgres = %q, want unhealthy", resp.Postgres)
	}
}

func TestReady_RedisUnhealthy(t *testing.T) {
	db := &mockPinger{pingFn: func(ctx context.Context) error { return nil }}
	redis := &mockPinger{pingFn: func(ctx context.Context) error {
		return errors.New("i/o timeout")
	}}
	handler := NewHealthHandler(db, redis, testLogger())

	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handler.Ready(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d: %s", w.Code, w.Body.String())
	}

	var resp ReadyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "degraded" {
		t.Errorf("status = %q, want degraded", resp.Status)
	}
	if resp.Redis != "unhealthy" {
		t.Errorf("redis = %q, want unhealthy", resp.Redis)
	}
	// Postgres should still be healthy.
	if resp.Postgres != "ok" {
		t.Errorf("postgres = %q, want ok", resp.Postgres)
	}
}

func TestReady_BothUnhealthy(t *testing.T) {
	db := &mockPinger{pingFn: func(ctx context.Context) error {
		return errors.New("connection refused")
	}}
	redis := &mockPinger{pingFn: func(ctx context.Context) error {
		return errors.New("i/o timeout")
	}}
	handler := NewHealthHandler(db, redis, testLogger())

	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handler.Ready(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d: %s", w.Code, w.Body.String())
	}

	var resp ReadyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "degraded" {
		t.Errorf("status = %q, want degraded", resp.Status)
	}
	if resp.Postgres != "unhealthy" {
		t.Errorf("postgres = %q, want unhealthy", resp.Postgres)
	}
	if resp.Redis != "unhealthy" {
		t.Errorf("redis = %q, want unhealthy", resp.Redis)
	}
}

func TestReady_RedisNotConfigured(t *testing.T) {
	db := &mockPinger{pingFn: func(ctx context.Context) error { return nil }}
	handler := NewHealthHandler(db, nil, testLogger()) // redis is nil

	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handler.Ready(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ReadyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Status)
	}
	if resp.Redis != "not_configured" {
		t.Errorf("redis = %q, want not_configured", resp.Redis)
	}
}
