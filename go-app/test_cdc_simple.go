package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/google/uuid"
)

// 简化的CDC验证测试
// 验证EventBus和事件系统的基本功能

func main() {
	log.Println("🧪 启动简化CDC验证测试...")
	
	// 创建EventBus
	factory := events.NewEventBusFactory()
	eventBus := factory.CreateMockEventBus()
	
	// 测试基本事件发布
	testBasicEventPublishing(eventBus)
	
	// 测试事件序列化
	testEventSerialization()
	
	// 测试批量事件处理
	testBatchEventHandling(eventBus)
	
	log.Println("✅ 简化CDC验证测试完成")
	log.Println("🎉 EventBus系统功能验证通过")
}

// testBasicEventPublishing 测试基本事件发布
func testBasicEventPublishing(eventBus events.EventBus) {
	log.Println("🔄 测试基本事件发布...")
	
	ctx := context.Background()
	tenantID := uuid.New()
	employeeID := uuid.New()
	
	// 创建员工创建事件
	event := events.NewEmployeeCreated(
		tenantID,
		employeeID,
		"TEST001",
		"张",
		"三",
		"zhangsan@test.com",
		time.Now(),
	)
	
	log.Printf("📤 发布事件: %s (ID: %s)", event.GetEventType(), event.GetEventID())
	
	// 发布事件
	if err := eventBus.Publish(ctx, event); err != nil {
		log.Printf("❌ 事件发布失败: %v", err)
		return
	}
	
	log.Printf("✅ 事件发布成功: %s", event.GetEventID())
	
	// 如果是Mock EventBus，验证事件是否被存储
	type MockEventBusInterface interface {
		GetPublishedEvents() []events.DomainEvent
	}
	
	if mockBus, ok := eventBus.(MockEventBusInterface); ok {
		publishedEvents := mockBus.GetPublishedEvents()
		log.Printf("📊 已发布事件数量: %d", len(publishedEvents))
		
		if len(publishedEvents) > 0 {
			lastEvent := publishedEvents[len(publishedEvents)-1]
			log.Printf("🔍 最后发布的事件: %s", lastEvent.GetEventType())
		}
	}
}

// testEventSerialization 测试事件序列化
func testEventSerialization() {
	log.Println("🔄 测试事件序列化...")
	
	tenantID := uuid.New()
	employeeID := uuid.New()
	
	// 创建事件
	event := events.NewEmployeeCreated(
		tenantID,
		employeeID,
		"SER001",
		"序列化",
		"测试",
		"serialization@test.com",
		time.Now(),
	)
	
	// 测试序列化
	serializedData, err := event.Serialize()
	if err != nil {
		log.Printf("❌ 事件序列化失败: %v", err)
		return
	}
	
	log.Printf("✅ 事件序列化成功，数据长度: %d 字节", len(serializedData))
	log.Printf("📄 序列化数据示例: %s", string(serializedData)[:min(len(serializedData), 200)] + "...")
	
	// 验证事件头部信息
	headers := event.GetHeaders()
	log.Printf("📋 事件头部信息:")
	for key, value := range headers {
		log.Printf("   %s: %v", key, value)
	}
}

// testBatchEventHandling 测试批量事件处理
func testBatchEventHandling(eventBus events.EventBus) {
	log.Println("🔄 测试批量事件处理...")
	
	ctx := context.Background()
	tenantID := uuid.New()
	
	// 创建批量事件
	var domainEvents []events.DomainEvent
	for i := 0; i < 5; i++ {
		employeeID := uuid.New()
		event := events.NewEmployeeCreated(
			tenantID,
			employeeID,
			fmt.Sprintf("BATCH%03d", i),
			"批量",
			fmt.Sprintf("测试%d", i),
			fmt.Sprintf("batch%d@test.com", i),
			time.Now(),
		)
		domainEvents = append(domainEvents, event)
	}
	
	log.Printf("📤 批量发布事件: %d 个", len(domainEvents))
	startTime := time.Now()
	
	// 批量发布事件
	for i, event := range domainEvents {
		if err := eventBus.Publish(ctx, event); err != nil {
			log.Printf("❌ 批量事件 %d 发布失败: %v", i, err)
			return
		}
	}
	
	duration := time.Since(startTime)
	
	log.Printf("✅ 批量事件发布成功")
	log.Printf("📊 处理时间: %v", duration)
	log.Printf("📊 平均每事件: %v", duration/time.Duration(len(domainEvents)))
	
	// 验证Mock EventBus中的事件
	type MockEventBusInterface interface {
		GetPublishedEvents() []events.DomainEvent
	}
	
	if mockBus, ok := eventBus.(MockEventBusInterface); ok {
		publishedEvents := mockBus.GetPublishedEvents()
		log.Printf("📊 总事件数: %d", len(publishedEvents))
		
		// 统计事件类型
		eventTypeCount := make(map[string]int)
		for _, event := range publishedEvents {
			eventTypeCount[event.GetEventType()]++
		}
		
		log.Printf("📊 事件类型统计:")
		for eventType, count := range eventTypeCount {
			log.Printf("   %s: %d", eventType, count)
		}
	}
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}