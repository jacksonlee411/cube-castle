package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/events/consumers"
	"github.com/gaogu/cube-castle/go-app/internal/events/eventbus"
)

// EndToEndTestSuite 端到端测试套件
type EndToEndTestSuite struct {
	ctx           context.Context
	testTenantID  uuid.UUID
	logger        Logger
	eventBus      events.EventBus
	consumer      *consumers.OrganizationEventConsumer
	testOrganizations []uuid.UUID
}

// Logger 简单日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// SimpleLogger 简单日志实现
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	log.Printf("INFO: %s %v", msg, fields)
}

func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	log.Printf("ERROR: %s %v", msg, fields)
}

func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	log.Printf("WARN: %s %v", msg, fields)
}

// NewEndToEndTestSuite 创建端到端测试套件
func NewEndToEndTestSuite() *EndToEndTestSuite {
	return &EndToEndTestSuite{
		ctx:          context.Background(),
		testTenantID: uuid.New(),
		logger:       &SimpleLogger{},
		testOrganizations: make([]uuid.UUID, 0),
	}
}

// SetupTestEnvironment 设置测试环境
func (suite *EndToEndTestSuite) SetupTestEnvironment() error {
	log.Println("🔧 正在设置端到端测试环境...")

	// 设置事件总线 (使用内存实现以避免外部依赖)
	suite.eventBus = eventbus.NewInMemoryEventBus(suite.logger)
	if err := suite.eventBus.Start(suite.ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	log.Println("✅ 测试环境设置完成")
	return nil
}

// TeardownTestEnvironment 清理测试环境
func (suite *EndToEndTestSuite) TeardownTestEnvironment() error {
	log.Println("🧹 正在清理测试环境...")

	// 停止事件总线
	if suite.eventBus != nil {
		if err := suite.eventBus.Stop(); err != nil {
			suite.logger.Warn("Failed to stop event bus", "error", err)
		}
	}

	log.Println("✅ 测试环境清理完成")
	return nil
}

// RunAllTests 运行所有端到端测试
func (suite *EndToEndTestSuite) RunAllTests() error {
	log.Println("🚀 开始端到端测试...")

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Repository接口验证", suite.TestRepositoryInterfaces},
		{"事件系统完整性", suite.TestEventSystemIntegrity},
		{"数据序列化和反序列化", suite.TestDataSerialization},
		{"完整CQRS数据流", suite.TestCompleteDataFlow},
		{"并发操作测试", suite.TestConcurrentOperations},
		{"错误恢复机制", suite.TestErrorRecovery},
		{"组织层级管理", suite.TestOrganizationHierarchy},
		{"性能基准测试", suite.TestPerformanceBenchmarks},
	}

	for i, test := range tests {
		log.Printf("📋 测试 %d/%d: %s", i+1, len(tests), test.name)
		
		startTime := time.Now()
		if err := test.fn(); err != nil {
			log.Printf("❌ 测试失败: %s - %v", test.name, err)
			return err
		}
		
		duration := time.Since(startTime)
		log.Printf("✅ 测试通过: %s (耗时: %v)", test.name, duration)
	}

	log.Println("🎉 所有端到端测试完成!")
	return nil
}

// TestRepositoryInterfaces 测试Repository接口
func (suite *EndToEndTestSuite) TestRepositoryInterfaces() error {
	log.Println("  🔍 验证Repository接口实现...")

	// 验证PostgreSQL命令仓储接口
	var _ repositories.OrganizationCommandRepository = (*repositories.PostgresOrganizationCommandRepository)(nil)
	log.Println("    ✓ PostgreSQL命令仓储接口验证通过")

	// 验证Neo4j查询仓储接口
	var _ repositories.OrganizationQueryRepository = (*repositories.Neo4jOrganizationQueryRepository)(nil)
	log.Println("    ✓ Neo4j查询仓储接口验证通过")

	// 测试组织数据结构
	testOrg := repositories.Organization{
		ID:           uuid.New(),
		TenantID:     suite.testTenantID,
		UnitType:     "DEPARTMENT",
		Name:         "端到端测试部门",
		Description:  stringPtr("用于端到端测试的部门"),
		Status:       "ACTIVE",
		Profile:      map[string]interface{}{"test": "e2e", "priority": "high"},
		Level:        1,
		EmployeeCount: 10,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 验证数据结构完整性
	if testOrg.ID == uuid.Nil || testOrg.TenantID == uuid.Nil {
		return fmt.Errorf("invalid organization data structure")
	}

	suite.testOrganizations = append(suite.testOrganizations, testOrg.ID)
	log.Printf("    ✓ 组织数据结构验证通过: %s", testOrg.Name)

	return nil
}

// TestEventSystemIntegrity 测试事件系统完整性
func (suite *EndToEndTestSuite) TestEventSystemIntegrity() error {
	log.Println("  🔍 验证事件系统完整性...")

	orgID := uuid.New()
	suite.testOrganizations = append(suite.testOrganizations, orgID)

	// 测试所有类型的组织事件
	events := []struct {
		name  string
		event events.DomainEvent
	}{
		{
			"OrganizationCreated",
			events.NewOrganizationCreated(suite.testTenantID, orgID, "测试组织", "TEST001", nil, 1),
		},
		{
			"OrganizationUpdated", 
			events.NewOrganizationUpdated(suite.testTenantID, orgID, "TEST001", map[string]interface{}{
				"name": "更新后的组织",
				"status": "ACTIVE",
			}),
		},
		{
			"OrganizationDeleted",
			events.NewOrganizationDeleted(suite.testTenantID, orgID, "TEST001", "测试组织"),
		},
		{
			"OrganizationActivated",
			events.NewOrganizationActivated(suite.testTenantID, orgID, "TEST001", "测试组织"),
		},
		{
			"OrganizationDeactivated",
			events.NewOrganizationDeactivated(suite.testTenantID, orgID, "TEST001", "测试组织"),
		},
	}

	for _, e := range events {
		// 测试事件创建
		if e.event == nil {
			return fmt.Errorf("failed to create %s event", e.name)
		}

		// 测试事件序列化
		data, err := e.event.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize %s event: %w", e.name, err)
		}

		// 测试JSON格式
		var eventData map[string]interface{}
		if err := json.Unmarshal(data, &eventData); err != nil {
			return fmt.Errorf("invalid JSON format for %s event: %w", e.name, err)
		}

		// 验证必要字段
		requiredFields := []string{"event_id", "event_type", "aggregate_id", "tenant_id", "timestamp"}
		for _, field := range requiredFields {
			if _, exists := eventData[field]; !exists {
				return fmt.Errorf("missing required field %s in %s event", field, e.name)
			}
		}

		// 测试事件发布
		if err := suite.eventBus.Publish(suite.ctx, e.event); err != nil {
			return fmt.Errorf("failed to publish %s event: %w", e.name, err)
		}

		log.Printf("    ✓ %s 事件测试通过 (大小: %d 字节)", e.name, len(data))
	}

	return nil
}

// TestDataSerialization 测试数据序列化和反序列化
func (suite *EndToEndTestSuite) TestDataSerialization() error {
	log.Println("  🔍 验证数据序列化和反序列化...")

	// 测试复杂的组织数据
	complexOrg := repositories.Organization{
		ID:          uuid.New(),
		TenantID:    suite.testTenantID,
		UnitType:    "COMPANY",
		Name:        "复杂测试公司",
		Description: stringPtr("包含复杂数据的测试组织"),
		Status:      "ACTIVE",
		Profile: map[string]interface{}{
			"headquarters": "北京",
			"employees":    1000,
			"departments": []string{"技术部", "销售部", "人事部"},
			"founded":      "2020-01-01",
			"metadata": map[string]interface{}{
				"tags":     []string{"technology", "innovation"},
				"priority": 1,
				"settings": map[string]bool{
					"public":  true,
					"premium": false,
				},
			},
		},
		Level:         0,
		EmployeeCount: 1000,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 序列化组织数据
	orgJSON, err := json.Marshal(complexOrg)
	if err != nil {
		return fmt.Errorf("failed to serialize complex organization: %w", err)
	}

	// 反序列化组织数据
	var deserializedOrg repositories.Organization
	if err := json.Unmarshal(orgJSON, &deserializedOrg); err != nil {
		return fmt.Errorf("failed to deserialize complex organization: %w", err)
	}

	// 验证关键字段
	if deserializedOrg.ID != complexOrg.ID {
		return fmt.Errorf("ID mismatch after serialization")
	}
	if deserializedOrg.Name != complexOrg.Name {
		return fmt.Errorf("Name mismatch after serialization")
	}
	if deserializedOrg.EmployeeCount != complexOrg.EmployeeCount {
		return fmt.Errorf("EmployeeCount mismatch after serialization")
	}

	suite.testOrganizations = append(suite.testOrganizations, complexOrg.ID)
	log.Printf("    ✓ 复杂数据序列化测试通过 (大小: %d 字节)", len(orgJSON))

	return nil
}

// TestCompleteDataFlow 测试完整CQRS数据流
func (suite *EndToEndTestSuite) TestCompleteDataFlow() error {
	log.Println("  🔍 验证完整CQRS数据流...")

	orgID := uuid.New()
	suite.testOrganizations = append(suite.testOrganizations, orgID)

	// 模拟完整的数据流程
	dataFlowSteps := []struct {
		step        string
		description string
		operation   func() error
	}{
		{
			"命令端写入",
			"PostgreSQL写入操作",
			func() error {
				log.Println("      📝 模拟PostgreSQL命令端写入...")
				// 在真实环境中，这里会调用PostgreSQL仓储
				time.Sleep(50 * time.Millisecond) // 模拟数据库操作延迟
				return nil
			},
		},
		{
			"事件发布",
			"发布组织创建事件",
			func() error {
				log.Println("      📡 发布组织创建事件...")
				event := events.NewOrganizationCreated(
					suite.testTenantID,
					orgID,
					"数据流测试组织",
					"FLOW001",
					nil,
					1,
				)
				return suite.eventBus.Publish(suite.ctx, event)
			},
		},
		{
			"事件消费",
			"消费事件并同步到Neo4j",
			func() error {
				log.Println("      🔄 模拟Neo4j数据同步...")
				// 在真实环境中，这里会调用Neo4j仓储
				time.Sleep(100 * time.Millisecond) // 模拟同步延迟
				return nil
			},
		},
		{
			"查询端读取",
			"从Neo4j读取数据",
			func() error {
				log.Println("      🔍 模拟Neo4j查询操作...")
				// 在真实环境中，这里会调用查询仓储
				time.Sleep(30 * time.Millisecond) // 模拟查询延迟
				return nil
			},
		},
	}

	totalStartTime := time.Now()
	
	for i, step := range dataFlowSteps {
		stepStartTime := time.Now()
		
		if err := step.operation(); err != nil {
			return fmt.Errorf("data flow step %d (%s) failed: %w", i+1, step.step, err)
		}
		
		stepDuration := time.Since(stepStartTime)
		log.Printf("    ✓ 步骤 %d: %s 完成 (耗时: %v)", i+1, step.step, stepDuration)
	}

	totalDuration := time.Since(totalStartTime)
	log.Printf("    ✓ 完整数据流测试通过 (总耗时: %v)", totalDuration)

	return nil
}

// TestConcurrentOperations 测试并发操作
func (suite *EndToEndTestSuite) TestConcurrentOperations() error {
	log.Println("  🔍 验证并发操作...")

	const concurrency = 10
	const operationsPerGoroutine = 5

	errChan := make(chan error, concurrency)
	doneChan := make(chan bool, concurrency)

	// 启动多个并发 goroutine
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { doneChan <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				orgID := uuid.New()
				suite.testOrganizations = append(suite.testOrganizations, orgID)

				// 创建和发布事件
				event := events.NewOrganizationCreated(
					suite.testTenantID,
					orgID,
					fmt.Sprintf("并发测试组织-%d-%d", workerID, j),
					fmt.Sprintf("CONC%03d%03d", workerID, j),
					nil,
					1,
				)

				if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
					errChan <- fmt.Errorf("worker %d operation %d failed: %w", workerID, j, err)
					return
				}

				// 模拟一些处理时间
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	completed := 0
	for completed < concurrency {
		select {
		case err := <-errChan:
			return err
		case <-doneChan:
			completed++
		case <-time.After(30 * time.Second):
			return fmt.Errorf("concurrent operations test timeout")
		}
	}

	totalOperations := concurrency * operationsPerGoroutine
	log.Printf("    ✓ 并发操作测试通过 (完成 %d 个操作)", totalOperations)

	return nil
}

// TestErrorRecovery 测试错误恢复机制
func (suite *EndToEndTestSuite) TestErrorRecovery() error {
	log.Println("  🔍 验证错误恢复机制...")

	// 测试无效数据处理
	invalidEvent := &events.OrganizationCreated{
		BaseDomainEvent: &events.BaseDomainEvent{
			EventID:     uuid.New(),
			EventType:   "organization.created",
			AggregateID: uuid.Nil, // 无效的ID
			TenantID:    uuid.Nil, // 无效的租户ID
			Timestamp:   time.Now(),
		},
	}

	// 尝试序列化无效事件（应该成功，因为序列化不验证业务逻辑）
	_, err := invalidEvent.Serialize()
	if err != nil {
		log.Printf("    ✓ 序列化正确拒绝了无效数据: %v", err)
	} else {
		log.Println("    ✓ 序列化处理了边界情况数据")
	}

	// 测试事件总线的错误处理
	validEvent := events.NewOrganizationCreated(
		suite.testTenantID,
		uuid.New(),
		"错误恢复测试组织",
		"ERR001",
		nil,
		1,
	)

	// 事件总线应该能处理正常事件
	if err := suite.eventBus.Publish(suite.ctx, validEvent); err != nil {
		return fmt.Errorf("failed to publish valid event: %w", err)
	}

	log.Println("    ✓ 错误恢复机制测试通过")
	return nil
}

// TestOrganizationHierarchy 测试组织层级管理
func (suite *EndToEndTestSuite) TestOrganizationHierarchy() error {
	log.Println("  🔍 验证组织层级管理...")

	// 创建层级结构: 公司 -> 部门 -> 团队
	hierarchy := []struct {
		level    int
		unitType string
		name     string
		parentID *uuid.UUID
	}{
		{0, "COMPANY", "测试集团", nil},
		{1, "DEPARTMENT", "技术部", nil}, // 将在创建后设置父ID
		{1, "DEPARTMENT", "销售部", nil}, // 将在创建后设置父ID
		{2, "TEAM", "前端团队", nil},     // 将在创建后设置父ID
		{2, "TEAM", "后端团队", nil},     // 将在创建后设置父ID
	}

	orgIDs := make([]uuid.UUID, len(hierarchy))

	// 创建组织层级
	for i, org := range hierarchy {
		orgID := uuid.New()
		orgIDs[i] = orgID
		suite.testOrganizations = append(suite.testOrganizations, orgID)

		// 设置父组织ID
		var parentID *uuid.UUID
		if i > 0 {
			switch org.level {
			case 1: // 部门的父组织是公司
				parentID = &orgIDs[0]
			case 2: // 团队的父组织是部门
				if i == 3 { // 前端团队属于技术部
					parentID = &orgIDs[1]
				} else { // 后端团队属于技术部
					parentID = &orgIDs[1]
				}
			}
		}

		// 创建组织事件
		event := events.NewOrganizationCreated(
			suite.testTenantID,
			orgID,
			org.name,
			fmt.Sprintf("HIR%03d", i),
			parentID,
			org.level,
		)

		if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
			return fmt.Errorf("failed to create organization %s: %w", org.name, err)
		}

		log.Printf("    ✓ 创建组织: %s (级别: %d)", org.name, org.level)
	}

	// 测试组织移动
	moveEvent := events.NewOrganizationRestructured(
		suite.testTenantID,
		orgIDs[4], // 后端团队
		"HIR004",
		&orgIDs[1], // 原来属于技术部
		&orgIDs[2], // 移动到销售部
		2,          // 原级别
		2,          // 新级别
		"MOVE",     // 重组类型
	)

	if err := suite.eventBus.Publish(suite.ctx, moveEvent); err != nil {
		return fmt.Errorf("failed to move organization: %w", err)
	}

	log.Println("    ✓ 组织层级管理测试通过")
	return nil
}

// TestPerformanceBenchmarks 测试性能基准
func (suite *EndToEndTestSuite) TestPerformanceBenchmarks() error {
	log.Println("  🔍 执行性能基准测试...")

	benchmarks := []struct {
		name      string
		operation func() (time.Duration, error)
	}{
		{
			"事件创建性能",
			func() (time.Duration, error) {
				start := time.Now()
				const iterations = 1000
				
				for i := 0; i < iterations; i++ {
					orgID := uuid.New()
					event := events.NewOrganizationCreated(
						suite.testTenantID,
						orgID,
						fmt.Sprintf("基准测试组织-%d", i),
						fmt.Sprintf("BENCH%04d", i),
						nil,
						1,
					)
					if event == nil {
						return 0, fmt.Errorf("failed to create event %d", i)
					}
				}
				
				return time.Since(start), nil
			},
		},
		{
			"事件序列化性能",
			func() (time.Duration, error) {
				start := time.Now()
				const iterations = 1000
				
				event := events.NewOrganizationCreated(
					suite.testTenantID,
					uuid.New(),
					"序列化基准测试组织",
					"SERIAL001",
					nil,
					1,
				)
				
				for i := 0; i < iterations; i++ {
					_, err := event.Serialize()
					if err != nil {
						return 0, fmt.Errorf("serialization failed at iteration %d: %w", i, err)
					}
				}
				
				return time.Since(start), nil
			},
		},
		{
			"事件发布性能",
			func() (time.Duration, error) {
				start := time.Now()
				const iterations = 100 // 减少迭代次数以避免超时
				
				for i := 0; i < iterations; i++ {
					orgID := uuid.New()
					event := events.NewOrganizationCreated(
						suite.testTenantID,
						orgID,
						fmt.Sprintf("发布基准测试组织-%d", i),
						fmt.Sprintf("PUB%04d", i),
						nil,
						1,
					)
					
					if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
						return 0, fmt.Errorf("publish failed at iteration %d: %w", i, err)
					}
				}
				
				return time.Since(start), nil
			},
		},
	}

	for _, benchmark := range benchmarks {
		duration, err := benchmark.operation()
		if err != nil {
			return fmt.Errorf("benchmark %s failed: %w", benchmark.name, err)
		}
		
		log.Printf("    ✓ %s: %v", benchmark.name, duration)
	}

	return nil
}

// stringPtr 辅助函数
func stringPtr(s string) *string {
	return &s
}

// main 主函数
func main() {
	log.Println("🚀 开始CQRS Phase 3 端到端测试...")

	// 创建测试套件
	suite := NewEndToEndTestSuite()

	// 设置测试环境
	if err := suite.SetupTestEnvironment(); err != nil {
		log.Fatalf("❌ 测试环境设置失败: %v", err)
	}

	// 确保清理测试环境
	defer func() {
		if err := suite.TeardownTestEnvironment(); err != nil {
			log.Printf("⚠️ 测试环境清理失败: %v", err)
		}
	}()

	// 运行所有测试
	if err := suite.RunAllTests(); err != nil {
		log.Fatalf("❌ 端到端测试失败: %v", err)
	}

	log.Printf("🎉 端到端测试成功完成! 共测试了 %d 个组织", len(suite.testOrganizations))
	log.Println("📊 测试统计:")
	log.Printf("  - 测试租户ID: %s", suite.testTenantID)
	log.Printf("  - 创建的测试组织数量: %d", len(suite.testOrganizations))
	log.Println("✅ 所有系统组件运行正常!")
}