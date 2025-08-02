package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/google/uuid"
)

// 简单的测试程序验证EventBus功能
func main() {
	log.Println("🧪 Testing EventBus integration...")

	// 创建Mock EventBus用于测试
	factory := events.NewEventBusFactory()
	mockEventBus := factory.CreateMockEventBus()

	// 测试发布员工创建事件
	tenantID := uuid.New()
	employeeID := uuid.New()
	
	event := events.NewEmployeeCreated(
		tenantID, 
		employeeID, 
		"EMP001",
		"张", 
		"三", 
		"zhangsan@example.com", 
		time.Now(),
	)

	log.Printf("📤 Publishing EmployeeCreated event: %s", event.GetEventID())
	
	ctx := context.Background()
	err := mockEventBus.Publish(ctx, event)
	if err != nil {
		log.Printf("❌ Failed to publish event: %v", err)
		return
	}

	log.Println("✅ Event published successfully!")

	// 验证事件内容
	eventData, err := event.Serialize()
	if err != nil {
		log.Printf("❌ Failed to serialize event: %v", err)
		return
	}

	log.Printf("📊 Event data: %s", string(eventData))

	// 测试事件头部信息
	headers := event.GetHeaders()
	headersData, _ := json.MarshalIndent(headers, "", "  ")
	log.Printf("📋 Event headers:\n%s", string(headersData))

	// 测试Mock EventBus的存储功能
	if mockBus, ok := mockEventBus.(*events.MockEventBus); ok {
		publishedEvents := mockBus.GetPublishedEvents()
		log.Printf("📈 Total events published: %d", len(publishedEvents))
		
		if len(publishedEvents) > 0 {
			lastEvent := publishedEvents[len(publishedEvents)-1]
			log.Printf("🔍 Last event type: %s", lastEvent.GetEventType())
			log.Printf("🔍 Last event aggregate: %s", lastEvent.GetAggregateType())
		}
	}

	log.Println("🎉 EventBus integration test completed successfully!")
}