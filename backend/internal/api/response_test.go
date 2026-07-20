package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// ── WriteJSON Tests ───────────────────────────────────────────────────────────────────

func TestWriteJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	WriteJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	var decoded map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if decoded["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", decoded["status"])
	}
}

func TestWriteJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusNoContent, nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body for nil data, got %d bytes", w.Body.Len())
	}
}

func TestWriteCreated(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "123"}

	WriteCreated(w, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestWriteNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	WriteNoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

// ── WriteError Tests ──────────────────────────────────────────────────────────────────

func TestWriteError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/warehouses/not-found", nil)

	err := pkgerrors.NewNotFound("warehouse", "not-found")
	WriteError(w, r, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var pd pkgerrors.ProblemDetail
	if err := json.Unmarshal(w.Body.Bytes(), &pd); err != nil {
		t.Fatalf("failed to decode problem detail: %v", err)
	}

	if pd.Code != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got %q", pd.Code)
	}
	if pd.Status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", pd.Status)
	}
	if pd.Instance != "/api/v1/warehouses/not-found" {
		t.Errorf("expected instance with request path, got %q", pd.Instance)
	}
}

func TestWriteError_InvalidInput(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/warehouses", nil)

	err := pkgerrors.NewInvalidInput("warehouse code is required")
	WriteError(w, r, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var pd pkgerrors.ProblemDetail
	json.Unmarshal(w.Body.Bytes(), &pd)

	if pd.Code != "INVALID_INPUT" {
		t.Errorf("expected code 'INVALID_INPUT', got %q", pd.Code)
	}
}

func TestWriteError_InvalidStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/orders/order-1/status", nil)

	err := pkgerrors.NewInvalidStatus("draft", "cancelled")
	WriteError(w, r, err)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	var pd pkgerrors.ProblemDetail
	json.Unmarshal(w.Body.Bytes(), &pd)

	if pd.Code != "INVALID_STATUS" {
		t.Errorf("expected code 'INVALID_STATUS', got %q", pd.Code)
	}
}

func TestWriteError_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/warehouses", nil)

	err := pkgerrors.NewAlreadyExists("warehouse", "wh-001")
	WriteError(w, r, err)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

func TestWriteError_Internal(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/inventory", nil)

	err := pkgerrors.NewInternal("db connection pool exhausted")
	WriteError(w, r, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var pd pkgerrors.ProblemDetail
	json.Unmarshal(w.Body.Bytes(), &pd)

	// Internal errors should not leak details.
	if strings.Contains(pd.Detail, "db connection") {
		t.Error("internal error should not leak implementation details")
	}
}

func TestWriteError_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Should not panic on nil error — NewProblemDetail handles nil.
	WriteError(w, r, nil)

	// nil error returns nil ProblemDetail, which WriteError handles as fallback 500.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 for nil error, got %d", w.Code)
	}
}

func TestWriteError_PlainError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	err := errors.New("something went wrong")
	WriteError(w, r, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 for unknown error, got %d", w.Code)
	}

	var pd pkgerrors.ProblemDetail
	json.Unmarshal(w.Body.Bytes(), &pd)

	if pd.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code 'INTERNAL_ERROR', got %q", pd.Code)
	}
}

// ── ReadJSON Tests ────────────────────────────────────────────────────────────────────

func TestReadJSON_Success(t *testing.T) {
	body := strings.NewReader(`{"code": "WH-001", "name": "Main Warehouse"}`)
	r := httptest.NewRequest("POST", "/test", body)

	var input struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := ReadJSON(r, &input); err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}

	if input.Code != "WH-001" {
		t.Errorf("expected code 'WH-001', got %q", input.Code)
	}
	if input.Name != "Main Warehouse" {
		t.Errorf("expected name 'Main Warehouse', got %q", input.Name)
	}
}

func TestReadJSON_InvalidJSON(t *testing.T) {
	body := strings.NewReader(`{invalid json}`)
	r := httptest.NewRequest("POST", "/test", body)

	var input struct {
		Code string `json:"code"`
	}

	err := ReadJSON(r, &input)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !pkgerrors.IsInvalidInput(err) {
		t.Errorf("expected invalid input error, got %v", err)
	}
}

func TestReadJSON_UnknownFields(t *testing.T) {
	body := strings.NewReader(`{"code": "WH-001", "unknown_field": "value"}`)
	r := httptest.NewRequest("POST", "/test", body)

	var input struct {
		Code string `json:"code"`
	}

	err := ReadJSON(r, &input)
	if err == nil {
		t.Fatal("expected error for unknown fields with DisallowUnknownFields")
	}

	if !pkgerrors.IsInvalidInput(err) {
		t.Errorf("expected invalid input error, got %v", err)
	}
}

// ── PathUUID Tests ────────────────────────────────────────────────────────────────────

func TestPathUUID_Success(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	// Go 1.22+ path values need to be set manually since httptest doesn't do routing.
	r.SetPathValue("id", "550e8400-e29b-41d4-a716-446655440000")

	id, err := PathUUID(r, "id")
	if err != nil {
		t.Fatalf("PathUUID failed: %v", err)
	}
	if id.String() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected UUID: %s", id.String())
	}
}

func TestPathUUID_Missing(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	// No path value set.

	_, err := PathUUID(r, "id")
	if err == nil {
		t.Fatal("expected error for missing path parameter")
	}
	if !pkgerrors.IsInvalidInput(err) {
		t.Errorf("expected invalid input error, got %v", err)
	}
}

func TestPathUUID_InvalidFormat(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.SetPathValue("id", "not-a-uuid")

	_, err := PathUUID(r, "id")
	if err == nil {
		t.Fatal("expected error for invalid UUID format")
	}
	if !pkgerrors.IsInvalidInput(err) {
		t.Errorf("expected invalid input error, got %v", err)
	}
}

// ── QueryParam Tests ──────────────────────────────────────────────────────────────────

func TestQueryParam_WithValue(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?key=value", nil)
	got := QueryParam(r, "key", "default")
	if got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
}

func TestQueryParam_Default(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	got := QueryParam(r, "key", "default")
	if got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestQueryParam_EmptyUsesDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?key=", nil)
	got := QueryParam(r, "key", "default")
	if got != "default" {
		t.Errorf("expected 'default' for empty value, got %q", got)
	}
}

// ── QueryParamInt Tests ────────────────────────────────────────────────────────────────

func TestQueryParamInt_WithValue(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=25", nil)
	got := QueryParamInt(r, "limit", 50)
	if got != 25 {
		t.Errorf("expected 25, got %d", got)
	}
}

func TestQueryParamInt_Default(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	got := QueryParamInt(r, "limit", 50)
	if got != 50 {
		t.Errorf("expected default 50, got %d", got)
	}
}

func TestQueryParamInt_InvalidReturnsDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=abc", nil)
	got := QueryParamInt(r, "limit", 50)
	if got != 50 {
		t.Errorf("expected default 50 for invalid value, got %d", got)
	}
}

func TestQueryParamInt_NegativeReturnsDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=-1", nil)
	got := QueryParamInt(r, "limit", 50)
	if got != 50 {
		t.Errorf("expected default 50 for negative value, got %d", got)
	}
}
