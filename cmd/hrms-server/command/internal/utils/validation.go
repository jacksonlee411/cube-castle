package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"cube-castle/cmd/hrms-server/command/internal/types"
)

var (
	organizationCodeRegex       = regexp.MustCompile(types.OrganizationCodePattern)
	organizationParentCodeRegex = regexp.MustCompile(types.OrganizationParentCodePattern)
	validUnitTypes              = map[types.UnitType]struct{}{
		types.UnitTypeDepartment:       {},
		types.UnitTypeOrganizationUnit: {},
		types.UnitTypeCompany:          {},
		types.UnitTypeProjectTeam:      {},
	}
	validStatuses = map[types.OrganizationStatus]struct{}{
		types.OrganizationStatusActive:   {},
		types.OrganizationStatusInactive: {},
		types.OrganizationStatusPlanned:  {},
		types.OrganizationStatusDeleted:  {},
	}
)

// ValidateCreateOrganization 验证创建组织请求
func ValidateCreateOrganization(req *types.CreateOrganizationRequest) error {
	// 1. 名称验证
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("组织名称不能为空")
	}

	if len(req.Name) > types.OrganizationNameMaxLength {
		return fmt.Errorf("组织名称不能超过%d个字符", types.OrganizationNameMaxLength)
	}

	// 名称格式验证（不能包含特殊字符）- 修复Unicode转义问题
	namePattern := regexp.MustCompile(`^[\p{L}\p{N}\s\-]+$`)
	if !namePattern.MatchString(req.Name) {
		return fmt.Errorf("组织名称包含无效字符，只允许字母、数字、中文、空格和连字符")
	}

	// 2. 组织类型验证
	if req.UnitType == "" {
		return fmt.Errorf("组织类型不能为空")
	}

	if _, ok := validUnitTypes[types.UnitType(req.UnitType)]; !ok {
		return fmt.Errorf("无效的组织类型: %s", req.UnitType)
	}

	// 3. 代码验证（如果提供）
	if req.Code != nil && *req.Code != "" {
		if !organizationCodeRegex.MatchString(*req.Code) {
			return fmt.Errorf("组织代码格式无效，必须为7位数字且首位不可为0")
		}
	}

	// 4. 父组织代码验证（如果提供）
	if req.ParentCode != nil {
		normalizedParent := NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent
		if normalizedParent != nil {
			if !organizationParentCodeRegex.MatchString(*normalizedParent) {
				return fmt.Errorf("父组织代码格式无效，需为0或合法的7位数字编码")
			}
		}
	}

	// 5. 排序顺序验证
	if req.SortOrder < 0 {
		return fmt.Errorf("排序顺序不能为负数")
	}

	if req.SortOrder > 9999 {
		return fmt.Errorf("排序顺序不能超过9999")
	}

	// 6. 描述验证
	if len(req.Description) > types.OrganizationDescriptionMaxLength {
		return fmt.Errorf("描述长度不能超过%d个字符", types.OrganizationDescriptionMaxLength)
	}

	// 7. 时态字段基本验证（不引入 isTemporal）
	if req.EffectiveDate != nil && req.EndDate != nil && req.EffectiveDate.Time.After(req.EndDate.Time) {
		return fmt.Errorf("生效日期不能晚于失效日期")
	}

	return nil
}

// ValidateUpdateOrganization 验证更新组织请求
func ValidateUpdateOrganization(req *types.UpdateOrganizationRequest) error {
	// 1. 名称验证
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return fmt.Errorf("组织名称不能为空")
		}
		if len(*req.Name) > types.OrganizationNameMaxLength {
			return fmt.Errorf("组织名称不能超过%d个字符", types.OrganizationNameMaxLength)
		}
		// 名称格式验证 - 修复正则表达式Unicode转义
		namePattern := regexp.MustCompile(`^[\p{L}\p{N}\s\-]+$`)
		if !namePattern.MatchString(*req.Name) {
			return fmt.Errorf("组织名称包含无效字符，只允许字母、数字、中文、空格和连字符")
		}
	}

	// 2. 组织类型验证
	if req.UnitType != nil {
		if _, ok := validUnitTypes[types.UnitType(*req.UnitType)]; !ok {
			return fmt.Errorf("无效的组织类型: %s", *req.UnitType)
		}
	}

	// 3. 父组织代码验证
	if req.ParentCode != nil {
		normalizedParent := NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent
		if normalizedParent != nil {
			if !organizationParentCodeRegex.MatchString(*normalizedParent) {
				return fmt.Errorf("父组织代码格式无效，需为0或合法的7位数字编码")
			}
		}
	}

	// 4. 排序顺序验证
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return fmt.Errorf("排序顺序不能为负数")
		}
		if *req.SortOrder > 9999 {
			return fmt.Errorf("排序顺序不能超过9999")
		}
	}

	// 5. 状态验证
	if req.Status != nil {
		if _, ok := validStatuses[types.OrganizationStatus(*req.Status)]; !ok {
			return fmt.Errorf("无效的状态: %s", *req.Status)
		}
	}

	// 6. 描述验证
	if req.Description != nil && len(*req.Description) > types.OrganizationDescriptionMaxLength {
		return fmt.Errorf("描述长度不能超过%d个字符", types.OrganizationDescriptionMaxLength)
	}

	// 7. 时态字段基本验证（不引入 isTemporal）
	if req.EffectiveDate != nil && req.EndDate != nil && req.EffectiveDate.Time.After(req.EndDate.Time) {
		return fmt.Errorf("生效日期不能晚于失效日期")
	}

	return nil
}

// ValidateOrganizationCode 验证组织代码格式
func ValidateOrganizationCode(code string) error {
	if code == "" {
		return fmt.Errorf("组织代码不能为空")
	}

	if !organizationCodeRegex.MatchString(code) {
		return fmt.Errorf("组织代码格式无效，必须为7位数字且首位不可为0")
	}

	return nil
}

// ValidateSuspendRequest 验证停用请求
func ValidateSuspendRequest(reason string) error {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return nil
	}

	if len(trimmed) < 5 {
		return fmt.Errorf("停用原因至少需要5个字符")
	}

	if len(trimmed) > 200 {
		return fmt.Errorf("停用原因不能超过200个字符")
	}

	return nil
}

// ValidateActivateRequest 验证激活请求
func ValidateActivateRequest(reason string) error {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return nil
	}

	if len(trimmed) < 5 {
		return fmt.Errorf("激活原因至少需要5个字符")
	}

	if len(trimmed) > 200 {
		return fmt.Errorf("激活原因不能超过200个字符")
	}

	return nil
}

// ValidateCreateVersionRequest 验证创建版本请求 (基于OpenAPI契约v4.4.0)
func ValidateCreateVersionRequest(req *types.CreateVersionRequest) error {
	// 1. 名称验证
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("组织名称不能为空")
	}

	if len(req.Name) > types.OrganizationNameMaxLength {
		return fmt.Errorf("组织名称不能超过%d个字符", types.OrganizationNameMaxLength)
	}

	// 名称格式验证 - 支持Unicode字符
	namePattern := regexp.MustCompile(`^[\p{L}\p{N}\s\-]+$`)
	if !namePattern.MatchString(req.Name) {
		return fmt.Errorf("组织名称包含无效字符，只允许字母、数字、中文、空格和连字符")
	}

	// 2. 组织类型验证
	if req.UnitType == "" {
		return fmt.Errorf("组织类型不能为空")
	}

	if _, ok := validUnitTypes[types.UnitType(req.UnitType)]; !ok {
		return fmt.Errorf("无效的组织类型: %s", req.UnitType)
	}

	// 3. 父组织代码验证（如果提供）
	if req.ParentCode != nil {
		normalizedParent := NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent
		if normalizedParent != nil {
			if !organizationParentCodeRegex.MatchString(*normalizedParent) {
				return fmt.Errorf("父组织代码格式无效，需为0或合法的7位数字编码")
			}
		}
	}

	// 4. 描述验证（如果提供）
	if req.Description != nil && len(*req.Description) > 1000 {
		return fmt.Errorf("描述长度不能超过1000个字符")
	}

	// 5. 排序顺序验证（如果提供）
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			return fmt.Errorf("排序顺序不能为负数")
		}
		if *req.SortOrder > 9999 {
			return fmt.Errorf("排序顺序不能超过9999")
		}
	}

	// 6. 生效日期验证
	if req.EffectiveDate == "" {
		return fmt.Errorf("生效日期不能为空")
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return fmt.Errorf("生效日期格式无效，必须为YYYY-MM-DD格式: %v", err)
	}

	// 7. 失效日期验证（如果提供）
	if req.EndDate != nil && *req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return fmt.Errorf("失效日期格式无效，必须为YYYY-MM-DD格式: %v", err)
		}

		if !endDate.After(effectiveDate) {
			return fmt.Errorf("失效日期必须晚于生效日期")
		}
	}

	// 8. 操作原因验证（可选）
	trimmedReason := strings.TrimSpace(req.OperationReason)
	if trimmedReason != "" {
		if len(trimmedReason) < 5 {
			return fmt.Errorf("操作原因至少需要5个字符")
		}

		if len(trimmedReason) > types.OrganizationOperationReasonMaxLength {
			return fmt.Errorf("操作原因不能超过%d个字符", types.OrganizationOperationReasonMaxLength)
		}
	}

	return nil
}
