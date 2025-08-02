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
}

// NewConnectionManager 创建Neo4j连接管理器
func NewConnectionManager(config *ConnectionConfig) (*ConnectionManager, error) {
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
	
	return &ConnectionManager{
		driver: driver,
		config: config,
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
	// TODO: 实现连接池统计信息
	return map[string]interface{}{
		"uri":              cm.config.URI,
		"database":         cm.config.Database,
		"max_connections":  cm.config.MaxConnections,
		"connection_timeout": cm.config.ConnectionTimeout.String(),
		"status":           "connected",
	}
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

// MockConnectionManager Mock连接管理器（开发环境）
type MockConnectionManager struct {
	connected bool
}

// NewMockConnectionManager 创建Mock连接管理器
func NewMockConnectionManager() *MockConnectionManager {
	log.Println("🔧 Using Mock Neo4j connection manager")
	return &MockConnectionManager{connected: true}
}

func (m *MockConnectionManager) GetSession(ctx context.Context) neo4j.SessionWithContext {
	return nil // Mock实现
}

func (m *MockConnectionManager) ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error) {
	log.Println("📝 Mock Neo4j write operation executed")
	return nil, nil
}

func (m *MockConnectionManager) ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error) {
	log.Println("📖 Mock Neo4j read operation executed")
	return nil, nil
}

func (m *MockConnectionManager) ExecuteWithRetry(ctx context.Context, work func(ctx context.Context) error) error {
	log.Println("🔄 Mock Neo4j retry operation executed")
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
		"type":   "mock",
		"status": "connected",
	}
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