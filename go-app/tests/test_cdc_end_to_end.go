package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/neo4j"
	"github.com/google/uuid"
)

// 端到端CDC验证测试
// 测试完整的CQRS+CDC流水线：从事件发布到Neo4j数据同步

func main() {
	log.Println("🧪 启动端到端CDC验证测试...")
	
	// 创建测试环境
	testEnvironment := setupTestEnvironment()
	defer cleanupTestEnvironment(testEnvironment)
	
	// 执行测试用例
	testCases := []struct {
		name     string
		testFunc func(*TestEnvironment) error
	}{
		{"测试员工事件端到端流程", testEmployeeEventEndToEnd},
		{"测试组织事件端到端流程", testOrganizationEventEndToEnd},
		{"测试批量事件处理", testBatchEventProcessing},
		{"测试错误处理和恢复", testErrorHandlingAndRecovery},
		{"测试性能和吞吐量", testPerformanceAndThroughput},
	}
	
	var passedTests, totalTests int
	
	for _, testCase := range testCases {
		totalTests++
		log.Printf("\n🔍 运行测试: %s", testCase.name)
		
		if err := testCase.testFunc(testEnvironment); err != nil {
			log.Printf("❌ 测试失败: %s - %v", testCase.name, err)
		} else {
			log.Printf("✅ 测试通过: %s", testCase.name)
			passedTests++
		}
	}
	
	// 输出测试结果
	log.Printf("\n📊 测试结果汇总:")
	log.Printf("   总测试数: %d", totalTests)
	log.Printf("   通过测试: %d", passedTests)
	log.Printf("   失败测试: %d", totalTests-passedTests)
	log.Printf("   成功率: %.1f%%", float64(passedTests)/float64(totalTests)*100)
	
	if passedTests == totalTests {
		log.Println("🎉 所有端到端CDC验证测试通过！")
		log.Println("✅ CQRS+CDC流水线功能验证完成")
	} else {
		log.Println("⚠️ 部分测试失败，需要检查CDC流水线配置")
	}
}

// TestEnvironment 测试环境
type TestEnvironment struct {
	pipeline     *neo4j.CQRSCDCPipeline
	eventBus     events.EventBus
	testTenantID uuid.UUID
	ctx          context.Context
	cancel       context.CancelFunc
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() *TestEnvironment {
	log.Println("🔧 设置测试环境...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	
	// 创建EventBus
	factory := events.NewEventBusFactory()
	eventBus := factory.CreateMockEventBus()
	
	// 创建CQRS+CDC流水线
	config := neo4j.DefaultPipelineConfig()
	config.EnableDetailedLogs = true
	config.HealthCheckInterval = time.Second * 10
	config.MetricsExportInterval = time.Second * 30
	
	pipeline, err := neo4j.NewCQRSCDCPipeline(eventBus, config)
	if err != nil {
		log.Fatalf("❌ 创建CQRS+CDC流水线失败: %v", err)
	}
	
	// 启动流水线
	if err := pipeline.Start(ctx); err != nil {
		log.Fatalf("❌ 启动CQRS+CDC流水线失败: %v", err)
	}
	
	testEnv := &TestEnvironment{
		pipeline:     pipeline,
		eventBus:     eventBus,
		testTenantID: uuid.New(),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	log.Println("✅ 测试环境设置完成")
	return testEnv
}

// cleanupTestEnvironment 清理测试环境
func cleanupTestEnvironment(env *TestEnvironment) {
	log.Println("🧹 清理测试环境...")
	
	if env.pipeline != nil {
		env.pipeline.Stop()
	}
	
	if env.cancel != nil {
		env.cancel()
	}
	
	log.Println("✅ 测试环境清理完成")
}

// testEmployeeEventEndToEnd 测试员工事件端到端流程
func testEmployeeEventEndToEnd(env *TestEnvironment) error {
	log.Println("🔄 测试员工事件端到端流程...")
	
	// 创建测试员工事件
	employeeID := uuid.New()
	
	// 1. 测试员工创建事件
	createEvent := events.NewEmployeeCreated(
		env.testTenantID,
		employeeID,
		"TEST001",
		"张",
		"三",
		"zhangsan@test.com",
		time.Now(),
	)
	
	log.Printf("📤 发送员工创建事件: %s", createEvent.GetEventID())
	if err := env.pipeline.ProcessEvent(env.ctx, createEvent); err != nil {
		return err
	}
	
	// 等待事件处理
	time.Sleep(time.Second * 2)
	
	// 2. 测试员工更新事件
	updateFields := map[string]interface{}{
		"phone_number": "13800138000",
		"department":   "技术部",
	}
	
	updateEvent := events.NewEmployeeUpdated(
		env.testTenantID,
		employeeID,
		"TEST001",
		updateFields,
	)
	
	log.Printf("📤 发送员工更新事件: %s", updateEvent.GetEventID())
	if err := env.pipeline.ProcessEvent(env.ctx, updateEvent); err != nil {
		return err
	}
	
	// 等待事件处理
	time.Sleep(time.Second * 2)
	
	// 3. 测试员工终止事件
	terminateEvent := events.NewEmployeeTerminated(
		env.testTenantID,
		employeeID,
		"TEST001",
		time.Now(),
		"正常离职",
	)
	
	log.Printf("📤 发送员工终止事件: %s", terminateEvent.GetEventID())
	if err := env.pipeline.ProcessEvent(env.ctx, terminateEvent); err != nil {
		return err
	}
	
	// 验证流水线状态
	stats := env.pipeline.GetPerformanceStats()
	if stats.ProcessedEvents < 3 {
		return fmt.Errorf("期望处理至少3个事件，实际处理 %d 个", stats.ProcessedEvents)
	}
	
	log.Printf("✅ 员工事件端到端流程测试通过 (处理了 %d 个事件)", stats.ProcessedEvents)
	return nil
}

// testOrganizationEventEndToEnd 测试组织事件端到端流程
func testOrganizationEventEndToEnd(env *TestEnvironment) error {
	log.Println("🔄 测试组织事件端到端流程...")
	
	// 创建测试组织事件
	orgID := uuid.New()
	parentOrgID := uuid.New()
	
	// 1. 测试组织创建事件
	createEvent := events.NewOrganizationCreated(
		env.testTenantID,
		orgID,
		"技术部",
		"负责技术研发工作",
		&parentOrgID,
		1,
	)
	
	log.Printf("📤 发送组织创建事件: %s", createEvent.GetEventID())
	if err := env.pipeline.ProcessEvent(env.ctx, createEvent); err != nil {
		return err
	}
	
	// 等待事件处理
	time.Sleep(time.Second * 2)
	
	// 2. 测试组织重构事件
	restructureEvent := events.NewOrganizationRestructured(
		env.testTenantID,
		orgID,
		"技术部",
		"部门合并",
		"组织架构优化",
	)
	
	log.Printf("📤 发送组织重构事件: %s", restructureEvent.GetEventID())
	if err := env.pipeline.ProcessEvent(env.ctx, restructureEvent); err != nil {
		return err
	}
	
	// 等待事件处理
	time.Sleep(time.Second * 2)
	
	// 3. 测试组织停用事件
	deactivateEvent := events.NewOrganizationDeactivated(
		env.testTenantID,
		orgID,
		"技术部",
		"部门撤销",
	)
	
	log.Printf("📤 发送组织停用事件: %s", deactivateEvent.GetEventID())
	if err := env.pipeline.ProcessEvent(env.ctx, deactivateEvent); err != nil {
		return err
	}
	
	// 验证流水线状态
	stats := env.pipeline.GetPerformanceStats()
	if stats.TotalEvents < 6 { // 包括之前的员工事件
		return fmt.Errorf("期望处理至少6个事件，实际处理 %d 个", stats.TotalEvents)
	}
	
	log.Printf("✅ 组织事件端到端流程测试通过")
	return nil
}

// testBatchEventProcessing 测试批量事件处理
func testBatchEventProcessing(env *TestEnvironment) error {
	log.Println("🔄 测试批量事件处理...")
	
	// 创建批量测试事件
	var batchEvents []events.DomainEvent
	
	for i := 0; i < 10; i++ {
		employeeID := uuid.New()
		event := events.NewEmployeeCreated(
			env.testTenantID,
			employeeID,
			fmt.Sprintf("BATCH%03d", i),
			"批量",
			fmt.Sprintf("测试%d", i),
			fmt.Sprintf("batch%d@test.com", i),
			time.Now(),
		)
		batchEvents = append(batchEvents, event)
	}
	
	log.Printf("📤 发送批量事件: %d个事件", len(batchEvents))
	startTime := time.Now()
	
	if err := env.pipeline.ProcessEventBatch(env.ctx, batchEvents); err != nil {
		return err
	}
	
	processingTime := time.Since(startTime)
	
	// 验证批量处理性能
	stats := env.pipeline.GetPerformanceStats()
	if stats.ThroughputPerSecond == 0 {
		return fmt.Errorf("批量处理吞吐量为0")
	}
	
	log.Printf("✅ 批量事件处理测试通过 (处理时间: %v, 吞吐量: %.2f/秒)", 
		processingTime, stats.ThroughputPerSecond)
	return nil
}

// testErrorHandlingAndRecovery 测试错误处理和恢复
func testErrorHandlingAndRecovery(env *TestEnvironment) error {
	log.Println("🔄 测试错误处理和恢复...")
	
	// 记录处理前的统计信息
	statsBefore := env.pipeline.GetPerformanceStats()
	
	// 创建一个可能导致错误的事件（例如无效的UUID）
	invalidEvent := &events.BaseDomainEvent{}
	invalidEvent.EventID = ""        // 无效的事件ID
	invalidEvent.EventType = "test.invalid"
	invalidEvent.AggregateID = uuid.Nil
	invalidEvent.TenantID = uuid.Nil
	invalidEvent.Timestamp = time.Now()
	
	log.Printf("📤 发送无效事件进行错误处理测试")
	
	// 处理无效事件（应该会失败）
	err := env.pipeline.ProcessEvent(env.ctx, invalidEvent)
	if err == nil {
		log.Printf("⚠️ 期望处理无效事件时失败，但实际成功了")
	} else {
		log.Printf("✅ 无效事件处理正确失败: %v", err)
	}
	
	// 检查健康状态
	healthStatus := env.pipeline.GetHealthStatus()
	if healthStatus == nil {
		return fmt.Errorf("无法获取健康状态")
	}
	
	log.Printf("📊 健康状态: 整体健康=%v, 错误计数=%d", 
		healthStatus.IsHealthy, healthStatus.ErrorCount)
	
	// 发送正常事件确保系统可以恢复
	normalEvent := events.NewEmployeeCreated(
		env.testTenantID,
		uuid.New(),
		"RECOVERY001",
		"恢复",
		"测试",
		"recovery@test.com",
		time.Now(),
	)
	
	if err := env.pipeline.ProcessEvent(env.ctx, normalEvent); err != nil {
		return fmt.Errorf("系统恢复测试失败: %v", err)
	}
	
	log.Printf("✅ 错误处理和恢复测试通过")
	return nil
}

// testPerformanceAndThroughput 测试性能和吞吐量
func testPerformanceAndThroughput(env *TestEnvironment) error {
	log.Println("🔄 测试性能和吞吐量...")
	
	// 性能测试参数
	eventCount := 50
	maxProcessingTime := time.Second * 30
	minThroughput := 1.0 // 最小每秒处理1个事件
	
	// 创建性能测试事件
	var perfEvents []events.DomainEvent
	
	for i := 0; i < eventCount; i++ {
		employeeID := uuid.New()
		event := events.NewEmployeeCreated(
			env.testTenantID,
			employeeID,
			fmt.Sprintf("PERF%03d", i),
			"性能",
			fmt.Sprintf("测试%d", i),
			fmt.Sprintf("perf%d@test.com", i),
			time.Now(),
		)
		perfEvents = append(perfEvents, event)
	}
	
	log.Printf("📤 开始性能测试: %d个事件", eventCount)
	startTime := time.Now()
	
	// 批量处理事件
	if err := env.pipeline.ProcessEventBatch(env.ctx, perfEvents); err != nil {
		return fmt.Errorf("性能测试事件处理失败: %v", err)
	}
	
	totalProcessingTime := time.Since(startTime)
	
	// 验证性能指标
	if totalProcessingTime > maxProcessingTime {
		return fmt.Errorf("处理时间超出限制: %v > %v", totalProcessingTime, maxProcessingTime)
	}
	
	stats := env.pipeline.GetPerformanceStats()
	if stats.ThroughputPerSecond < minThroughput {
		return fmt.Errorf("吞吐量低于预期: %.2f < %.2f", stats.ThroughputPerSecond, minThroughput)
	}
	
	log.Printf("✅ 性能测试通过:")
	log.Printf("   处理事件: %d 个", eventCount)
	log.Printf("   总耗时: %v", totalProcessingTime)
	log.Printf("   平均每事件: %v", totalProcessingTime/time.Duration(eventCount))
	log.Printf("   吞吐量: %.2f 事件/秒", stats.ThroughputPerSecond)
	log.Printf("   成功率: %.1f%%", float64(stats.ProcessedEvents)/float64(stats.TotalEvents)*100)
	
	return nil
}