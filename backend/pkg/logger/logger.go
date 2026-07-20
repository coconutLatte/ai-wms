// Package logger provides structured logging for the WMS application.
package logger

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with convenience methods.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger with the given log level.
func New(level string) *Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		// Add source in development for debugging
		AddSource: level == "debug",
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{slog.New(handler)}
}

// With creates a child logger with additional fields.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}
