package auth

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"

	pkglogger "cube-castle/pkg/logger"
)

// PBACPermissionChecker 基于策略的访问控制检查器
type PBACPermissionChecker struct {
	db     *sql.DB
	logger pkglogger.Logger
}

//go:embed generated/graphql-permissions.json
var embeddedGraphQLPermissions []byte

// GraphQLQueryPermissions 从生成制品加载 Query→scope 映射（由 docs/api/schema.graphql 注释生成）。
var GraphQLQueryPermissions = func() map[string]string {
	var m map[string]string
	if len(embeddedGraphQLPermissions) > 0 {
		if err := json.Unmarshal(embeddedGraphQLPermissions, &m); err == nil && len(m) > 0 {
			return m
		}
	}
	// 回退：内置映射（仅用于生成制品缺失时的兜底）
	return defaultGraphQLQueryPermissions
}()

// defaultGraphQLQueryPermissions 为构建期生成失败时的最小兜底映射
// 注意：唯一事实来源为 docs/api/schema.graphql 的 "Permissions Required:" 注释；本映射仅作回退使用。
var defaultGraphQLQueryPermissions = map[string]string{
	// 基础查询
	"organizations":     "org:read",
	"organization":      "org:read",
	"organizationStats": "org:read:stats",

	// 时态查询
	"organizationAtDate":   "org:read:history",
	"organizationHistory":  "org:read:history",
	"organizationVersions": "org:read:history",

	// 职位查询
	"positions":               "position:read",
	"position":                "position:read",
	"positionTimeline":        "position:read:history",
	"positionVersions":        "position:read:history",
	"positionAssignments":     "position:read",
	"assignments":             "position:read",
	"assignmentHistory":       "position:read:history",
	"assignmentStats":         "position:read:stats",
	"positionAssignmentAudit": "position:assignments:audit",
	"positionTransfers":       "position:read:history",
	"vacantPositions":         "position:read",
	"positionHeadcountStats":  "position:read:stats",

	// 层级查询
	"organizationHierarchy": "org:read:hierarchy",
	"organizationSubtree":   "org:read:hierarchy",
	"hierarchyStatistics":   "org:read:hierarchy",

	// 审计
	"auditHistory": "org:read:audit",
	"auditLog":     "org:read:audit",
	// 作业目录
	"jobFamilyGroups": "job-catalog:read",
	"jobFamilies":     "job-catalog:read",
	"jobRoles":        "job-catalog:read",
	"jobLevels":       "job-catalog:read",
}

// NewPBACPermissionChecker 返回 PBAC 检查器实例。
func NewPBACPermissionChecker(db *sql.DB, logger pkglogger.Logger) *PBACPermissionChecker {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	return &PBACPermissionChecker{
		db:     db,
		logger: logger,
	}
}

// RolePermissions 提供角色到权限的映射（兜底用途）。
var RolePermissions = map[string][]string{
	"ADMIN": {
		"READ_ORGANIZATION",
		"READ_ORGANIZATION_HISTORY",
		"READ_ORGANIZATION_HIERARCHY",
		"READ_ORGANIZATION_STATISTICS",
		"READ_ORGANIZATION_AUDIT",
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
	scopes := GetUserScopes(ctx)

	// 如果没有认证信息，拒绝访问
	if tenantID == "" || userID == "" {
		return fmt.Errorf("authentication required")
	}

	// 获取查询所需权限
	requiredPermission, exists := GraphQLQueryPermissions[resource]
	if !exists {
		p.logger.WithFields(pkglogger.Fields{"query": resource}).Warn("unknown GraphQL query")
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
			p.logger.WithFields(pkglogger.Fields{"role": role, "query": resource}).Info("access granted via role")
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
func (p *PBACPermissionChecker) checkUserPermission(_ context.Context, _ string, userID, _ string) bool {
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
func (p *PBACPermissionChecker) checkInheritedPermission(_ context.Context, _ string, _ string, _ string) bool {
	// 实现基于组织层级的权限继承
	// 这里可以查询用户所在的组织，然后检查组织层级权限
	// 简化实现：暂时返回false
	return false
}

// CheckGraphQLQuery 检查GraphQL查询权限
func (p *PBACPermissionChecker) CheckGraphQLQuery(ctx context.Context, queryName string) error {
	return p.CheckPermission(ctx, queryName)
}

// MockPermissionCheck 是面向开发环境的简化检查逻辑。
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
	// TODO-TEMPORARY(2025-12-15): 兼容老 scope org:write → org:update/org:create，收敛后移除
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
