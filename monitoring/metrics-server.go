package main

import (
	"log"
	"net/http"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "status"},
	)
	
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
	
	organizationOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "organization_operations_total",
			Help: "Total organization CRUD operations",
		},
		[]string{"operation", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(organizationOperations)
}

// MetricsMiddleware wraps HTTP handlers with metrics collection
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap the ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, http.StatusText(wrapped.statusCode)).Inc()
	})
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
func RecordOrganizationOperation(operation, status string) {
	organizationOperations.WithLabelValues(operation, status).Inc()
}

func main() {
	// 在独立端口提供 metrics
	http.Handle("/metrics", promhttp.Handler())
	
	log.Println("📊 Metrics server starting on :9999")
	log.Println("🔍 Metrics endpoint: http://localhost:9999/metrics")
	
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal("Failed to start metrics server:", err)
	}
}