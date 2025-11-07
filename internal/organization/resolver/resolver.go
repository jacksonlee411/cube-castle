package resolver

import (
	"context"
	"fmt"
	"strings"

	"cube-castle/internal/auth"
	"cube-castle/internal/organization/dto"
	pkglogger "cube-castle/pkg/logger"
	sharedconfig "cube-castle/shared/config"
	"github.com/google/uuid"
)

type QueryRepository interface {
	GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *dto.OrganizationFilter, pagination *dto.PaginationInput) (*dto.OrganizationConnection, error)
	GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*dto.Organization, error)
	GetOrganizationAtDate(ctx context.Context, tenantID uuid.UUID, code string, date string) (*dto.Organization, error)
	GetOrganizationHistory(ctx context.Context, tenantID uuid.UUID, code string, fromDate string, toDate string) ([]dto.Organization, error)
	GetOrganizationVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]dto.Organization, error)
	GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*dto.OrganizationStats, error)
	GetOrganizationHierarchy(ctx context.Context, tenantID uuid.UUID, code string) (*dto.OrganizationHierarchyData, error)
	GetOrganizationSubtree(ctx context.Context, tenantID uuid.UUID, code string, maxDepth int) (*dto.OrganizationHierarchyData, error)
	GetPositions(ctx context.Context, tenantID uuid.UUID, filter *dto.PositionFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionSortInput) (*dto.PositionConnection, error)
	GetPositionByCode(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*dto.Position, error)
	GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetPositionAssignmentAudit(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID *string, dateRange *dto.DateRangeInput, pagination *dto.PaginationInput) (*dto.PositionAssignmentAuditConnection, error)
	GetPositionTimeline(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]dto.PositionTimelineEntry, error)
	GetPositionVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]dto.Position, error)
	GetVacantPositionConnection(ctx context.Context, tenantID uuid.UUID, filter *dto.VacantPositionFilterInput, pagination *dto.PaginationInput, sorting []dto.VacantPositionSortInput) (*dto.VacantPositionConnection, error)
	GetPositionTransfers(ctx context.Context, tenantID uuid.UUID, positionCode *string, organizationCode *string, pagination *dto.PaginationInput) (*dto.PositionTransferConnection, error)
	GetPositionHeadcountStats(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*dto.HeadcountStats, error)
	GetJobFamilyGroups(ctx context.Context, tenantID uuid.UUID, includeInactive bool, asOfDate *string) ([]dto.JobFamilyGroup, error)
	GetJobFamilies(ctx context.Context, tenantID uuid.UUID, groupCode string, includeInactive bool, asOfDate *string) ([]dto.JobFamily, error)
	GetJobRoles(ctx context.Context, tenantID uuid.UUID, familyCode string, includeInactive bool, asOfDate *string) ([]dto.JobRole, error)
	GetJobLevels(ctx context.Context, tenantID uuid.UUID, roleCode string, includeInactive bool, asOfDate *string) ([]dto.JobLevel, error)
	GetAuditHistory(ctx context.Context, tenantID uuid.UUID, recordID string, startDate, endDate, operation, userID *string, limit int) ([]dto.AuditRecordData, error)
	GetAuditLog(ctx context.Context, auditID string) (*dto.AuditRecordData, error)
	GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
}

type AssignmentProvider interface {
	GetAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
}

type PermissionChecker interface {
	CheckQueryPermission(ctx context.Context, queryName string) error
}

type Resolver struct {
	repo         QueryRepository
	logger       pkglogger.Logger
	permissions  PermissionChecker
	assignFacade AssignmentProvider
}

func NewResolver(repo QueryRepository, logger pkglogger.Logger, permissions PermissionChecker) *Resolver {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	return &Resolver{
		repo: repo,
		logger: logger.WithFields(pkglogger.Fields{
			"component": "query-resolver",
		}),
		permissions: permissions,
	}
}

func NewResolverWithAssignments(repo QueryRepository, assignments AssignmentProvider, logger pkglogger.Logger, permissions PermissionChecker) *Resolver {
	res := NewResolver(repo, logger, permissions)
	res.assignFacade = assignments
	return res
}

func (r *Resolver) loggerFor(resolverName, operation string, fields pkglogger.Fields) pkglogger.Logger {
	log := r.logger
	if resolverName != "" {
		log = log.WithFields(pkglogger.Fields{"resolver": resolverName})
	}
	if operation != "" {
		log = log.WithFields(pkglogger.Fields{"operation": operation})
	}
	if len(fields) > 0 {
		log = log.WithFields(fields)
	}
	return log
}

func (r *Resolver) authorize(ctx context.Context, queryName string, log pkglogger.Logger) error {
	if err := r.permissions.CheckQueryPermission(ctx, queryName); err != nil {
		log.WithFields(pkglogger.Fields{"error": err}).Warn("permission denied")
		return fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}
	return nil
}

// 当前组织列表查询 - 符合API契约v4.2.1 (camelCase方法名)
func (r *Resolver) Organizations(ctx context.Context, args struct {
	Filter     *dto.OrganizationFilter
	Pagination *dto.PaginationInput
}) (*dto.OrganizationConnection, error) {
	log := r.loggerFor("organizations", "list", pkglogger.Fields{
		"tenantId": sharedconfig.DefaultTenantID.String(),
	})
	if err := r.authorize(ctx, "organizations", log); err != nil {
		return nil, err
	}
	log.Info("处理组织列表查询")

	if args.Filter != nil {
		log.WithFields(pkglogger.Fields{"filter": args.Filter}).Info("附带过滤参数")
	}
	if args.Pagination != nil {
		log.WithFields(pkglogger.Fields{"pagination": args.Pagination}).Info("附带分页参数")
	}

	return r.repo.GetOrganizations(ctx, sharedconfig.DefaultTenantID, args.Filter, args.Pagination)
}

// 单个组织查询
func (r *Resolver) Organization(ctx context.Context, args struct {
	Code     string
	AsOfDate *string
}) (*dto.Organization, error) {
	log := r.loggerFor("organization", "get", pkglogger.Fields{
		"tenantId": sharedconfig.DefaultTenantID.String(),
		"code":     args.Code,
	})
	if err := r.authorize(ctx, "organization", log); err != nil {
		return nil, err
	}
	log.Info("查询单个组织")
	return r.repo.GetOrganization(ctx, sharedconfig.DefaultTenantID, args.Code)
}

// 时态查询 - 时间点
func (r *Resolver) OrganizationAtDate(ctx context.Context, args struct {
	Code string
	Date string
}) (*dto.Organization, error) {
	log := r.loggerFor("organization", "temporal", pkglogger.Fields{
		"tenantId": sharedconfig.DefaultTenantID.String(),
		"code":     args.Code,
		"date":     args.Date,
	})
	if err := r.authorize(ctx, "organizationAtDate", log); err != nil {
		return nil, err
	}
	log.Info("执行组织时态查询")
	return r.repo.GetOrganizationAtDate(ctx, sharedconfig.DefaultTenantID, args.Code, args.Date)
}

// 时态查询 - 历史范围
func (r *Resolver) OrganizationHistory(ctx context.Context, args struct {
	Code     string
	FromDate string
	ToDate   string
}) ([]dto.Organization, error) {
	log := r.loggerFor("organization", "history", pkglogger.Fields{
		"tenantId": sharedconfig.DefaultTenantID.String(),
		"code":     args.Code,
		"from":     args.FromDate,
		"to":       args.ToDate,
	})
	if err := r.authorize(ctx, "organizationHistory", log); err != nil {
		return nil, err
	}
	log.Info("执行组织历史查询")
	return r.repo.GetOrganizationHistory(ctx, sharedconfig.DefaultTenantID, args.Code, args.FromDate, args.ToDate)
}

// 组织版本查询 - 按计划实现，支持includeDeleted参数
func (r *Resolver) OrganizationVersions(ctx context.Context, args struct {
	Code           string
	IncludeDeleted *bool
}) ([]dto.Organization, error) {
	includeDeleted := false
	if args.IncludeDeleted != nil {
		includeDeleted = *args.IncludeDeleted
	}
	log := r.loggerFor("organization", "versions", pkglogger.Fields{
		"tenantId":       sharedconfig.DefaultTenantID.String(),
		"code":           args.Code,
		"includeDeleted": includeDeleted,
	})
	if err := r.authorize(ctx, "organizationVersions", log); err != nil {
		return nil, err
	}
	log.Info("执行组织版本查询")
	return r.repo.GetOrganizationVersions(ctx, sharedconfig.DefaultTenantID, args.Code, includeDeleted)
}

// 组织统计 (camelCase方法名)
func (r *Resolver) OrganizationStats(ctx context.Context, args struct {
	AsOfDate          *string
	IncludeHistorical bool
}) (*dto.OrganizationStats, error) {
	log := r.loggerFor("organization", "stats", pkglogger.Fields{
		"tenantId": sharedconfig.DefaultTenantID.String(),
	})
	if err := r.authorize(ctx, "organizationStats", log); err != nil {
		return nil, err
	}
	log.Info("执行组织统计查询")
	return r.repo.GetOrganizationStats(ctx, sharedconfig.DefaultTenantID)
}

// 高级层级结构查询 - 严格遵循API规范v4.2.1
func (r *Resolver) OrganizationHierarchy(ctx context.Context, args struct {
	Code     string
	TenantId string
}) (*dto.OrganizationHierarchyData, error) {
	log := r.loggerFor("organization", "hierarchy", pkglogger.Fields{
		"code": args.Code,
	})
	if err := r.authorize(ctx, "organizationHierarchy", log); err != nil {
		return nil, err
	}
	log.WithFields(pkglogger.Fields{"tenantId": args.TenantId}).Info("执行组织层级查询")

	tenantID, err := uuid.Parse(args.TenantId)
	if err != nil {
		log.WithFields(pkglogger.Fields{"error": err}).Warn("invalid tenant ID")
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	return r.repo.GetOrganizationHierarchy(ctx, tenantID, args.Code)
}

func (r *Resolver) OrganizationSubtree(ctx context.Context, args struct {
	Code            string
	TenantId        string
	MaxDepth        int32
	IncludeInactive bool
}) ([]dto.OrganizationHierarchyData, error) {
	log := r.loggerFor("organization", "subtree", pkglogger.Fields{
		"code":     args.Code,
		"maxDepth": args.MaxDepth,
	})
	if err := r.authorize(ctx, "organizationSubtree", log); err != nil {
		return nil, err
	}
	log.WithFields(pkglogger.Fields{"tenantId": args.TenantId, "includeInactive": args.IncludeInactive}).Info("执行组织子树查询")

	tenantID, err := uuid.Parse(args.TenantId)
	if err != nil {
		log.WithFields(pkglogger.Fields{"error": err}).Warn("invalid tenant ID")
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
		return []dto.OrganizationHierarchyData{}, nil
	}

	// 先转换根节点
	root := dto.OrganizationHierarchyData{
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
		ChildrenField:       []dto.OrganizationHierarchyData{}, // 简化实现，先不递归转换
	}

	return []dto.OrganizationHierarchyData{root}, nil
}

// 层级统计查询
func (r *Resolver) HierarchyStatistics(ctx context.Context, args struct {
	TenantId              string
	IncludeIntegrityCheck bool
}) (*dto.HierarchyStatistics, error) {
	log := r.loggerFor("organization", "hierarchyStats", pkglogger.Fields{"tenantId": args.TenantId, "includeIntegrityCheck": args.IncludeIntegrityCheck})
	if err := r.authorize(ctx, "hierarchyStatistics", log); err != nil {
		return nil, err
	}

	// TODO: 实现实际的层级统计逻辑
	return &dto.HierarchyStatistics{
		TenantIdField:           args.TenantId,
		TotalOrganizationsField: 0,
		MaxDepthField:           0,
		AvgDepthField:           0.0,
		DepthDistributionField:  []dto.DepthDistribution{},
		RootOrganizationsField:  0,
		LeafOrganizationsField:  0,
		IntegrityIssuesField:    []dto.IntegrityIssue{},
		LastAnalyzedField:       "",
	}, nil
}

// Positions 查询
func (r *Resolver) Positions(ctx context.Context, args struct {
	Filter     *dto.PositionFilterInput
	Pagination *dto.PaginationInput
	Sorting    *[]dto.PositionSortInput
}) (*dto.PositionConnection, error) {
	log := r.loggerFor("position", "list", pkglogger.Fields{"tenantId": sharedconfig.DefaultTenantID.String()})
	if err := r.authorize(ctx, "positions", log); err != nil {
		return nil, err
	}

	var sorting []dto.PositionSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}
	log.WithFields(pkglogger.Fields{
		"filter":     args.Filter,
		"pagination": args.Pagination,
		"sortCount":  len(sorting),
	}).Info("查询职位列表")

	return r.repo.GetPositions(ctx, sharedconfig.DefaultTenantID, args.Filter, args.Pagination, sorting)
}

// Position 查询单个职位
func (r *Resolver) Position(ctx context.Context, args struct {
	Code     string
	AsOfDate *string
}) (*dto.Position, error) {
	log := r.loggerFor("position", "get", pkglogger.Fields{
		"tenantId": sharedconfig.DefaultTenantID.String(),
		"code":     args.Code,
		"asOfDate": args.AsOfDate,
	})
	if err := r.authorize(ctx, "position", log); err != nil {
		return nil, err
	}
	log.Info("查询职位详情")

	return r.repo.GetPositionByCode(ctx, sharedconfig.DefaultTenantID, args.Code, args.AsOfDate)
}

// PositionAssignments 查询职位任职记录
func (r *Resolver) PositionAssignments(ctx context.Context, args struct {
	PositionCode string
	Filter       *dto.PositionAssignmentFilterInput
	Pagination   *dto.PaginationInput
	Sorting      *[]dto.PositionAssignmentSortInput
}) (*dto.PositionAssignmentConnection, error) {
	log := r.loggerFor("position", "assignments", pkglogger.Fields{
		"positionCode": args.PositionCode,
	})
	if err := r.authorize(ctx, "positionAssignments", log); err != nil {
		return nil, err
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("invalid tenant id")
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	var sorting []dto.PositionAssignmentSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}

	log.WithFields(pkglogger.Fields{
		"tenantId":   tenantID.String(),
		"filter":     args.Filter,
		"pagination": args.Pagination,
		"sortCount":  len(sorting),
	}).Info("查询职位任职记录")

	return r.repo.GetPositionAssignments(ctx, tenantID, args.PositionCode, args.Filter, args.Pagination, sorting)
}

// Assignments 查询多职位任职记录，兼容组织维度过滤
func (r *Resolver) Assignments(ctx context.Context, args struct {
	OrganizationCode *string
	PositionCode     *string
	Filter           *dto.PositionAssignmentFilterInput
	Pagination       *dto.PaginationInput
	Sorting          *[]dto.PositionAssignmentSortInput
}) (*dto.PositionAssignmentConnection, error) {
	log := r.loggerFor("assignments", "list", nil)
	if err := r.authorize(ctx, "assignments", log); err != nil {
		return nil, err
	}

	if r.assignFacade == nil {
		return nil, fmt.Errorf("ASSIGNMENT_QUERY_FACADE_NOT_CONFIGURED")
	}

	positionCode := ""
	if args.PositionCode != nil {
		positionCode = strings.TrimSpace(*args.PositionCode)
	}
	if positionCode == "" {
		return nil, fmt.Errorf("POSITION_CODE_REQUIRED")
	}

	tenantID := r.resolveTenant(ctx, log)
	var sorting []dto.PositionAssignmentSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}
	return r.assignFacade.GetAssignments(ctx, tenantID, positionCode, args.Filter, args.Pagination, sorting)
}

// AssignmentHistory 查询职位任职历史
func (r *Resolver) AssignmentHistory(ctx context.Context, args struct {
	PositionCode string
	Filter       *dto.PositionAssignmentFilterInput
	Pagination   *dto.PaginationInput
	Sorting      *[]dto.PositionAssignmentSortInput
}) (*dto.PositionAssignmentConnection, error) {
	log := r.loggerFor("assignments", "history", pkglogger.Fields{"positionCode": args.PositionCode})
	if err := r.authorize(ctx, "assignmentHistory", log); err != nil {
		return nil, err
	}
	if r.assignFacade == nil {
		return nil, fmt.Errorf("ASSIGNMENT_QUERY_FACADE_NOT_CONFIGURED")
	}
	tenantID := r.resolveTenant(ctx, log)
	var sorting []dto.PositionAssignmentSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}
	return r.assignFacade.GetAssignmentHistory(ctx, tenantID, args.PositionCode, args.Filter, args.Pagination, sorting)
}

// AssignmentStats 查询职位或组织的任职统计
func (r *Resolver) AssignmentStats(ctx context.Context, args struct {
	OrganizationCode *string
	PositionCode     *string
}) (*dto.AssignmentStats, error) {
	log := r.loggerFor("assignments", "stats", pkglogger.Fields{
		"positionCode":     args.PositionCode,
		"organizationCode": args.OrganizationCode,
	})
	if err := r.authorize(ctx, "assignmentStats", log); err != nil {
		return nil, err
	}
	if r.assignFacade == nil {
		return nil, fmt.Errorf("ASSIGNMENT_QUERY_FACADE_NOT_CONFIGURED")
	}
	positionCode := ""
	if args.PositionCode != nil {
		positionCode = strings.TrimSpace(*args.PositionCode)
	}
	orgCode := ""
	if args.OrganizationCode != nil {
		orgCode = strings.TrimSpace(*args.OrganizationCode)
	}
	if positionCode == "" && orgCode == "" {
		return nil, fmt.Errorf("POSITION_OR_ORGANIZATION_REQUIRED")
	}
	tenantID := r.resolveTenant(ctx, log)
	return r.assignFacade.GetAssignmentStats(ctx, tenantID, positionCode, orgCode)
}

func (r *Resolver) PositionAssignmentAudit(ctx context.Context, args struct {
	PositionCode string
	AssignmentId *string
	DateRange    *dto.DateRangeInput
	Pagination   *dto.PaginationInput
}) (*dto.PositionAssignmentAuditConnection, error) {
	log := r.loggerFor("position", "assignmentAudit", pkglogger.Fields{
		"positionCode": args.PositionCode,
		"assignmentId": args.AssignmentId,
	})
	if err := r.authorize(ctx, "positionAssignmentAudit", log); err != nil {
		return nil, err
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("invalid tenant id")
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	log.WithFields(pkglogger.Fields{"tenantId": tenantID.String(), "dateRange": args.DateRange}).Info("查询任职审计记录")
	return r.repo.GetPositionAssignmentAudit(ctx, tenantID, args.PositionCode, args.AssignmentId, args.DateRange, args.Pagination)
}

// PositionTimeline 查询职位时间线
func (r *Resolver) PositionTimeline(ctx context.Context, args struct {
	Code      string
	StartDate *string
	EndDate   *string
}) ([]dto.PositionTimelineEntry, error) {
	log := r.loggerFor("position", "timeline", pkglogger.Fields{
		"code":      args.Code,
		"startDate": args.StartDate,
		"endDate":   args.EndDate,
	})
	if err := r.authorize(ctx, "positionTimeline", log); err != nil {
		return nil, err
	}
	log.Info("查询职位时间线")

	return r.repo.GetPositionTimeline(ctx, sharedconfig.DefaultTenantID, args.Code, args.StartDate, args.EndDate)
}

func (r *Resolver) resolveTenant(ctx context.Context, log pkglogger.Logger) uuid.UUID {
	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("invalid tenant id")
			return sharedconfig.DefaultTenantID
		}
		return parsed
	}
	return tenantID
}

// PositionVersions 查询职位版本列表
func (r *Resolver) PositionVersions(ctx context.Context, args struct {
	Code           string
	IncludeDeleted *bool
}) ([]dto.Position, error) {
	includeDeleted := false
	if args.IncludeDeleted != nil {
		includeDeleted = *args.IncludeDeleted
	}
	log := r.loggerFor("position", "versions", pkglogger.Fields{
		"code":           args.Code,
		"includeDeleted": includeDeleted,
		"tenantId":       sharedconfig.DefaultTenantID.String(),
	})
	if err := r.authorize(ctx, "positionVersions", log); err != nil {
		return nil, err
	}
	log.Info("查询职位版本列表")

	return r.repo.GetPositionVersions(ctx, sharedconfig.DefaultTenantID, args.Code, includeDeleted)
}

// VacantPositions 查询空缺职位
func (r *Resolver) VacantPositions(ctx context.Context, args struct {
	Filter     *dto.VacantPositionFilterInput
	Pagination *dto.PaginationInput
	Sorting    *[]dto.VacantPositionSortInput
}) (*dto.VacantPositionConnection, error) {
	log := r.loggerFor("position", "vacant", nil)
	if err := r.authorize(ctx, "vacantPositions", log); err != nil {
		return nil, err
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("invalid tenant id")
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	var sorting []dto.VacantPositionSortInput
	if args.Sorting != nil {
		sorting = *args.Sorting
	}

	log.WithFields(pkglogger.Fields{
		"tenantId":   tenantID.String(),
		"filter":     args.Filter,
		"pagination": args.Pagination,
		"sortCount":  len(sorting),
	}).Info("查询空缺职位")

	return r.repo.GetVacantPositionConnection(ctx, tenantID, args.Filter, args.Pagination, sorting)
}

// PositionTransfers 查询职位转移记录
func (r *Resolver) PositionTransfers(ctx context.Context, args struct {
	PositionCode     *string
	OrganizationCode *string
	Pagination       *dto.PaginationInput
}) (*dto.PositionTransferConnection, error) {
	log := r.loggerFor("position", "transfers", pkglogger.Fields{
		"positionCode":     args.PositionCode,
		"organizationCode": args.OrganizationCode,
	})
	if err := r.authorize(ctx, "positionTransfers", log); err != nil {
		return nil, err
	}

	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("invalid tenant id")
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}

	log.WithFields(pkglogger.Fields{
		"tenantId":   tenantID.String(),
		"pagination": args.Pagination,
	}).Info("查询职位转移记录")

	return r.repo.GetPositionTransfers(ctx, tenantID, args.PositionCode, args.OrganizationCode, args.Pagination)
}

// PositionHeadcountStats 查询编制统计
func (r *Resolver) PositionHeadcountStats(ctx context.Context, args struct {
	OrganizationCode    string
	IncludeSubordinates *bool
}) (*dto.HeadcountStats, error) {
	log := r.loggerFor("position", "headcountStats", pkglogger.Fields{"organizationCode": args.OrganizationCode})
	if err := r.authorize(ctx, "positionHeadcountStats", log); err != nil {
		return nil, err
	}
	includeSubordinates := true
	if args.IncludeSubordinates != nil {
		includeSubordinates = *args.IncludeSubordinates
	}
	tenantID := sharedconfig.DefaultTenantID
	if tenantStr := auth.GetTenantID(ctx); tenantStr != "" {
		parsed, err := uuid.Parse(tenantStr)
		if err != nil {
			log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("invalid tenant id")
			return nil, fmt.Errorf("INVALID_TENANT")
		}
		tenantID = parsed
	}
	log.WithFields(pkglogger.Fields{
		"tenantId":           tenantID.String(),
		"includeSubordinate": includeSubordinates,
	}).Info("查询职位编制统计")

	return r.repo.GetPositionHeadcountStats(ctx, tenantID, args.OrganizationCode, includeSubordinates)
}

// JobFamilyGroups 查询职类
func (r *Resolver) JobFamilyGroups(ctx context.Context, args struct {
	IncludeInactive *bool
	AsOfDate        *string
}) ([]dto.JobFamilyGroup, error) {
	log := r.loggerFor("jobCatalog", "familyGroups", nil)
	if err := r.authorize(ctx, "jobFamilyGroups", log); err != nil {
		return nil, err
	}
	includeInactive := false
	if args.IncludeInactive != nil {
		includeInactive = *args.IncludeInactive
	}
	log.WithFields(pkglogger.Fields{"includeInactive": includeInactive, "asOfDate": args.AsOfDate}).Info("查询职类")

	return r.repo.GetJobFamilyGroups(ctx, sharedconfig.DefaultTenantID, includeInactive, args.AsOfDate)
}

// JobFamilies 查询职种
func (r *Resolver) JobFamilies(ctx context.Context, args struct {
	GroupCode       string
	IncludeInactive *bool
	AsOfDate        *string
}) ([]dto.JobFamily, error) {
	log := r.loggerFor("jobCatalog", "families", pkglogger.Fields{"groupCode": args.GroupCode})
	if err := r.authorize(ctx, "jobFamilies", log); err != nil {
		return nil, err
	}
	includeInactive := false
	if args.IncludeInactive != nil {
		includeInactive = *args.IncludeInactive
	}
	log.WithFields(pkglogger.Fields{"includeInactive": includeInactive, "asOfDate": args.AsOfDate}).Info("查询职种")

	return r.repo.GetJobFamilies(ctx, sharedconfig.DefaultTenantID, args.GroupCode, includeInactive, args.AsOfDate)
}

// JobRoles 查询职务
func (r *Resolver) JobRoles(ctx context.Context, args struct {
	FamilyCode      string
	IncludeInactive *bool
	AsOfDate        *string
}) ([]dto.JobRole, error) {
	log := r.loggerFor("jobCatalog", "roles", pkglogger.Fields{"familyCode": args.FamilyCode})
	if err := r.authorize(ctx, "jobRoles", log); err != nil {
		return nil, err
	}
	includeInactive := false
	if args.IncludeInactive != nil {
		includeInactive = *args.IncludeInactive
	}
	log.WithFields(pkglogger.Fields{"includeInactive": includeInactive, "asOfDate": args.AsOfDate}).Info("查询职务")

	return r.repo.GetJobRoles(ctx, sharedconfig.DefaultTenantID, args.FamilyCode, includeInactive, args.AsOfDate)
}

// JobLevels 查询职级
func (r *Resolver) JobLevels(ctx context.Context, args struct {
	RoleCode        string
	IncludeInactive *bool
	AsOfDate        *string
}) ([]dto.JobLevel, error) {
	log := r.loggerFor("jobCatalog", "levels", pkglogger.Fields{"roleCode": args.RoleCode})
	if err := r.authorize(ctx, "jobLevels", log); err != nil {
		return nil, err
	}
	includeInactive := false
	if args.IncludeInactive != nil {
		includeInactive = *args.IncludeInactive
	}
	log.WithFields(pkglogger.Fields{"includeInactive": includeInactive, "asOfDate": args.AsOfDate}).Info("查询职级")

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
}) ([]dto.AuditRecordData, error) {
	log := r.loggerFor("audit", "history", pkglogger.Fields{
		"recordId": args.RecordId,
		"limit":    args.Limit,
	})
	if err := r.authorize(ctx, "auditHistory", log); err != nil {
		return nil, err
	}

	limit := int32(50) // 默认限制
	if args.Limit > 0 {
		limit = args.Limit
		if limit > 200 { // API规范限制最大200
			limit = 200
		}
	}

	tenantStr := auth.GetTenantID(ctx)
	if tenantStr == "" {
		log.Warn("缺少租户ID，拒绝审计历史查询")
		return nil, fmt.Errorf("TENANT_REQUIRED")
	}
	tenantUUID, err := uuid.Parse(tenantStr)
	if err != nil {
		log.WithFields(pkglogger.Fields{"tenantId": tenantStr, "error": err}).Warn("无效租户ID")
		return nil, fmt.Errorf("INVALID_TENANT")
	}

	log.WithFields(pkglogger.Fields{
		"tenantId":  tenantUUID.String(),
		"startDate": args.StartDate,
		"endDate":   args.EndDate,
		"operation": args.Operation,
		"userId":    args.UserId,
		"limit":     limit,
	}).Info("执行审计历史查询")

	return r.repo.GetAuditHistory(ctx, tenantUUID, args.RecordId, args.StartDate, args.EndDate, args.Operation, args.UserId, int(limit))
}

// 单条审计记录查询 - v4.6.0
func (r *Resolver) AuditLog(ctx context.Context, args struct {
	AuditId string
}) (*dto.AuditRecordData, error) {
	log := r.loggerFor("audit", "log", pkglogger.Fields{"auditId": args.AuditId})
	if err := r.authorize(ctx, "auditLog", log); err != nil {
		return nil, err
	}
	log.Info("单条审计记录查询")
	return r.repo.GetAuditLog(ctx, args.AuditId)
}
