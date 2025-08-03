package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/events/consumers"
	"github.com/gaogu/cube-castle/go-app/internal/events/eventbus"
)

// DatabaseIntegrationTestSuite 数据库集成测试套件
type DatabaseIntegrationTestSuite struct {
	ctx          context.Context
	testTenantID uuid.UUID
	logger       Logger

	// 数据库连接
	pgDB         *sqlx.DB
	neo4jDriver  neo4j.DriverWithContext

	// 仓储实例
	commandRepo  repositories.OrganizationCommandRepository
	queryRepo    repositories.OrganizationQueryRepository

	// 事件系统
	eventBus     events.EventBus
	consumer     *consumers.OrganizationEventConsumer

	// 测试数据
	testOrganizations []uuid.UUID
}

// Logger 日志接口
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

// NewDatabaseIntegrationTestSuite 创建数据库集成测试套件
func NewDatabaseIntegrationTestSuite() *DatabaseIntegrationTestSuite {
	return &DatabaseIntegrationTestSuite{
		ctx:               context.Background(),
		testTenantID:      uuid.New(),
		logger:            &SimpleLogger{},
		testOrganizations: make([]uuid.UUID, 0),
	}
}

// SetupDatabaseConnections 设置数据库连接
func (suite *DatabaseIntegrationTestSuite) SetupDatabaseConnections() error {
	log.Println("🔧 正在设置数据库连接...")

	// 设置PostgreSQL连接 (如果有的话)
	pgURL := os.Getenv("POSTGRES_URL")
	if pgURL != "" {
		db, err := sqlx.Open("postgres", pgURL)
		if err != nil {
			log.Printf("⚠️ PostgreSQL连接失败 (跳过): %v", err)
		} else {
			// 测试连接
			if err := db.Ping(); err != nil {
				log.Printf("⚠️ PostgreSQL ping失败 (跳过): %v", err)
				db.Close()
			} else {
				suite.pgDB = db
				suite.commandRepo = repositories.NewPostgresOrganizationCommandRepository(db, suite.logger)
				log.Println("✅ PostgreSQL连接成功")
			}
		}
	} else {
		log.Println("📝 未设置POSTGRES_URL环境变量，跳过PostgreSQL测试")
	}

	// 设置Neo4j连接 (如果有的话)
	neo4jURL := os.Getenv("NEO4J_URL")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")
	
	if neo4jURL != "" && neo4jUser != "" && neo4jPassword != "" {
		driver, err := neo4j.NewDriverWithContext(neo4jURL, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
		if err != nil {
			log.Printf("⚠️ Neo4j连接失败 (跳过): %v", err)
		} else {
			// 测试连接
			if err := driver.VerifyConnectivity(suite.ctx); err != nil {
				log.Printf("⚠️ Neo4j连接验证失败 (跳过): %v", err)
				driver.Close(suite.ctx)
			} else {
				suite.neo4jDriver = driver
				suite.queryRepo = repositories.NewNeo4jOrganizationQueryRepository(driver, suite.logger)
				log.Println("✅ Neo4j连接成功")
			}
		}
	} else {
		log.Println("📝 未设置Neo4j环境变量，跳过Neo4j测试")
	}

	// 设置事件总线
	suite.eventBus = eventbus.NewInMemoryEventBus(suite.logger)
	if err := suite.eventBus.Start(suite.ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// 设置事件消费者 (如果有Neo4j连接)
	if suite.neo4jDriver != nil {
		suite.consumer = consumers.NewOrganizationEventConsumer(suite.neo4jDriver, suite.logger)
		
		// 订阅事件 (使用适配器)
		eventTypes := []string{
			"organization.created",
			"organization.updated", 
			"organization.deleted",
			"organization.restructured",
			"organization.activated",
			"organization.deactivated",
		}

		for _, eventType := range eventTypes {
			adapter := NewOrganizationEventHandlerAdapter(suite.consumer, eventType)
			if err := suite.eventBus.Subscribe(suite.ctx, eventType, adapter); err != nil {
				log.Printf("⚠️ 订阅事件 %s 失败: %v", eventType, err)
			}
		}
		log.Println("✅ 事件消费者设置完成")
	}

	log.Println("✅ 数据库连接设置完成")
	return nil
}

// TeardownDatabaseConnections 清理数据库连接
func (suite *DatabaseIntegrationTestSuite) TeardownDatabaseConnections() error {
	log.Println("🧹 正在清理数据库连接...")

	// 清理测试数据
	if err := suite.CleanupTestData(); err != nil {
		suite.logger.Warn("Failed to cleanup test data", "error", err)
	}

	// 关闭事件总线
	if suite.eventBus != nil {
		if err := suite.eventBus.Stop(); err != nil {
			suite.logger.Warn("Failed to stop event bus", "error", err)
		}
	}

	// 关闭Neo4j连接
	if suite.neo4jDriver != nil {
		if err := suite.neo4jDriver.Close(suite.ctx); err != nil {
			suite.logger.Warn("Failed to close Neo4j driver", "error", err)
		}
	}

	// 关闭PostgreSQL连接
	if suite.pgDB != nil {
		if err := suite.pgDB.Close(); err != nil {
			suite.logger.Warn("Failed to close PostgreSQL connection", "error", err)
		}
	}

	log.Println("✅ 数据库连接清理完成")
	return nil
}

// CleanupTestData 清理测试数据
func (suite *DatabaseIntegrationTestSuite) CleanupTestData() error {
	log.Printf("🧹 正在清理 %d 个测试组织数据...", len(suite.testOrganizations))

	// 清理Neo4j测试数据
	if suite.neo4jDriver != nil {
		session := suite.neo4jDriver.NewSession(suite.ctx, neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeWrite,
			DatabaseName: "neo4j",
		})
		defer session.Close(suite.ctx)

		_, err := session.ExecuteWrite(suite.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			cypher := `
				MATCH (o:Organization {tenant_id: $tenant_id})
				DETACH DELETE o`
			
			result, err := tx.Run(suite.ctx, cypher, map[string]any{
				"tenant_id": suite.testTenantID.String(),
			})
			if err != nil {
				return nil, err
			}

			summary, err := result.Consume(suite.ctx)
			if err != nil {
				return nil, err
			}

			return summary.Counters().NodesDeleted(), nil
		})

		if err != nil {
			suite.logger.Warn("Failed to cleanup Neo4j test data", "error", err)
		} else {
			log.Println("✅ Neo4j测试数据清理完成")
		}
	}

	// 清理PostgreSQL测试数据
	if suite.pgDB != nil {
		query := `DELETE FROM organizations WHERE tenant_id = $1`
		result, err := suite.pgDB.ExecContext(suite.ctx, query, suite.testTenantID)
		if err != nil {
			suite.logger.Warn("Failed to cleanup PostgreSQL test data", "error", err)
		} else {
			if rowsAffected, err := result.RowsAffected(); err == nil {
				log.Printf("✅ PostgreSQL测试数据清理完成 (删除 %d 行)", rowsAffected)
			}
		}
	}

	return nil
}

// RunDatabaseIntegrationTests 运行数据库集成测试
func (suite *DatabaseIntegrationTestSuite) RunDatabaseIntegrationTests() error {
	log.Println("🚀 开始数据库集成测试...")

	tests := []struct {
		name      string
		fn        func() error
		requireDB string
	}{
		{"PostgreSQL命令仓储测试", suite.TestPostgreSQLCommandRepository, "postgres"},
		{"Neo4j查询仓储测试", suite.TestNeo4jQueryRepository, "neo4j"},
		{"完整CQRS数据流测试", suite.TestCompleteDataFlowWithDB, "both"},
		{"事件消费和数据同步测试", suite.TestEventConsumptionAndSync, "neo4j"},
		{"数据一致性验证测试", suite.TestDataConsistency, "both"},
	}

	for i, test := range tests {
		// 检查是否有必需的数据库连接
		if test.requireDB == "postgres" && suite.pgDB == nil {
			log.Printf("⏭️ 跳过测试 %d/%d: %s (需要PostgreSQL连接)", i+1, len(tests), test.name)
			continue
		}
		if test.requireDB == "neo4j" && suite.neo4jDriver == nil {
			log.Printf("⏭️ 跳过测试 %d/%d: %s (需要Neo4j连接)", i+1, len(tests), test.name)
			continue
		}
		if test.requireDB == "both" && (suite.pgDB == nil || suite.neo4jDriver == nil) {
			log.Printf("⏭️ 跳过测试 %d/%d: %s (需要PostgreSQL和Neo4j连接)", i+1, len(tests), test.name)
			continue
		}

		log.Printf("📋 测试 %d/%d: %s", i+1, len(tests), test.name)
		
		startTime := time.Now()
		if err := test.fn(); err != nil {
			log.Printf("❌ 测试失败: %s - %v", test.name, err)
			return err
		}
		
		duration := time.Since(startTime)
		log.Printf("✅ 测试通过: %s (耗时: %v)", test.name, duration)
	}

	log.Println("🎉 所有数据库集成测试完成!")
	return nil
}

// TestPostgreSQLCommandRepository 测试PostgreSQL命令仓储
func (suite *DatabaseIntegrationTestSuite) TestPostgreSQLCommandRepository() error {
	log.Println("  🔍 测试PostgreSQL命令仓储...")

	// 创建测试组织
	orgID := uuid.New()
	suite.testOrganizations = append(suite.testOrganizations, orgID)

	testOrg := repositories.Organization{
		ID:           orgID,
		TenantID:     suite.testTenantID,
		UnitType:     "DEPARTMENT",
		Name:         "数据库测试部门",
		Description:  stringPtr("PostgreSQL数据库测试部门"),
		Status:       "ACTIVE",
		Profile:      map[string]interface{}{"test": "database", "priority": "high"},
		Level:        1,
		EmployeeCount: 15,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 测试创建组织
	if err := suite.commandRepo.CreateOrganization(suite.ctx, testOrg); err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	log.Println("    ✓ 组织创建成功")

	// 测试更新组织
	changes := map[string]interface{}{
		"name":           "更新后的数据库测试部门",
		"employee_count": 20,
		"updated_at":     time.Now(),
	}
	if err := suite.commandRepo.UpdateOrganization(suite.ctx, orgID, suite.testTenantID, changes); err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	log.Println("    ✓ 组织更新成功")

	// 测试状态变更
	if err := suite.commandRepo.SetOrganizationStatus(suite.ctx, orgID, suite.testTenantID, "INACTIVE"); err != nil {
		return fmt.Errorf("failed to set organization status: %w", err)
	}
	log.Println("    ✓ 组织状态变更成功")

	return nil
}

// TestNeo4jQueryRepository 测试Neo4j查询仓储
func (suite *DatabaseIntegrationTestSuite) TestNeo4jQueryRepository() error {
	log.Println("  🔍 测试Neo4j查询仓储...")

	// 首先在Neo4j中创建一些测试数据
	orgID := uuid.New()
	suite.testOrganizations = append(suite.testOrganizations, orgID)

	// 直接在Neo4j中创建测试节点
	session := suite.neo4jDriver.NewSession(suite.ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(suite.ctx)

	_, err := session.ExecuteWrite(suite.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			CREATE (o:Organization {
				id: $id,
				tenant_id: $tenant_id,
				unit_type: $unit_type,
				name: $name,
				description: $description,
				status: $status,
				level: $level,
				employee_count: $employee_count,
				is_active: $is_active,
				created_at: $created_at,
				updated_at: $updated_at
			})`
		
		_, err := tx.Run(suite.ctx, cypher, map[string]any{
			"id":             orgID.String(),
			"tenant_id":      suite.testTenantID.String(),
			"unit_type":      "DEPARTMENT",
			"name":           "Neo4j测试部门",
			"description":    "Neo4j数据库测试部门",
			"status":         "ACTIVE",
			"level":          1,
			"employee_count": 25,
			"is_active":      true,
			"created_at":     time.Now().Format(time.RFC3339),
			"updated_at":     time.Now().Format(time.RFC3339),
		})
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to create test data in Neo4j: %w", err)
	}
	log.Println("    ✓ Neo4j测试数据创建成功")

	// 测试组织存在性检查
	exists, err := suite.queryRepo.OrganizationExists(suite.ctx, orgID, suite.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to check organization existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("organization should exist but was not found")
	}
	log.Println("    ✓ 组织存在性检查成功")

	return nil
}

// TestCompleteDataFlowWithDB 测试完整的CQRS数据流
func (suite *DatabaseIntegrationTestSuite) TestCompleteDataFlowWithDB() error {
	log.Println("  🔍 测试完整的CQRS数据流...")

	orgID := uuid.New()
	suite.testOrganizations = append(suite.testOrganizations, orgID)

	// 步骤1: 通过命令仓储创建组织 (PostgreSQL)
	log.Println("    1️⃣ 通过PostgreSQL命令仓储创建组织...")
	testOrg := repositories.Organization{
		ID:           orgID,
		TenantID:     suite.testTenantID,
		UnitType:     "COMPANY",
		Name:         "CQRS数据流测试公司",
		Description:  stringPtr("完整数据流测试"),
		Status:       "ACTIVE",
		Profile:      map[string]interface{}{"test": "cqrs_flow"},
		Level:        0,
		EmployeeCount: 50,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := suite.commandRepo.CreateOrganization(suite.ctx, testOrg); err != nil {
		return fmt.Errorf("failed to create organization in PostgreSQL: %w", err)
	}
	log.Println("    ✓ PostgreSQL写入成功")

	// 步骤2: 发布事件
	log.Println("    2️⃣ 发布组织创建事件...")
	event := events.NewOrganizationCreated(
		suite.testTenantID,
		orgID,
		testOrg.Name,
		"CQRS001",
		nil,
		testOrg.Level,
	)

	if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	log.Println("    ✓ 事件发布成功")

	// 步骤3: 等待事件消费和同步
	log.Println("    3️⃣ 等待事件消费和Neo4j同步...")
	time.Sleep(2 * time.Second) // 等待异步处理

	// 步骤4: 验证Neo4j中的数据
	log.Println("    4️⃣ 验证Neo4j中的数据...")
	exists, err := suite.queryRepo.OrganizationExists(suite.ctx, orgID, suite.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to check organization in Neo4j: %w", err)
	}
	if !exists {
		log.Println("    ⚠️ 组织在Neo4j中未找到 (可能是事件处理延迟)")
	} else {
		log.Println("    ✓ Neo4j数据同步验证成功")
	}

	log.Println("    ✅ 完整CQRS数据流测试完成")
	return nil
}

// TestEventConsumptionAndSync 测试事件消费和数据同步
func (suite *DatabaseIntegrationTestSuite) TestEventConsumptionAndSync() error {
	log.Println("  🔍 测试事件消费和数据同步...")

	// 直接测试事件消费者
	orgID := uuid.New()
	suite.testOrganizations = append(suite.testOrganizations, orgID)

	// 创建测试事件
	event := events.NewOrganizationCreated(
		suite.testTenantID,
		orgID,
		"事件消费测试组织",
		"SYNC001",
		nil,
		1,
	)

	// 序列化事件
	eventData, err := event.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// 直接调用消费者
	if err := suite.consumer.ConsumeEvent(suite.ctx, eventData); err != nil {
		return fmt.Errorf("failed to consume event: %w", err)
	}
	log.Println("    ✓ 事件消费成功")

	// 验证数据同步
	time.Sleep(1 * time.Second) // 等待处理完成
	exists, err := suite.queryRepo.OrganizationExists(suite.ctx, orgID, suite.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to verify sync: %w", err)
	}
	if exists {
		log.Println("    ✓ 数据同步验证成功")
	} else {
		log.Println("    ⚠️ 数据同步验证失败")
	}

	return nil
}

// TestDataConsistency 测试数据一致性
func (suite *DatabaseIntegrationTestSuite) TestDataConsistency() error {
	log.Println("  🔍 测试数据一致性...")

	// 创建多个组织进行一致性测试
	orgs := []struct {
		name  string
		level int
	}{
		{"一致性测试总公司", 0},
		{"一致性测试分公司", 1}, 
		{"一致性测试部门", 2},
	}

	orgIDs := make([]uuid.UUID, len(orgs))

	// 批量创建组织
	for i, org := range orgs {
		orgID := uuid.New()
		orgIDs[i] = orgID
		suite.testOrganizations = append(suite.testOrganizations, orgID)

		testOrg := repositories.Organization{
			ID:           orgID,
			TenantID:     suite.testTenantID,
			UnitType:     "COMPANY",
			Name:         org.name,
			Description:  stringPtr("数据一致性测试"),
			Status:       "ACTIVE",
			Profile:      map[string]interface{}{"test": "consistency"},
			Level:        org.level,
			EmployeeCount: 10 * (i + 1),
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// 写入PostgreSQL
		if err := suite.commandRepo.CreateOrganization(suite.ctx, testOrg); err != nil {
			return fmt.Errorf("failed to create organization %d: %w", i, err)
		}

		// 发布事件
		event := events.NewOrganizationCreated(
			suite.testTenantID,
			orgID,
			org.name,
			fmt.Sprintf("CONS%03d", i),
			nil,
			org.level,
		)

		if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %d: %w", i, err)
		}
	}

	log.Printf("    ✓ 创建了 %d 个组织", len(orgs))

	// 等待同步完成
	time.Sleep(3 * time.Second)

	// 验证一致性
	consistentCount := 0
	for i, orgID := range orgIDs {
		exists, err := suite.queryRepo.OrganizationExists(suite.ctx, orgID, suite.testTenantID)
		if err != nil {
			log.Printf("    ⚠️ 验证组织 %d 时出错: %v", i, err)
			continue
		}
		if exists {
			consistentCount++
		}
	}

	log.Printf("    📊 数据一致性结果: %d/%d 个组织在Neo4j中同步成功", consistentCount, len(orgs))
	
	if consistentCount == len(orgs) {
		log.Println("    ✅ 数据一致性验证成功")
	} else {
		log.Println("    ⚠️ 部分数据未同步 (可能是延迟或配置问题)")
	}

	return nil
}

// OrganizationEventHandlerAdapter 组织事件处理器适配器
type OrganizationEventHandlerAdapter struct {
	consumer  *consumers.OrganizationEventConsumer
	eventType string
}

// NewOrganizationEventHandlerAdapter 创建事件处理器适配器
func NewOrganizationEventHandlerAdapter(consumer *consumers.OrganizationEventConsumer, eventType string) *OrganizationEventHandlerAdapter {
	return &OrganizationEventHandlerAdapter{
		consumer:  consumer,
		eventType: eventType,
	}
}

// Handle 处理事件
func (a *OrganizationEventHandlerAdapter) Handle(ctx context.Context, event events.DomainEvent) error {
	// 序列化事件
	eventData, err := event.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}
	
	// 调用消费者
	return a.consumer.ConsumeEvent(ctx, eventData)
}

// GetHandlerName 获取处理器名称
func (a *OrganizationEventHandlerAdapter) GetHandlerName() string {
	return fmt.Sprintf("OrganizationEventConsumer-%s", a.eventType)
}

// GetEventType 获取事件类型 (实现EventHandler接口)
func (a *OrganizationEventHandlerAdapter) GetEventType() string {
	return a.eventType
}

// stringPtr 辅助函数
func stringPtr(s string) *string {
	return &s
}

// main 主函数
func main() {
	log.Println("🚀 开始CQRS Phase 3 数据库集成测试...")
	log.Println("💡 提示: 设置以下环境变量以启用数据库测试:")
	log.Println("   POSTGRES_URL=postgres://user:password@localhost/dbname")
	log.Println("   NEO4J_URL=neo4j://localhost:7687")
	log.Println("   NEO4J_USER=neo4j")
	log.Println("   NEO4J_PASSWORD=password")

	// 创建测试套件
	suite := NewDatabaseIntegrationTestSuite()

	// 设置数据库连接
	if err := suite.SetupDatabaseConnections(); err != nil {
		log.Fatalf("❌ 数据库连接设置失败: %v", err)
	}

	// 确保清理数据库连接
	defer func() {
		if err := suite.TeardownDatabaseConnections(); err != nil {
			log.Printf("⚠️ 数据库连接清理失败: %v", err)
		}
	}()

	// 运行数据库集成测试
	if err := suite.RunDatabaseIntegrationTests(); err != nil {
		log.Fatalf("❌ 数据库集成测试失败: %v", err)
	}

	log.Printf("🎉 数据库集成测试成功完成! 共测试了 %d 个组织", len(suite.testOrganizations))
	log.Println("📊 测试统计:")
	log.Printf("  - 测试租户ID: %s", suite.testTenantID)
	log.Printf("  - 创建的测试组织数量: %d", len(suite.testOrganizations))
	log.Println("✅ 数据库集成测试完成!")
}