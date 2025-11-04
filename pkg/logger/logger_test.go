package logger

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestLoggerWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	fixed := time.Date(2024, 12, 1, 10, 30, 45, int(123*time.Millisecond), time.UTC)

	base := NewLogger(
		WithWriter(buf),
		WithLevel(LevelDebug),
		WithTimestampFunc(func() time.Time { return fixed }),
	)

	logger := base.WithFields(Fields{"component": "repository"})
	logger.Info("organization created")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	if got := entry["level"]; got != "INFO" {
		t.Fatalf("expected level INFO, got %v", got)
	}

	if got := entry["message"]; got != "organization created" {
		t.Fatalf("expected message 'organization created', got %v", got)
	}

	fields, ok := entry["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected fields object, got %T", entry["fields"])
	}
	if fields["component"] != "repository" {
		t.Fatalf("expected component field, got %v", fields["component"])
	}

	if entry["timestamp"] != fixed.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("unexpected timestamp: %v", entry["timestamp"])
	}

	if caller, ok := entry["caller"].(string); !ok || caller == "" {
		t.Fatalf("expected caller to be populated, got %v", entry["caller"])
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(
		WithWriter(buf),
		WithLevel(LevelWarn),
		WithTimestampFunc(func() time.Time { return time.Unix(0, 0) }),
	)

	logger.Info("info should be suppressed")
	if buf.Len() != 0 {
		t.Fatalf("expected no output for INFO when level WARN, got %s", buf.String())
	}

	logger.Warn("warn should pass")
	if buf.Len() == 0 {
		t.Fatal("expected WARN log to be written")
	}
}

func TestStdLoggerBridge(t *testing.T) {
	buf := &bytes.Buffer{}
	base := NewLogger(
		WithWriter(buf),
		WithLevel(LevelDebug),
		WithTimestampFunc(func() time.Time { return time.Unix(0, 0) }),
	)

	std := NewStdLogger(base.WithFields(Fields{"bridge": "std"}), LevelInfo)
	std.Printf("legacy message: %s", "ok")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal bridged log entry: %v", err)
	}

	if entry["level"] != "INFO" {
		t.Fatalf("expected INFO level, got %v", entry["level"])
	}
	if entry["message"] != "legacy message: ok" {
		t.Fatalf("unexpected message: %v", entry["message"])
	}
	fields := entry["fields"].(map[string]interface{})
	if fields["bridge"] != "std" {
		t.Fatalf("expected bridge field, got %v", fields["bridge"])
	}
}

func TestParseLevel(t *testing.T) {
	lvl, err := ParseLevel("warn")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if lvl != LevelWarn {
		t.Fatalf("expected LevelWarn, got %v", lvl)
	}

	if _, err := ParseLevel("invalid"); err == nil {
		t.Fatal("expected error for invalid level")
	}
}
