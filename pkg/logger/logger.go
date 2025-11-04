package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the minimum severity that will be recorded by a Logger.
type Level int

// Fields stores structured attributes attached to logger entries.
type Fields map[string]interface{}

const (
	// LevelDebug captures verbose diagnostic output.
	LevelDebug Level = iota
	// LevelInfo captures high-level state changes and successful operations.
	LevelInfo
	// LevelWarn captures recoverable issues or unexpected states.
	LevelWarn
	// LevelError captures unrecoverable failures that require attention.
	LevelError
)

// String converts a Level to its textual representation.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a textual level (case-insensitive) into a Level value.
func ParseLevel(s string) (Level, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("logger: unsupported level %q", s)
	}
}

// Logger defines the structured logging contract used across the codebase.
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	WithFields(fields Fields) Logger
}

// Option configures a structured logger instance.
type Option func(*structuredLogger)

type structuredLogger struct {
	writer     io.Writer
	level      Level
	fields     Fields
	now        func() time.Time
	callerSkip int
	lock       *sync.Mutex
}

// NewLogger constructs a new structured logger instance.
func NewLogger(opts ...Option) Logger {
	l := &structuredLogger{
		writer:     os.Stdout,
		level:      LevelInfo,
		fields:     Fields{},
		now:        time.Now,
		callerSkip: 0,
		lock:       &sync.Mutex{},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// WithLevel configures the minimum log level.
func WithLevel(level Level) Option {
	return func(l *structuredLogger) {
		l.level = level
	}
}

// WithLevelString parses and configures the minimum log level using textual input.
func WithLevelString(level string) Option {
	return func(l *structuredLogger) {
		if parsed, err := ParseLevel(level); err == nil {
			l.level = parsed
		}
	}
}

// WithWriter configures the output writer.
func WithWriter(w io.Writer) Option {
	return func(l *structuredLogger) {
		if w != nil {
			l.writer = w
		}
	}
}

// WithTimestampFunc overrides the timestamp provider (primarily for testing).
func WithTimestampFunc(fn func() time.Time) Option {
	return func(l *structuredLogger) {
		if fn != nil {
			l.now = fn
		}
	}
}

// WithCallerSkip adjusts the number of frames skipped when capturing caller information.
func WithCallerSkip(skip int) Option {
	return func(l *structuredLogger) {
		if skip >= 0 {
			l.callerSkip = skip
		}
	}
}

func (l *structuredLogger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

func (l *structuredLogger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}

func (l *structuredLogger) Info(msg string) {
	l.log(LevelInfo, msg)
}

func (l *structuredLogger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

func (l *structuredLogger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

func (l *structuredLogger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *structuredLogger) Error(msg string) {
	l.log(LevelError, msg)
}

func (l *structuredLogger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

func (l *structuredLogger) WithFields(fields Fields) Logger {
	if len(fields) == 0 {
		return l
	}

	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}

	return &structuredLogger{
		writer:     l.writer,
		level:      l.level,
		fields:     merged,
		now:        l.now,
		callerSkip: l.callerSkip,
		lock:       l.lock,
	}
}

func (l *structuredLogger) log(level Level, message string) {
	if level < l.level || l == nil {
		return
	}

	entry := logEntry{
		Timestamp: l.now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   message,
		Caller:    l.caller(),
	}

	if len(l.fields) > 0 {
		entry.Fields = make(map[string]interface{}, len(l.fields))
		for k, v := range l.fields {
			entry.Fields[k] = v
		}
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		l.writeFallback(level, message, err)
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	_, _ = l.writer.Write(payload)
	_, _ = l.writer.Write([]byte("\n"))
}

func (l *structuredLogger) caller() string {
	const defaultSkip = 4
	pcs := make([]uintptr, 4)
	n := runtime.Callers(defaultSkip+l.callerSkip, pcs)
	if n == 0 {
		return ""
	}
	frame, _ := runtime.CallersFrames(pcs[:n]).Next()
	file := frame.File
	if idx := strings.LastIndex(file, "/"); idx != -1 && idx+1 < len(file) {
		file = file[idx+1:]
	}
	return fmt.Sprintf("%s:%d", file, frame.Line)
}

func (l *structuredLogger) writeFallback(level Level, message string, marshalErr error) {
	fallback := fmt.Sprintf(
		`{"timestamp":"%s","level":"%s","message":"%s","fields":{"marshal_error":"%s"}}`,
		l.now().UTC().Format(time.RFC3339Nano),
		level.String(),
		safeJSONString(message),
		safeJSONString(marshalErr.Error()),
	)

	l.lock.Lock()
	defer l.lock.Unlock()
	_, _ = l.writer.Write([]byte(fallback))
	_, _ = l.writer.Write([]byte("\n"))
}

func safeJSONString(s string) string {
	b, err := json.Marshal(s)
	if err != nil || len(b) < 2 {
		return s
	}
	return string(b[1 : len(b)-1])
}

type logEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
}

type noopLogger struct{}

// NewNoopLogger returns a Logger implementation that discards all output.
func NewNoopLogger() Logger {
	return noopLogger{}
}

func (noopLogger) Debug(string)                  {}
func (noopLogger) Debugf(string, ...interface{}) {}
func (noopLogger) Info(string)                   {}
func (noopLogger) Infof(string, ...interface{})  {}
func (noopLogger) Warn(string)                   {}
func (noopLogger) Warnf(string, ...interface{})  {}
func (noopLogger) Error(string)                  {}
func (noopLogger) Errorf(string, ...interface{}) {}
func (noopLogger) WithFields(Fields) Logger      { return noopLogger{} }
