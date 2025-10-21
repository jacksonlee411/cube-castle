package graphql

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaLoaderLoadAndCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.graphql")
	content := "type Query { ping: String! }\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	loader := NewSchemaLoader(path)
	first, err := loader.LoadSchema()
	if err != nil {
		t.Fatalf("expected schema to load: %v", err)
	}
	second, err := loader.LoadSchema()
	if err != nil {
		t.Fatalf("expected cached load to succeed: %v", err)
	}
	if first != second || first != content {
		t.Fatalf("expected cached schema to match original")
	}
}

func TestSchemaLoaderMissingFile(t *testing.T) {
	loader := NewSchemaLoader(filepath.Join(t.TempDir(), "missing.graphql"))
	if _, err := loader.LoadSchema(); err == nil {
		t.Fatalf("expected error when schema file missing")
	}
}

func TestMustLoadSchema(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected MustLoadSchema to panic when file missing")
		}
	}()
	MustLoadSchema(filepath.Join(t.TempDir(), "absent.graphql"))
}

func TestValidateSchemaConsistency(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.graphql")
	if err := os.WriteFile(path, []byte("type Query { ping: String! }\n"), 0o600); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	if err := ValidateSchemaConsistency(path, "type Query { ping: String! }\n"); err != nil {
		t.Fatalf("expected schema validation to succeed: %v", err)
	}

	err := ValidateSchemaConsistency(path, "type Query { pong: String! }\n")
	if err == nil || !strings.Contains(err.Error(), "schema inconsistency") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestGetDefaultSchemaPath(t *testing.T) {
	if p := GetDefaultSchemaPath(); !strings.HasSuffix(p, filepath.Join("docs", "api", "schema.graphql")) {
		t.Fatalf("unexpected default path: %s", p)
	}
}
