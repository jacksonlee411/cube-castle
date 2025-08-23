package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cube-castle/cmd/organization-command-service-simplified/internal/handlers"
	"github.com/cube-castle/cmd/organization-command-service-simplified/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
)

func main() {
	logger := log.New(os.Stdout, "[COMMAND-SERVICE] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 启动组织命令服务...")

	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 验证数据库连接
	if err := db.Ping(); err != nil {
		logger.Fatalf("数据库连接验证失败: %v", err)
	}

	logger.Println("✅ 数据库连接成功")

	// 初始化仓储和处理器
	orgRepo := repository.NewOrganizationRepository(db, logger)
	orgHandler := handlers.NewOrganizationHandler(orgRepo, logger)

	// 设置路由
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS设置
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "service": "organization-command-service", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// 设置组织相关路由
	orgHandler.SetupRoutes(r)

	// 服务启动
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		logger.Printf("🎯 组织命令服务启动在端口 %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("🛑 正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("服务关闭错误: %v", err)
	} else {
		logger.Println("✅ 服务已安全关闭")
	}
}
