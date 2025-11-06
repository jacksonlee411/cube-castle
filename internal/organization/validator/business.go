package validator

import (
	"context"
	"fmt"
	"regexp"
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

var validOrganizationStatuses = map[string]struct{}{
	string(types.OrganizationStatusActive):   {},
	string(types.OrganizationStatusInactive): {},
	string(types.OrganizationStatusPlanned):  {},
	string(types.OrganizationStatusDeleted):  {},
}

var (
	organizationCodeRegex       = regexp.MustCompile(types.OrganizationCodePattern)
	organizationParentCodeRegex = regexp.MustCompile(types.OrganizationParentCodePattern)
	organizationNameRegex       = regexp.MustCompile(`^[\p{L}\p{N}\s\-\(\)（）]+$`)
)

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

	v.validateCreateRequestBasics(req, result)

	selfReferentialParent := false
	if req.Code != nil && req.ParentCode != nil && strings.EqualFold(*req.Code, *req.ParentCode) {
		selfReferentialParent = true
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORG_CYCLE_DETECTED",
			Message:  "Organization cannot be its own parent",
			Field:    "parentCode",
			Severity: string(SeverityCritical),
			Context: map[string]interface{}{
				"ruleId":          "ORG-CIRC",
				"attemptedParent": *req.ParentCode,
			},
		})
	}

	if req.ParentCode != nil && *req.ParentCode != "" {
		if selfReferentialParent {
			result.Valid = false
			return result
		}

		parent, err := v.hierarchyRepo.GetOrganization(ctx, *req.ParentCode, tenantID)
		if err != nil || parent == nil {
			result.Errors = append(result.Errors, ValidationError{
				Code:     ErrorCodeInvalidParent,
				Message:  fmt.Sprintf("父组织不存在或无效: %s", *req.ParentCode),
				Field:    "parentCode",
				Severity: string(SeverityHigh),
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
			Severity: string(SeverityCritical),
		})
		result.Valid = false
		return result
	}

	result.Context["existing_organization"] = existingOrg

	v.validateUpdateRequestBasics(req, result, existingOrg)

	if req.EffectiveDate != nil || req.EndDate != nil {
		if err := v.validateTemporalData(req.EffectiveDate, req.EndDate, result); err != nil {
			v.logger.Warnf("时态数据验证失败: %v", err)
		}
	}

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
	chain := NewValidationChain(
		v.logger,
		WithOperationLabel("TemporalParentAvailability"),
		WithBaseContext(map[string]interface{}{
			"operation": "TemporalParentAvailability",
		}),
	)
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
			Severity: string(SeverityHigh),
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

func (v *BusinessRuleValidator) validateCreateRequestBasics(req *types.CreateOrganizationRequest, result *ValidationResult) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORG_NAME_REQUIRED",
			Message:  "组织名称不能为空",
			Field:    "name",
			Severity: string(SeverityHigh),
		})
	} else {
		if len(req.Name) > types.OrganizationNameMaxLength {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_NAME_TOO_LONG",
				Message:  fmt.Sprintf("组织名称不能超过%d个字符", types.OrganizationNameMaxLength),
				Field:    "name",
				Severity: string(SeverityHigh),
			})
		}
		if !organizationNameRegex.MatchString(req.Name) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_NAME_INVALID",
				Message:  "组织名称包含无效字符，只允许字母、数字、中文、空格、连字符和括号",
				Field:    "name",
				Severity: string(SeverityHigh),
			})
		}
	}

	unitType := strings.TrimSpace(req.UnitType)
	if unitType == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORG_UNIT_TYPE_REQUIRED",
			Message:  "组织类型不能为空",
			Field:    "unitType",
			Severity: string(SeverityHigh),
		})
	} else {
		unitTypeUpper := strings.ToUpper(unitType)
		if _, ok := validUnitTypes[unitTypeUpper]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_UNIT_TYPE_INVALID",
				Message:  fmt.Sprintf("无效的组织类型: %s", unitType),
				Field:    "unitType",
				Value:    unitType,
				Severity: string(SeverityHigh),
			})
		} else {
			req.UnitType = unitTypeUpper
		}
	}

	if req.Code != nil {
		trimmed := strings.TrimSpace(*req.Code)
		if trimmed == "" {
			req.Code = nil
		} else if !organizationCodeRegex.MatchString(trimmed) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_CODE_INVALID",
				Message:  "组织代码格式无效，必须为7位数字且首位不可为0",
				Field:    "code",
				Value:    trimmed,
				Severity: string(SeverityHigh),
			})
		} else {
			req.Code = &trimmed
		}
	}

	if req.ParentCode != nil {
		trimmedParent := strings.TrimSpace(*req.ParentCode)
		if trimmedParent == "" {
			req.ParentCode = nil
		} else if !organizationParentCodeRegex.MatchString(trimmedParent) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_PARENT_INVALID",
				Message:  "父组织代码格式无效，需为0或合法的7位数字编码",
				Field:    "parentCode",
				Value:    trimmedParent,
				Severity: string(SeverityHigh),
			})
			req.ParentCode = nil
		} else {
			req.ParentCode = &trimmedParent
		}
	}

	if req.SortOrder < 0 || req.SortOrder > 9999 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORG_SORT_ORDER_INVALID",
			Message:  "排序顺序需在0到9999之间",
			Field:    "sortOrder",
			Value:    req.SortOrder,
			Severity: string(SeverityHigh),
		})
	}

	req.Description = strings.TrimSpace(req.Description)
	if len(req.Description) > types.OrganizationDescriptionMaxLength {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORG_DESCRIPTION_TOO_LONG",
			Message:  fmt.Sprintf("描述长度不能超过%d个字符", types.OrganizationDescriptionMaxLength),
			Field:    "description",
			Severity: string(SeverityHigh),
		})
	}
}

func (v *BusinessRuleValidator) validateUpdateRequestBasics(req *types.UpdateOrganizationRequest, result *ValidationResult, existing *types.Organization) {
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_NAME_REQUIRED",
				Message:  "组织名称不能为空",
				Field:    "name",
				Severity: string(SeverityHigh),
			})
		} else {
			if len(trimmed) > types.OrganizationNameMaxLength {
				result.Errors = append(result.Errors, ValidationError{
					Code:     "ORG_NAME_TOO_LONG",
					Message:  fmt.Sprintf("组织名称不能超过%d个字符", types.OrganizationNameMaxLength),
					Field:    "name",
					Severity: string(SeverityHigh),
				})
			}
			if !organizationNameRegex.MatchString(trimmed) {
				result.Errors = append(result.Errors, ValidationError{
					Code:     "ORG_NAME_INVALID",
					Message:  "组织名称包含无效字符，只允许字母、数字、中文、空格、连字符和括号",
					Field:    "name",
					Severity: string(SeverityHigh),
				})
			}
			req.Name = &trimmed
		}
	}

	if req.UnitType != nil {
		trimmed := strings.TrimSpace(*req.UnitType)
		if trimmed == "" {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_UNIT_TYPE_REQUIRED",
				Message:  "组织类型不能为空",
				Field:    "unitType",
				Severity: string(SeverityHigh),
			})
		} else {
			unitTypeUpper := strings.ToUpper(trimmed)
			if _, ok := validUnitTypes[unitTypeUpper]; !ok {
				result.Errors = append(result.Errors, ValidationError{
					Code:     "ORG_UNIT_TYPE_INVALID",
					Message:  fmt.Sprintf("无效的组织类型: %s", trimmed),
					Field:    "unitType",
					Value:    trimmed,
					Severity: string(SeverityHigh),
				})
			} else {
				req.UnitType = &unitTypeUpper
			}
		}
	}

	if req.ParentCode != nil {
		trimmedParent := strings.TrimSpace(*req.ParentCode)
		if trimmedParent == "" {
			req.ParentCode = nil
		} else if !organizationParentCodeRegex.MatchString(trimmedParent) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_PARENT_INVALID",
				Message:  "父组织代码格式无效，需为0或合法的7位数字编码",
				Field:    "parentCode",
				Value:    trimmedParent,
				Severity: string(SeverityHigh),
			})
			req.ParentCode = nil
		} else {
			req.ParentCode = &trimmedParent
		}
	}

	if req.SortOrder != nil {
		if *req.SortOrder < 0 || *req.SortOrder > 9999 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_SORT_ORDER_INVALID",
				Message:  "排序顺序需在0到9999之间",
				Field:    "sortOrder",
				Value:    *req.SortOrder,
				Severity: string(SeverityHigh),
			})
		}
	}

	if req.Status != nil {
		trimmedStatus := strings.ToUpper(strings.TrimSpace(*req.Status))
		if trimmedStatus == "" {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_STATUS_REQUIRED",
				Message:  "状态不能为空",
				Field:    "status",
				Severity: string(SeverityHigh),
			})
		} else if _, ok := validOrganizationStatuses[trimmedStatus]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_STATUS_INVALID",
				Message:  fmt.Sprintf("无效的组织状态: %s", trimmedStatus),
				Field:    "status",
				Value:    trimmedStatus,
				Severity: string(SeverityHigh),
			})
		} else {
			req.Status = &trimmedStatus
		}
	}

	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if len(trimmed) > types.OrganizationDescriptionMaxLength {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ORG_DESCRIPTION_TOO_LONG",
				Message:  fmt.Sprintf("描述长度不能超过%d个字符", types.OrganizationDescriptionMaxLength),
				Field:    "description",
				Severity: string(SeverityHigh),
			})
		}
		req.Description = &trimmed
	}

	// 当未提供新状态时，保持现有状态写回上下文，便于后续规则判断。
	if req.Status == nil && existing != nil {
		result.Context["existingStatus"] = strings.ToUpper(strings.TrimSpace(existing.Status))
	}
}
