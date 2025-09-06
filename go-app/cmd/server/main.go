// Cube Castle 组织API服务器主程序
// 集成Prometheus监控指标采集

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	
	"cube-castle/internal/api/handlers"
	"cube-castle/internal/api/middleware"
	"cube-castle/internal/metrics"
	"cube-castle/internal/service"
)

func main() {
	// 初始化Gin路由器
	router := gin.New()

	// 添加基础中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加Prometheus监控中间件
	router.Use(metrics.PrometheusMiddleware())

	// 注册自定义业务指标
	metrics.RegisterCustomMetrics()

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "organization-api",
			"version":   "v4.2.1",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Prometheus指标端点
	router.GET("/metrics", metrics.Handler())

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 组织单元管理路由
		orgRoutes := v1.Group("/organization-units")
		{
			orgRoutes.POST("", handlers.CreateOrganization)
			orgRoutes.GET("/:code", handlers.GetOrganization)
			orgRoutes.PUT("/:code", handlers.UpdateOrganization)
			orgRoutes.PATCH("/:code", handlers.PatchOrganization)
			orgRoutes.DELETE("/:code", handlers.DeleteOrganization)
			
			// 核心业务操作 (ADR-008统一端点)
			orgRoutes.POST("/:code/activate", handlers.ActivateOrganization)
			orgRoutes.POST("/:code/suspend", handlers.SuspendOrganization)
		}
	}

	// ADR-008: 弃用端点处理中间件
	router.Use(middleware.DeprecatedEndpointGuard())

	// 启动服务器
	server := &http.Server{
		Addr:    ":9090",
		Handler: router,
	}

	// 优雅关闭处理
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Println("🚀 Organization API Server started on :9090")
	log.Println("📊 Metrics endpoint available at http://localhost:9090/metrics")
	log.Println("❤️  Health check available at http://localhost:9090/health")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// 5秒优雅关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}