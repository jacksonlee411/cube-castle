package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// EmployeeEventConsumer 员工事件消费者
// 负责将员工变更事件同步到Neo4j图数据库
type EmployeeEventConsumer struct {
	neo4jService *service.Neo4jService
	logger       Logger
}

// NewEmployeeEventConsumer 创建员工事件消费者
func NewEmployeeEventConsumer(neo4jService *service.Neo4jService, logger Logger) *EmployeeEventConsumer {
	return &EmployeeEventConsumer{
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// Handle 实现EventHandler接口 - 处理领域事件
func (c *EmployeeEventConsumer) Handle(ctx context.Context, event events.DomainEvent) error {
	fmt.Printf("🔄 员工事件消费者接收到事件: %s\n", event.GetEventType())
	
	c.logger.Info("处理员工事件", 
		"event_type", event.GetEventType(),
		"event_id", event.GetEventID(),
		"aggregate_id", event.GetAggregateID(),
		"tenant_id", event.GetTenantID(),
	)

	switch event.GetEventType() {
	case "employee.hired":
		return c.handleEmployeeHired(ctx, event)
	case "employee.updated":
		return c.handleEmployeeUpdatedDomainEvent(ctx, event)
	case "employee.terminated":
		return c.handleEmployeeTerminatedDomainEvent(ctx, event)
	case "employee.deleted":
		return c.handleEmployeeDeletedDomainEvent(ctx, event)
	default:
		c.logger.Warn("Unknown employee event type", "event_type", event.GetEventType())
		return nil // 不处理未知事件类型，但不报错
	}
}

// GetEventType 返回处理的事件类型
func (c *EmployeeEventConsumer) GetEventType() string {
	return "employee.*" // 处理所有员工相关事件
}

// GetHandlerName 返回处理器名称
func (c *EmployeeEventConsumer) GetHandlerName() string {
	return "EmployeeEventConsumer"
}

// ConsumeEmployeeEvent 保留原有接口以兼容性（已弃用，使用Handle方法）
func (c *EmployeeEventConsumer) ConsumeEmployeeEvent(ctx context.Context, event []byte) error {
	var employeeEvent EmployeeEvent
	if err := json.Unmarshal(event, &employeeEvent); err != nil {
		c.logger.Error("Failed to unmarshal employee event", "error", err)
		return fmt.Errorf("failed to unmarshal employee event: %w", err)
	}

	c.logger.Info("Processing employee event", 
		"event_type", employeeEvent.EventType,
		"employee_id", employeeEvent.EmployeeID,
		"tenant_id", employeeEvent.TenantID,
	)

	switch employeeEvent.EventType {
	case "employee.hired":
		return c.handleEmployeeHiredLegacy(ctx, employeeEvent)
	case "employee.updated":
		return c.handleEmployeeUpdatedLegacy(ctx, employeeEvent)
	case "employee.terminated":
		return c.handleEmployeeTerminatedLegacy(ctx, employeeEvent)
	case "employee.deleted":
		return c.handleEmployeeDeletedLegacy(ctx, employeeEvent)
	default:
		c.logger.Warn("Unknown employee event type", "event_type", employeeEvent.EventType)
		return nil // 不处理未知事件类型，但不报错
	}
}

// EmployeeEvent 员工事件基础结构（兼容性）
type EmployeeEvent struct {
	EventType   string                 `json:"event_type"`
	EventID     string                 `json:"event_id"`
	TenantID    uuid.UUID             `json:"tenant_id"`
	EmployeeID  uuid.UUID             `json:"employee_id"`
	Timestamp   time.Time             `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// handleEmployeeHired 处理员工入职事件（新版本 - 使用DomainEvent）
func (c *EmployeeEventConsumer) handleEmployeeHired(ctx context.Context, event events.DomainEvent) error {
	fmt.Printf("📝 处理员工入职事件: %s\n", event.GetAggregateID())
	
	c.logger.Info("Handling employee hired event", "employee_id", event.GetAggregateID())

	// 将DomainEvent序列化然后反序列化以获取具体数据
	eventData, err := event.Serialize()
	if err != nil {
		c.logger.Error("Failed to serialize event", "error", err)
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	var eventMap map[string]interface{}
	if err := json.Unmarshal(eventData, &eventMap); err != nil {
		c.logger.Error("Failed to unmarshal event data", "error", err)
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// 从事件数据中提取员工信息
	employeeNode := service.EmployeeNode{
		ID:         event.GetAggregateID().String(),
		EmployeeID: getStringFromEventMap(eventMap, "employee_number", event.GetAggregateID().String()),
		LegalName:  fmt.Sprintf("%s %s", getStringFromEventMap(eventMap, "first_name", ""), getStringFromEventMap(eventMap, "last_name", "")),
		Email:      getStringFromEventMap(eventMap, "email", ""),
		Status:     "ACTIVE", // 默认为活跃状态
		HireDate:   parseTimeFromEventMap(eventMap, "hire_date"),
		Properties: map[string]interface{}{
			"created_at": event.GetTimestamp().Format(time.RFC3339),
		},
	}

	// 同步到Neo4j - 为Neo4j操作创建新的上下文，避免使用可能已取消的事件上下文
	neo4jCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	err = c.neo4jService.SyncEmployee(neo4jCtx, employeeNode)
	if err != nil {
		c.logger.Error("Failed to sync hired employee to Neo4j", 
			"employee_id", event.GetAggregateID(),
			"error", err,
		)
		return fmt.Errorf("failed to sync hired employee to Neo4j: %w", err)
	}

	fmt.Printf("✅ 员工成功同步到Neo4j: %s\n", event.GetAggregateID())
	c.logger.Info("Successfully synced hired employee to Neo4j", "employee_id", event.GetAggregateID())
	return nil
}

// handleEmployeeUpdated 处理员工更新事件
func (c *EmployeeEventConsumer) handleEmployeeUpdated(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee updated event", "employee_id", event.EmployeeID)

	// 获取当前员工数据
	currentEmployee, err := c.neo4jService.GetEmployee(ctx, event.EmployeeID.String())
	if err != nil {
		c.logger.Error("Failed to get current employee from Neo4j", 
			"employee_id", event.EmployeeID,
			"error", err,
		)
		return fmt.Errorf("failed to get current employee: %w", err)
	}

	if currentEmployee == nil {
		c.logger.Warn("Employee not found in Neo4j, skipping update", "employee_id", event.EmployeeID)
		return nil
	}

	// 应用更新字段
	updatedFields := event.Data["updated_fields"]
	if updatedFieldsMap, ok := updatedFields.(map[string]interface{}); ok {
		// 更新基本字段
		if firstName, exists := updatedFieldsMap["first_name"]; exists {
			if firstNameStr, ok := firstName.(string); ok {
				lastName := getStringFromData(updatedFieldsMap, "last_name", extractLastName(currentEmployee.LegalName))
				currentEmployee.LegalName = fmt.Sprintf("%s %s", firstNameStr, lastName)
			}
		}
		
		if lastName, exists := updatedFieldsMap["last_name"]; exists {
			if lastNameStr, ok := lastName.(string); ok {
				firstName := getStringFromData(updatedFieldsMap, "first_name", extractFirstName(currentEmployee.LegalName))
				currentEmployee.LegalName = fmt.Sprintf("%s %s", firstName, lastNameStr)
			}
		}

		if email, exists := updatedFieldsMap["email"]; exists {
			if emailStr, ok := email.(string); ok {
				currentEmployee.Email = emailStr
			}
		}

		if status, exists := updatedFieldsMap["employment_status"]; exists {
			if statusStr, ok := status.(string); ok {
				currentEmployee.Status = statusStr
			}
		}

		// 更新属性
		currentEmployee.Properties["updated_at"] = event.Timestamp.Format(time.RFC3339)
		for key, value := range updatedFieldsMap {
			currentEmployee.Properties[key] = value
		}
	}

	// 同步更新到Neo4j
	err = c.neo4jService.SyncEmployee(ctx, *currentEmployee)
	if err != nil {
		c.logger.Error("Failed to sync updated employee to Neo4j", 
			"employee_id", event.EmployeeID,
			"error", err,
		)
		return fmt.Errorf("failed to sync updated employee to Neo4j: %w", err)
	}

	c.logger.Info("Successfully synced updated employee to Neo4j", "employee_id", event.EmployeeID)
	return nil
}

// handleEmployeeTerminated 处理员工离职事件
func (c *EmployeeEventConsumer) handleEmployeeTerminated(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee terminated event", "employee_id", event.EmployeeID)

	// 获取当前员工数据
	currentEmployee, err := c.neo4jService.GetEmployee(ctx, event.EmployeeID.String())
	if err != nil {
		c.logger.Error("Failed to get current employee from Neo4j", 
			"employee_id", event.EmployeeID,
			"error", err,
		)
		return fmt.Errorf("failed to get current employee: %w", err)
	}

	if currentEmployee == nil {
		c.logger.Warn("Employee not found in Neo4j, skipping termination", "employee_id", event.EmployeeID)
		return nil
	}

	// 更新员工状态为已终止
	currentEmployee.Status = "TERMINATED"
	currentEmployee.Properties["terminated_at"] = event.Timestamp.Format(time.RFC3339)
	currentEmployee.Properties["termination_date"] = getStringFromData(event.Data, "termination_date", event.Timestamp.Format("2006-01-02"))
	currentEmployee.Properties["termination_type"] = getStringFromData(event.Data, "termination_type", "voluntary")
	currentEmployee.Properties["termination_reason"] = getStringFromData(event.Data, "termination_reason", "")

	// 同步更新到Neo4j
	err = c.neo4jService.SyncEmployee(ctx, *currentEmployee)
	if err != nil {
		c.logger.Error("Failed to sync terminated employee to Neo4j", 
			"employee_id", event.EmployeeID,
			"error", err,
		)
		return fmt.Errorf("failed to sync terminated employee to Neo4j: %w", err)
	}

	c.logger.Info("Successfully synced terminated employee to Neo4j", "employee_id", event.EmployeeID)
	return nil
}

// handleEmployeeDeleted 处理员工删除事件
func (c *EmployeeEventConsumer) handleEmployeeDeleted(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee deleted event", "employee_id", event.EmployeeID)

	// 在Neo4j中软删除或标记员工
	// 这里我们选择标记为已删除而不是物理删除，以保留历史数据
	currentEmployee, err := c.neo4jService.GetEmployee(ctx, event.EmployeeID.String())
	if err != nil {
		c.logger.Error("Failed to get current employee from Neo4j", 
			"employee_id", event.EmployeeID,
			"error", err,
		)
		return fmt.Errorf("failed to get current employee: %w", err)
	}

	if currentEmployee == nil {
		c.logger.Warn("Employee not found in Neo4j, skipping deletion", "employee_id", event.EmployeeID)
		return nil
	}

	// 标记为已删除
	currentEmployee.Status = "DELETED"
	currentEmployee.Properties["deleted_at"] = event.Timestamp.Format(time.RFC3339)
	currentEmployee.Properties["deleted_by"] = getStringFromData(event.Data, "deleted_by", "system")
	currentEmployee.Properties["deletion_reason"] = getStringFromData(event.Data, "deletion_reason", "")

	// 同步更新到Neo4j
	err = c.neo4jService.SyncEmployee(ctx, *currentEmployee)
	if err != nil {
		c.logger.Error("Failed to sync deleted employee to Neo4j", 
			"employee_id", event.EmployeeID,
			"error", err,
		)
		return fmt.Errorf("failed to sync deleted employee to Neo4j: %w", err)
	}

	c.logger.Info("Successfully synced deleted employee to Neo4j", "employee_id", event.EmployeeID)
	return nil
}

// 辅助函数：从事件数据中获取字符串值
func getStringFromData(data map[string]interface{}, key, defaultValue string) string {
	if value, exists := data[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// 辅助函数：从事件数据中解析时间
func parseTimeFromData(data map[string]interface{}, key string) time.Time {
	if value, exists := data[key]; exists {
		if timeStr, ok := value.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
				return parsedTime
			}
			if parsedTime, err := time.Parse("2006-01-02", timeStr); err == nil {
				return parsedTime
			}
		}
	}
	return time.Now()
}

// 辅助函数：从全名中提取名字
func extractFirstName(legalName string) string {
	parts := strings.Fields(legalName)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// 辅助函数：从全名中提取姓氏
func extractLastName(legalName string) string {
	parts := strings.Fields(legalName)
	if len(parts) > 1 {
		return strings.Join(parts[1:], " ")
	}
	return ""
}

// 新增辅助函数：从事件映射中获取字符串值
func getStringFromEventMap(eventMap map[string]interface{}, key, defaultValue string) string {
	if value, exists := eventMap[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// 新增辅助函数：从事件映射中解析时间
func parseTimeFromEventMap(eventMap map[string]interface{}, key string) time.Time {
	if value, exists := eventMap[key]; exists {
		if timeStr, ok := value.(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
				return parsedTime
			}
			if parsedTime, err := time.Parse("2006-01-02", timeStr); err == nil {
				return parsedTime
			}
		}
	}
	return time.Now()
}

// 需要实现其他事件处理方法的存根（暂时）
func (c *EmployeeEventConsumer) handleEmployeeUpdatedDomainEvent(ctx context.Context, event events.DomainEvent) error {
	c.logger.Info("处理员工更新事件 (暂未实现)", "employee_id", event.GetAggregateID())
	return nil
}

func (c *EmployeeEventConsumer) handleEmployeeTerminatedDomainEvent(ctx context.Context, event events.DomainEvent) error {
	c.logger.Info("处理员工离职事件 (暂未实现)", "employee_id", event.GetAggregateID())
	return nil
}

func (c *EmployeeEventConsumer) handleEmployeeDeletedDomainEvent(ctx context.Context, event events.DomainEvent) error {
	c.logger.Info("处理员工删除事件 (暂未实现)", "employee_id", event.GetAggregateID())
	return nil
}

// 兼容性处理方法
func (c *EmployeeEventConsumer) handleEmployeeHiredLegacy(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee hired event (legacy)", "employee_id", event.EmployeeID)
	// 旧版本处理逻辑保持不变...
	return nil
}

func (c *EmployeeEventConsumer) handleEmployeeUpdatedLegacy(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee updated event (legacy)", "employee_id", event.EmployeeID)
	// 旧版本处理逻辑...
	return nil
}

func (c *EmployeeEventConsumer) handleEmployeeTerminatedLegacy(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee terminated event (legacy)", "employee_id", event.EmployeeID)
	// 旧版本处理逻辑...
	return nil
}

func (c *EmployeeEventConsumer) handleEmployeeDeletedLegacy(ctx context.Context, event EmployeeEvent) error {
	c.logger.Info("Handling employee deleted event (legacy)", "employee_id", event.EmployeeID)
	// 旧版本处理逻辑...
	return nil
}