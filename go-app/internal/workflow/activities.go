package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
	"go.temporal.io/sdk/activity"
)

// Activities 活动处理器
type Activities struct {
	logger *logging.StructuredLogger
}

// NewActivities 创建新的活动处理器
func NewActivities(logger *logging.StructuredLogger) *Activities {
	return &Activities{
		logger: logger,
	}
}

// === 员工入职相关活动 ===

// CreateEmployeeAccountActivity 创建员工账户活动
func (a *Activities) CreateEmployeeAccountActivity(ctx context.Context, req CreateAccountRequest) (*CreateAccountResult, error) {
	start := time.Now()
	logger := activity.GetLogger(ctx)

	logger.Info("Creating employee account",
		"employee_id", req.EmployeeID,
		"email", req.Email)

	// 模拟账户创建过程
	// 在实际实现中，这里会调用身份认证系统的API

	// 模拟处理时间
	time.Sleep(2 * time.Second)

	// 记录业务日志
	a.logger.LogEmployeeCreated(req.EmployeeID, req.TenantID, req.Email)

	// 记录指标
	duration := time.Since(start)
	metrics.RecordDatabaseOperation("CREATE", "user_accounts", "success", duration)

	result := &CreateAccountResult{
		AccountID: fmt.Sprintf("acc_%s", req.EmployeeID.String()[:8]),
		Success:   true,
	}

	logger.Info("Employee account created successfully",
		"employee_id", req.EmployeeID,
		"account_id", result.AccountID)

	return result, nil
}

// AssignEquipmentAndPermissionsActivity 分配设备和权限活动
func (a *Activities) AssignEquipmentAndPermissionsActivity(ctx context.Context, req AssignEquipmentRequest) (*AssignEquipmentResult, error) {
	start := time.Now()
	logger := activity.GetLogger(ctx)

	logger.Info("Assigning equipment and permissions",
		"employee_id", req.EmployeeID,
		"department", req.Department,
		"position", req.Position)

	// 模拟设备分配过程
	// 在实际实现中，这里会调用设备管理系统和权限系统的API

	// 根据部门和职位确定设备清单
	var assignedItems []string

	switch req.Department {
	case "技术部", "Technology":
		assignedItems = append(assignedItems, "laptop", "monitor", "keyboard", "mouse")
		if req.Position == "Senior Developer" || req.Position == "Tech Lead" {
			assignedItems = append(assignedItems, "additional_monitor")
		}
	case "销售部", "Sales":
		assignedItems = append(assignedItems, "laptop", "mobile_phone")
	case "人事部", "HR":
		assignedItems = append(assignedItems, "laptop", "printer_access")
	default:
		assignedItems = append(assignedItems, "laptop")
	}

	// 模拟处理时间
	time.Sleep(3 * time.Second)

	// 记录指标
	duration := time.Since(start)
	metrics.RecordDatabaseOperation("INSERT", "equipment_assignments", "success", duration)

	result := &AssignEquipmentResult{
		AssignedItems: assignedItems,
		Success:       true,
	}

	logger.Info("Equipment and permissions assigned successfully",
		"employee_id", req.EmployeeID,
		"assigned_items", assignedItems)

	return result, nil
}

// SendWelcomeEmailActivity 发送欢迎邮件活动
func (a *Activities) SendWelcomeEmailActivity(ctx context.Context, req WelcomeEmailRequest) (*SendEmailResult, error) {
	start := time.Now()
	logger := activity.GetLogger(ctx)

	logger.Info("Sending welcome email",
		"employee_id", req.EmployeeID,
		"email", req.Email,
		"first_name", req.FirstName)

	// 构造欢迎邮件内容
	emailContent := fmt.Sprintf(`
Dear %s,

Welcome to Cube Castle! We are excited to have you join our team in the %s department.

Your start date is %s. Please arrive at 9:00 AM for your orientation session.

If you have any questions before your start date, please don't hesitate to reach out to HR.

Best regards,
The Cube Castle Team
	`, req.FirstName, req.Department, req.StartDate.Format("January 2, 2006"))

	// 模拟邮件发送过程
	// 在实际实现中，这里会调用邮件服务的API
	time.Sleep(1 * time.Second)

	// 记录指标
	duration := time.Since(start)
	metrics.RecordDatabaseOperation("INSERT", "email_logs", "success", duration)

	result := &SendEmailResult{
		MessageID: fmt.Sprintf("msg_%d", time.Now().Unix()),
		Success:   true,
	}

	logger.Info("Welcome email sent successfully",
		"employee_id", req.EmployeeID,
		"message_id", result.MessageID)

	// 记录到业务日志
	a.logger.LogDebug("welcome_email", "Email sent to new employee", map[string]interface{}{
		"employee_id":     req.EmployeeID,
		"email":           req.Email,
		"message_id":      result.MessageID,
		"content_preview": emailContent[:100] + "...",
	})

	return result, nil
}

// NotifyManagerActivity 通知经理活动
func (a *Activities) NotifyManagerActivity(ctx context.Context, req NotifyManagerRequest) (*NotifyManagerResult, error) {
	start := time.Now()
	logger := activity.GetLogger(ctx)

	logger.Info("Notifying manager",
		"manager_id", req.ManagerID,
		"new_employee_id", req.NewEmployeeID,
		"employee_name", req.EmployeeName)

	// 构造通知内容
	_ = fmt.Sprintf(`
Hello,

A new employee, %s, will be joining your team in the %s department as a %s starting on %s.

Please prepare for their arrival and consider scheduling a welcome meeting.

Employee ID: %s

Best regards,
HR Team
	`, req.EmployeeName, req.Department, req.Position,
		req.StartDate.Format("January 2, 2006"), req.NewEmployeeID)

	// 模拟通知发送过程
	// 在实际实现中，这里会调用通知系统的API
	time.Sleep(1 * time.Second)

	// 记录指标
	duration := time.Since(start)
	metrics.RecordDatabaseOperation("INSERT", "notifications", "success", duration)

	result := &NotifyManagerResult{
		NotificationID: fmt.Sprintf("notif_%d", time.Now().Unix()),
		Success:        true,
	}

	logger.Info("Manager notification sent successfully",
		"manager_id", req.ManagerID,
		"notification_id", result.NotificationID)

	return result, nil
}

// === 休假审批相关活动 ===

// ValidateLeaveRequestActivity 验证休假请求活动
func (a *Activities) ValidateLeaveRequestActivity(ctx context.Context, req ValidateLeaveRequestRequest) (*ValidateLeaveRequestResult, error) {
	start := time.Now()
	logger := activity.GetLogger(ctx)

	logger.Info("Validating leave request",
		"request_id", req.RequestID,
		"employee_id", req.EmployeeID,
		"leave_type", req.LeaveType,
		"start_date", req.StartDate,
		"end_date", req.EndDate)

	// 执行各种验证
	result := &ValidateLeaveRequestResult{
		IsValid: true,
		Reason:  "",
	}

	// 验证1：检查日期有效性
	if req.StartDate.After(req.EndDate) {
		result.IsValid = false
		result.Reason = "Start date must be before end date"
		return result, nil
	}

	// 验证2：检查是否为过去的日期
	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		result.IsValid = false
		result.Reason = "Cannot request leave for past dates"
		return result, nil
	}

	// 验证3：检查休假类型
	validLeaveTypes := []string{"annual", "sick", "personal", "maternity", "paternity"}
	isValidType := false
	for _, validType := range validLeaveTypes {
		if req.LeaveType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		result.IsValid = false
		result.Reason = "Invalid leave type"
		return result, nil
	}

	// 验证4：检查休假长度（简化版本）
	duration := req.EndDate.Sub(req.StartDate).Hours() / 24
	if duration > 30 {
		result.IsValid = false
		result.Reason = "Leave duration cannot exceed 30 days"
		return result, nil
	}

	// 模拟数据库查询检查冲突
	time.Sleep(500 * time.Millisecond)

	// 记录指标
	processingDuration := time.Since(start)
	metrics.RecordDatabaseOperation("SELECT", "leave_requests", "success", processingDuration)

	logger.Info("Leave request validation completed",
		"request_id", req.RequestID,
		"is_valid", result.IsValid,
		"reason", result.Reason)

	return result, nil
}

// NotifyManagerForApprovalActivity 通知经理审批活动
func (a *Activities) NotifyManagerForApprovalActivity(ctx context.Context, req NotifyManagerForApprovalRequest) (*NotifyManagerForApprovalResult, error) {
	start := time.Now()
	logger := activity.GetLogger(ctx)

	logger.Info("Notifying manager for approval",
		"request_id", req.RequestID,
		"manager_id", req.ManagerID,
		"employee_id", req.EmployeeID)

	// 构造审批通知内容
	_ = fmt.Sprintf(`
You have a new leave request to review:

Request ID: %s
Employee ID: %s
Leave Type: %s
Start Date: %s
End Date: %s
Reason: %s
Requested At: %s

Please review and approve or reject this request.
	`, req.RequestID, req.EmployeeID, req.LeaveType,
		req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02"),
		req.Reason, req.RequestedAt.Format("2006-01-02 15:04:05"))

	// 模拟通知发送过程
	time.Sleep(1 * time.Second)

	// 记录指标
	duration := time.Since(start)
	metrics.RecordDatabaseOperation("INSERT", "approval_notifications", "success", duration)

	result := &NotifyManagerForApprovalResult{
		NotificationID: fmt.Sprintf("approval_notif_%d", time.Now().Unix()),
		Success:        true,
	}

	logger.Info("Manager approval notification sent successfully",
		"request_id", req.RequestID,
		"notification_id", result.NotificationID)

	return result, nil
}

// WaitForManagerApprovalActivity 等待经理审批活动
func (a *Activities) WaitForManagerApprovalActivity(ctx context.Context, req WaitForManagerApprovalRequest) (*ManagerApprovalResult, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Waiting for manager approval",
		"request_id", req.RequestID,
		"manager_id", req.ManagerID,
		"timeout_hours", req.TimeoutHours)

	// 在实际实现中，这里会等待外部信号或查询审批状态
	// 为了演示目的，我们模拟一个快速的自动审批

	// 模拟审批处理时间
	time.Sleep(5 * time.Second)

	// 模拟审批决策（80%概率通过）
	approved := time.Now().Unix()%5 != 0 // 80%概率

	result := &ManagerApprovalResult{
		ApproverID: req.ManagerID,
		ApprovedAt: time.Now(),
	}

	if approved {
		result.Decision = "approved"
		result.Comments = "Request approved automatically for demo purposes"
	} else {
		result.Decision = "rejected"
		result.Comments = "Request rejected automatically for demo purposes"
	}

	logger.Info("Manager approval completed",
		"request_id", req.RequestID,
		"decision", result.Decision)

	return result, nil
}

// SendLeaveApprovedNotificationActivity 发送休假审批通过通知活动
func (a *Activities) SendLeaveApprovedNotificationActivity(ctx context.Context, req LeaveApprovedNotificationRequest) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Sending leave approved notification",
		"request_id", req.RequestID,
		"employee_id", req.EmployeeID)

	// 构造通知内容
	_ = fmt.Sprintf(`
Good news! Your leave request has been approved.

Request ID: %s
Start Date: %s
End Date: %s
Approved By: %s

Please make sure to update your calendar and inform your team.

Best regards,
HR Team
	`, req.RequestID, req.StartDate.Format("2006-01-02"),
		req.EndDate.Format("2006-01-02"), req.ApproverID)

	// 模拟通知发送
	time.Sleep(1 * time.Second)

	logger.Info("Leave approved notification sent successfully",
		"request_id", req.RequestID)

	return nil
}

// SendLeaveRejectedNotificationActivity 发送休假审批拒绝通知活动
func (a *Activities) SendLeaveRejectedNotificationActivity(ctx context.Context, req LeaveRejectedNotificationRequest) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Sending leave rejected notification",
		"request_id", req.RequestID,
		"employee_id", req.EmployeeID)

	// 构造通知内容
	_ = fmt.Sprintf(`
We regret to inform you that your leave request has been rejected.

Request ID: %s
Reason: %s
Reviewed By: %s

If you have any questions, please contact your manager or HR.

Best regards,
HR Team
	`, req.RequestID, req.Reason, req.ApproverID)

	// 模拟通知发送
	time.Sleep(1 * time.Second)

	logger.Info("Leave rejected notification sent successfully",
		"request_id", req.RequestID)

	return nil
}

// === 批量员工处理相关活动 ===

// ProcessSingleEmployeeActivity 处理单个员工的活动
func (a *Activities) ProcessSingleEmployeeActivity(ctx context.Context, req ProcessSingleEmployeeRequest) (ProcessSingleEmployeeResult, error) {
	start := time.Now()

	result := ProcessSingleEmployeeResult{
		EmployeeID: req.EmployeeID,
		Status:     "success",
	}

	a.logger.Info("Processing single employee",
		"batch_id", req.BatchID,
		"employee_id", req.EmployeeID,
		"operation", req.Operation)

	// 根据操作类型执行不同的处理
	switch req.Operation {
	case "onboard":
		err := a.processEmployeeOnboard(ctx, req)
		if err != nil {
			result.Status = "failed"
			result.ErrorMessage = err.Error()
			a.logger.LogError("employee_onboard", "Failed to onboard employee", err, map[string]interface{}{
				"employee_id": req.EmployeeID,
				"batch_id":    req.BatchID,
			})
		}

	case "offboard":
		err := a.processEmployeeOffboard(ctx, req)
		if err != nil {
			result.Status = "failed"
			result.ErrorMessage = err.Error()
			a.logger.LogError("employee_offboard", "Failed to offboard employee", err, map[string]interface{}{
				"employee_id": req.EmployeeID,
				"batch_id":    req.BatchID,
			})
		}

	case "update":
		err := a.processEmployeeUpdate(ctx, req)
		if err != nil {
			result.Status = "failed"
			result.ErrorMessage = err.Error()
			a.logger.LogError("employee_update", "Failed to update employee", err, map[string]interface{}{
				"employee_id": req.EmployeeID,
				"batch_id":    req.BatchID,
			})
		}

	default:
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Unknown operation: %s", req.Operation)
		a.logger.LogError("employee_process", "Unknown operation", nil, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"operation":   req.Operation,
		})
	}

	// 记录处理时间
	duration := time.Since(start)
	metrics.RecordAIRequest(fmt.Sprintf("process_employee_%s", req.Operation), result.Status, duration)

	a.logger.Info("Single employee processing completed",
		"employee_id", req.EmployeeID,
		"status", result.Status,
		"duration", duration)

	return result, nil
}

// processEmployeeOnboard 处理员工入职
func (a *Activities) processEmployeeOnboard(ctx context.Context, req ProcessSingleEmployeeRequest) error {
	// 模拟员工入职处理逻辑
	// 在实际实现中，这里会调用具体的业务逻辑

	time.Sleep(100 * time.Millisecond) // 模拟处理时间

	// 检查必需字段
	if firstName, ok := req.Data["first_name"].(string); !ok || firstName == "" {
		return fmt.Errorf("missing required field: first_name")
	}

	if lastName, ok := req.Data["last_name"].(string); !ok || lastName == "" {
		return fmt.Errorf("missing required field: last_name")
	}

	if email, ok := req.Data["email"].(string); !ok || email == "" {
		return fmt.Errorf("missing required field: email")
	}

	// 记录入职事件
	a.logger.Info("Business event: employee_onboarded",
		"employee_id", req.EmployeeID.String(),
		"tenant_id", req.TenantID.String(),
		"batch_id", req.BatchID,
		"first_name", req.Data["first_name"],
		"last_name", req.Data["last_name"],
		"email", req.Data["email"],
		"department", req.Data["department"],
		"position", req.Data["position"],
	)

	return nil
}

// processEmployeeOffboard 处理员工离职
func (a *Activities) processEmployeeOffboard(ctx context.Context, req ProcessSingleEmployeeRequest) error {
	// 模拟员工离职处理逻辑
	time.Sleep(150 * time.Millisecond) // 模拟处理时间

	// 检查离职日期
	if lastWorkingDay, ok := req.Data["last_working_day"].(string); !ok || lastWorkingDay == "" {
		return fmt.Errorf("missing required field: last_working_day")
	}

	// 记录离职事件
	a.logger.Info("Business event: employee_offboarded",
		"employee_id", req.EmployeeID.String(),
		"tenant_id", req.TenantID.String(),
		"batch_id", req.BatchID,
		"last_working_day", req.Data["last_working_day"],
		"reason", req.Data["reason"],
	)

	return nil
}

// processEmployeeUpdate 处理员工信息更新
func (a *Activities) processEmployeeUpdate(ctx context.Context, req ProcessSingleEmployeeRequest) error {
	// 模拟员工信息更新处理逻辑
	time.Sleep(80 * time.Millisecond) // 模拟处理时间

	// 记录更新事件
	a.logger.Info("Business event: employee_updated",
		"employee_id", req.EmployeeID.String(),
		"tenant_id", req.TenantID.String(),
		"batch_id", req.BatchID,
		"updated_data", fmt.Sprintf("%+v", req.Data),
	)

	return nil
}

// === 入职工作流活动函数 ===

// CreateEmployeeAccountActivity 创建员工账户活动
func CreateEmployeeAccountActivity(ctx context.Context, req EmployeeOnboardingRequest) error {
	// 模拟创建员工账户
	return nil
}

// AssignEquipmentAndPermissionsActivity 分配设备和权限活动
func AssignEquipmentAndPermissionsActivity(ctx context.Context, req EmployeeOnboardingRequest) error {
	// 模拟分配设备和权限
	return nil
}

// SendWelcomeEmailActivity 发送欢迎邮件活动
func SendWelcomeEmailActivity(ctx context.Context, req EmployeeOnboardingRequest) error {
	// 模拟发送欢迎邮件
	return nil
}

// NotifyManagerActivity 通知经理活动
func NotifyManagerActivity(ctx context.Context, req EmployeeOnboardingRequest) error {
	// 模拟通知经理
	return nil
}

// === 休假审批工作流活动函数 ===

// ValidateLeaveRequestActivity 验证休假请求活动
func ValidateLeaveRequestActivity(ctx context.Context, req LeaveApprovalRequest) error {
	// 模拟验证休假请求
	return nil
}

// NotifyManagerForApprovalActivity 通知经理审批活动
func NotifyManagerForApprovalActivity(ctx context.Context, req LeaveApprovalRequest) error {
	// 模拟通知经理审批
	return nil
}

// WaitForManagerApprovalActivity 等待经理审批活动
func WaitForManagerApprovalActivity(ctx context.Context, req LeaveApprovalRequest) (LeaveApprovalResponse, error) {
	// 模拟等待经理审批，实际应该是长时间运行的任务
	return LeaveApprovalResponse{Approved: true, Comments: "Approved"}, nil
}

// SendLeaveApprovedNotificationActivity 发送休假批准通知活动
func SendLeaveApprovedNotificationActivity(ctx context.Context, req LeaveApprovalRequest) error {
	// 模拟发送批准通知
	return nil
}

// SendLeaveRejectedNotificationActivity 发送休假拒绝通知活动
func SendLeaveRejectedNotificationActivity(ctx context.Context, req LeaveApprovalRequest) error {
	// 模拟发送拒绝通知
	return nil
}

// ProcessSingleEmployeeActivity 处理单个员工数据活动
func ProcessSingleEmployeeActivity(ctx context.Context, req ProcessSingleEmployeeRequest) error {
	// 模拟处理单个员工数据
	return nil
}
