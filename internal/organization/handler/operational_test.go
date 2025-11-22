package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/organization/middleware"
	scheduler "cube-castle/internal/organization/scheduler"
	pkglogger "cube-castle/pkg/logger"
)

type fakeMonitor struct {
	metrics *scheduler.MonitoringMetrics
	alerts  []string
	err     error
}

func (m *fakeMonitor) CollectMetrics(_ context.Context) (*scheduler.MonitoringMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.metrics != nil {
		return m.metrics, nil
	}
	return &scheduler.MonitoringMetrics{}, nil
}

func (m *fakeMonitor) CheckAlerts(_ context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.alerts, nil
}

type fakeScheduler struct {
	running bool
	tasks   []scheduler.TaskSnapshot
	runErr  error
}

func (s *fakeScheduler) ListTasks() []scheduler.TaskSnapshot { return s.tasks }
func (s *fakeScheduler) IsRunning() bool                     { return s.running }
func (s *fakeScheduler) RunTask(_ interface{}, _ string) error {
	return s.runErr
}

func newOperationalHandlerForTest(mon monitoringCollector, _ *fakeScheduler, rl *middleware.RateLimitMiddleware) *OperationalHandler {
	return &OperationalHandler{
		monitor:   mon,
		scheduler: (*scheduler.OperationalScheduler)(nil),
		rateLimit: rl,
		logger:    pkglogger.NewNoopLogger(),
	}
}

func TestOperationalHandler_RateLimitStats(t *testing.T) {
	rl := middleware.NewRateLimitMiddleware(nil, pkglogger.NewNoopLogger())
	rl.UpdateConfig(&middleware.RateLimitConfig{RequestsPerMinute: 10})
	rl.BlockIP("1.1.1.1", time.Minute)
	rl.UpdateConfig(rl.Config())

	h := newOperationalHandlerForTest(nil, nil, rl)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operational/rate-limit/stats", nil)
	h.GetRateLimitStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !contains(rr.Body.String(), "totalRequests") {
		t.Fatalf("expected stats payload, got %s", rr.Body.String())
	}
}

func TestOperationalHandler_GetHealth_Failure(t *testing.T) {
	monitor := &fakeMonitor{err: errors.New("collect failed")}
	h := newOperationalHandlerForTest(monitor, nil, middleware.NewRateLimitMiddleware(nil, pkglogger.NewNoopLogger()))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operational/health", nil)

	h.GetHealth(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestOperationalHandler_GetTasks_DisabledScheduler(t *testing.T) {
	h := newOperationalHandlerForTest(nil, nil, middleware.NewRateLimitMiddleware(nil, pkglogger.NewNoopLogger()))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/operational/tasks", nil)

	h.GetTasks(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when scheduler missing, got %d", rr.Code)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
