package main

import (
	"context"
	"crypto/md5"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 简化的时态业务实体（移除版本字段） =====

type Organization struct {
	RecordID    string    `json:"record_id" db:"record_id"` // UUID唯一标识符
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Code        string    `json:"code" db:"code"`
	ParentCode  *string   `json:"parent_code,omitempty" db:"parent_code"`
	Name        string    `json:"name" db:"name"`
	UnitType    string    `json:"unit_type" db:"unit_type"`
	Status      string    `json:"status" db:"status"`
	Level       int       `json:"level" db:"level"`
	Path        string    `json:"path" db:"path"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 时态字段（符合行业标准）
	EffectiveDate *time.Time `json:"effective_date,omitempty" db:"effective_date"`
	EndDate       *time.Time `json:"end_date,omitempty" db:"end_date"`
	ChangeReason  *string    `json:"change_reason,omitempty" db:"change_reason"`
	IsCurrent     *bool      `json:"is_current,omitempty" db:"is_current"`
}

// 时态查询选项（移除版本相关参数）
type TemporalQueryOptions struct {
	AsOfDate         *time.Time `json:"as_of_date,omitempty"`        // 时间点查询
	EffectiveDate    *time.Time `json:"effective_date,omitempty"`    // 生效日期过滤
	EndDate          *time.Time `json:"end_date,omitempty"`          // 结束日期过滤
	IncludeHistory   bool       `json:"include_history,omitempty"`   // 包含历史版本
	IncludeFuture    bool       `json:"include_future,omitempty"`    // 包含未来版本
	IncludeDissolved bool       `json:"include_dissolved,omitempty"` // 包含已解散组织
	MaxRecords       int        `json:"max_records,omitempty"`       // 最大记录数量
}

// 组织变更事件请求
type OrganizationChangeEvent struct {
	EventType     string                 `json:"event_type"`         // CREATE, UPDATE, RESTRUCTURE, DISSOLVE
	EffectiveDate time.Time              `json:"effective_date"`     // 生效日期
	EndDate       *time.Time             `json:"end_date,omitempty"` // 结束日期(特殊场景)
	ChangeData    map[string]interface{} `json:"change_data"`        // 变更内容
	ChangeReason  string                 `json:"change_reason"`      // 变更原因
}

// 组织事件实体
type OrganizationEvent struct {
	EventID          string     `json:"event_id" db:"event_id"`
	OrganizationCode string     `json:"organization_code" db:"organization_code"`
	EventType        string     `json:"event_type" db:"event_type"`
	EventData        []byte     `json:"event_data" db:"event_data"`
	EffectiveDate    time.Time  `json:"effective_date" db:"effective_date"`
	EndDate          *time.Time `json:"end_date" db:"end_date"`
	CreatedBy        string     `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	TenantID         string     `json:"tenant_id" db:"tenant_id"`
}

// ===== 时态仓储层 =====

type TemporalOrganizationRepository struct {
	db *sql.DB
}

func NewTemporalOrganizationRepository(db *sql.DB) *TemporalOrganizationRepository {
	return &TemporalOrganizationRepository{db: db}
}

// HTTP查询参数解析（移除版本参数）
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

	// 解析effective_date和end_date
	if effectiveDateStr := r.URL.Query().Get("effective_date"); effectiveDateStr != "" {
		if effectiveDate, err := time.Parse("2006-01-02", effectiveDateStr); err == nil {
			opts.EffectiveDate = &effectiveDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			opts.EndDate = &endDate
		}
	}

	// 解析布尔参数
	opts.IncludeHistory = r.URL.Query().Get("include_history") == "true"
	opts.IncludeFuture = r.URL.Query().Get("include_future") == "true"
	opts.IncludeDissolved = r.URL.Query().Get("include_dissolved") == "true"

	return opts, nil
}

// 时态查询实现（基于纯日期模型）
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

	// 时间点查询：查询在指定日期有效的记录，优化NULL值处理
	if opts.AsOfDate != nil {
		conditions = append(conditions, fmt.Sprintf(
			"COALESCE(effective_date, CURRENT_TIMESTAMP) <= $%d AND (end_date IS NULL OR end_date >= $%d)",
			argIndex, argIndex))
		args = append(args, *opts.AsOfDate)
		argIndex++
	}

	// 日期范围查询，优化NULL值处理
	if opts.EffectiveDate != nil {
		conditions = append(conditions, fmt.Sprintf("COALESCE(effective_date, CURRENT_TIMESTAMP) >= $%d", argIndex))
		args = append(args, *opts.EffectiveDate)
		argIndex++
	}

	if opts.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("COALESCE(end_date, '9999-12-31'::timestamp) <= $%d", argIndex))
		args = append(args, *opts.EndDate)
		argIndex++
	}

	// 当前记录过滤 - 如果既没有时间点查询，也没有明确包含历史，则只返回当前记录
	if !opts.IncludeHistory && opts.AsOfDate == nil {
		conditions = append(conditions, "is_current = true")
	}

	// 未来记录过滤 - 只在明确不包含未来记录时过滤，但不影响当前记录
	if !opts.IncludeFuture && opts.AsOfDate == nil && opts.IncludeHistory {
		conditions = append(conditions, "COALESCE(effective_date, CURRENT_TIMESTAMP) <= CURRENT_TIMESTAMP")
	}

	// 已解散组织过滤 - 当包含历史记录时，不应该过滤已解散组织
	if !opts.IncludeDissolved && !opts.IncludeHistory && opts.AsOfDate == nil {
		conditions = append(conditions, "(end_date IS NULL OR end_date > CURRENT_DATE)")
	}

	// 特殊处理：当明确要求包含历史记录时，确保不过滤任何历史记录
	if opts.IncludeHistory {
		// 如果包含历史记录，则移除可能的已解散组织过滤条件
		// 不添加任何关于end_date的过滤条件
	}

	// 构建查询（按日期排序）- 使用COALESCE处理NULL值，优化扫描性能
	query := fmt.Sprintf(`
		SELECT record_id, tenant_id, code, 
		       COALESCE(parent_code, '') as parent_code,
		       name, unit_type, status, level, path, sort_order,
		       COALESCE(description, '') as description,
		       created_at, updated_at,
		       COALESCE(effective_date, CURRENT_TIMESTAMP) as effective_date,
		       end_date,
		       COALESCE(change_reason, '') as change_reason,
		       COALESCE(is_current, false) as is_current
		FROM organization_units 
		WHERE %s
		ORDER BY COALESCE(effective_date, CURRENT_TIMESTAMP) DESC
		%s
	`, strings.Join(conditions, " AND "),
		func() string {
			if opts.MaxRecords > 0 {
				return fmt.Sprintf("LIMIT %d", opts.MaxRecords)
			}
			return ""
		}())

	// 调试：打印查询条件和参数
	log.Printf("[DEBUG] 时态查询 - code: %s, conditions: %v, args: %v", code, conditions, args)
	log.Printf("[DEBUG] 查询选项 - IncludeHistory: %v, IncludeFuture: %v, IncludeDissolved: %v",
		opts.IncludeHistory, opts.IncludeFuture, opts.IncludeDissolved)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("时态查询失败: %w", err)
	}
	defer rows.Close()

	var organizations []*Organization
	for rows.Next() {
		org := &Organization{}
		var parentCode, changeReason string
		var endDate sql.NullTime
		var isCurrent bool
		var effectiveDate time.Time

		err := rows.Scan(
			&org.RecordID, &org.TenantID, &org.Code, &parentCode, &org.Name,
			&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
			&org.Description, &org.CreatedAt, &org.UpdatedAt,
			&effectiveDate, &endDate, &changeReason, &isCurrent,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描时态查询结果失败: %w", err)
		}

		// 处理字段赋值
		if parentCode != "" {
			org.ParentCode = &parentCode
		}
		org.EffectiveDate = &effectiveDate
		if endDate.Valid {
			org.EndDate = &endDate.Time
		}
		if changeReason != "" {
			org.ChangeReason = &changeReason
		}
		org.IsCurrent = &isCurrent

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

// 创建组织历史记录（使用统一的organization_units表）
func (r *TemporalOrganizationRepository) CreateOrganizationHistory(ctx context.Context, tx *sql.Tx, org *Organization) error {
	// 历史记录已经通过INSERT到organization_units表创建，这里只需记录日志
	log.Printf("✅ 组织历史记录已创建: %s (生效日期: %v)",
		org.Code,
		func() string {
			if org.EffectiveDate != nil {
				return org.EffectiveDate.Format("2006-01-02")
			}
			return "当前时间"
		}())

	// 不需要额外操作，organization_units表本身就是时态数据存储
	return nil
}

// ===== HTTP处理器 =====

type TemporalOrganizationHandler struct {
	repo        *TemporalOrganizationRepository
	db          *sql.DB
	redisClient *redis.Client
	cacheTTL    time.Duration
}

func NewTemporalOrganizationHandler(db *sql.DB) *TemporalOrganizationHandler {
	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &TemporalOrganizationHandler{
		repo:        NewTemporalOrganizationRepository(db),
		db:          db,
		redisClient: redisClient,
		cacheTTL:    5 * time.Minute, // 5分钟缓存TTL
	}
}

// 生成时态查询缓存键
func (h *TemporalOrganizationHandler) getCacheKey(tenantID, code string, opts *TemporalQueryOptions) string {
	hasher := md5.New()
	optsStr := ""
	if opts != nil {
		if opts.AsOfDate != nil {
			optsStr += fmt.Sprintf("asof:%v", opts.AsOfDate.Format("2006-01-02"))
		}
		if opts.EffectiveDate != nil {
			optsStr += fmt.Sprintf("effdate:%v", opts.EffectiveDate.Format("2006-01-02"))
		}
		if opts.EndDate != nil {
			optsStr += fmt.Sprintf("enddate:%v", opts.EndDate.Format("2006-01-02"))
		}
		if opts.IncludeHistory {
			optsStr += ":hist"
		}
		if opts.IncludeFuture {
			optsStr += ":future"
		}
	}
	hasher.Write([]byte(fmt.Sprintf("temporal:%s:%s:%s", tenantID, code, optsStr)))
	return fmt.Sprintf("cache:%x", hasher.Sum(nil))
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

	// 生成缓存键
	cacheKey := h.getCacheKey(tenantID.String(), code, opts)

	// 尝试从缓存获取
	if h.redisClient != nil {
		cachedData, err := h.redisClient.Get(r.Context(), cacheKey).Result()
		if err == nil {
			var cachedResponse map[string]interface{}
			if json.Unmarshal([]byte(cachedData), &cachedResponse) == nil {
				log.Printf("[CACHE HIT] 时态查询缓存命中 - 键: %s, 组织: %s", cacheKey, code)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(cachedResponse)
				return
			}
		}
		log.Printf("[CACHE MISS] 时态查询缓存未命中，查询数据库 - 键: %s", cacheKey)
	}

	// 执行时态查询
	organizations, err := h.repo.GetByCodeTemporal(r.Context(), tenantID, code, opts)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "TEMPORAL_QUERY_ERROR", "时态查询失败", err)
		return
	}

	if len(organizations) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "未找到匹配的组织记录", nil)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"organizations": organizations,
		"query_options": opts,
		"result_count":  len(organizations),
		"queried_at":    time.Now().Format(time.RFC3339),
	}

	// 将结果写入缓存
	if h.redisClient != nil {
		if cacheData, err := json.Marshal(response); err == nil {
			h.redisClient.Set(r.Context(), cacheKey, string(cacheData), h.cacheTTL)
			log.Printf("[CACHE SET] 时态查询结果已缓存 - 键: %s, 组织: %s, TTL: %v", cacheKey, code, h.cacheTTL)
		}
	}

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

	// 2. 处理不同类型的事件（简化处理，不使用版本号）
	switch req.EventType {
	case "UPDATE":
		err = h.handleUpdateEvent(r.Context(), tx, tenantID, code, &req)
	case "RESTRUCTURE":
		err = h.handleRESTRUCTUREEvent(r.Context(), tx, tenantID, code, &req)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// 处理更新事件（无版本号逻辑）
func (h *TemporalOrganizationHandler) handleUpdateEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *OrganizationChangeEvent) error {
	// 获取当前记录
	currentOrg, err := h.getCurrentRecord(ctx, tx, tenantID, code)
	if err != nil {
		return fmt.Errorf("获取当前记录失败: %w", err)
	}

	// 设置当前记录的结束日期
	endDate := req.EffectiveDate.AddDate(0, 0, -1)
	_, err = tx.ExecContext(ctx,
		"UPDATE organization_units SET end_date = $1, is_current = false WHERE code = $2 AND tenant_id = $3 AND is_current = true",
		endDate, code, tenantID.String())
	if err != nil {
		return fmt.Errorf("更新当前记录结束日期失败: %w", err)
	}

	// 创建新记录
	updatedOrg := *currentOrg
	updatedOrg.EffectiveDate = &req.EffectiveDate
	updatedOrg.EndDate = req.EndDate
	updatedOrg.ChangeReason = &req.ChangeReason
	isCurrent := true
	updatedOrg.IsCurrent = &isCurrent

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
		case "parent_code":
			if parentCode, ok := value.(string); ok && parentCode != "" {
				updatedOrg.ParentCode = &parentCode
				// 当父组织变更时，需要重新计算层级信息
				level, path, err := h.calculateHierarchy(ctx, tx, tenantID, parentCode, code)
				if err != nil {
					return fmt.Errorf("重新计算层级信息失败: %w", err)
				}
				updatedOrg.Level = level
				updatedOrg.Path = path
			} else if parentCode == "" {
				// 设置为根组织
				updatedOrg.ParentCode = nil
				updatedOrg.Level = 1
				updatedOrg.Path = "/" + code
			}
		}
	}

	// 插入新记录 - 优化：让触发器处理层级计算，但提供充足的信息
	_, err = tx.ExecContext(ctx, `
		INSERT INTO organization_units (
			code, parent_code, tenant_id, name, unit_type, status, level, path, 
			sort_order, description, effective_date, end_date, change_reason, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		updatedOrg.Code, updatedOrg.ParentCode, updatedOrg.TenantID,
		updatedOrg.Name, updatedOrg.UnitType, updatedOrg.Status,
		updatedOrg.Level, updatedOrg.Path, updatedOrg.SortOrder,
		updatedOrg.Description, updatedOrg.EffectiveDate, updatedOrg.EndDate,
		updatedOrg.ChangeReason, updatedOrg.IsCurrent)

	if err != nil {
		return fmt.Errorf("插入新记录失败: %w", err)
	}

	// 创建历史记录
	if err := h.repo.CreateOrganizationHistory(ctx, tx, &updatedOrg); err != nil {
		return fmt.Errorf("创建历史记录失败: %w", err)
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

	// 更新当前记录的结束日期和状态
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

	// 支持基于effective_date的历史记录状态变更
	var updateQuery string
	var args []interface{}

	if req.EffectiveDate.IsZero() {
		// 如果没有指定生效日期，则更新当前记录
		updateQuery = "UPDATE organization_units SET status = $1, updated_at = NOW() WHERE code = $2 AND tenant_id = $3 AND is_current = true"
		args = []interface{}{newStatus, code, tenantID.String()}
	} else {
		// 如果指定了生效日期，则更新特定日期的记录
		updateQuery = "UPDATE organization_units SET status = $1, updated_at = NOW() WHERE code = $2 AND tenant_id = $3 AND effective_date = $4"
		args = []interface{}{newStatus, code, tenantID.String(), req.EffectiveDate}
	}

	result, err := tx.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("状态变更失败: %w", err)
	}

	// 检查是否有记录被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		if req.EffectiveDate.IsZero() {
			return fmt.Errorf("未找到组织 %s 的当前记录", code)
		} else {
			return fmt.Errorf("未找到组织 %s 在日期 %s 的记录", code, req.EffectiveDate.Format("2006-01-02"))
		}
	}

	log.Printf("✅ 状态变更成功: 组织=%s, 日期=%v, 新状态=%s, 影响记录=%d条",
		code,
		func() string {
			if req.EffectiveDate.IsZero() {
				return "当前记录"
			}
			return req.EffectiveDate.Format("2006-01-02")
		}(),
		newStatus,
		rowsAffected)

	// 如果是DEACTIVATE操作且指定了生效日期，触发gap填充
	if req.EventType == "DEACTIVATE" && !req.EffectiveDate.IsZero() && newStatus == "INACTIVE" {
		log.Printf("🔄 触发gap填充: 组织=%s 的 %s 记录已作废，开始填充时间空洞", code, req.EffectiveDate.Format("2006-01-02"))

		// 执行gap填充 - 使用我们优化过的smart_timeline_fill函数
		_, err := tx.ExecContext(ctx, "SELECT smart_timeline_fill($1)", code)
		if err != nil {
			log.Printf("⚠️ Gap填充失败: %v", err)
			// 不返回错误，允许状态变更成功，但记录gap填充失败
		} else {
			log.Printf("✅ Gap填充完成: 组织=%s 时间轴已优化", code)
		}
	}

	return nil
}

// 处理重组事件
func (h *TemporalOrganizationHandler) handleRESTRUCTUREEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *OrganizationChangeEvent) error {
	// 获取当前记录
	currentOrg, err := h.getCurrentRecord(ctx, tx, tenantID, code)
	if err != nil {
		return fmt.Errorf("获取当前记录失败: %w", err)
	}

	// 正确计算当前记录的结束日期：新记录生效日期前一天
	endDate := req.EffectiveDate.AddDate(0, 0, -1)

	// 时态连续性检查：确保不会产生时间线间隙
	if currentOrg.EffectiveDate != nil && endDate.Before(*currentOrg.EffectiveDate) {
		return fmt.Errorf("时态连续性违反: 结束日期(%s)不能早于当前记录生效日期(%s)",
			endDate.Format("2006-01-02"), currentOrg.EffectiveDate.Format("2006-01-02"))
	}

	// 更新所有当前记录的状态
	_, err = tx.ExecContext(ctx,
		`UPDATE organization_units 
		 SET end_date = $1, is_current = false 
		 WHERE code = $2 AND tenant_id = $3 AND is_current = true`,
		endDate, code, tenantID.String())
	if err != nil {
		return fmt.Errorf("更新当前记录结束日期失败: %w", err)
	}

	// 创建重组后的新记录
	newOrg := *currentOrg
	newOrg.EffectiveDate = &req.EffectiveDate
	newOrg.EndDate = req.EndDate // 可为nil，表示当前生效
	newOrg.ChangeReason = &req.ChangeReason
	isCurrent := true
	newOrg.IsCurrent = &isCurrent

	// 应用重组变更数据
	if changeData, ok := req.ChangeData["unit_type"]; ok {
		if unitType, ok := changeData.(string); ok {
			newOrg.UnitType = unitType
		}
	}
	if changeData, ok := req.ChangeData["name"]; ok {
		if name, ok := changeData.(string); ok {
			newOrg.Name = name
		}
	}
	if changeData, ok := req.ChangeData["parent_code"]; ok {
		if parentCode, ok := changeData.(string); ok && parentCode != "" {
			newOrg.ParentCode = &parentCode
		} else {
			newOrg.ParentCode = nil
		}
	}

	// 插入新的重组记录
	_, err = tx.ExecContext(ctx, `
		INSERT INTO organization_units (
			code, parent_code, tenant_id, name, unit_type, status, level, path, 
			sort_order, description, effective_date, end_date, change_reason, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		newOrg.Code, newOrg.ParentCode, newOrg.TenantID,
		newOrg.Name, newOrg.UnitType, newOrg.Status,
		newOrg.Level, newOrg.Path, newOrg.SortOrder,
		newOrg.Description, newOrg.EffectiveDate, newOrg.EndDate,
		newOrg.ChangeReason, newOrg.IsCurrent)

	if err != nil {
		return fmt.Errorf("插入重组记录失败: %w", err)
	}

	// 创建历史记录
	if err := h.repo.CreateOrganizationHistory(ctx, tx, &newOrg); err != nil {
		return fmt.Errorf("创建重组历史记录失败: %w", err)
	}

	return nil
}

// 获取当前记录
func (h *TemporalOrganizationHandler) getCurrentRecord(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*Organization, error) {
	query := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, change_reason, is_current
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND is_current = true
	`

	org := &Organization{}
	var changeReason, endDate sql.NullString
	var isCurrent sql.NullBool
	var effectiveDate sql.NullTime

	err := tx.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason, &isCurrent,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("未找到组织 %s 的当前记录", code)
	} else if err != nil {
		return nil, fmt.Errorf("查询当前记录失败: %w", err)
	}

	// 处理NULL值
	if effectiveDate.Valid {
		org.EffectiveDate = &effectiveDate.Time
	}
	if endDate.Valid {
		t, _ := time.Parse("2006-01-02", endDate.String)
		org.EndDate = &t
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}
	if isCurrent.Valid {
		org.IsCurrent = &isCurrent.Bool
	}

	return org, nil
}

// 计算组织层级信息 - 为时态记录创建提供准确的层级数据
func (h *TemporalOrganizationHandler) calculateHierarchy(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, parentCode, currentCode string) (int, string, error) {
	if parentCode == "" {
		// 根组织
		return 1, "/" + currentCode, nil
	}

	// 查询父组织的当前记录
	query := `
		SELECT level, path 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND is_current = true
	`

	var parentLevel int
	var parentPath string
	err := tx.QueryRowContext(ctx, query, tenantID.String(), parentCode).Scan(&parentLevel, &parentPath)

	if err == sql.ErrNoRows {
		// 如果父组织当前记录不存在，查找最新的记录
		query = `
			SELECT level, path 
			FROM organization_units 
			WHERE tenant_id = $1 AND code = $2 
			ORDER BY effective_date DESC 
			LIMIT 1
		`
		err = tx.QueryRowContext(ctx, query, tenantID.String(), parentCode).Scan(&parentLevel, &parentPath)

		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("父组织 %s 不存在", parentCode)
		} else if err != nil {
			return 0, "", fmt.Errorf("查询父组织层级信息失败: %w", err)
		}
	} else if err != nil {
		return 0, "", fmt.Errorf("查询父组织当前记录失败: %w", err)
	}

	// 计算当前组织的层级和路径
	currentLevel := parentLevel + 1
	currentPath := parentPath + "/" + currentCode

	return currentLevel, currentPath, nil
}

// 历史记录更新请求结构
type UpdateHistoryRecordRequest struct {
	Name           string  `json:"name"`
	UnitType       string  `json:"unit_type"`
	Status         string  `json:"status"`
	Description    string  `json:"description"`
	EffectiveDate  string  `json:"effective_date"`
	ParentCode     *string `json:"parent_code,omitempty"`
	ChangeReason   string  `json:"change_reason"`
}

// 历史记录直接更新处理器
func (h *TemporalOrganizationHandler) UpdateHistoryRecord(w http.ResponseWriter, r *http.Request) {
	recordID := chi.URLParam(r, "record_id")
	if recordID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_RECORD_ID", "缺少记录ID", nil)
		return
	}

	var req UpdateHistoryRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 验证必填字段
	if req.Name == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_NAME", "组织名称是必填项", nil)
		return
	}
	if req.UnitType == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_UNIT_TYPE", "组织类型是必填项", nil)
		return
	}
	if req.Status == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_STATUS", "组织状态是必填项", nil)
		return
	}

	tenantID := h.getTenantID(r)

	// 解析生效日期
	var effectiveDate time.Time
	var err error
	if req.EffectiveDate != "" {
		effectiveDate, err = time.Parse("2006-01-02", req.EffectiveDate)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_EFFECTIVE_DATE", "生效日期格式无效", err)
			return
		}
	}

	// 开始事务
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "TRANSACTION_ERROR", "开始事务失败", err)
		return
	}
	defer tx.Rollback()

	// 首先检查记录是否存在
	var existingOrg Organization
	checkQuery := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, change_reason, is_current
		FROM organization_units 
		WHERE record_id = $1 AND tenant_id = $2
	`

	var changeReason, endDate sql.NullString
	var isCurrent sql.NullBool
	var effectiveDateDB sql.NullTime

	err = tx.QueryRowContext(r.Context(), checkQuery, recordID, tenantID.String()).Scan(
		&existingOrg.RecordID, &existingOrg.TenantID, &existingOrg.Code, &existingOrg.ParentCode, &existingOrg.Name,
		&existingOrg.UnitType, &existingOrg.Status, &existingOrg.Level, &existingOrg.Path, &existingOrg.SortOrder,
		&existingOrg.Description, &existingOrg.CreatedAt, &existingOrg.UpdatedAt,
		&effectiveDateDB, &endDate, &changeReason, &isCurrent,
	)

	if err == sql.ErrNoRows {
		h.writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "未找到指定的历史记录", nil)
		return
	} else if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "QUERY_ERROR", "查询记录失败", err)
		return
	}

	// 处理NULL值
	if effectiveDateDB.Valid {
		existingOrg.EffectiveDate = &effectiveDateDB.Time
	}
	if endDate.Valid {
		t, _ := time.Parse("2006-01-02", endDate.String)
		existingOrg.EndDate = &t
	}
	if changeReason.Valid {
		existingOrg.ChangeReason = &changeReason.String
	}
	if isCurrent.Valid {
		existingOrg.IsCurrent = &isCurrent.Bool
	}

	// 构建更新语句
	updateQuery := `
		UPDATE organization_units 
		SET name = $1, unit_type = $2, status = $3, description = $4, 
		    parent_code = $5, effective_date = $6, change_reason = $7, updated_at = NOW()
		WHERE record_id = $8 AND tenant_id = $9
	`

	// 执行更新
	result, err := tx.ExecContext(r.Context(), updateQuery,
		req.Name, req.UnitType, req.Status, req.Description,
		req.ParentCode, effectiveDate, req.ChangeReason,
		recordID, tenantID.String())

	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "UPDATE_ERROR", "更新历史记录失败", err)
		return
	}

	// 检查是否有记录被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "UPDATE_CHECK_ERROR", "检查更新结果失败", err)
		return
	}

	if rowsAffected == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, "NOT_UPDATED", "没有记录被更新", nil)
		return
	}

	// 如果父组织变更，需要重新计算层级信息
	if req.ParentCode != nil && (existingOrg.ParentCode == nil || *req.ParentCode != *existingOrg.ParentCode) {
		level, path, err := h.calculateHierarchy(r.Context(), tx, tenantID, *req.ParentCode, existingOrg.Code)
		if err != nil {
			log.Printf("⚠️ 重新计算层级信息失败: %v", err)
			// 不返回错误，允许更新继续完成
		} else {
			// 更新层级信息
			_, err = tx.ExecContext(r.Context(),
				"UPDATE organization_units SET level = $1, path = $2 WHERE record_id = $3",
				level, path, recordID)
			if err != nil {
				log.Printf("⚠️ 更新层级信息失败: %v", err)
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "COMMIT_ERROR", "提交事务失败", err)
		return
	}

	// 清除相关缓存 - 使用组织代码进行缓存失效
	if h.redisClient != nil {
		// 模糊匹配并删除与该组织相关的所有缓存
		ctx := r.Context()
		keys, err := h.redisClient.Keys(ctx, fmt.Sprintf("cache:*:%s:*", existingOrg.Code)).Result()
		if err == nil && len(keys) > 0 {
			h.redisClient.Del(ctx, keys...)
			log.Printf("[CACHE CLEAR] 历史记录更新后清除缓存 - 组织: %s, 清除键数: %d", existingOrg.Code, len(keys))
		}
	}

	// 构建响应
	response := map[string]interface{}{
		"record_id":      recordID,
		"code":           existingOrg.Code,
		"name":           req.Name,
		"unit_type":      req.UnitType,
		"status":         req.Status,
		"effective_date": req.EffectiveDate,
		"updated_at":     time.Now().Format(time.RFC3339),
		"message":        "历史记录更新成功",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ 历史记录更新成功: 记录ID=%s, 组织=%s, 名称=%s", recordID, existingOrg.Code, req.Name)
}

// 时间线事件结构
type TimelineEvent struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	EventType     string                 `json:"event_type"`
	EventDate     string                 `json:"event_date"`
	EffectiveDate string                 `json:"effective_date"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	TriggeredBy   string                 `json:"triggered_by,omitempty"`
}

// 获取组织时间线
func (h *TemporalOrganizationHandler) GetOrganizationTimeline(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)
	limit := 50 // 默认限制50条记录
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// 查询时间线事件
	events, err := h.getTimelineEvents(r.Context(), tenantID, code, limit)
	if err != nil {
		log.Printf("获取时间线事件失败: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "QUERY_FAILED", "获取时间线失败", err)
		return
	}

	response := map[string]interface{}{
		"timeline":     events,
		"count":        len(events),
		"organization": code,
		"queried_at":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取时间线事件数据
func (h *TemporalOrganizationHandler) getTimelineEvents(ctx context.Context, tenantID uuid.UUID, code string, limit int) ([]TimelineEvent, error) {
	// 查询组织的所有历史记录，按创建时间倒序
	query := `
		SELECT 
			record_id,
			code,
			name,
			unit_type,
			status,
			effective_date,
			end_date,
			change_reason,
			created_at,
			updated_at,
			is_current
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
		ORDER BY created_at DESC, effective_date DESC
		LIMIT $3`

	rows, err := h.db.QueryContext(ctx, query, tenantID.String(), code, limit)
	if err != nil {
		return nil, fmt.Errorf("查询历史记录失败: %w", err)
	}
	defer rows.Close()

	var events []TimelineEvent
	for rows.Next() {
		var (
			recordID       string
			orgCode        string
			name           string
			unitType       string
			status         string
			effectiveDate  *time.Time
			endDate        *time.Time
			changeReason   *string
			createdAt      time.Time
			updatedAt      time.Time
			isCurrent      *bool
		)

		err := rows.Scan(
			&recordID, &orgCode, &name, &unitType, &status,
			&effectiveDate, &endDate, &changeReason,
			&createdAt, &updatedAt, &isCurrent,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描记录失败: %w", err)
		}

		// 确定事件类型和描述
		eventType, title, description := h.determineEventType(
			name, unitType, status, effectiveDate, endDate, 
			changeReason, isCurrent, createdAt, updatedAt,
		)

		// 构建时间线事件
		event := TimelineEvent{
			ID:            fmt.Sprintf("%s_%d", orgCode, createdAt.Unix()),
			Title:         title,
			Description:   description,
			EventType:     eventType,
			EventDate:     createdAt.Format(time.RFC3339),
			EffectiveDate: formatTimePtr(effectiveDate),
			Status:        status,
			Metadata: map[string]interface{}{
				"name":        name,
				"unit_type":   unitType,
				"end_date":    formatTimePtr(endDate),
				"is_current":  isCurrent,
				"updated_at":  updatedAt.Format(time.RFC3339),
			},
			TriggeredBy: "系统用户", // 可以后续扩展为实际用户信息
		}

		if changeReason != nil {
			event.Metadata["change_reason"] = *changeReason
		}

		events = append(events, event)
	}

	return events, nil
}

// 确定事件类型和描述
func (h *TemporalOrganizationHandler) determineEventType(
	name, unitType, status string,
	effectiveDate, endDate *time.Time,
	changeReason *string,
	isCurrent *bool,
	createdAt, updatedAt time.Time,
) (eventType, title, description string) {
	
	// 根据时间和状态判断事件类型
	now := time.Now()
	isActive := status == "ACTIVE"
	isPlanned := status == "PLANNED"
	
	// 判断是否是创建事件（通常创建时间和更新时间相近）
	isCreation := updatedAt.Sub(createdAt).Seconds() < 5
	
	// 判断是否已结束
	isEnded := endDate != nil && endDate.Before(now)

	switch {
	case isCreation && isPlanned:
		eventType = "create"
		title = fmt.Sprintf("创建计划组织: %s", name)
		description = fmt.Sprintf("新建了%s类型的计划组织，预计于%s生效", 
			h.getUnitTypeName(unitType), formatTimePtr(effectiveDate))
			
	case isCreation && isActive:
		eventType = "create"
		title = fmt.Sprintf("创建组织: %s", name)
		description = fmt.Sprintf("新建了%s类型的组织单元，立即生效", 
			h.getUnitTypeName(unitType))
			
	case !isCreation && isActive && isCurrent != nil && *isCurrent:
		eventType = "activate"
		title = fmt.Sprintf("激活组织: %s", name)
		description = fmt.Sprintf("组织单元状态变更为激活")
		
	case !isCreation && status == "INACTIVE":
		eventType = "deactivate"
		title = fmt.Sprintf("停用组织: %s", name)
		description = fmt.Sprintf("组织单元状态变更为停用")
		
	case !isCreation && isEnded:
		eventType = "dissolve"
		title = fmt.Sprintf("解散组织: %s", name)
		description = fmt.Sprintf("组织单元于%s解散", formatTimePtr(endDate))
		
	case !isCreation:
		eventType = "update"
		title = fmt.Sprintf("更新组织: %s", name)
		description = fmt.Sprintf("组织信息发生变更")
		
	default:
		eventType = "update"
		title = fmt.Sprintf("组织变更: %s", name)
		description = fmt.Sprintf("组织单元信息更新")
	}
	
	// 添加变更原因到描述中
	if changeReason != nil && *changeReason != "" {
		description += fmt.Sprintf("，变更原因：%s", *changeReason)
	}
	
	return eventType, title, description
}

// 获取组织类型中文名
func (h *TemporalOrganizationHandler) getUnitTypeName(unitType string) string {
	typeNames := map[string]string{
		"COMPANY":      "公司",
		"DEPARTMENT":   "部门",
		"COST_CENTER":  "成本中心",
		"PROJECT_TEAM": "项目团队",
	}
	if name, exists := typeNames[unitType]; exists {
		return name
	}
	return unitType
}

// 格式化时间指针
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
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
			"service":      "Organization Temporal Command Service",
			"version":      "2.0.0",
			"status":       "healthy",
			"timestamp":    time.Now().Format(time.RFC3339),
			"architecture": "CQRS Temporal Side - 时态查询和事件管理",
			"features":     []string{"temporal-queries", "event-driven-changes", "date-based-versioning"},
		})
	})

	// 根路径信息 - 时态服务完整接口文档
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":      "Organization Temporal Command Service",
			"version":      "2.0.0",
			"architecture": "CQRS Temporal Side - 时态查询和事件管理",
			"endpoints": map[string]string{
				"temporal_query": "GET /api/v1/organization-units/{code}/temporal?as_of_date=YYYY-MM-DD",
				"create_event":   "POST /api/v1/organization-units/{code}/events",
				"health":         "GET /health",
				"metrics":        "GET /metrics",
			},
			"query_parameters": map[string]string{
				"as_of_date":     "查询指定日期的组织状态 (YYYY-MM-DD)",
				"effective_from": "查询时间范围起始日期 (YYYY-MM-DD)",
				"effective_to":   "查询时间范围结束日期 (YYYY-MM-DD)",
			},
			"temporal_features": []string{
				"纯日期生效模型 - 符合行业标准",
				"时间点查询 - as_of_date参数支持",
				"时间范围查询 - effective_from/to参数支持",
				"事件驱动变更 - UPDATE/RESTRUCTURE/DISSOLVE支持",
				"缓存优化 - Redis缓存提升查询性能",
			},
			"note": "本服务专注时态查询，常规CRUD操作请使用命令服务(9090)或查询服务(8090)",
		})
	})

	// 监控指标
	r.Handle("/metrics", promhttp.Handler())

	// API路由
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		// 时态查询端点
		r.Get("/{code}/temporal", handler.GetOrganizationTemporal)

		// 时间线可视化端点 - 新增
		r.Get("/{code}/timeline", handler.GetOrganizationTimeline)

		// 事件驱动变更端点
		r.Post("/{code}/events", handler.CreateOrganizationEvent)

		// 历史记录直接更新端点 - 新增
		r.Put("/history/{record_id}", handler.UpdateHistoryRecord)

		// 时态查询端点的查询字符串版本
		r.Get("/{code}", handler.GetOrganizationTemporal) // 支持时态查询参数
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "9091" // 使用9091端口避免与命令服务冲突
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		log.Printf("🚀 时态组织命令服务启动在端口 %s (无版本号模式)", port)
		log.Println("📋 支持的功能:")
		log.Println("  - 时态查询 (as_of_date, effective_from, effective_to)")
		log.Println("  - 事件驱动变更 (UPDATE, RESTRUCTURE, DISSOLVE)")
		log.Println("  - 纯日期生效管理（符合行业标准）")
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
