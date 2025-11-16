package scheduler

import (
	"testing"

	pkglogger "cube-castle/pkg/logger"
)

func TestTemporalMonitor_DefaultRules(t *testing.T) {
	m := NewTemporalMonitor(nil, pkglogger.NewNoopLogger())
	rules := m.GetDefaultAlertRules()
	if len(rules) == 0 {
		t.Fatalf("expected default alert rules")
	}
	found := false
	for _, r := range rules {
		if r.Name == "HEALTH_SCORE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected HEALTH_SCORE rule")
	}
}

