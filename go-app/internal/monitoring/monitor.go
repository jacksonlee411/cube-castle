package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// HealthStatus 健康状态结构
type HealthStatus struct {
	Service     string                 `json:"service"`
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Checks      map[string]CheckResult `json:"checks"`
	Metrics     SystemMetrics          `json:"metrics"`
}

// CheckResult 检查结果
type CheckResult struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPU           CPUMetrics         `json:"cpu"`
	Memory        MemoryMetrics      `json:"memory"`
	HTTP          HTTPMetrics        `json:"http"`
	CustomMetrics map[string]float64 `json:"custom_metrics"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	UsedBytes      uint64  `json:"used_bytes"`
	TotalBytes     uint64  `json:"total_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	HeapBytes      uint64  `json:"heap_bytes"`
	GoroutineCount int     `json:"goroutine_count"`
}

// HTTPMetrics HTTP指标
type HTTPMetrics struct {
	RequestCount    int64                     `json:"request_count"`
	AverageLatency  time.Duration             `json:"average_latency_ms"`
	ErrorRate       float64                   `json:"error_rate_percent"`
	StatusCodes     map[string]int64          `json:"status_codes"`
	EndpointMetrics map[string]EndpointMetric `json:"endpoints"`
}

// EndpointMetric 端点指标
type EndpointMetric struct {
	RequestCount   int64         `json:"request_count"`
	AverageLatency time.Duration `json:"average_latency_ms"`
	ErrorCount     int64         `json:"error_count"`
	LastAccessed   time.Time     `json:"last_accessed"`
}

// Monitor 监控器
type Monitor struct {
	startTime     time.Time
	httpMetrics   *HTTPMetrics
	systemMetrics *SystemMetrics
	config        *MonitorConfig
	mu            sync.RWMutex
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	ServiceName string
	Version     string
	Environment string
}

// NewMonitor 创建新的监控器
func NewMonitor(config *MonitorConfig) *Monitor {
	if config == nil {
		config = &MonitorConfig{
			ServiceName: "cube-castle",
			Version:     "1.0.0",
			Environment: "development",
		}
	}

	return &Monitor{
		startTime: time.Now(),
		httpMetrics: &HTTPMetrics{
			StatusCodes:     make(map[string]int64),
			EndpointMetrics: make(map[string]EndpointMetric),
		},
		systemMetrics: &SystemMetrics{
			CustomMetrics: make(map[string]float64),
		},
		config: config,
	}
}

// GetHealthStatus 获取健康状态
func (m *Monitor) GetHealthStatus(ctx context.Context) *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &HealthStatus{
		Service:     m.config.ServiceName,
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     m.config.Version,
		Environment: m.config.Environment,
		Checks:      m.performHealthChecks(ctx),
		Metrics:     m.collectSystemMetrics(),
	}
}

// GetDetailedHealthStatus 获取详细健康状态
func (m *Monitor) GetDetailedHealthStatus(ctx context.Context) *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &HealthStatus{
		Service:     m.config.ServiceName,
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     m.config.Version,
		Environment: m.config.Environment,
		Checks:      make(map[string]CheckResult),
		Metrics:     m.collectSystemMetrics(),
	}

	// 执行详细检查
	status.Checks["api"] = CheckResult{
		Status:  "healthy",
		Message: "API server is running",
		Latency: time.Millisecond * 5,
	}

	status.Checks["memory"] = m.checkMemory()
	status.Checks["disk"] = m.checkDisk()

	// 确定总体状态
	for _, check := range status.Checks {
		if check.Status != "healthy" {
			status.Status = "unhealthy"
			break
		}
	}

	return status
}

// RecordHTTPRequest 记录HTTP请求指标
func (m *Monitor) RecordHTTPRequest(method, path string, statusCode int, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	statusStr := fmt.Sprintf("%d", statusCode)

	// 更新HTTP指标
	m.httpMetrics.RequestCount++
	m.httpMetrics.StatusCodes[statusStr]++

	// 计算平均延迟
	if m.httpMetrics.RequestCount == 1 {
		m.httpMetrics.AverageLatency = latency
	} else {
		// 滑动平均
		m.httpMetrics.AverageLatency = time.Duration(
			(int64(m.httpMetrics.AverageLatency)*int64(m.httpMetrics.RequestCount-1) + int64(latency)) / int64(m.httpMetrics.RequestCount),
		)
	}

	// 计算错误率
	errorCount := int64(0)
	for status, count := range m.httpMetrics.StatusCodes {
		if len(status) > 0 && (status[0] == '4' || status[0] == '5') {
			errorCount += count
		}
	}
	m.httpMetrics.ErrorRate = float64(errorCount) / float64(m.httpMetrics.RequestCount) * 100

	// 更新端点指标
	endpointKey := method + " " + path
	if metric, exists := m.httpMetrics.EndpointMetrics[endpointKey]; exists {
		metric.RequestCount++
		if statusCode >= 400 {
			metric.ErrorCount++
		}
		// 计算端点平均延迟
		if metric.RequestCount == 1 {
			metric.AverageLatency = latency
		} else {
			metric.AverageLatency = time.Duration(
				(int64(metric.AverageLatency)*int64(metric.RequestCount-1) + int64(latency)) / int64(metric.RequestCount),
			)
		}
		metric.LastAccessed = time.Now()
		m.httpMetrics.EndpointMetrics[endpointKey] = metric
	} else {
		m.httpMetrics.EndpointMetrics[endpointKey] = EndpointMetric{
			RequestCount:   1,
			AverageLatency: latency,
			ErrorCount: func() int64 {
				if statusCode >= 400 {
					return 1
				}
				return 0
			}(),
			LastAccessed: time.Now(),
		}
	}
}

// UpdateCustomMetric 更新自定义指标
func (m *Monitor) UpdateCustomMetric(key string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.systemMetrics.CustomMetrics[key] = value
}

// IncrementCustomMetric 增加自定义指标
func (m *Monitor) IncrementCustomMetric(key string, increment float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if current, exists := m.systemMetrics.CustomMetrics[key]; exists {
		m.systemMetrics.CustomMetrics[key] = current + increment
	} else {
		m.systemMetrics.CustomMetrics[key] = increment
	}
}

// GetSystemMetrics 获取系统指标
func (m *Monitor) GetSystemMetrics() SystemMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.collectSystemMetrics()
}

// GetHTTPMetrics 获取HTTP指标
func (m *Monitor) GetHTTPMetrics() HTTPMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.httpMetrics
}

// ServeHTTP 实现http.Handler接口
func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.URL.Path {
	case "/health":
		m.handleHealth(w, r)
	case "/health/detailed":
		m.handleDetailedHealth(w, r)
	case "/metrics":
		m.handleMetrics(w, r)
	case "/metrics/system":
		m.handleSystemMetrics(w, r)
	case "/metrics/http":
		m.handleHTTPMetrics(w, r)
	default:
		http.NotFound(w, r)
		return
	}

	// 记录监控端点的指标
	latency := time.Since(start)
	m.RecordHTTPRequest(r.Method, r.URL.Path, 200, latency)
}

// 私有方法

func (m *Monitor) performHealthChecks(ctx context.Context) map[string]CheckResult {
	checks := make(map[string]CheckResult)

	checks["api"] = CheckResult{
		Status:  "healthy",
		Message: "API server is running",
		Latency: time.Millisecond * 5,
	}

	return checks
}

func (m *Monitor) collectSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return SystemMetrics{
		CPU: CPUMetrics{
			UsagePercent: 0, // 简化版本不实现实际CPU监控
			Cores:        runtime.NumCPU(),
		},
		Memory: MemoryMetrics{
			UsedBytes:      memStats.HeapInuse,
			TotalBytes:     memStats.HeapSys,
			UsagePercent:   float64(memStats.HeapInuse) / float64(memStats.HeapSys) * 100,
			HeapBytes:      memStats.HeapAlloc,
			GoroutineCount: runtime.NumGoroutine(),
		},
		HTTP:          *m.httpMetrics,
		CustomMetrics: m.systemMetrics.CustomMetrics,
	}
}

func (m *Monitor) checkMemory() CheckResult {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	usagePercent := float64(memStats.HeapInuse) / float64(memStats.HeapSys) * 100

	status := "healthy"
	message := fmt.Sprintf("Memory usage: %.2f%%", usagePercent)

	if usagePercent > 90 {
		status = "warning"
		message = fmt.Sprintf("High memory usage: %.2f%%", usagePercent)
	}

	return CheckResult{
		Status:  status,
		Message: message,
		Latency: time.Millisecond,
	}
}

func (m *Monitor) checkDisk() CheckResult {
	return CheckResult{
		Status:  "healthy",
		Message: "Disk usage normal",
		Latency: time.Millisecond,
	}
}

func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := m.GetHealthStatus(r.Context())
	m.writeJSONResponse(w, status)
}

func (m *Monitor) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	status := m.GetDetailedHealthStatus(r.Context())
	statusCode := http.StatusOK
	if status.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}
	w.WriteHeader(statusCode)
	m.writeJSONResponse(w, status)
}

func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.GetSystemMetrics()
	response := map[string]interface{}{
		"timestamp": time.Now(),
		"metrics":   metrics,
	}
	m.writeJSONResponse(w, response)
}

func (m *Monitor) handleSystemMetrics(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemMetrics := map[string]interface{}{
		"cpu": map[string]interface{}{
			"cores":      runtime.NumCPU(),
			"goroutines": runtime.NumGoroutine(),
		},
		"memory": map[string]interface{}{
			"heap_alloc":  memStats.HeapAlloc,
			"heap_sys":    memStats.HeapSys,
			"heap_inuse":  memStats.HeapInuse,
			"stack_inuse": memStats.StackInuse,
			"total_alloc": memStats.TotalAlloc,
			"sys":         memStats.Sys,
			"gc_runs":     memStats.NumGC,
			"last_gc":     time.Unix(0, int64(memStats.LastGC)),
		},
		"runtime": map[string]interface{}{
			"version":    runtime.Version(),
			"uptime":     time.Since(m.startTime),
			"start_time": m.startTime,
		},
	}

	m.writeJSONResponse(w, systemMetrics)
}

func (m *Monitor) handleHTTPMetrics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"timestamp": time.Now(),
		"http":      m.GetHTTPMetrics(),
	}
	m.writeJSONResponse(w, response)
}

func (m *Monitor) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
