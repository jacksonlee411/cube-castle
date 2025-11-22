// Package events 定义 HRMS 命令模块的 Outbox 事件 helper。
package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cube-castle/pkg/database"
	"github.com/google/uuid"
)

const (
	// DefaultSourceCommand 是组织命令服务在事件中的 source 值。
	DefaultSourceCommand = "command-service"

	aggregateAssignment = "assignment"
	aggregatePosition   = "position"
	aggregateJobLevel   = "jobLevel"

	// EventAssignmentFilled 表示任命占用。
	EventAssignmentFilled = "assignment.filled"
	// EventAssignmentVacated 表示任命释放。
	EventAssignmentVacated = "assignment.vacated"
	// EventAssignmentUpdated 表示任命变更。
	EventAssignmentUpdated = "assignment.updated"
	// EventAssignmentClosed 表示任命关闭。
	EventAssignmentClosed = "assignment.closed"

	// EventPositionCreated 表示职位创建。
	EventPositionCreated = "position.created"
	// EventPositionUpdated 表示职位更新。
	EventPositionUpdated = "position.updated"

	// EventJobLevelVersionCreated 表示职级版本创建。
	EventJobLevelVersionCreated = "jobLevel.versionCreated"
	// EventJobLevelVersionConflict 表示职级版本冲突。
	EventJobLevelVersionConflict = "jobLevel.versionConflict"
)

// Context 描述 outbox 事件的通用上下文。
type Context struct {
	TenantID      uuid.UUID
	RequestID     string
	CorrelationID string
	Operation     string
	Source        string
}

// NewAssignmentEvent 构造 assignment.* 事件。
func NewAssignmentEvent(eventType string, ctx Context, assignmentID, positionCode string, payload map[string]interface{}) (*database.OutboxEvent, error) {
	aggregateID := strings.TrimSpace(assignmentID)
	if aggregateID == "" {
		aggregateID = strings.TrimSpace(positionCode)
	}
	if aggregateID == "" {
		aggregateID = ctx.TenantID.String()
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["positionCode"] = strings.TrimSpace(positionCode)
	if assignmentID != "" {
		payload["assignmentId"] = strings.TrimSpace(assignmentID)
	}
	return newOutboxEvent(eventType, aggregateAssignment, aggregateID, ctx, payload)
}

// NewPositionEvent 构造 position.* 事件。
func NewPositionEvent(eventType string, ctx Context, positionCode string, payload map[string]interface{}) (*database.OutboxEvent, error) {
	aggregateID := strings.TrimSpace(positionCode)
	if aggregateID == "" {
		aggregateID = ctx.TenantID.String()
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["positionCode"] = strings.TrimSpace(positionCode)
	return newOutboxEvent(eventType, aggregatePosition, aggregateID, ctx, payload)
}

// NewJobLevelEvent 构造 jobLevel.* 事件。
func NewJobLevelEvent(eventType string, ctx Context, jobLevelCode string, payload map[string]interface{}) (*database.OutboxEvent, error) {
	aggregateID := strings.TrimSpace(jobLevelCode)
	if aggregateID == "" {
		aggregateID = ctx.TenantID.String()
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["jobLevelCode"] = strings.TrimSpace(jobLevelCode)
	return newOutboxEvent(eventType, aggregateJobLevel, aggregateID, ctx, payload)
}

func newOutboxEvent(eventType, aggregateType, aggregateID string, ctx Context, attributes map[string]interface{}) (*database.OutboxEvent, error) {
	aggregateID = strings.TrimSpace(aggregateID)
	if aggregateID == "" {
		return nil, fmt.Errorf("aggregateID is required for event %s", eventType)
	}

	data := make(map[string]interface{}, len(attributes)+8)
	for k, v := range attributes {
		if k == "" || v == nil {
			continue
		}
		data[k] = v
	}

	if ctx.TenantID != uuid.Nil {
		data["tenantId"] = ctx.TenantID.String()
	}
	if rid := strings.TrimSpace(ctx.RequestID); rid != "" {
		data["requestId"] = rid
	}
	if cid := strings.TrimSpace(ctx.CorrelationID); cid != "" {
		data["correlationId"] = cid
	}
	if op := strings.TrimSpace(ctx.Operation); op != "" {
		data["operation"] = op
	}

	source := strings.TrimSpace(ctx.Source)
	if source == "" {
		source = DefaultSourceCommand
	}
	data["source"] = source
	data["aggregateType"] = aggregateType
	data["aggregateId"] = aggregateID
	data["eventType"] = eventType
	data["occurredAt"] = time.Now().UTC().Format(time.RFC3339Nano)

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}

	event := database.NewOutboxEvent()
	event.AggregateID = aggregateID
	event.AggregateType = aggregateType
	event.EventType = eventType
	event.Payload = string(raw)
	return event, nil
}
