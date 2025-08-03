package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// OrganizationEventConsumer 组织事件消费者
// 负责将组织变更事件同步到Neo4j图数据库
type OrganizationEventConsumer struct {
	driver neo4j.DriverWithContext
	logger Logger
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// OrganizationEvent 组织事件结构
type OrganizationEvent struct {
	EventID      uuid.UUID              `json:"event_id"`
	EventType    string                 `json:"event_type"`
	AggregateID  uuid.UUID              `json:"aggregate_id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	Timestamp    time.Time              `json:"timestamp"`
	EventVersion string                 `json:"event_version"`
	Payload      map[string]interface{} `json:"payload"`
}

// NewOrganizationEventConsumer 创建组织事件消费者
func NewOrganizationEventConsumer(driver neo4j.DriverWithContext, logger Logger) *OrganizationEventConsumer {
	return &OrganizationEventConsumer{
		driver: driver,
		logger: logger,
	}
}

// ConsumeEvent 处理组织事件
func (c *OrganizationEventConsumer) ConsumeEvent(ctx context.Context, eventData []byte) error {
	var event OrganizationEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		c.logger.Error("Failed to unmarshal organization event", "error", err)
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	c.logger.Info("Processing organization event", "event_type", event.EventType, "aggregate_id", event.AggregateID)

	switch event.EventType {
	case "organization.created":
		return c.handleOrganizationCreated(ctx, event)
	case "organization.updated":
		return c.handleOrganizationUpdated(ctx, event)
	case "organization.deleted":
		return c.handleOrganizationDeleted(ctx, event)
	case "organization.moved":
		return c.handleOrganizationMoved(ctx, event)
	case "organization.activated":
		return c.handleOrganizationActivated(ctx, event)
	case "organization.deactivated":
		return c.handleOrganizationDeactivated(ctx, event)
	default:
		c.logger.Warn("Unknown organization event type", "event_type", event.EventType)
		return nil // 忽略未知事件类型
	}
}

// handleOrganizationCreated 处理组织创建事件
func (c *OrganizationEventConsumer) handleOrganizationCreated(ctx context.Context, event OrganizationEvent) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	// 从事件负载中提取组织信息
	org, err := c.extractOrganizationFromPayload(event.Payload, event.AggregateID, event.TenantID)
	if err != nil {
		return fmt.Errorf("failed to extract organization from payload: %w", err)
	}

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 创建组织节点
		createNodeCypher := `
			CREATE (o:Organization {
				id: $id,
				tenant_id: $tenant_id,
				unit_type: $unit_type,
				name: $name,
				description: $description,
				status: $status,
				profile: $profile,
				level: $level,
				employee_count: $employee_count,
				is_active: $is_active,
				created_at: $created_at,
				updated_at: $updated_at
			})`

		profileJSON, _ := json.Marshal(org.Profile)
		params := map[string]any{
			"id":             org.ID.String(),
			"tenant_id":      org.TenantID.String(),
			"unit_type":      org.UnitType,
			"name":           org.Name,
			"description":    org.Description,
			"status":         org.Status,
			"profile":        string(profileJSON),
			"level":          org.Level,
			"employee_count": org.EmployeeCount,
			"is_active":      org.IsActive,
			"created_at":     org.CreatedAt.Format(time.RFC3339),
			"updated_at":     org.UpdatedAt.Format(time.RFC3339),
		}

		_, err := tx.Run(ctx, createNodeCypher, params)
		if err != nil {
			return nil, err
		}

		// 如果有父组织，创建关系
		if org.ParentUnitID != nil {
			createRelationshipCypher := `
				MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
				MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
				CREATE (parent)-[:PARENT_OF]->(child)`

			_, err := tx.Run(ctx, createRelationshipCypher, map[string]any{
				"parent_id":  org.ParentUnitID.String(),
				"child_id":   org.ID.String(),
				"tenant_id":  org.TenantID.String(),
			})
			if err != nil {
				c.logger.Warn("Failed to create parent relationship", "error", err, "parent_id", org.ParentUnitID, "child_id", org.ID)
				// 不返回错误，允许节点创建成功
			}
		}

		return "created", nil
	})

	if err != nil {
		c.logger.Error("Failed to create organization in Neo4j", "error", err, "org_id", org.ID)
		return fmt.Errorf("failed to create organization in Neo4j: %w", err)
	}

	c.logger.Info("Organization created successfully in Neo4j", "org_id", org.ID, "name", org.Name)
	return nil
}

// handleOrganizationUpdated 处理组织更新事件
func (c *OrganizationEventConsumer) handleOrganizationUpdated(ctx context.Context, event OrganizationEvent) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	// 提取更新字段
	changes := event.Payload
	if len(changes) == 0 {
		c.logger.Warn("No changes in organization update event", "org_id", event.AggregateID)
		return nil
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 构建动态更新语句
		setParts := []string{}
		params := map[string]any{
			"id":        event.AggregateID.String(),
			"tenant_id": event.TenantID.String(),
		}

		for field, value := range changes {
			switch field {
			case "name":
				setParts = append(setParts, "o.name = $name")
				params["name"] = value
			case "description":
				setParts = append(setParts, "o.description = $description")
				params["description"] = value
			case "status":
				setParts = append(setParts, "o.status = $status")
				params["status"] = value
			case "is_active":
				setParts = append(setParts, "o.is_active = $is_active")
				params["is_active"] = value
			case "profile":
				profileJSON, _ := json.Marshal(value)
				setParts = append(setParts, "o.profile = $profile")
				params["profile"] = string(profileJSON)
			case "level":
				setParts = append(setParts, "o.level = $level")
				params["level"] = value
			case "employee_count":
				setParts = append(setParts, "o.employee_count = $employee_count")
				params["employee_count"] = value
			}
		}

		if len(setParts) > 0 {
			setParts = append(setParts, "o.updated_at = $updated_at")
			params["updated_at"] = event.Timestamp.Format(time.RFC3339)

			cypher := fmt.Sprintf(`
				MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
				SET %s`, strings.Join(setParts, ", "))

			_, err := tx.Run(ctx, cypher, params)
			return "updated", err
		}

		return "no_changes", nil
	})

	if err != nil {
		c.logger.Error("Failed to update organization in Neo4j", "error", err, "org_id", event.AggregateID)
		return fmt.Errorf("failed to update organization in Neo4j: %w", err)
	}

	c.logger.Info("Organization updated successfully in Neo4j", "org_id", event.AggregateID)
	return nil
}

// handleOrganizationDeleted 处理组织删除事件
func (c *OrganizationEventConsumer) handleOrganizationDeleted(ctx context.Context, event OrganizationEvent) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	deleteResult, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 首先删除所有相关关系，然后删除节点
		cypher := `
			MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
			DETACH DELETE o`

		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":        event.AggregateID.String(),
			"tenant_id": event.TenantID.String(),
		})

		if err != nil {
			return nil, err
		}

		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}

		return summary.Counters().NodesDeleted(), nil
	})

	if err != nil {
		c.logger.Error("Failed to delete organization in Neo4j", "error", err, "org_id", event.AggregateID)
		return fmt.Errorf("failed to delete organization in Neo4j: %w", err)
	}

	nodesDeleted := deleteResult.(int)
	if nodesDeleted == 0 {
		c.logger.Warn("Organization not found for deletion in Neo4j", "org_id", event.AggregateID)
	} else {
		c.logger.Info("Organization deleted successfully in Neo4j", "org_id", event.AggregateID, "nodes_deleted", nodesDeleted)
	}

	return nil
}

// handleOrganizationMoved 处理组织移动事件
func (c *OrganizationEventConsumer) handleOrganizationMoved(ctx context.Context, event OrganizationEvent) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	newParentID, _ := event.Payload["new_parent_id"].(string)
	newLevel, _ := event.Payload["new_level"].(float64)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 删除旧的父子关系
		deleteOldRelCypher := `
			MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
			MATCH (parent)-[:PARENT_OF]->(o)
			DELETE r`

		_, err := tx.Run(ctx, deleteOldRelCypher, map[string]any{
			"id":        event.AggregateID.String(),
			"tenant_id": event.TenantID.String(),
		})
		if err != nil {
			c.logger.Warn("Failed to delete old parent relationship", "error", err)
		}

		// 更新组织层级
		updateLevelCypher := `
			MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
			SET o.level = $level, o.updated_at = $updated_at`

		_, err = tx.Run(ctx, updateLevelCypher, map[string]any{
			"id":         event.AggregateID.String(),
			"tenant_id":  event.TenantID.String(),
			"level":      int(newLevel),
			"updated_at": event.Timestamp.Format(time.RFC3339),
		})
		if err != nil {
			return nil, err
		}

		// 如果有新父组织，创建新关系
		if newParentID != "" {
			createNewRelCypher := `
				MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
				MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
				CREATE (parent)-[:PARENT_OF]->(child)`

			_, err = tx.Run(ctx, createNewRelCypher, map[string]any{
				"parent_id": newParentID,
				"child_id":  event.AggregateID.String(),
				"tenant_id": event.TenantID.String(),
			})
			if err != nil {
				return nil, err
			}
		}

		return "moved", nil
	})

	if err != nil {
		c.logger.Error("Failed to move organization in Neo4j", "error", err, "org_id", event.AggregateID)
		return fmt.Errorf("failed to move organization in Neo4j: %w", err)
	}

	c.logger.Info("Organization moved successfully in Neo4j", "org_id", event.AggregateID, "new_parent_id", newParentID)
	return nil
}

// handleOrganizationActivated 处理组织激活事件
func (c *OrganizationEventConsumer) handleOrganizationActivated(ctx context.Context, event OrganizationEvent) error {
	return c.updateOrganizationStatus(ctx, event, "ACTIVE", true)
}

// handleOrganizationDeactivated 处理组织停用事件
func (c *OrganizationEventConsumer) handleOrganizationDeactivated(ctx context.Context, event OrganizationEvent) error {
	return c.updateOrganizationStatus(ctx, event, "INACTIVE", false)
}

// updateOrganizationStatus 更新组织状态
func (c *OrganizationEventConsumer) updateOrganizationStatus(ctx context.Context, event OrganizationEvent, status string, isActive bool) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
			SET o.status = $status, o.is_active = $is_active, o.updated_at = $updated_at`

		_, err := tx.Run(ctx, cypher, map[string]any{
			"id":         event.AggregateID.String(),
			"tenant_id":  event.TenantID.String(),
			"status":     status,
			"is_active":  isActive,
			"updated_at": event.Timestamp.Format(time.RFC3339),
		})

		return "status_updated", err
	})

	if err != nil {
		c.logger.Error("Failed to update organization status in Neo4j", "error", err, "org_id", event.AggregateID, "status", status)
		return fmt.Errorf("failed to update organization status in Neo4j: %w", err)
	}

	c.logger.Info("Organization status updated successfully in Neo4j", "org_id", event.AggregateID, "status", status)
	return nil
}

// extractOrganizationFromPayload 从事件负载中提取组织信息
func (c *OrganizationEventConsumer) extractOrganizationFromPayload(payload map[string]interface{}, aggregateID, tenantID uuid.UUID) (*repositories.Organization, error) {
	org := &repositories.Organization{
		ID:       aggregateID,
		TenantID: tenantID,
	}

	// 提取基本字段
	if unitType, ok := payload["unit_type"].(string); ok {
		org.UnitType = unitType
	}

	if name, ok := payload["name"].(string); ok {
		org.Name = name
	}

	if description, ok := payload["description"].(string); ok {
		org.Description = &description
	}

	if status, ok := payload["status"].(string); ok {
		org.Status = status
	}

	if level, ok := payload["level"].(float64); ok {
		org.Level = int(level)
	}

	if employeeCount, ok := payload["employee_count"].(float64); ok {
		org.EmployeeCount = int(employeeCount)
	}

	if isActive, ok := payload["is_active"].(bool); ok {
		org.IsActive = isActive
	}

	// 提取父组织ID
	if parentUnitIDStr, ok := payload["parent_unit_id"].(string); ok && parentUnitIDStr != "" {
		if parentUnitID, err := uuid.Parse(parentUnitIDStr); err == nil {
			org.ParentUnitID = &parentUnitID
		}
	}

	// 提取Profile
	if profile, ok := payload["profile"].(map[string]interface{}); ok {
		org.Profile = profile
	}

	// 提取时间字段
	if createdAtStr, ok := payload["created_at"].(string); ok {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			org.CreatedAt = createdAt
		}
	}

	if updatedAtStr, ok := payload["updated_at"].(string); ok {
		if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			org.UpdatedAt = updatedAt
		}
	}

	// 设置默认值
	if org.CreatedAt.IsZero() {
		org.CreatedAt = time.Now()
	}
	if org.UpdatedAt.IsZero() {
		org.UpdatedAt = time.Now()
	}

	return org, nil
}