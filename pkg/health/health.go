package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// HealthStatus 表示服务健康状态
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthCheck 表示单个组件的健康检查
type HealthCheck struct {
	Name     string                 `json:"name"`
	Status   HealthStatus           `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Duration time.Duration          `json:"duration"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// ServiceHealth 表示整个服务的健康状态
type ServiceHealth struct {
	Service    string        `json:"service"`
	Version    string        `json:"version,omitempty"`
	Status     HealthStatus  `json:"status"`
	Timestamp  time.Time     `json:"timestamp"`
	Uptime     time.Duration `json:"uptime"`
	Checks     []HealthCheck `json:"checks"`
	Summary    Summary       `json:"summary"`
}

// Summary 提供健康检查的汇总信息
type Summary struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Degraded int `json:"degraded"`
	Failed   int `json:"failed"`
}

// Checker 定义健康检查器接口
type Checker interface {
	Check(ctx context.Context) HealthCheck
}

// PostgreSQLChecker PostgreSQL数据库健康检查器
type PostgreSQLChecker struct {
	Name string
	DB   *sql.DB
}

func (c *PostgreSQLChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name: c.Name,
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 执行健康检查查询
	var result int
	err := c.DB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	
	check.Duration = time.Since(start)
	
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Database query failed: %v", err)
		return check
	}

	// 检查连接池状态
	stats := c.DB.Stats()
	check.Status = StatusHealthy
	check.Message = "Database connection healthy"
	check.Details = map[string]interface{}{
		"open_connections": stats.OpenConnections,
		"in_use":          stats.InUse,
		"idle":            stats.Idle,
		"max_open":        stats.MaxOpenConnections,
	}

	// 如果连接使用率过高，标记为降级
	if stats.InUse > 0 && float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.8 {
		check.Status = StatusDegraded
		check.Message = "Database connection pool usage high"
	}

	return check
}

// RedisChecker Redis健康检查器
type RedisChecker struct {
	Name   string
	Client *redis.Client
}

func (c *RedisChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name: c.Name,
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 执行PING命令
	pong, err := c.Client.Ping(ctx).Result()
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Redis ping failed: %v", err)
		return check
	}

	// 获取Redis信息
	info, err := c.Client.Info(ctx, "memory").Result()
	if err != nil {
		check.Status = StatusDegraded
		check.Message = "Redis ping successful but info query failed"
		check.Details = map[string]interface{}{
			"ping_response": pong,
		}
		return check
	}

	check.Status = StatusHealthy
	check.Message = "Redis connection healthy"
	check.Details = map[string]interface{}{
		"ping_response": pong,
		"info_available": len(info) > 0,
	}

	return check
}

// Neo4jChecker Neo4j数据库健康检查器
type Neo4jChecker struct {
	Name   string
	Driver neo4j.DriverWithContext
}

func (c *Neo4jChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name: c.Name,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 验证连接
	err := c.Driver.VerifyConnectivity(ctx)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Neo4j connectivity failed: %v", err)
		return check
	}

	// 创建会话并执行简单查询
	session := c.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err = session.Run(ctx, "RETURN 1 as test", nil)
	if err != nil {
		check.Status = StatusDegraded
		check.Message = "Neo4j connected but query failed"
		return check
	}

	check.Status = StatusHealthy
	check.Message = "Neo4j connection healthy"
	check.Details = map[string]interface{}{
		"connectivity": "verified",
		"query_test":   "passed",
	}

	return check
}

// DependencyChecker 服务依赖检查器
type DependencyChecker struct {
	Name           string
	URL            string
	Required       bool
	CheckInterval  time.Duration
	MaxRetries     int
}

func (c *DependencyChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name: c.Name,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 尝试多次检查依赖服务
	var lastErr error
	for i := 0; i < c.MaxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if i < c.MaxRetries-1 {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // 退避策略
			}
			continue
		}
		defer resp.Body.Close()

		check.Duration = time.Since(start)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			check.Status = StatusHealthy
			check.Message = fmt.Sprintf("Dependency %s is healthy", c.Name)
			check.Details = map[string]interface{}{
				"url":          c.URL,
				"status_code":  resp.StatusCode,
				"required":     c.Required,
				"retry_count":  i,
			}
			return check
		}

		lastErr = fmt.Errorf("dependency returned status %d", resp.StatusCode)
	}

	check.Duration = time.Since(start)
	
	if c.Required {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Required dependency %s is unavailable: %v", c.Name, lastErr)
	} else {
		check.Status = StatusDegraded
		check.Message = fmt.Sprintf("Optional dependency %s is unavailable: %v", c.Name, lastErr)
	}

	check.Details = map[string]interface{}{
		"url":        c.URL,
		"required":   c.Required,
		"max_retries": c.MaxRetries,
		"error":      lastErr.Error(),
	}

	return check
}

// StartupChecker 启动时依赖检查器
type StartupChecker struct {
	Name           string
	CheckFunction  func(ctx context.Context) error
	Description    string
}

func (c *StartupChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name: c.Name,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := c.CheckFunction(ctx)
	check.Duration = time.Since(start)

	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Startup check failed: %v", err)
		check.Details = map[string]interface{}{
			"description": c.Description,
			"error":       err.Error(),
		}
	} else {
		check.Status = StatusHealthy
		check.Message = "Startup check passed"
		check.Details = map[string]interface{}{
			"description": c.Description,
		}
	}

	return check
}

// HealthManager 健康检查管理器
type HealthManager struct {
	serviceName string
	version     string
	startTime   time.Time
	checkers    []Checker
}

// NewHealthManager 创建新的健康检查管理器
func NewHealthManager(serviceName, version string) *HealthManager {
	return &HealthManager{
		serviceName: serviceName,
		version:     version,
		startTime:   time.Now(),
		checkers:    make([]Checker, 0),
	}
}

// AddChecker 添加健康检查器
func (hm *HealthManager) AddChecker(checker Checker) {
	hm.checkers = append(hm.checkers, checker)
}

// Check 执行所有健康检查
func (hm *HealthManager) Check(ctx context.Context) ServiceHealth {
	checks := make([]HealthCheck, 0, len(hm.checkers))
	summary := Summary{Total: len(hm.checkers)}

	// 并发执行所有健康检查
	checkChan := make(chan HealthCheck, len(hm.checkers))
	
	for _, checker := range hm.checkers {
		go func(c Checker) {
			checkChan <- c.Check(ctx)
		}(checker)
	}

	// 收集结果
	for i := 0; i < len(hm.checkers); i++ {
		check := <-checkChan
		checks = append(checks, check)

		switch check.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusDegraded:
			summary.Degraded++
		case StatusUnhealthy:
			summary.Failed++
		}
	}

	// 确定整体服务状态
	var overallStatus HealthStatus
	if summary.Failed > 0 {
		overallStatus = StatusUnhealthy
	} else if summary.Degraded > 0 {
		overallStatus = StatusDegraded
	} else {
		overallStatus = StatusHealthy
	}

	return ServiceHealth{
		Service:   hm.serviceName,
		Version:   hm.version,
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(hm.startTime),
		Checks:    checks,
		Summary:   summary,
	}
}

// Handler 创建HTTP健康检查处理器
func (hm *HealthManager) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		health := hm.Check(ctx)

		// 设置适当的HTTP状态码
		statusCode := http.StatusOK
		if health.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if health.Status == StatusDegraded {
			statusCode = http.StatusPartialContent
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		json.NewEncoder(w).Encode(health)
	}
}