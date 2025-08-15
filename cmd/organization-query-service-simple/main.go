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

// ===== GraphQL相关类型 =====
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message   string                 `json:"message"`
	Locations []GraphQLErrorLocation `json:"locations,omitempty"`
	Path      []interface{}          `json:"path,omitempty"`
}

type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
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

type OrganizationStats struct {
	TotalCount      int `json:"total_count"`
	ActiveCount     int `json:"active_count"`
	InactiveCount   int `json:"inactive_count"`
	CompanyCount    int `json:"company_count"`
	DepartmentCount int `json:"department_count"`
}

// ===== 数据库仓储 =====
type OrganizationQueryRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewOrganizationQueryRepository(db *sql.DB, logger *log.Logger) *OrganizationQueryRepository {
	return &OrganizationQueryRepository{db: db, logger: logger}
}

func (r *OrganizationQueryRepository) GetAll(ctx context.Context, tenantID uuid.UUID) ([]Organization, error) {
	query := `
		SELECT tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, is_temporal, change_reason, is_current
		FROM organization_units 
		WHERE tenant_id = $1
		ORDER BY level ASC, sort_order ASC, code ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("查询组织列表失败: %w", err)
	}
	defer rows.Close()
	
	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(
			&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
			&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
			&org.Description, &org.CreatedAt, &org.UpdatedAt,
			&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason, &org.IsCurrent,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描组织数据失败: %w", err)
		}
		organizations = append(organizations, org)
	}
	
	return organizations, nil
}

func (r *OrganizationQueryRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
	query := `
		SELECT tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, is_temporal, change_reason, is_current
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
	`
	
	var org Organization
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason, &org.IsCurrent,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 不存在
		}
		return nil, fmt.Errorf("查询组织失败: %w", err)
	}
	
	return &org, nil
}

func (r *OrganizationQueryRepository) GetStats(ctx context.Context, tenantID uuid.UUID) (*OrganizationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active_count,
			COUNT(CASE WHEN status = 'INACTIVE' THEN 1 END) as inactive_count,
			COUNT(CASE WHEN unit_type = 'COMPANY' THEN 1 END) as company_count,
			COUNT(CASE WHEN unit_type = 'DEPARTMENT' THEN 1 END) as department_count
		FROM organization_units 
		WHERE tenant_id = $1
	`
	
	var stats OrganizationStats
	err := r.db.QueryRowContext(ctx, query, tenantID.String()).Scan(
		&stats.TotalCount, &stats.ActiveCount, &stats.InactiveCount,
		&stats.CompanyCount, &stats.DepartmentCount,
	)
	
	if err != nil {
		return nil, fmt.Errorf("查询组织统计失败: %w", err)
	}
	
	return &stats, nil
}

// ===== GraphQL处理器 =====
type GraphQLHandler struct {
	repo   *OrganizationQueryRepository
	logger *log.Logger
}

func NewGraphQLHandler(repo *OrganizationQueryRepository, logger *log.Logger) *GraphQLHandler {
	return &GraphQLHandler{repo: repo, logger: logger}
}

func (h *GraphQLHandler) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "Invalid JSON request", nil)
		return
	}

	// 简化的GraphQL查询解析
	tenantID := h.getTenantID(r)
	
	// 处理不同的查询类型
	if strings.Contains(req.Query, "organizations") && !strings.Contains(req.Query, "organization(") {
		// 查询所有组织
		organizations, err := h.repo.GetAll(r.Context(), tenantID)
		if err != nil {
			h.writeErrorResponse(w, "查询组织列表失败", err)
			return
		}
		
		response := GraphQLResponse{
			Data: map[string]interface{}{
				"organizations": organizations,
			},
		}
		h.writeResponse(w, response)
		
	} else if strings.Contains(req.Query, "organization(") {
		// 根据code查询单个组织
		code := h.extractCodeFromQuery(req.Query, req.Variables)
		if code == "" {
			h.writeErrorResponse(w, "缺少组织代码参数", nil)
			return
		}
		
		org, err := h.repo.GetByCode(r.Context(), tenantID, code)
		if err != nil {
			h.writeErrorResponse(w, "查询组织失败", err)
			return
		}
		
		response := GraphQLResponse{
			Data: map[string]interface{}{
				"organization": org,
			},
		}
		h.writeResponse(w, response)
		
	} else if strings.Contains(req.Query, "organizationStats") {
		// 查询组织统计
		stats, err := h.repo.GetStats(r.Context(), tenantID)
		if err != nil {
			h.writeErrorResponse(w, "查询组织统计失败", err)
			return
		}
		
		response := GraphQLResponse{
			Data: map[string]interface{}{
				"organizationStats": stats,
			},
		}
		h.writeResponse(w, response)
		
	} else {
		h.writeErrorResponse(w, "不支持的查询类型", nil)
	}
}

func (h *GraphQLHandler) extractCodeFromQuery(query string, variables map[string]interface{}) string {
	// 简单解析GraphQL查询中的code参数
	if variables != nil {
		if code, ok := variables["code"].(string); ok {
			return code
		}
	}
	
	// 从查询字符串中提取code (简化版本)
	if strings.Contains(query, "code:") {
		parts := strings.Split(query, "code:")
		if len(parts) > 1 {
			codePart := strings.TrimSpace(parts[1])
			codePart = strings.Split(codePart, ")")[0]
			codePart = strings.Trim(codePart, "\" ")
			return codePart
		}
	}
	
	return ""
}

func (h *GraphQLHandler) getTenantID(r *http.Request) uuid.UUID {
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

func (h *GraphQLHandler) writeResponse(w http.ResponseWriter, response GraphQLResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *GraphQLHandler) writeErrorResponse(w http.ResponseWriter, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	
	graphqlErr := GraphQLError{
		Message: message,
	}
	
	if err != nil {
		graphqlErr.Message = fmt.Sprintf("%s: %v", message, err)
		h.logger.Printf("GraphQL错误: %v", err)
	}
	
	response := GraphQLResponse{
		Errors: []GraphQLError{graphqlErr},
	}
	
	json.NewEncoder(w).Encode(response)
}

// ===== 主程序 =====
func main() {
	logger := log.New(os.Stdout, "[简化查询服务] ", log.LstdFlags)

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

	repo := NewOrganizationQueryRepository(db, logger)
	handler := NewGraphQLHandler(repo, logger)

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

	// GraphQL路由
	r.Post("/graphql", handler.HandleGraphQL)
	
	// GraphiQL界面 (开发环境)
	r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GraphiQL - Cube Castle</title>
    <style>
        body { margin: 0; height: 100vh; overflow: hidden; }
        #graphiql { height: 100vh; }
    </style>
    <script crossorigin src="https://unpkg.com/react@17/umd/react.development.js"></script>
    <script crossorigin src="https://unpkg.com/react-dom@17/umd/react-dom.development.js"></script>
    <script src="https://unpkg.com/graphiql@1.4.7/graphiql.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/graphiql@1.4.7/graphiql.min.css" />
</head>
<body>
    <div id="graphiql"></div>
    <script>
        function graphQLFetcher(graphQLParams) {
            return fetch('/graphql', {
                method: 'post',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(graphQLParams),
            }).then(function (response) {
                return response.text();
            }).then(function (responseBody) {
                try {
                    return JSON.parse(responseBody);
                } catch (error) {
                    return responseBody;
                }
            });
        }

        ReactDOM.render(
            React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
            document.getElementById('graphiql')
        );
    </script>
</body>
</html>
		`)
	})

	// 简化的健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "organization-query-service",
			"version":   "dev-simplified",
			"timestamp": time.Now(),
		})
	})

	// 根路径信息
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":  "Cube Castle 组织查询服务 (开发版)",
			"version":  "dev-simplified",
			"status":   "running",
			"endpoints": map[string]string{
				"graphql":  "POST /graphql",
				"graphiql": "GET /graphiql",
				"health":   "GET /health",
			},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
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

	logger.Printf("🚀 组织查询服务启动成功 - 端口 :%s", port)
	logger.Printf("📍 GraphQL端点: http://localhost:%s/graphql", port)
	logger.Printf("📍 GraphiQL界面: http://localhost:%s/graphiql", port)
	logger.Printf("📍 健康检查: http://localhost:%s/health", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}

	logger.Println("服务已关闭")
}