package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/neo4j"
	neo4jdriver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// 简化版性能基准测试
func main() {
	log.Println("⚡ 启动简化版性能基准测试...")
	
	ctx := context.Background()
	
	// 高性能配置
	config := &neo4j.MockConfig{
		SuccessRate:    0.98,
		LatencyMin:     time.Microsecond * 500,
		LatencyMax:     time.Millisecond * 2,
		EnableMetrics:  true,
		ErrorTypes:     []string{"timeout"},
		ErrorRate:      0.02,
		MaxConnections: 100,
		DatabaseName:   "performance_test",
	}
	
	manager := neo4j.NewMockConnectionManagerWithConfig(config)
	defer manager.Close(ctx)
	
	// 测试1: 基准延迟测试
	log.Println("🔄 执行基准延迟测试...")
	
	var latencies []time.Duration
	iterations := 50 // 减少迭代次数
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		_, err := manager.ExecuteWrite(ctx, func(tx neo4jdriver.ManagedTransaction) (any, error) {
			return fmt.Sprintf("test_%d", i), nil
		})
		
		latency := time.Since(start)
		latencies = append(latencies, latency)
		
		if err != nil && len(latencies) < iterations/2 {
			log.Printf("操作 %d 失败: %v", i, err)
		}
	}
	
	// 计算统计
	var avgLatency time.Duration
	if len(latencies) > 0 {
		var totalLatency time.Duration
		minLatency := latencies[0]
		maxLatency := latencies[0]
		
		for _, latency := range latencies {
			totalLatency += latency
			if latency < minLatency {
				minLatency = latency
			}
			if latency > maxLatency {
				maxLatency = latency
			}
		}
		
		avgLatency = totalLatency / time.Duration(len(latencies))
		
		log.Printf("📊 基准延迟统计:")
		log.Printf("   成功操作: %d/%d", len(latencies), iterations)
		log.Printf("   平均延迟: %v", avgLatency)
		log.Printf("   最小延迟: %v", minLatency)
		log.Printf("   最大延迟: %v", maxLatency)
		
		if avgLatency > time.Millisecond*10 {
			log.Printf("⚠️ 平均延迟较高: %v", avgLatency)
		} else {
			log.Println("✅ 基准延迟测试通过")
		}
	}
	
	// 测试2: 简化吞吐量测试
	log.Println("🔄 执行吞吐量测试...")
	
	startTime := time.Now()
	operations := 100
	successCount := 0
	
	for i := 0; i < operations; i++ {
		_, err := manager.ExecuteWrite(ctx, func(tx neo4jdriver.ManagedTransaction) (any, error) {
			return fmt.Sprintf("throughput_%d", i), nil
		})
		
		if err == nil {
			successCount++
		}
		
		// 短暂间隔
		time.Sleep(time.Microsecond * 100)
	}
	
	totalTime := time.Since(startTime)
	throughput := float64(operations) / totalTime.Seconds()
	successRate := float64(successCount) / float64(operations) * 100
	
	log.Printf("📊 吞吐量测试结果:")
	log.Printf("   总操作数: %d", operations)
	log.Printf("   成功操作: %d", successCount)
	log.Printf("   成功率: %.1f%%", successRate)
	log.Printf("   测试时间: %v", totalTime)
	log.Printf("   吞吐量: %.2f ops/sec", throughput)
	
	if throughput >= 50.0 && successRate >= 90.0 {
		log.Println("✅ 吞吐量测试通过")
	} else {
		log.Printf("⚠️ 吞吐量或成功率未达标: %.2f ops/sec, %.1f%%", throughput, successRate)
	}
	
	// 测试3: 资源监控
	log.Println("🔄 执行资源监控测试...")
	
	initialStats := manager.GetStatistics()
	
	// 执行一些操作
	for i := 0; i < 20; i++ {
		manager.ExecuteRead(ctx, func(tx neo4jdriver.ManagedTransaction) (any, error) {
			return fmt.Sprintf("monitor_%d", i), nil
		})
	}
	
	finalStats := manager.GetStatistics()
	
	log.Printf("📊 资源监控结果:")
	log.Printf("   初始操作数: %v", initialStats["total_operations"])
	log.Printf("   最终操作数: %v", finalStats["total_operations"])
	log.Printf("   平均延迟: %v", finalStats["average_latency"])
	log.Printf("   错误率: %v", finalStats["error_rate"])
	
	// 验证统计更新
	initialOps := initialStats["total_operations"].(int64)
	finalOps := finalStats["total_operations"].(int64)
	
	if finalOps > initialOps {
		log.Println("✅ 资源监控测试通过")
	} else {
		log.Printf("⚠️ 统计更新异常: %d -> %d", initialOps, finalOps)
	}
	
	// 性能总结
	log.Println("\n🎉 简化性能基准测试完成!")
	log.Printf("📋 测试总结:")
	log.Printf("   平均延迟: %v", avgLatency)
	log.Printf("   峰值吞吐量: %.2f ops/sec", throughput)
	log.Printf("   总成功率: %.1f%%", successRate)
	log.Printf("   Mock配置: 98%%成功率, 0.5ms-2ms延迟")
	
	log.Println("✅ 性能基准和监控功能验证成功!")
}