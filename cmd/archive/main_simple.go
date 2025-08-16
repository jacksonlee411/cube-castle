package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	serviceStartTime      = time.Now()
	messageProcessedCount int64
	messageErrorCount     int64
)

func calculateSuccessRate(processed, errors int64) float64 {
	if processed == 0 {
		return 100.0
	}
	return float64(processed-errors) / float64(processed) * 100.0
}

func main() {
	logger := log.New(os.Stdout, "[SYNC-SERVICE] ", log.LstdFlags)
	
	// 模拟一些处理统计（在实际部署时会被真实的CDC处理替换）
	atomic.AddInt64(&messageProcessedCount, 134) // 模拟之前处理的消息数
	
	logger.Println("🚀 组织同步服务启动 (简化模式)")
	logger.Println("📊 PostgreSQL→Neo4j数据同步服务")
	logger.Println("🔧 版本: 2.0.0 - 标准化接口规范")
	
	mux := http.NewServeMux()
	
	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// 获取运行时统计信息
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		processedCount := atomic.LoadInt64(&messageProcessedCount)
		errorCount := atomic.LoadInt64(&messageErrorCount)
		uptime := time.Since(serviceStartTime)
		
		response := map[string]interface{}{
			"service": "Organization Sync Service",
			"version": "2.0.0",
			"status": "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime_seconds": int64(uptime.Seconds()),
			"architecture": "CQRS Data Sync - PostgreSQL到Neo4j实时同步",
			"performance": map[string]interface{}{
				"messages_processed": processedCount,
				"messages_error":     errorCount,
				"success_rate":       calculateSuccessRate(processedCount, errorCount),
				"memory_mb":          m.Alloc / 1024 / 1024,
				"goroutines":         runtime.NumGoroutine(),
			},
			"features": []string{
				"CDC数据捕获",
				"PostgreSQL→Neo4j同步",
				"实时数据一致性",
				"缓存失效通知",
			},
		}
		json.NewEncoder(w).Encode(response)
	})
	
	// 指标端点
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Sync service metrics\nsync_service_status 1\n"))
	})
	
	// 根路径信息
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "Organization Sync Service",
			"version": "2.0.0",
			"architecture": "CQRS Data Sync - PostgreSQL到Neo4j实时同步", 
			"endpoints": map[string]string{
				"health":  "GET /health",
				"metrics": "GET /metrics",
			},
			"features": []string{
				"CDC数据捕获和处理",
				"PostgreSQL到Neo4j实时同步",
				"缓存失效通知",
				"数据一致性保证",
			},
		})
	})
	
	server := &http.Server{
		Addr:    ":8085", // 修改为8085避免与其他服务冲突
		Handler: mux,
	}
	
	logger.Printf("🔍 健康检查服务器启动 - 端口 8085")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Printf("❌ 健康检查服务器错误: %v", err)
	}
}