package utils

import (
	"fmt"
	"regexp"
	"strings"

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

	// 名称格式验证（不能包含特殊字符）
	namePattern := regexp.MustCompile(`^[\w\s\u4e00-\u9fff-]+$`)
	if !namePattern.MatchString(req.Name) {
		return fmt.Errorf("组织名称包含无效字符，只允许字母、数字、中文、空格和连字符")
	}

	// 2. 组织类型验证
	if req.UnitType == "" {
		return fmt.Errorf("组织类型不能为空")
	}

	validTypes := map[string]bool{
		"COMPANY": true, "DEPARTMENT": true, "TEAM": true, "POSITION": true,
	}
	if !validTypes[req.UnitType] {
		return fmt.Errorf("无效的组织类型: %s，允许的类型: COMPANY, DEPARTMENT, TEAM, POSITION", req.UnitType)
	}

	// 3. 代码验证（如果提供）
	if req.Code != nil && *req.Code != "" {
		if len(*req.Code) < 3 || len(*req.Code) > 20 {
			return fmt.Errorf("组织代码长度必须在3-20个字符之间")
		}
		codePattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
		if !codePattern.MatchString(*req.Code) {
			return fmt.Errorf("组织代码格式无效，必须以大写字母开头，只能包含大写字母、数字和下划线")
		}
	}

	// 4. 父组织代码验证（如果提供）
	if req.ParentCode != nil && *req.ParentCode != "" {
		if len(*req.ParentCode) < 3 || len(*req.ParentCode) > 20 {
			return fmt.Errorf("父组织代码格式无效，长度必须在3-20个字符之间")
		}
		codePattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
		if !codePattern.MatchString(*req.ParentCode) {
			return fmt.Errorf("父组织代码格式无效，必须以大写字母开头，只能包含大写字母、数字和下划线")
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

	// 7. 时态管理验证
	if req.IsTemporal {
		if req.EffectiveDate == nil {
			return fmt.Errorf("时态组织必须设置生效日期")
		}
		if req.EndDate != nil && req.EffectiveDate.Time.After(req.EndDate.Time) {
			return fmt.Errorf("生效日期不能晚于失效日期")
		}
		if req.ChangeReason == "" {
			return fmt.Errorf("时态组织必须提供变更原因")
		}
		if len(req.ChangeReason) > 200 {
			return fmt.Errorf("变更原因不能超过200个字符")
		}
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
		// 名称格式验证
		namePattern := regexp.MustCompile(`^[\w\s\u4e00-\u9fff-]+$`)
		if !namePattern.MatchString(*req.Name) {
			return fmt.Errorf("组织名称包含无效字符，只允许字母、数字、中文、空格和连字符")
		}
	}

	// 2. 组织类型验证
	if req.UnitType != nil {
		validTypes := map[string]bool{
			"COMPANY": true, "DEPARTMENT": true, "TEAM": true, "POSITION": true,
		}
		if !validTypes[*req.UnitType] {
			return fmt.Errorf("无效的组织类型: %s，允许的类型: COMPANY, DEPARTMENT, TEAM, POSITION", *req.UnitType)
		}
	}

	// 3. 父组织代码验证
	if req.ParentCode != nil && *req.ParentCode != "" {
		if len(*req.ParentCode) < 3 || len(*req.ParentCode) > 20 {
			return fmt.Errorf("父组织代码格式无效，长度必须在3-20个字符之间")
		}
		codePattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
		if !codePattern.MatchString(*req.ParentCode) {
			return fmt.Errorf("父组织代码格式无效，必须以大写字母开头，只能包含大写字母、数字和下划线")
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
			"ACTIVE": true, "INACTIVE": true, "DELETED": true,
		}
		if !validStatuses[*req.Status] {
			return fmt.Errorf("无效的状态: %s，允许的状态: ACTIVE, INACTIVE, DELETED", *req.Status)
		}
	}

	// 6. 描述验证
	if req.Description != nil && len(*req.Description) > 500 {
		return fmt.Errorf("描述长度不能超过500个字符")
	}

	// 7. 时态管理验证
	if req.IsTemporal != nil && *req.IsTemporal {
		if req.EffectiveDate == nil {
			return fmt.Errorf("启用时态管理时必须设置生效日期")
		}
		if req.EndDate != nil && req.EffectiveDate != nil && req.EffectiveDate.Time.After(req.EndDate.Time) {
			return fmt.Errorf("生效日期不能晚于失效日期")
		}
		if req.ChangeReason == nil || *req.ChangeReason == "" {
			return fmt.Errorf("时态更新必须提供变更原因")
		}
		if req.ChangeReason != nil && len(*req.ChangeReason) > 200 {
			return fmt.Errorf("变更原因不能超过200个字符")
		}
	}

	return nil
}

// ValidateOrganizationCode 验证组织代码格式
func ValidateOrganizationCode(code string) error {
	if code == "" {
		return fmt.Errorf("组织代码不能为空")
	}
	
	if len(code) < 3 || len(code) > 20 {
		return fmt.Errorf("组织代码长度必须在3-20个字符之间")
	}
	
	codePattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	if !codePattern.MatchString(code) {
		return fmt.Errorf("组织代码格式无效，必须以大写字母开头，只能包含大写字母、数字和下划线")
	}
	
	return nil
}

// ValidateSuspendRequest 验证停用请求
func ValidateSuspendRequest(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("停用原因不能为空")
	}
	
	if len(reason) < 5 {
		return fmt.Errorf("停用原因至少需要5个字符")
	}
	
	if len(reason) > 200 {
		return fmt.Errorf("停用原因不能超过200个字符")
	}
	
	return nil
}

// ValidateActivateRequest 验证激活请求
func ValidateActivateRequest(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("激活原因不能为空")
	}
	
	if len(reason) < 5 {
		return fmt.Errorf("激活原因至少需要5个字符")
	}
	
	if len(reason) > 200 {
		return fmt.Errorf("激活原因不能超过200个字符")
	}
	
	return nil
}
