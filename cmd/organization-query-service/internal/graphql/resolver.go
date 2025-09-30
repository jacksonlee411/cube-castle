package graphql

import (
	"context"
	"fmt"
	"log"

	"cube-castle-deployment-test/cmd/organization-query-service/internal/model"
	"cube-castle-deployment-test/cmd/organization-query-service/internal/repository"
	"cube-castle-deployment-test/internal/auth"
	sharedconfig "shared/config"
	"github.com/google/uuid"
)

type Resolver struct {
	repo   *repository.PostgreSQLRepository
	logger *log.Logger
	authMW *auth.GraphQLPermissionMiddleware
}

func NewResolver(repo *repository.PostgreSQLRepository, logger *log.Logger, authMW *auth.GraphQLPermissionMiddleware) *Resolver {
	return &Resolver{repo: repo, logger: logger, authMW: authMW}
}

// 当前组织列表查询 - 符合API契约v4.2.1 (camelCase方法名)
func (r *Resolver) Organizations(ctx context.Context, args struct {
	Filter     *model.OrganizationFilter
	Pagination *model.PaginationInput
}) (*model.OrganizationConnection, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizations"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizations: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 查询组织列表 - API契约v4.2.1")

	// 记录查询参数用于调试
	if args.Filter != nil {
		r.logger.Printf("[GraphQL] 过滤条件: %+v", *args.Filter)
	}
	if args.Pagination != nil {
		r.logger.Printf("[GraphQL] 分页参数: %+v", *args.Pagination)
	}

	return r.repo.GetOrganizations(ctx, sharedconfig.DefaultTenantID, args.Filter, args.Pagination)
}

// 单个组织查询
func (r *Resolver) Organization(ctx context.Context, args struct {
	Code     string
	AsOfDate *string
}) (*model.Organization, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organization"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organization: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 查询单个组织 - code: %s", args.Code)
	return r.repo.GetOrganization(ctx, sharedconfig.DefaultTenantID, args.Code)
}

// 时态查询 - 时间点
func (r *Resolver) OrganizationAtDate(ctx context.Context, args struct {
	Code string
	Date string
}) (*model.Organization, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationAtDate"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationAtDate: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 时态查询 - code: %s, date: %s", args.Code, args.Date)
	return r.repo.GetOrganizationAtDate(ctx, sharedconfig.DefaultTenantID, args.Code, args.Date)
}

// 时态查询 - 历史范围
func (r *Resolver) OrganizationHistory(ctx context.Context, args struct {
	Code     string
	FromDate string
	ToDate   string
}) ([]model.Organization, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationHistory"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationHistory: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 历史查询 - code: %s, range: %s~%s", args.Code, args.FromDate, args.ToDate)
	return r.repo.GetOrganizationHistory(ctx, sharedconfig.DefaultTenantID, args.Code, args.FromDate, args.ToDate)
}

// 组织版本查询 - 按计划实现，支持includeDeleted参数
func (r *Resolver) OrganizationVersions(ctx context.Context, args struct {
	Code           string
	IncludeDeleted *bool
}) ([]model.Organization, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationVersions"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationVersions: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	includeDeleted := false
	if args.IncludeDeleted != nil {
		includeDeleted = *args.IncludeDeleted
	}

	r.logger.Printf("[GraphQL] 版本查询 - code: %s, includeDeleted: %v", args.Code, includeDeleted)
	return r.repo.GetOrganizationVersions(ctx, sharedconfig.DefaultTenantID, args.Code, includeDeleted)
}

// 组织统计 (camelCase方法名)
func (r *Resolver) OrganizationStats(ctx context.Context, args struct {
	AsOfDate          *string
	IncludeHistorical bool
}) (*model.OrganizationStats, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationStats"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationStats: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 统计查询")
	return r.repo.GetOrganizationStats(ctx, sharedconfig.DefaultTenantID)
}

// 高级层级结构查询 - 严格遵循API规范v4.2.1
func (r *Resolver) OrganizationHierarchy(ctx context.Context, args struct {
	Code     string
	TenantId string
}) (*model.OrganizationHierarchyData, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationHierarchy"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationHierarchy: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 层级结构查询 - code: %s, tenantId: %s", args.Code, args.TenantId)

	tenantID, err := uuid.Parse(args.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	return r.repo.GetOrganizationHierarchy(ctx, tenantID, args.Code)
}

func (r *Resolver) OrganizationSubtree(ctx context.Context, args struct {
	Code            string
	TenantId        string
	MaxDepth        int32
	IncludeInactive bool
}) ([]model.OrganizationHierarchyData, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationSubtree"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationSubtree: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 子树查询 - code: %s, tenantId: %s, maxDepth: %v", args.Code, args.TenantId, args.MaxDepth)

	tenantID, err := uuid.Parse(args.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	maxDepth := 10 // 默认深度
	if args.MaxDepth > 0 {
		maxDepth = int(args.MaxDepth)
	}

	subtree, err := r.repo.GetOrganizationSubtree(ctx, tenantID, args.Code, maxDepth)
	if err != nil {
		return nil, err
	}

	// 将单个子树转换为数组（Schema期望数组返回）
	if subtree == nil {
		return []model.OrganizationHierarchyData{}, nil
	}

	// 先转换根节点
	root := model.OrganizationHierarchyData{
		CodeField:           subtree.CodeField,
		NameField:           subtree.NameField,
		LevelField:          subtree.LevelField,
		HierarchyDepthField: subtree.HierarchyDepthField,
		CodePathField:       subtree.CodePathField,
		NamePathField:       subtree.NamePathField,
		ParentChainField:    []string{}, // 根节点没有父级链
		ChildrenCountField:  len(subtree.ChildrenField),
		IsRootField:         subtree.LevelField == 1,
		IsLeafField:         len(subtree.ChildrenField) == 0,
		ChildrenField:       []model.OrganizationHierarchyData{}, // 简化实现，先不递归转换
	}

	return []model.OrganizationHierarchyData{root}, nil
}

// 层级统计查询
func (r *Resolver) HierarchyStatistics(ctx context.Context, args struct {
	TenantId              string
	IncludeIntegrityCheck bool
}) (*model.HierarchyStatistics, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "hierarchyStatistics"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: hierarchyStatistics: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	// TODO: 实现实际的层级统计逻辑
	return &model.HierarchyStatistics{
		TenantIdField:           args.TenantId,
		TotalOrganizationsField: 0,
		MaxDepthField:           0,
		AvgDepthField:           0.0,
		DepthDistributionField:  []model.DepthDistribution{},
		RootOrganizationsField:  0,
		LeafOrganizationsField:  0,
		IntegrityIssuesField:    []model.IntegrityIssue{},
		LastAnalyzedField:       "",
	}, nil
}

// 审计历史查询 - v4.6.0 基于record_id
func (r *Resolver) AuditHistory(ctx context.Context, args struct {
	RecordId  string
	StartDate *string
	EndDate   *string
	Operation *string
	UserId    *string
	Limit     int32
}) ([]model.AuditRecordData, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "auditHistory"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: auditHistory: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 审计历史查询 - recordId: %s", args.RecordId)

	limit := int32(50) // 默认限制
	if args.Limit > 0 {
		limit = args.Limit
		if limit > 200 { // API规范限制最大200
			limit = 200
		}
	}

	// 从上下文获取租户ID，强制租户隔离
	tenantStr := auth.GetTenantID(ctx)
	if tenantStr == "" {
		r.logger.Printf("[AUTH] 缺少租户ID，拒绝审计历史查询")
		return nil, fmt.Errorf("TENANT_REQUIRED")
	}
	tenantUUID, err := uuid.Parse(tenantStr)
	if err != nil {
		r.logger.Printf("[AUTH] 无效租户ID: %s", tenantStr)
		return nil, fmt.Errorf("INVALID_TENANT")
	}

	return r.repo.GetAuditHistory(ctx, tenantUUID, args.RecordId, args.StartDate, args.EndDate, args.Operation, args.UserId, int(limit))
}

// 单条审计记录查询 - v4.6.0
func (r *Resolver) AuditLog(ctx context.Context, args struct {
	AuditId string
}) (*model.AuditRecordData, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "auditLog"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: auditLog: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 单条审计记录查询 - auditId: %s", args.AuditId)
	return r.repo.GetAuditLog(ctx, args.AuditId)
}
