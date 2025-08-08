package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

// SlogLogger implements Logger interface using Go's structured logging
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new structured logger
func NewSlogLogger(level, format, output, timeFormat string) (*SlogLogger, error) {
	// Parse log level
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Determine output destination
	var writer io.Writer
	switch strings.ToLower(output) {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		writer = os.Stdout
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time attribute
			if a.Key == slog.TimeKey {
				if timeFormat != "" {
					return slog.Attr{
						Key:   a.Key,
						Value: slog.StringValue(a.Value.Time().Format(timeFormat)),
					}
				}
			}
			return a
		},
	}

	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	logger := slog.New(handler)
	
	return &SlogLogger{logger: logger}, nil
}

// Debug logs a debug message with optional fields
func (l *SlogLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message with optional fields
func (l *SlogLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message with optional fields
func (l *SlogLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message with optional fields
func (l *SlogLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, fields...)
}

// With creates a new logger with the given fields
func (l *SlogLogger) With(fields ...interface{}) Logger {
	return &SlogLogger{
		logger: l.logger.With(fields...),
	}
}

// LogOrganizationCreated logs organization creation with structured fields
func LogOrganizationCreated(logger Logger, code, name, unitType string, commandID, tenantID, parentCode string) {
	logger.Info("organization created",
		"event", "organization_created",
		"organization_code", code,
		"organization_name", name,
		"unit_type", unitType,
		"command_id", commandID,
		"tenant_id", tenantID,
		"parent_code", parentCode,
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)
}

// LogOrganizationUpdated logs organization update with structured fields
func LogOrganizationUpdated(logger Logger, code string, changes map[string]interface{}, commandID, tenantID string) {
	logger.Info("organization updated",
		"event", "organization_updated",
		"organization_code", code,
		"changes", changes,
		"command_id", commandID,
		"tenant_id", tenantID,
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)
}

// LogOrganizationDeleted logs organization deletion with structured fields
func LogOrganizationDeleted(logger Logger, code, commandID, tenantID string) {
	logger.Info("organization deleted",
		"event", "organization_deleted",
		"organization_code", code,
		"command_id", commandID,
		"tenant_id", tenantID,
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)
}