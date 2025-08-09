package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "status", "service"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"method", "endpoint", "service"},
	)

	organizationOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "organization_operations_total",
			Help: "Total organization CRUD operations",
		},
		[]string{"operation", "status", "service"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(organizationOperations)
}

// MetricsMiddleware wraps HTTP handlers with metrics collection
func MetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start).Seconds()
			httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, serviceName).Observe(duration)
			httpRequestsTotal.WithLabelValues(r.Method, http.StatusText(wrapped.statusCode), serviceName).Inc()
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecordOrganizationOperation records business metrics
func RecordOrganizationOperation(operation, status, service string) {
	organizationOperations.WithLabelValues(operation, status, service).Inc()
}