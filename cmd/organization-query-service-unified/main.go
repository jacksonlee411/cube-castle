package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq"
)

// 数据模型
type Organization struct {
	Code            string     `json:"code" db:"code"`
	ParentCode      *string    `json:"parentCode" db:"parent_code"`
	TenantID        string     `json:"tenantId" db:"tenant_id"`
	Name            string     `json:"name" db:"name"`
	UnitType        string     `json:"unitType" db:"unit_type"`
	Status          string     `json:"status" db:"status"`
	IsDeleted       bool       `json:"isDeleted" db:"is_deleted"`
	Level           int        `json:"level" db:"level"`
	HierarchyDepth  int        `json:"hierarchyDepth" db:"hierarchy_depth"`
	CodePath        string     `json:"codePath" db:"code_path"`
	NamePath        string     `json:"namePath" db:"name_path"`
	SortOrder       int        `json:"sortOrder" db:"sort_order"`
	Description     *string    `json:"description" db:"description"`
	Profile         string     `json:"profile" db:"profile"` // JSONB as string
	EffectiveDate   string     `json:"effectiveDate" db:"effective_date"`
	EndDate         *string    `json:"endDate" db:"end_date"`
	IsCurrent       bool       `json:"isCurrent" db:"is_current"`
	IsFuture        bool       `json:"isFuture" db:"is_future"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" db:"updated_at"`
	OperationType   string     `json:"operationType" db:"operation_type"`
	OperatedByID    string     `json:"-" db:"operated_by_id"`
	OperatedByName  string     `json:"-" db:"operated_by_name"`
	OperationReason *string    `json:"operationReason" db:"operation_reason"`
	RecordID        string     `json:"recordId" db:"record_id"`
}

type OperatedBy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (o *Organization) GetOperatedBy() *OperatedBy {
	return &OperatedBy{
		ID:   o.OperatedByID,
		Name: o.OperatedByName,
	}
}

type OrganizationStats struct {
	TotalCount      int `json:"totalCount"`
	ActiveCount     int `json:"activeCount"`
	InactiveCount   int `json:"inactiveCount"`
	DepartmentCount int `json:"departmentCount"`
	CompanyCount    int `json:"companyCount"`
	ProjectCount    int `json:"projectCount"`
}

// 数据库连接
var db *sql.DB

func init() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}
	
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	log.Println("✅ Database connected successfully")
}

// GraphQL Schema定义
func createSchema() (graphql.Schema, error) {
	// 操作人类型
	operatedByType := graphql.NewObject(graphql.ObjectConfig{
		Name: "OperatedBy",
		Fields: graphql.Fields{
			"id":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"name": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	// 组织类型
	organizationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Organization",
		Fields: graphql.Fields{
			"code":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"parentCode":      &graphql.Field{Type: graphql.String},
			"tenantId":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"name":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"unitType":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"status":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"isDeleted":       &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"level":           &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"hierarchyDepth":  &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"codePath":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"namePath":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"sortOrder":       &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"description":     &graphql.Field{Type: graphql.String},
			"profile":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"effectiveDate":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"endDate":         &graphql.Field{Type: graphql.String},
			"isCurrent":       &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"isFuture":        &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"createdAt":       &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
			"updatedAt":       &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
			"operationType":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"operatedBy":      &graphql.Field{Type: graphql.NewNonNull(operatedByType)},
			"operationReason": &graphql.Field{Type: graphql.String},
			"recordId":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	// 统计类型
	statsType := graphql.NewObject(graphql.ObjectConfig{
		Name: "OrganizationStats",
		Fields: graphql.Fields{
			"totalCount":      &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"activeCount":     &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"inactiveCount":   &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"departmentCount": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"companyCount":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"projectCount":    &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
		},
	})

	// 连接类型 (分页)
	connectionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "OrganizationConnection",
		Fields: graphql.Fields{
			"data": &graphql.Field{
				Type: graphql.NewList(organizationType),
			},
			"totalCount": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"hasMore":    &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		},
	})

	// 查询类型
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// 1. organizations - 基础分页查询
			"organizations": &graphql.Field{
				Type: connectionType,
				Args: graphql.FieldConfigArgument{
					"first":  &graphql.ArgumentConfig{Type: graphql.Int},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int},
					"filter": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					first := 10 // 默认
					if val, ok := p.Args["first"].(int); ok {
						first = val
					}
					
					offset := 0
					if val, ok := p.Args["offset"].(int); ok {
						offset = val
					}

					query := `
						SELECT code, parent_code, tenant_id, name, unit_type, status, is_deleted,
							   level, hierarchy_depth, code_path, name_path, sort_order,
							   description, profile::text, effective_date, end_date, is_current, is_future,
							   created_at, updated_at, operation_type, operated_by_id, operated_by_name,
							   operation_reason, record_id
						FROM organization_units 
						WHERE is_current = true AND NOT is_deleted
						ORDER BY code
						LIMIT $1 OFFSET $2
					`

					rows, err := db.Query(query, first, offset)
					if err != nil {
						return nil, err
					}
					defer rows.Close()

					var organizations []*Organization
					for rows.Next() {
						org := &Organization{}
						err := rows.Scan(
							&org.Code, &org.ParentCode, &org.TenantID, &org.Name, &org.UnitType,
							&org.Status, &org.IsDeleted, &org.Level, &org.HierarchyDepth,
							&org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
							&org.Profile, &org.EffectiveDate, &org.EndDate, &org.IsCurrent,
							&org.IsFuture, &org.CreatedAt, &org.UpdatedAt, &org.OperationType,
							&org.OperatedByID, &org.OperatedByName, &org.OperationReason, &org.RecordID,
						)
						if err != nil {
							return nil, err
						}
						organizations = append(organizations, org)
					}

					// 获取总数
					var totalCount int
					countQuery := "SELECT COUNT(*) FROM organization_units WHERE is_current = true AND NOT is_deleted"
					err = db.QueryRow(countQuery).Scan(&totalCount)
					if err != nil {
						return nil, err
					}

					return map[string]interface{}{
						"data":       organizations,
						"totalCount": totalCount,
						"hasMore":    offset+len(organizations) < totalCount,
					}, nil
				},
			},

			// 2. organization - 单记录查询
			"organization": &graphql.Field{
				Type: organizationType,
				Args: graphql.FieldConfigArgument{
					"code":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"asOfDate": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					code := p.Args["code"].(string)
					
					query := `
						SELECT code, parent_code, tenant_id, name, unit_type, status, is_deleted,
							   level, hierarchy_depth, code_path, name_path, sort_order,
							   description, profile::text, effective_date, end_date, is_current, is_future,
							   created_at, updated_at, operation_type, operated_by_id, operated_by_name,
							   operation_reason, record_id
						FROM organization_units 
						WHERE code = $1 AND is_current = true AND NOT is_deleted
						LIMIT 1
					`

					org := &Organization{}
					err := db.QueryRow(query, code).Scan(
						&org.Code, &org.ParentCode, &org.TenantID, &org.Name, &org.UnitType,
						&org.Status, &org.IsDeleted, &org.Level, &org.HierarchyDepth,
						&org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
						&org.Profile, &org.EffectiveDate, &org.EndDate, &org.IsCurrent,
						&org.IsFuture, &org.CreatedAt, &org.UpdatedAt, &org.OperationType,
						&org.OperatedByID, &org.OperatedByName, &org.OperationReason, &org.RecordID,
					)
					if err != nil {
						if err == sql.ErrNoRows {
							return nil, nil
						}
						return nil, err
					}

					return org, nil
				},
			},

			// 3. organizationStats - 统计查询
			"organizationStats": &graphql.Field{
				Type: statsType,
				Args: graphql.FieldConfigArgument{
					"asOfDate": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					query := `
						SELECT 
							COUNT(*) as total_count,
							COUNT(*) FILTER (WHERE status = 'ACTIVE') as active_count,
							COUNT(*) FILTER (WHERE status = 'INACTIVE') as inactive_count,
							COUNT(*) FILTER (WHERE unit_type = 'DEPARTMENT') as department_count,
							COUNT(*) FILTER (WHERE unit_type = 'COMPANY') as company_count,
							COUNT(*) FILTER (WHERE unit_type = 'PROJECT_TEAM') as project_count
						FROM organization_units 
						WHERE is_current = true AND NOT is_deleted
					`

					stats := &OrganizationStats{}
					err := db.QueryRow(query).Scan(
						&stats.TotalCount, &stats.ActiveCount, &stats.InactiveCount,
						&stats.DepartmentCount, &stats.CompanyCount, &stats.ProjectCount,
					)
					if err != nil {
						return nil, err
					}

					return stats, nil
				},
			},
		},
	})

	// 创建schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		return schema, err
	}

	return schema, nil
}

func main() {
	// 创建GraphQL schema
	schema, err := createSchema()
	if err != nil {
		log.Fatal("Failed to create GraphQL schema:", err)
	}

	// 创建Gin router
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "GraphQL Query Service",
			"version":   "v4.2.1",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// GraphQL端点
	graphQLHandler := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: false,
	})

	r.POST("/graphql", gin.WrapH(graphQLHandler))
	r.GET("/graphql", gin.WrapH(graphQLHandler))

	// GraphiQL界面 (开发模式)
	r.GET("/graphiql", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, graphiqlHTML)
	})

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("🚀 GraphQL Query Service启动在端口 %s", port)
	log.Printf("📊 GraphQL端点: http://localhost:%s/graphql", port)
	log.Printf("🔧 GraphiQL界面: http://localhost:%s/graphiql", port)
	log.Printf("🏥 健康检查: http://localhost:%s/health", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

const graphiqlHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>GraphiQL - Cube Castle API</title>
    <link href="https://unpkg.com/graphiql/graphiql.min.css" rel="stylesheet" />
</head>
<body style="margin: 0;">
    <div id="graphiql" style="height: 100vh;"></div>
    <script src="https://unpkg.com/react@17/umd/react.production.min.js"></script>
    <script src="https://unpkg.com/react-dom@17/umd/react-dom.production.min.js"></script>
    <script src="https://unpkg.com/graphiql/graphiql.min.js"></script>
    <script>
        const graphQLFetcher = (graphQLParams) => {
            return fetch('/graphql', {
                method: 'post',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(graphQLParams),
                credentials: 'include',
            }).then(response => response.json()).catch(() => response.text());
        };

        ReactDOM.render(
            React.createElement(GraphiQL, {
                fetcher: graphQLFetcher,
                defaultQuery: '# Cube Castle Organization API\n# Version: v4.2.1\n\nquery {\n  organizations(first: 5) {\n    data {\n      code\n      name\n      unitType\n      status\n      level\n      codePath\n    }\n    totalCount\n    hasMore\n  }\n}\n\n# 统计查询示例\n# query {\n#   organizationStats {\n#     totalCount\n#     activeCount\n#     departmentCount\n#   }\n# }\n\n# 单记录查询示例\n# query {\n#   organization(code: "1000001") {\n#     code\n#     name\n#     unitType\n#     profile\n#     operatedBy {\n#       id\n#       name\n#     }\n#   }\n# }',
            }),
            document.getElementById('graphiql')
        );
    </script>
</body>
</html>
`