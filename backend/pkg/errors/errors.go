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

// IsInvalidInput checks if an error is an invalid input error.
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsInvalidStatus checks if an error is an invalid status transition error.
func IsInvalidStatus(err error) bool {
	return errors.Is(err, ErrInvalidStatus)
}

// Is checks if err matches target using errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// NewAlreadyExists creates an already exists error.
func NewAlreadyExists(resource string, id string) error {
	return &DomainError{
		Err:     ErrAlreadyExists,
		Message: fmt.Sprintf("%s with id=%s already exists", resource, id),
		Code:    "ALREADY_EXISTS",
	}
}

// NewConflict creates a conflict error with context.
func NewConflict(msg string) error {
	return &DomainError{
		Err:     ErrConflict,
		Message: msg,
		Code:    "CONFLICT",
	}
}

// NewLocationOccupied creates a location occupied error.
func NewLocationOccupied(locationID string) error {
	return &DomainError{
		Err:     ErrLocationOccupied,
		Message: fmt.Sprintf("location %s is already occupied", locationID),
		Code:    "LOCATION_OCCUPIED",
	}
}

// NewLocationFull creates a location full error.
func NewLocationFull(locationID string, capacity float64) error {
	return &DomainError{
		Err:     ErrLocationFull,
		Message: fmt.Sprintf("location %s is full (capacity: %.2f)", locationID, capacity),
		Code:    "LOCATION_FULL",
	}
}

// NewInternal creates an internal error.
func NewInternal(msg string) error {
	return &DomainError{
		Err:     ErrInternal,
		Message: msg,
		Code:    "INTERNAL_ERROR",
	}
}
