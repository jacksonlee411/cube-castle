package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 自定义日期类型 =====
type Date struct {
	time.Time
}

func NewDate(year int, month time.Month, day int) *Date {
	return &Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func ParseDate(s string) (*Date, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &Date{t}, nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format("2006-01-02"))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" || s == "null" {
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = *parsed
	return nil
}

func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*d = Date{v}
		return nil
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = *parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Date", value)
	}
}

func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}

func (d *Date) String() string {
	if d == nil {
		return ""
	}
	return d.Format("2006-01-02")
}

// ===== 业务实体 =====
type Organization struct {
	TenantID      string    `json:"tenant_id" db:"tenant_id"`
	Code          string    `json:"code" db:"code"`
	ParentCode    *string   `json:"parent_code,omitempty" db:"parent_code"`
	Name          string    `json:"name" db:"name"`
	UnitType      string    `json:"unit_type" db:"unit_type"`
	Status        string    `json:"status" db:"status"`
	Level         int       `json:"level" db:"level"`
	Path          string    `json:"path" db:"path"`
	SortOrder     int       `json:"sort_order" db:"sort_order"`
	Description   string    `json:"description" db:"description"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	EffectiveDate *Date     `json:"effective_date,omitempty" db:"effective_date"`
	EndDate       *Date     `json:"end_date,omitempty" db:"end_date"`
	IsTemporal    bool      `json:"is_temporal" db:"is_temporal"`
	ChangeReason  *string   `json:"change_reason,omitempty" db:"change_reason"`
	IsCurrent     bool      `json:"is_current" db:"is_current"`
}

// ===== 请求/响应模型 =====
type CreateOrganizationRequest struct {
	Name          string  `json:"name"`
	UnitType      string  `json:"unit_type"`
	ParentCode    *string `json:"parent_code,omitempty"`
	SortOrder     int     `json:"sort_order"`
	Description   string  `json:"description"`
	EffectiveDate *Date   `json:"effective_date,omitempty"`
	EndDate       *Date   `json:"end_date,omitempty"`
	IsTemporal    bool    `json:"is_temporal"`
	ChangeReason  string  `json:"change_reason,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name          *string `json:"name,omitempty"`
	UnitType      *string `json:"unit_type,omitempty"`
	Status        *string `json:"status,omitempty"`
	SortOrder     *int    `json:"sort_order,omitempty"`
	Description   *string `json:"description,omitempty"`
	ParentCode    *string `json:"parent_code,omitempty"`
	EffectiveDate *Date   `json:"effective_date,omitempty"`
	EndDate       *Date   `json:"end_date,omitempty"`
	IsTemporal    *bool   `json:"is_temporal,omitempty"`
	ChangeReason  *string `json:"change_reason,omitempty"`
}

type OrganizationResponse struct {
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	UnitType      string    `json:"unit_type"`
	Status        string    `json:"status"`
	Level         int       `json:"level"`
	Path          string    `json:"path"`
	SortOrder     int       `json:"sort_order"`
	Description   string    `json:"description"`
	ParentCode    *string   `json:"parent_code,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	EffectiveDate *Date     `json:"effective_date,omitempty"`
	EndDate       *Date     `json:"end_date,omitempty"`
	IsTemporal    bool      `json:"is_temporal"`
	ChangeReason  *string   `json:"change_reason,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// ===== 数据库仓储 =====
type OrganizationRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewOrganizationRepository(db *sql.DB, logger *log.Logger) *OrganizationRepository {
	return &OrganizationRepository{db: db, logger: logger}
}

func (r *OrganizationRepository) GenerateCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	query := `
		SELECT COALESCE(MAX(CAST(code AS INTEGER)), 1000000) + 1 as next_code
		FROM organization_units 
		WHERE tenant_id = $1 AND code ~ '^[0-9]{7}$'
	`
	
	var nextCode int
	err := r.db.QueryRowContext(ctx, query, tenantID.String()).Scan(&nextCode)
	if err != nil {
		return "", fmt.Errorf("生成组织代码失败: %w", err)
	}
	
	return fmt.Sprintf("%07d", nextCode), nil
}

func (r *OrganizationRepository) Create(ctx context.Context, org *Organization) (*Organization, error) {
	query := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`
	
	var createdAt, updatedAt time.Time
	
	var effectiveDate *Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
	} else {
		now := time.Now()
		effectiveDate = NewDate(now.Year(), now.Month(), now.Day())
	}

	err := r.db.QueryRowContext(ctx, query,
		org.TenantID, org.Code, org.ParentCode, org.Name, org.UnitType, org.Status,
		org.Level, org.Path, org.SortOrder, org.Description, time.Now(), time.Now(),
		effectiveDate, org.EndDate, org.IsTemporal, org.ChangeReason,
	).Scan(&createdAt, &updatedAt)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503":
				return nil, fmt.Errorf("父组织不存在: %s", *org.ParentCode)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}
	
	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate
	
	return org, nil
}

func (r *OrganizationRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
	query := `
		SELECT tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, is_temporal, change_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
	`
	
	var org Organization
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在: %s", code)
		}
		return nil, fmt.Errorf("查询组织失败: %w", err)
	}
	
	return &org, nil
}

func (r *OrganizationRepository) CalculatePath(ctx context.Context, tenantID uuid.UUID, parentCode *string, code string) (string, int, error) {
	if parentCode == nil {
		return "/" + code, 1, nil
	}
	
	query := `SELECT path, level FROM organization_units WHERE tenant_id = $1 AND code = $2`
	
	var parentPath string
	var parentLevel int
	
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), *parentCode).Scan(&parentPath, &parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("父组织不存在: %s", *parentCode)
		}
		return "", 0, fmt.Errorf("查询父组织失败: %w", err)
	}
	
	path := parentPath + "/" + code
	level := parentLevel + 1
	
	return path, level, nil
}

// ===== HTTP处理器 =====
type OrganizationHandler struct {
	repo   *OrganizationRepository
	logger *log.Logger
}

func NewOrganizationHandler(repo *OrganizationRepository, logger *log.Logger) *OrganizationHandler {
	return &OrganizationHandler{repo: repo, logger: logger}
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	if err := h.validateCreateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)
	
	code, err := h.repo.GenerateCode(r.Context(), tenantID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "CODE_GENERATION_ERROR", "生成组织代码失败", err)
		return
	}

	path, level, err := h.repo.CalculatePath(r.Context(), tenantID, req.ParentCode, code)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "PARENT_ERROR", "父组织处理失败", err)
		return
	}

	now := time.Now()
	org := &Organization{
		TenantID:      tenantID.String(),
		Code:          code,
		ParentCode:    req.ParentCode,
		Name:          req.Name,
		UnitType:      req.UnitType,
		Status:        "ACTIVE",
		Level:         level,
		Path:          path,
		SortOrder:     req.SortOrder,
		Description:   req.Description,
		EffectiveDate: req.EffectiveDate,
		EndDate:       req.EndDate,
		IsTemporal:    req.IsTemporal,
		ChangeReason:  func() *string { if req.ChangeReason == "" { return nil } else { return &req.ChangeReason } }(),
	}

	if org.EffectiveDate == nil {
		today := NewDate(now.Year(), now.Month(), now.Day())
		org.EffectiveDate = today
	}

	createdOrg, err := h.repo.Create(r.Context(), org)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "CREATE_ERROR", "创建组织失败", err)
		return
	}

	response := h.toOrganizationResponse(createdOrg)
	
	h.logger.Printf("组织创建成功: %s - %s", response.Code, response.Name)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)

	org, err := h.repo.GetByCode(r.Context(), tenantID, code)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "组织不存在", err)
		return
	}

	response := h.toOrganizationResponse(org)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 验证方法
func (h *OrganizationHandler) validateCreateRequest(req *CreateOrganizationRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("组织名称不能为空")
	}
	
	if len(req.Name) > 100 {
		return fmt.Errorf("组织名称不能超过100个字符")
	}
	
	if req.UnitType == "" {
		return fmt.Errorf("组织类型不能为空")
	}
	
	validTypes := map[string]bool{
		"COMPANY": true, "DEPARTMENT": true, "COST_CENTER": true, "PROJECT_TEAM": true,
	}
	if !validTypes[req.UnitType] {
		return fmt.Errorf("无效的组织类型: %s", req.UnitType)
	}
	
	return nil
}

func (h *OrganizationHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return DefaultTenantID
	}
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.logger.Printf("无效的租户ID，使用默认值: %s", tenantIDStr)
		return DefaultTenantID
	}
	
	return tenantID
}

func (h *OrganizationHandler) toOrganizationResponse(org *Organization) *OrganizationResponse {
	return &OrganizationResponse{
		Code:          org.Code,
		Name:          org.Name,
		UnitType:      org.UnitType,
		Status:        org.Status,
		Level:         org.Level,
		Path:          org.Path,
		SortOrder:     org.SortOrder,
		Description:   org.Description,
		ParentCode:    org.ParentCode,
		CreatedAt:     org.CreatedAt,
		UpdatedAt:     org.UpdatedAt,
		EffectiveDate: org.EffectiveDate,
		EndDate:       org.EndDate,
		IsTemporal:    org.IsTemporal,
		ChangeReason:  org.ChangeReason,
	}
}

func (h *OrganizationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{
		Code:    code,
		Message: message,
	}
	
	if err != nil {
		errorResp.Error = err.Error()
		h.logger.Printf("错误响应 [%d %s]: %v", statusCode, code, err)
	}
	
	json.NewEncoder(w).Encode(errorResp)
}

// ===== 主程序 =====
func main() {
	logger := log.New(os.Stdout, "[简化命令服务] ", log.LstdFlags)

	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
	logger.Println("PostgreSQL连接成功")

	repo := NewOrganizationRepository(db, logger)
	handler := NewOrganizationHandler(repo, logger)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API路由
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/organization-units", func(r chi.Router) {
			r.Post("/", handler.CreateOrganization)
			r.Get("/{code}", handler.GetOrganization)
		})
	})

	// 简化的健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "organization-command-service",
			"version":   "dev-simplified",
			"timestamp": time.Now(),
		})
	})

	// 根路径信息
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":  "Cube Castle 组织命令服务 (开发版)",
			"version":  "dev-simplified",
			"status":   "running",
			"endpoints": map[string]string{
				"create": "POST /api/v1/organization-units",
				"get":    "GET /api/v1/organization-units/{code}",
				"health": "GET /health",
			},
		})
	})

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
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("正在关闭服务...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("服务关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 组织命令服务启动成功 - 端口 :%s", port)
	logger.Printf("📍 API端点: http://localhost:%s/api/v1/organization-units", port)
	logger.Printf("📍 健康检查: http://localhost:%s/health", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}

	logger.Println("服务已关闭")
}