package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
)

// BusinessLogic 业务逻辑层 - 可独立测试
type BusinessLogic struct {
	logger *logging.StructuredLogger
}

// NewBusinessLogic 创建业务逻辑实例
func NewBusinessLogic(logger *logging.StructuredLogger) *BusinessLogic {
	return &BusinessLogic{
		logger: logger,
	}
}

// EmployeeAccountService 员工账户服务
func (bl *BusinessLogic) CreateEmployeeAccount(ctx context.Context, req CreateAccountRequest) (*CreateAccountResult, error) {
	// 验证输入
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return nil, fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
	}

	// 模拟账户创建业务逻辑
	accountID := fmt.Sprintf("acc_%s", req.EmployeeID.String()[:8])
	
	// 记录业务事件
	bl.logger.Info("Creating employee account",
		"employee_id", req.EmployeeID,
		"email", req.Email,
		"account_id", accountID)

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	return &CreateAccountResult{
		AccountID: accountID,
		Success:   true,
	}, nil
}

// EquipmentService 设备分配服务
func (bl *BusinessLogic) AssignEquipmentAndPermissions(ctx context.Context, req AssignEquipmentRequest) (*AssignEquipmentResult, error) {
	// 验证输入
	if req.Department == "" {
		return nil, fmt.Errorf("department is required")
	}
	if req.Position == "" {
		return nil, fmt.Errorf("position is required")
	}

	// 根据部门和职位确定设备清单
	var assignedItems []string
	
	switch req.Department {
	case "技术部", "Technology", "Engineering":
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

	bl.logger.Info("Assigning equipment",
		"employee_id", req.EmployeeID,
		"department", req.Department,
		"assigned_items", assignedItems)

	return &AssignEquipmentResult{
		AssignedItems: assignedItems,
		Success:       true,
	}, nil
}

// EmailService 邮件服务
func (bl *BusinessLogic) SendWelcomeEmail(ctx context.Context, req WelcomeEmailRequest) (*SendEmailResult, error) {
	// 验证输入
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return nil, fmt.Errorf("first name is required")
	}

	// 构造邮件内容
	emailContent := fmt.Sprintf(`
Dear %s,

Welcome to Cube Castle! We are excited to have you join our team in the %s department.

Your start date is %s. Please arrive at 9:00 AM for your orientation session.

Best regards,
The Cube Castle Team
	`, req.FirstName, req.Department, req.StartDate.Format("January 2, 2006"))

	messageID := fmt.Sprintf("msg_%d", time.Now().Unix())

	bl.logger.Info("Sending welcome email",
		"employee_id", req.EmployeeID,
		"email", req.Email,
		"message_id", messageID,
		"content_preview", emailContent[:100]+"...")

	return &SendEmailResult{
		MessageID: messageID,
		Success:   true,
	}, nil
}

// NotificationService 通知服务
func (bl *BusinessLogic) NotifyManager(ctx context.Context, req NotifyManagerRequest) (*NotifyManagerResult, error) {
	// 验证输入
	if req.ManagerID == uuid.Nil {
		return nil, fmt.Errorf("manager ID is required")
	}
	if req.EmployeeName == "" {
		return nil, fmt.Errorf("employee name is required")
	}

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

	notificationID := fmt.Sprintf("notif_%d", time.Now().Unix())

	bl.logger.Info("Sending manager notification",
		"manager_id", req.ManagerID,
		"new_employee_id", req.NewEmployeeID,
		"notification_id", notificationID)

	return &NotifyManagerResult{
		NotificationID: notificationID,
		Success:        true,
	}, nil
}

// LeaveValidationService 休假验证服务
func (bl *BusinessLogic) ValidateLeaveRequest(ctx context.Context, req ValidateLeaveRequestRequest) (*ValidateLeaveRequestResult, error) {
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

	// 验证4：检查休假长度
	duration := req.EndDate.Sub(req.StartDate).Hours() / 24
	if duration > 30 {
		result.IsValid = false
		result.Reason = "Leave duration cannot exceed 30 days"
		return result, nil
	}

	bl.logger.Info("Leave request validation completed",
		"request_id", req.RequestID,
		"is_valid", result.IsValid)

	return result, nil
}

// 重构后的Activities - 只作为Temporal适配器
func (a *Activities) CreateEmployeeAccountActivityV2(ctx context.Context, req CreateAccountRequest) (*CreateAccountResult, error) {
	// 使用业务逻辑层
	businessLogic := NewBusinessLogic(a.logger)
	return businessLogic.CreateEmployeeAccount(ctx, req)
}

func (a *Activities) AssignEquipmentAndPermissionsActivityV2(ctx context.Context, req AssignEquipmentRequest) (*AssignEquipmentResult, error) {
	businessLogic := NewBusinessLogic(a.logger)
	return businessLogic.AssignEquipmentAndPermissions(ctx, req)
}

func (a *Activities) SendWelcomeEmailActivityV2(ctx context.Context, req WelcomeEmailRequest) (*SendEmailResult, error) {
	businessLogic := NewBusinessLogic(a.logger)
	return businessLogic.SendWelcomeEmail(ctx, req)
}

func (a *Activities) NotifyManagerActivityV2(ctx context.Context, req NotifyManagerRequest) (*NotifyManagerResult, error) {
	businessLogic := NewBusinessLogic(a.logger)
	return businessLogic.NotifyManager(ctx, req)
}

func (a *Activities) ValidateLeaveRequestActivityV2(ctx context.Context, req ValidateLeaveRequestRequest) (*ValidateLeaveRequestResult, error) {
	businessLogic := NewBusinessLogic(a.logger)
	return businessLogic.ValidateLeaveRequest(ctx, req)
}