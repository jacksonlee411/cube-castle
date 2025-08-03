package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/events/eventbus"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// PerformanceBenchmarkSuite 性能基准测试套件
type PerformanceBenchmarkSuite struct {
	ctx       context.Context
	tenantID  uuid.UUID
	logger    Logger
	eventBus  events.EventBus
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
	// 在性能测试中减少日志输出
}

func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	log.Printf("ERROR: %s %v", msg, fields)
}

func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	// 在性能测试中减少日志输出
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Name          string
	Operations    int
	Duration      time.Duration
	OpsPerSecond  float64
	AvgLatency    time.Duration
	Memory        uint64
	Allocations   uint64
	GoroutinesBefore int
	GoroutinesAfter  int
}

// NewPerformanceBenchmarkSuite 创建性能基准测试套件
func NewPerformanceBenchmarkSuite() *PerformanceBenchmarkSuite {
	return &PerformanceBenchmarkSuite{
		ctx:      context.Background(),
		tenantID: uuid.New(),
		logger:   &SimpleLogger{},
	}
}

// Setup 设置测试环境
func (suite *PerformanceBenchmarkSuite) Setup() error {
	log.Println("🔧 设置性能基准测试环境...")

	// 设置事件总线
	suite.eventBus = eventbus.NewInMemoryEventBus(suite.logger)
	if err := suite.eventBus.Start(suite.ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	log.Println("✅ 性能基准测试环境设置完成")
	return nil
}

// Teardown 清理测试环境
func (suite *PerformanceBenchmarkSuite) Teardown() error {
	log.Println("🧹 清理性能基准测试环境...")

	if suite.eventBus != nil {
		if err := suite.eventBus.Stop(); err != nil {
			suite.logger.Warn("Failed to stop event bus", "error", err)
		}
	}

	log.Println("✅ 性能基准测试环境清理完成")
	return nil
}

// RunAllBenchmarks 运行所有性能基准测试
func (suite *PerformanceBenchmarkSuite) RunAllBenchmarks() error {
	log.Println("🚀 开始性能基准测试...")

	benchmarks := []struct {
		name string
		fn   func() (*BenchmarkResult, error)
	}{
		{"事件创建性能", suite.BenchmarkEventCreation},
		{"事件序列化性能", suite.BenchmarkEventSerialization},
		{"事件发布性能", suite.BenchmarkEventPublishing},
		{"组织数据结构性能", suite.BenchmarkOrganizationDataStructure},
		{"并发事件处理性能", suite.BenchmarkConcurrentEventHandling},
		{"内存使用和GC性能", suite.BenchmarkMemoryAndGC},
		{"高负载压力测试", suite.BenchmarkHighLoadStress},
	}

	var allResults []*BenchmarkResult

	for i, benchmark := range benchmarks {
		log.Printf("📊 基准测试 %d/%d: %s", i+1, len(benchmarks), benchmark.name)
		
		// 强制GC以获得更准确的内存测量
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
		
		result, err := benchmark.fn()
		if err != nil {
			log.Printf("❌ 基准测试失败: %s - %v", benchmark.name, err)
			return err
		}
		
		allResults = append(allResults, result)
		suite.printBenchmarkResult(result)
	}

	// 生成综合报告
	suite.generateSummaryReport(allResults)

	log.Println("🎉 所有性能基准测试完成!")
	return nil
}

// BenchmarkEventCreation 事件创建性能基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkEventCreation() (*BenchmarkResult, error) {
	const operations = 10000
	
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	
	for i := 0; i < operations; i++ {
		orgID := uuid.New()
		event := events.NewOrganizationCreated(
			suite.tenantID,
			orgID,
			fmt.Sprintf("基准测试组织-%d", i),
			fmt.Sprintf("BENCH%05d", i),
			nil,
			1,
		)
		if event == nil {
			return nil, fmt.Errorf("failed to create event %d", i)
		}
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "事件创建性能",
		Operations:    operations,
		Duration:      duration,
		OpsPerSecond:  float64(operations) / duration.Seconds(),
		AvgLatency:    duration / operations,
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// BenchmarkEventSerialization 事件序列化性能基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkEventSerialization() (*BenchmarkResult, error) {
	const operations = 5000
	
	// 预创建事件
	event := events.NewOrganizationCreated(
		suite.tenantID,
		uuid.New(),
		"序列化基准测试组织",
		"SERIAL001",
		nil,
		1,
	)
	
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	
	for i := 0; i < operations; i++ {
		_, err := event.Serialize()
		if err != nil {
			return nil, fmt.Errorf("serialization failed at iteration %d: %w", i, err)
		}
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "事件序列化性能",
		Operations:    operations,
		Duration:      duration,
		OpsPerSecond:  float64(operations) / duration.Seconds(),
		AvgLatency:    duration / operations,
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// BenchmarkEventPublishing 事件发布性能基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkEventPublishing() (*BenchmarkResult, error) {
	const operations = 1000
	
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	
	for i := 0; i < operations; i++ {
		orgID := uuid.New()
		event := events.NewOrganizationCreated(
			suite.tenantID,
			orgID,
			fmt.Sprintf("发布基准测试组织-%d", i),
			fmt.Sprintf("PUB%05d", i),
			nil,
			1,
		)
		
		if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
			return nil, fmt.Errorf("publish failed at iteration %d: %w", i, err)
		}
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "事件发布性能",
		Operations:    operations,
		Duration:      duration,
		OpsPerSecond:  float64(operations) / duration.Seconds(),
		AvgLatency:    duration / operations,
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// BenchmarkOrganizationDataStructure 组织数据结构性能基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkOrganizationDataStructure() (*BenchmarkResult, error) {
	const operations = 5000
	
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	
	for i := 0; i < operations; i++ {
		org := repositories.Organization{
			ID:           uuid.New(),
			TenantID:     suite.tenantID,
			UnitType:     "DEPARTMENT",
			Name:         fmt.Sprintf("数据结构基准测试组织-%d", i),
			Description:  stringPtr("复杂数据结构性能测试"),
			Status:       "ACTIVE",
			Profile: map[string]interface{}{
				"department":    "技术部",
				"location":      "北京",
				"employees":     100 + i,
				"budget":        float64(1000000 + i*1000),
				"established":   "2020-01-01",
				"tags":          []string{"technology", "innovation", "growth"},
				"settings": map[string]interface{}{
					"public":     true,
					"recruiting": i%2 == 0,
					"priority":   i % 5,
				},
				"metadata": map[string]interface{}{
					"created_by":  "system",
					"version":     "1.0",
					"checksum":    fmt.Sprintf("hash-%d", i),
				},
			},
			Level:         i % 5,
			EmployeeCount: 100 + i,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		
		// 序列化测试复杂度
		_, err := json.Marshal(org)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal organization %d: %w", i, err)
		}
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "组织数据结构性能",
		Operations:    operations,
		Duration:      duration,
		OpsPerSecond:  float64(operations) / duration.Seconds(),
		AvgLatency:    duration / operations,
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// BenchmarkConcurrentEventHandling 并发事件处理性能基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkConcurrentEventHandling() (*BenchmarkResult, error) {
	const concurrency = 20
	const operationsPerGoroutine = 100
	const totalOperations = concurrency * operationsPerGoroutine
	
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	
	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				orgID := uuid.New()
				event := events.NewOrganizationCreated(
					suite.tenantID,
					orgID,
					fmt.Sprintf("并发基准测试组织-%d-%d", workerID, j),
					fmt.Sprintf("CONC%03d%03d", workerID, j),
					nil,
					1,
				)
				
				if err := suite.eventBus.Publish(suite.ctx, event); err != nil {
					errChan <- fmt.Errorf("worker %d operation %d failed: %w", workerID, j, err)
					return
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errChan)
	
	// 检查错误
	if len(errChan) > 0 {
		return nil, <-errChan
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "并发事件处理性能",
		Operations:    totalOperations,
		Duration:      duration,
		OpsPerSecond:  float64(totalOperations) / duration.Seconds(),
		AvgLatency:    duration / totalOperations,
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// BenchmarkMemoryAndGC 内存使用和GC性能基准测试
func (suite *PerformanceBenchmarkSuite) BenchmarkMemoryAndGC() (*BenchmarkResult, error) {
	const operations = 2000
	
	var memBefore, memAfter runtime.MemStats
	runtime.GC() // 强制GC
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	
	// 创建大量对象以测试GC性能
	var eventList []events.DomainEvent
	
	for i := 0; i < operations; i++ {
		orgID := uuid.New()
		event := events.NewOrganizationCreated(
			suite.tenantID,
			orgID,
			fmt.Sprintf("GC基准测试组织-%d", i),
			fmt.Sprintf("GC%05d", i),
			nil,
			1,
		)
		
		eventList = append(eventList, event)
		
		// 定期发布和清理以测试GC
		if i%100 == 0 {
			for _, e := range eventList {
				suite.eventBus.Publish(suite.ctx, e)
			}
			eventList = eventList[:0] // 清空切片但保留容量
			
			if i%500 == 0 {
				runtime.GC() // 触发GC
			}
		}
	}
	
	// 发布剩余事件
	for _, e := range eventList {
		suite.eventBus.Publish(suite.ctx, e)
	}
	
	runtime.GC() // 最终GC
	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "内存使用和GC性能",
		Operations:    operations,
		Duration:      duration,
		OpsPerSecond:  float64(operations) / duration.Seconds(),
		AvgLatency:    duration / operations,
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// BenchmarkHighLoadStress 高负载压力测试
func (suite *PerformanceBenchmarkSuite) BenchmarkHighLoadStress() (*BenchmarkResult, error) {
	const duration = 5 * time.Second
	const concurrency = 50
	
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	goroutinesBefore := runtime.NumGoroutine()
	
	start := time.Now()
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	var totalOps int64
	var mu sync.Mutex
	
	// 启动多个goroutine进行高负载测试
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			localOps := 0
			
			for {
				select {
				case <-stopChan:
					mu.Lock()
					totalOps += int64(localOps)
					mu.Unlock()
					return
				default:
					orgID := uuid.New()
					event := events.NewOrganizationCreated(
						suite.tenantID,
						orgID,
						fmt.Sprintf("压力测试组织-%d-%d", workerID, localOps),
						fmt.Sprintf("STRESS%02d%05d", workerID, localOps),
						nil,
						1,
					)
					
					suite.eventBus.Publish(suite.ctx, event)
					localOps++
				}
			}
		}(i)
	}
	
	// 运行指定时间后停止
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()
	
	actualDuration := time.Since(start)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	return &BenchmarkResult{
		Name:          "高负载压力测试",
		Operations:    int(totalOps),
		Duration:      actualDuration,
		OpsPerSecond:  float64(totalOps) / actualDuration.Seconds(),
		AvgLatency:    actualDuration / time.Duration(totalOps),
		Memory:        memAfter.Alloc - memBefore.Alloc,
		Allocations:   memAfter.Mallocs - memBefore.Mallocs,
		GoroutinesBefore: goroutinesBefore,
		GoroutinesAfter:  goroutinesAfter,
	}, nil
}

// printBenchmarkResult 打印基准测试结果
func (suite *PerformanceBenchmarkSuite) printBenchmarkResult(result *BenchmarkResult) {
	log.Printf("  📊 %s:", result.Name)
	log.Printf("    操作数量: %d", result.Operations)
	log.Printf("    总耗时: %v", result.Duration)
	log.Printf("    每秒操作数: %.2f ops/sec", result.OpsPerSecond)
	log.Printf("    平均延迟: %v", result.AvgLatency)
	log.Printf("    内存使用: %s", formatBytes(result.Memory))
	log.Printf("    分配次数: %d", result.Allocations)
	log.Printf("    Goroutine数量: %d → %d", result.GoroutinesBefore, result.GoroutinesAfter)
}

// generateSummaryReport 生成综合报告
func (suite *PerformanceBenchmarkSuite) generateSummaryReport(results []*BenchmarkResult) {
	log.Println("\n📈 性能基准测试综合报告:")
	log.Println("================================================================")
	
	var totalOps int
	var totalDuration time.Duration
	var totalMemory uint64
	var totalAllocations uint64
	
	log.Printf("%-25s %10s %15s %12s %12s", "测试名称", "操作数", "每秒操作数", "平均延迟", "内存使用")
	log.Println("----------------------------------------------------------------")
	
	for _, result := range results {
		totalOps += result.Operations
		totalDuration += result.Duration
		totalMemory += result.Memory
		totalAllocations += result.Allocations
		
		log.Printf("%-25s %10d %12.2f/s %12v %12s", 
			result.Name, 
			result.Operations, 
			result.OpsPerSecond, 
			result.AvgLatency,
			formatBytes(result.Memory))
	}
	
	log.Println("----------------------------------------------------------------")
	log.Printf("%-25s %10d %12.2f/s %12v %12s", 
		"总计", 
		totalOps, 
		float64(totalOps)/totalDuration.Seconds(), 
		totalDuration/time.Duration(len(results)),
		formatBytes(totalMemory))
	
	log.Println("\n🎯 性能评估:")
	
	// 性能评估
	for _, result := range results {
		var performance string
		switch {
		case result.OpsPerSecond > 10000:
			performance = "🟢 优秀"
		case result.OpsPerSecond > 5000:
			performance = "🟡 良好"
		case result.OpsPerSecond > 1000:
			performance = "🟠 一般"
		default:
			performance = "🔴 需要优化"
		}
		log.Printf("  %s: %s (%.0f ops/sec)", result.Name, performance, result.OpsPerSecond)
	}
	
	log.Println("\n📊 系统资源使用:")
	log.Printf("  总内存分配: %s", formatBytes(totalMemory))
	log.Printf("  总分配次数: %d", totalAllocations)
	log.Printf("  平均每次分配: %s", formatBytes(totalMemory/uint64(max(int(totalAllocations), 1))))
}

// formatBytes 格式化字节数
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// max 返回两个整数的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// stringPtr 辅助函数
func stringPtr(s string) *string {
	return &s
}

// main 主函数
func main() {
	log.Println("🚀 开始CQRS Phase 3 性能基准测试...")
	
	// 创建性能基准测试套件
	suite := NewPerformanceBenchmarkSuite()
	
	// 设置测试环境
	if err := suite.Setup(); err != nil {
		log.Fatalf("❌ 性能基准测试环境设置失败: %v", err)
	}
	
	// 确保清理测试环境
	defer func() {
		if err := suite.Teardown(); err != nil {
			log.Printf("⚠️ 性能基准测试环境清理失败: %v", err)
		}
	}()
	
	// 显示系统信息
	log.Printf("💻 系统信息:")
	log.Printf("  Go版本: %s", runtime.Version())
	log.Printf("  操作系统: %s", runtime.GOOS)
	log.Printf("  架构: %s", runtime.GOARCH)
	log.Printf("  CPU核心数: %d", runtime.NumCPU())
	log.Printf("  初始Goroutine数: %d\n", runtime.NumGoroutine())
	
	// 运行性能基准测试
	if err := suite.RunAllBenchmarks(); err != nil {
		log.Fatalf("❌ 性能基准测试失败: %v", err)
	}
	
	log.Println("🎉 性能基准测试成功完成!")
}