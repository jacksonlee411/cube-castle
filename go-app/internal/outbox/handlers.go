package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// EmployeeEventHandler 员工事件处理器
type EmployeeEventHandler struct {
	employeeRepo EmployeeRepository
}

// NewEmployeeEventHandler 创建员工事件处理器
func NewEmployeeEventHandler(employeeRepo EmployeeRepository) *EmployeeEventHandler {
	return &EmployeeEventHandler{
		employeeRepo: employeeRepo,
	}
}

// GetEventType 获取事件类型
func (h *EmployeeEventHandler) GetEventType() string {
	return EventTypeEmployeeCreated
}

// HandleEvent 处理员工创建事件
func (h *EmployeeEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("👤 Processing employee created event: %s", event.ID)

	// 这里可以添加具体的业务逻辑，比如：
	// - 发送欢迎邮件
	// - 创建用户账户
	// - 分配默认权限
	// - 发送通知给HR部门

	log.Printf("✅ Employee created event processed: %s", event.ID)
	return nil
}

// EmployeeUpdatedEventHandler 员工更新事件处理器
type EmployeeUpdatedEventHandler struct {
	employeeRepo EmployeeRepository
}

// NewEmployeeUpdatedEventHandler 创建员工更新事件处理器
func NewEmployeeUpdatedEventHandler(employeeRepo EmployeeRepository) *EmployeeUpdatedEventHandler {
	return &EmployeeUpdatedEventHandler{
		employeeRepo: employeeRepo,
	}
}

// GetEventType 获取事件类型
func (h *EmployeeUpdatedEventHandler) GetEventType() string {
	return EventTypeEmployeeUpdated
}

// HandleEvent 处理员工更新事件
func (h *EmployeeUpdatedEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("✏️ Processing employee updated event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		EmployeeID    string                 `json:"employee_id"`
		UpdatedFields map[string]interface{} `json:"updated_fields"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 根据更新的字段执行相应的业务逻辑
	if _, ok := payload.UpdatedFields["email"]; ok {
		log.Printf("📧 Employee email updated, sending notification")
		// 发送邮件变更通知
	}

	if _, ok := payload.UpdatedFields["department"]; ok {
		log.Printf("🏢 Employee department changed, updating permissions")
		// 更新部门权限
	}

	log.Printf("✅ Employee updated event processed: %s", event.ID)
	return nil
}

// EmployeePhoneUpdatedEventHandler 员工电话更新事件处理器
type EmployeePhoneUpdatedEventHandler struct {
	employeeRepo EmployeeRepository
}

// NewEmployeePhoneUpdatedEventHandler 创建员工电话更新事件处理器
func NewEmployeePhoneUpdatedEventHandler(employeeRepo EmployeeRepository) *EmployeePhoneUpdatedEventHandler {
	return &EmployeePhoneUpdatedEventHandler{
		employeeRepo: employeeRepo,
	}
}

// GetEventType 获取事件类型
func (h *EmployeePhoneUpdatedEventHandler) GetEventType() string {
	return EventTypeEmployeePhoneUpdated
}

// HandleEvent 处理员工电话更新事件
func (h *EmployeePhoneUpdatedEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("📱 Processing employee phone updated event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		EmployeeID     string `json:"employee_id"`
		OldPhoneNumber string `json:"old_phone_number"`
		NewPhoneNumber string `json:"new_phone_number"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 发送短信验证码到新手机号
	log.Printf("📲 Sending verification SMS to new phone number: %s", payload.NewPhoneNumber)

	// 更新相关系统的手机号信息
	log.Printf("🔄 Updating phone number in external systems")

	log.Printf("✅ Employee phone updated event processed: %s", event.ID)
	return nil
}

// OrganizationEventHandler 组织事件处理器
type OrganizationEventHandler struct {
	organizationRepo OrganizationRepository
}

// NewOrganizationEventHandler 创建组织事件处理器
func NewOrganizationEventHandler(organizationRepo OrganizationRepository) *OrganizationEventHandler {
	return &OrganizationEventHandler{
		organizationRepo: organizationRepo,
	}
}

// GetEventType 获取事件类型
func (h *OrganizationEventHandler) GetEventType() string {
	return EventTypeOrganizationCreated
}

// HandleEvent 处理组织创建事件
func (h *OrganizationEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("🏢 Processing organization created event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		OrganizationID string `json:"organization_id"`
		Name           string `json:"name"`
		Code           string `json:"code"`
		ParentID       string `json:"parent_id,omitempty"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 创建组织相关的默认配置
	log.Printf("⚙️ Creating default configurations for organization: %s", payload.Name)

	// 发送通知给相关管理员
	log.Printf("📢 Notifying administrators about new organization")

	log.Printf("✅ Organization created event processed: %s", event.ID)
	return nil
}

// LeaveRequestEventHandler 休假申请事件处理器
type LeaveRequestEventHandler struct {
	leaveRequestRepo LeaveRequestRepository
}

// NewLeaveRequestEventHandler 创建休假申请事件处理器
func NewLeaveRequestEventHandler(leaveRequestRepo LeaveRequestRepository) *LeaveRequestEventHandler {
	return &LeaveRequestEventHandler{
		leaveRequestRepo: leaveRequestRepo,
	}
}

// GetEventType 获取事件类型
func (h *LeaveRequestEventHandler) GetEventType() string {
	return EventTypeLeaveRequestCreated
}

// HandleEvent 处理休假申请创建事件
func (h *LeaveRequestEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("🏖️ Processing leave request created event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		RequestID  string `json:"request_id"`
		EmployeeID string `json:"employee_id"`
		ManagerID  string `json:"manager_id"`
		StartDate  string `json:"start_date"`
		EndDate    string `json:"end_date"`
		LeaveType  string `json:"leave_type"`
		Reason     string `json:"reason"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 发送通知给经理
	log.Printf("📧 Sending notification to manager: %s", payload.ManagerID)

	// 创建审批工作流
	log.Printf("🔄 Creating approval workflow for leave request")

	// 更新员工休假余额
	log.Printf("📊 Updating employee leave balance")

	log.Printf("✅ Leave request created event processed: %s", event.ID)
	return nil
}

// LeaveRequestApprovedEventHandler 休假申请批准事件处理器
type LeaveRequestApprovedEventHandler struct {
	leaveRequestRepo LeaveRequestRepository
}

// NewLeaveRequestApprovedEventHandler 创建休假申请批准事件处理器
func NewLeaveRequestApprovedEventHandler(leaveRequestRepo LeaveRequestRepository) *LeaveRequestApprovedEventHandler {
	return &LeaveRequestApprovedEventHandler{
		leaveRequestRepo: leaveRequestRepo,
	}
}

// GetEventType 获取事件类型
func (h *LeaveRequestApprovedEventHandler) GetEventType() string {
	return EventTypeLeaveRequestApproved
}

// HandleEvent 处理休假申请批准事件
func (h *LeaveRequestApprovedEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("✅ Processing leave request approved event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		RequestID  string `json:"request_id"`
		EmployeeID string `json:"employee_id"`
		ApprovedBy string `json:"approved_by"`
		ApprovedAt string `json:"approved_at"`
		Comment    string `json:"comment,omitempty"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 发送批准通知给员工
	log.Printf("📧 Sending approval notification to employee: %s", payload.EmployeeID)

	// 更新休假申请状态
	log.Printf("🔄 Updating leave request status to approved")

	// 扣除休假余额
	log.Printf("📊 Deducting leave balance")

	// 创建日历事件
	log.Printf("📅 Creating calendar event for approved leave")

	log.Printf("✅ Leave request approved event processed: %s", event.ID)
	return nil
}

// LeaveRequestRejectedEventHandler 休假申请拒绝事件处理器
type LeaveRequestRejectedEventHandler struct {
	leaveRequestRepo LeaveRequestRepository
}

// NewLeaveRequestRejectedEventHandler 创建休假申请拒绝事件处理器
func NewLeaveRequestRejectedEventHandler(leaveRequestRepo LeaveRequestRepository) *LeaveRequestRejectedEventHandler {
	return &LeaveRequestRejectedEventHandler{
		leaveRequestRepo: leaveRequestRepo,
	}
}

// GetEventType 获取事件类型
func (h *LeaveRequestRejectedEventHandler) GetEventType() string {
	return EventTypeLeaveRequestRejected
}

// HandleEvent 处理休假申请拒绝事件
func (h *LeaveRequestRejectedEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("❌ Processing leave request rejected event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		RequestID  string `json:"request_id"`
		EmployeeID string `json:"employee_id"`
		RejectedBy string `json:"rejected_by"`
		RejectedAt string `json:"rejected_at"`
		Reason     string `json:"reason"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 发送拒绝通知给员工
	log.Printf("📧 Sending rejection notification to employee: %s", payload.EmployeeID)

	// 更新休假申请状态
	log.Printf("🔄 Updating leave request status to rejected")

	// 退还休假余额（如果已扣除）
	log.Printf("📊 Restoring leave balance if deducted")

	log.Printf("✅ Leave request rejected event processed: %s", event.ID)
	return nil
}

// NotificationEventHandler 通知事件处理器
type NotificationEventHandler struct{}

// NewNotificationEventHandler 创建通知事件处理器
func NewNotificationEventHandler() *NotificationEventHandler {
	return &NotificationEventHandler{}
}

// GetEventType 获取事件类型
func (h *NotificationEventHandler) GetEventType() string {
	return "notification.sent"
}

// HandleEvent 处理通知事件
func (h *NotificationEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	log.Printf("📢 Processing notification event: %s", event.ID)

	// 解析事件载荷
	var payload struct {
		Type        string `json:"type"`
		RecipientID string `json:"recipient_id"`
		Subject     string `json:"subject"`
		Content     string `json:"content"`
		Channel     string `json:"channel"` // email, sms, push, etc.
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// 根据渠道发送通知
	switch payload.Channel {
	case "email":
		log.Printf("📧 Sending email notification to: %s", payload.RecipientID)
	case "sms":
		log.Printf("📱 Sending SMS notification to: %s", payload.RecipientID)
	case "push":
		log.Printf("📱 Sending push notification to: %s", payload.RecipientID)
	default:
		log.Printf("⚠️ Unknown notification channel: %s", payload.Channel)
	}

	log.Printf("✅ Notification event processed: %s", event.ID)
	return nil
}
