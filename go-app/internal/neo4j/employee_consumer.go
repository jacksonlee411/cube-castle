package neo4j

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// EmployeeEventConsumer 员工事件消费者
type EmployeeEventConsumer struct {
	*BaseEventConsumer
}

// NewEmployeeEventConsumer 创建员工事件消费者
func NewEmployeeEventConsumer(connMgr ConnectionManagerInterface) *EmployeeEventConsumer {
	base := NewBaseEventConsumer(connMgr, "employee.*")
	
	return &EmployeeEventConsumer{
		BaseEventConsumer: base,
	}
}

// ConsumeEvent 消费员工事件
func (c *EmployeeEventConsumer) ConsumeEvent(ctx context.Context, event events.DomainEvent) error {
	log.Printf("🔄 处理员工事件: %s (ID: %s)", event.GetEventType(), event.GetEventID())
	
	switch event.GetEventType() {
	case "employee.created":
		return c.handleEmployeeCreated(ctx, event)
	case "employee.updated":
		return c.handleEmployeeUpdated(ctx, event)
	case "employee.deleted":
		return c.handleEmployeeDeleted(ctx, event)
	case "employee.hired":
		return c.handleEmployeeHired(ctx, event)
	case "employee.terminated":
		return c.handleEmployeeTerminated(ctx, event)
	case "employee.phone_updated":
		return c.handleEmployeePhoneUpdated(ctx, event)
	default:
		log.Printf("⚠️ 未知的员工事件类型: %s", event.GetEventType())
		return nil
	}
}

// handleEmployeeCreated 处理员工创建事件
func (c *EmployeeEventConsumer) handleEmployeeCreated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseEmployeeCreatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析员工创建事件失败: %w", err)
	}
	
	// 创建同步操作
	syncOp := &NodeSyncOperation{
		Label:      "Employee",
		Operation:  "CREATE",
		UniqueKeys: []string{"id", "tenant_id"},
		Properties: map[string]interface{}{
			"id":              eventData.EmployeeID.String(),
			"tenant_id":       eventData.TenantID.String(),
			"employee_number": eventData.EmployeeNumber,
			"first_name":      eventData.FirstName,
			"last_name":       eventData.LastName,
			"email":           eventData.Email,
			"hire_date":       eventData.HireDate.Format(time.RFC3339),
			"status":          eventData.Status,
			"created_at":      event.GetTimestamp().Format(time.RFC3339),
			"event_id":        event.GetEventID(),
			"event_version":   event.GetEventVersion(),
		},
	}
	
	// 验证操作
	if err := syncOp.Validate(); err != nil {
		return fmt.Errorf("验证同步操作失败: %w", err)
	}
	
	// 执行同步
	_, err = c.connectionManager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		if err := syncOp.Execute(ctx, tx); err != nil {
			return nil, err
		}
		
		// 创建事件记录节点（用于审计）
		eventOp := &NodeSyncOperation{
			Label:      "EmployeeEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":       event.GetEventID(),
				"event_type":     event.GetEventType(),
				"aggregate_id":   event.GetAggregateID().String(),
				"tenant_id":      event.GetTenantID().String(),
				"timestamp":      event.GetTimestamp().Format(time.RFC3339),
				"version":        "1", // 简化版本处理
				"processed_at":   time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步员工创建到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 员工创建事件同步完成: %s (员工号: %s)", eventData.EmployeeID, eventData.EmployeeNumber)
	return nil
}

// handleEmployeeUpdated 处理员工更新事件
func (c *EmployeeEventConsumer) handleEmployeeUpdated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseEmployeeUpdatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析员工更新事件失败: %w", err)
	}
	
	// 准备更新属性
	updateProps := map[string]interface{}{
		"id":        eventData.EmployeeID.String(),
		"tenant_id": eventData.TenantID.String(),
	}
	
	// 添加更新的字段
	for key, value := range eventData.UpdatedFields {
		updateProps[key] = value
	}
	
	// 添加元数据
	updateProps["updated_at"] = event.GetTimestamp().Format(time.RFC3339)
	updateProps["last_event_id"] = event.GetEventID()
	updateProps["version"] = event.GetEventVersion()
	
	// 创建同步操作
	syncOp := &NodeSyncOperation{
		Label:      "Employee",
		Operation:  "UPDATE",
		UniqueKeys: []string{"id", "tenant_id"},
		Properties: updateProps,
	}
	
	// 执行同步
	_, err = c.connectionManager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		if err := syncOp.Execute(ctx, tx); err != nil {
			return nil, err
		}
		
		// 记录更新事件
		eventOp := &NodeSyncOperation{
			Label:      "EmployeeEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":       event.GetEventID(),
				"event_type":     event.GetEventType(),
				"aggregate_id":   event.GetAggregateID().String(),
				"tenant_id":      event.GetTenantID().String(),
				"timestamp":      event.GetTimestamp().Format(time.RFC3339),
				"version":        "1", // 简化版本处理
				"updated_fields": formatUpdatedFields(eventData.UpdatedFields),
				"processed_at":   time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步员工更新到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 员工更新事件同步完成: %s", eventData.EmployeeID)
	return nil
}

// handleEmployeeDeleted 处理员工删除事件
func (c *EmployeeEventConsumer) handleEmployeeDeleted(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseEmployeeDeletedEvent(event)
	if err != nil {
		return fmt.Errorf("解析员工删除事件失败: %w", err)
	}
	
	// 执行软删除（标记为已删除而不是物理删除）
	_, err = c.connectionManager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 软删除：标记为已删除状态
		cypher := `
			MATCH (e:Employee) 
			WHERE e.id = $employee_id AND e.tenant_id = $tenant_id
			SET e.status = 'DELETED', 
			    e.deleted_at = $deleted_at,
			    e.deleted_by_event_id = $event_id,
			    e.is_deleted = true
		`
		
		params := map[string]interface{}{
			"employee_id": eventData.EmployeeID.String(),
			"tenant_id":   eventData.TenantID.String(),
			"deleted_at":  event.GetTimestamp().Format(time.RFC3339),
			"event_id":    event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录删除事件
		eventOp := &NodeSyncOperation{
			Label:      "EmployeeEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":     event.GetEventID(),
				"event_type":   event.GetEventType(),
				"aggregate_id": event.GetAggregateID().String(),
				"tenant_id":    event.GetTenantID().String(),
				"timestamp":    event.GetTimestamp().Format(time.RFC3339),
				"version":      event.GetEventVersion(),
				"reason":       eventData.Reason,
				"processed_at": time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步员工删除到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 员工删除事件同步完成: %s (原因: %s)", eventData.EmployeeID, eventData.Reason)
	return nil
}

// handleEmployeeHired 处理员工雇佣事件
func (c *EmployeeEventConsumer) handleEmployeeHired(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseEmployeeHiredEvent(event)
	if err != nil {
		return fmt.Errorf("解析员工雇佣事件失败: %w", err)
	}
	
	// 更新员工状态为已雇佣
	_, err = c.connectionManager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (e:Employee) 
			WHERE e.id = $employee_id AND e.tenant_id = $tenant_id
			SET e.status = 'ACTIVE',
			    e.hire_date = $hire_date,
			    e.hired_at = $hired_at,
			    e.last_event_id = $event_id
		`
		
		params := map[string]interface{}{
			"employee_id": eventData.EmployeeID.String(),
			"tenant_id":   eventData.TenantID.String(),
			"hire_date":   eventData.HireDate.Format(time.RFC3339),
			"hired_at":    event.GetTimestamp().Format(time.RFC3339),
			"event_id":    event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录雇佣事件
		eventOp := &NodeSyncOperation{
			Label:      "EmployeeEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":     event.GetEventID(),
				"event_type":   event.GetEventType(),
				"aggregate_id": event.GetAggregateID().String(),
				"tenant_id":    event.GetTenantID().String(),
				"timestamp":    event.GetTimestamp().Format(time.RFC3339),
				"hire_date":    eventData.HireDate.Format(time.RFC3339),
				"processed_at": time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步员工雇佣到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 员工雇佣事件同步完成: %s", eventData.EmployeeID)
	return nil
}

// handleEmployeeTerminated 处理员工终止事件
func (c *EmployeeEventConsumer) handleEmployeeTerminated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseEmployeeTerminatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析员工终止事件失败: %w", err)
	}
	
	// 更新员工状态为已终止
	_, err = c.connectionManager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (e:Employee) 
			WHERE e.id = $employee_id AND e.tenant_id = $tenant_id
			SET e.status = 'TERMINATED',
			    e.termination_date = $termination_date,
			    e.termination_reason = $reason,
			    e.terminated_at = $terminated_at,
			    e.last_event_id = $event_id
		`
		
		params := map[string]interface{}{
			"employee_id":       eventData.EmployeeID.String(),
			"tenant_id":         eventData.TenantID.String(),
			"termination_date":  eventData.TerminationDate.Format(time.RFC3339),
			"reason":            eventData.Reason,
			"terminated_at":     event.GetTimestamp().Format(time.RFC3339),
			"event_id":          event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录终止事件
		eventOp := &NodeSyncOperation{
			Label:      "EmployeeEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":          event.GetEventID(),
				"event_type":        event.GetEventType(),
				"aggregate_id":      event.GetAggregateID().String(),
				"tenant_id":         event.GetTenantID().String(),
				"timestamp":         event.GetTimestamp().Format(time.RFC3339),
				"termination_date":  eventData.TerminationDate.Format(time.RFC3339),
				"reason":           eventData.Reason,
				"processed_at":     time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步员工终止到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 员工终止事件同步完成: %s (原因: %s)", eventData.EmployeeID, eventData.Reason)
	return nil
}

// handleEmployeePhoneUpdated 处理员工电话更新事件
func (c *EmployeeEventConsumer) handleEmployeePhoneUpdated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseEmployeePhoneUpdatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析员工电话更新事件失败: %w", err)
	}
	
	// 更新员工电话信息
	_, err = c.connectionManager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (e:Employee) 
			WHERE e.id = $employee_id AND e.tenant_id = $tenant_id
			SET e.phone_number = $phone_number,
			    e.phone_updated_at = $updated_at,
			    e.last_event_id = $event_id
		`
		
		params := map[string]interface{}{
			"employee_id":   eventData.EmployeeID.String(),
			"tenant_id":     eventData.TenantID.String(),
			"phone_number":  eventData.PhoneNumber,
			"updated_at":    event.GetTimestamp().Format(time.RFC3339),
			"event_id":      event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录电话更新事件
		eventOp := &NodeSyncOperation{
			Label:      "EmployeeEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":      event.GetEventID(),
				"event_type":    event.GetEventType(),
				"aggregate_id":  event.GetAggregateID().String(),
				"tenant_id":     event.GetTenantID().String(),
				"timestamp":     event.GetTimestamp().Format(time.RFC3339),
				"phone_number":  eventData.PhoneNumber,
				"processed_at":  time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步员工电话更新到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 员工电话更新事件同步完成: %s", eventData.EmployeeID)
	return nil
}

// 事件数据解析结构体
type EmployeeCreatedEventData struct {
	TenantID       uuid.UUID
	EmployeeID     uuid.UUID
	EmployeeNumber string
	FirstName      string
	LastName       string
	Email          string
	HireDate       time.Time
	Status         string
}

type EmployeeUpdatedEventData struct {
	TenantID       uuid.UUID
	EmployeeID     uuid.UUID
	EmployeeNumber string
	UpdatedFields  map[string]interface{}
}

type EmployeeDeletedEventData struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
	Reason     string
}

type EmployeeHiredEventData struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
	HireDate   time.Time
}

type EmployeeTerminatedEventData struct {
	TenantID        uuid.UUID
	EmployeeID      uuid.UUID
	TerminationDate time.Time
	Reason          string
}

type EmployeePhoneUpdatedEventData struct {
	TenantID    uuid.UUID
	EmployeeID  uuid.UUID
	PhoneNumber string
}

// 事件解析函数
func parseEmployeeCreatedEvent(event events.DomainEvent) (*EmployeeCreatedEventData, error) {
	// 这里应该根据实际的事件结构进行解析
	// 暂时使用简化的解析逻辑
	return &EmployeeCreatedEventData{
		TenantID:       event.GetTenantID(),
		EmployeeID:     event.GetAggregateID(),
		EmployeeNumber: "解析中", // 需要从事件payload中解析
		FirstName:      "解析中",
		LastName:       "解析中",
		Email:          "解析中",
		HireDate:       event.GetTimestamp(),
		Status:         "active",
	}, nil
}

func parseEmployeeUpdatedEvent(event events.DomainEvent) (*EmployeeUpdatedEventData, error) {
	return &EmployeeUpdatedEventData{
		TenantID:       event.GetTenantID(),
		EmployeeID:     event.GetAggregateID(),
		EmployeeNumber: "解析中",
		UpdatedFields:  make(map[string]interface{}),
	}, nil
}

func parseEmployeeDeletedEvent(event events.DomainEvent) (*EmployeeDeletedEventData, error) {
	return &EmployeeDeletedEventData{
		TenantID:   event.GetTenantID(),
		EmployeeID: event.GetAggregateID(),
		Reason:     "标准删除流程",
	}, nil
}

func parseEmployeeHiredEvent(event events.DomainEvent) (*EmployeeHiredEventData, error) {
	return &EmployeeHiredEventData{
		TenantID:   event.GetTenantID(),
		EmployeeID: event.GetAggregateID(),
		HireDate:   event.GetTimestamp(),
	}, nil
}

func parseEmployeeTerminatedEvent(event events.DomainEvent) (*EmployeeTerminatedEventData, error) {
	return &EmployeeTerminatedEventData{
		TenantID:        event.GetTenantID(),
		EmployeeID:      event.GetAggregateID(),
		TerminationDate: event.GetTimestamp(),
		Reason:          "标准终止流程",
	}, nil
}

func parseEmployeePhoneUpdatedEvent(event events.DomainEvent) (*EmployeePhoneUpdatedEventData, error) {
	return &EmployeePhoneUpdatedEventData{
		TenantID:    event.GetTenantID(),
		EmployeeID:  event.GetAggregateID(),
		PhoneNumber: "解析中",
	}, nil
}

// 辅助函数
func formatUpdatedFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}
	
	result := ""
	first := true
	for key := range fields {
		if !first {
			result += ", "
		}
		result += key
		first = false
	}
	return result
}