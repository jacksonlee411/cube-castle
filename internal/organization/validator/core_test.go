package validator

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"

	pkglogger "cube-castle/pkg/logger"
)

func TestValidatorCoreSmoke(t *testing.T) {
	logger := pkglogger.NewLogger(
		pkglogger.WithWriter(io.Discard),
		pkglogger.WithLevel(pkglogger.LevelError),
	)

	chain := NewValidationChain(logger)

	executionOrder := make([]string, 0)

	if err := chain.Register(&Rule{
		ID:       "ORG-DEPTH",
		Priority: 10,
		Severity: SeverityHigh,
		Handler: func(_ context.Context, _ interface{}) (*RuleOutcome, error) {
			executionOrder = append(executionOrder, "ORG-DEPTH")
			return &RuleOutcome{
				Warnings: []ValidationWarning{
					{
						Code:    "DEPTH_WARNING",
						Message: "depth close to threshold",
					},
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("register rule ORG-DEPTH failed: %v", err)
	}

	if err := chain.Register(&Rule{
		ID:           "POS-HEADCOUNT",
		Priority:     20,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler: func(_ context.Context, _ interface{}) (*RuleOutcome, error) {
			executionOrder = append(executionOrder, "POS-HEADCOUNT")
			return &RuleOutcome{
				Errors: []ValidationError{
					{
						Code:    "POS_HEADCOUNT_EXCEEDED",
						Message: "headcount exceeds limit",
						Field:   "headcount",
					},
				},
				Context: map[string]interface{}{
					"current":  3,
					"capacity": 2,
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("register rule POS-HEADCOUNT failed: %v", err)
	}

	if err := chain.Register(&Rule{
		ID:       "ASSIGN-FTE",
		Priority: 30,
		Severity: SeverityMedium,
		Handler: func(_ context.Context, _ interface{}) (*RuleOutcome, error) {
			executionOrder = append(executionOrder, "ASSIGN-FTE")
			return &RuleOutcome{
				Errors: []ValidationError{
					{
						Code:    "ASSIGN_FTE_LIMIT",
						Message: "FTE limit exceeded",
					},
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("register rule ASSIGN-FTE failed: %v", err)
	}

	result := chain.Execute(context.Background(), map[string]interface{}{
		"headcount": 3,
		"capacity":  2,
	})

	if result.Valid {
		t.Fatalf("expected result to be invalid")
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	errItem := result.Errors[0]
	if errItem.Code != "POS_HEADCOUNT_EXCEEDED" {
		t.Fatalf("unexpected error code %q", errItem.Code)
	}
	if errItem.Severity != string(SeverityHigh) {
		t.Fatalf("expected severity %q, got %q", SeverityHigh, errItem.Severity)
	}
	if errItem.Context == nil || errItem.Context["ruleId"] != "POS-HEADCOUNT" {
		t.Fatalf("expected ruleId POS-HEADCOUNT in error context, got %#v", errItem.Context)
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}

	if got := executionOrder; !reflect.DeepEqual(got, []string{"ORG-DEPTH", "POS-HEADCOUNT"}) {
		t.Fatalf("expected execution order [ORG-DEPTH POS-HEADCOUNT], got %#v", got)
	}

	executed, ok := result.Context["executedRules"].([]string)
	if !ok {
		t.Fatalf("expected executedRules in context, got %#v", result.Context["executedRules"])
	}
	if !reflect.DeepEqual(executed, []string{"ORG-DEPTH", "POS-HEADCOUNT"}) {
		t.Fatalf("unexpected executedRules %#v", executed)
	}
}

func TestSeverityToHTTPStatus(t *testing.T) {
	tests := map[string]int{
		string(SeverityCritical): http.StatusBadRequest,
		string(SeverityHigh):     http.StatusBadRequest,
		string(SeverityMedium):   http.StatusUnprocessableEntity,
		string(SeverityLow):      http.StatusOK,
		"unknown":                http.StatusBadRequest,
	}

	for input, want := range tests {
		got := SeverityToHTTPStatus(input)
		if got != want {
			t.Fatalf("severity %q expected status %d, got %d", input, want, got)
		}
	}
}

func TestValidationChainWithBaseContext(t *testing.T) {
	chain := NewValidationChain(nil, WithBaseContext(map[string]interface{}{"tenant": "T-1"}))

	result := chain.Execute(context.Background(), struct{}{})
	if tenant, ok := result.Context["tenant"]; !ok || tenant != "T-1" {
		t.Fatalf("expected base context to be merged, got %#v", result.Context)
	}
	if _, ok := result.Context["executedRules"]; !ok {
		t.Fatalf("expected executedRules key to exist")
	}
}

func TestValidationChainContextCancelled(t *testing.T) {
	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	chain := NewValidationChain(logger)
	called := false
	_ = chain.Register(&Rule{
		ID:       "TEST-CANCEL",
		Severity: SeverityHigh,
		Handler: func(_ context.Context, _ interface{}) (*RuleOutcome, error) {
			called = true
			return nil, nil
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := chain.Execute(ctx, struct{}{})
	if called {
		t.Fatalf("expected handler not to be called when context cancelled")
	}
	if cancelled, ok := result.Context["cancelled"].(bool); !ok || !cancelled {
		t.Fatalf("expected cancelled flag in context, got %#v", result.Context)
	}
	executed := result.Context["executedRules"].([]string)
	if len(executed) != 0 {
		t.Fatalf("expected no executed rules, got %#v", executed)
	}
}

func TestValidationChainTelemetryOnly(t *testing.T) {
	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	chain := NewValidationChain(logger)
	var order []string
	_ = chain.Register(&Rule{
		ID:            "TELEMETRY",
		Severity:      SeverityHigh,
		Priority:      5,
		ShortCircuit:  true,
		TelemetryOnly: true,
		Handler: func(_ context.Context, _ interface{}) (*RuleOutcome, error) {
			order = append(order, "TELEMETRY")
			return &RuleOutcome{Errors: []ValidationError{{Code: "TELEMETRY_FAILURE"}}}, nil
		},
	})
	_ = chain.Register(&Rule{
		ID:       "NEXT",
		Priority: 10,
		Handler: func(_ context.Context, _ interface{}) (*RuleOutcome, error) {
			order = append(order, "NEXT")
			return nil, nil
		},
	})

	result := chain.Execute(context.Background(), struct{}{})
	if !reflect.DeepEqual(order, []string{"TELEMETRY", "NEXT"}) {
		t.Fatalf("expected both rules to execute, got %#v", order)
	}
	if result.Valid {
		t.Fatalf("expected result to be invalid due to telemetry error")
	}
}

func TestValidationChainRegisterNilRule(t *testing.T) {
	chain := NewValidationChain(nil)
	if err := chain.Register(nil); err == nil {
		t.Fatalf("expected error when registering nil rule")
	}
}
