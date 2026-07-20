// Package errors provides domain-specific error types and API-layer error formatting.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ── RFC 7807 Problem Details ─────────────────────────────────────────────────────────────

// ProblemDetail represents an RFC 7807 Problem Details response.
// https://datatracker.ietf.org/doc/html/rfc7807
type ProblemDetail struct {
	// Type is a URI reference identifying the problem type.
	Type string `json:"type"`
	// Title is a short, human-readable summary of the problem.
	Title string `json:"title"`
	// Status is the HTTP status code.
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail"`
	// Instance is a URI reference identifying the specific occurrence (e.g., the request path).
	Instance string `json:"instance,omitempty"`
	// Code is an application-specific error code.
	Code string `json:"code,omitempty"`
	// Errors is a list of validation errors (only present for validation failures).
	Errors []ValidationField `json:"errors,omitempty"`
}

// ── Validation Errors ────────────────────────────────────────────────────────────────────

// ValidationField represents a single field-level validation error.
type ValidationField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ValidationError contains one or more field-level validation errors.
type ValidationError struct {
	Fields []ValidationField
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	msgs := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		if f.Code != "" {
			msgs[i] = fmt.Sprintf("%s: %s (%s)", f.Field, f.Message, f.Code)
		} else {
			msgs[i] = fmt.Sprintf("%s: %s", f.Field, f.Message)
		}
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

// Add adds a validation error for a field.
func (e *ValidationError) Add(field, message string) {
	e.Fields = append(e.Fields, ValidationField{Field: field, Message: message})
}

// AddWithCode adds a validation error for a field with an error code.
func (e *ValidationError) AddWithCode(field, message, code string) {
	e.Fields = append(e.Fields, ValidationField{Field: field, Message: message, Code: code})
}

// HasErrors returns true if there are any validation errors.
func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

// NewValidationError creates a new ValidationError.
func NewValidationError() *ValidationError {
	return &ValidationError{}
}

// ── Error → HTTP Status Mapping ──────────────────────────────────────────────────────────

// StatusCode returns the HTTP status code for a given domain error.
// This centralizes the error-to-status mapping so handlers don't need
// to know which status code maps to which error type.
func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Validation errors → 400 Bad Request
	var valErr *ValidationError
	if AsValidationError(err, &valErr) {
		return http.StatusBadRequest
	}

	// DomainError codes map to specific HTTP statuses.
	var domainErr *DomainError
	if AsDomainError(err, &domainErr) {
		return domainCodeToStatus(domainErr.Code)
	}

	// Sentinel error checks.
	switch {
	case IsNotFound(err):
		return http.StatusNotFound
	case IsInvalidInput(err):
		return http.StatusBadRequest
	case Is(err, ErrAlreadyExists):
		return http.StatusConflict
	case Is(err, ErrInsufficientQty):
		return http.StatusUnprocessableEntity
	case Is(err, ErrLocationOccupied):
		return http.StatusConflict
	case Is(err, ErrLocationFull):
		return http.StatusUnprocessableEntity
	case Is(err, ErrInvalidStatus):
		return http.StatusUnprocessableEntity
	case Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case Is(err, ErrForbidden):
		return http.StatusForbidden
	case Is(err, ErrConflict):
		return http.StatusConflict
	case Is(err, ErrInternal):
		return http.StatusInternalServerError
	}

	// Default: 500 Internal Server Error for unknown errors.
	return http.StatusInternalServerError
}

// domainCodeToStatus maps a DomainError code to an HTTP status code.
func domainCodeToStatus(code string) int {
	switch code {
	case "NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_INPUT":
		return http.StatusBadRequest
	case "INSUFFICIENT_QTY":
		return http.StatusUnprocessableEntity
	case "INVALID_STATUS":
		return http.StatusUnprocessableEntity
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "ALREADY_EXISTS":
		return http.StatusConflict
	case "CONFLICT":
		return http.StatusConflict
	case "LOCATION_OCCUPIED":
		return http.StatusConflict
	case "LOCATION_FULL":
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// AsDomainError extracts a *DomainError from an error chain.
func AsDomainError(err error, target **DomainError) bool {
	return As(err, target)
}

// AsValidationError extracts a *ValidationError from an error chain.
func AsValidationError(err error, target **ValidationError) bool {
	return As(err, target)
}

// ── Problem Detail Construction ──────────────────────────────────────────────────────────

// ProblemTypeBase is the base URI for problem types defined by this application.
const ProblemTypeBase = "https://api.ai-wms.io/problems/"

// problemTitles maps error codes to human-readable titles.
var problemTitles = map[string]string{
	"NOT_FOUND":          "Resource Not Found",
	"INVALID_INPUT":      "Invalid Input",
	"INSUFFICIENT_QTY":   "Insufficient Quantity",
	"INVALID_STATUS":     "Invalid Status Transition",
	"UNAUTHORIZED":       "Unauthorized",
	"FORBIDDEN":          "Forbidden",
	"ALREADY_EXISTS":     "Resource Already Exists",
	"CONFLICT":           "Resource Conflict",
	"LOCATION_OCCUPIED":  "Location Occupied",
	"LOCATION_FULL":      "Location Full",
}

// NewProblemDetail creates an RFC 7807 Problem Detail from an error.
// instance is typically the request URI path.
func NewProblemDetail(err error, instance string) *ProblemDetail {
	if err == nil {
		return nil
	}

	status := StatusCode(err)
	code := extractCode(err)
	detail := extractDetail(err, status)

	pd := &ProblemDetail{
		Type:     ProblemTypeBase + strings.ToLower(strings.ReplaceAll(code, "_", "-")),
		Title:    problemTitle(code),
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Code:     code,
	}

	// Attach validation errors if present.
	var valErr *ValidationError
	if AsValidationError(err, &valErr) && valErr.HasErrors() {
		pd.Errors = valErr.Fields
	}

	return pd
}

// ToJSON serializes the ProblemDetail to JSON bytes.
func (pd *ProblemDetail) ToJSON() ([]byte, error) {
	return json.Marshal(pd)
}

// problemTitle returns a human-readable title for an error code.
func problemTitle(code string) string {
	if title, ok := problemTitles[code]; ok {
		return title
	}
	return "Internal Server Error"
}

// extractCode extracts an application error code from an error chain.
func extractCode(err error) string {
	var domainErr *DomainError
	if AsDomainError(err, &domainErr) && domainErr.Code != "" {
		return domainErr.Code
	}

	var valErr *ValidationError
	if AsValidationError(err, &valErr) {
		return "INVALID_INPUT"
	}

	switch {
	case IsNotFound(err):
		return "NOT_FOUND"
	case Is(err, ErrInvalidInput):
		return "INVALID_INPUT"
	case Is(err, ErrAlreadyExists):
		return "ALREADY_EXISTS"
	case Is(err, ErrInsufficientQty):
		return "INSUFFICIENT_QTY"
	case Is(err, ErrLocationOccupied):
		return "LOCATION_OCCUPIED"
	case Is(err, ErrLocationFull):
		return "LOCATION_FULL"
	case Is(err, ErrInvalidStatus):
		return "INVALID_STATUS"
	case Is(err, ErrUnauthorized):
		return "UNAUTHORIZED"
	case Is(err, ErrForbidden):
		return "FORBIDDEN"
	case Is(err, ErrConflict):
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}

// extractDetail returns a user-facing detail message from the error.
// For 5xx errors it returns a generic message to avoid leaking internals.
func extractDetail(err error, status int) string {
	if status >= 500 {
		return "An unexpected error occurred. Please try again later or contact support."
	}

	// For validation errors, aggregate field messages.
	var valErr *ValidationError
	if AsValidationError(err, &valErr) && valErr.HasErrors() {
		msgs := make([]string, len(valErr.Fields))
		for i, f := range valErr.Fields {
			msgs[i] = f.Message
		}
		return strings.Join(msgs, "; ")
	}

	// Unwrap to get the innermost message for user-facing detail.
	inner := err
	for {
		unwrapped := Unwrap(inner)
		if unwrapped == nil {
			break
		}
		inner = unwrapped
	}

	msg := inner.Error()
	// Strip the sentinel prefix from domain errors for cleaner output.
	var domainErr *DomainError
	if AsDomainError(err, &domainErr) {
		return domainErr.Message
	}

	return msg
}

// Unwrap is a helper that unwraps an error (wrapping errors.Unwrap for convenience).
func Unwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}
	u, ok := err.(unwrapper)
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// As is a helper wrapping errors.As for convenience.
func As(err error, target interface{}) bool {
	type aser interface {
		As(target interface{}) bool
	}
	// Use standard library's errors.As through a simple type assertion approach.
	// We need to import "errors" but we're already in the errors package.
	// Let's just use the standard approach.
	return asImpl(err, target)
}

// asImpl implements errors.As using the reflect-free approach.
func asImpl(err error, target interface{}) bool {
	// We match the *ValidationError or *DomainError by type assertion through the chain.
	switch t := target.(type) {
	case **ValidationError:
		for {
			if ve, ok := err.(*ValidationError); ok {
				*t = ve
				return true
			}
			u, ok := err.(interface{ Unwrap() error })
			if !ok {
				return false
			}
			err = u.Unwrap()
		}
	case **DomainError:
		for {
			if de, ok := err.(*DomainError); ok {
				*t = de
				return true
			}
			u, ok := err.(interface{ Unwrap() error })
			if !ok {
				return false
			}
			err = u.Unwrap()
		}
	}
	return false
}
