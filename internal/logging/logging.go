package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level represents the logging level
type Level string

const (
	// LevelDebug represents debug level logging
	LevelDebug Level = "debug"
	// LevelInfo represents info level logging
	LevelInfo Level = "info"
	// LevelWarn represents warning level logging
	LevelWarn Level = "warn"
	// LevelError represents error level logging
	LevelError Level = "error"
)

// Logger is the interface for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field represents a logging field
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new Field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// zapLogger implements the Logger interface using zap
type zapLogger struct {
	logger *zap.Logger
	fields []Field
}

// NewLogger creates a new logger
func NewLogger(level Level) (Logger, error) {
	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.Level(getZapLevel(level)),
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &zapLogger{logger: logger}, nil
}

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, l.getZapFields(fields)...)
}

// Info logs an info message
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, l.getZapFields(fields)...)
}

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, l.getZapFields(fields)...)
}

// Error logs an error message
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, l.getZapFields(fields)...)
}

// With returns a logger with the given fields
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(l.getZapFields(fields)...),
		fields: append(l.fields, fields...),
	}
}

// WithContext returns a logger with context fields
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	fields := []Field{
		F("request_id", ctx.Value("request_id")),
		F("user_id", ctx.Value("user_id")),
		F("conversation_id", ctx.Value("conversation_id")),
	}
	return l.With(fields...)
}

// getZapFields converts Field to zap.Field
func (l *zapLogger) getZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)+len(l.fields))

	// Add logger fields
	for _, f := range l.fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}

	// Add message fields
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}

	return zapFields
}

// getZapLevel converts Level to zapcore.Level
func getZapLevel(level Level) zapcore.Level {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// LogEvent represents a structured log event
type LogEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Stacktrace  string                 `json:"stacktrace,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	ConversationID string             `json:"conversation_id,omitempty"`
}

// FormatEvent formats a log event as JSON
func FormatEvent(event LogEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal log event: %w", err)
	}
	return string(data), nil
} 