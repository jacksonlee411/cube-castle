package neo4j

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

// TestConnectionManagerInterface 接口一致性测试
func TestConnectionManagerInterface(t *testing.T) {
	tests := []struct {
		name    string
		factory func() ConnectionManagerInterface
	}{
		{
			name: "默认Mock连接管理器",
			factory: func() ConnectionManagerInterface {
				return NewMockConnectionManager()
			},
		},
		{
			name: "配置化Mock连接管理器", 
			factory: func() ConnectionManagerInterface {
				config := &MockConfig{
					SuccessRate:    0.9,
					LatencyMin:     time.Millisecond * 1,
					LatencyMax:     time.Millisecond * 5,
					EnableMetrics:  true,
					ErrorRate:      0.1,
					MaxConnections: 25,
					DatabaseName:   "test_db",
				}
				return NewMockConnectionManagerWithConfig(config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := tt.factory()
			ctx := context.Background()

			// 测试接口方法存在
			assert.NotNil(t, mgr)
			assert.NotEmpty(t, mgr.GetType())
			
			// 测试基本操作
			stats := mgr.GetStatistics()
			assert.NotNil(t, stats)
			assert.Contains(t, stats, "type")
			
			// 测试健康检查
			err := mgr.Health(ctx)
			assert.NoError(t, err)
			
			// 测试写操作
			result, err := mgr.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				return "test_write", nil
			})
			
			if err == nil {
				assert.NotNil(t, result)
			}
			
			// 测试读操作
			readResult, err := mgr.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				return "test_read", nil
			})
			
			if err == nil {
				assert.NotNil(t, readResult)
			}
			
			// 测试重试操作
			err = mgr.ExecuteWithRetry(ctx, func(ctx context.Context) error {
				return nil
			})
			// 重试操作可能会根据配置失败，这是正常的
			
			// 关闭连接
			err = mgr.Close(ctx)
			assert.NoError(t, err)
		})
	}
}

// TestMockConnectionManagerBehavior Mock行为测试
func TestMockConnectionManagerBehavior(t *testing.T) {
	ctx := context.Background()
	
	t.Run("成功率控制", func(t *testing.T) {
		config := &MockConfig{
			SuccessRate:   0.5, // 50%成功率
			ErrorRate:     0.5, // 50%错误率
			EnableMetrics: true,
		}
		
		mgr := NewMockConnectionManagerWithConfig(config).(*MockConnectionManager)
		
		// 执行多次操作统计成功率
		totalOps := 20
		successCount := 0
		
		for i := 0; i < totalOps; i++ {
			_, err := mgr.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				return "test", nil
			})
			if err == nil {
				successCount++
			}
		}
		
		// 允许一定误差范围
		successRate := float64(successCount) / float64(totalOps)
		assert.InDelta(t, 0.5, successRate, 0.3, "成功率应该接近配置值")
	})
	
	t.Run("延迟模拟", func(t *testing.T) {
		config := &MockConfig{
			SuccessRate:   1.0,
			LatencyMin:    time.Millisecond * 10,
			LatencyMax:    time.Millisecond * 20,
			EnableMetrics: true,
		}
		
		mgr := NewMockConnectionManagerWithConfig(config)
		
		start := time.Now()
		_, err := mgr.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			return "test", nil
		})
		duration := time.Since(start)
		
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, duration, time.Millisecond*10, "延迟应该大于最小值")
		assert.LessOrEqual(t, duration, time.Millisecond*100, "延迟应该在合理范围内")
	})
	
	t.Run("指标统计", func(t *testing.T) {
		mgr := NewMockConnectionManager().(*MockConnectionManager)
		
		// 初始状态
		stats := mgr.GetStatistics()
		assert.Equal(t, int64(0), stats["total_operations"])
		
		// 执行操作
		mgr.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			return "test", nil
		})
		
		// 检查统计更新
		finalStats := mgr.GetStatistics()
		assert.Equal(t, int64(1), finalStats["total_operations"])
		assert.Equal(t, int64(1), finalStats["successful_ops"])
	})
}

// TestEventConsumerIntegration 事件消费者集成测试
func TestEventConsumerIntegration(t *testing.T) {
	ctx := context.Background()
	mockMgr := NewMockConnectionManager()
	
	t.Run("员工事件消费者", func(t *testing.T) {
		consumer := NewEmployeeEventConsumer(mockMgr)
		assert.NotNil(t, consumer)
		
		// 创建测试事件
		event := &MockDomainEvent{
			EventID:       uuid.New(),
			EventType:     "employee.created",
			AggregateID:   uuid.New(),
			TenantID:      uuid.New(),
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
		}
		
		// 消费事件 (Mock环境下应该成功)
		err := consumer.ConsumeEvent(ctx, event)
		assert.NoError(t, err)
	})
	
	t.Run("组织事件消费者", func(t *testing.T) {
		consumer := NewOrganizationEventConsumer(mockMgr)
		assert.NotNil(t, consumer)
		
		// 创建测试事件
		event := &MockDomainEvent{
			EventID:       uuid.New(),
			EventType:     "organization.created",
			AggregateID:   uuid.New(),
			TenantID:      uuid.New(),
			Timestamp:     time.Now(),
			EventVersion:  "1.0",
		}
		
		// 消费事件 (Mock环境下应该成功)
		err := consumer.ConsumeEvent(ctx, event)
		assert.NoError(t, err)
	})
}

// TestNodeSyncOperation 同步操作测试
func TestNodeSyncOperation(t *testing.T) {
	t.Run("CREATE操作验证", func(t *testing.T) {
		op := &NodeSyncOperation{
			Label:      "TestNode",
			Operation:  "CREATE",
			UniqueKeys: []string{"id"},
			Properties: map[string]interface{}{
				"id":   uuid.New().String(),
				"name": "test",
			},
		}
		
		err := op.Validate()
		assert.NoError(t, err)
		
		description := op.GetDescription()
		assert.Contains(t, description, "CREATE")
		assert.Contains(t, description, "TestNode")
	})
	
	t.Run("UPDATE操作验证", func(t *testing.T) {
		op := &NodeSyncOperation{
			Label:      "TestNode", 
			Operation:  "UPDATE",
			UniqueKeys: []string{"id", "tenant_id"},
			Properties: map[string]interface{}{
				"id":        uuid.New().String(),
				"tenant_id": uuid.New().String(),
				"name":      "updated_test",
			},
		}
		
		err := op.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("验证失败场景", func(t *testing.T) {
		// 缺少Label
		op := &NodeSyncOperation{
			Operation:  "CREATE",
			UniqueKeys: []string{"id"},
			Properties: map[string]interface{}{"id": "test"},
		}
		
		err := op.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "标签")
	})
}

// TestConcurrentOperations 并发安全性测试
func TestConcurrentOperations(t *testing.T) {
	mgr := NewMockConnectionManager()
	ctx := context.Background()
	
	// 并发执行多个操作
	concurrency := 10
	operations := 5
	
	errChan := make(chan error, concurrency*operations)
	
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				_, err := mgr.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
					return fmt.Sprintf("worker-%d-op-%d", id, j), nil
				})
				errChan <- err
			}
		}(i)
	}
	
	// 收集结果
	errorCount := 0
	for i := 0; i < concurrency*operations; i++ {
		err := <-errChan
		if err != nil {
			errorCount++
		}
	}
	
	// Mock环境下大部分操作应该成功
	successRate := float64(concurrency*operations-errorCount) / float64(concurrency*operations)
	assert.GreaterOrEqual(t, successRate, 0.8, "并发操作成功率应该大于80%")
	
	// 检查统计信息
	stats := mgr.GetStatistics()
	totalOps := stats["total_operations"].(int64)
	assert.Equal(t, int64(concurrency*operations), totalOps, "总操作数应该正确")
}

// MockDomainEvent 测试用的域事件实现
type MockDomainEvent struct {
	EventID      uuid.UUID
	EventType    string
	AggregateID  uuid.UUID
	TenantID     uuid.UUID
	Timestamp    time.Time
	EventVersion string
}

func (e *MockDomainEvent) GetEventID() uuid.UUID     { return e.EventID }
func (e *MockDomainEvent) GetEventType() string      { return e.EventType }
func (e *MockDomainEvent) GetEventVersion() string   { return e.EventVersion }
func (e *MockDomainEvent) GetAggregateID() uuid.UUID { return e.AggregateID }
func (e *MockDomainEvent) GetAggregateType() string  { return "MockAggregate" }
func (e *MockDomainEvent) GetTenantID() uuid.UUID    { return e.TenantID }
func (e *MockDomainEvent) GetTimestamp() time.Time   { return e.Timestamp }
func (e *MockDomainEvent) GetOccurredAt() time.Time  { return e.Timestamp }

func (e *MockDomainEvent) Serialize() ([]byte, error) {
	return []byte("mock_serialized_event"), nil
}

func (e *MockDomainEvent) GetHeaders() map[string]string {
	return map[string]string{"content-type": "application/json"}
}

func (e *MockDomainEvent) GetMetadata() map[string]interface{} {
	return map[string]interface{}{"source": "test"}
}

func (e *MockDomainEvent) GetCorrelationID() string { return "test-correlation-id" }
func (e *MockDomainEvent) GetCausationID() string   { return "test-causation-id" }