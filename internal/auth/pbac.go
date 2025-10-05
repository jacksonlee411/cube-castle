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

// GraphQL查询 → PBAC scope 映射（与 docs/api/schema.graphql 声明一致）
var GraphQLQueryPermissions = map[string]string{
	// 基础查询
	"organizations":     "org:read",
	"organization":      "org:read",
	"organizationStats": "org:read:stats",

	// 时态查询
	"organizationAtDate":   "org:read:history",
	"organizationHistory":  "org:read:history",
	"organizationVersions": "org:read:history",

	// 层级查询
	"organizationHierarchy": "org:read:hierarchy",
	"organizationSubtree":   "org:read:hierarchy",

	// 审计
	"auditHistory": "org:read:audit",
	"auditLog":     "org:read:audit",
}

// 角色权限预设映射（使用与 GraphQLQueryPermissions 一致的 scope 格式）
var RolePermissions = map[string][]string{
	"ADMIN": {
		"org:read",
		"org:read:history",
		"org:read:hierarchy",
		"org:read:stats",
		"org:read:audit",
		"org:write",
	},
	"MANAGER": {
		"org:read",
		"org:read:history",
		"org:read:hierarchy",
	},
	"EMPLOYEE": {
		"org:read",
	},
	"GUEST": {},
}

// CheckPermission 检查权限的主方法
func (p *PBACPermissionChecker) CheckPermission(ctx context.Context, resource string) error {
	tenantID := GetTenantID(ctx)
	userID := GetUserID(ctx)
	roles := GetUserRoles(ctx)
	scopes := GetUserScopes(ctx)

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

	// 1. PBAC scope 检查
	if hasScope(scopes, requiredPermission) {
		return nil
	}

	// 2. 检查直接用户权限（保留兜底逻辑）
	if p.checkUserPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	// 3. 角色权限（向后兼容开发期）
	for _, role := range roles {
		if p.checkRolePermission(role, requiredPermission) {
			p.logger.Printf("Access granted via role %s for query %s", role, resource)
			return nil
		}
	}

	// 4. 检查继承权限（基于组织层级）
	if p.checkInheritedPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	return fmt.Errorf("access denied for query: %s", resource)
}

// checkUserPermission 检查用户直接权限
func (p *PBACPermissionChecker) checkUserPermission(ctx context.Context, tenantID, userID, permission string) bool {
	// 这里可以查询用户权限表
	// 简化实现：现在只检查是否为系统管理员
	return userID == "admin"
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

func hasScope(scopes []string, required string) bool {
	if required == "" {
		return true
	}
	// 兼容老scope org:write → org:update/org:create，但读取侧基本无需
	for _, s := range scopes {
		if s == required {
			return true
		}
		if s == "org:write" && (required == "org:update" || required == "org:create") {
			return true
		}
	}
	return false
}
