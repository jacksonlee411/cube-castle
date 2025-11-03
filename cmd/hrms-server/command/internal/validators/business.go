package validators

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"cube-castle/cmd/hrms-server/command/internal/repository"
	"cube-castle/internal/types"
)

// BusinessRuleValidator 业务规则验证器
type BusinessRuleValidator struct {
	hierarchyRepo *repository.HierarchyRepository
	orgRepo       *repository.OrganizationRepository
	logger        *log.Logger
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

func NewBusinessRuleValidator(hierarchyRepo *repository.HierarchyRepository, orgRepo *repository.OrganizationRepository, logger *log.Logger) *BusinessRuleValidator {
	return &BusinessRuleValidator{
		hierarchyRepo: hierarchyRepo,
		orgRepo:       orgRepo,
		logger:        logger,
	}
}

// ValidateOrganizationCreation 验证组织创建
func (v *BusinessRuleValidator) ValidateOrganizationCreation(ctx context.Context, req *types.CreateOrganizationRequest, tenantID uuid.UUID) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Context:  make(map[string]interface{}),
	}

	// 1. 验证父组织存在性和有效性
	if req.ParentCode != nil && *req.ParentCode != "" {
		if err := v.validateParentOrganization(ctx, *req.ParentCode, tenantID, result); err != nil {
			v.logger.Printf("父组织验证失败: %v", err)
		}
	}

	// 2. 验证层级深度限制
	if req.ParentCode != nil {
		if err := v.validateDepthLimit(ctx, req.ParentCode, tenantID, result); err != nil {
			v.logger.Printf("深度限制验证失败: %v", err)
		}
	}

	if req.ParentCode != nil && *req.ParentCode != "" && req.EffectiveDate != nil {
		if err := v.validateTemporalParentAvailability(ctx, tenantID, *req.ParentCode, req.EffectiveDate.Time, result); err != nil {
			v.logger.Printf("时态父级验证失败: %v", err)
		}
	}

	// 3. 验证代码唯一性
	if req.Code != nil {
		if err := v.validateCodeUniqueness(ctx, *req.Code, tenantID, result); err != nil {
			v.logger.Printf("代码唯一性验证失败: %v", err)
		}
	}

	// 4. 验证时态数据一致性
	if err := v.validateTemporalData(req.EffectiveDate, req.EndDate, result); err != nil {
		v.logger.Printf("时态数据验证失败: %v", err)
	}

	// 5. 验证业务逻辑规则
	if err := v.validateBusinessLogic(ctx, req, tenantID, result); err != nil {
		v.logger.Printf("业务逻辑验证失败: %v", err)
	}

	result.Valid = len(result.Errors) == 0
	v.logger.Printf("组织创建验证完成: 有效=%t, 错误数=%d, 警告数=%d",
		result.Valid, len(result.Errors), len(result.Warnings))

	return result
}

// ValidateOrganizationUpdate 验证组织更新
func (v *BusinessRuleValidator) ValidateOrganizationUpdate(ctx context.Context, code string, req *types.UpdateOrganizationRequest, tenantID uuid.UUID) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Context:  make(map[string]interface{}),
	}

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

	// 2. 如果更改了父组织，需要验证循环引用
	if req.ParentCode != nil {
		if err := v.validateCircularReference(ctx, code, *req.ParentCode, tenantID, result); err != nil {
			v.logger.Printf("循环引用验证失败: %v", err)
		}
	}

	// 3. 验证状态转换合法性
	if req.Status != nil {
		if err := v.validateStatusTransition(existingOrg.Status, *req.Status, result); err != nil {
			v.logger.Printf("状态转换验证失败: %v", err)
		}
	}

	// 4. 验证父组织变更影响
	if req.ParentCode != nil && (existingOrg.ParentCode == nil || *existingOrg.ParentCode != *req.ParentCode) {
		if err := v.validateParentChange(ctx, code, *req.ParentCode, tenantID, result); err != nil {
			v.logger.Printf("父组织变更验证失败: %v", err)
		}
	}

	if req.ParentCode != nil && *req.ParentCode != "" {
		var (
			effectiveAt  time.Time
			hasEffective bool
		)
		if req.EffectiveDate != nil {
			effectiveAt = req.EffectiveDate.Time
			hasEffective = true
		} else if existingOrg != nil && existingOrg.EffectiveDate != nil {
			effectiveAt = existingOrg.EffectiveDate.Time
			hasEffective = true
		}
		if hasEffective {
			if err := v.validateTemporalParentAvailability(ctx, tenantID, *req.ParentCode, effectiveAt, result); err != nil {
				v.logger.Printf("时态父级可用性验证失败: %v", err)
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// validateParentOrganization 验证父组织
func (v *BusinessRuleValidator) validateParentOrganization(ctx context.Context, parentCode string, tenantID uuid.UUID, result *ValidationResult) error {
	parent, err := v.hierarchyRepo.GetOrganization(ctx, parentCode, tenantID)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     ErrorCodeInvalidParent,
			Message:  fmt.Sprintf("父组织不存在或无效: %s", parentCode),
			Field:    "parentCode",
			Value:    parentCode,
			Severity: "HIGH",
		})
		return err
	}

	// 检查父组织状态
	if parent.Status == "INACTIVE" || parent.Status == "DELETED" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "PARENT_INACTIVE",
			Message: fmt.Sprintf("父组织状态为 %s，可能影响子组织功能", parent.Status),
			Field:   "parentCode",
			Value:   parentCode,
		})
	}

	result.Context["parent_organization"] = parent
	return nil
}

// validateDepthLimit 验证层级深度限制 (最大17级)
func (v *BusinessRuleValidator) validateDepthLimit(ctx context.Context, parentCode *string, tenantID uuid.UUID, result *ValidationResult) error {
	if parentCode == nil || *parentCode == "" {
		return nil // 根组织无深度限制
	}

	depth, err := v.hierarchyRepo.GetOrganizationDepth(ctx, *parentCode, tenantID)
	if err != nil {
		return fmt.Errorf("获取组织深度失败: %w", err)
	}

	result.Context["parent_depth"] = depth

	if depth >= 17 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     ErrorCodeDepthExceeded,
			Message:  fmt.Sprintf("超过最大层级深度限制 (17级)，当前父组织深度: %d", depth),
			Field:    "parentCode",
			Value:    *parentCode,
			Severity: "HIGH",
		})
		return fmt.Errorf("depth limit exceeded: %d", depth)
	}

	// 深度接近限制时发出警告
	if depth >= 15 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "DEPTH_WARNING",
			Message: fmt.Sprintf("组织深度接近限制 (当前: %d/17)", depth+1),
			Field:   "parentCode",
			Value:   depth + 1,
		})
	}

	return nil
}

func (v *BusinessRuleValidator) validateTemporalParentAvailability(ctx context.Context, tenantID uuid.UUID, parentCode string, effectiveDate time.Time, result *ValidationResult) error {
	parentAtDate, err := v.hierarchyRepo.GetOrganizationAtDate(ctx, parentCode, tenantID, effectiveDate)
	if err != nil {
		return fmt.Errorf("查询父组织时态失败: %w", err)
	}

	if parentAtDate == nil || strings.ToUpper(parentAtDate.Status) != "ACTIVE" {
		context := map[string]interface{}{}
		message := fmt.Sprintf("上级组织 %s 在指定生效日期 %s 不存在或未激活。", parentCode, effectiveDate.Format("2006-01-02"))

		if latestParent, latestErr := v.hierarchyRepo.GetOrganization(ctx, parentCode, tenantID); latestErr == nil && latestParent != nil {
			context["parentName"] = latestParent.Name
			if latestParent.EffectiveDate != nil {
				suggested := latestParent.EffectiveDate.String()
				if suggested != "" {
					context["suggestedDate"] = suggested
					message += fmt.Sprintf(" 可选择在 %s 之后生效，或更换上级组织。", suggested)
				}
			}
		}

		result.Errors = append(result.Errors, ValidationError{
			Code:     "TEMPORAL_PARENT_UNAVAILABLE",
			Message:  message,
			Field:    "parentCode",
			Value:    parentCode,
			Severity: "HIGH",
			Context:  context,
		})
		result.Valid = false
	}

	return nil
}

// ValidateTemporalParentAvailability 导出校验结果，供Handler复用
func (v *BusinessRuleValidator) ValidateTemporalParentAvailability(ctx context.Context, tenantID uuid.UUID, parentCode string, effectiveDate time.Time) *ValidationResult {
	result := NewValidationResult()
	if err := v.validateTemporalParentAvailability(ctx, tenantID, parentCode, effectiveDate, result); err != nil {
		v.logger.Printf("时态父级可用性验证失败: %v", err)
	}
	result.Valid = len(result.Errors) == 0
	return result
}

// validateCircularReference 检测循环引用
func (v *BusinessRuleValidator) validateCircularReference(ctx context.Context, code, parentCode string, tenantID uuid.UUID, result *ValidationResult) error {
	if parentCode == "" {
		return nil // 根组织无循环引用风险
	}

	if code == parentCode {
		result.Errors = append(result.Errors, ValidationError{
			Code:     ErrorCodeCircularReference,
			Message:  "组织不能将自己设为父组织",
			Field:    "parentCode",
			Value:    parentCode,
			Severity: "CRITICAL",
		})
		return fmt.Errorf("self reference detected")
	}

	// 获取目标父组织的祖先链
	ancestors, err := v.hierarchyRepo.GetAncestorChain(ctx, parentCode, tenantID)
	if err != nil {
		return fmt.Errorf("获取祖先链失败: %w", err)
	}

	// 检查当前组织是否在祖先链中
	for _, ancestor := range ancestors {
		if ancestor.Code == code {
			result.Errors = append(result.Errors, ValidationError{
				Code:     ErrorCodeCircularReference,
				Message:  fmt.Sprintf("检测到循环引用：组织 %s 不能成为其子孙组织 %s 的父组织", code, parentCode),
				Field:    "parentCode",
				Value:    parentCode,
				Severity: "CRITICAL",
			})
			return fmt.Errorf("circular reference detected")
		}
	}

	result.Context["ancestor_count"] = len(ancestors)
	return nil
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

// validateStatusTransition 验证状态转换
func (v *BusinessRuleValidator) validateStatusTransition(currentStatus, newStatus string, result *ValidationResult) error {
	// 定义合法的状态转换矩阵
	validTransitions := map[string][]string{
		"ACTIVE":   {"INACTIVE", "DELETED"},
		"INACTIVE": {"ACTIVE", "DELETED"},
		"PLANNED":  {"ACTIVE", "DELETED"},
		"DELETED":  {}, // 已删除状态不能转换到其他状态
	}

	validTargets, exists := validTransitions[currentStatus]
	if !exists {
		result.Errors = append(result.Errors, ValidationError{
			Code:     ErrorCodeInvalidStatus,
			Message:  fmt.Sprintf("无效的当前状态: %s", currentStatus),
			Value:    currentStatus,
			Severity: "HIGH",
		})
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	// 状态未发生变化时允许幂等更新
	if currentStatus == newStatus {
		return nil
	}

	// 检查目标状态是否合法
	isValidTransition := false
	for _, validStatus := range validTargets {
		if validStatus == newStatus {
			isValidTransition = true
			break
		}
	}

	if !isValidTransition {
		result.Errors = append(result.Errors, ValidationError{
			Code:     ErrorCodeStatusConflict,
			Message:  fmt.Sprintf("不能从状态 %s 转换到 %s", currentStatus, newStatus),
			Field:    "status",
			Value:    newStatus,
			Severity: "HIGH",
		})
		return fmt.Errorf("invalid status transition: %s -> %s", currentStatus, newStatus)
	}

	return nil
}

// validateParentChange 验证父组织变更影响
func (v *BusinessRuleValidator) validateParentChange(ctx context.Context, code, newParentCode string, tenantID uuid.UUID, result *ValidationResult) error {
	// 检查是否有子组织
	children, err := v.hierarchyRepo.GetDirectChildren(ctx, code, tenantID)
	if err != nil {
		return fmt.Errorf("检查子组织失败: %w", err)
	}

	if len(children) > 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Code:    "PARENT_CHANGE_WITH_CHILDREN",
			Message: fmt.Sprintf("该组织有 %d 个子组织，变更父组织将影响整个子树结构", len(children)),
			Field:   "parentCode",
			Value:   len(children),
		})
	}

	// 计算变更后的深度影响
	newParentDepth, err := v.hierarchyRepo.GetOrganizationDepth(ctx, newParentCode, tenantID)
	if err == nil {
		maxChildDepth := 0
		for _, child := range children {
			childDepth := child.Depth
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}

		// 计算变更后最大深度
		newMaxDepth := newParentDepth + 1 + maxChildDepth
		if newMaxDepth > 17 {
			result.Errors = append(result.Errors, ValidationError{
				Code:     ErrorCodeDepthExceeded,
				Message:  fmt.Sprintf("父组织变更将导致子树超过最大深度限制 (预计深度: %d)", newMaxDepth),
				Field:    "parentCode",
				Value:    newMaxDepth,
				Severity: "HIGH",
			})
		}
	}

	return nil
}

// ValidateHierarchyConsistency 验证层级一致性
func (v *BusinessRuleValidator) ValidateHierarchyConsistency(ctx context.Context, code string, tenantID uuid.UUID) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Context:  make(map[string]interface{}),
	}

	org, err := v.hierarchyRepo.GetOrganization(ctx, code, tenantID)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "ORGANIZATION_NOT_FOUND",
			Message:  fmt.Sprintf("组织不存在: %s", code),
			Severity: "CRITICAL",
		})
		result.Valid = false
		return result
	}

	// 验证路径一致性
	if org.ParentCode != nil {
		ancestors, err := v.hierarchyRepo.GetAncestorChain(ctx, code, tenantID)
		if err == nil {
			// 构建预期的代码路径
			expectedCodePath := ""
			for i, ancestor := range ancestors {
				if i > 0 {
					expectedCodePath += "/"
				}
				expectedCodePath += ancestor.Code
			}

			actualCodePath := strings.TrimSpace(org.CodePath)

			if actualCodePath != expectedCodePath {
				result.Errors = append(result.Errors, ValidationError{
					Code:     ErrorCodePathInconsistency,
					Message:  fmt.Sprintf("代码路径不一致: 期望=%s, 实际=%s", expectedCodePath, actualCodePath),
					Field:    "codePath",
					Severity: "MEDIUM",
				})
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}
