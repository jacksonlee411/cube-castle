package neo4j

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ConnectionManagerInterface Neo4j连接管理器统一接口
// 简化版本：从3层接口简化到2层，减少复杂性
type ConnectionManagerInterface interface {
	// 核心事务操作
	ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error)
	ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error)
	ExecuteWithRetry(ctx context.Context, work func(ctx context.Context) error) error
	
	// 连接管理
	GetSession(ctx context.Context) neo4j.SessionWithContext
	Health(ctx context.Context) error
	Close(ctx context.Context) error
	
	// 监控统计
	GetStatistics() map[string]interface{}
	
	// 连接管理器类型标识
	GetType() ConnectionManagerType
}

// ConnectionManagerType 连接管理器类型
type ConnectionManagerType string

const (
	ConnectionManagerTypeReal ConnectionManagerType = "real"
	ConnectionManagerTypeMock ConnectionManagerType = "mock"
)

// ConnectionManagerMetrics 连接管理器指标
type ConnectionManagerMetrics struct {
	// 连接统计
	TotalConnections    int64         `json:"total_connections"`
	ActiveConnections   int64         `json:"active_connections"`
	ConnectionsCreated  int64         `json:"connections_created"`
	ConnectionsDestroyed int64        `json:"connections_destroyed"`
	
	// 操作统计
	TotalOperations     int64         `json:"total_operations"`
	SuccessfulOps       int64         `json:"successful_operations"`
	FailedOps           int64         `json:"failed_operations"`
	
	// 性能统计
	AverageLatency      time.Duration `json:"average_latency"`
	MaxLatency          time.Duration `json:"max_latency"`
	MinLatency          time.Duration `json:"min_latency"`
	
	// 错误统计
	ErrorRate           float64       `json:"error_rate"`
	LastError           string        `json:"last_error,omitempty"`
	LastErrorTime       time.Time     `json:"last_error_time,omitempty"`
	
	// 重试统计
	TotalRetries        int64         `json:"total_retries"`
	RetrySuccessRate    float64       `json:"retry_success_rate"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	BaseDelay       time.Duration `json:"base_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	Multiplier      float64       `json:"multiplier"`
	EnableJitter    bool          `json:"enable_jitter"`
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		BaseDelay:    time.Millisecond * 100,
		MaxDelay:     time.Second * 30,
		Multiplier:   2.0,
		EnableJitter: true,
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status      string                 `json:"status"`       // "healthy", "degraded", "unhealthy"
	Timestamp   time.Time             `json:"timestamp"`
	Latency     time.Duration         `json:"latency"`
	Details     map[string]interface{} `json:"details,omitempty"`
	LastCheck   time.Time             `json:"last_check"`
}

// ConnectionManagerFactory 连接管理器工厂
type ConnectionManagerFactory interface {
	CreateReal(config *ConnectionConfig) (ConnectionManagerInterface, error)
	CreateMock(config *MockConfig) ConnectionManagerInterface
}

// standardFactory 标准工厂实现
type standardFactory struct{}

// NewConnectionManagerFactory 创建连接管理器工厂
func NewConnectionManagerFactory() ConnectionManagerFactory {
	return &standardFactory{}
}

func (f *standardFactory) CreateReal(config *ConnectionConfig) (ConnectionManagerInterface, error) {
	return NewConnectionManager(config)
}

func (f *standardFactory) CreateMock(config *MockConfig) ConnectionManagerInterface {
	return NewMockConnectionManagerWithConfig(config)
}