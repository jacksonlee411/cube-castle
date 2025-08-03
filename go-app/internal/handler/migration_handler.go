package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// MigrationHandler 处理迁移相关的HTTP请求
type MigrationHandler struct{}

// NewMigrationHandler 创建迁移处理器
func NewMigrationHandler() *MigrationHandler {
	return &MigrationHandler{}
}

// MigrationGuide 迁移指南结构
type MigrationGuide struct {
	Title         string                   `json:"title"`
	Version       string                   `json:"version"`
	LastUpdated   time.Time               `json:"last_updated"`
	DeprecatedAPIs []DeprecatedAPI         `json:"deprecated_apis"`
	NewAPIs       []NewAPI                 `json:"new_apis"`
	Examples      []MigrationExample      `json:"examples"`
}

// DeprecatedAPI 废弃API信息
type DeprecatedAPI struct {
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	RemovalDate string    `json:"removal_date"`
	Replacement string    `json:"replacement"`
	Reason      string    `json:"reason"`
}

// NewAPI 新API信息
type NewAPI struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "command" 或 "query"
	Description string `json:"description"`
}

// MigrationExample 迁移示例
type MigrationExample struct {
	Title       string `json:"title"`
	OldRequest  string `json:"old_request"`
	NewRequest  string `json:"new_request"`
	Description string `json:"description"`
}

// GetMigrationGuide 处理 GET /api/v1/migration-guide
func (h *MigrationHandler) GetMigrationGuide() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guide := MigrationGuide{
			Title:       "员工管理API CQRS迁移指南",
			Version:     "v1.0.0",
			LastUpdated: time.Now(),
			DeprecatedAPIs: []DeprecatedAPI{
				{
					Method:      "GET",
					Path:        "/api/v1/employees",
					RemovalDate: "2024-12-31",
					Replacement: "/api/v1/queries/employees",
					Reason:      "迁移到CQRS查询架构",
				},
				{
					Method:      "POST",
					Path:        "/api/v1/employees",
					RemovalDate: "2024-12-31",
					Replacement: "/api/v1/commands/hire-employee",
					Reason:      "迁移到CQRS命令架构",
				},
				{
					Method:      "GET",
					Path:        "/api/v1/employees/{id}",
					RemovalDate: "2024-12-31",
					Replacement: "/api/v1/queries/employees/{id}",
					Reason:      "迁移到CQRS查询架构",
				},
				{
					Method:      "PUT",
					Path:        "/api/v1/employees/{id}",
					RemovalDate: "2024-12-31",
					Replacement: "/api/v1/commands/update-employee",
					Reason:      "迁移到CQRS命令架构",
				},
				{
					Method:      "DELETE",
					Path:        "/api/v1/employees/{id}",
					RemovalDate: "2024-12-31",
					Replacement: "/api/v1/commands/terminate-employee",
					Reason:      "迁移到CQRS命令架构",
				},
			},
			NewAPIs: []NewAPI{
				{
					Method:      "GET",
					Path:        "/api/v1/queries/employees",
					Type:        "query",
					Description: "查询员工列表，支持搜索和过滤",
				},
				{
					Method:      "GET",
					Path:        "/api/v1/queries/employees/{id}",
					Type:        "query",
					Description: "查询单个员工详细信息",
				},
				{
					Method:      "POST",
					Path:        "/api/v1/commands/hire-employee",
					Type:        "command",
					Description: "雇佣新员工命令",
				},
				{
					Method:      "PUT",
					Path:        "/api/v1/commands/update-employee",
					Type:        "command",
					Description: "更新员工信息命令",
				},
				{
					Method:      "POST",
					Path:        "/api/v1/commands/terminate-employee",
					Type:        "command",
					Description: "终止员工雇佣关系命令",
				},
			},
			Examples: []MigrationExample{
				{
					Title:       "获取员工列表",
					OldRequest:  "GET /api/v1/employees?search=张三&department=技术部",
					NewRequest:  "GET /api/v1/queries/employees?search=张三&department=技术部",
					Description: "查询端点路径从 /employees 迁移到 /queries/employees",
				},
				{
					Title:       "创建新员工",
					OldRequest:  "POST /api/v1/employees",
					NewRequest:  "POST /api/v1/commands/hire-employee",
					Description: "创建员工从REST端点迁移到雇佣员工命令",
				},
				{
					Title:       "更新员工信息",
					OldRequest:  "PUT /api/v1/employees/{id}",
					NewRequest:  "PUT /api/v1/commands/update-employee",
					Description: "更新操作从REST端点迁移到更新员工命令",
				},
				{
					Title:       "删除员工",
					OldRequest:  "DELETE /api/v1/employees/{id}",
					NewRequest:  "POST /api/v1/commands/terminate-employee",
					Description: "删除操作迁移到终止员工命令，更符合业务语义",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(guide)
	}
}