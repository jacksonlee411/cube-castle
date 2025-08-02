package repositories

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
)

// Organization 组织模型
type Organization struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id"`
	Status       string                 `json:"status"`
	Profile      map[string]interface{} `json:"profile"`
	Level        int                    `json:"level"`
	EmployeeCount int                   `json:"employee_count"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// OrganizationCommandRepository PostgreSQL命令仓储接口
type OrganizationCommandRepository interface {
	// 创建组织
	CreateOrganization(ctx context.Context, org Organization) error
	
	// 更新组织
	UpdateOrganization(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, changes map[string]interface{}) error
	
	// 删除组织
	DeleteOrganization(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	
	// 移动组织
	MoveOrganization(ctx context.Context, id uuid.UUID, newParentID *uuid.UUID, tenantID uuid.UUID) error
	
	// 激活/停用组织
	SetOrganizationStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error
	
	// 批量更新组织
	BulkUpdateOrganizations(ctx context.Context, ids []uuid.UUID, tenantID uuid.UUID, changes map[string]interface{}) error
	
	// 事务操作支持
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// OrganizationQueryRepository Neo4j查询仓储接口
type OrganizationQueryRepository interface {
	// 获取单个组织
	GetOrganization(ctx context.Context, query queries.GetOrganizationQuery) (*Organization, error)
	
	// 组织列表查询
	ListOrganizations(ctx context.Context, query queries.ListOrganizationsQuery) ([]*Organization, *PaginationInfo, error)
	
	// 组织树查询
	GetOrganizationTree(ctx context.Context, query queries.GetOrganizationTreeQuery) ([]*Organization, error)
	
	// 组织统计查询
	GetOrganizationStats(ctx context.Context, query queries.GetOrganizationStatsQuery) (*OrganizationStats, error)
	
	// 搜索组织
	SearchOrganizations(ctx context.Context, query queries.SearchOrganizationsQuery) ([]*Organization, error)
	
	// 组织层级查询
	GetOrganizationHierarchy(ctx context.Context, targetID uuid.UUID, direction string, maxDepth int, tenantID uuid.UUID) ([]*Organization, error)
	
	// 组织路径查询
	GetOrganizationPath(ctx context.Context, fromID, toID uuid.UUID, tenantID uuid.UUID) ([]*Organization, error)
	
	// 获取同级组织
	GetSiblingOrganizations(ctx context.Context, unitID uuid.UUID, includeSelf bool, tenantID uuid.UUID) ([]*Organization, error)
	
	// 获取子组织
	GetChildOrganizations(ctx context.Context, parentID uuid.UUID, tenantID uuid.UUID) ([]*Organization, error)
	
	// 检查组织存在性
	OrganizationExists(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, error)
}

// 辅助结构体定义
type PaginationInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

type OrganizationStats struct {
	TotalOrganizations  int `json:"total_organizations"`
	ActiveOrganizations int `json:"active_organizations"`
	Companies           int `json:"companies"`
	Departments         int `json:"departments"`
	Teams               int `json:"teams"`
	MaxDepth            int `json:"max_depth"`
	TotalEmployees      int `json:"total_employees"`
}

