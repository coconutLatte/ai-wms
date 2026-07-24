// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSwaggerSpecEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	// Test GET /api/docs/swagger.json
	req := httptest.NewRequest("GET", "/api/docs/swagger.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Verify we get non-empty JSON body
	body := rec.Body.Bytes()
	if len(body) < 100 {
		t.Errorf("expected substantial JSON body, got %d bytes", len(body))
	}

	// Verify it starts with a JSON object
	if body[0] != '{' {
		t.Errorf("expected JSON object starting with '{', got '%c'", body[0])
	}
}

func TestSwaggerUIEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	// Test GET /api/docs (Swagger UI HTML page)
	req := httptest.NewRequest("GET", "/api/docs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}

	// Verify it contains Swagger UI HTML
	body := rec.Body.String()
	if body == "" {
		t.Error("expected non-empty HTML body")
	}

	// Check for key Swagger UI identifiers
	if !contains(body, "SwaggerUIBundle") {
		t.Error("Swagger UI page should contain 'SwaggerUIBundle'")
	}
	if !contains(body, "swagger-ui") {
		t.Error("Swagger UI page should contain 'swagger-ui'")
	}
}

func TestSwaggerCORSEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	// Test CORS header on spec endpoint
	req := httptest.NewRequest("GET", "/api/docs/swagger.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	cors := resp.Header.Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("expected CORS header '*', got '%s'", cors)
	}
}

func TestSwaggerUIAccessibleWithoutAuth(t *testing.T) {
	// Verify that /api/docs and /api/docs/swagger.json don't require auth.
	// These routes are registered directly on the mux, not through auth middleware.
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	// Without Authorization header
	req := httptest.NewRequest("GET", "/api/docs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("swagger UI should be accessible without auth, got %d", rec.Code)
	}

	req2 := httptest.NewRequest("GET", "/api/docs/swagger.json", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("swagger spec should be accessible without auth, got %d", rec2.Code)
	}
}

func TestSwaggerSpecContainsExpectedTags(t *testing.T) {
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	req := httptest.NewRequest("GET", "/api/docs/swagger.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Verify the spec contains expected tags for all documented resource groups
	expectedTags := []string{
		"warehouses",
		"zones",
		"locations",
		"skus",
		"inventory",
		"orders",
		"asns",
		"tasks",
		"waves",
		"shipments",
		"cycle-counts",
		"dashboard",
		"auth",
		"users",
		"roles",
		"audit-logs",
		"settings",
	}

	for _, tag := range expectedTags {
		if !contains(body, `"name": "`+tag+`"`) {
			t.Errorf("swagger spec should contain tag '%s'", tag)
		}
	}
}

func TestSwaggerSpecIsValidJSON(t *testing.T) {
	// The spec should be parseable as valid JSON.
	// We just check it starts with { and has the openapi key.
	mux := http.NewServeMux()
	RegisterSwaggerRoutes(mux)

	req := httptest.NewRequest("GET", "/api/docs/swagger.json", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !contains(body, `"openapi"`) {
		t.Error("swagger spec should contain 'openapi' version field")
	}
	if !contains(body, `"paths"`) {
		t.Error("swagger spec should contain 'paths' section")
	}
	if !contains(body, `"components"`) {
		t.Error("swagger spec should contain 'components' section")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
