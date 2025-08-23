package utils

import (
	"fmt"
	"strings"

	"organization-command-service/internal/types"
)

// ValidateCreateOrganization 验证创建组织请求
func ValidateCreateOrganization(req *types.CreateOrganizationRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("组织名称不能为空")
	}

	if len(req.Name) > 100 {
		return fmt.Errorf("组织名称不能超过100个字符")
	}

	if req.UnitType == "" {
		return fmt.Errorf("组织类型不能为空")
	}

	validTypes := map[string]bool{
		"ORGANIZATION_UNIT": true, "DEPARTMENT": true, "PROJECT_TEAM": true,
	}
	if !validTypes[req.UnitType] {
		return fmt.Errorf("无效的组织类型: %s", req.UnitType)
	}

	if req.SortOrder < 0 {
		return fmt.Errorf("排序顺序不能为负数")
	}

	// 时态管理验证
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
	}

	return nil
}

// ValidateUpdateOrganization 验证更新组织请求
func ValidateUpdateOrganization(req *types.UpdateOrganizationRequest) error {
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return fmt.Errorf("组织名称不能为空")
		}
		if len(*req.Name) > 100 {
			return fmt.Errorf("组织名称不能超过100个字符")
		}
	}

	if req.UnitType != nil {
		validTypes := map[string]bool{
			"ORGANIZATION_UNIT": true, "DEPARTMENT": true, "PROJECT_TEAM": true,
		}
		if !validTypes[*req.UnitType] {
			return fmt.Errorf("无效的组织类型: %s", *req.UnitType)
		}
	}

	if req.SortOrder != nil && *req.SortOrder < 0 {
		return fmt.Errorf("排序顺序不能为负数")
	}

	// 时态管理验证
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
	}

	return nil
}
