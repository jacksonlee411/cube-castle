package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// PBACPermissionChecker 基于策略的访问控制检查器
type PBACPermissionChecker struct {
	db     *sql.DB
	logger *log.Logger
}

func NewPBACPermissionChecker(db *sql.DB, logger *log.Logger) *PBACPermissionChecker {
	return &PBACPermissionChecker{
		db:     db,
		logger: logger,
	}
}

// GraphQL查询权限映射表
var GraphQLQueryPermissions = map[string]string{
	"organizations":             "READ_ORGANIZATION",
	"organization":              "READ_ORGANIZATION", 
	"organizationHistory":       "READ_ORGANIZATION_HISTORY",
	"organizationHierarchy":     "READ_ORGANIZATION_HIERARCHY",
	"organizationStatistics":    "READ_ORGANIZATION_STATISTICS",
	"organizationVersions":      "READ_ORGANIZATION_HISTORY",
	"organizationsByDateRange":  "READ_ORGANIZATION_HISTORY",
	"organizationsByParent":     "READ_ORGANIZATION",
	"organizationsByType":       "READ_ORGANIZATION",
	"organizationsByStatus":     "READ_ORGANIZATION",
}

// 角色权限预设映射
var RolePermissions = map[string][]string{
	"ADMIN": {
		"READ_ORGANIZATION",
		"READ_ORGANIZATION_HISTORY", 
		"READ_ORGANIZATION_HIERARCHY",
		"READ_ORGANIZATION_STATISTICS",
		"WRITE_ORGANIZATION",
	},
	"MANAGER": {
		"READ_ORGANIZATION",
		"READ_ORGANIZATION_HISTORY",
		"READ_ORGANIZATION_HIERARCHY",
	},
	"EMPLOYEE": {
		"READ_ORGANIZATION",
	},
	"GUEST": {},
}

// CheckPermission 检查权限的主方法
func (p *PBACPermissionChecker) CheckPermission(ctx context.Context, resource string) error {
	tenantID := GetTenantID(ctx)
	userID := GetUserID(ctx)
	roles := GetUserRoles(ctx)

	// 如果没有认证信息，拒绝访问
	if tenantID == "" || userID == "" {
		return fmt.Errorf("authentication required")
	}

	// 获取查询所需权限
	requiredPermission, exists := GraphQLQueryPermissions[resource]
	if !exists {
		p.logger.Printf("Unknown GraphQL query: %s", resource)
		return fmt.Errorf("unknown query: %s", resource)
	}

	// 1. 检查直接用户权限
	if p.checkUserPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	// 2. 检查角色权限
	for _, role := range roles {
		if p.checkRolePermission(role, requiredPermission) {
			p.logger.Printf("Access granted via role %s for query %s", role, resource)
			return nil
		}
	}

	// 3. 检查继承权限（基于组织层级）
	if p.checkInheritedPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	return fmt.Errorf("access denied for query: %s", resource)
}

// checkUserPermission 检查用户直接权限
func (p *PBACPermissionChecker) checkUserPermission(ctx context.Context, tenantID, userID, permission string) bool {
	// 这里可以查询用户权限表
	// 简化实现：现在只检查是否为系统管理员
	if userID == "admin" {
		return true
	}
	return false
}

// checkRolePermission 检查角色权限
func (p *PBACPermissionChecker) checkRolePermission(role, permission string) bool {
	if permissions, exists := RolePermissions[role]; exists {
		for _, perm := range permissions {
			if perm == permission {
				return true
			}
		}
	}
	return false
}

// checkInheritedPermission 检查继承权限
func (p *PBACPermissionChecker) checkInheritedPermission(ctx context.Context, tenantID, userID, permission string) bool {
	// 实现基于组织层级的权限继承
	// 这里可以查询用户所在的组织，然后检查组织层级权限
	// 简化实现：暂时返回false
	return false
}

// CheckGraphQLQuery 检查GraphQL查询权限
func (p *PBACPermissionChecker) CheckGraphQLQuery(ctx context.Context, queryName string) error {
	return p.CheckPermission(ctx, queryName)
}

// 模拟权限检查（用于开发和测试）
func (p *PBACPermissionChecker) MockPermissionCheck(ctx context.Context, queryName string) error {
	roles := GetUserRoles(ctx)
	userID := GetUserID(ctx)
	
	// 开发模式：管理员用户直接通过
	if userID == "admin" || contains(roles, "ADMIN") {
		return nil
	}
	
	// 检查角色权限
	requiredPermission, exists := GraphQLQueryPermissions[queryName]
	if !exists {
		return fmt.Errorf("unknown query: %s", queryName)
	}
	
	for _, role := range roles {
		if p.checkRolePermission(role, requiredPermission) {
			return nil
		}
	}
	
	return fmt.Errorf("access denied for query: %s", queryName)
}

// contains 检查字符串数组是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}