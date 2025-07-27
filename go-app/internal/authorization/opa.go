package authorization

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
)

// OPAAuthorizer OPA授权器
type OPAAuthorizer struct {
	logger  *logging.StructuredLogger
	queries map[string]rego.PreparedEvalQuery
}

// AuthorizationInput 授权输入
type AuthorizationInput struct {
	UserID     string                 `json:"user_id"`
	TenantID   string                 `json:"tenant_id"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	User       UserInfo               `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string   `json:"id"`
	Role     string   `json:"role"`
	TenantID string   `json:"tenant_id"`
	Groups   []string `json:"groups"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// AuthorizationResult 授权结果
type AuthorizationResult struct {
	Allowed bool     `json:"allowed"`
	Reason  string   `json:"reason"`
	Errors  []string `json:"errors,omitempty"`
}

// NewOPAAuthorizer 创建新的OPA授权器
func NewOPAAuthorizer(logger *logging.StructuredLogger) (*OPAAuthorizer, error) {
	authorizer := &OPAAuthorizer{
		logger:  logger,
		queries: make(map[string]rego.PreparedEvalQuery),
	}

	// 加载授权策略
	err := authorizer.loadPolicies()
	if err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	logger.Info("OPA授权器初始化成功")
	return authorizer, nil
}

// loadPolicies 加载授权策略
func (o *OPAAuthorizer) loadPolicies() error {
	policies := map[string]string{
		"corehr":       coreHRPolicy,
		"admin":        adminPolicy,
		"tenant":       tenantPolicy,
		"workflow":     workflowPolicy,
		"intelligence": intelligencePolicy,
	}

	for name, policy := range policies {
		query, err := rego.New(
			rego.Query("data."+name+".allow"),
			rego.Module(name+".rego", policy),
		).PrepareForEval(context.Background())

		if err != nil {
			return fmt.Errorf("failed to prepare policy %s: %w", name, err)
		}

		o.queries[name] = query
		o.logger.Info("已加载授权策略", "policy", name)
	}

	return nil
}

// Authorize 执行授权检查
func (o *OPAAuthorizer) Authorize(ctx context.Context, input AuthorizationInput) (*AuthorizationResult, error) {
	start := time.Now()
	
	o.logger.LogAccessAttempt(input.UserID, input.Resource, input.Action, false, "checking")

	// 确定使用哪个策略
	policyName := o.determinePolicyForResource(input.Resource)
	
	query, exists := o.queries[policyName]
	if !exists {
		return &AuthorizationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("No policy found for resource: %s", input.Resource),
		}, nil
	}

	// 执行策略评估
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		metrics.RecordError("authorization", "policy_eval_error")
		o.logger.LogError("authorization", "Policy evaluation failed", err, map[string]interface{}{
			"user_id":  input.UserID,
			"resource": input.Resource,
			"action":   input.Action,
			"policy":   policyName,
		})
		return nil, fmt.Errorf("policy evaluation failed: %w", err)
	}

	// 解析结果
	allowed := false
	reason := "Access denied by policy"

	if len(results) > 0 && len(results[0].Expressions) > 0 {
		if result, ok := results[0].Expressions[0].Value.(bool); ok {
			allowed = result
			if allowed {
				reason = "Access granted by policy"
			}
		}
	}

	// 记录授权结果
	duration := time.Since(start)
	o.logger.LogAccessAttempt(input.UserID, input.Resource, input.Action, allowed, reason)
	
	if allowed {
		metrics.RecordAIRequest("authorization_success", "success", duration)
	} else {
		metrics.RecordAIRequest("authorization_denied", "denied", duration)
	}

	return &AuthorizationResult{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}

// AuthorizeHTTPRequest 为HTTP请求执行授权检查
func (o *OPAAuthorizer) AuthorizeHTTPRequest(ctx context.Context, userID, tenantID, method, path string, user UserInfo) (*AuthorizationResult, error) {
	// 从HTTP方法和路径解析资源和动作
	resource, action := o.parseHTTPRequest(method, path)
	
	input := AuthorizationInput{
		UserID:   userID,
		TenantID: tenantID,
		Resource: resource,
		Action:   action,
		User:     user,
		Context: map[string]interface{}{
			"method": method,
			"path":   path,
		},
	}

	return o.Authorize(ctx, input)
}

// determinePolicyForResource 根据资源确定使用的策略
func (o *OPAAuthorizer) determinePolicyForResource(resource string) string {
	switch {
	case strings.HasPrefix(resource, "employee") || strings.HasPrefix(resource, "organization"):
		return "corehr"
	case strings.HasPrefix(resource, "workflow"):
		return "workflow"
	case strings.HasPrefix(resource, "intelligence"):
		return "intelligence"
	case strings.HasPrefix(resource, "admin"):
		return "admin"
	default:
		return "tenant" // 默认使用租户策略
	}
}

// parseHTTPRequest 解析HTTP请求到资源和动作
func (o *OPAAuthorizer) parseHTTPRequest(method, path string) (resource, action string) {
	// 简化的路径解析逻辑
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	
	// 移除API版本前缀
	if len(pathParts) > 2 && pathParts[0] == "api" {
		pathParts = pathParts[2:] // 跳过 "api/v1"
	}

	if len(pathParts) == 0 {
		return "unknown", "unknown"
	}

	// 根据HTTP方法映射动作
	switch method {
	case "GET":
		action = "read"
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = "access"
	}

	// 提取资源名称
	if len(pathParts) >= 2 {
		module := pathParts[0]     // corehr, intelligence, etc.
		resource = pathParts[1]    // employees, organizations, etc.
		
		// 组合模块和资源
		return fmt.Sprintf("%s:%s", module, resource), action
	} else if len(pathParts) == 1 {
		return pathParts[0], action
	}

	return "unknown", action
}

// ValidateUser 验证用户信息
func (o *OPAAuthorizer) ValidateUser(ctx context.Context, userID, tenantID string) (*UserInfo, error) {
	// 这里应该从用户服务或数据库获取用户信息
	// 简化实现，返回基础用户信息
	
	user := &UserInfo{
		ID:       userID,
		TenantID: tenantID,
		Role:     o.determineUserRole(userID), // 简化的角色确定逻辑
		Groups:   []string{},
		Attributes: map[string]interface{}{
			"verified": true,
		},
	}

	return user, nil
}

// determineUserRole 确定用户角色（简化实现）
func (o *OPAAuthorizer) determineUserRole(userID string) string {
	// 这里应该从数据库或用户服务查询真实角色
	// 简化实现：根据用户ID模式返回角色
	switch {
	case strings.Contains(userID, "admin"):
		return "admin"
	case strings.Contains(userID, "hr"):
		return "hr"
	case strings.Contains(userID, "manager"):
		return "manager"
	default:
		return "employee"
	}
}

// ReloadPolicies 重新加载策略
func (o *OPAAuthorizer) ReloadPolicies() error {
	o.logger.Info("重新加载授权策略")
	return o.loadPolicies()
}

// GetPolicyDecision 获取策略决策详情（用于调试）
func (o *OPAAuthorizer) GetPolicyDecision(ctx context.Context, input AuthorizationInput) (map[string]interface{}, error) {
	policyName := o.determinePolicyForResource(input.Resource)
	
	query, exists := o.queries[policyName]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyName)
	}

	// 获取完整的评估结果
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, err
	}

	// 返回详细结果用于调试
	return map[string]interface{}{
		"policy":  policyName,
		"input":   input,
		"results": results,
	}, nil
}

// 策略定义

const coreHRPolicy = `
package corehr

import rego.v1

# 默认拒绝访问
default allow := false

# 管理员可以进行所有操作
allow if {
    input.user.role == "admin"
}

# HR可以管理员工和组织信息
allow if {
    input.user.role == "hr"
    input.resource in ["corehr:employees", "corehr:organizations"]
    input.action in ["create", "read", "update"]
}

# 经理可以查看和更新自己团队的员工信息
allow if {
    input.user.role == "manager"
    input.resource == "corehr:employees"
    input.action in ["read", "update"]
    # 在实际实现中，这里应该检查员工是否属于该经理的团队
}

# 员工可以查看自己的信息
allow if {
    input.user.role == "employee"
    input.resource == "corehr:employees"
    input.action == "read"
    input.resource_id == input.user.id
}

# 所有用户都可以查看组织架构
allow if {
    input.resource == "corehr:organizations"
    input.action == "read"
}

# 租户隔离：只能访问自己租户的数据
allow if {
    input.tenant_id == input.user.tenant_id
    basic_access_allowed
}

basic_access_allowed if {
    input.action == "read"
    input.resource in ["corehr:employees", "corehr:organizations"]
}
`

const adminPolicy = `
package admin

import rego.v1

default allow := false

# 只有管理员可以访问管理功能
allow if {
    input.user.role == "admin"
    input.resource == "admin"
}

# 管理员可以访问所有监控和管理端点
allow if {
    input.user.role == "admin"
    startswith(input.resource, "admin:")
}
`

const tenantPolicy = `
package tenant

import rego.v1

default allow := false

# 基本的租户隔离策略
allow if {
    input.tenant_id == input.user.tenant_id
    input.user.role in ["admin", "hr", "manager", "employee"]
}

# 跨租户访问需要特殊权限
allow if {
    input.user.role == "admin"
    "cross_tenant_access" in input.user.groups
}
`

const workflowPolicy = `
package workflow

import rego.v1

default allow := false

# 管理员可以管理所有工作流
allow if {
    input.user.role == "admin"
    input.resource == "workflow"
}

# HR可以启动员工相关工作流
allow if {
    input.user.role == "hr"
    input.resource == "workflow"
    input.action in ["create", "read"]
    startswith(input.context.workflow_type, "employee")
}

# 经理可以审批工作流
allow if {
    input.user.role == "manager"
    input.resource == "workflow"
    input.action == "update"
    input.context.workflow_type == "approval"
}

# 员工可以查看自己相关的工作流
allow if {
    input.user.role == "employee"
    input.resource == "workflow"
    input.action == "read"
    input.context.employee_id == input.user.id
}
`

const intelligencePolicy = `
package intelligence

import rego.v1

default allow := false

# 所有认证用户都可以使用AI服务
allow if {
    input.resource == "intelligence"
    input.action in ["create", "read"]
    input.user.role in ["admin", "hr", "manager", "employee"]
}

# 管理员可以管理AI服务
allow if {
    input.user.role == "admin"
    input.resource == "intelligence"
}

# 租户隔离：只能访问自己租户的AI会话
allow if {
    input.resource == "intelligence"
    input.tenant_id == input.user.tenant_id
}
`