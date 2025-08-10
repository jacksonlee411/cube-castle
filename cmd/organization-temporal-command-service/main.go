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
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"cube-castle-deployment-test/pkg/monitoring"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 扩展的时态业务实体 =====

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
	
	// 新增时态字段
	EffectiveDate     *time.Time `json:"effective_date,omitempty" db:"effective_date"`
	EndDate           *time.Time `json:"end_date,omitempty" db:"end_date"`
	Version           *int       `json:"version,omitempty" db:"version"`
	SupersedesVersion *int       `json:"supersedes_version,omitempty" db:"supersedes_version"`
	ChangeReason      *string    `json:"change_reason,omitempty" db:"change_reason"`
	IsCurrent         *bool      `json:"is_current,omitempty" db:"is_current"`
}

// 时态查询选项
type TemporalQueryOptions struct {
	AsOfDate        *time.Time `json:"as_of_date,omitempty"`        // 时间点查询
	EffectiveFrom   *time.Time `json:"effective_from,omitempty"`    // 生效起始时间
	EffectiveTo     *time.Time `json:"effective_to,omitempty"`      // 生效结束时间  
	IncludeHistory  bool       `json:"include_history,omitempty"`   // 包含历史版本
	IncludeFuture   bool       `json:"include_future,omitempty"`    // 包含未来版本
	IncludeDissolved bool      `json:"include_dissolved,omitempty"` // 包含已解散组织
	Version         *int       `json:"version,omitempty"`           // 特定版本
	MaxVersions     int        `json:"max_versions,omitempty"`      // 最大版本数量
}

// 组织变更事件请求
type OrganizationChangeEvent struct {
	EventType     string                 `json:"event_type"`      // CREATE, UPDATE, RESTRUCTURE, DISSOLVE
	EffectiveDate time.Time              `json:"effective_date"`  // 生效日期
	EndDate       *time.Time             `json:"end_date,omitempty"` // 结束日期(特殊场景)
	ChangeData    map[string]interface{} `json:"change_data"`     // 变更内容
	ChangeReason  string                 `json:"change_reason"`   // 变更原因
}

// 组织事件实体
type OrganizationEvent struct {
	EventID           string     `json:"event_id" db:"event_id"`
	OrganizationCode  string     `json:"organization_code" db:"organization_code"`
	EventType         string     `json:"event_type" db:"event_type"`
	EventData         []byte     `json:"event_data" db:"event_data"`
	EffectiveDate     time.Time  `json:"effective_date" db:"effective_date"`
	EndDate           *time.Time `json:"end_date" db:"end_date"`
	CreatedBy         string     `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	TenantID          string     `json:"tenant_id" db:"tenant_id"`
}

// 时间线操作请求
type TimelineOperationRequest struct {
	Operation     string                 `json:"operation"`               // CORRECT, CANCEL, VOID
	TargetDate    time.Time              `json:"target_date"`             // 操作目标日期
	TargetVersion *int                   `json:"target_version,omitempty"` // 目标版本
	NewData       map[string]interface{} `json:"new_data,omitempty"`      // 校正数据
	Reason        string                 `json:"reason"`                  // 操作原因
}

// ===== 时态仓储层 =====

type TemporalOrganizationRepository struct {
	db *sql.DB
}

func NewTemporalOrganizationRepository(db *sql.DB) *TemporalOrganizationRepository {
	return &TemporalOrganizationRepository{db: db}
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
	
	// 解析最大版本数
	if maxVersionsStr := r.URL.Query().Get("max_versions"); maxVersionsStr != "" {
		if maxVersions, err := strconv.Atoi(maxVersionsStr); err == nil {
			opts.MaxVersions = maxVersions
		}
	}
	
	return opts, nil
}

// 时态查询实现
func (r *TemporalOrganizationRepository) GetByCodeTemporal(ctx context.Context, tenantID uuid.UUID, code string, opts *TemporalQueryOptions) ([]*Organization, error) {
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
	
	// 时间点查询：查询在指定日期有效的版本
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
	
	rows, err := r.db.QueryContext(ctx, query, args...)
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

// 创建组织事件
func (r *TemporalOrganizationRepository) CreateOrganizationEvent(ctx context.Context, tx *sql.Tx, event *OrganizationEvent) (string, error) {
	var eventID string
	query := `
		INSERT INTO organization_events (
			organization_code, event_type, event_data, effective_date, 
			end_date, created_by, tenant_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING event_id
	`
	
	err := tx.QueryRowContext(ctx, query,
		event.OrganizationCode, event.EventType, event.EventData,
		event.EffectiveDate, event.EndDate, event.CreatedBy, event.TenantID,
	).Scan(&eventID)
	
	if err != nil {
		return "", fmt.Errorf("创建组织事件失败: %w", err)
	}
	
	return eventID, nil
}

// 创建组织版本历史记录
func (r *TemporalOrganizationRepository) CreateOrganizationVersion(ctx context.Context, tx *sql.Tx, org *Organization) error {
	// 序列化组织数据为JSON
	orgData, err := json.Marshal(org)
	if err != nil {
		return fmt.Errorf("序列化组织数据失败: %w", err)
	}
	
	query := `
		INSERT INTO organization_versions (
			organization_code, version, effective_date, end_date,
			snapshot_data, change_reason, tenant_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err = tx.ExecContext(ctx, query,
		org.Code, *org.Version, *org.EffectiveDate, org.EndDate,
		orgData, *org.ChangeReason, org.TenantID,
	)
	
	if err != nil {
		return fmt.Errorf("创建组织版本历史失败: %w", err)
	}
	
	return nil
}

// 获取组织的下一个版本号
func (r *TemporalOrganizationRepository) GetNextVersion(ctx context.Context, tx *sql.Tx, code string) (int, error) {
	var maxVersion int
	query := `SELECT COALESCE(MAX(version), 0) + 1 FROM organization_units WHERE code = $1`
	
	err := tx.QueryRowContext(ctx, query, code).Scan(&maxVersion)
	if err != nil {
		return 0, fmt.Errorf("获取下一个版本号失败: %w", err)
	}
	
	return maxVersion, nil
}

// ===== HTTP处理器 =====

type TemporalOrganizationHandler struct {
	repo *TemporalOrganizationRepository
	db   *sql.DB
}

func NewTemporalOrganizationHandler(db *sql.DB) *TemporalOrganizationHandler {
	return &TemporalOrganizationHandler{
		repo: NewTemporalOrganizationRepository(db),
		db:   db,
	}
}

func (h *TemporalOrganizationHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantHeader := r.Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		if tenantID, err := uuid.Parse(tenantHeader); err == nil {
			return tenantID
		}
	}
	return DefaultTenantID
}

func (h *TemporalOrganizationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details error) {
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

// 时态查询处理器
func (h *TemporalOrganizationHandler) GetOrganizationTemporal(w http.ResponseWriter, r *http.Request) {
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
	organizations, err := h.repo.GetByCodeTemporal(r.Context(), tenantID, code, opts)
	if err != nil {
		monitoring.RecordOrganizationOperation("temporal_get", "failed", "command-service")
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
	
	monitoring.RecordOrganizationOperation("temporal_get", "success", "command-service")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 事件驱动变更处理器
func (h *TemporalOrganizationHandler) CreateOrganizationEvent(w http.ResponseWriter, r *http.Request) {
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
	
	// 1. 记录事件
	eventData, _ := json.Marshal(req.ChangeData)
	eventID, err := h.repo.CreateOrganizationEvent(r.Context(), tx, &OrganizationEvent{
		OrganizationCode: code,
		EventType:        req.EventType,
		EventData:        eventData,
		EffectiveDate:    req.EffectiveDate,
		EndDate:          req.EndDate,
		CreatedBy:        "system", // 从认证上下文获取
		TenantID:         tenantID.String(),
	})
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "EVENT_CREATE_ERROR", "创建事件失败", err)
		return
	}
	
	// 2. 处理不同类型的事件
	switch req.EventType {
	case "UPDATE", "RESTRUCTURE":
		err = h.handleUpdateEvent(r.Context(), tx, tenantID, code, &req)
	case "DISSOLVE":
		err = h.handleDissolveEvent(r.Context(), tx, tenantID, code, &req)
	case "ACTIVATE", "DEACTIVATE":
		err = h.handleStatusEvent(r.Context(), tx, tenantID, code, &req)
	default:
		err = fmt.Errorf("未支持的事件类型: %s", req.EventType)
	}
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "EVENT_PROCESS_ERROR", "处理事件失败", err)
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
	
	monitoring.RecordOrganizationOperation("event_create", "success", "command-service")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// 处理更新事件
func (h *TemporalOrganizationHandler) handleUpdateEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *OrganizationChangeEvent) error {
	// 获取当前版本
	currentOrg, err := h.getCurrentVersion(ctx, tx, tenantID, code)
	if err != nil {
		return fmt.Errorf("获取当前版本失败: %w", err)
	}
	
	// 创建新版本
	newVersion, err := h.repo.GetNextVersion(ctx, tx, code)
	if err != nil {
		return fmt.Errorf("获取新版本号失败: %w", err)
	}
	
	// 应用变更数据
	updatedOrg := *currentOrg
	updatedOrg.Version = &newVersion
	updatedOrg.EffectiveDate = &req.EffectiveDate
	updatedOrg.EndDate = req.EndDate
	updatedOrg.ChangeReason = &req.ChangeReason
	updatedOrg.SupersedesVersion = currentOrg.Version
	
	// 应用具体的字段变更
	for field, value := range req.ChangeData {
		switch field {
		case "name":
			if name, ok := value.(string); ok {
				updatedOrg.Name = name
			}
		case "unit_type":
			if unitType, ok := value.(string); ok {
				updatedOrg.UnitType = unitType
			}
		case "status":
			if status, ok := value.(string); ok {
				updatedOrg.Status = status
			}
		case "description":
			if desc, ok := value.(string); ok {
				updatedOrg.Description = desc
			}
		}
	}
	
	// 创建版本历史记录
	if err := h.repo.CreateOrganizationVersion(ctx, tx, &updatedOrg); err != nil {
		return fmt.Errorf("创建版本历史记录失败: %w", err)
	}
	
	return nil
}

// 处理解散事件
func (h *TemporalOrganizationHandler) handleDissolveEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *OrganizationChangeEvent) error {
	endDate := req.EndDate
	if endDate == nil {
		// 默认使用生效日期作为结束日期
		endDate = &req.EffectiveDate
	}
	
	// 更新当前版本的结束日期和状态
	_, err := tx.ExecContext(ctx,
		"UPDATE organization_units SET end_date = $1, status = 'INACTIVE', is_current = false WHERE code = $2 AND tenant_id = $3 AND is_current = true",
		*endDate, code, tenantID.String())
		
	return err
}

// 处理状态变更事件
func (h *TemporalOrganizationHandler) handleStatusEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *OrganizationChangeEvent) error {
	var newStatus string
	switch req.EventType {
	case "ACTIVATE":
		newStatus = "ACTIVE"
	case "DEACTIVATE":
		newStatus = "INACTIVE"
	}
	
	// 直接更新当前版本的状态
	_, err := tx.ExecContext(ctx,
		"UPDATE organization_units SET status = $1, updated_at = NOW() WHERE code = $2 AND tenant_id = $3 AND is_current = true",
		newStatus, code, tenantID.String())
		
	return err
}

// 获取当前版本
func (h *TemporalOrganizationHandler) getCurrentVersion(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*Organization, error) {
	query := `
		SELECT tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, version, supersedes_version, change_reason, is_current
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND is_current = true
	`
	
	org := &Organization{}
	err := tx.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.Version, &org.SupersedesVersion,
		&org.ChangeReason, &org.IsCurrent,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("未找到组织 %s 的当前版本", code)
	} else if err != nil {
		return nil, fmt.Errorf("查询当前版本失败: %w", err)
	}
	
	return org, nil
}

// ===== 主程序 =====

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
	handler := NewTemporalOrganizationHandler(db)
	
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
	
	// 监控指标
	r.Handle("/metrics", promhttp.Handler())
	
	// API路由
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		// 时态查询端点
		r.Get("/{code}/temporal", handler.GetOrganizationTemporal)
		
		// 事件驱动变更端点
		r.Post("/{code}/events", handler.CreateOrganizationEvent)
		
		// 时态查询端点的查询字符串版本
		r.Get("/{code}", handler.GetOrganizationTemporal) // 支持时态查询参数
	})
	
	// 启动服务器
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
		log.Printf("🚀 时态组织命令服务启动在端口 %s", port)
		log.Println("📋 支持的功能:")
		log.Println("  - 时态查询 (as_of_date, effective_from, effective_to)")
		log.Println("  - 事件驱动变更 (UPDATE, RESTRUCTURE, DISSOLVE)")
		log.Println("  - 版本历史管理")
		log.Println("  - 时间线一致性保证")
		
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