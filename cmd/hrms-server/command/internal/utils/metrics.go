package utils

import (
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	OperationCreate     = "create"
	OperationUpdate     = "update"
	OperationDelete     = "delete"
	OperationSuspend    = "suspend"
	OperationReactivate = "reactivate"

	StatusSuccess = "success"
	StatusError   = "error"
)

var (
	registerOnce sync.Once

	temporalOperationsTotal *prometheus.CounterVec
	auditWritesTotal        *prometheus.CounterVec
	httpRequestsTotal       *prometheus.CounterVec
)

func ensureRegistered() {
	registerOnce.Do(func() {
		temporalOperationsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "temporal_operations_total",
				Help: "Total number of temporal operations grouped by operation type and outcome.",
			},
			[]string{"operation", "status"},
		)

		auditWritesTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "audit_writes_total",
				Help: "Total number of audit log writes grouped by outcome.",
			},
			[]string{"status"},
		)

		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests handled by the command service grouped by method, route, and status code.",
			},
			[]string{"method", "route", "status"},
		)

		prometheus.MustRegister(temporalOperationsTotal, auditWritesTotal, httpRequestsTotal)
	})
}

func RecordTemporalOperation(operation string, err error) {
	ensureRegistered()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	temporalOperationsTotal.WithLabelValues(operation, status).Inc()
}

func RecordAuditWrite(err error) {
	ensureRegistered()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	auditWritesTotal.WithLabelValues(status).Inc()
}

func RecordHTTPRequest(method, route string, statusCode int) {
	ensureRegistered()

	code := strconv.Itoa(statusCode)
	httpRequestsTotal.WithLabelValues(method, route, code).Inc()
}
