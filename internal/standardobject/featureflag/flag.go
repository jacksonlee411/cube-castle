package featureflag

import (
	"context"
	"os"
	"strings"
	"sync"
)

// Toggle is the minimal contract consumed by the standardobject adapter.
type Toggle interface {
	Enabled(ctx context.Context) bool
}

// StaticToggle returns a Toggle whose value never changes.
type StaticToggle struct {
	enabled bool
}

// NewStaticToggle builds a Toggle that always returns the provided value.
func NewStaticToggle(enabled bool) Toggle {
	return StaticToggle{enabled: enabled}
}

// Enabled implements the Toggle interface.
func (t StaticToggle) Enabled(_ context.Context) bool {
	return t.enabled
}

// EnvToggle reads a boolean flag from the provided environment variable once per process.
type EnvToggle struct {
	envVar       string
	defaultValue bool

	once   sync.Once
	parsed bool
	value  bool
}

// NewEnvToggle returns a Toggle source using STANDARD_OBJECTS_ENABLED (or other env var) semantics.
func NewEnvToggle(envVar string, defaultValue bool) *EnvToggle {
	return &EnvToggle{
		envVar:       envVar,
		defaultValue: defaultValue,
	}
}

// Enabled implements Toggle using lazy env parsing.
func (t *EnvToggle) Enabled(_ context.Context) bool {
	t.once.Do(func() {
		val := strings.TrimSpace(os.Getenv(t.envVar))
		switch strings.ToLower(val) {
		case "1", "true", "on", "yes":
			t.value = true
			t.parsed = true
		case "0", "false", "off", "no":
			t.value = false
			t.parsed = true
		default:
			t.value = t.defaultValue
		}
	})

	if !t.parsed {
		return t.defaultValue
	}
	return t.value
}
