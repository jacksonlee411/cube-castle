package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// RESTAPIPermissions 定义 REST 端点与权限映射
var RESTAPIPermissions = map[string]string{
	"POST /api/v1/organization-units":            "WRITE_ORGANIZATION",
	"PUT /api/v1/organization-units/*":           "UPDATE_ORGANIZATION",
	"POST /api/v1/organization-units/*/suspend":  "SUSPEND_ORGANIZATION",
	"POST /api/v1/organization-units/*/activate": "ACTIVATE_ORGANIZATION",
	"POST /api/v1/organization-units/*/events":   "MANAGE_ORGANIZATION_EVENTS",
	"POST /api/v1/organization-units/*/versions": "CREATE_TEMPORAL_VERSION",
	"PUT /api/v1/organization-units/*/history/*": "UPDATE_ORGANIZATION_HISTORY",
	"GET /api/v1/operational/health":             "SYSTEM_MONITOR_READ",
	"GET /api/v1/operational/metrics":            "SYSTEM_MONITOR_READ",
	"GET /api/v1/operational/alerts":             "SYSTEM_MONITOR_READ",
	"GET /api/v1/operational/rate-limit/stats":   "SYSTEM_MONITOR_READ",
	"GET /api/v1/operational/tasks":              "SYSTEM_OPS_READ",
	"GET /api/v1/operational/tasks/status":       "SYSTEM_OPS_READ",
	"POST /api/v1/operational/tasks/*/trigger":   "SYSTEM_OPS_WRITE",
	"POST /api/v1/operational/cutover":           "SYSTEM_OPS_WRITE",
	"POST /api/v1/operational/consistency-check": "SYSTEM_OPS_WRITE",
	"POST /api/v1/job-family-groups":             "job-catalog:write",
	"PUT /api/v1/job-family-groups/*":            "job-catalog:write",
	"POST /api/v1/job-family-groups/*/versions":  "job-catalog:write",
	"POST /api/v1/job-families":                  "job-catalog:write",
	"PUT /api/v1/job-families/*":                 "job-catalog:write",
	"POST /api/v1/job-families/*/versions":       "job-catalog:write",
	"POST /api/v1/job-roles":                     "job-catalog:write",
	"PUT /api/v1/job-roles/*":                    "job-catalog:write",
	"POST /api/v1/job-roles/*/versions":          "job-catalog:write",
	"POST /api/v1/job-levels":                    "job-catalog:write",
	"PUT /api/v1/job-levels/*":                   "job-catalog:write",
	"POST /api/v1/job-levels/*/versions":         "job-catalog:write",
}

// restRolePermissions 定义 REST 角色权限
var restRolePermissions = map[string][]string{
	"ADMIN": {
		"WRITE_ORGANIZATION",
		"UPDATE_ORGANIZATION",
		"SUSPEND_ORGANIZATION",
		"ACTIVATE_ORGANIZATION",
		"MANAGE_ORGANIZATION_EVENTS",
		"CREATE_TEMPORAL_VERSION",
		"UPDATE_ORGANIZATION_HISTORY",
		"SYSTEM_MONITOR_READ",
		"SYSTEM_OPS_READ",
		"SYSTEM_OPS_WRITE",
		"job-catalog:write",
	},
	"MANAGER": {
		"WRITE_ORGANIZATION",
		"UPDATE_ORGANIZATION",
		"SUSPEND_ORGANIZATION",
		"ACTIVATE_ORGANIZATION",
		"job-catalog:write",
	},
	"HR_STAFF": {
		"WRITE_ORGANIZATION",
		"UPDATE_ORGANIZATION",
		"job-catalog:write",
	},
	"EMPLOYEE": {},
	"GUEST":    {},
}

// CheckRESTPermission 按 method/path 检查权限
func (p *PBACPermissionChecker) CheckRESTPermission(ctx context.Context, method, path string) error {
	tenantID := GetTenantID(ctx)
	userID := GetUserID(ctx)
	roles := GetUserRoles(ctx)

	if tenantID == "" || userID == "" {
		return fmt.Errorf("authentication required")
	}

	key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	requiredPermission, found := resolveRESTPermission(key)
	if !found {
		p.logger.Printf("Unknown API endpoint: %s %s", method, path)
		return fmt.Errorf("unknown endpoint: %s %s", method, path)
	}

	if p.checkUserPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	for _, role := range roles {
		if checkRESTRolePermission(role, requiredPermission) {
			p.logger.Printf("Access granted via role %s for %s %s", role, method, path)
			return nil
		}
	}

	if p.checkInheritedPermission(ctx, tenantID, userID, requiredPermission) {
		return nil
	}

	return fmt.Errorf("access denied for: %s %s", method, path)
}

// CheckRESTAPI 检查HTTP请求的权限
func (p *PBACPermissionChecker) CheckRESTAPI(r *http.Request) error {
	return p.CheckRESTPermission(r.Context(), r.Method, r.URL.Path)
}

// MockRESTPermissionCheck 开发模式下允许角色兜底
func (p *PBACPermissionChecker) MockRESTPermissionCheck(ctx context.Context, method, path string) error {
	roles := GetUserRoles(ctx)
	userID := GetUserID(ctx)

	if userID == "admin" || contains(roles, "ADMIN") {
		return nil
	}

	key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	requiredPermission, found := resolveRESTPermission(key)
	if !found {
		return fmt.Errorf("unknown endpoint: %s %s", method, path)
	}

	for _, role := range roles {
		if checkRESTRolePermission(role, requiredPermission) {
			return nil
		}
	}

	return fmt.Errorf("access denied for: %s %s", method, path)
}

func resolveRESTPermission(key string) (string, bool) {
	for pattern, permission := range RESTAPIPermissions {
		if matchRESTPattern(pattern, key) {
			return permission, true
		}
	}
	return "", false
}

func matchRESTPattern(pattern, actual string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == actual
	}
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false
	}
	prefix := parts[0]
	suffix := parts[1]
	return strings.HasPrefix(actual, prefix) && strings.HasSuffix(actual, suffix) && len(actual) >= len(prefix)+len(suffix)
}

func checkRESTRolePermission(role, permission string) bool {
	if permissions, exists := restRolePermissions[role]; exists {
		for _, perm := range permissions {
			if perm == permission {
				return true
			}
		}
	}
	return false
}
