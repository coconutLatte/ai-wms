// Package errors provides domain-specific error types for the WMS application.
package errors

import (
	"errors"
	"fmt"
)

// Common sentinel errors.
var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrInsufficientQty  = errors.New("insufficient inventory quantity")
	ErrLocationOccupied = errors.New("location is already occupied")
	ErrLocationFull     = errors.New("location capacity exceeded")
	ErrInvalidStatus    = errors.New("invalid status transition")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("resource conflict")
	ErrInternal         = errors.New("internal error")
)

// DomainError wraps a sentinel error with context.
type DomainError struct {
	Err     error
	Message string
	Code    string
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Err.Error(), e.Message)
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewNotFound creates a not found error with context.
func NewNotFound(resource string, id string) error {
	return &DomainError{
		Err:     ErrNotFound,
		Message: fmt.Sprintf("%s with id=%s not found", resource, id),
		Code:    "NOT_FOUND",
	}
}

// NewInvalidInput creates an invalid input error.
func NewInvalidInput(msg string) error {
	return &DomainError{
		Err:     ErrInvalidInput,
		Message: msg,
		Code:    "INVALID_INPUT",
	}
}

// NewInsufficientQty creates an insufficient quantity error.
func NewInsufficientQty(skuID string, requested, available float64) error {
	return &DomainError{
		Err:     ErrInsufficientQty,
		Message: fmt.Sprintf("SKU %s: requested %.2f, available %.2f", skuID, requested, available),
		Code:    "INSUFFICIENT_QTY",
	}
}

// NewInvalidStatus creates an invalid status transition error.
func NewInvalidStatus(current, target string) error {
	return &DomainError{
		Err:     ErrInvalidStatus,
		Message: fmt.Sprintf("cannot transition from %s to %s", current, target),
		Code:    "INVALID_STATUS",
	}
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInsufficientQty checks if an error is an insufficient quantity error.
func IsInsufficientQty(err error) bool {
	return errors.Is(err, ErrInsufficientQty)
}
