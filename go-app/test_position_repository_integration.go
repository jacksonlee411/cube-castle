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

	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// TestPositionRepositoryIntegration 职位仓储集成测试
type TestPositionRepositoryIntegration struct {
	db           *sqlx.DB
	positionRepo repositories.PositionCommandRepository
	outboxRepo   repositories.OutboxRepository
	testTenantID uuid.UUID
}

// SetupTest 设置测试环境
func SetupTest() (*TestPositionRepositoryIntegration, error) {
	// 连接数据库
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}

	db, err := sqlx.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建仓储
	outboxRepo := repositories.NewPostgresOutboxRepository(db)
	positionRepo := repositories.NewPostgresPositionRepository(db, outboxRepo)

	testTenantID := uuid.New()

	return &TestPositionRepositoryIntegration{
		db:           db,
		positionRepo: positionRepo,
		outboxRepo:   outboxRepo,
		testTenantID: testTenantID,
	}, nil
}

// Cleanup 清理测试环境
func (t *TestPositionRepositoryIntegration) Cleanup() {
	ctx := context.Background()

	// 清理测试数据
	queries := []string{
		"DELETE FROM assignment_history WHERE assignment_id IN (SELECT id FROM position_assignments WHERE tenant_id = $1)",
		"DELETE FROM assignment_details WHERE assignment_id IN (SELECT id FROM position_assignments WHERE tenant_id = $1)",
		"DELETE FROM position_assignments WHERE tenant_id = $1",
		"DELETE FROM positions WHERE tenant_id = $1",
		"DELETE FROM outbox_events WHERE tenant_id = $1",
	}

	for _, query := range queries {
		_, err := t.db.ExecContext(ctx, query, t.testTenantID)
		if err != nil {
			log.Printf("Cleanup warning: %v", err)
		}
	}

	if t.db != nil {
		t.db.Close()
	}
}

// TestCreatePosition 测试创建职位
func (t *TestPositionRepositoryIntegration) TestCreatePosition() error {
	log.Println("🧪 Testing Position Creation...")

	ctx := context.Background()
	position := repositories.Position{
		ID:            uuid.New(),
		TenantID:      t.testTenantID,
		PositionType:  "REGULAR",
		JobProfileID:  uuid.New(),
		DepartmentID:  uuid.New(),
		Status:        "ACTIVE",
		BudgetedFTE:   1.0,
		Details: map[string]interface{}{
			"title":       "Test Software Engineer",
			"description": "Test position for integration testing",
		},
	}

	// 创建职位
	err := t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}

	// 验证职位是否创建成功
	var count int
	err = t.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM positions WHERE id = $1 AND tenant_id = $2",
		position.ID, t.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to verify position creation: %w", err)
	}

	if count != 1 {
		return fmt.Errorf("expected 1 position, got %d", count)
	}

	log.Println("✅ Position creation test passed")
	return nil
}

// TestCreatePositionAssignment 测试创建职位分配
func (t *TestPositionRepositoryIntegration) TestCreatePositionAssignment() error {
	log.Println("🧪 Testing Position Assignment Creation...")

	ctx := context.Background()

	// 先创建一个职位
	position := repositories.Position{
		ID:           uuid.New(),
		TenantID:     t.testTenantID,
		PositionType: "REGULAR",
		JobProfileID: uuid.New(),
		DepartmentID: uuid.New(),
		Status:       "ACTIVE",
		BudgetedFTE:  1.0,
		Details: map[string]interface{}{
			"title": "Test Position for Assignment",
		},
	}

	err := t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to create position for assignment test: %w", err)
	}

	// 创建职位分配
	assignment := repositories.PositionAssignment{
		ID:             uuid.New(),
		TenantID:       t.testTenantID,
		PositionID:     position.ID,
		EmployeeID:     uuid.New(),
		StartDate:      time.Now(),
		IsCurrent:      true,
		FTE:            1.0,
		AssignmentType: "PRIMARY",
	}

	err = t.positionRepo.CreatePositionAssignment(ctx, assignment)
	if err != nil {
		return fmt.Errorf("failed to create position assignment: %w", err)
	}

	// 验证分配是否创建成功
	var count int
	err = t.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM position_assignments WHERE id = $1 AND tenant_id = $2",
		assignment.ID, t.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to verify assignment creation: %w", err)
	}

	if count != 1 {
		return fmt.Errorf("expected 1 assignment, got %d", count)
	}

	// 验证业务约束：同一员工只能有一个当前的主要职位分配
	var primaryCount int
	err = t.db.GetContext(ctx, &primaryCount,
		"SELECT COUNT(*) FROM position_assignments WHERE employee_id = $1 AND is_current = true AND assignment_type = 'PRIMARY'",
		assignment.EmployeeID)
	if err != nil {
		return fmt.Errorf("failed to verify primary assignment constraint: %w", err)
	}

	if primaryCount != 1 {
		return fmt.Errorf("expected 1 primary assignment, got %d", primaryCount)
	}

	log.Println("✅ Position assignment creation test passed")
	return nil
}

// TestOutboxEvents 测试Outbox事件
func (t *TestPositionRepositoryIntegration) TestOutboxEvents() error {
	log.Println("🧪 Testing Outbox Events...")

	ctx := context.Background()

	// 创建职位（应该生成Outbox事件）
	position := repositories.Position{
		ID:           uuid.New(),
		TenantID:     t.testTenantID,
		PositionType: "REGULAR",
		JobProfileID: uuid.New(),
		DepartmentID: uuid.New(),
		Status:       "ACTIVE",
		BudgetedFTE:  1.0,
		Details: map[string]interface{}{
			"title": "Test Position for Outbox",
		},
	}

	err := t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to create position for outbox test: %w", err)
	}

	// 验证Outbox事件是否创建
	var eventCount int
	err = t.db.GetContext(ctx, &eventCount,
		"SELECT COUNT(*) FROM outbox_events WHERE tenant_id = $1",
		t.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to verify outbox events: %w", err)
	}

	if eventCount < 1 {
		return fmt.Errorf("expected at least 1 outbox event, got %d", eventCount)
	}

	// 验证事件数据
	var event struct {
		EventType   string `db:"event_type"`
		AggregateID string `db:"aggregate_id"`
		Status      string `db:"status"`
	}

	err = t.db.GetContext(ctx, &event,
		"SELECT event_type, aggregate_id, status FROM outbox_events WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 1",
		t.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to get outbox event details: %w", err)
	}

	if event.Status != "PENDING" {
		return fmt.Errorf("expected event status 'PENDING', got '%s'", event.Status)
	}

	log.Printf("✅ Outbox events test passed (event_type: %s)", event.EventType)
	return nil
}

// TestPositionUpdates 测试职位更新
func (t *TestPositionRepositoryIntegration) TestPositionUpdates() error {
	log.Println("🧪 Testing Position Updates...")

	ctx := context.Background()

	// 创建职位
	position := repositories.Position{
		ID:           uuid.New(),
		TenantID:     t.testTenantID,
		PositionType: "REGULAR",
		JobProfileID: uuid.New(),
		DepartmentID: uuid.New(),
		Status:       "ACTIVE",
		BudgetedFTE:  1.0,
		Details: map[string]interface{}{
			"title":       "Original Title",
			"description": "Original Description",
		},
	}

	err := t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to create position for update test: %w", err)
	}

	// 更新职位
	updatedPosition := position
	updatedPosition.Status = "FROZEN"
	updatedPosition.Details = map[string]interface{}{
		"title":       "Updated Title",
		"description": "Updated Description",
		"reason":      "Position updated for testing",
	}

	err = t.positionRepo.UpdatePosition(ctx, updatedPosition)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	// 验证更新是否成功
	var dbPosition struct {
		Status  string                 `db:"status"`
		Details map[string]interface{} `db:"details"`
	}

	err = t.db.GetContext(ctx, &dbPosition,
		"SELECT status, details FROM positions WHERE id = $1",
		position.ID)
	if err != nil {
		return fmt.Errorf("failed to verify position update: %w", err)
	}

	if dbPosition.Status != "FROZEN" {
		return fmt.Errorf("expected updated status 'FROZEN', got '%s'", dbPosition.Status)
	}

	if title, ok := dbPosition.Details["title"].(string); !ok || title != "Updated Title" {
		return fmt.Errorf("expected updated title 'Updated Title', got '%v'", dbPosition.Details["title"])
	}

	log.Println("✅ Position updates test passed")
	return nil
}

// TestDatabaseConstraints 测试数据库约束
func (t *TestPositionRepositoryIntegration) TestDatabaseConstraints() error {
	log.Println("🧪 Testing Database Constraints...")

	ctx := context.Background()

	// 测试无效的职位状态
	position := repositories.Position{
		ID:           uuid.New(),
		TenantID:     t.testTenantID,
		PositionType: "REGULAR",
		JobProfileID: uuid.New(),
		DepartmentID: uuid.New(),
		Status:       "INVALID_STATUS", // 无效状态
		BudgetedFTE:  1.0,
	}

	err := t.positionRepo.CreatePosition(ctx, position)
	if err == nil {
		return fmt.Errorf("expected constraint violation for invalid status, but creation succeeded")
	}

	// 测试无效的FTE值
	position.Status = "ACTIVE"
	position.BudgetedFTE = 10.0 // 超出约束范围

	err = t.positionRepo.CreatePosition(ctx, position)
	if err == nil {
		return fmt.Errorf("expected constraint violation for invalid FTE, but creation succeeded")
	}

	// 测试有效的职位创建
	position.BudgetedFTE = 1.0
	err = t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("valid position creation failed: %w", err)
	}

	log.Println("✅ Database constraints test passed")
	return nil
}

// RunAllTests 运行所有测试
func (t *TestPositionRepositoryIntegration) RunAllTests() error {
	tests := []struct {
		name string
		test func() error
	}{
		{"Database Constraints", t.TestDatabaseConstraints},
		{"Position Creation", t.TestCreatePosition},
		{"Position Assignment", t.TestCreatePositionAssignment},
		{"Outbox Events", t.TestOutboxEvents},
		{"Position Updates", t.TestPositionUpdates},
	}

	log.Println("🚀 Starting Position Repository Integration Tests...")
	log.Printf("Test Tenant ID: %s", t.testTenantID)

	passed := 0
	failed := 0

	for _, test := range tests {
		log.Printf("\n--- Running Test: %s ---", test.name)

		if err := test.test(); err != nil {
			log.Printf("❌ Test '%s' FAILED: %v", test.name, err)
			failed++
		} else {
			log.Printf("✅ Test '%s' PASSED", test.name)
			passed++
		}
	}

	log.Printf("\n🏁 Integration Test Results:")
	log.Printf("✅ Passed: %d", passed)
	log.Printf("❌ Failed: %d", failed)
	log.Printf("📊 Total: %d", passed+failed)

	if failed > 0 {
		return fmt.Errorf("integration tests failed: %d/%d tests failed", failed, passed+failed)
	}

	log.Println("🎉 All Position Repository Integration Tests Passed!")
	return nil
}

// main 主函数
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🧪 Position Repository Integration Test Suite")

	// 设置测试环境
	test, err := SetupTest()
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}
	defer test.Cleanup()

	// 运行所有测试
	if err := test.RunAllTests(); err != nil {
		log.Fatalf("Integration tests failed: %v", err)
	}

	log.Println("✅ Position Repository Integration Test Suite Completed Successfully!")
}