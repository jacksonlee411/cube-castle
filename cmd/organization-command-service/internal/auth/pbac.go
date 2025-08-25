package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// REST API权限映射表
var RESTAPIPermissions = map[string]map[string]string{
	"POST /api/v1/organization-units":                    {"method": "POST", "permission": "WRITE_ORGANIZATION"},
	"PUT /api/v1/organization-units/*":                   {"method": "PUT", "permission": "UPDATE_ORGANIZATION"},
	"DELETE /api/v1/organization-units/*":                {"method": "DELETE", "permission": "DELETE_ORGANIZATION"},
	"POST /api/v1/organization-units/*/suspend":          {"method": "POST", "permission": "SUSPEND_ORGANIZATION"},
	"POST /api/v1/organization-units/*/activate":         {"method": "POST", "permission": "ACTIVATE_ORGANIZATION"},
	"POST /api/v1/organization-units/*/events":           {"method": "POST", "permission": "MANAGE_ORGANIZATION_EVENTS"},
	"PUT /api/v1/organization-units/*/history/*":         {"method": "PUT", "permission": "UPDATE_ORGANIZATION_HISTORY"},
}

// 角色权限预设映射
var RolePermissions = map[string][]string{
	"ADMIN": {
		"WRITE_ORGANIZATION",
		"UPDATE_ORGANIZATION",
		"DELETE_ORGANIZATION",
		"SUSPEND_ORGANIZATION",
		"ACTIVATE_ORGANIZATION",
		"MANAGE_ORGANIZATION_EVENTS",
		"UPDATE_ORGANIZATION_HISTORY",
	},
	"MANAGER": {
		"WRITE_ORGANIZATION",
		"UPDATE_ORGANIZATION",
		"SUSPEND_ORGANIZATION",
		"ACTIVATE_ORGANIZATION",
	},
	"HR_STAFF": {
		"WRITE_ORGANIZATION",
		"UPDATE_ORGANIZATION",
	},
	"EMPLOYEE": {},
	"GUEST":    {},
}

// CheckPermission 检查权限的主方法
func (p *PBACPermissionChecker) CheckPermission(ctx context.Context, method, path string) error {
	tenantID := GetTenantID(ctx)
	userID := GetUserID(ctx)
	roles := GetUserRoles(ctx)

	// 如果没有认证信息，拒绝访问
	if tenantID == "" || userID == "" {
		return fmt.Errorf("authentication required")
	}

	// 构建权限检查key
	permissionKey := fmt.Sprintf("%s %s", method, path)
	
	// 查找匹配的权限映射
	var requiredPermission string
	found := false
	
	for pattern, config := range RESTAPIPermissions {
		if p.matchPattern(pattern, permissionKey) {
			requiredPermission = config["permission"]
			found = true
			break
		}
	}

	if !found {
		p.logger.Printf("Unknown API endpoint: %s %s", method, path)
		return fmt.Errorf("unknown endpoint: %s %s", method, path)
	}

	// 1. 检查直接用户权限
	if p.checkUserPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	// 2. 检查角色权限
	for _, role := range roles {
		if p.checkRolePermission(role, requiredPermission) {
			p.logger.Printf("Access granted via role %s for %s %s", role, method, path)
			return nil
		}
	}

	// 3. 检查继承权限（基于组织层级）
	if p.checkInheritedPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	return fmt.Errorf("access denied for: %s %s", method, path)
}

// matchPattern 简单的模式匹配，支持*通配符
func (p *PBACPermissionChecker) matchPattern(pattern, actual string) bool {
	// 简单实现：替换*为任意匹配
	if strings.Contains(pattern, "*") {
		// 更精确的模式匹配可以使用正则表达式
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(actual, prefix)
	}
	return pattern == actual
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

// CheckRESTAPI 检查REST API权限
func (p *PBACPermissionChecker) CheckRESTAPI(r *http.Request) error {
	return p.CheckPermission(r.Context(), r.Method, r.URL.Path)
}

// 模拟权限检查（用于开发和测试）
func (p *PBACPermissionChecker) MockPermissionCheck(ctx context.Context, method, path string) error {
	roles := GetUserRoles(ctx)
	userID := GetUserID(ctx)

	// 开发模式：管理员用户直接通过
	if userID == "admin" || contains(roles, "ADMIN") {
		return nil
	}

	// 检查角色权限
	permissionKey := fmt.Sprintf("%s %s", method, path)
	
	for pattern, config := range RESTAPIPermissions {
		if p.matchPattern(pattern, permissionKey) {
			requiredPermission := config["permission"]
			
			for _, role := range roles {
				if p.checkRolePermission(role, requiredPermission) {
					return nil
				}
			}
			
			return fmt.Errorf("access denied for: %s %s", method, path)
		}
	}

	return fmt.Errorf("unknown endpoint: %s %s", method, path)
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