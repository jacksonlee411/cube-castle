package neo4j

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ConnectionConfig Neo4j连接配置
type ConnectionConfig struct {
	URI      string
	Username string
	Password string
	Database string
	
	// 连接池配置
	MaxConnections     int
	ConnectionTimeout  time.Duration
	MaxTransactionTime time.Duration
	
	// 重试配置
	MaxRetries   int
	RetryBackoff time.Duration
}

// ConnectionManager Neo4j连接管理器
type ConnectionManager struct {
	driver neo4j.DriverWithContext
	config *ConnectionConfig
	
	// 新增：指标统计
	metrics *ConnectionManagerMetrics
	
	// 新增：重试配置
	retryConfig *RetryConfig
}

// NewConnectionManager 创建Neo4j连接管理器
func NewConnectionManager(config *ConnectionConfig) (ConnectionManagerInterface, error) {
	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Neo4j configuration: %w", err)
	}
	
	// 创建驱动配置
	driverConfig := func(conf *neo4j.Config) {
		conf.MaxConnectionPoolSize = config.MaxConnections
		conf.ConnectionAcquisitionTimeout = config.ConnectionTimeout
		conf.MaxTransactionRetryTime = config.MaxTransactionTime
		
		// Neo4j Go Driver v5不再支持设置Encrypted属性
		// 加密由URI scheme决定 (bolt:// 或 bolt+s://)
	}
	
	// 创建驱动
	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.Username, config.Password, ""),
		driverConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}
	
	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}
	
	log.Printf("✅ Neo4j connection established successfully: %s", config.URI)
	
	// 初始化指标和重试配置
	metrics := &ConnectionManagerMetrics{
		LastErrorTime: time.Time{},
	}
	
	return &ConnectionManager{
		driver:      driver,
		config:      config,
		metrics:     metrics,
		retryConfig: DefaultRetryConfig(),
	}, nil
}

// GetSession 获取Neo4j会话
func (cm *ConnectionManager) GetSession(ctx context.Context) neo4j.SessionWithContext {
	return cm.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: cm.config.Database,
	})
}

// ExecuteWrite 执行写事务
func (cm *ConnectionManager) ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error) {
	session := cm.GetSession(ctx)
	defer session.Close(ctx)
	
	return session.ExecuteWrite(ctx, work)
}

// ExecuteRead 执行读事务
func (cm *ConnectionManager) ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error) {
	session := cm.GetSession(ctx)
	defer session.Close(ctx)
	
	return session.ExecuteRead(ctx, work)
}

// ExecuteWithRetry 带重试的事务执行
func (cm *ConnectionManager) ExecuteWithRetry(ctx context.Context, work func(ctx context.Context) error) error {
	var lastErr error
	
	for attempt := 0; attempt <= cm.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoff := time.Duration(attempt) * cm.config.RetryBackoff
			time.Sleep(backoff)
			log.Printf("🔄 Retrying Neo4j operation (attempt %d/%d)", attempt, cm.config.MaxRetries)
		}
		
		if err := work(ctx); err != nil {
			lastErr = err
			
			// 判断是否为可重试错误
			if !isRetryableError(err) {
				return err
			}
			continue
		}
		
		return nil
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", cm.config.MaxRetries, lastErr)
}

// Health 健康检查
func (cm *ConnectionManager) Health(ctx context.Context) error {
	session := cm.GetSession(ctx)
	defer session.Close(ctx)
	
	_, err := session.Run(ctx, "RETURN 1", nil)
	return err
}

// Close 关闭连接
func (cm *ConnectionManager) Close(ctx context.Context) error {
	if cm.driver != nil {
		return cm.driver.Close(ctx)
	}
	return nil
}

// GetStatistics 获取连接统计信息
func (cm *ConnectionManager) GetStatistics() map[string]interface{} {
	stats := map[string]interface{}{
		"type":                "real",
		"uri":                 cm.config.URI,
		"database":            cm.config.Database,
		"max_connections":     cm.config.MaxConnections,
		"connection_timeout":  cm.config.ConnectionTimeout.String(),
		"status":              "connected",
		
		// 指标统计
		"total_operations":    cm.metrics.TotalOperations,
		"successful_ops":      cm.metrics.SuccessfulOps,
		"failed_ops":          cm.metrics.FailedOps,
		"error_rate":          cm.metrics.ErrorRate,
		"average_latency":     cm.metrics.AverageLatency.String(),
		"total_retries":       cm.metrics.TotalRetries,
		"retry_success_rate":  cm.metrics.RetrySuccessRate,
	}
	
	if !cm.metrics.LastErrorTime.IsZero() {
		stats["last_error"] = cm.metrics.LastError
		stats["last_error_time"] = cm.metrics.LastErrorTime.Format(time.RFC3339)
	}
	
	return stats
}

// GetType 获取连接管理器类型
func (cm *ConnectionManager) GetType() ConnectionManagerType {
	return ConnectionManagerTypeReal
}

// Neo4jConfigFromEnv 从环境变量创建配置
func Neo4jConfigFromEnv() *ConnectionConfig {
	config := &ConnectionConfig{
		URI:      getEnvString("NEO4J_URI", "bolt://localhost:7687"),
		Username: getEnvString("NEO4J_USERNAME", "neo4j"),
		Password: getEnvString("NEO4J_PASSWORD", "password"),
		Database: getEnvString("NEO4J_DATABASE", "neo4j"),
		
		MaxConnections:     getEnvInt("NEO4J_MAX_CONNECTIONS", 50),
		ConnectionTimeout:  getEnvDuration("NEO4J_CONNECTION_TIMEOUT", "30s"),
		MaxTransactionTime: getEnvDuration("NEO4J_MAX_TRANSACTION_TIME", "60s"),
		
		MaxRetries:   getEnvInt("NEO4J_MAX_RETRIES", 3),
		RetryBackoff: getEnvDuration("NEO4J_RETRY_BACKOFF", "1s"),
	}
	
	return config
}

// MockConfig Mock连接管理器配置
type MockConfig struct {
	// 行为配置
	SuccessRate    float64       `json:"success_rate"`    // 成功率 0.0-1.0
	LatencyMin     time.Duration `json:"latency_min"`     // 最小延迟
	LatencyMax     time.Duration `json:"latency_max"`     // 最大延迟
	EnableMetrics  bool          `json:"enable_metrics"`  // 启用统计
	
	// 错误模拟
	ErrorTypes     []string      `json:"error_types"`     // 错误类型列表
	ErrorRate      float64       `json:"error_rate"`      // 错误率
	
	// 连接配置
	MaxConnections int           `json:"max_connections"` // 模拟最大连接数
	DatabaseName   string        `json:"database_name"`   // 数据库名称
}

// DefaultMockConfig 默认Mock配置
func DefaultMockConfig() *MockConfig {
	return &MockConfig{
		SuccessRate:    1.0,
		LatencyMin:     time.Millisecond * 1,
		LatencyMax:     time.Millisecond * 10,
		EnableMetrics:  true,
		ErrorTypes:     []string{},
		ErrorRate:      0.0,
		MaxConnections: 50,
		DatabaseName:   "mock_neo4j",
	}
}

// MockConnectionManager Mock连接管理器（开发环境）
type MockConnectionManager struct {
	connected bool
	config    *MockConfig
	metrics   *ConnectionManagerMetrics
	
	// 新增：操作计数器
	operationCount int64
}

// NewMockConnectionManager 创建Mock连接管理器（使用默认配置）
func NewMockConnectionManager() ConnectionManagerInterface {
	return NewMockConnectionManagerWithConfig(DefaultMockConfig())
}

// NewMockConnectionManagerWithConfig 创建带配置的Mock连接管理器
func NewMockConnectionManagerWithConfig(config *MockConfig) ConnectionManagerInterface {
	log.Printf("🔧 Using Mock Neo4j connection manager (success_rate: %.2f)", config.SuccessRate)
	
	metrics := &ConnectionManagerMetrics{
		LastErrorTime: time.Time{},
	}
	
	return &MockConnectionManager{
		connected:      true,
		config:         config,
		metrics:        metrics,
		operationCount: 0,
	}
}

func (m *MockConnectionManager) GetSession(ctx context.Context) neo4j.SessionWithContext {
	return nil // Mock实现
}

func (m *MockConnectionManager) ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error) {
	start := time.Now()
	m.operationCount++
	m.metrics.TotalOperations++
	
	// 模拟延迟
	m.simulateLatency()
	
	// 模拟错误
	if err := m.simulateError("write"); err != nil {
		m.metrics.FailedOps++
		m.updateMetrics(time.Since(start), err)
		return nil, err
	}
	
	log.Println("📝 Mock Neo4j write operation executed")
	m.metrics.SuccessfulOps++
	m.updateMetrics(time.Since(start), nil)
	return "mock_write_result", nil
}

func (m *MockConnectionManager) ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error) {
	start := time.Now()
	m.operationCount++
	m.metrics.TotalOperations++
	
	// 模拟延迟
	m.simulateLatency()
	
	// 模拟错误
	if err := m.simulateError("read"); err != nil {
		m.metrics.FailedOps++
		m.updateMetrics(time.Since(start), err)
		return nil, err
	}
	
	log.Println("📖 Mock Neo4j read operation executed")
	m.metrics.SuccessfulOps++
	m.updateMetrics(time.Since(start), nil)
	return "mock_read_result", nil
}

func (m *MockConnectionManager) ExecuteWithRetry(ctx context.Context, work func(ctx context.Context) error) error {
	start := time.Now()
	m.operationCount++
	m.metrics.TotalOperations++
	m.metrics.TotalRetries++
	
	// 模拟延迟
	m.simulateLatency()
	
	// 模拟错误
	if err := m.simulateError("retry"); err != nil {
		m.metrics.FailedOps++
		m.updateMetrics(time.Since(start), err)
		return err
	}
	
	log.Println("🔄 Mock Neo4j retry operation executed")
	m.metrics.SuccessfulOps++
	m.updateMetrics(time.Since(start), nil)
	return nil
}

func (m *MockConnectionManager) Health(ctx context.Context) error {
	return nil // Mock始终健康
}

func (m *MockConnectionManager) Close(ctx context.Context) error {
	m.connected = false
	log.Println("🔌 Mock Neo4j connection closed")
	return nil
}

func (m *MockConnectionManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"type":                "mock",
		"status":              "connected",
		"database_name":       m.config.DatabaseName,
		"max_connections":     m.config.MaxConnections,
		"success_rate":        m.config.SuccessRate,
		"latency_range":       fmt.Sprintf("%v-%v", m.config.LatencyMin, m.config.LatencyMax),
		
		// 指标统计
		"total_operations":    m.metrics.TotalOperations,
		"successful_ops":      m.metrics.SuccessfulOps,
		"failed_ops":          m.metrics.FailedOps,
		"error_rate":          m.metrics.ErrorRate,
		"average_latency":     m.metrics.AverageLatency.String(),
		"total_retries":       m.metrics.TotalRetries,
		"retry_success_rate":  m.metrics.RetrySuccessRate,
		"operation_count":     m.operationCount,
	}
}

// GetType 获取连接管理器类型
func (m *MockConnectionManager) GetType() ConnectionManagerType {
	return ConnectionManagerTypeMock
}

// 辅助函数
func validateConfig(config *ConnectionConfig) error {
	if config.URI == "" {
		return fmt.Errorf("URI is required")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("password is required")
	}
	if config.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}
	if config.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}
	return nil
}

func isRetryableError(err error) bool {
	// 判断错误是否可重试
	// 例如：网络错误、临时性错误等
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	retryableErrors := []string{
		"connection",
		"timeout",
		"temporary",
		"network",
		"unavailable",
	}
	
	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if defaultDuration, err := time.ParseDuration(defaultValue); err == nil {
		return defaultDuration
	}
	return time.Second * 30 // 默认30秒
}

// Mock模拟方法

// simulateLatency 模拟延迟
func (m *MockConnectionManager) simulateLatency() {
	if m.config.LatencyMin <= 0 && m.config.LatencyMax <= 0 {
		return
	}
	
	var latency time.Duration
	if m.config.LatencyMax > m.config.LatencyMin {
		diff := m.config.LatencyMax - m.config.LatencyMin
		latency = m.config.LatencyMin + time.Duration(float64(diff)*randomFloat())
	} else {
		latency = m.config.LatencyMin
	}
	
	time.Sleep(latency)
}

// simulateError 模拟错误
func (m *MockConnectionManager) simulateError(operation string) error {
	if m.config.ErrorRate <= 0 {
		return nil
	}
	
	if randomFloat() < m.config.ErrorRate {
		errorMsg := fmt.Sprintf("mock %s operation failed", operation)
		if len(m.config.ErrorTypes) > 0 {
			errorType := m.config.ErrorTypes[int(randomFloat()*float64(len(m.config.ErrorTypes)))]
			errorMsg = fmt.Sprintf("mock %s error: %s", operation, errorType)
		}
		return fmt.Errorf(errorMsg)
	}
	
	return nil
}

// updateMetrics 更新指标
func (m *MockConnectionManager) updateMetrics(duration time.Duration, err error) {
	if !m.config.EnableMetrics {
		return
	}
	
	// 更新延迟统计
	if m.metrics.TotalOperations == 1 {
		m.metrics.AverageLatency = duration
		m.metrics.MinLatency = duration
		m.metrics.MaxLatency = duration
	} else {
		// 计算平均延迟
		totalTime := time.Duration(float64(m.metrics.AverageLatency) * float64(m.metrics.TotalOperations-1))
		m.metrics.AverageLatency = (totalTime + duration) / time.Duration(m.metrics.TotalOperations)
		
		if duration < m.metrics.MinLatency {
			m.metrics.MinLatency = duration
		}
		if duration > m.metrics.MaxLatency {
			m.metrics.MaxLatency = duration
		}
	}
	
	// 更新错误统计
	if err != nil {
		m.metrics.LastError = err.Error()
		m.metrics.LastErrorTime = time.Now()
	}
	
	// 计算错误率
	if m.metrics.TotalOperations > 0 {
		m.metrics.ErrorRate = float64(m.metrics.FailedOps) / float64(m.metrics.TotalOperations)
	}
	
	// 计算重试成功率
	if m.metrics.TotalRetries > 0 {
		successfulRetries := m.metrics.TotalRetries - m.metrics.FailedOps
		m.metrics.RetrySuccessRate = float64(successfulRetries) / float64(m.metrics.TotalRetries)
	}
}

// randomFloat 生成0-1之间的随机浮点数
func randomFloat() float64 {
	// 简单的伪随机数生成
	return float64(time.Now().UnixNano()%1000) / 1000.0
}