package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// 简化的事件结构用于调试
type DebugEvent struct {
	Schema  interface{} `json:"schema"`
	Payload struct {
		Before *map[string]interface{} `json:"before"`
		After  *map[string]interface{} `json:"after"`
		Source struct {
			Connector string `json:"connector"`
			Name      string `json:"name"`
		} `json:"source"`
		Op   string `json:"op"`
		TsMs int64  `json:"ts_ms"`
	} `json:"payload"`
}

func main() {
	log.Println("🔍 启动CDC消息诊断工具...")
	
	// Kafka消费者配置
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	
	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatalf("创建Kafka消费者失败: %v", err)
	}
	defer consumer.Close()
	
	// 消费最新消息
	partitionConsumer, err := consumer.ConsumePartition("organization_db.public.organization_units", 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("创建分区消费者失败: %v", err)
	}
	defer partitionConsumer.Close()
	
	log.Println("⏳ 等待新的CDC消息...")
	
	timeout := time.After(30 * time.Second)
	for {
		select {
		case message := <-partitionConsumer.Messages():
			log.Printf("📬 收到消息: offset=%d, 大小=%d bytes", message.Offset, len(message.Value))
			
			// 尝试解析消息
			var event DebugEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("❌ JSON解析失败: %v", err)
				log.Printf("📄 原始消息前200字符: %s", string(message.Value[:min(200, len(message.Value))]))
			} else {
				log.Printf("✅ 消息解析成功!")
				log.Printf("   Schema存在: %v", event.Schema != nil)
				log.Printf("   操作类型: '%s'", event.Payload.Op)
				log.Printf("   连接器: %s", event.Payload.Source.Connector)
				log.Printf("   时间戳: %d", event.Payload.TsMs)
				
				if event.Payload.After != nil {
					afterData, _ := json.Marshal(event.Payload.After)
					log.Printf("   After数据: %s", string(afterData))
				}
			}
			return
			
		case <-timeout:
			log.Println("⏰ 30秒内没有收到新消息")
			return
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}