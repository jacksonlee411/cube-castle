package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

// ===== 数据模型 =====

type OrganizationUnit struct {
	Code              string     `json:"code"`
	TenantID          string     `json:"tenantId"`
	Name              string     `json:"name"`
	UnitType          string     `json:"unitType"`
	Status            string     `json:"status"`
	ParentCode        *string    `json:"parentCode,omitempty"`
	Description       *string    `json:"description,omitempty"`
	Profile           *string    `json:"profile,omitempty"`
	EffectiveDate     string     `json:"effectiveDate"`
	EndDate           *string    `json:"endDate,omitempty"`
	ChangeReason      *string    `json:"changeReason,omitempty"`
	IsCurrent         bool       `json:"isCurrent"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// 时态命令请求结构
type CreateTemporalOrganizationRequest struct {
	Name          string     `json:"name" validate:"required"`
	UnitType      string     `json:"unitType" validate:"required"`
	Status        string     `json:"status" validate:"required"`
	ParentCode    *string    `json:"parentCode,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Profile       *string    `json:"profile,omitempty"`
	EffectiveDate string     `json:"effectiveDate" validate:"required"`
	EndDate       *string    `json:"endDate,omitempty"`
	ChangeReason  *string    `json:"changeReason,omitempty"`
}

type UpdateTemporalOrganizationRequest struct {
	Name          *string    `json:"name,omitempty"`
	UnitType      *string    `json:"unitType,omitempty"`
	Status        *string    `json:"status,omitempty"`
	ParentCode    *string    `json:"parentCode,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Profile       *string    `json:"profile,omitempty"`
	EffectiveDate string     `json:"effectiveDate" validate:"required"`
	EndDate       *string    `json:"endDate,omitempty"`
	ChangeReason  *string    `json:"changeReason,omitempty"`
}

type TemporalEventRequest struct {
	EventType      string                 `json:"eventType" validate:"required"` // CREATE, UPDATE, RESTRUCTURE, DISSOLVE
	EffectiveDate  string                 `json:"effectiveDate" validate:"required"`
	EndDate        *string                `json:"endDate,omitempty"`
	ChangeReason   *string                `json:"changeReason,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"` // 变更数据
}

// API响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type TemporalQueryResponse struct {
	Organizations []OrganizationUnit     `json:"organizations"`
	QueriedAt     string                 `json:"queriedAt"`
	QueryOptions  map[string]interface{} `json:"queryOptions"`
}

// ===== 服务结构 =====

type TemporalCommandService struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewTemporalCommandService(db *sql.DB, redisClient *redis.Client) *TemporalCommandService {
	return &TemporalCommandService{
		db:          db,
		redisClient: redisClient,
	}
}

// ===== 时态CRUD操作 =====

// CreateTemporalOrganization 创建时态组织
func (s *TemporalCommandService) CreateTemporalOrganization(ctx context.Context, req *CreateTemporalOrganizationRequest) (*OrganizationUnit, error) {
	// 生成新的组织代码
	code := s.generateOrganizationCode()

	// 验证有效日期格式
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("无效的生效日期格式: %w", err)
	}

	// 如果指定了结束日期，验证其格式
	var endDate *time.Time
	if req.EndDate != nil {
		parsedEndDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("无效的结束日期格式: %w", err)
		}
		if parsedEndDate.Before(effectiveDate) || parsedEndDate.Equal(effectiveDate) {
			return nil, fmt.Errorf("结束日期必须晚于生效日期")
		}
		endDate = &parsedEndDate
	}

	// 检查父组织是否存在（如果指定了父组织）
	if req.ParentCode != nil && *req.ParentCode != "" {
		var exists bool
		err := s.db.QueryRowContext(ctx, 
			`SELECT EXISTS(SELECT 1 FROM organization_units 
			 WHERE code = $1 AND tenant_id = $2 AND is_current = true)`,
			*req.ParentCode, DefaultTenantIDString).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("检查父组织失败: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("父组织代码 %s 不存在", *req.ParentCode)
		}
	}

	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 插入新的组织记录
	now := time.Now()
	query := `
		INSERT INTO organization_units (
			code, tenant_id, name, unit_type, status, parent_code,
			description, profile, effective_date, end_date, 
			is_current, change_reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err = tx.ExecContext(ctx, query,
		code, DefaultTenantIDString, req.Name, req.UnitType, req.Status,
		req.ParentCode, req.Description, req.Profile,
		effectiveDate, endDate, true, req.ChangeReason, now, now)

	if err != nil {
		return nil, fmt.Errorf("插入组织记录失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 返回创建的组织
	org := &OrganizationUnit{
		Code:          code,
		TenantID:      DefaultTenantIDString,
		Name:          req.Name,
		UnitType:      req.UnitType,
		Status:        req.Status,
		ParentCode:    req.ParentCode,
		Description:   req.Description,
		Profile:       req.Profile,
		EffectiveDate: req.EffectiveDate,
		EndDate:       req.EndDate,
		ChangeReason:  req.ChangeReason,
		IsCurrent:     true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 清除相关缓存
	s.invalidateCache(ctx, code)

	log.Printf("✅ 创建时态组织成功: %s (%s) 生效日期: %s", code, req.Name, req.EffectiveDate)
	return org, nil
}

// UpdateTemporalOrganization 更新时态组织（创建新版本）
func (s *TemporalCommandService) UpdateTemporalOrganization(ctx context.Context, code string, req *UpdateTemporalOrganizationRequest) (*OrganizationUnit, error) {
	// 验证有效日期格式
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("无效的生效日期格式: %w", err)
	}

	// 获取当前版本
	currentOrg, err := s.getCurrentOrganization(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("获取当前组织失败: %w", err)
	}
	if currentOrg == nil {
		return nil, fmt.Errorf("组织代码 %s 不存在", code)
	}

	// 检查新的生效日期是否合理
	currentEffectiveDate, _ := time.Parse("2006-01-02", currentOrg.EffectiveDate)
	if effectiveDate.Before(currentEffectiveDate) || effectiveDate.Equal(currentEffectiveDate) {
		return nil, fmt.Errorf("新版本生效日期必须晚于当前版本 (%s)", currentOrg.EffectiveDate)
	}

	// 如果指定了结束日期，验证其格式
	var endDate *time.Time
	if req.EndDate != nil {
		parsedEndDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("无效的结束日期格式: %w", err)
		}
		if parsedEndDate.Before(effectiveDate) || parsedEndDate.Equal(effectiveDate) {
			return nil, fmt.Errorf("结束日期必须晚于生效日期")
		}
		endDate = &parsedEndDate
	}

	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 更新当前版本为非当前
	_, err = tx.ExecContext(ctx,
		`UPDATE organization_units 
		 SET is_current = false, end_date = $1, updated_at = $2
		 WHERE code = $3 AND tenant_id = $4 AND is_current = true`,
		effectiveDate.Format("2006-01-02"), time.Now(), code, DefaultTenantIDString)

	if err != nil {
		return nil, fmt.Errorf("更新当前版本状态失败: %w", err)
	}

	// 创建新版本（合并当前值和更新值）
	newOrg := &OrganizationUnit{
		Code:        code,
		TenantID:    DefaultTenantIDString,
		Name:        currentOrg.Name,
		UnitType:    currentOrg.UnitType,
		Status:      currentOrg.Status,
		ParentCode:  currentOrg.ParentCode,
		Description: currentOrg.Description,
		Profile:     currentOrg.Profile,
	}

	// 应用更新字段
	if req.Name != nil {
		newOrg.Name = *req.Name
	}
	if req.UnitType != nil {
		newOrg.UnitType = *req.UnitType
	}
	if req.Status != nil {
		newOrg.Status = *req.Status
	}
	if req.ParentCode != nil {
		newOrg.ParentCode = req.ParentCode
	}
	if req.Description != nil {
		newOrg.Description = req.Description
	}
	if req.Profile != nil {
		newOrg.Profile = req.Profile
	}

	// 插入新版本记录
	now := time.Now()
	query := `
		INSERT INTO organization_units (
			code, tenant_id, name, unit_type, status, parent_code,
			description, profile, effective_date, end_date, 
			is_current, change_reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err = tx.ExecContext(ctx, query,
		code, DefaultTenantIDString, newOrg.Name, newOrg.UnitType, newOrg.Status,
		newOrg.ParentCode, newOrg.Description, newOrg.Profile,
		effectiveDate, endDate, true, req.ChangeReason, now, now)

	if err != nil {
		return nil, fmt.Errorf("插入新版本记录失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新返回对象
	newOrg.EffectiveDate = req.EffectiveDate
	newOrg.EndDate = req.EndDate
	newOrg.ChangeReason = req.ChangeReason
	newOrg.IsCurrent = true
	newOrg.CreatedAt = now
	newOrg.UpdatedAt = now

	// 清除相关缓存
	s.invalidateCache(ctx, code)

	log.Printf("✅ 更新时态组织成功: %s 新生效日期: %s", code, req.EffectiveDate)
	return newOrg, nil
}

// DissolveOrganization 解散组织（设置结束日期）
func (s *TemporalCommandService) DissolveOrganization(ctx context.Context, code string, endDate string, changeReason *string) error {
	// 验证日期格式
	parsedEndDate, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("无效的结束日期格式: %w", err)
	}

	// 获取当前版本
	currentOrg, err := s.getCurrentOrganization(ctx, code)
	if err != nil {
		return fmt.Errorf("获取当前组织失败: %w", err)
	}
	if currentOrg == nil {
		return fmt.Errorf("组织代码 %s 不存在", code)
	}

	// 检查结束日期是否合理
	currentEffectiveDate, _ := time.Parse("2006-01-02", currentOrg.EffectiveDate)
	if parsedEndDate.Before(currentEffectiveDate) || parsedEndDate.Equal(currentEffectiveDate) {
		return fmt.Errorf("解散日期必须晚于当前生效日期 (%s)", currentOrg.EffectiveDate)
	}

	// 更新当前版本的结束日期
	_, err = s.db.ExecContext(ctx,
		`UPDATE organization_units 
		 SET end_date = $1, change_reason = $2, updated_at = $3
		 WHERE code = $4 AND tenant_id = $5 AND is_current = true`,
		endDate, changeReason, time.Now(), code, DefaultTenantIDString)

	if err != nil {
		return fmt.Errorf("更新解散日期失败: %w", err)
	}

	// 清除相关缓存
	s.invalidateCache(ctx, code)

	log.Printf("✅ 解散组织成功: %s 解散日期: %s", code, endDate)
	return nil
}

// ===== 时态查询操作 =====

// GetTemporalOrganization 获取组织的时态信息
func (s *TemporalCommandService) GetTemporalOrganization(ctx context.Context, code string, asOfDate *string, effectiveFrom *string, effectiveTo *string) (*TemporalQueryResponse, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 基本条件
	conditions = append(conditions, fmt.Sprintf("code = $%d", argIndex))
	args = append(args, code)
	argIndex++

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, DefaultTenantIDString)
	argIndex++

	queryOptions := make(map[string]interface{})

	// 时间点查询
	if asOfDate != nil {
		// 验证日期格式
		_, err := time.Parse("2006-01-02", *asOfDate)
		if err != nil {
			return nil, fmt.Errorf("无效的查询日期格式: %w", err)
		}

		conditions = append(conditions, 
			fmt.Sprintf("effective_date <= $%d", argIndex),
			fmt.Sprintf("(end_date IS NULL OR end_date > $%d)", argIndex))
		args = append(args, *asOfDate)
		argIndex++

		queryOptions["as_of_date"] = *asOfDate + "T00:00:00Z"
	} else {
		// 默认查询当前版本
		conditions = append(conditions, "is_current = true")
	}

	// 时间范围查询
	if effectiveFrom != nil {
		_, err := time.Parse("2006-01-02", *effectiveFrom)
		if err != nil {
			return nil, fmt.Errorf("无效的起始日期格式: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("effective_date >= $%d", argIndex))
		args = append(args, *effectiveFrom)
		argIndex++
		queryOptions["effective_from"] = *effectiveFrom
	}

	if effectiveTo != nil {
		_, err := time.Parse("2006-01-02", *effectiveTo)
		if err != nil {
			return nil, fmt.Errorf("无效的结束日期格式: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("effective_date <= $%d", argIndex))
		args = append(args, *effectiveTo)
		argIndex++
		queryOptions["effective_to"] = *effectiveTo
	}

	// 构建查询
	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
		SELECT code, tenant_id, name, unit_type, status, parent_code,
		       description, profile, effective_date, end_date, 
		       is_current, change_reason, created_at, updated_at
		FROM organization_units 
		WHERE %s
		ORDER BY effective_date DESC`, whereClause)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询时态组织失败: %w", err)
	}
	defer rows.Close()

	var organizations []OrganizationUnit
	for rows.Next() {
		org := OrganizationUnit{}
		var endDate sql.NullString
		var changeReason sql.NullString
		var parentCode sql.NullString
		var description sql.NullString
		var profile sql.NullString

		err := rows.Scan(
			&org.Code, &org.TenantID, &org.Name, &org.UnitType, &org.Status,
			&parentCode, &description, &profile,
			&org.EffectiveDate, &endDate, &org.IsCurrent, &changeReason,
			&org.CreatedAt, &org.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("扫描组织记录失败: %w", err)
		}

		if endDate.Valid {
			org.EndDate = &endDate.String
		}
		if changeReason.Valid {
			org.ChangeReason = &changeReason.String
		}
		if parentCode.Valid && parentCode.String != "" {
			org.ParentCode = &parentCode.String
		}
		if description.Valid && description.String != "" {
			org.Description = &description.String
		}
		if profile.Valid && profile.String != "" {
			org.Profile = &profile.String
		}

		organizations = append(organizations, org)
	}

	return &TemporalQueryResponse{
		Organizations: organizations,
		QueriedAt:    time.Now().Format(time.RFC3339),
		QueryOptions: queryOptions,
	}, nil
}

// ===== 辅助方法 =====

// getCurrentOrganization 获取组织的当前版本
func (s *TemporalCommandService) getCurrentOrganization(ctx context.Context, code string) (*OrganizationUnit, error) {
	query := `
		SELECT code, tenant_id, name, unit_type, status, parent_code,
		       description, profile, effective_date, end_date, 
		       is_current, change_reason, created_at, updated_at
		FROM organization_units 
		WHERE code = $1 AND tenant_id = $2 AND is_current = true`

	org := &OrganizationUnit{}
	var endDate sql.NullString
	var changeReason sql.NullString
	var parentCode sql.NullString
	var description sql.NullString
	var profile sql.NullString

	err := s.db.QueryRowContext(ctx, query, code, DefaultTenantIDString).Scan(
		&org.Code, &org.TenantID, &org.Name, &org.UnitType, &org.Status,
		&parentCode, &description, &profile,
		&org.EffectiveDate, &endDate, &org.IsCurrent, &changeReason,
		&org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if endDate.Valid {
		org.EndDate = &endDate.String
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}
	if parentCode.Valid && parentCode.String != "" {
		org.ParentCode = &parentCode.String
	}
	if description.Valid && description.String != "" {
		org.Description = &description.String
	}
	if profile.Valid && profile.String != "" {
		org.Profile = &profile.String
	}

	return org, nil
}

// generateOrganizationCode 生成组织代码
func (s *TemporalCommandService) generateOrganizationCode() string {
	// 查询当前最大代码
	var maxCode int
	err := s.db.QueryRow(
		"SELECT COALESCE(MAX(CAST(code AS INTEGER)), 1000000) FROM organization_units WHERE tenant_id = $1",
		DefaultTenantIDString).Scan(&maxCode)
	
	if err != nil {
		// 如果查询失败，从1000001开始
		maxCode = 1000000
	}
	
	return fmt.Sprintf("%07d", maxCode+1)
}

// invalidateCache 清除相关缓存
func (s *TemporalCommandService) invalidateCache(ctx context.Context, code string) {
	// 清除相关的缓存键
	cacheKeys := []string{
		fmt.Sprintf("orgs:all:%s", DefaultTenantIDString),
		fmt.Sprintf("org:%s:%s", DefaultTenantIDString, code),
		fmt.Sprintf("org_temporal:%s:%s", DefaultTenantIDString, code),
	}

	for _, key := range cacheKeys {
		s.redisClient.Del(ctx, key)
	}
	
	log.Printf("🗑️ 清除缓存: %v", cacheKeys)
}

// ===== HTTP处理器 =====

func (s *TemporalCommandService) createTemporalOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTemporalOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "无效的请求数据", err)
		return
	}

	org, err := s.CreateTemporalOrganization(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "格式") {
			s.sendErrorResponse(w, http.StatusBadRequest, "创建失败", err)
		} else {
			s.sendErrorResponse(w, http.StatusInternalServerError, "创建失败", err)
		}
		return
	}

	s.sendSuccessResponse(w, http.StatusCreated, org, "创建时态组织成功")
}

func (s *TemporalCommandService) updateTemporalOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "缺少组织代码", nil)
		return
	}

	var req UpdateTemporalOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "无效的请求数据", err)
		return
	}

	org, err := s.UpdateTemporalOrganization(r.Context(), code, &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "格式") {
			s.sendErrorResponse(w, http.StatusBadRequest, "更新失败", err)
		} else {
			s.sendErrorResponse(w, http.StatusInternalServerError, "更新失败", err)
		}
		return
	}

	s.sendSuccessResponse(w, http.StatusOK, org, "更新时态组织成功")
}

func (s *TemporalCommandService) dissolveOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "缺少组织代码", nil)
		return
	}

	var req struct {
		EndDate      string  `json:"endDate" validate:"required"`
		ChangeReason *string `json:"changeReason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "无效的请求数据", err)
		return
	}

	err := s.DissolveOrganization(r.Context(), code, req.EndDate, req.ChangeReason)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || strings.Contains(err.Error(), "格式") {
			s.sendErrorResponse(w, http.StatusBadRequest, "解散失败", err)
		} else {
			s.sendErrorResponse(w, http.StatusInternalServerError, "解散失败", err)
		}
		return
	}

	s.sendSuccessResponse(w, http.StatusOK, nil, "组织解散成功")
}

func (s *TemporalCommandService) getTemporalOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "缺少组织代码", nil)
		return
	}

	// 解析查询参数
	query := r.URL.Query()
	var asOfDate, effectiveFrom, effectiveTo *string

	if date := query.Get("as_of_date"); date != "" {
		asOfDate = &date
	}
	if from := query.Get("effective_from"); from != "" {
		effectiveFrom = &from
	}
	if to := query.Get("effective_to"); to != "" {
		effectiveTo = &to
	}

	result, err := s.GetTemporalOrganization(r.Context(), code, asOfDate, effectiveFrom, effectiveTo)
	if err != nil {
		if strings.Contains(err.Error(), "格式") {
			s.sendErrorResponse(w, http.StatusBadRequest, "查询失败", err)
		} else {
			s.sendErrorResponse(w, http.StatusInternalServerError, "查询失败", err)
		}
		return
	}

	s.sendSuccessResponse(w, http.StatusOK, result, "查询时态组织成功")
}

// ===== 响应辅助方法 =====

func (s *TemporalCommandService) sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	json.NewEncoder(w).Encode(response)
}

func (s *TemporalCommandService) sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	
	response := APIResponse{
		Success:   false,
		Message:   message,
		Error:     errorMsg,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	json.NewEncoder(w).Encode(response)
}

// ===== 主程序 =====

func main() {
	// 数据库连接
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "cubecastle")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
	log.Println("✅ PostgreSQL连接成功")

	// Redis连接
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 测试Redis连接
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("连接Redis失败: %v", err)
	}
	log.Println("✅ Redis连接成功")

	// 创建服务
	service := NewTemporalCommandService(db, redisClient)

	// 设置路由
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "organization-temporal-command-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"features":  []string{"temporal-crud", "version-management", "organization-lifecycle"},
		})
	})

	// 监控指标
	r.Handle("/metrics", promhttp.Handler())

	// API路由
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		// 时态命令操作
		r.Post("/", service.createTemporalOrganizationHandler)
		r.Put("/{code}", service.updateTemporalOrganizationHandler)
		r.Post("/{code}/dissolve", service.dissolveOrganizationHandler)
		
		// 时态查询操作
		r.Get("/{code}/temporal", service.getTemporalOrganizationHandler)
	})

	// 启动服务器
	port := getEnv("PORT", "9092")
	
	log.Printf("🚀 时态命令服务启动在端口 %s", port)
	log.Println("📋 支持的功能:")
	log.Println("  - 时态组织创建 (POST /api/v1/organization-units)")
	log.Println("  - 时态组织更新 (PUT /api/v1/organization-units/{code})")
	log.Println("  - 组织解散 (POST /api/v1/organization-units/{code}/dissolve)")
	log.Println("  - 时态查询 (GET /api/v1/organization-units/{code}/temporal)")
	log.Printf("🌐 健康检查: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}