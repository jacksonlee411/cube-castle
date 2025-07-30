package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// StructuredLogger 结构化日志器
type StructuredLogger struct {
	*slog.Logger
}

// NewStructuredLogger 创建新的结构化日志器
func NewStructuredLogger() *StructuredLogger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	return &StructuredLogger{logger}
}

// WithRequestContext 添加请求上下文
func (l *StructuredLogger) WithRequestContext(requestID, userID, tenantID string) *StructuredLogger {
	return &StructuredLogger{
		l.With(
			"request_id", requestID,
			"user_id", userID,
			"tenant_id", tenantID,
		),
	}
}

// WithContext 从context中提取并添加上下文信息
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if userID := ctx.Value("user_id"); userID != nil {
			if tenantID := ctx.Value("tenant_id"); tenantID != nil {
				return l.WithRequestContext(
					requestID.(string),
					userID.(string),
					tenantID.(string),
				)
			}
		}
	}
	return l
}

// === 业务事件日志方法 ===

// LogEmployeeCreated 记录员工创建事件
func (l *StructuredLogger) LogEmployeeCreated(employeeID, tenantID uuid.UUID, employeeNumber string) {
	l.Info("employee_created",
		"event_type", "employee_created",
		"employee_id", employeeID,
		"tenant_id", tenantID,
		"employee_number", employeeNumber,
		"timestamp", time.Now().Unix(),
	)
}

// LogEmployeeUpdated 记录员工更新事件
func (l *StructuredLogger) LogEmployeeUpdated(employeeID, tenantID uuid.UUID, updatedFields map[string]interface{}) {
	fieldsJSON, _ := json.Marshal(updatedFields)
	l.Info("employee_updated",
		"event_type", "employee_updated",
		"employee_id", employeeID,
		"tenant_id", tenantID,
		"updated_fields", string(fieldsJSON),
		"field_count", len(updatedFields),
		"timestamp", time.Now().Unix(),
	)
}

// LogEmployeeDeleted 记录员工删除事件
func (l *StructuredLogger) LogEmployeeDeleted(employeeID, tenantID uuid.UUID, employeeNumber string) {
	l.Info("employee_deleted",
		"event_type", "employee_deleted",
		"employee_id", employeeID,
		"tenant_id", tenantID,
		"employee_number", employeeNumber,
		"timestamp", time.Now().Unix(),
	)
}

// LogOrganizationCreated 记录组织创建事件
func (l *StructuredLogger) LogOrganizationCreated(orgID, tenantID uuid.UUID, orgName, orgCode string) {
	l.Info("organization_created",
		"event_type", "organization_created",
		"organization_id", orgID,
		"tenant_id", tenantID,
		"organization_name", orgName,
		"organization_code", orgCode,
		"timestamp", time.Now().Unix(),
	)
}

// LogAIRequest 记录AI请求处理事件
func (l *StructuredLogger) LogAIRequest(sessionID, intent string, processingTime time.Duration, success bool) {
	l.Info("ai_request_processed",
		"event_type", "ai_request_processed",
		"session_id", sessionID,
		"intent", intent,
		"processing_time_ms", processingTime.Milliseconds(),
		"success", success,
		"timestamp", time.Now().Unix(),
	)
}

// LogDatabaseOperation 记录数据库操作事件
func (l *StructuredLogger) LogDatabaseOperation(operation, table string, recordCount int, duration time.Duration, success bool) {
	l.Info("database_operation",
		"event_type", "database_operation",
		"operation", operation,
		"table", table,
		"record_count", recordCount,
		"duration_ms", duration.Milliseconds(),
		"success", success,
		"timestamp", time.Now().Unix(),
	)
}

// LogAPIRequest 记录API请求事件
func (l *StructuredLogger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, userAgent string) {
	l.Info("api_request",
		"event_type", "api_request",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"user_agent", userAgent,
		"timestamp", time.Now().Unix(),
	)
}

// LogAuthEvent 记录认证事件
func (l *StructuredLogger) LogAuthEvent(eventType, userID, tenantID string, success bool, reason string) {
	l.Info("auth_event",
		"event_type", "auth_event",
		"auth_type", eventType,
		"user_id", userID,
		"tenant_id", tenantID,
		"success", success,
		"reason", reason,
		"timestamp", time.Now().Unix(),
	)
}

// LogOutboxEvent 记录发件箱事件处理
func (l *StructuredLogger) LogOutboxEvent(eventID uuid.UUID, eventType string, aggregateID uuid.UUID, success bool, retryCount int) {
	l.Info("outbox_event_processed",
		"event_type", "outbox_event_processed",
		"event_id", eventID,
		"outbox_event_type", eventType,
		"aggregate_id", aggregateID,
		"success", success,
		"retry_count", retryCount,
		"timestamp", time.Now().Unix(),
	)
}

// LogWorkflowEvent 记录工作流事件
func (l *StructuredLogger) LogWorkflowEvent(workflowID, workflowType, status string, duration time.Duration) {
	l.Info("workflow_event",
		"event_type", "workflow_event",
		"workflow_id", workflowID,
		"workflow_type", workflowType,
		"status", status,
		"duration_ms", duration.Milliseconds(),
		"timestamp", time.Now().Unix(),
	)
}

// === 系统事件日志方法 ===

// LogError 记录错误事件
func (l *StructuredLogger) LogError(errorType, message string, err error, context map[string]interface{}) {
	contextJSON, _ := json.Marshal(context)

	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}

	l.Error("error_occurred",
		"event_type", "error_occurred",
		"error_type", errorType,
		"message", message,
		"error", errorStr,
		"context", string(contextJSON),
		"timestamp", time.Now().Unix(),
	)
}

// LogPerformanceMetric 记录性能指标
func (l *StructuredLogger) LogPerformanceMetric(metricName string, value float64, unit string, tags map[string]string) {
	tagsJSON, _ := json.Marshal(tags)

	l.Info("performance_metric",
		"event_type", "performance_metric",
		"metric_name", metricName,
		"value", value,
		"unit", unit,
		"tags", string(tagsJSON),
		"timestamp", time.Now().Unix(),
	)
}

// LogHealthCheck 记录健康检查事件
func (l *StructuredLogger) LogHealthCheck(component string, status string, checkDuration time.Duration, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)

	l.Info("health_check",
		"event_type", "health_check",
		"component", component,
		"status", status,
		"check_duration_ms", checkDuration.Milliseconds(),
		"details", string(detailsJSON),
		"timestamp", time.Now().Unix(),
	)
}

// === 安全事件日志方法 ===

// LogSecurityEvent 记录安全事件
func (l *StructuredLogger) LogSecurityEvent(eventType, userID, sourceIP, description string, severity string) {
	l.Warn("security_event",
		"event_type", "security_event",
		"security_event_type", eventType,
		"user_id", userID,
		"source_ip", sourceIP,
		"description", description,
		"severity", severity,
		"timestamp", time.Now().Unix(),
	)
}

// LogAccessAttempt 记录访问尝试
func (l *StructuredLogger) LogAccessAttempt(userID, resource, action string, allowed bool, reason string) {
	level := slog.LevelInfo
	if !allowed {
		level = slog.LevelWarn
	}

	l.Log(context.Background(), level, "access_attempt",
		"event_type", "access_attempt",
		"user_id", userID,
		"resource", resource,
		"action", action,
		"allowed", allowed,
		"reason", reason,
		"timestamp", time.Now().Unix(),
	)
}

// === 调试和诊断方法 ===

// LogDebug 记录调试信息
func (l *StructuredLogger) LogDebug(component, message string, data map[string]interface{}) {
	dataJSON, _ := json.Marshal(data)

	l.Debug("debug_info",
		"event_type", "debug_info",
		"component", component,
		"message", message,
		"data", string(dataJSON),
		"timestamp", time.Now().Unix(),
	)
}

// LogServiceStartup 记录服务启动事件
func (l *StructuredLogger) LogServiceStartup(serviceName, version string, config map[string]interface{}) {
	configJSON, _ := json.Marshal(config)

	l.Info("service_startup",
		"event_type", "service_startup",
		"service_name", serviceName,
		"version", version,
		"config", string(configJSON),
		"timestamp", time.Now().Unix(),
	)
}

// LogServiceShutdown 记录服务关闭事件
func (l *StructuredLogger) LogServiceShutdown(serviceName string, reason string, uptime time.Duration) {
	l.Info("service_shutdown",
		"event_type", "service_shutdown",
		"service_name", serviceName,
		"reason", reason,
		"uptime_seconds", uptime.Seconds(),
		"timestamp", time.Now().Unix(),
	)
}
