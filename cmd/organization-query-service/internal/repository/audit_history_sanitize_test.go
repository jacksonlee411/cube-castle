package repository

import (
	"io"
	"log"
	"reflect"
	"testing"
)

func TestSanitizeModifiedFields(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       []string
		wantIssues int
		wantErr    bool
	}{
		{
			name:       "empty string returns empty slice",
			input:      "",
			want:       []string{},
			wantIssues: 0,
		},
		{
			name:       "null replaced with empty slice and issue",
			input:      "null",
			want:       []string{},
			wantIssues: 1,
		},
		{
			name:       "valid array preserved",
			input:      "[\"name\",\"status\"]",
			want:       []string{"name", "status"},
			wantIssues: 0,
		},
		{
			name:       "non string coerced",
			input:      "[123,\"name\"]",
			want:       []string{"123", "name"},
			wantIssues: 1,
		},
		{
			name:    "invalid json",
			input:   "{",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, issues, err := sanitizeModifiedFields(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil slice")
			}
			if tc.want != nil && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("unexpected sanitized result: got %v, want %v", got, tc.want)
			}
			if len(issues) != tc.wantIssues {
				t.Fatalf("unexpected issues count: got %d, want %d", len(issues), tc.wantIssues)
			}
		})
	}
}

func TestSanitizeChanges(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantLen        int
		wantDataType   string
		wantIssues     int
		firstFieldName string
		wantErr        bool
	}{
		{
			name:           "empty array",
			input:          "[]",
			wantLen:        0,
			wantIssues:     0,
			firstFieldName: "",
		},
		{
			name:           "missing dataType produces unknown",
			input:          `[{"field":"name","oldValue":"A","newValue":"B"}]`,
			wantLen:        1,
			wantIssues:     1,
			wantDataType:   "string",
			firstFieldName: "name",
		},
		{
			name:    "invalid json",
			input:   "{",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, issues, err := sanitizeChanges(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil slice")
			}
			if len(got) != tc.wantLen {
				t.Fatalf("unexpected length: got %d, want %d", len(got), tc.wantLen)
			}
			if tc.wantLen > 0 {
				if got[0].FieldField != tc.firstFieldName {
					t.Fatalf("unexpected field name: got %s, want %s", got[0].FieldField, tc.firstFieldName)
				}
				if got[0].DataTypeField != tc.wantDataType {
					t.Fatalf("unexpected dataType: got %s, want %s", got[0].DataTypeField, tc.wantDataType)
				}
			}
			if len(issues) != tc.wantIssues {
				t.Fatalf("unexpected issues count: got %d, want %d", len(issues), tc.wantIssues)
			}
		})
	}
}

func TestNormalizeChangeValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{"nil", nil, nil},
		{"string", "abc", "abc"},
		{"bool", true, "true"},
		{"float", 12.5, "12.5"},
		{"map", map[string]interface{}{"k": "v"}, "{\"k\":\"v\"}"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeChangeValue(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("unexpected normalize result: got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRegisterValidationFailure(t *testing.T) {
	repo := &PostgreSQLRepository{
		logger: log.New(io.Discard, "", 0),
		auditConfig: AuditHistoryConfig{
			CircuitBreakerThreshold: 2,
		},
	}

	if repo.registerValidationFailure() {
		t.Fatalf("expected circuit to remain closed on first failure")
	}
	if !repo.registerValidationFailure() {
		t.Fatalf("expected circuit to open on second failure when threshold=2")
	}

	repo.registerValidationSuccess()

	if repo.registerValidationFailure() {
		t.Fatalf("expected circuit to remain closed after reset")
	}
}
