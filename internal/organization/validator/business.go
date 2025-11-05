package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// BusinessRuleValidator 业务规则验证器
type BusinessRuleValidator struct {
	hierarchyRepo hierarchyRepository
	orgRepo       organizationRepository
	logger        pkglogger.Logger
}

type hierarchyRepository interface {
	GetOrganization(ctx context.Context, code string, tenantID uuid.UUID) (*types.Organization, error)
	GetOrganizationDepth(ctx context.Context, code string, tenantID uuid.UUID) (int, error)
	GetAncestorChain(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error)
	GetDirectChildren(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error)
	GetOrganizationAtDate(ctx context.Context, code string, tenantID uuid.UUID, ts time.Time) (*repository.OrganizationNode, error)
}

type organizationRepository interface {
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*types.Organization, error)
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid    bool                   `json:"valid"`
	Errors   []ValidationError      `json:"errors"`
	Warnings []ValidationWarning    `json:"warnings"`
	Context  map[string]interface{} `json:"context"`
}

// NewValidationResult 创建默认有效的验证结果
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Context:  make(map[string]interface{}),
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Field    string                 `json:"field,omitempty"`
	Value    interface{}            `json:"value,omitempty"`
	Severity string                 `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	Context  map[string]interface{} `json:"context,omitempty"`
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Field   string      `json:"field,omitempty"`
	Value   interface{} `json:"value,omitempty"`
}

// 业务规则错误代码
const (
	// 层级结构规则
	ErrorCodeDepthExceeded      = "DEPTH_EXCEEDED"
	ErrorCodeCircularReference  = "CIRCULAR_REFERENCE"
	ErrorCodeOrphanOrganization = "ORPHAN_ORGANIZATION"
	ErrorCodeInvalidParent      = "INVALID_PARENT"

	// 数据一致性规则
	ErrorCodePathInconsistency  = "PATH_INCONSISTENCY"
	ErrorCodeLevelInconsistency = "LEVEL_INCONSISTENCY"
	ErrorCodeDuplicateCode      = "DUPLICATE_CODE"

	// 业务逻辑规则
	ErrorCodeInvalidStatus    = "INVALID_STATUS"
	ErrorCodeStatusConflict   = "STATUS_CONFLICT"
	ErrorCodeTemporalConflict = "TEMPORAL_CONFLICT"
	ErrorCodePermissionDenied = "PERMISSION_DENIED"
)

var validUnitTypes = map[string]struct{}{
	string(types.UnitTypeDepartment):       {},
	string(types.UnitTypeOrganizationUnit): {},
	string(types.UnitTypeCompany):          {},
	string(types.UnitTypeProjectTeam):      {},
}

func NewBusinessRuleValidator(hierarchyRepo *repository.HierarchyRepository, orgRepo *repository.OrganizationRepository, baseLogger pkglogger.Logger) *BusinessRuleValidator {
	return &BusinessRuleValidator{
		hierarchyRepo: hierarchyRepo,
		orgRepo:       orgRepo,
		logger: baseLogger.WithFields(pkglogger.Fields{
			"component": "validator",
			"module":    "organization",
		}),
	}
}

// ValidateOrganizationCreation 验证组织创建
func (v *BusinessRuleValidator) ValidateOrganizationCreation(ctx context.Context, req *types.CreateOrganizationRequest, tenantID uuid.UUID) *ValidationResult {
	result := NewValidationResult()

	if req.ParentCode != nil && *req.ParentCode != "" {
		parent, err := v.hierarchyRepo.GetOrganization(ctx, *req.ParentCode, tenantID)
		if err != nil || parent == nil {
			result.Errors = append(result.Errors, ValidationError{
				Code:     ErrorCodeInvalidParent,
				Message:  fmt.Sprintf("父组织不存在或无效: %s", *req.ParentCode),
				Field:    "parentCode",
				Severity: "HIGH",
			})
			result.Valid = false
			return result
		}

		result.Context["parent_organization"] = parent
		if strings.EqualFold(parent.Status, "INACTIVE") || strings.EqualFold(parent.Status, "DELETED") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    "PARENT_INACTIVE",
				Message: fmt.Sprintf("父组织状态为 %s，可能影响子组织功能", parent.Status),
				Field:   "parentCode",
				Value:   parent.Status,
			})
		}
	}

	chain := v.buildOrganizationCreateChain(req)
	subject := &organizationCreateSubject{
		TenantID: tenantID,
		Request:  req,
	}

	outcome := chain.Execute(ctx, subject)
	result.Errors = append(result.Errors, outcome.Errors...)
	result.Warnings = append(result.Warnings, outcome.Warnings...)
	for k, v := range outcome.Context {
		result.Context[k] = v
	}

	if req.Code != nil {
		if err := v.validateCodeUniqueness(ctx, *req.Code, tenantID, result); err != nil {
			v.logger.Warnf("代码唯一性验证失败: %v", err)
		}
	}

	if err := v.validateTemporalData(req.EffectiveDate, req.EndDate, result); err != nil {
		v.logger.Warnf("时态数据验证失败: %v", err)
	}

	if err := v.validateBusinessLogic(ctx, req, tenantID, result); err != nil {
		v.logger.Warnf("业务逻辑验证失败: %v", err)
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateOrganizationUpdate 验证组织更新
func (v *BusinessRuleValidator) ValidateOrganizationUpdate(ctx context.Context, code string, req *types.UpdateOrganizationRequest, tenantID uuid.UUID) *ValidationResult {
	result := NewValidationResult()

	// 1. 验证目标组织存在
	existingOrg, err := v.hierarchyRepo.GetOrganization(ctx, code, tenantID)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORGANIZATION_NOT_FOUND",
			Message:  fmt.Sprintf("目标组织不存在: %s", code),
			Value:    code,
			Severity: "CRITICAL",
		})
		result.Valid = false
		return result
	}

	result.Context["existing_organization"] = existingOrg

	chain := v.buildOrganizationUpdateChain(existingOrg, req)

	subject := &organizationUpdateSubject{
		TenantID: tenantID,
		Code:     code,
		Request:  req,
		Existing: existingOrg,
	}

	outcome := chain.Execute(ctx, subject)
	result.Errors = append(result.Errors, outcome.Errors...)
	result.Warnings = append(result.Warnings, outcome.Warnings...)
	for k, v := range outcome.Context {
		result.Context[k] = v
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateTemporalParentAvailability 提供给 handler 的时态父级校验入口
func (v *BusinessRuleValidator) ValidateTemporalParentAvailability(ctx context.Context, tenantID uuid.UUID, parentCode string, effectiveDate time.Time) *ValidationResult {
	parent := strings.TrimSpace(parentCode)
	chain := NewValidationChain(v.logger)
	chain.Register(&Rule{
		ID:           "ORG-TEMPORAL",
		Priority:     10,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      v.newOrgTemporalRule(parent, &effectiveDate),
	})

	result := chain.Execute(ctx, &organizationUpdateSubject{
		TenantID: tenantID,
	})
	result.Valid = len(result.Errors) == 0
	return result
}

// validateCodeUniqueness 验证代码唯一性
func (v *BusinessRuleValidator) validateCodeUniqueness(ctx context.Context, code string, tenantID uuid.UUID, result *ValidationResult) error {
	_, err := v.hierarchyRepo.GetOrganization(ctx, code, tenantID)
	if err == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     ErrorCodeDuplicateCode,
			Message:  fmt.Sprintf("组织代码 %s 已存在", code),
			Field:    "code",
			Value:    code,
			Severity: "HIGH",
		})
		return fmt.Errorf("duplicate code: %s", code)
	}

	return nil // 组织不存在，代码可用
}

// validateTemporalData 验证时态数据
func (v *BusinessRuleValidator) validateTemporalData(effectiveDate, endDate *types.Date, result *ValidationResult) error {
	if effectiveDate == nil {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "MISSING_EFFECTIVE_DATE",
			Message: "未指定生效日期，将使用当前日期",
			Field:   "effectiveDate",
		})
		return nil
	}

	if endDate != nil && effectiveDate != nil {
		if endDate.Before(effectiveDate.Time) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     ErrorCodeTemporalConflict,
				Message:  "结束日期不能早于生效日期",
				Field:    "endDate",
				Value:    endDate.String(),
				Severity: "MEDIUM",
			})
			return fmt.Errorf("temporal conflict: endDate < effectiveDate")
		}
	}

	return nil
}

// validateBusinessLogic 验证业务逻辑规则
func (v *BusinessRuleValidator) validateBusinessLogic(ctx context.Context, req *types.CreateOrganizationRequest, tenantID uuid.UUID, result *ValidationResult) error {
	// 1. 验证单位类型有效性
	if _, ok := validUnitTypes[strings.ToUpper(req.UnitType)]; !ok {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INVALID_UNIT_TYPE",
			Message:  fmt.Sprintf("无效的单位类型: %s", req.UnitType),
			Field:    "unitType",
			Value:    req.UnitType,
			Severity: "MEDIUM",
		})
	}

	// 2. 验证名称规则
	if len(req.Name) > types.OrganizationNameMaxLength {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "NAME_TOO_LONG",
			Message:  fmt.Sprintf("组织名称长度不能超过%d个字符", types.OrganizationNameMaxLength),
			Field:    "name",
			Value:    len(req.Name),
			Severity: "MEDIUM",
		})
	}

	// 3. 验证排序规则
	if req.SortOrder < 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "NEGATIVE_SORT_ORDER",
			Message: "排序值为负数，可能影响显示顺序",
			Field:   "sortOrder",
			Value:   req.SortOrder,
		})
	}

	return nil
}
