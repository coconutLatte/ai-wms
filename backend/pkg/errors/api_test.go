package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// ── ValidationError Tests ──────────────────────────────────────────────────────────────

func TestValidationError_Add(t *testing.T) {
	ve := NewValidationError()
	ve.Add("code", "warehouse code is required")
	ve.Add("name", "warehouse name is required")

	if !ve.HasErrors() {
		t.Fatal("expected HasErrors to be true")
	}
	if len(ve.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(ve.Fields))
	}
	if ve.Fields[0].Field != "code" {
		t.Errorf("expected field 'code', got %q", ve.Fields[0].Field)
	}
	if ve.Fields[1].Field != "name" {
		t.Errorf("expected field 'name', got %q", ve.Fields[1].Field)
	}
}

func TestValidationError_AddWithCode(t *testing.T) {
	ve := NewValidationError()
	ve.AddWithCode("qty", "must be positive", "POSITIVE_REQUIRED")

	if ve.Fields[0].Code != "POSITIVE_REQUIRED" {
		t.Errorf("expected code 'POSITIVE_REQUIRED', got %q", ve.Fields[0].Code)
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := NewValidationError()
	ve.Add("code", "warehouse code is required")
	ve.Add("name", "warehouse name is required")

	errStr := ve.Error()
	expectedContain := []string{"validation failed", "code: warehouse code is required", "name: warehouse name is required"}
	for _, s := range expectedContain {
		if !contains(errStr, s) {
			t.Errorf("expected error string to contain %q, got %q", s, errStr)
		}
	}
}

func TestValidationError_HasErrors_Empty(t *testing.T) {
	ve := NewValidationError()
	if ve.HasErrors() {
		t.Error("expected HasErrors to be false for empty validation error")
	}
}

// ── StatusCode Tests ───────────────────────────────────────────────────────────────────

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"nil error", nil, http.StatusOK},
		{"not found", NewNotFound("warehouse", "wh-001"), http.StatusNotFound},
		{"invalid input", NewInvalidInput("name is required"), http.StatusBadRequest},
		{"validation error", func() error {
			ve := NewValidationError()
			ve.Add("code", "required")
			return ve
		}(), http.StatusBadRequest},
		{"invalid status", NewInvalidStatus("draft", "cancelled"), http.StatusUnprocessableEntity},
		{"insufficient qty", NewInsufficientQty("sku-1", 10, 5), http.StatusUnprocessableEntity},
		{"already exists", NewAlreadyExists("warehouse", "wh-001"), http.StatusConflict},
		{"conflict", NewConflict("resource is locked"), http.StatusConflict},
		{"location occupied", NewLocationOccupied("loc-1"), http.StatusConflict},
		{"location full", NewLocationFull("loc-1", 100), http.StatusUnprocessableEntity},
		{"internal", NewInternal("db connection failed"), http.StatusInternalServerError},
		{"sentinel not found", ErrNotFound, http.StatusNotFound},
		{"sentinel invalid input", ErrInvalidInput, http.StatusBadRequest},
		{"sentinel already exists", ErrAlreadyExists, http.StatusConflict},
		{"sentinel unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"sentinel forbidden", ErrForbidden, http.StatusForbidden},
		{"sentinel conflict", ErrConflict, http.StatusConflict},
		{"sentinel internal", ErrInternal, http.StatusInternalServerError},
		{"sentinel insufficient qty", ErrInsufficientQty, http.StatusUnprocessableEntity},
		{"sentinel invalid status", ErrInvalidStatus, http.StatusUnprocessableEntity},
		{"wrapped not found", &DomainError{Err: ErrNotFound, Code: "NOT_FOUND"}, http.StatusNotFound},
		{"unknown error", errors.New("something went wrong"), http.StatusInternalServerError},
		{"wrapped unknown", &DomainError{Err: errors.New("unknown"), Code: "CUSTOM"}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusCode(tt.err)
			if got != tt.wantStatus {
				t.Errorf("StatusCode() = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

// ── ProblemDetail Tests ────────────────────────────────────────────────────────────────

func TestNewProblemDetail_NotFound(t *testing.T) {
	err := NewNotFound("warehouse", "wh-001")
	pd := NewProblemDetail(err, "/api/v1/warehouses/wh-001")

	if pd.Status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", pd.Status)
	}
	if pd.Code != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got %q", pd.Code)
	}
	if pd.Title != "Resource Not Found" {
		t.Errorf("expected title 'Resource Not Found', got %q", pd.Title)
	}
	if pd.Instance != "/api/v1/warehouses/wh-001" {
		t.Errorf("expected instance '/api/v1/warehouses/wh-001', got %q", pd.Instance)
	}
	if !contains(pd.Detail, "wh-001") {
		t.Errorf("expected detail to contain 'wh-001', got %q", pd.Detail)
	}
	if !contains(pd.Type, "not-found") {
		t.Errorf("expected type to contain 'not-found', got %q", pd.Type)
	}
}

func TestNewProblemDetail_InvalidInput(t *testing.T) {
	err := NewInvalidInput("warehouse code is required")
	pd := NewProblemDetail(err, "/api/v1/warehouses")

	if pd.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", pd.Status)
	}
	if pd.Code != "INVALID_INPUT" {
		t.Errorf("expected code 'INVALID_INPUT', got %q", pd.Code)
	}
	if pd.Title != "Invalid Input" {
		t.Errorf("expected title 'Invalid Input', got %q", pd.Title)
	}
	if pd.Detail != "warehouse code is required" {
		t.Errorf("expected detail 'warehouse code is required', got %q", pd.Detail)
	}
}

func TestNewProblemDetail_ValidationErrors(t *testing.T) {
	ve := NewValidationError()
	ve.Add("code", "warehouse code is required")
	ve.AddWithCode("name", "warehouse name is required", "REQUIRED")

	pd := NewProblemDetail(ve, "/api/v1/warehouses")

	if pd.Status != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", pd.Status)
	}
	if len(pd.Errors) != 2 {
		t.Fatalf("expected 2 validation errors, got %d", len(pd.Errors))
	}
	if pd.Errors[0].Field != "code" {
		t.Errorf("expected field 'code', got %q", pd.Errors[0].Field)
	}
	if pd.Errors[1].Code != "REQUIRED" {
		t.Errorf("expected code 'REQUIRED', got %q", pd.Errors[1].Code)
	}
}

func TestNewProblemDetail_Internal(t *testing.T) {
	err := NewInternal("db connection pool exhausted")
	pd := NewProblemDetail(err, "/api/v1/inventory")

	if pd.Status != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", pd.Status)
	}
	// Internal errors should NOT leak implementation details.
	if pd.Detail == "db connection pool exhausted" {
		t.Error("internal error detail should not leak implementation details")
	}
	if !contains(pd.Detail, "unexpected") && !contains(pd.Detail, "try again") {
		t.Errorf("expected generic error detail, got %q", pd.Detail)
	}
}

func TestNewProblemDetail_NilError(t *testing.T) {
	pd := NewProblemDetail(nil, "/test")
	if pd != nil {
		t.Error("expected nil ProblemDetail for nil error")
	}
}

func TestProblemDetail_ToJSON(t *testing.T) {
	err := NewNotFound("zone", "zone-001")
	pd := NewProblemDetail(err, "/api/v1/zones/zone-001")

	data, jsonErr := pd.ToJSON()
	if jsonErr != nil {
		t.Fatalf("ToJSON() error: %v", jsonErr)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify all RFC 7807 fields are present.
	requiredFields := []string{"type", "title", "status", "detail", "instance", "code"}
	for _, field := range requiredFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("expected field %q in JSON output", field)
		}
	}

	if int(decoded["status"].(float64)) != http.StatusNotFound {
		t.Errorf("expected status 404 in JSON, got %v", decoded["status"])
	}
}

func TestProblemDetail_JSONSerializationRoundTrip(t *testing.T) {
	err := NewInvalidStatus("draft", "completed")
	pd := NewProblemDetail(err, "/api/v1/orders/order-1/status")

	data, _ := pd.ToJSON()

	var decoded ProblemDetail
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Status != http.StatusUnprocessableEntity {
		t.Errorf("round-trip: expected status 422, got %d", decoded.Status)
	}
	if decoded.Code != "INVALID_STATUS" {
		t.Errorf("round-trip: expected code 'INVALID_STATUS', got %q", decoded.Code)
	}
}

// ── New Helper Constructor Tests ───────────────────────────────────────────────────────

func TestNewAlreadyExists(t *testing.T) {
	err := NewAlreadyExists("warehouse", "wh-001")
	var de *DomainError
	if !AsDomainError(err, &de) {
		t.Fatal("expected *DomainError")
	}
	if de.Code != "ALREADY_EXISTS" {
		t.Errorf("expected code 'ALREADY_EXISTS', got %q", de.Code)
	}
	if !Is(err, ErrAlreadyExists) {
		t.Error("expected error to wrap ErrAlreadyExists")
	}
}

func TestNewConflict(t *testing.T) {
	err := NewConflict("resource is locked by another operation")
	var de *DomainError
	if !AsDomainError(err, &de) {
		t.Fatal("expected *DomainError")
	}
	if de.Code != "CONFLICT" {
		t.Errorf("expected code 'CONFLICT', got %q", de.Code)
	}
	if !Is(err, ErrConflict) {
		t.Error("expected error to wrap ErrConflict")
	}
}

func TestNewLocationOccupied(t *testing.T) {
	err := NewLocationOccupied("loc-001")
	if !Is(err, ErrLocationOccupied) {
		t.Error("expected error to wrap ErrLocationOccupied")
	}
}

func TestNewLocationFull(t *testing.T) {
	err := NewLocationFull("loc-001", 100.0)
	if !Is(err, ErrLocationFull) {
		t.Error("expected error to wrap ErrLocationFull")
	}
}

func TestNewInternal(t *testing.T) {
	err := NewInternal("something broke")
	if !Is(err, ErrInternal) {
		t.Error("expected error to wrap ErrInternal")
	}
	// Internal errors should have the INTERNAL_ERROR code.
	var de *DomainError
	if AsDomainError(err, &de) {
		if de.Code != "INTERNAL_ERROR" {
			t.Errorf("expected code 'INTERNAL_ERROR', got %q", de.Code)
		}
	}
}

// ── Is / IsInvalidInput / IsInvalidStatus Tests ────────────────────────────────────────

func TestIsInvalidInput_Helper(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{"direct sentinel", ErrInvalidInput, true},
		{"NewInvalidInput", NewInvalidInput("bad"), true},
		{"wrapped", &DomainError{Err: ErrInvalidInput, Code: "INVALID_INPUT"}, true},
		{"not found", ErrNotFound, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidInput(tt.err); got != tt.expect {
				t.Errorf("IsInvalidInput() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestIsInvalidStatus_Helper(t *testing.T) {
	if !IsInvalidStatus(ErrInvalidStatus) {
		t.Error("expected IsInvalidStatus to return true for ErrInvalidStatus")
	}
	if IsInvalidStatus(ErrNotFound) {
		t.Error("expected IsInvalidStatus to return false for ErrNotFound")
	}
}

func TestIs_Helper(t *testing.T) {
	err := NewNotFound("warehouse", "wh-001")
	if !Is(err, ErrNotFound) {
		t.Error("Is should match through wrapping")
	}
	if Is(err, ErrAlreadyExists) {
		t.Error("Is should not match unrelated error")
	}
}

// ── AsDomainError / AsValidationError Tests ────────────────────────────────────────────

func TestAsDomainError_Helper(t *testing.T) {
	err := NewNotFound("zone", "z-1")
	var de *DomainError
	if !AsDomainError(err, &de) {
		t.Fatal("expected AsDomainError to return true")
	}
	if de.Code != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got %q", de.Code)
	}

	// Should fail for non-DomainError.
	plain := errors.New("plain error")
	var de2 *DomainError
	if AsDomainError(plain, &de2) {
		t.Error("expected AsDomainError to return false for plain error")
	}
}

func TestAsValidationError_Helper(t *testing.T) {
	ve := NewValidationError()
	ve.Add("field", "error")
	var extracted *ValidationError
	if !AsValidationError(ve, &extracted) {
		t.Fatal("expected AsValidationError to return true")
	}
	if len(extracted.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(extracted.Fields))
	}

	// Should fail for non-ValidationError.
	plain := errors.New("plain error")
	var extracted2 *ValidationError
	if AsValidationError(plain, &extracted2) {
		t.Error("expected AsValidationError to return false for plain error")
	}
}

// ── extractCode Tests ─────────────────────────────────────────────────────────────────

func TestExtractCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{"domain error with code", NewNotFound("w", "1"), "NOT_FOUND"},
		{"domain error without code", &DomainError{Err: ErrNotFound}, "NOT_FOUND"},
		{"validation error", func() error { ve := NewValidationError(); ve.Add("f", "m"); return ve }(), "INVALID_INPUT"},
		{"sentinel not found", ErrNotFound, "NOT_FOUND"},
		{"sentinel conflict", ErrConflict, "CONFLICT"},
		{"sentinel internal", ErrInternal, "INTERNAL_ERROR"},
		{"unknown error", errors.New("unknown"), "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("extractCode() = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

// ── extractDetail Tests ────────────────────────────────────────────────────────────────

func TestExtractDetail_InternalHiding(t *testing.T) {
	err := NewInternal("secret db password leaked")
	detail := extractDetail(err, http.StatusInternalServerError)
	if detail == "secret db password leaked" {
		t.Error("extractDetail should hide internal error details")
	}
}

func TestExtractDetail_ValidationAggregation(t *testing.T) {
	ve := NewValidationError()
	ve.Add("code", "required")
	ve.Add("name", "too short")
	detail := extractDetail(ve, http.StatusBadRequest)
	if !contains(detail, "required") || !contains(detail, "too short") {
		t.Errorf("extractDetail should aggregate validation messages, got %q", detail)
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────────────────

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
