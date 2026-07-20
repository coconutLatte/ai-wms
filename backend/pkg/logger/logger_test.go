package logger

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	log := New("info")
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	if log.Logger == nil {
		t.Fatal("expected non-nil underlying slog.Logger")
	}
}

func TestNew_AllLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		log := New(level)
		if log == nil {
			t.Errorf("level=%s: expected non-nil logger", level)
		}
	}
}

func TestNew_InvalidLevelDefaultsToInfo(t *testing.T) {
	log := New("invalid")
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	// Should not panic and should default to info.
	log.Info("test default level")
}

func TestWith(t *testing.T) {
	log := New("info")
	child := log.With("component", "test")
	if child == nil {
		t.Fatal("expected non-nil child logger")
	}
	// Ensure child is independent.
	if log == child {
		t.Fatal("expected With() to return a new logger instance")
	}
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	// Verify the context value is set.
	val := ctx.Value(requestIDKey)
	if val == nil {
		t.Fatal("expected request_id in context")
	}
	if val != "req-123" {
		t.Errorf("expected req-123, got %v", val)
	}
}

func TestWithRequestID_EmptyID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "")

	val := ctx.Value(requestIDKey)
	if val == nil {
		t.Fatal("expected empty request_id in context")
	}
}

func TestContextMethods_NoPanic(t *testing.T) {
	log := New("info")
	ctx := context.Background()

	// These should not panic.
	log.DebugContext(ctx, "debug message", "key", "val")
	log.InfoContext(ctx, "info message", "key", "val")
	log.WarnContext(ctx, "warn message", "key", "val")
	log.ErrorContext(ctx, "error message", "key", "val")
}

func TestContextMethods_WithRequestID(t *testing.T) {
	log := New("info")
	ctx := WithRequestID(context.Background(), "req-456")

	// These should not panic and should include request_id.
	log.InfoContext(ctx, "info with request")
	log.ErrorContext(ctx, "error with request")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected string // string representation of slog.Level
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"error", "ERROR"},
		{"invalid", "INFO"}, // defaults to info
		{"", "INFO"},
	}

	for _, tt := range tests {
		level := parseLevel(tt.input)
		if level.String() != tt.expected {
			t.Errorf("parseLevel(%q) = %s, want %s", tt.input, level.String(), tt.expected)
		}
	}
}
