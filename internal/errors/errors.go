package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation_error"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "not_found_error"
	// ErrorTypeConflict represents conflict errors
	ErrorTypeConflict ErrorType = "conflict_error"
	// ErrorTypeRateLimit represents rate limit errors
	ErrorTypeRateLimit ErrorType = "rate_limit_error"
	// ErrorTypeInternal represents internal errors
	ErrorTypeInternal ErrorType = "internal_error"
	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout_error"
)

// BotError represents a bot-specific error
type BotError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *BotError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *BotError) Unwrap() error {
	return e.Err
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	var botErr *BotError
	if err == nil {
		return false
	}
	if ok := errors.As(err, &botErr); !ok {
		return false
	}
	return botErr.Type == ErrorTypeValidation
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	var botErr *BotError
	if err == nil {
		return false
	}
	if ok := errors.As(err, &botErr); !ok {
		return false
	}
	return botErr.Type == ErrorTypeNotFound
}

// IsConflictError checks if the error is a conflict error
func IsConflictError(err error) bool {
	var botErr *BotError
	if err == nil {
		return false
	}
	if ok := errors.As(err, &botErr); !ok {
		return false
	}
	return botErr.Type == ErrorTypeConflict
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	var botErr *BotError
	if err == nil {
		return false
	}
	if ok := errors.As(err, &botErr); !ok {
		return false
	}
	return botErr.Type == ErrorTypeRateLimit
}

// IsInternalError checks if the error is an internal error
func IsInternalError(err error) bool {
	var botErr *BotError
	if err == nil {
		return false
	}
	if ok := errors.As(err, &botErr); !ok {
		return false
	}
	return botErr.Type == ErrorTypeInternal
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	var botErr *BotError
	if err == nil {
		return false
	}
	if ok := errors.As(err, &botErr); !ok {
		return false
	}
	return botErr.Type == ErrorTypeTimeout
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) error {
	return &BotError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) error {
	return &BotError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Err:     err,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string, err error) error {
	return &BotError{
		Type:    ErrorTypeConflict,
		Message: message,
		Err:     err,
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string, err error) error {
	return &BotError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Err:     err,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) error {
	return &BotError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, err error) error {
	return &BotError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Err:     err,
	}
} 