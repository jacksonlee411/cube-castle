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

// OrganizationEventConsumer 组织事件消费者
type OrganizationEventConsumer struct {
	*BaseEventConsumer
}

// NewOrganizationEventConsumer 创建组织事件消费者
func NewOrganizationEventConsumer(connMgr ConnectionManagerInterface) *OrganizationEventConsumer {
	base := NewBaseEventConsumer(connMgr, "organization.*")
	
	return &OrganizationEventConsumer{
		BaseEventConsumer: base,
	}
}

// ConsumeEvent 消费组织事件
func (c *OrganizationEventConsumer) ConsumeEvent(ctx context.Context, event events.DomainEvent) error {
	log.Printf("🔄 处理组织事件: %s (ID: %s)", event.GetEventType(), event.GetEventID())
	
	switch event.GetEventType() {
	case "organization.created":
		return c.handleOrganizationCreated(ctx, event)
	case "organization.updated":
		return c.handleOrganizationUpdated(ctx, event)
	case "organization.deleted":
		return c.handleOrganizationDeleted(ctx, event)
	case "organization.restructured":
		return c.handleOrganizationRestructured(ctx, event)
	case "organization.activated":
		return c.handleOrganizationActivated(ctx, event)
	case "organization.deactivated":
		return c.handleOrganizationDeactivated(ctx, event)
	default:
		log.Printf("⚠️ 未知的组织事件类型: %s", event.GetEventType())
		return nil
	}
}

// handleOrganizationCreated 处理组织创建事件
func (c *OrganizationEventConsumer) handleOrganizationCreated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseOrganizationCreatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析组织创建事件失败: %w", err)
	}
	
	// 执行组织节点创建和层级关系建立
	_, err = c.connectionManager.ExecuteWrite(ctx, func(ctx context.Context, tx neo4j.ManagedTransaction) (any, error) {
		// 创建组织节点
		orgSyncOp := &NodeSyncOperation{
			Label:      "Organization",
			Operation:  "CREATE",
			UniqueKeys: []string{"id", "tenant_id"},
			Properties: map[string]interface{}{
				"id":              eventData.OrganizationID.String(),
				"tenant_id":       eventData.TenantID.String(),
				"name":            eventData.Name,
				"description":     eventData.Description,
				"org_type":        eventData.OrgType,
				"level":           eventData.Level,
				"is_active":       true,
				"created_at":      event.GetTimestamp().Format(time.RFC3339),
				"event_id":        event.GetEventID(),
				"event_version":   event.GetEventVersion(),
			},
		}
		
		if err := orgSyncOp.Execute(ctx, tx); err != nil {
			return nil, fmt.Errorf("创建组织节点失败: %w", err)
		}
		
		// 如果有父组织，建立层级关系
		if eventData.ParentOrgID != nil {
			relationshipCypher := `
				MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
				MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
				CREATE (parent)-[:CONTAINS {
					created_at: $created_at,
					created_by_event: $event_id
				}]->(child)
			`
			
			relationshipParams := map[string]interface{}{
				"parent_id":  eventData.ParentOrgID.String(),
				"child_id":   eventData.OrganizationID.String(),
				"tenant_id":  eventData.TenantID.String(),
				"created_at": event.GetTimestamp().Format(time.RFC3339),
				"event_id":   event.GetEventID(),
			}
			
			_, err := tx.Run(ctx, relationshipCypher, relationshipParams)
			if err != nil {
				return nil, fmt.Errorf("创建组织层级关系失败: %w", err)
			}
		}
		
		// 创建事件记录节点
		eventOp := &NodeSyncOperation{
			Label:      "OrganizationEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":       event.GetEventID(),
				"event_type":     event.GetEventType(),
				"aggregate_id":   event.GetAggregateID().String(),
				"tenant_id":      event.GetTenantID().String(),
				"timestamp":      event.GetTimestamp().Format(time.RFC3339),
				"version":        event.GetEventVersion(),
				"processed_at":   time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步组织创建到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 组织创建事件同步完成: %s (名称: %s)", eventData.OrganizationID, eventData.Name)
	return nil
}

// handleOrganizationUpdated 处理组织更新事件
func (c *OrganizationEventConsumer) handleOrganizationUpdated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseOrganizationUpdatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析组织更新事件失败: %w", err)
	}
	
	// 准备更新属性
	updateProps := map[string]interface{}{
		"id":        eventData.OrganizationID.String(),
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
		Label:      "Organization",
		Operation:  "UPDATE",
		UniqueKeys: []string{"id", "tenant_id"},
		Properties: updateProps,
	}
	
	// 执行同步
	_, err = c.connectionManager.ExecuteWrite(ctx, func(ctx context.Context, tx neo4j.ManagedTransaction) (any, error) {
		if err := syncOp.Execute(ctx, tx); err != nil {
			return nil, err
		}
		
		// 记录更新事件
		eventOp := &NodeSyncOperation{
			Label:      "OrganizationEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":       event.GetEventID(),
				"event_type":     event.GetEventType(),
				"aggregate_id":   event.GetAggregateID().String(),
				"tenant_id":      event.GetTenantID().String(),
				"timestamp":      event.GetTimestamp().Format(time.RFC3339),
				"version":        event.GetEventVersion(),
				"updated_fields": formatUpdatedFields(eventData.UpdatedFields),
				"processed_at":   time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步组织更新到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 组织更新事件同步完成: %s", eventData.OrganizationID)
	return nil
}

// handleOrganizationDeleted 处理组织删除事件
func (c *OrganizationEventConsumer) handleOrganizationDeleted(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseOrganizationDeletedEvent(event)
	if err != nil {
		return fmt.Errorf("解析组织删除事件失败: %w", err)
	}
	
	// 执行软删除并处理级联关系
	_, err = c.connectionManager.ExecuteWrite(ctx, func(ctx context.Context, tx neo4j.ManagedTransaction) (any, error) {
		// 软删除组织节点
		cypher := `
			MATCH (org:Organization) 
			WHERE org.id = $org_id AND org.tenant_id = $tenant_id
			SET org.is_active = false,
			    org.is_deleted = true,
			    org.deleted_at = $deleted_at,
			    org.deleted_by_event_id = $event_id,
			    org.deletion_reason = $reason
		`
		
		params := map[string]interface{}{
			"org_id":     eventData.OrganizationID.String(),
			"tenant_id":  eventData.TenantID.String(),
			"deleted_at": event.GetTimestamp().Format(time.RFC3339),
			"event_id":   event.GetEventID(),
			"reason":     eventData.Reason,
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 标记与此组织相关的关系为已删除
		relationshipCypher := `
			MATCH (org:Organization {id: $org_id, tenant_id: $tenant_id})-[r]-(related)
			SET r.is_deleted = true,
			    r.deleted_at = $deleted_at,
			    r.deleted_by_event_id = $event_id
		`
		
		_, err = tx.Run(ctx, relationshipCypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录删除事件
		eventOp := &NodeSyncOperation{
			Label:      "OrganizationEvent",
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
		return fmt.Errorf("同步组织删除到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 组织删除事件同步完成: %s (原因: %s)", eventData.OrganizationID, eventData.Reason)
	return nil
}

// handleOrganizationRestructured 处理组织重构事件
func (c *OrganizationEventConsumer) handleOrganizationRestructured(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseOrganizationRestructuredEvent(event)
	if err != nil {
		return fmt.Errorf("解析组织重构事件失败: %w", err)
	}
	
	// 执行组织结构重组
	_, err = c.connectionManager.ExecuteWrite(ctx, func(ctx context.Context, tx neo4j.ManagedTransaction) (any, error) {
		// 如果有新的父组织关系，先删除旧关系
		if eventData.NewParentOrgID != nil {
			// 删除现有的父级关系
			deleteOldRelationCypher := `
				MATCH (org:Organization {id: $org_id, tenant_id: $tenant_id})<-[r:CONTAINS]-(parent:Organization)
				SET r.is_deleted = true,
				    r.deleted_at = $restructured_at,
				    r.deleted_by_event_id = $event_id
			`
			
			deleteParams := map[string]interface{}{
				"org_id":          eventData.OrganizationID.String(),
				"tenant_id":       eventData.TenantID.String(),
				"restructured_at": event.GetTimestamp().Format(time.RFC3339),
				"event_id":        event.GetEventID(),
			}
			
			_, err := tx.Run(ctx, deleteOldRelationCypher, deleteParams)
			if err != nil {
				return nil, fmt.Errorf("删除旧父级关系失败: %w", err)
			}
			
			// 创建新的父级关系
			createNewRelationCypher := `
				MATCH (newParent:Organization {id: $new_parent_id, tenant_id: $tenant_id})
				MATCH (org:Organization {id: $org_id, tenant_id: $tenant_id})
				CREATE (newParent)-[:CONTAINS {
					created_at: $restructured_at,
					created_by_event: $event_id,
					restructure_type: $restructure_type
				}]->(org)
			`
			
			createParams := map[string]interface{}{
				"new_parent_id":     eventData.NewParentOrgID.String(),
				"org_id":            eventData.OrganizationID.String(),
				"tenant_id":         eventData.TenantID.String(),
				"restructured_at":   event.GetTimestamp().Format(time.RFC3339),
				"event_id":          event.GetEventID(),
				"restructure_type":  eventData.RestructureType,
			}
			
			_, err = tx.Run(ctx, createNewRelationCypher, createParams)
			if err != nil {
				return nil, fmt.Errorf("创建新父级关系失败: %w", err)
			}
		}
		
		// 更新组织节点的重构信息
		updateOrgCypher := `
			MATCH (org:Organization {id: $org_id, tenant_id: $tenant_id})
			SET org.last_restructured_at = $restructured_at,
			    org.restructure_type = $restructure_type,
			    org.restructure_reason = $reason,
			    org.last_event_id = $event_id
		`
		
		updateParams := map[string]interface{}{
			"org_id":           eventData.OrganizationID.String(),
			"tenant_id":        eventData.TenantID.String(),
			"restructured_at":  event.GetTimestamp().Format(time.RFC3339),
			"restructure_type": eventData.RestructureType,
			"reason":           eventData.Reason,
			"event_id":         event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, updateOrgCypher, updateParams)
		if err != nil {
			return nil, fmt.Errorf("更新组织重构信息失败: %w", err)
		}
		
		// 记录重构事件
		eventOp := &NodeSyncOperation{
			Label:      "OrganizationEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":         event.GetEventID(),
				"event_type":       event.GetEventType(),
				"aggregate_id":     event.GetAggregateID().String(),
				"tenant_id":        event.GetTenantID().String(),
				"timestamp":        event.GetTimestamp().Format(time.RFC3339),
				"restructure_type": eventData.RestructureType,
				"reason":           eventData.Reason,
				"processed_at":     time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步组织重构到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 组织重构事件同步完成: %s (类型: %s)", eventData.OrganizationID, eventData.RestructureType)
	return nil
}

// handleOrganizationActivated 处理组织激活事件
func (c *OrganizationEventConsumer) handleOrganizationActivated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseOrganizationActivatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析组织激活事件失败: %w", err)
	}
	
	// 激活组织
	_, err = c.connectionManager.ExecuteWrite(ctx, func(ctx context.Context, tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (org:Organization) 
			WHERE org.id = $org_id AND org.tenant_id = $tenant_id
			SET org.is_active = true,
			    org.activated_at = $activated_at,
			    org.activation_reason = $reason,
			    org.last_event_id = $event_id
		`
		
		params := map[string]interface{}{
			"org_id":      eventData.OrganizationID.String(),
			"tenant_id":   eventData.TenantID.String(),
			"activated_at": event.GetTimestamp().Format(time.RFC3339),
			"reason":      eventData.Reason,
			"event_id":    event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录激活事件
		eventOp := &NodeSyncOperation{
			Label:      "OrganizationEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":     event.GetEventID(),
				"event_type":   event.GetEventType(),
				"aggregate_id": event.GetAggregateID().String(),
				"tenant_id":    event.GetTenantID().String(),
				"timestamp":    event.GetTimestamp().Format(time.RFC3339),
				"reason":       eventData.Reason,
				"processed_at": time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步组织激活到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 组织激活事件同步完成: %s", eventData.OrganizationID)
	return nil
}

// handleOrganizationDeactivated 处理组织停用事件
func (c *OrganizationEventConsumer) handleOrganizationDeactivated(ctx context.Context, event events.DomainEvent) error {
	// 解析事件数据
	eventData, err := parseOrganizationDeactivatedEvent(event)
	if err != nil {
		return fmt.Errorf("解析组织停用事件失败: %w", err)
	}
	
	// 停用组织
	_, err = c.connectionManager.ExecuteWrite(ctx, func(ctx context.Context, tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (org:Organization) 
			WHERE org.id = $org_id AND org.tenant_id = $tenant_id
			SET org.is_active = false,
			    org.deactivated_at = $deactivated_at,
			    org.deactivation_reason = $reason,
			    org.last_event_id = $event_id
		`
		
		params := map[string]interface{}{
			"org_id":        eventData.OrganizationID.String(),
			"tenant_id":     eventData.TenantID.String(),
			"deactivated_at": event.GetTimestamp().Format(time.RFC3339),
			"reason":        eventData.Reason,
			"event_id":      event.GetEventID(),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 记录停用事件
		eventOp := &NodeSyncOperation{
			Label:      "OrganizationEvent",
			Operation:  "CREATE",
			UniqueKeys: []string{"event_id"},
			Properties: map[string]interface{}{
				"event_id":     event.GetEventID(),
				"event_type":   event.GetEventType(),
				"aggregate_id": event.GetAggregateID().String(),
				"tenant_id":    event.GetTenantID().String(),
				"timestamp":    event.GetTimestamp().Format(time.RFC3339),
				"reason":       eventData.Reason,
				"processed_at": time.Now().Format(time.RFC3339),
			},
		}
		
		return nil, eventOp.Execute(ctx, tx)
	})
	
	if err != nil {
		return fmt.Errorf("同步组织停用到Neo4j失败: %w", err)
	}
	
	log.Printf("✅ 组织停用事件同步完成: %s", eventData.OrganizationID)
	return nil
}

// 组织事件数据结构体
type OrganizationCreatedEventData struct {
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Description    string
	OrgType        string
	Level          int
	ParentOrgID    *uuid.UUID
}

type OrganizationUpdatedEventData struct {
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	UpdatedFields  map[string]interface{}
}

type OrganizationDeletedEventData struct {
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	Reason         string
}

type OrganizationRestructuredEventData struct {
	TenantID          uuid.UUID
	OrganizationID    uuid.UUID
	NewParentOrgID    *uuid.UUID
	RestructureType   string
	Reason            string
}

type OrganizationActivatedEventData struct {
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	Reason         string
}

type OrganizationDeactivatedEventData struct {
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	Reason         string
}

// 组织事件解析函数
func parseOrganizationCreatedEvent(event events.DomainEvent) (*OrganizationCreatedEventData, error) {
	// 这里应该根据实际的事件结构进行解析
	return &OrganizationCreatedEventData{
		TenantID:       event.GetTenantID(),
		OrganizationID: event.GetAggregateID(),
		Name:           "解析中", // 需要从事件payload中解析
		Description:    "解析中",
		OrgType:        "部门",
		Level:          1,
		ParentOrgID:    nil,
	}, nil
}

func parseOrganizationUpdatedEvent(event events.DomainEvent) (*OrganizationUpdatedEventData, error) {
	return &OrganizationUpdatedEventData{
		TenantID:       event.GetTenantID(),
		OrganizationID: event.GetAggregateID(),
		UpdatedFields:  make(map[string]interface{}),
	}, nil
}

func parseOrganizationDeletedEvent(event events.DomainEvent) (*OrganizationDeletedEventData, error) {
	return &OrganizationDeletedEventData{
		TenantID:       event.GetTenantID(),
		OrganizationID: event.GetAggregateID(),
		Reason:         "标准删除流程",
	}, nil
}

func parseOrganizationRestructuredEvent(event events.DomainEvent) (*OrganizationRestructuredEventData, error) {
	return &OrganizationRestructuredEventData{
		TenantID:          event.GetTenantID(),
		OrganizationID:    event.GetAggregateID(),
		NewParentOrgID:    nil,
		RestructureType:   "重组",
		Reason:           "组织结构优化",
	}, nil
}

func parseOrganizationActivatedEvent(event events.DomainEvent) (*OrganizationActivatedEventData, error) {
	return &OrganizationActivatedEventData{
		TenantID:       event.GetTenantID(),
		OrganizationID: event.GetAggregateID(),
		Reason:         "标准激活流程",
	}, nil
}

func parseOrganizationDeactivatedEvent(event events.DomainEvent) (*OrganizationDeactivatedEventData, error) {
	return &OrganizationDeactivatedEventData{
		TenantID:       event.GetTenantID(),
		OrganizationID: event.GetAggregateID(),
		Reason:         "标准停用流程",
	}, nil
}