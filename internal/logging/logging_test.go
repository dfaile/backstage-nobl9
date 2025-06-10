package logging

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"backstage-nobl9/internal/logging"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{
			name:  "debug level",
			level: LevelDebug,
		},
		{
			name:  "info level",
			level: LevelInfo,
		},
		{
			name:  "warn level",
			level: LevelWarn,
		},
		{
			name:  "error level",
			level: LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level)
			assert.NoError(t, err)
			assert.NotNil(t, logger)
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	logger, err := NewLogger(LevelInfo)
	assert.NoError(t, err)

	// Test with single field
	loggerWithField := logger.With(F("key", "value"))
	assert.NotNil(t, loggerWithField)

	// Test with multiple fields
	loggerWithFields := logger.With(
		F("key1", "value1"),
		F("key2", "value2"),
	)
	assert.NotNil(t, loggerWithFields)
}

func TestLoggerWithContext(t *testing.T) {
	logger, err := NewLogger(LevelInfo)
	assert.NoError(t, err)

	// Create context with values
	ctx := context.WithValue(context.Background(), "request_id", "123")
	ctx = context.WithValue(ctx, "user_id", "user123")
	ctx = context.WithValue(ctx, "conversation_id", "conv123")

	// Test with context
	loggerWithContext := logger.WithContext(ctx)
	assert.NotNil(t, loggerWithContext)
}

func TestFormatEvent(t *testing.T) {
	now := time.Now()
	event := LogEvent{
		Timestamp:      now,
		Level:          "info",
		Message:        "test message",
		Fields:         map[string]interface{}{"key": "value"},
		Error:          "test error",
		Stacktrace:     "test stacktrace",
		RequestID:      "123",
		UserID:         "user123",
		ConversationID: "conv123",
	}

	// Format event
	formatted, err := FormatEvent(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, formatted)

	// Parse formatted event
	var parsed LogEvent
	err = json.Unmarshal([]byte(formatted), &parsed)
	assert.NoError(t, err)

	// Verify fields
	assert.Equal(t, event.Level, parsed.Level)
	assert.Equal(t, event.Message, parsed.Message)
	assert.Equal(t, event.Fields, parsed.Fields)
	assert.Equal(t, event.Error, parsed.Error)
	assert.Equal(t, event.Stacktrace, parsed.Stacktrace)
	assert.Equal(t, event.RequestID, parsed.RequestID)
	assert.Equal(t, event.UserID, parsed.UserID)
	assert.Equal(t, event.ConversationID, parsed.ConversationID)
}

func TestField(t *testing.T) {
	field := F("key", "value")
	assert.Equal(t, "key", field.Key)
	assert.Equal(t, "value", field.Value)
}

func TestGetZapLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected zapcore.Level
	}{
		{
			name:     "debug level",
			level:    LevelDebug,
			expected: zapcore.DebugLevel,
		},
		{
			name:     "info level",
			level:    LevelInfo,
			expected: zapcore.InfoLevel,
		},
		{
			name:     "warn level",
			level:    LevelWarn,
			expected: zapcore.WarnLevel,
		},
		{
			name:     "error level",
			level:    LevelError,
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "unknown level",
			level:    "unknown",
			expected: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := getZapLevel(tt.level)
			assert.Equal(t, tt.expected, level)
		})
	}
} 