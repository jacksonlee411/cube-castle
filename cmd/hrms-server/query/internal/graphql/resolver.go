package graphql

import (
	"context"
	"fmt"
	"log"

	"cube-castle/cmd/hrms-server/query/internal/model"
	"cube-castle/internal/auth"
	"github.com/google/uuid"
	graphqlgo "github.com/graph-gophers/graphql-go"
	sharedconfig "cube-castle/shared/config"
)

type QueryRepository interface {
	GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *model.OrganizationFilter, pagination *model.PaginationInput) (*model.OrganizationConnection, error)
	GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error)
	GetOrganizationAtDate(ctx context.Context, tenantID uuid.UUID, code string, date string) (*model.Organization, error)
	GetOrganizationHistory(ctx context.Context, tenantID uuid.UUID, code string, fromDate string, toDate string) ([]model.Organization, error)
	GetOrganizationVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]model.Organization, error)
	GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*model.OrganizationStats, error)
	GetOrganizationHierarchy(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationHierarchyData, error)
	GetOrganizationSubtree(ctx context.Context, tenantID uuid.UUID, code string, maxDepth int) (*model.OrganizationHierarchyData, error)
	GetPositions(ctx context.Context, tenantID uuid.UUID, filter *model.PositionFilterInput, pagination *model.PaginationInput, sorting []model.PositionSortInput) (*model.PositionConnection, error)
	GetPositionByCode(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*model.Position, error)
	GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *model.PositionAssignmentFilterInput, pagination *model.PaginationInput, sorting []model.PositionAssignmentSortInput) (*model.PositionAssignmentConnection, error)
	GetPositionAssignmentAudit(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID *string, dateRange *model.DateRangeInput, pagination *model.PaginationInput) (*model.PositionAssignmentAuditConnection, error)
	GetPositionTimeline(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]model.PositionTimelineEntry, error)
	GetPositionVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]model.Position, error)
	GetVacantPositionConnection(ctx context.Context, tenantID uuid.UUID, filter *model.VacantPositionFilterInput, pagination *model.PaginationInput, sorting []model.VacantPositionSortInput) (*model.VacantPositionConnection, error)
	GetPositionTransfers(ctx context.Context, tenantID uuid.UUID, positionCode *string, organizationCode *string, pagination *model.PaginationInput) (*model.PositionTransferConnection, error)
	GetPositionHeadcountStats(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*model.HeadcountStats, error)
	GetJobFamilyGroups(ctx context.Context, tenantID uuid.UUID, includeInactive bool, asOfDate *string) ([]model.JobFamilyGroup, error)
	GetJobFamilies(ctx context.Context, tenantID uuid.UUID, groupCode string, includeInactive bool, asOfDate *string) ([]model.JobFamily, error)
	GetJobRoles(ctx context.Context, tenantID uuid.UUID, familyCode string, includeInactive bool, asOfDate *string) ([]model.JobRole, error)
	GetJobLevels(ctx context.Context, tenantID uuid.UUID, roleCode string, includeInactive bool, asOfDate *string) ([]model.JobLevel, error)
	GetAuditHistory(ctx context.Context, tenantID uuid.UUID, recordID string, startDate, endDate, operation, userID *string, limit int) ([]model.AuditRecordData, error)
	GetAuditLog(ctx context.Context, auditID string) (*model.AuditRecordData, error)
}

type PermissionChecker interface {
	CheckQueryPermission(ctx context.Context, queryName string) error
}

type Resolver struct {
	repo        QueryRepository
	logger      *log.Logger
	permissions PermissionChecker
}

func NewResolver(repo QueryRepository, logger *log.Logger, permissions PermissionChecker) *Resolver {
	return &Resolver{repo: repo, logger: logger, permissions: permissions}
}

// 当前组织列表查询 - 符合API契约v4.2.1 (camelCase方法名)
func (r *Resolver) Organizations(ctx context.Context, args struct {
	Filter     *model.OrganizationFilter
	Pagination *model.PaginationInput
}) (*model.OrganizationConnection, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "organizations"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "organization"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "organizationAtDate"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "organizationHistory"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationHistory: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 历史查询 - code: %s, range: %s~%s", args.Code, args.FromDate, args.ToDate)
	return r.repo.GetOrganizationHistory(ctx, sharedconfig.DefaultTenantID, args.Code, args.FromDate, args.ToDate)
}

// 组织版本查询 - 按计划实现，支持includeDeleted参数
func (r *Resolver) OrganizationVersions(ctx context.Context, args struct {
	Code           string
	IncludeDeleted graphqlgo.NullBool
}) ([]model.Organization, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "organizationVersions"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationVersions: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	includeDeleted := false
	if args.IncludeDeleted.Set && args.IncludeDeleted.Value != nil {
		includeDeleted = *args.IncludeDeleted.Value
	}

	r.logger.Printf("[GraphQL] 版本查询 - code: %s, includeDeleted: %v", args.Code, includeDeleted)
	return r.repo.GetOrganizationVersions(ctx, sharedconfig.DefaultTenantID, args.Code, includeDeleted)
}

// 组织统计 (camelCase方法名)
func (r *Resolver) OrganizationStats(ctx context.Context, args struct {
	AsOfDate          *string
	IncludeHistorical bool
}) (*model.OrganizationStats, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "organizationStats"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "organizationHierarchy"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "organizationSubtree"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "hierarchyStatistics"); err != nil {
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

// Positions 查询
func (r *Resolver) Positions(ctx context.Context, args struct {
	Filter     *model.PositionFilterInput
	Pagination *model.PaginationInput
	Sorting    *[]model.PositionSortInput
}) (*model.PositionConnection, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positions"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positions: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	var sorting []model.PositionSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}
	r.logger.Printf("[GraphQL] 查询职位列表 filter=%+v pagination=%+v sort=%d", args.Filter, args.Pagination, len(sorting))

	return r.repo.GetPositions(ctx, sharedconfig.DefaultTenantID, args.Filter, args.Pagination, sorting)
}

// Position 查询单个职位
func (r *Resolver) Position(ctx context.Context, args struct {
	Code     string
	AsOfDate *string
}) (*model.Position, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "position"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: position: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 查询职位详情 code=%s asOfDate=%v", args.Code, args.AsOfDate)

	return r.repo.GetPositionByCode(ctx, sharedconfig.DefaultTenantID, args.Code, args.AsOfDate)
}

// PositionAssignments 查询职位任职记录
func (r *Resolver) PositionAssignments(ctx context.Context, args struct {
	PositionCode string
	Filter       *model.PositionAssignmentFilterInput
	Pagination   *model.PaginationInput
	Sorting      *[]model.PositionAssignmentSortInput
}) (*model.PositionAssignmentConnection, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positionAssignments"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positionAssignments: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	var (
		tenantID = sharedconfig.DefaultTenantID
	)
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			r.logger.Printf("[AUTH] 非法租户ID: %s", tenantStr)
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	var sorting []model.PositionAssignmentSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}

	r.logger.Printf("[GraphQL] 查询职位任职 positionCode=%s filter=%+v pagination=%+v sort=%d tenant=%s",
		args.PositionCode, args.Filter, args.Pagination, len(sorting), tenantID.String())

	return r.repo.GetPositionAssignments(ctx, tenantID, args.PositionCode, args.Filter, args.Pagination, sorting)
}

func (r *Resolver) PositionAssignmentAudit(ctx context.Context, args struct {
	PositionCode string
	AssignmentId *string
	DateRange    *model.DateRangeInput
	Pagination   *model.PaginationInput
}) (*model.PositionAssignmentAuditConnection, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positionAssignmentAudit"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positionAssignmentAudit: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			r.logger.Printf("[AUTH] 非法租户ID: %s", tenantStr)
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	r.logger.Printf("[GraphQL] 查询任职审计 positionCode=%s assignmentId=%v", args.PositionCode, args.AssignmentId)
	return r.repo.GetPositionAssignmentAudit(ctx, tenantID, args.PositionCode, args.AssignmentId, args.DateRange, args.Pagination)
}

// PositionTimeline 查询职位时间线
func (r *Resolver) PositionTimeline(ctx context.Context, args struct {
	Code      string
	StartDate *string
	EndDate   *string
}) ([]model.PositionTimelineEntry, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positionTimeline"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positionTimeline: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 查询职位时间线 code=%s start=%v end=%v", args.Code, args.StartDate, args.EndDate)

	return r.repo.GetPositionTimeline(ctx, sharedconfig.DefaultTenantID, args.Code, args.StartDate, args.EndDate)
}

// PositionVersions 查询职位版本列表
func (r *Resolver) PositionVersions(ctx context.Context, args struct {
	Code           string
	IncludeDeleted graphqlgo.NullBool
}) ([]model.Position, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positionVersions"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positionVersions: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	includeDeleted := false
	if args.IncludeDeleted.Set && args.IncludeDeleted.Value != nil {
		includeDeleted = *args.IncludeDeleted.Value
	}

	r.logger.Printf("[GraphQL] 查询职位版本 code=%s includeDeleted=%v", args.Code, includeDeleted)

	return r.repo.GetPositionVersions(ctx, sharedconfig.DefaultTenantID, args.Code, includeDeleted)
}

// VacantPositions 查询空缺职位
func (r *Resolver) VacantPositions(ctx context.Context, args struct {
	Filter     *model.VacantPositionFilterInput
	Pagination *model.PaginationInput
	Sorting    *[]model.VacantPositionSortInput
}) (*model.VacantPositionConnection, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "vacantPositions"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: vacantPositions: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			r.logger.Printf("[AUTH] 无效租户ID: %s", tenantStr)
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	var sorting []model.VacantPositionSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}

	r.logger.Printf("[GraphQL] 查询空缺职位 filter=%+v pagination=%+v sort=%d tenant=%s",
		args.Filter, args.Pagination, len(sorting), tenantID.String())

	return r.repo.GetVacantPositionConnection(ctx, tenantID, args.Filter, args.Pagination, sorting)
}

// PositionTransfers 查询职位转移记录
func (r *Resolver) PositionTransfers(ctx context.Context, args struct {
	PositionCode     *string
	OrganizationCode *string
	Pagination       *model.PaginationInput
}) (*model.PositionTransferConnection, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positionTransfers"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positionTransfers: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			r.logger.Printf("[AUTH] 无效租户ID: %s", tenantStr)
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	r.logger.Printf("[GraphQL] 查询职位转移 positionCode=%v organizationCode=%v pagination=%+v tenant=%s",
		args.PositionCode, args.OrganizationCode, args.Pagination, tenantID.String())

	return r.repo.GetPositionTransfers(ctx, tenantID, args.PositionCode, args.OrganizationCode, args.Pagination)
}

// PositionHeadcountStats 查询编制统计
func (r *Resolver) PositionHeadcountStats(ctx context.Context, args struct {
	OrganizationCode    string
	IncludeSubordinates graphqlgo.NullBool
}) (*model.HeadcountStats, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "positionHeadcountStats"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: positionHeadcountStats: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	includeSubordinates := true
	if args.IncludeSubordinates.Set && args.IncludeSubordinates.Value != nil {
		includeSubordinates = *args.IncludeSubordinates.Value
	}
	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			r.logger.Printf("[AUTH] 无效租户ID: %s", tenantStr)
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}
	r.logger.Printf("[GraphQL] 查询职位编制统计 org=%s includeSub=%v tenant=%s", args.OrganizationCode, includeSubordinates, tenantID.String())

	return r.repo.GetPositionHeadcountStats(ctx, tenantID, args.OrganizationCode, includeSubordinates)
}

// JobFamilyGroups 查询职类
func (r *Resolver) JobFamilyGroups(ctx context.Context, args struct {
	IncludeInactive graphqlgo.NullBool
	AsOfDate        *string
}) ([]model.JobFamilyGroup, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "jobFamilyGroups"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: jobFamilyGroups: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	includeInactive := false
	if args.IncludeInactive.Set && args.IncludeInactive.Value != nil {
		includeInactive = *args.IncludeInactive.Value
	}
	r.logger.Printf("[GraphQL] 查询职类 includeInactive=%v asOf=%v", includeInactive, args.AsOfDate)

	return r.repo.GetJobFamilyGroups(ctx, sharedconfig.DefaultTenantID, includeInactive, args.AsOfDate)
}

// JobFamilies 查询职种
func (r *Resolver) JobFamilies(ctx context.Context, args struct {
	GroupCode       string
	IncludeInactive graphqlgo.NullBool
	AsOfDate        *string
}) ([]model.JobFamily, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "jobFamilies"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: jobFamilies: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	includeInactive := false
	if args.IncludeInactive.Set && args.IncludeInactive.Value != nil {
		includeInactive = *args.IncludeInactive.Value
	}
	r.logger.Printf("[GraphQL] 查询职种 group=%s includeInactive=%v asOf=%v", args.GroupCode, includeInactive, args.AsOfDate)

	return r.repo.GetJobFamilies(ctx, sharedconfig.DefaultTenantID, args.GroupCode, includeInactive, args.AsOfDate)
}

// JobRoles 查询职务
func (r *Resolver) JobRoles(ctx context.Context, args struct {
	FamilyCode      string
	IncludeInactive graphqlgo.NullBool
	AsOfDate        *string
}) ([]model.JobRole, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "jobRoles"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: jobRoles: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	includeInactive := false
	if args.IncludeInactive.Set && args.IncludeInactive.Value != nil {
		includeInactive = *args.IncludeInactive.Value
	}
	r.logger.Printf("[GraphQL] 查询职务 family=%s includeInactive=%v asOf=%v", args.FamilyCode, includeInactive, args.AsOfDate)

	return r.repo.GetJobRoles(ctx, sharedconfig.DefaultTenantID, args.FamilyCode, includeInactive, args.AsOfDate)
}

// JobLevels 查询职级
func (r *Resolver) JobLevels(ctx context.Context, args struct {
	RoleCode        string
	IncludeInactive graphqlgo.NullBool
	AsOfDate        *string
}) ([]model.JobLevel, error) {
	if err := r.permissions.CheckQueryPermission(ctx, "jobLevels"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: jobLevels: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	includeInactive := false
	if args.IncludeInactive.Set && args.IncludeInactive.Value != nil {
		includeInactive = *args.IncludeInactive.Value
	}
	r.logger.Printf("[GraphQL] 查询职级 role=%s includeInactive=%v asOf=%v", args.RoleCode, includeInactive, args.AsOfDate)

	return r.repo.GetJobLevels(ctx, sharedconfig.DefaultTenantID, args.RoleCode, includeInactive, args.AsOfDate)
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
	if err := r.permissions.CheckQueryPermission(ctx, "auditHistory"); err != nil {
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
	if err := r.permissions.CheckQueryPermission(ctx, "auditLog"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: auditLog: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	r.logger.Printf("[GraphQL] 单条审计记录查询 - auditId: %s", args.AuditId)
	return r.repo.GetAuditLog(ctx, args.AuditId)
}
