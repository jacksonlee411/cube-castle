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

// SimplePositionTest 简化的职位测试
type SimplePositionTest struct {
	db           *sqlx.DB
	positionRepo *repositories.PostgresPositionRepository
	outboxRepo   repositories.OutboxRepository
	testTenantID uuid.UUID
}

// Setup 设置测试环境  
func Setup() (*SimplePositionTest, error) {
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

	outboxRepo := repositories.NewPostgresOutboxRepository(db)
	positionRepo := repositories.NewPostgresPositionRepository(db, outboxRepo)

	return &SimplePositionTest{
		db:           db,
		positionRepo: positionRepo,
		outboxRepo:   outboxRepo,
		testTenantID: uuid.New(),
	}, nil
}

// Cleanup 清理测试数据
func (t *SimplePositionTest) Cleanup() {
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

	t.db.Close()
}

// TestBasicPositionCRUD 测试基本的职位CRUD操作
func (t *SimplePositionTest) TestBasicPositionCRUD() error {
	log.Println("🧪 Testing Basic Position CRUD...")

	ctx := context.Background()
	
	// 创建职位
	position := repositories.Position{
		ID:                 uuid.New(),
		TenantID:           t.testTenantID,
		PositionType:       "REGULAR",
		JobProfileID:       uuid.New(),
		DepartmentID:       uuid.New(),
		ManagerPositionID:  nil,
		Status:             "ACTIVE",
		BudgetedFTE:        1.0,
		Details: map[string]interface{}{
			"title":       "Test Software Engineer",
			"description": "Test position for CQRS integration",
			"level":       "Senior",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. 测试创建职位
	log.Println("  → Creating position...")
	err := t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}

	// 2. 验证职位是否创建成功
	log.Println("  → Verifying position creation...")
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
	
	// 3. 测试职位更新
	log.Println("  → Updating position...")
	updateFields := map[string]interface{}{
		"status": "FROZEN",
		"details": map[string]interface{}{
			"title":       "Updated Test Software Engineer",
			"description": "Updated test position",
			"level":       "Principal",
			"reason":      "Position updated via CQRS",
		},
	}
	
	err = t.positionRepo.UpdatePosition(ctx, position.ID, t.testTenantID, updateFields)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	// 4. 验证更新是否成功
	log.Println("  → Verifying position update...")
	var updatedStatus string
	err = t.db.GetContext(ctx, &updatedStatus,
		"SELECT status FROM positions WHERE id = $1", position.ID)
	if err != nil {
		return fmt.Errorf("failed to verify position update: %w", err)
	}

	if updatedStatus != "FROZEN" {
		return fmt.Errorf("expected updated status 'FROZEN', got '%s'", updatedStatus)
	}

	log.Println("✅ Basic Position CRUD test passed")
	return nil
}

// TestOutboxIntegration 测试Outbox模式集成
func (t *SimplePositionTest) TestOutboxIntegration() error {
	log.Println("🧪 Testing Outbox Pattern Integration...")

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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建职位
	log.Println("  → Creating position with Outbox events...")
	err := t.positionRepo.CreatePosition(ctx, position)
	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}

	// 验证Outbox事件是否创建
	log.Println("  → Verifying Outbox events...")
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

	// 检查事件详情
	var eventType string
	err = t.db.GetContext(ctx, &eventType,
		"SELECT event_type FROM outbox_events WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 1",
		t.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to get event type: %w", err)
	}

	log.Printf("✅ Outbox integration test passed (event_type: %s)", eventType)
	return nil
}

// TestSchemaConstraints 测试数据库约束
func (t *SimplePositionTest) TestSchemaConstraints() error {
	log.Println("🧪 Testing Database Schema Constraints...")

	ctx := context.Background()

	// 测试无效的职位状态约束
	log.Println("  → Testing invalid position status constraint...")
	_, err := t.db.ExecContext(ctx, `
		INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id, status, budgeted_fte, created_at, updated_at)
		VALUES ($1, $2, 'REGULAR', $3, $4, 'INVALID_STATUS', 1.0, NOW(), NOW())
	`, uuid.New(), t.testTenantID, uuid.New(), uuid.New())

	if err == nil {
		return fmt.Errorf("expected constraint violation for invalid status, but insert succeeded")
	}

	// 测试无效的FTE约束
	log.Println("  → Testing invalid FTE constraint...")
	_, err = t.db.ExecContext(ctx, `
		INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id, status, budgeted_fte, created_at, updated_at)
		VALUES ($1, $2, 'REGULAR', $3, $4, 'ACTIVE', 10.0, NOW(), NOW())
	`, uuid.New(), t.testTenantID, uuid.New(), uuid.New())

	if err == nil {
		return fmt.Errorf("expected constraint violation for invalid FTE, but insert succeeded")
	}

	// 测试有效的职位创建
	log.Println("  → Testing valid position creation...")
	validPositionID := uuid.New()
	_, err = t.db.ExecContext(ctx, `
		INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id, status, budgeted_fte, created_at, updated_at)
		VALUES ($1, $2, 'REGULAR', $3, $4, 'ACTIVE', 1.0, NOW(), NOW())
	`, validPositionID, t.testTenantID, uuid.New(), uuid.New())

	if err != nil {
		return fmt.Errorf("valid position creation failed: %w", err)
	}

	log.Println("✅ Schema constraints test passed")
	return nil
}

// TestPositionAssignmentTables 测试职位分配相关表
func (t *SimplePositionTest) TestPositionAssignmentTables() error {
	log.Println("🧪 Testing Position Assignment Tables...")

	ctx := context.Background()

	// 1. 先创建一个职位
	positionID := uuid.New()
	_, err := t.db.ExecContext(ctx, `
		INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id, status, budgeted_fte, created_at, updated_at)
		VALUES ($1, $2, 'REGULAR', $3, $4, 'ACTIVE', 1.0, NOW(), NOW())
	`, positionID, t.testTenantID, uuid.New(), uuid.New())
	if err != nil {
		return fmt.Errorf("failed to create position for assignment test: %w", err)
	}

	// 2. 测试创建职位分配
	log.Println("  → Creating position assignment...")
	assignmentID := uuid.New()
	employeeID := uuid.New()
	
	_, err = t.db.ExecContext(ctx, `
		INSERT INTO position_assignments (id, tenant_id, position_id, employee_id, start_date, is_current, fte, assignment_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, 1.0, 'PRIMARY', NOW(), NOW())
	`, assignmentID, t.testTenantID, positionID, employeeID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create position assignment: %w", err)
	}

	// 3. 测试分配详情
	log.Println("  → Creating assignment details...")
	detailsID := uuid.New()
	_, err = t.db.ExecContext(ctx, `
		INSERT INTO assignment_details (id, assignment_id, effective_date, reason, approval_status, created_at, updated_at)
		VALUES ($1, $2, $3, 'Test assignment details', 'APPROVED', NOW(), NOW())
	`, detailsID, assignmentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create assignment details: %w", err)
	}

	// 4. 测试分配历史
	log.Println("  → Creating assignment history...")
	historyID := uuid.New()
	_, err = t.db.ExecContext(ctx, `
		INSERT INTO assignment_history (id, assignment_id, change_type, new_values, changed_by, effective_date, created_at)
		VALUES ($1, $2, 'CREATED', '{}', $3, $4, NOW())
	`, historyID, assignmentID, uuid.New(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create assignment history: %w", err)
	}

	// 5. 验证数据完整性
	log.Println("  → Verifying assignment data integrity...")
	var assignmentCount, detailsCount, historyCount int
	
	err = t.db.GetContext(ctx, &assignmentCount,
		"SELECT COUNT(*) FROM position_assignments WHERE tenant_id = $1", t.testTenantID)
	if err != nil {
		return fmt.Errorf("failed to count assignments: %w", err)
	}
	
	err = t.db.GetContext(ctx, &detailsCount,
		"SELECT COUNT(*) FROM assignment_details WHERE assignment_id = $1", assignmentID)
	if err != nil {
		return fmt.Errorf("failed to count assignment details: %w", err)
	}
	
	err = t.db.GetContext(ctx, &historyCount,
		"SELECT COUNT(*) FROM assignment_history WHERE assignment_id = $1", assignmentID)
	if err != nil {
		return fmt.Errorf("failed to count assignment history: %w", err)
	}

	if assignmentCount != 1 || detailsCount != 1 || historyCount != 1 {
		return fmt.Errorf("expected 1 record in each table, got assignments: %d, details: %d, history: %d",
			assignmentCount, detailsCount, historyCount)
	}

	log.Println("✅ Position assignment tables test passed")
	return nil
}

// RunAllTests 运行所有测试
func (t *SimplePositionTest) RunAllTests() error {
	tests := []struct {
		name string
		test func() error
	}{
		{"Schema Constraints", t.TestSchemaConstraints},
		{"Basic Position CRUD", t.TestBasicPositionCRUD},
		{"Outbox Integration", t.TestOutboxIntegration},
		{"Position Assignment Tables", t.TestPositionAssignmentTables},
	}

	log.Println("🚀 Starting Position CQRS Schema Integration Tests...")
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

	log.Println("🎉 All Position CQRS Integration Tests Passed!")
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🧪 Position CQRS Schema Integration Test Suite")

	test, err := Setup()
	if err != nil {
		log.Fatalf("Failed to setup test: %v", err)
	}
	defer test.Cleanup()

	if err := test.RunAllTests(); err != nil {
		log.Fatalf("Tests failed: %v", err)
	}

	log.Println("✅ Position CQRS Schema Integration Test Suite Completed Successfully!")
}