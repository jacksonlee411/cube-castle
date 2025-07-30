package outbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOutboxService_CreateEvent(t *testing.T) {
	// 跳过实际数据库测试
	t.Skip("Skipping database tests in unit test mode")
}

func TestOutboxService_CreateEmployeeCreatedEvent(t *testing.T) {
	t.Skip("Skipping database dependent test")
}

func TestOutboxService_CreateEmployeeUpdatedEvent(t *testing.T) {
	t.Skip("Skipping database dependent test")
}

func TestOutboxService_CreateOrganizationCreatedEvent(t *testing.T) {
	t.Skip("Skipping database dependent test")
}

func TestOutboxService_CreateLeaveRequestCreatedEvent(t *testing.T) {
	t.Skip("Skipping database dependent test")
}

func TestOutboxService_CreateNotificationEvent(t *testing.T) {
	t.Skip("Skipping database dependent test")
}

func TestEventPayloadMarshaling(t *testing.T) {
	// 测试事件载荷的JSON序列化
	employeeData := map[string]interface{}{
		"employee_id": "123e4567-e89b-12d3-a456-426614174000",
		"employee_data": map[string]interface{}{
			"first_name": "张三",
			"last_name":  "李",
			"email":      "zhangsan@example.com",
		},
		"created_at": time.Now().Format(time.RFC3339),
	}

	payload, err := json.Marshal(employeeData)
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}

	// 验证JSON可以正确反序列化
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(payload, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	assert.Equal(t, employeeData["employee_id"], unmarshaled["employee_id"])
	assert.NotNil(t, unmarshaled["employee_data"])
	assert.NotNil(t, unmarshaled["created_at"])
}

func TestEventTypes(t *testing.T) {
	// 测试事件类型常量
	assert.NotEmpty(t, EventTypeEmployeeCreated)
	assert.NotEmpty(t, EventTypeEmployeeUpdated)
	assert.NotEmpty(t, EventTypeEmployeePhoneUpdated)
	assert.NotEmpty(t, EventTypeOrganizationCreated)
	assert.NotEmpty(t, EventTypeLeaveRequestCreated)
	assert.NotEmpty(t, EventTypeLeaveRequestApproved)
	assert.NotEmpty(t, EventTypeLeaveRequestRejected)
	assert.NotEmpty(t, EventTypeNotification)
}

func TestAggregateTypes(t *testing.T) {
	// 测试聚合类型常量
	assert.NotEmpty(t, AggregateTypeEmployee)
	assert.NotEmpty(t, AggregateTypeOrganization)
	assert.NotEmpty(t, AggregateTypeLeaveRequest)
	assert.NotEmpty(t, AggregateTypeNotification)
}

func TestCreateEventRequest(t *testing.T) {
	// 测试创建事件请求结构
	req := &CreateEventRequest{
		AggregateID:   uuid.New(),
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeeCreated,
		EventVersion:  1,
		Payload:       []byte(`{"test": "data"}`),
	}

	assert.NotNil(t, req.AggregateID)
	assert.Equal(t, AggregateTypeEmployee, req.AggregateType)
	assert.Equal(t, EventTypeEmployeeCreated, req.EventType)
	assert.Equal(t, 1, req.EventVersion)
	assert.NotEmpty(t, req.Payload)
}

func TestEventStructure(t *testing.T) {
	// 测试事件结构
	event := &Event{
		ID:            uuid.New(),
		AggregateID:   uuid.New(),
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeeCreated,
		EventVersion:  1,
		Payload:       []byte(`{"test": "data"}`),
		Metadata:      nil,
		ProcessedAt:   nil,
		CreatedAt:     time.Now(),
	}

	assert.NotNil(t, event.ID)
	assert.NotNil(t, event.AggregateID)
	assert.Equal(t, AggregateTypeEmployee, event.AggregateType)
	assert.Equal(t, EventTypeEmployeeCreated, event.EventType)
	assert.Equal(t, 1, event.EventVersion)
	assert.NotEmpty(t, event.Payload)
	assert.NotNil(t, event.CreatedAt)
}

// 集成测试（需要实际数据库）
func TestOutboxServiceIntegration(t *testing.T) {
	t.Skip("Integration test - requires actual database")

	// 这个测试需要实际的数据库连接
	// 在实际环境中，应该使用测试数据库进行集成测试
	// ctx := context.Background()

	// 1. 创建数据库连接
	// 2. 初始化发件箱服务
	// 3. 创建事件
	// 4. 验证事件被正确存储
	// 5. 处理事件
	// 6. 验证事件被标记为已处理
	// 7. 测试事件重放
}
