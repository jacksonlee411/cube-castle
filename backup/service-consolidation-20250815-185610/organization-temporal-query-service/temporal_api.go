/**
 * GraphQL时态查询API
 * 为组织架构查询服务添加时态管理能力
 */
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq"
)

// 时态组织单元类型
type TemporalOrganizationUnit struct {
	Code            string     `json:"code"`
	Name            string     `json:"name"`
	UnitType        string     `json:"unitType"`
	Status          string     `json:"status"`
	Level           int        `json:"level"`
	Path            string     `json:"path"`
	SortOrder       int        `json:"sortOrder"`
	Description     *string    `json:"description"`
	ParentCode      *string    `json:"parentCode"`
	EffectiveFrom   *time.Time `json:"effectiveFrom"`
	EffectiveTo     *time.Time `json:"effectiveTo"`
	IsTemporal      bool       `json:"isTemporal"`
	Version         int        `json:"version"`
	ChangeReason    *string    `json:"changeReason"`
	TemporalStatus  string     `json:"temporalStatus"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// 数据库连接
var db *sql.DB

// GraphQL类型定义
var temporalOrganizationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "TemporalOrganizationUnit",
	Fields: graphql.Fields{
		"code": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"name": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"unitType": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"status": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"level": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"path": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"sortOrder": &graphql.Field{
			Type: graphql.Int,
		},
		"description": &graphql.Field{
			Type: graphql.String,
		},
		"parentCode": &graphql.Field{
			Type: graphql.String,
		},
		"effectiveFrom": &graphql.Field{
			Type: graphql.DateTime,
		},
		"effectiveTo": &graphql.Field{
			Type: graphql.DateTime,
		},
		"isTemporal": &graphql.Field{
			Type: graphql.Boolean,
		},
		"version": &graphql.Field{
			Type: graphql.Int,
		},
		"changeReason": &graphql.Field{
			Type: graphql.String,
		},
		"temporalStatus": &graphql.Field{
			Type: graphql.String,
		},
		"createdAt": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
		},
		"updatedAt": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
		},
	},
})

// 时态查询函数
func getOrganizationsAsByDate(ctx context.Context, asOfDate *time.Time, includeInactive bool, limit int, offset int) ([]*TemporalOrganizationUnit, error) {
	query := `
		SELECT ou.code, ou.name, ou.unit_type, ou.status, ou.level, ou.path, ou.sort_order,
			   ou.description, ou.parent_code, ou.effective_from, ou.effective_to,
			   COALESCE(ou.version, 1) as version, ou.change_reason, ou.created_at, ou.updated_at,
			   COALESCE(ou.is_temporal, false) as is_temporal,
			   CASE 
					WHEN ou.effective_from IS NULL AND ou.effective_to IS NULL THEN 'always_active'
					WHEN ou.effective_from <= NOW() AND (ou.effective_to IS NULL OR ou.effective_to > NOW()) THEN 'currently_active'
					WHEN ou.effective_from > NOW() THEN 'future_active' 
					WHEN ou.effective_to <= NOW() THEN 'expired'
					ELSE 'unknown'
			   END as temporal_status
		FROM organization_units ou
		WHERE ($1 = true OR ou.status != 'INACTIVE')
		ORDER BY ou.level, ou.sort_order, ou.name
		LIMIT $2 OFFSET $3
	`
	args := []interface{}{includeInactive, limit, offset}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query organizations failed: %w", err)
	}
	defer rows.Close()

	var organizations []*TemporalOrganizationUnit
	for rows.Next() {
		org := &TemporalOrganizationUnit{}

		err := rows.Scan(
			&org.Code, &org.Name, &org.UnitType, &org.Status, &org.Level, &org.Path,
			&org.SortOrder, &org.Description, &org.ParentCode, &org.EffectiveFrom,
			&org.EffectiveTo, &org.Version, &org.ChangeReason, &org.CreatedAt,
			&org.UpdatedAt, &org.IsTemporal, &org.TemporalStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("scan organization failed: %w", err)
		}

		organizations = append(organizations, org)
	}

	return organizations, nil
}

// GraphQL查询字段
var queryFields = graphql.Fields{
	// 时态组织查询
	"organizations": &graphql.Field{
		Type: graphql.NewList(temporalOrganizationType),
		Args: graphql.FieldConfigArgument{
			"first": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				DefaultValue: 50,
			},
			"offset": &graphql.ArgumentConfig{
				Type:         graphql.Int,
				DefaultValue: 0,
			},
			"asOfDate": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"temporalMode": &graphql.ArgumentConfig{
				Type:         graphql.String,
				DefaultValue: "current",
			},
			"includeInactive": &graphql.ArgumentConfig{
				Type:         graphql.Boolean,
				DefaultValue: false,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			includeInactive := p.Args["includeInactive"].(bool)
			limit := p.Args["first"].(int)
			offset := p.Args["offset"].(int)

			var asOfDate *time.Time
			if asOfDateStr, ok := p.Args["asOfDate"].(string); ok && asOfDateStr != "" {
				if parsed, err := time.Parse(time.RFC3339, asOfDateStr); err == nil {
					asOfDate = &parsed
				}
			}

			return getOrganizationsAsByDate(p.Context, asOfDate, includeInactive, limit, offset)
		},
	},

	// 单个时态组织查询
	"organization": &graphql.Field{
		Type: temporalOrganizationType,
		Args: graphql.FieldConfigArgument{
			"code": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"asOfDate": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"temporalMode": &graphql.ArgumentConfig{
				Type:         graphql.String,
				DefaultValue: "current",
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			code := p.Args["code"].(string)

			query := `
				SELECT code, name, unit_type, status, level, path, sort_order,
					   description, parent_code, effective_from, effective_to,
					   COALESCE(version, 1) as version, change_reason, created_at, updated_at,
					   COALESCE(is_temporal, false) as is_temporal,
					   'currently_active' as temporal_status
				FROM organization_units
				WHERE code = $1
			`

			row := db.QueryRowContext(p.Context, query, code)
			org := &TemporalOrganizationUnit{}

			err := row.Scan(
				&org.Code, &org.Name, &org.UnitType, &org.Status, &org.Level, &org.Path,
				&org.SortOrder, &org.Description, &org.ParentCode, &org.EffectiveFrom,
				&org.EffectiveTo, &org.Version, &org.ChangeReason, &org.CreatedAt,
				&org.UpdatedAt, &org.IsTemporal, &org.TemporalStatus,
			)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, nil
				}
				return nil, fmt.Errorf("query single organization failed: %w", err)
			}

			return org, nil
		},
	},

	// 组织统计信息
	"organizationStats": &graphql.Field{
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name: "OrganizationStats",
			Fields: graphql.Fields{
				"totalCount": &graphql.Field{Type: graphql.Int},
				"byType": &graphql.Field{Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
					Name: "TypeCount",
					Fields: graphql.Fields{
						"unitType": &graphql.Field{Type: graphql.String},
						"count":    &graphql.Field{Type: graphql.Int},
					},
				}))},
				"byStatus": &graphql.Field{Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
					Name: "StatusCount",
					Fields: graphql.Fields{
						"status": &graphql.Field{Type: graphql.String},
						"count":  &graphql.Field{Type: graphql.Int},
					},
				}))},
			},
		}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 获取总数
			var totalCount int
			err := db.QueryRowContext(p.Context, "SELECT COUNT(*) FROM organization_units").Scan(&totalCount)
			if err != nil {
				return nil, fmt.Errorf("query total count failed: %w", err)
			}

			// 获取按类型统计
			typeQuery := `SELECT unit_type, COUNT(*) FROM organization_units GROUP BY unit_type`
			typeRows, err := db.QueryContext(p.Context, typeQuery)
			if err != nil {
				return nil, fmt.Errorf("query type stats failed: %w", err)
			}
			defer typeRows.Close()

			var byType []map[string]interface{}
			for typeRows.Next() {
				var unitType string
				var count int
				err := typeRows.Scan(&unitType, &count)
				if err != nil {
					return nil, fmt.Errorf("scan type stats failed: %w", err)
				}
				byType = append(byType, map[string]interface{}{
					"unitType": unitType,
					"count":    count,
				})
			}

			// 获取按状态统计
			statusQuery := `SELECT status, COUNT(*) FROM organization_units GROUP BY status`
			statusRows, err := db.QueryContext(p.Context, statusQuery)
			if err != nil {
				return nil, fmt.Errorf("query status stats failed: %w", err)
			}
			defer statusRows.Close()

			var byStatus []map[string]interface{}
			for statusRows.Next() {
				var status string
				var count int
				err := statusRows.Scan(&status, &count)
				if err != nil {
					return nil, fmt.Errorf("scan status stats failed: %w", err)
				}
				byStatus = append(byStatus, map[string]interface{}{
					"status": status,
					"count":  count,
				})
			}

			return map[string]interface{}{
				"totalCount": totalCount,
				"byType":     byType,
				"byStatus":   byStatus,
			}, nil
		},
	},
}

func main() {
	// 数据库连接
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=user password=password dbname=cubecastle sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// 创建GraphQL Schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: queryFields,
		}),
	})
	if err != nil {
		log.Fatal("Failed to create GraphQL schema:", err)
	}

	// 创建GraphQL处理器
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// 设置路由
	http.Handle("/graphql", h)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "temporal-graphql-api"})
	})

	log.Println("Temporal GraphQL API Server started on :8090")
	log.Println("GraphiQL UI available at http://localhost:8090/graphql")
	log.Fatal(http.ListenAndServe(":8090", nil))
}