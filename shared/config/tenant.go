package config

import "github.com/google/uuid"

// 默认租户配置 - 项目级别统一配置
const (
	// DefaultTenantIDString 默认租户ID字符串
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	
	// DefaultTenantName 默认租户名称
	DefaultTenantName = "高谷集团"
	
	// DefaultTenantCode 默认租户代码
	DefaultTenantCode = "GAOYAGU"
)

var (
	// DefaultTenantID 默认租户UUID
	DefaultTenantID = uuid.MustParse(DefaultTenantIDString)
)

// GetDefaultTenantID 获取默认租户ID
func GetDefaultTenantID() uuid.UUID {
	return DefaultTenantID
}

// GetDefaultTenantIDString 获取默认租户ID字符串
func GetDefaultTenantIDString() string {
	return DefaultTenantIDString
}

// IsDefaultTenant 检查是否为默认租户
func IsDefaultTenant(tenantID uuid.UUID) bool {
	return tenantID == DefaultTenantID
}

// IsDefaultTenantString 检查字符串是否为默认租户ID
func IsDefaultTenantString(tenantIDStr string) bool {
	return tenantIDStr == DefaultTenantIDString
}