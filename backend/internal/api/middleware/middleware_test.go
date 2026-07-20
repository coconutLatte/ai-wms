package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_ExtractsHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id != "test-request-123" {
			t.Errorf("expected request ID 'test-request-123', got '%s'", id)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") != "test-request-123" {
		t.Errorf("expected response header X-Request-ID to be 'test-request-123', got '%s'",
			rec.Header().Get("X-Request-ID"))
	}
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected a generated request ID, got empty string")
	}

	// Verify it's a valid UUID format (36 chars with dashes).
	if len(capturedID) != 36 {
		t.Errorf("expected UUID length 36, got %d: '%s'", len(capturedID), capturedID)
	}

	// Response should echo back the generated ID.
	if rec.Header().Get("X-Request-ID") != capturedID {
		t.Errorf("response header X-Request-ID '%s' does not match context ID '%s'",
			rec.Header().Get("X-Request-ID"), capturedID)
	}
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	id := GetRequestID(req.Context())
	if id != "" {
		t.Errorf("expected empty string for bare context, got '%s'", id)
	}
}

func TestLogger_LogsRequest(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	handler := RequestID(Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/test?page=1", nil)
	req.Header.Set("X-Request-ID", "log-test-1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "request completed") {
		t.Errorf("expected log message 'request completed', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "log-test-1") {
		t.Errorf("expected request_id in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "/api/test") {
		t.Errorf("expected path in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "GET") {
		t.Errorf("expected method in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "page=1") {
		t.Errorf("expected query in log, got: %s", logOutput)
	}
}

func TestLogger_LogsErrorStatusAtErrorLevel(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := RequestID(Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})))

	req := httptest.NewRequest(http.MethodPost, "/api/error", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, `"level":"ERROR"`) {
		t.Errorf("expected ERROR level for 500 status, got: %s", logOutput)
	}
}

func TestLogger_LogsClientErrorAtWarnLevel(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := RequestID(Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/bad", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, `"level":"WARN"`) {
		t.Errorf("expected WARN level for 400 status, got: %s", logOutput)
	}
}

func TestRecovery_RecoversFromPanic(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := RequestID(Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/panic", nil)
	req.Header.Set("X-Request-ID", "panic-test-1")
	rec := httptest.NewRecorder()

	// Should not panic — it should recover and return 500.
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "internal server error") {
		t.Errorf("expected error JSON body, got: %s", body)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "panic recovered") {
		t.Errorf("expected log message 'panic recovered', got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "panic-test-1") {
		t.Errorf("expected request_id in panic log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "something went wrong") {
		t.Errorf("expected panic value in log, got: %s", logOutput)
	}
}

func TestRecovery_NoPanicPassesThrough(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := RequestID(Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("all good"))
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/ok", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "all good" {
		t.Errorf("expected 'all good', got '%s'", rec.Body.String())
	}
	// No panic log should be present.
	if strings.Contains(buf.String(), "panic recovered") {
		t.Error("unexpected panic recovered log for normal request")
	}
}

func TestCORS_SetsHeaders(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposedHeaders: []string{"X-Request-ID"},
		MaxAge:         600,
	}

	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if !strings.Contains(origin, "http://localhost:5173") {
		t.Errorf("expected origin to contain localhost:5173, got: %s", origin)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "GET") {
		t.Errorf("expected methods to contain GET, got: %s", methods)
	}

	headers := rec.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(headers, "Content-Type") {
		t.Errorf("expected headers to contain Content-Type, got: %s", headers)
	}

	exposed := rec.Header().Get("Access-Control-Expose-Headers")
	if !strings.Contains(exposed, "X-Request-ID") {
		t.Errorf("expected exposed headers to contain X-Request-ID, got: %s", exposed)
	}

	maxAge := rec.Header().Get("Access-Control-Max-Age")
	if maxAge != "600" {
		t.Errorf("expected Max-Age 600, got: %s", maxAge)
	}
}

func TestCORS_HandlesPreflight(t *testing.T) {
	config := DefaultCORSConfig()

	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for preflight, got %d", rec.Code)
	}
}

func TestCORS_AllowCredentials(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}

	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	creds := rec.Header().Get("Access-Control-Allow-Credentials")
	if creds != "true" {
		t.Errorf("expected Allow-Credentials true, got: %s", creds)
	}
}

func TestCORS_DefaultConfig(t *testing.T) {
	config := DefaultCORSConfig()
	if len(config.AllowedMethods) < 3 {
		t.Error("default config should have at least 3 allowed methods")
	}
	if len(config.AllowedHeaders) < 3 {
		t.Error("default config should have at least 3 allowed headers")
	}
	if config.MaxAge != 300 {
		t.Errorf("expected MaxAge 300, got %d", config.MaxAge)
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	config := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	}

	handler := CORS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected origin '*', got: %s", origin)
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty", nil, ""},
		{"single", []string{"a"}, "a"},
		{"multiple", []string{"a", "b", "c"}, "a, b, c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinStrings(tt.input)
			if got != tt.expected {
				t.Errorf("joinStrings(%v) = '%s', want '%s'", tt.input, got, tt.expected)
			}
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 300, "300"},
		{"negative", -1, "-1"},
		{"large", 86400, "86400"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := itoa(tt.input)
			if got != tt.expected {
				t.Errorf("itoa(%d) = '%s', want '%s'", tt.input, got, tt.expected)
			}
		})
	}
}

// TestFullMiddlewareChain verifies all middleware work together correctly.
func TestFullMiddlewareChain(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := RequestID(
		Recovery(logger)(
			Logger(logger)(
				CORS(DefaultCORSConfig())(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						id := GetRequestID(r.Context())
						if id == "" {
							t.Error("expected request ID in context, got empty")
						}
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"status":"ok"}`))
					}),
				),
			),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "chain-test-1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") != "chain-test-1" {
		t.Errorf("expected X-Request-ID in response, got '%s'", rec.Header().Get("X-Request-ID"))
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS origin '*', got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "chain-test-1") {
		t.Errorf("expected request_id in log: %s", logOutput)
	}
}

// TestNoOp logger for benchmarks — suppresses output.
var noopLogger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

func BenchmarkMiddlewareChain(b *testing.B) {
	handler := RequestID(
		Recovery(noopLogger)(
			Logger(noopLogger)(
				CORS(DefaultCORSConfig())(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}),
				),
			),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-Request-ID", "bench-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
