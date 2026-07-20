// Package middleware provides HTTP middleware for the WMS API layer.
// Includes request ID propagation, structured request logging, panic recovery, and CORS.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Context keys for request-scoped values.
type contextKey string

const (
	// RequestIDKey is the context key for the request ID.
	RequestIDKey contextKey = "request_id"
)

// RequestID is a middleware that extracts the X-Request-ID header from incoming
// requests, or generates a new UUID if none is provided. The request ID is
// stored in the request context and set on the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}

		// Store in context for downstream handlers and middleware.
		ctx := context.WithValue(r.Context(), RequestIDKey, id)

		// Echo back to the client.
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}

// Logger returns a middleware that logs each HTTP request using structured logging.
// It records the method, path, status code, duration, and request ID.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := GetRequestID(r.Context())

			// Wrap response writer to capture status code.
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", duration),
				slog.String("request_id", requestID),
			}

			// Include query string if present (but not for sensitive paths in the future).
			if r.URL.RawQuery != "" {
				attrs = append(attrs, slog.String("query", r.URL.RawQuery))
			}

			// Include remote address.
			attrs = append(attrs, slog.String("remote_addr", r.RemoteAddr))

			// Log at appropriate level based on status code.
			if rw.statusCode >= 500 {
				logger.Error("request completed", attrs...)
			} else if rw.statusCode >= 400 {
				logger.Warn("request completed", attrs...)
			} else {
				logger.Info("request completed", attrs...)
			}
		})
	}
}

// Recovery returns a middleware that recovers from panics, logs the stack trace,
// and responds with a 500 Internal Server Error.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := GetRequestID(r.Context())
					stack := debug.Stack()

					logger.Error("panic recovered",
						slog.Any("panic", rec),
						slog.String("request_id", requestID),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("stack", string(stack)),
					)

					// Only write header if not already written.
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins allowed to make cross-origin requests.
	// Use "*" to allow all origins (not recommended with credentials).
	AllowedOrigins []string

	// AllowedMethods is a list of HTTP methods allowed for cross-origin requests.
	// Defaults to GET, POST, PUT, DELETE, PATCH, OPTIONS.
	AllowedMethods []string

	// AllowedHeaders is a list of non-simple headers allowed.
	// Defaults to Accept, Authorization, Content-Type, X-Request-ID.
	AllowedHeaders []string

	// ExposedHeaders is a list of headers exposed to the browser.
	ExposedHeaders []string

	// AllowCredentials indicates whether to include the Access-Control-Allow-Credentials header.
	AllowCredentials bool

	// MaxAge is the duration (in seconds) that preflight responses may be cached.
	MaxAge int
}

// DefaultCORSConfig returns a sensible default CORS configuration for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}
}

// CORS returns a middleware that sets CORS headers on every response and handles
// preflight (OPTIONS) requests.
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	// Pre-compute joined header values.
	origins := joinStrings(config.AllowedOrigins)
	methods := joinStrings(config.AllowedMethods)
	headers := joinStrings(config.AllowedHeaders)
	exposed := joinStrings(config.ExposedHeaders)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers on every response.
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if exposed != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposed)
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", itoa(config.MaxAge))
			}

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// joinStrings joins a slice of strings with ", " separator.
func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(ss[0])
	for i := 1; i < len(ss); i++ {
		b.WriteString(", ")
		b.WriteString(ss[i])
	}
	return b.String()
}

// itoa is a simple int-to-string helper (avoids importing strconv for a single use).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	// Handle negative (shouldn't happen for MaxAge but be safe).
	neg := n < 0
	if neg {
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
