package recovery

import (
	"fmt"
	"time"

	"github.com/dfaile/backstage-nobl9/internal/errors"
)

// Strategy defines the type of recovery strategy
type Strategy string

const (
	// StrategyRetry indicates a retry strategy
	StrategyRetry Strategy = "retry"
	// StrategyFallback indicates a fallback strategy
	StrategyFallback Strategy = "fallback"
	// StrategyCancel indicates a cancellation strategy
	StrategyCancel Strategy = "cancel"
)

// Recovery defines a recovery action
type Recovery struct {
	Strategy    Strategy
	MaxAttempts int
	Delay       time.Duration
	Message     string
}

// NewRecovery creates a new recovery action
func NewRecovery(strategy Strategy, maxAttempts int, delay time.Duration, message string) *Recovery {
	return &Recovery{
		Strategy:    strategy,
		MaxAttempts: maxAttempts,
		Delay:       delay,
		Message:     message,
	}
}

// Format returns a formatted message for the recovery action
func (r *Recovery) Format() string {
	switch r.Strategy {
	case StrategyRetry:
		return fmt.Sprintf("Retrying... (%d attempts remaining)", r.MaxAttempts)
	case StrategyFallback:
		return fmt.Sprintf("Using fallback strategy: %s", r.Message)
	case StrategyCancel:
		return fmt.Sprintf("Operation cancelled: %s", r.Message)
	default:
		return r.Message
	}
}

// GetRecoveryForError returns a recovery action for the given error
func GetRecoveryForError(err error) *Recovery {
	switch {
	case errors.IsRateLimitError(err):
		return NewRecovery(
			StrategyRetry,
			3,
			5*time.Second,
			"Rate limit exceeded, retrying...",
		)
	case errors.IsTimeoutError(err):
		return NewRecovery(
			StrategyRetry,
			2,
			2*time.Second,
			"Operation timed out, retrying...",
		)
	case errors.IsNotFoundError(err):
		return NewRecovery(
			StrategyFallback,
			1,
			0,
			"Resource not found, using default values",
		)
	case errors.IsConflictError(err):
		return NewRecovery(
			StrategyCancel,
			1,
			0,
			"Resource already exists",
		)
	case errors.IsValidationError(err):
		return NewRecovery(
			StrategyCancel,
			1,
			0,
			"Invalid input, please check your request",
		)
	default:
		return NewRecovery(
			StrategyCancel,
			1,
			0,
			"An unexpected error occurred",
		)
	}
}

// ShouldRetry determines if an operation should be retried
func ShouldRetry(err error, attempts int) bool {
	recovery := GetRecoveryForError(err)
	return recovery.Strategy == StrategyRetry && attempts < recovery.MaxAttempts
}

// GetRetryDelay returns the delay before the next retry attempt
func GetRetryDelay(err error) time.Duration {
	recovery := GetRecoveryForError(err)
	return recovery.Delay
} 