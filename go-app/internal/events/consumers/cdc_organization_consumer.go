package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// CDCOrganizationConsumer CDC组织事件消费者
// 负责将Debezium CDC组织变更事件同步到Neo4j图数据库
type CDCOrganizationConsumer struct {
	neo4jService *service.Neo4jService
	logger       Logger
}

// DebeziumEvent Debezium CDC事件结构
type DebeziumEvent struct {
	Before      map[string]interface{} `json:"before"`
	After       map[string]interface{} `json:"after"`
	Source      map[string]interface{} `json:"source"`
	Op          string                 `json:"op"`  // c=create, u=update, d=delete, r=read(snapshot)
	TsMs        int64                  `json:"ts_ms"`
	Transaction *map[string]interface{} `json:"transaction"`
}

// NewCDCOrganizationConsumer 创建CDC组织事件消费者
func NewCDCOrganizationConsumer(neo4jService *service.Neo4jService, logger Logger) *CDCOrganizationConsumer {
	return &CDCOrganizationConsumer{
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// Handle 实现EventHandler接口 - 处理CDC事件
// 这里我们接收的是来自Kafka的原始字节数据，需要解析为Debezium格式
func (c *CDCOrganizationConsumer) Handle(ctx context.Context, event events.DomainEvent) error {
	c.logger.Info("Received CDC organization event", "event_type", event.GetEventType())
	
	// 将DomainEvent序列化获取原始数据
	eventData, err := event.Serialize()
	if err != nil {
		c.logger.Error("Failed to serialize domain event", "error", err)
		return fmt.Errorf("failed to serialize domain event: %w", err)
	}
	
	// 解析为Debezium格式
	var cdcEvent DebeziumEvent
	if err := json.Unmarshal(eventData, &cdcEvent); err != nil {
		c.logger.Error("Failed to unmarshal CDC event", "error", err)
		return fmt.Errorf("failed to unmarshal CDC event: %w", err)
	}

	c.logger.Info("Processing CDC organization event", 
		"op", cdcEvent.Op, 
		"table", c.getTableName(cdcEvent),
		"event_id", event.GetEventID(),
	)

	switch cdcEvent.Op {
	case "c": // create
		return c.handleOrganizationCreated(ctx, cdcEvent)
	case "u": // update
		return c.handleOrganizationUpdated(ctx, cdcEvent)
	case "d": // delete
		return c.handleOrganizationDeleted(ctx, cdcEvent)
	case "r": // read (snapshot)
		return c.handleOrganizationCreated(ctx, cdcEvent) // 处理快照数据
	default:
		c.logger.Warn("Unknown CDC operation", "op", cdcEvent.Op)
		return nil
	}
}

// GetEventType 返回处理的事件类型
func (c *CDCOrganizationConsumer) GetEventType() string {
	return "organization_db.public.organization_units"
}

// GetHandlerName 返回处理器名称
func (c *CDCOrganizationConsumer) GetHandlerName() string {
	return "CDCOrganizationConsumer"
}

// ConsumeRawEvent 处理原始CDC事件数据（直接从Kafka消费）
func (c *CDCOrganizationConsumer) ConsumeRawEvent(ctx context.Context, eventData []byte) error {
	var cdcEvent DebeziumEvent
	if err := json.Unmarshal(eventData, &cdcEvent); err != nil {
		c.logger.Error("Failed to unmarshal CDC event", "error", err)
		return fmt.Errorf("failed to unmarshal CDC event: %w", err)
	}

	c.logger.Info("Processing raw CDC organization event", "op", cdcEvent.Op, "table", c.getTableName(cdcEvent))

	switch cdcEvent.Op {
	case "c": // create
		return c.handleOrganizationCreated(ctx, cdcEvent)
	case "u": // update
		return c.handleOrganizationUpdated(ctx, cdcEvent)
	case "d": // delete
		return c.handleOrganizationDeleted(ctx, cdcEvent)
	case "r": // read (snapshot)
		return c.handleOrganizationCreated(ctx, cdcEvent) // 处理快照数据
	default:
		c.logger.Warn("Unknown CDC operation", "op", cdcEvent.Op)
		return nil
	}
}

// handleOrganizationCreated 处理组织创建事件
func (c *CDCOrganizationConsumer) handleOrganizationCreated(ctx context.Context, event DebeziumEvent) error {
	if event.After == nil {
		c.logger.Warn("No 'after' data in create event")
		return nil
	}

	// 从after字段中提取组织信息
	org, err := c.extractOrganizationFromCDC(event.After)
	if err != nil {
		return fmt.Errorf("failed to extract organization from CDC event: %w", err)
	}

	// 转换为Neo4j服务所需的OrganizationNode结构
	orgNode := service.OrganizationNode{
		ID:          org.ID.String(),
		TenantID:    org.TenantID.String(),
		UnitType:    org.UnitType,
		Name:        org.Name,
		Description: c.getStringValue(org.Description),
		Status:      org.Status,
		IsActive:    org.IsActive,
		Properties: map[string]interface{}{
			"created_at": org.CreatedAt.Format(time.RFC3339),
			"updated_at": org.UpdatedAt.Format(time.RFC3339),
		},
	}

	if org.ParentUnitID != nil {
		orgNode.ParentUnitID = org.ParentUnitID.String()
	}

	if org.Profile != nil {
		orgNode.Properties["profile"] = org.Profile
	}

	// 为Neo4j操作创建新的上下文，避免使用可能已取消的事件上下文
	neo4jCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用Neo4j服务同步组织
	err = c.neo4jService.SyncOrganization(neo4jCtx, orgNode)
	if err != nil {
		c.logger.Error("Failed to sync organization to Neo4j", 
			"org_id", org.ID,
			"error", err,
		)
		return fmt.Errorf("failed to sync organization to Neo4j: %w", err)
	}

	c.logger.Info("Organization created successfully in Neo4j", "org_id", org.ID, "name", org.Name)
	return nil
}

// handleOrganizationUpdated 处理组织更新事件
func (c *CDCOrganizationConsumer) handleOrganizationUpdated(ctx context.Context, event DebeziumEvent) error {
	if event.After == nil {
		c.logger.Warn("No 'after' data in update event")
		return nil
	}

	org, err := c.extractOrganizationFromCDC(event.After)
	if err != nil {
		return fmt.Errorf("failed to extract organization from CDC event: %w", err)
	}

	// 转换为Neo4j服务所需的OrganizationNode结构
	orgNode := service.OrganizationNode{
		ID:          org.ID.String(),
		TenantID:    org.TenantID.String(),
		UnitType:    org.UnitType,
		Name:        org.Name,
		Description: c.getStringValue(org.Description),
		Status:      org.Status,
		IsActive:    org.IsActive,
		Properties: map[string]interface{}{
			"created_at": org.CreatedAt.Format(time.RFC3339),
			"updated_at": org.UpdatedAt.Format(time.RFC3339),
		},
	}

	if org.ParentUnitID != nil {
		orgNode.ParentUnitID = org.ParentUnitID.String()
	}

	if org.Profile != nil {
		orgNode.Properties["profile"] = org.Profile
	}

	// 为Neo4j操作创建新的上下文
	neo4jCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用Neo4j服务同步组织
	err = c.neo4jService.SyncOrganization(neo4jCtx, orgNode)
	if err != nil {
		c.logger.Error("Failed to sync updated organization to Neo4j", 
			"org_id", org.ID,
			"error", err,
		)
		return fmt.Errorf("failed to sync updated organization to Neo4j: %w", err)
	}

	c.logger.Info("Organization updated successfully in Neo4j", "org_id", org.ID, "name", org.Name)
	return nil
}

// handleOrganizationDeleted 处理组织删除事件
func (c *CDCOrganizationConsumer) handleOrganizationDeleted(ctx context.Context, event DebeziumEvent) error {
	if event.Before == nil {
		c.logger.Warn("No 'before' data in delete event")
		return nil
	}

	// 从before字段中获取要删除的组织ID
	idStr, ok := event.Before["id"].(string)
	if !ok {
		return fmt.Errorf("invalid organization ID in delete event")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("failed to parse organization ID: %w", err)
	}

	tenantIDStr, ok := event.Before["tenant_id"].(string)
	if !ok {
		return fmt.Errorf("invalid tenant ID in delete event")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	// 为Neo4j操作创建新的上下文
	neo4jCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用Neo4j服务删除组织
	err = c.neo4jService.DeleteOrganization(neo4jCtx, id.String(), tenantID.String())
	if err != nil {
		c.logger.Error("Failed to delete organization from Neo4j", 
			"org_id", id,
			"tenant_id", tenantID,
			"error", err,
		)
		return fmt.Errorf("failed to delete organization from Neo4j: %w", err)
	}

	c.logger.Info("Organization deleted from Neo4j", "org_id", id, "tenant_id", tenantID)
	return nil
}

// extractOrganizationFromCDC 从CDC事件中提取组织信息
func (c *CDCOrganizationConsumer) extractOrganizationFromCDC(data map[string]interface{}) (*OrganizationData, error) {
	// 提取ID
	idStr, ok := data["id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid organization ID")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse organization ID: %w", err)
	}

	// 提取TenantID
	tenantIDStr, ok := data["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid tenant ID")
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	org := &OrganizationData{
		ID:       id,
		TenantID: tenantID,
	}

	// 提取其他字段
	if unitType, ok := data["unit_type"].(string); ok {
		org.UnitType = unitType
	}

	if name, ok := data["name"].(string); ok {
		org.Name = name
	}

	if description, ok := data["description"].(string); ok && description != "" {
		org.Description = &description
	}

	// 处理status字段
	if status, ok := data["status"].(string); ok && status != "" {
		org.Status = status
	} else if isActive, ok := data["is_active"].(bool); ok {
		if isActive {
			org.Status = "ACTIVE"
		} else {
			org.Status = "INACTIVE"
		}
	} else {
		org.Status = "ACTIVE" // 默认值
	}

	if isActive, ok := data["is_active"].(bool); ok {
		org.IsActive = isActive
	} else {
		org.IsActive = true // 默认值
	}

	// 处理父组织ID
	if parentUnitIDStr, ok := data["parent_unit_id"].(string); ok && parentUnitIDStr != "" {
		if parentUnitID, err := uuid.Parse(parentUnitIDStr); err == nil {
			org.ParentUnitID = &parentUnitID
		}
	}

	// 处理时间字段
	if createdAtStr, ok := data["created_at"].(string); ok {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			org.CreatedAt = createdAt
		}
	}

	if updatedAtStr, ok := data["updated_at"].(string); ok {
		if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			org.UpdatedAt = updatedAt
		}
	}

	// 设置默认时间
	if org.CreatedAt.IsZero() {
		org.CreatedAt = time.Now()
	}
	if org.UpdatedAt.IsZero() {
		org.UpdatedAt = time.Now()
	}

	// 处理profile字段
	if profile, ok := data["profile"]; ok && profile != nil {
		org.Profile = profile
	}

	return org, nil
}

// OrganizationData 组织数据结构
type OrganizationData struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	Status       string                 `json:"status"`
	IsActive     bool                   `json:"is_active"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id"`
	Profile      interface{}            `json:"profile"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// getTableName 获取表名
func (c *CDCOrganizationConsumer) getTableName(event DebeziumEvent) string {
	if source, ok := event.Source["table"]; ok {
		return source.(string)
	}
	return "unknown"
}

// getStringValue 安全获取字符串值
func (c *CDCOrganizationConsumer) getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}