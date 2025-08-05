package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// 简化版的模型和处理器
type OrganizationUnit struct {
	Code        string    `json:"code" db:"code"`
	ParentCode  *string   `json:"parent_code,omitempty" db:"parent_code"`
	Name        string    `json:"name" db:"name"`
	UnitType    string    `json:"unit_type" db:"unit_type"`
	Status      string    `json:"status" db:"status"`
	Level       int       `json:"level" db:"level"`
	Path        string    `json:"path" db:"path"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateOrganizationUnitRequest struct {
	Name        string  `json:"name"`
	ParentCode  *string `json:"parent_code,omitempty"`
	UnitType    string  `json:"unit_type"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

type ListOrganizationUnitsResponse struct {
	Organizations []OrganizationUnit `json:"organizations"`
	TotalCount    int64              `json:"total_count"`
	Page          int                `json:"page"`
	PageSize      int                `json:"page_size"`
}

// API处理器
type OrganizationHandler struct {
	db *sqlx.DB
}

func NewOrganizationHandler(db *sqlx.DB) *OrganizationHandler {
	return &OrganizationHandler{db: db}
}

func (h *OrganizationHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9" // 使用实际存在的租户ID
	}

	// 解析查询参数
	limit := 50
	offset := 0
	unitType := r.URL.Query().Get("unit_type")
	status := r.URL.Query().Get("status")

	// 构建查询
	query := `
		SELECT code, parent_code, name, unit_type, status, level, path, 
		       sort_order, description, created_at, updated_at
		FROM organization_units 
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIndex := 2

	if unitType != "" {
		query += fmt.Sprintf(" AND unit_type = $%d", argIndex)
		args = append(args, unitType)
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += " ORDER BY path, sort_order, code"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var units []OrganizationUnit
	err := h.db.Select(&units, query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取总数
	countQuery := `SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1`
	var totalCount int64
	err = h.db.Get(&totalCount, countQuery, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Count error: %v", err), http.StatusInternalServerError)
		return
	}

	response := ListOrganizationUnitsResponse{
		Organizations: units,
		TotalCount:    totalCount,
		Page:          1,
		PageSize:      limit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// 简单的JSON序列化
	fmt.Fprintf(w, `{
		"organizations": [`)
	
	for i, unit := range units {
		if i > 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `{
			"code": "%s",
			"name": "%s",
			"unit_type": "%s",
			"status": "%s",
			"level": %d,
			"path": "%s",
			"sort_order": %d,
			"created_at": "%s",
			"updated_at": "%s"`,
			unit.Code, unit.Name, unit.UnitType, unit.Status, 
			unit.Level, unit.Path, unit.SortOrder,
			unit.CreatedAt.Format(time.RFC3339),
			unit.UpdatedAt.Format(time.RFC3339))
		
		if unit.ParentCode != nil {
			fmt.Fprintf(w, `,"parent_code": "%s"`, *unit.ParentCode)
		}
		if unit.Description != nil {
			fmt.Fprintf(w, `,"description": "%s"`, *unit.Description)
		}
		fmt.Fprintf(w, "}")
	}
	
	fmt.Fprintf(w, `],
		"total_count": %d,
		"page": %d,
		"page_size": %d
	}`, totalCount, response.Page, response.PageSize)
}

func (h *OrganizationHandler) GetOrganizationByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	}

	// 验证编码格式
	if len(code) != 7 {
		http.Error(w, "Invalid code format: must be 7 digits", http.StatusBadRequest)
		return
	}

	var unit OrganizationUnit
	query := `
		SELECT code, parent_code, name, unit_type, status, level, path, 
		       sort_order, description, created_at, updated_at
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
	`
	
	err := h.db.Get(&unit, query, tenantID, code)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "Organization unit not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprintf(w, `{
		"code": "%s",
		"name": "%s",
		"unit_type": "%s",
		"status": "%s",
		"level": %d,
		"path": "%s",
		"sort_order": %d,
		"created_at": "%s",
		"updated_at": "%s"`,
		unit.Code, unit.Name, unit.UnitType, unit.Status,
		unit.Level, unit.Path, unit.SortOrder,
		unit.CreatedAt.Format(time.RFC3339),
		unit.UpdatedAt.Format(time.RFC3339))
	
	if unit.ParentCode != nil {
		fmt.Fprintf(w, `,"parent_code": "%s"`, *unit.ParentCode)
	}
	if unit.Description != nil {
		fmt.Fprintf(w, `,"description": "%s"`, *unit.Description)
	}
	fmt.Fprintf(w, "}")
}

func (h *OrganizationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	}

	// 获取总数
	var totalCount int64
	err := h.db.Get(&totalCount, "SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1", tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// 按类型统计
	typeStats := make(map[string]int64)
	rows, err := h.db.Query("SELECT unit_type, COUNT(*) FROM organization_units WHERE tenant_id = $1 GROUP BY unit_type", tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var unitType string
		var count int64
		rows.Scan(&unitType, &count)
		typeStats[unitType] = count
	}

	// 按状态统计
	statusStats := make(map[string]int64)
	rows, err = h.db.Query("SELECT status, COUNT(*) FROM organization_units WHERE tenant_id = $1 GROUP BY status", tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		rows.Scan(&status, &count)
		statusStats[status] = count
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	fmt.Fprintf(w, `{
		"total_count": %d,
		"by_type": {`, totalCount)
	
	first := true
	for unitType, count := range typeStats {
		if !first {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `"%s": %d`, unitType, count)
		first = false
	}
	
	fmt.Fprintf(w, `},
		"by_status": {`)
	
	first = true
	for status, count := range statusStats {
		if !first {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `"%s": %d`, status, count)
		first = false
	}
	
	fmt.Fprintf(w, `}
	}`)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
		"status": "healthy",
		"timestamp": "%s",
		"version": "v2.0.0",
		"service": "organization-units-api"
	}`, time.Now().Format(time.RFC3339))
}

func main() {
	// 数据库连接
	dbURL := "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connection established")

	// 创建处理器
	orgHandler := NewOrganizationHandler(db)

	// 设置路由
	r := chi.NewRouter()
	
	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	
	// CORS设置
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", HealthCheck)

	// API路由
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/organization-units", func(r chi.Router) {
			r.Get("/", orgHandler.GetOrganizations)
			r.Get("/stats", orgHandler.GetStats)
			r.Get("/{code}", orgHandler.GetOrganizationByCode)
		})
	})

	// 启动服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("🛑 Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("🚀 Organization Units API v2.0 starting on :8080")
	log.Printf("📊 Health check: http://localhost:8080/health")
	log.Printf("📋 API endpoint: http://localhost:8080/api/v1/organization-units")
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}