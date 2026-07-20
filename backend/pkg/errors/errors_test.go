package errors

import (
	"errors"
	"testing"
)

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("warehouse", "wh-001")

	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		t.Fatal("expected error to be *DomainError")
	}

	if domainErr.Code != "NOT_FOUND" {
		t.Errorf("expected Code=%q, got %q", "NOT_FOUND", domainErr.Code)
	}

	if domainErr.Message != "warehouse with id=wh-001 not found" {
		t.Errorf("unexpected Message: %q", domainErr.Message)
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("expected error to wrap ErrNotFound")
	}
}

func TestNewInvalidInput(t *testing.T) {
	err := NewInvalidInput("name is required")

	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		t.Fatal("expected error to be *DomainError")
	}

	if domainErr.Code != "INVALID_INPUT" {
		t.Errorf("expected Code=%q, got %q", "INVALID_INPUT", domainErr.Code)
	}

	if domainErr.Message != "name is required" {
		t.Errorf("unexpected Message: %q", domainErr.Message)
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Error("expected error to wrap ErrInvalidInput")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		isNotFound bool
	}{
		{
			name:       "not found error from NewNotFound",
			err:        NewNotFound("warehouse", "wh-001"),
			isNotFound: true,
		},
		{
			name:       "plain ErrNotFound",
			err:        ErrNotFound,
			isNotFound: true,
		},
		{
			name:       "wrapped not found error",
			err:        &DomainError{Err: ErrNotFound, Message: "custom", Code: "NOT_FOUND"},
			isNotFound: true,
		},
		{
			name:       "invalid input error",
			err:        NewInvalidInput("bad request"),
			isNotFound: false,
		},
		{
			name:       "nil error",
			err:        nil,
			isNotFound: false,
		},
		{
			name:       "unrelated error",
			err:        errors.New("some other error"),
			isNotFound: false,
		},
		{
			name:       "ErrAlreadyExists is not not-found",
			err:        ErrAlreadyExists,
			isNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.isNotFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.isNotFound)
			}
		})
	}
}

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name    string
		de      DomainError
		wantStr string
	}{
		{
			name:    "with message",
			de:      DomainError{Err: ErrNotFound, Message: "warehouse with id=wh-001 not found", Code: "NOT_FOUND"},
			wantStr: "resource not found: warehouse with id=wh-001 not found",
		},
		{
			name:    "without message",
			de:      DomainError{Err: ErrNotFound, Code: "NOT_FOUND"},
			wantStr: "resource not found",
		},
		{
			name:    "invalid input error",
			de:      DomainError{Err: ErrInvalidInput, Message: "name is required", Code: "INVALID_INPUT"},
			wantStr: "invalid input: name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.de.Error()
			if got != tt.wantStr {
				t.Errorf("Error() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	wrapped := errors.New("wrapped error")
	de := &DomainError{
		Err:     wrapped,
		Message: "context",
		Code:    "TEST",
	}

	unwrapped := de.Unwrap()
	if unwrapped != wrapped {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrapped)
	}

	// Verify errors.Is works through the chain.
	if !errors.Is(de, wrapped) {
		t.Error("expected errors.Is to find the wrapped error")
	}
}
