package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMonitor(t *testing.T) {
	tests := []struct {
		name     string
		config   *MonitorConfig
		expected *MonitorConfig
	}{
		{
			name:   "with nil config",
			config: nil,
			expected: &MonitorConfig{
				ServiceName: "cube-castle",
				Version:     "1.0.0",
				Environment: "development",
			},
		},
		{
			name: "with custom config",
			config: &MonitorConfig{
				ServiceName: "test-service",
				Version:     "2.0.0",
				Environment: "production",
			},
			expected: &MonitorConfig{
				ServiceName: "test-service",
				Version:     "2.0.0",
				Environment: "production",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewMonitor(tt.config)
			
			if monitor == nil {
				t.Fatal("Expected monitor to be created, got nil")
			}
			
			if monitor.config.ServiceName != tt.expected.ServiceName {
				t.Errorf("Expected ServiceName %s, got %s", tt.expected.ServiceName, monitor.config.ServiceName)
			}
			
			if monitor.config.Version != tt.expected.Version {
				t.Errorf("Expected Version %s, got %s", tt.expected.Version, monitor.config.Version)
			}
			
			if monitor.config.Environment != tt.expected.Environment {
				t.Errorf("Expected Environment %s, got %s", tt.expected.Environment, monitor.config.Environment)
			}
			
			if monitor.httpMetrics == nil {
				t.Error("Expected httpMetrics to be initialized")
			}
			
			if monitor.systemMetrics == nil {
				t.Error("Expected systemMetrics to be initialized")
			}
		})
	}
}

func TestMonitor_GetHealthStatus(t *testing.T) {
	monitor := NewMonitor(nil)
	ctx := context.Background()
	
	status := monitor.GetHealthStatus(ctx)
	
	if status == nil {
		t.Fatal("Expected health status to be returned, got nil")
	}
	
	if status.Service != "cube-castle" {
		t.Errorf("Expected service name 'cube-castle', got %s", status.Service)
	}
	
	if status.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", status.Status)
	}
	
	if status.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", status.Version)
	}
	
	if status.Environment != "development" {
		t.Errorf("Expected environment 'development', got %s", status.Environment)
	}
	
	if status.Checks == nil {
		t.Error("Expected checks to be initialized")
	}
	
	if len(status.Checks) == 0 {
		t.Error("Expected at least one health check")
	}
}

func TestMonitor_GetDetailedHealthStatus(t *testing.T) {
	monitor := NewMonitor(nil)
	ctx := context.Background()
	
	status := monitor.GetDetailedHealthStatus(ctx)
	
	if status == nil {
		t.Fatal("Expected detailed health status to be returned, got nil")
	}
	
	expectedChecks := []string{"api", "memory", "disk"}
	for _, checkName := range expectedChecks {
		if check, exists := status.Checks[checkName]; !exists {
			t.Errorf("Expected check '%s' to exist", checkName)
		} else {
			if check.Status == "" {
				t.Errorf("Expected check '%s' to have a status", checkName)
			}
		}
	}
}

func TestMonitor_RecordHTTPRequest(t *testing.T) {
	monitor := NewMonitor(nil)
	
	// Test recording a single request
	monitor.RecordHTTPRequest("GET", "/test", 200, time.Millisecond*100)
	
	metrics := monitor.GetHTTPMetrics()
	
	if metrics.RequestCount != 1 {
		t.Errorf("Expected request count 1, got %d", metrics.RequestCount)
	}
	
	if metrics.AverageLatency != time.Millisecond*100 {
		t.Errorf("Expected average latency %v, got %v", time.Millisecond*100, metrics.AverageLatency)
	}
	
	if metrics.StatusCodes["200"] != 1 {
		t.Errorf("Expected status code 200 count to be 1, got %d", metrics.StatusCodes["200"])
	}
	
	if metrics.ErrorRate != 0 {
		t.Errorf("Expected error rate 0, got %f", metrics.ErrorRate)
	}
	
	// Test endpoint metrics
	endpointKey := "GET /test"
	if endpoint, exists := metrics.EndpointMetrics[endpointKey]; !exists {
		t.Errorf("Expected endpoint metrics for '%s' to exist", endpointKey)
	} else {
		if endpoint.RequestCount != 1 {
			t.Errorf("Expected endpoint request count 1, got %d", endpoint.RequestCount)
		}
		if endpoint.ErrorCount != 0 {
			t.Errorf("Expected endpoint error count 0, got %d", endpoint.ErrorCount)
		}
	}
}

func TestMonitor_RecordHTTPRequest_ErrorTracking(t *testing.T) {
	monitor := NewMonitor(nil)
	
	// Record some successful and error requests
	monitor.RecordHTTPRequest("GET", "/test", 200, time.Millisecond*50)
	monitor.RecordHTTPRequest("GET", "/test", 404, time.Millisecond*30)
	monitor.RecordHTTPRequest("GET", "/test", 500, time.Millisecond*200)
	monitor.RecordHTTPRequest("POST", "/api", 201, time.Millisecond*75)
	
	metrics := monitor.GetHTTPMetrics()
	
	if metrics.RequestCount != 4 {
		t.Errorf("Expected request count 4, got %d", metrics.RequestCount)
	}
	
	// Error rate should be 50% (2 errors out of 4 requests)
	expectedErrorRate := 50.0
	if metrics.ErrorRate != expectedErrorRate {
		t.Errorf("Expected error rate %f, got %f", expectedErrorRate, metrics.ErrorRate)
	}
	
	// Check status code counts
	if metrics.StatusCodes["200"] != 1 {
		t.Errorf("Expected status code 200 count to be 1, got %d", metrics.StatusCodes["200"])
	}
	if metrics.StatusCodes["404"] != 1 {
		t.Errorf("Expected status code 404 count to be 1, got %d", metrics.StatusCodes["404"])
	}
	if metrics.StatusCodes["500"] != 1 {
		t.Errorf("Expected status code 500 count to be 1, got %d", metrics.StatusCodes["500"])
	}
	if metrics.StatusCodes["201"] != 1 {
		t.Errorf("Expected status code 201 count to be 1, got %d", metrics.StatusCodes["201"])
	}
	
	// Check endpoint-specific error tracking
	getEndpoint := metrics.EndpointMetrics["GET /test"]
	if getEndpoint.RequestCount != 3 {
		t.Errorf("Expected GET /test request count 3, got %d", getEndpoint.RequestCount)
	}
	if getEndpoint.ErrorCount != 2 {
		t.Errorf("Expected GET /test error count 2, got %d", getEndpoint.ErrorCount)
	}
	
	postEndpoint := metrics.EndpointMetrics["POST /api"]
	if postEndpoint.RequestCount != 1 {
		t.Errorf("Expected POST /api request count 1, got %d", postEndpoint.RequestCount)
	}
	if postEndpoint.ErrorCount != 0 {
		t.Errorf("Expected POST /api error count 0, got %d", postEndpoint.ErrorCount)
	}
}

func TestMonitor_CustomMetrics(t *testing.T) {
	monitor := NewMonitor(nil)
	
	// Test UpdateCustomMetric
	monitor.UpdateCustomMetric("test_metric", 42.5)
	
	metrics := monitor.GetSystemMetrics()
	if metrics.CustomMetrics["test_metric"] != 42.5 {
		t.Errorf("Expected custom metric value 42.5, got %f", metrics.CustomMetrics["test_metric"])
	}
	
	// Test IncrementCustomMetric
	monitor.IncrementCustomMetric("counter", 1.0)
	monitor.IncrementCustomMetric("counter", 2.5)
	
	metrics = monitor.GetSystemMetrics()
	if metrics.CustomMetrics["counter"] != 3.5 {
		t.Errorf("Expected counter value 3.5, got %f", metrics.CustomMetrics["counter"])
	}
	
	// Test incrementing non-existent metric
	monitor.IncrementCustomMetric("new_counter", 10.0)
	metrics = monitor.GetSystemMetrics()
	if metrics.CustomMetrics["new_counter"] != 10.0 {
		t.Errorf("Expected new_counter value 10.0, got %f", metrics.CustomMetrics["new_counter"])
	}
}

func TestMonitor_ServeHTTP(t *testing.T) {
	monitor := NewMonitor(nil)
	
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedContent string
	}{
		{
			name:           "health endpoint",
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedContent: "cube-castle",
		},
		{
			name:           "detailed health endpoint",
			path:           "/health/detailed",
			expectedStatus: http.StatusOK,
			expectedContent: "cube-castle",
		},
		{
			name:           "metrics endpoint",
			path:           "/metrics",
			expectedStatus: http.StatusOK,
			expectedContent: "metrics",
		},
		{
			name:           "system metrics endpoint",
			path:           "/metrics/system",
			expectedStatus: http.StatusOK,
			expectedContent: "cpu",
		},
		{
			name:           "http metrics endpoint",
			path:           "/metrics/http",
			expectedStatus: http.StatusOK,
			expectedContent: "http",
		},
		{
			name:           "not found endpoint",
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedContent: "404",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			
			monitor.ServeHTTP(w, req)
			
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
			
			body := w.Body.String()
			if tt.expectedContent != "" && tt.expectedContent != "404" {
				if body == "" {
					t.Error("Expected response body to contain content")
				}
				// For non-404 responses, we expect JSON content
				if w.Header().Get("Content-Type") != "application/json" && tt.expectedStatus == http.StatusOK {
					t.Error("Expected Content-Type to be application/json")
				}
			}
		})
	}
}

func TestMonitor_ConcurrentAccess(t *testing.T) {
	monitor := NewMonitor(nil)
	
	// Test concurrent HTTP request recording
	done := make(chan bool, 100)
	
	for i := 0; i < 100; i++ {
		go func(i int) {
			monitor.RecordHTTPRequest("GET", "/test", 200, time.Millisecond*time.Duration(i%10))
			monitor.UpdateCustomMetric("concurrent_test", float64(i))
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}
	
	metrics := monitor.GetHTTPMetrics()
	if metrics.RequestCount != 100 {
		t.Errorf("Expected request count 100, got %d", metrics.RequestCount)
	}
	
	systemMetrics := monitor.GetSystemMetrics()
	if _, exists := systemMetrics.CustomMetrics["concurrent_test"]; !exists {
		t.Error("Expected concurrent_test custom metric to exist")
	}
}

func TestMonitor_AverageLatencyCalculation(t *testing.T) {
	monitor := NewMonitor(nil)
	
	// Record requests with known latencies
	latencies := []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
	}
	
	for _, latency := range latencies {
		monitor.RecordHTTPRequest("GET", "/test", 200, latency)
	}
	
	metrics := monitor.GetHTTPMetrics()
	
	// Expected average: (100 + 200 + 300) / 3 = 200ms
	expectedAverage := time.Millisecond * 200
	if metrics.AverageLatency != expectedAverage {
		t.Errorf("Expected average latency %v, got %v", expectedAverage, metrics.AverageLatency)
	}
}

func BenchmarkMonitor_RecordHTTPRequest(b *testing.B) {
	monitor := NewMonitor(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordHTTPRequest("GET", "/test", 200, time.Millisecond*10)
	}
}

func BenchmarkMonitor_GetSystemMetrics(b *testing.B) {
	monitor := NewMonitor(nil)
	
	// Add some data first
	for i := 0; i < 100; i++ {
		monitor.RecordHTTPRequest("GET", "/test", 200, time.Millisecond*10)
		monitor.UpdateCustomMetric("test_metric", float64(i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetSystemMetrics()
	}
}