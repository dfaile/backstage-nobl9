package errors

import (
	"errors"
	"testing"
	"backstage-nobl9/internal/errors"
)

func TestBotError(t *testing.T) {
	// Test error without underlying error
	err := NewValidationError("invalid input", nil)
	if err.Error() != "validation_error: invalid input" {
		t.Errorf("Expected error message 'validation_error: invalid input', got '%s'", err.Error())
	}

	// Test error with underlying error
	underlyingErr := errors.New("underlying error")
	err = NewValidationError("invalid input", underlyingErr)
	if err.Error() != "validation_error: invalid input (underlying error)" {
		t.Errorf("Expected error message 'validation_error: invalid input (underlying error)', got '%s'", err.Error())
	}

	// Test error unwrapping
	var botErr *BotError
	if !errors.As(err, &botErr) {
		t.Error("Expected error to be a BotError")
	}
	if botErr.Type != ErrorTypeValidation {
		t.Errorf("Expected error type 'validation_error', got '%s'", botErr.Type)
	}
	if botErr.Message != "invalid input" {
		t.Errorf("Expected error message 'invalid input', got '%s'", botErr.Message)
	}
	if botErr.Err != underlyingErr {
		t.Error("Expected underlying error to match")
	}
}

func TestErrorTypeChecks(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "validation error",
			err:      NewValidationError("test", nil),
			checkFn:  IsValidationError,
			expected: true,
		},
		{
			name:     "not found error",
			err:      NewNotFoundError("test", nil),
			checkFn:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "conflict error",
			err:      NewConflictError("test", nil),
			checkFn:  IsConflictError,
			expected: true,
		},
		{
			name:     "rate limit error",
			err:      NewRateLimitError("test", nil),
			checkFn:  IsRateLimitError,
			expected: true,
		},
		{
			name:     "internal error",
			err:      NewInternalError("test", nil),
			checkFn:  IsInternalError,
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			checkFn:  IsValidationError,
			expected: false,
		},
		{
			name:     "non-bot error",
			err:      errors.New("test"),
			checkFn:  IsValidationError,
			expected: false,
		},
		{
			name:     "wrong error type",
			err:      NewValidationError("test", nil),
			checkFn:  IsNotFoundError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkFn(tt.err); got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Create a chain of errors
	err1 := errors.New("error 1")
	err2 := NewValidationError("error 2", err1)
	err3 := NewInternalError("error 3", err2)

	// Test unwrapping
	var botErr *BotError
	if !errors.As(err3, &botErr) {
		t.Error("Expected error to be a BotError")
	}
	if botErr.Type != ErrorTypeInternal {
		t.Errorf("Expected error type 'internal_error', got '%s'", botErr.Type)
	}

	// Test unwrapping to get the validation error
	if !errors.As(err3, &botErr) {
		t.Error("Expected error to be a BotError")
	}
	if !IsValidationError(botErr.Err) {
		t.Error("Expected underlying error to be a validation error")
	}

	// Test unwrapping to get the original error
	if !errors.Is(err3, err1) {
		t.Error("Expected error to wrap the original error")
	}
} 