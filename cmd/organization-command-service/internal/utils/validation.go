package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"organization-command-service/internal/types"
)

// ValidateCreateOrganization 验证创建组织请求
func ValidateCreateOrganization(req *types.CreateOrganizationRequest) error {
	// 1. 名称验证
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("组织名称不能为空")
	}

	if len(req.Name) < 2 {
		return fmt.Errorf("组织名称长度不能少于2个字符")
	}

	if len(req.Name) > 100 {
		return fmt.Errorf("组织名称不能超过100个字符")
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

	validTypes := map[string]bool{
		"DEPARTMENT": true, "ORGANIZATION_UNIT": true, "PROJECT_TEAM": true,
	}
	if !validTypes[req.UnitType] {
		return fmt.Errorf("无效的组织类型: %s，允许的类型: DEPARTMENT, ORGANIZATION_UNIT, PROJECT_TEAM", req.UnitType)
	}

	// 3. 代码验证（如果提供）
	if req.Code != nil && *req.Code != "" {
		if len(*req.Code) < 3 || len(*req.Code) > 10 {
			return fmt.Errorf("组织代码长度必须在3-10个字符之间")
		}
		// 修复：支持数字开头的代码格式，兼容现有数据
		codePattern := regexp.MustCompile(`^[A-Z0-9][A-Z0-9_]*$`)
		if !codePattern.MatchString(*req.Code) {
			return fmt.Errorf("组织代码格式无效，必须以大写字母或数字开头，只能包含大写字母、数字和下划线")
		}
	}

	// 4. 父组织代码验证（如果提供）
	if req.ParentCode != nil {
		normalizedParent := NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent
		if normalizedParent != nil {
			if len(*normalizedParent) < 3 || len(*normalizedParent) > 10 {
				return fmt.Errorf("父组织代码格式无效，长度必须在3-10个字符之间")
			}
			// 修复：支持数字开头的父组织代码格式，兼容现有数据
			codePattern := regexp.MustCompile(`^[A-Z0-9][A-Z0-9_]*$`)
			if !codePattern.MatchString(*normalizedParent) {
				return fmt.Errorf("父组织代码格式无效，必须以大写字母或数字开头，只能包含大写字母、数字和下划线")
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
	if len(req.Description) > 500 {
		return fmt.Errorf("描述长度不能超过500个字符")
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
		if len(*req.Name) < 2 {
			return fmt.Errorf("组织名称长度不能少于2个字符")
		}
		if len(*req.Name) > 100 {
			return fmt.Errorf("组织名称不能超过100个字符")
		}
		// 名称格式验证 - 修复正则表达式Unicode转义
		namePattern := regexp.MustCompile(`^[\p{L}\p{N}\s\-]+$`)
		if !namePattern.MatchString(*req.Name) {
			return fmt.Errorf("组织名称包含无效字符，只允许字母、数字、中文、空格和连字符")
		}
	}

	// 2. 组织类型验证
	if req.UnitType != nil {
		validTypes := map[string]bool{
			"DEPARTMENT": true, "ORGANIZATION_UNIT": true, "PROJECT_TEAM": true,
		}
		if !validTypes[*req.UnitType] {
			return fmt.Errorf("无效的组织类型: %s，允许的类型: DEPARTMENT, ORGANIZATION_UNIT, PROJECT_TEAM", *req.UnitType)
		}
	}

	// 3. 父组织代码验证
	if req.ParentCode != nil {
		normalizedParent := NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent
		if normalizedParent != nil {
			if len(*normalizedParent) < 3 || len(*normalizedParent) > 10 {
				return fmt.Errorf("父组织代码格式无效，长度必须在3-10个字符之间")
			}
			// 修复：支持数字开头的父组织代码格式，兼容现有数据
			codePattern := regexp.MustCompile(`^[A-Z0-9][A-Z0-9_]*$`)
			if !codePattern.MatchString(*normalizedParent) {
				return fmt.Errorf("父组织代码格式无效，必须以大写字母或数字开头，只能包含大写字母、数字和下划线")
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
		validStatuses := map[string]bool{
			"ACTIVE":   true,
			"INACTIVE": true,
			"PLANNED":  true,
			"DELETED":  true,
		}
		if !validStatuses[*req.Status] {
			return fmt.Errorf("无效的状态: %s，允许的状态: ACTIVE, INACTIVE, PLANNED, DELETED", *req.Status)
		}
	}

	// 6. 描述验证
	if req.Description != nil && len(*req.Description) > 500 {
		return fmt.Errorf("描述长度不能超过500个字符")
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

	if len(code) < 3 || len(code) > 10 {
		return fmt.Errorf("组织代码长度必须在3-10个字符之间")
	}

	// 修复：支持数字开头的组织代码格式，兼容现有数据
	codePattern := regexp.MustCompile(`^[A-Z0-9][A-Z0-9_]*$`)
	if !codePattern.MatchString(code) {
		return fmt.Errorf("组织代码格式无效，必须以大写字母或数字开头，只能包含大写字母、数字和下划线")
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

	if len(req.Name) < 2 {
		return fmt.Errorf("组织名称长度不能少于2个字符")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("组织名称不能超过255个字符")
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

	validTypes := map[string]bool{
		"DEPARTMENT": true, "ORGANIZATION_UNIT": true, "PROJECT_TEAM": true,
	}
	if !validTypes[req.UnitType] {
		return fmt.Errorf("无效的组织类型: %s，允许的类型: DEPARTMENT, ORGANIZATION_UNIT, PROJECT_TEAM", req.UnitType)
	}

	// 3. 父组织代码验证（如果提供）
	if req.ParentCode != nil {
		normalizedParent := NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent
		if normalizedParent != nil {
			if len(*normalizedParent) != 7 {
				return fmt.Errorf("父组织代码长度必须为7个字符")
			}
			// 支持数字开头的组织代码格式，兼容现有数据
			codePattern := regexp.MustCompile(`^[A-Z0-9][A-Z0-9_]*$`)
			if !codePattern.MatchString(*normalizedParent) {
				return fmt.Errorf("父组织代码格式无效，必须以大写字母或数字开头，只能包含大写字母、数字和下划线")
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

		if len(trimmedReason) > 500 {
			return fmt.Errorf("操作原因不能超过500个字符")
		}
	}

	return nil
}
