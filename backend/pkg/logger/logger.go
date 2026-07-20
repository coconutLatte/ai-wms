// Package logger provides structured logging for the WMS application.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with convenience methods.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger with the given log level.
// Supported levels: debug, info, warn, error.
// Logs are written as JSON to stdout.
func New(level string) *Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(level),
		AddSource: level == "debug",
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{slog.New(handler)}
}

// With creates a child logger with additional structured fields.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

// ── Context-aware convenience methods ──────────────────────────────────────────────────────

// DebugContext logs at debug level with context-derived fields (e.g., request_id).
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.withContext(ctx).Debug(msg, args...)
}

// InfoContext logs at info level with context-derived fields.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.withContext(ctx).Info(msg, args...)
}

// WarnContext logs at warn level with context-derived fields.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.withContext(ctx).Warn(msg, args...)
}

// ErrorContext logs at error level with context-derived fields.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.withContext(ctx).Error(msg, args...)
}

// withContext extracts common fields from context (e.g., request_id)
// and returns a child logger with those fields attached.
func (l *Logger) withContext(ctx context.Context) *slog.Logger {
	args := make([]any, 0)

	if reqID := ctx.Value(requestIDKey); reqID != nil {
		if id, ok := reqID.(string); ok && id != "" {
			args = append(args, slog.String("request_id", id))
		}
	}

	if len(args) == 0 {
		return l.Logger
	}
	return l.Logger.With(args...)
}

// contextKey is the unexported type used for context keys to avoid collisions.
type contextKey struct{ name string }

var requestIDKey = contextKey{name: "request_id"}

// WithRequestID returns a context with the given request ID stored.
// This is picked up by the context-aware logging methods (DebugContext, etc.).
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// ── Helpers ────────────────────────────────────────────────────────────────────────────────

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
