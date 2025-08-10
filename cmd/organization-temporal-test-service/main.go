package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// 时态组织结构
type Organization struct {
	TenantID          string     `json:"tenant_id" db:"tenant_id"`
	Code              string     `json:"code" db:"code"`
	ParentCode        *string    `json:"parent_code,omitempty" db:"parent_code"`
	Name              string     `json:"name" db:"name"`
	UnitType          string     `json:"unit_type" db:"unit_type"`
	Status            string     `json:"status" db:"status"`
	Level             int        `json:"level" db:"level"`
	Path              string     `json:"path" db:"path"`
	SortOrder         int        `json:"sort_order" db:"sort_order"`
	Description       string     `json:"description" db:"description"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	
	// 时态字段
	EffectiveDate     *time.Time `json:"effective_date,omitempty" db:"effective_date"`
	EndDate           *time.Time `json:"end_date,omitempty" db:"end_date"`
	Version           *int       `json:"version,omitempty" db:"version"`
	SupersedesVersion *int       `json:"supersedes_version,omitempty" db:"supersedes_version"`
	ChangeReason      *string    `json:"change_reason,omitempty" db:"change_reason"`
	IsCurrent         *bool      `json:"is_current,omitempty" db:"is_current"`
}

// 时态查询选项
type TemporalQueryOptions struct {
	AsOfDate        *time.Time `json:"as_of_date,omitempty"`
	EffectiveFrom   *time.Time `json:"effective_from,omitempty"`
	EffectiveTo     *time.Time `json:"effective_to,omitempty"`
	IncludeHistory  bool       `json:"include_history,omitempty"`
	IncludeFuture   bool       `json:"include_future,omitempty"`
	IncludeDissolved bool      `json:"include_dissolved,omitempty"`
	Version         *int       `json:"version,omitempty"`
	MaxVersions     int        `json:"max_versions,omitempty"`
}

// 组织变更事件请求
type OrganizationChangeEvent struct {
	EventType     string                 `json:"event_type"`
	EffectiveDate time.Time              `json:"effective_date"`
	EndDate       *time.Time             `json:"end_date,omitempty"`
	ChangeData    map[string]interface{} `json:"change_data"`
	ChangeReason  string                 `json:"change_reason"`
}

// 处理器
type TemporalHandler struct {
	db *sql.DB
}

func NewTemporalHandler(db *sql.DB) *TemporalHandler {
	return &TemporalHandler{db: db}
}

func (h *TemporalHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantHeader := r.Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		if tenantID, err := uuid.Parse(tenantHeader); err == nil {
			return tenantID
		}
	}
	return DefaultTenantID
}

func (h *TemporalHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error_code": errorCode,
		"message":    message,
	}
	
	if details != nil {
		response["details"] = details.Error()
	}
	
	json.NewEncoder(w).Encode(response)
}

// HTTP查询参数解析
func ParseTemporalQuery(r *http.Request) (*TemporalQueryOptions, error) {
	opts := &TemporalQueryOptions{}
	
	// 解析as_of_date参数
	if asOfStr := r.URL.Query().Get("as_of_date"); asOfStr != "" {
		if asOfDate, err := time.Parse("2006-01-02", asOfStr); err == nil {
			opts.AsOfDate = &asOfDate
		} else {
			return nil, fmt.Errorf("无效的as_of_date格式，期望：YYYY-MM-DD")
		}
	}
	
	// 解析日期范围
	if fromStr := r.URL.Query().Get("effective_from"); fromStr != "" {
		if from, err := time.Parse("2006-01-02", fromStr); err == nil {
			opts.EffectiveFrom = &from
		}
	}
	
	if toStr := r.URL.Query().Get("effective_to"); toStr != "" {
		if to, err := time.Parse("2006-01-02", toStr); err == nil {
			opts.EffectiveTo = &to
		}
	}
	
	// 解析布尔参数
	opts.IncludeHistory = r.URL.Query().Get("include_history") == "true"
	opts.IncludeFuture = r.URL.Query().Get("include_future") == "true" 
	opts.IncludeDissolved = r.URL.Query().Get("include_dissolved") == "true"
	
	// 解析版本参数
	if versionStr := r.URL.Query().Get("version"); versionStr != "" {
		if version, err := strconv.Atoi(versionStr); err == nil {
			opts.Version = &version
		}
	}
	
	if maxVersionsStr := r.URL.Query().Get("max_versions"); maxVersionsStr != "" {
		if maxVersions, err := strconv.Atoi(maxVersionsStr); err == nil {
			opts.MaxVersions = maxVersions
		}
	}
	
	return opts, nil
}

// 时态查询实现
func (h *TemporalHandler) GetByCodeTemporal(ctx context.Context, tenantID uuid.UUID, code string, opts *TemporalQueryOptions) ([]*Organization, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	// 基础条件
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, tenantID.String())
	argIndex++
	
	conditions = append(conditions, fmt.Sprintf("code = $%d", argIndex))
	args = append(args, code)
	argIndex++
	
	// 时间点查询
	if opts.AsOfDate != nil {
		conditions = append(conditions, fmt.Sprintf(
			"effective_date <= $%d AND (end_date IS NULL OR end_date >= $%d)", 
			argIndex, argIndex))
		args = append(args, *opts.AsOfDate)
		argIndex++
	}
	
	// 日期范围查询
	if opts.EffectiveFrom != nil {
		conditions = append(conditions, fmt.Sprintf("effective_date >= $%d", argIndex))
		args = append(args, *opts.EffectiveFrom)
		argIndex++
	}
	
	if opts.EffectiveTo != nil {
		conditions = append(conditions, fmt.Sprintf("effective_date <= $%d", argIndex))
		args = append(args, *opts.EffectiveTo)
		argIndex++
	}
	
	// 特定版本查询
	if opts.Version != nil {
		conditions = append(conditions, fmt.Sprintf("version = $%d", argIndex))
		args = append(args, *opts.Version)
		argIndex++
	}
	
	// 当前版本过滤
	if !opts.IncludeHistory && opts.AsOfDate == nil && opts.Version == nil {
		conditions = append(conditions, "is_current = true")
	}
	
	// 未来版本过滤
	if !opts.IncludeFuture {
		conditions = append(conditions, "effective_date <= CURRENT_DATE")
	}
	
	// 已解散组织过滤
	if !opts.IncludeDissolved {
		conditions = append(conditions, "(end_date IS NULL OR end_date > CURRENT_DATE)")
	}
	
	// 构建查询
	query := fmt.Sprintf(`
		SELECT tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, version, supersedes_version, change_reason, is_current
		FROM organization_units 
		WHERE %s
		ORDER BY version DESC
		%s
	`, strings.Join(conditions, " AND "), 
	   func() string {
		   if opts.MaxVersions > 0 {
			   return fmt.Sprintf("LIMIT %d", opts.MaxVersions)
		   }
		   return ""
	   }())
	
	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("时态查询失败: %w", err)
	}
	defer rows.Close()
	
	var organizations []*Organization
	for rows.Next() {
		org := &Organization{}
		err := rows.Scan(
			&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
			&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
			&org.Description, &org.CreatedAt, &org.UpdatedAt,
			&org.EffectiveDate, &org.EndDate, &org.Version, &org.SupersedesVersion,
			&org.ChangeReason, &org.IsCurrent,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描时态查询结果失败: %w", err)
		}
		organizations = append(organizations, org)
	}
	
	return organizations, nil
}

// 时态查询处理器
func (h *TemporalHandler) GetOrganizationTemporal(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}
	
	// 解析时态查询参数
	opts, err := ParseTemporalQuery(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_TEMPORAL_PARAMS", "时态查询参数无效", err)
		return
	}
	
	tenantID := h.getTenantID(r)
	
	// 执行时态查询
	organizations, err := h.GetByCodeTemporal(r.Context(), tenantID, code, opts)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "TEMPORAL_QUERY_ERROR", "时态查询失败", err)
		return
	}
	
	if len(organizations) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "未找到匹配的组织版本", nil)
		return
	}
	
	// 构建响应
	response := map[string]interface{}{
		"organizations": organizations,
		"query_options": opts,
		"result_count":  len(organizations),
		"queried_at":    time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 创建组织事件处理器
func (h *TemporalHandler) CreateOrganizationEvent(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	
	var req OrganizationChangeEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	
	// 验证事件类型
	validEventTypes := map[string]bool{
		"CREATE": true, "UPDATE": true, "RESTRUCTURE": true, "DISSOLVE": true,
		"ACTIVATE": true, "DEACTIVATE": true,
	}
	if !validEventTypes[req.EventType] {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_EVENT_TYPE", "无效的事件类型", nil)
		return
	}
	
	tenantID := h.getTenantID(r)
	
	// 开始事务
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "TRANSACTION_ERROR", "开始事务失败", err)
		return
	}
	defer tx.Rollback()
	
	// 记录事件到events表
	eventData, _ := json.Marshal(req.ChangeData)
	var eventID string
	err = tx.QueryRowContext(r.Context(), `
		INSERT INTO organization_events (
			organization_code, event_type, event_data, effective_date, 
			end_date, created_by, tenant_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING event_id`,
		code, req.EventType, eventData,
		req.EffectiveDate, req.EndDate, "system", tenantID.String(),
	).Scan(&eventID)
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "EVENT_CREATE_ERROR", "创建事件失败", err)
		return
	}
	
	// 提交事务
	if err := tx.Commit(); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "COMMIT_ERROR", "提交事务失败", err)
		return
	}
	
	response := map[string]interface{}{
		"event_id":       eventID,
		"event_type":     req.EventType,
		"organization":   code,
		"effective_date": req.EffectiveDate,
		"status":         "processed",
		"processed_at":   time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()
	
	if err = db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}
	
	log.Println("✅ 数据库连接成功")
	
	// 创建处理器
	handler := NewTemporalHandler(db)
	
	// 设置路由
	r := chi.NewRouter()
	
	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	
	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"service": "organization-temporal-command-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"features": []string{"temporal-queries", "event-driven-changes", "timeline-management"},
		})
	})
	
	// API路由
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		// 时态查询端点
		r.Get("/{code}/temporal", handler.GetOrganizationTemporal)
		r.Get("/{code}", handler.GetOrganizationTemporal) // 支持时态查询参数
		
		// 事件驱动变更端点
		r.Post("/{code}/events", handler.CreateOrganizationEvent)
	})
	
	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "9091"
	}
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	
	// 优雅关闭
	go func() {
		log.Printf("🚀 时态组织命令服务启动在端口 %s", port)
		log.Println("📋 支持的功能:")
		log.Println("  - 时态查询 (as_of_date, effective_from, effective_to)")
		log.Println("  - 事件驱动变更 (UPDATE, RESTRUCTURE, DISSOLVE)")
		log.Println("  - 版本历史管理")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败:", err)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("正在关闭服务器...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}
	
	log.Println("服务器已关闭")
}