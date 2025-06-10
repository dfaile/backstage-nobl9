package recovery_test

import (
	"testing"
	"time"

	"github.com/dfaile/backstage-nobl9/internal/errors"
)

func TestNewRecovery(t *testing.T) {
	recovery := NewRecovery(StrategyRetry, 3, 5*time.Second, "test message")
	if recovery.Strategy != StrategyRetry {
		t.Errorf("Expected strategy %v, got %v", StrategyRetry, recovery.Strategy)
	}
	if recovery.MaxAttempts != 3 {
		t.Errorf("Expected max attempts 3, got %d", recovery.MaxAttempts)
	}
	if recovery.Delay != 5*time.Second {
		t.Errorf("Expected delay 5s, got %v", recovery.Delay)
	}
	if recovery.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", recovery.Message)
	}
}

func TestRecoveryFormat(t *testing.T) {
	tests := []struct {
		name     string
		recovery *Recovery
		want     string
	}{
		{
			name: "retry strategy",
			recovery: NewRecovery(
				StrategyRetry,
				3,
				5*time.Second,
				"test message",
			),
			want: "Retrying... (3 attempts remaining)",
		},
		{
			name: "fallback strategy",
			recovery: NewRecovery(
				StrategyFallback,
				1,
				0,
				"test message",
			),
			want: "Using fallback strategy: test message",
		},
		{
			name: "cancel strategy",
			recovery: NewRecovery(
				StrategyCancel,
				1,
				0,
				"test message",
			),
			want: "Operation cancelled: test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.recovery.Format()
			if got != tt.want {
				t.Errorf("Recovery.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRecoveryForError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    Strategy
		attempts int
		delay   time.Duration
	}{
		{
			name:    "rate limit error",
			err:     errors.NewRateLimitError("rate limit exceeded"),
			want:    StrategyRetry,
			attempts: 3,
			delay:   5 * time.Second,
		},
		{
			name:    "timeout error",
			err:     errors.NewTimeoutError("operation timed out"),
			want:    StrategyRetry,
			attempts: 2,
			delay:   2 * time.Second,
		},
		{
			name:    "not found error",
			err:     errors.NewNotFoundError("resource not found"),
			want:    StrategyFallback,
			attempts: 1,
			delay:   0,
		},
		{
			name:    "conflict error",
			err:     errors.NewConflictError("resource already exists"),
			want:    StrategyCancel,
			attempts: 1,
			delay:   0,
		},
		{
			name:    "validation error",
			err:     errors.NewValidationError("invalid input"),
			want:    StrategyCancel,
			attempts: 1,
			delay:   0,
		},
		{
			name:    "unknown error",
			err:     errors.NewInternalError("unknown error"),
			want:    StrategyCancel,
			attempts: 1,
			delay:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recovery := GetRecoveryForError(tt.err)
			if recovery.Strategy != tt.want {
				t.Errorf("GetRecoveryForError() strategy = %v, want %v", recovery.Strategy, tt.want)
			}
			if recovery.MaxAttempts != tt.attempts {
				t.Errorf("GetRecoveryForError() attempts = %v, want %v", recovery.MaxAttempts, tt.attempts)
			}
			if recovery.Delay != tt.delay {
				t.Errorf("GetRecoveryForError() delay = %v, want %v", recovery.Delay, tt.delay)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		attempts int
		want     bool
	}{
		{
			name:     "rate limit error with attempts remaining",
			err:      errors.NewRateLimitError("rate limit exceeded"),
			attempts: 2,
			want:     true,
		},
		{
			name:     "rate limit error with no attempts remaining",
			err:      errors.NewRateLimitError("rate limit exceeded"),
			attempts: 3,
			want:     false,
		},
		{
			name:     "timeout error with attempts remaining",
			err:      errors.NewTimeoutError("operation timed out"),
			attempts: 1,
			want:     true,
		},
		{
			name:     "timeout error with no attempts remaining",
			err:      errors.NewTimeoutError("operation timed out"),
			attempts: 2,
			want:     false,
		},
		{
			name:     "not found error",
			err:      errors.NewNotFoundError("resource not found"),
			attempts: 1,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldRetry(tt.err, tt.attempts)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRetryDelay(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want time.Duration
	}{
		{
			name: "rate limit error",
			err:  errors.NewRateLimitError("rate limit exceeded"),
			want: 5 * time.Second,
		},
		{
			name: "timeout error",
			err:  errors.NewTimeoutError("operation timed out"),
			want: 2 * time.Second,
		},
		{
			name: "not found error",
			err:  errors.NewNotFoundError("resource not found"),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRetryDelay(tt.err)
			if got != tt.want {
				t.Errorf("GetRetryDelay() = %v, want %v", got, tt.want)
			}
		})
	}
} 