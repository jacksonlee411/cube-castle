package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// EventBusService 事件总线服务
type EventBusService struct {
	eventBus   EventBus
	serializer EventSerializer
	validator  *EventValidator
	config     *EventBusConfig
}

// EventBusServiceInterface 事件总线服务接口
type EventBusServiceInterface interface {
	GetEventBus() EventBus
	GetSerializer() EventSerializer
	GetValidator() *EventValidator
	Start(ctx context.Context) error
	Stop() error
	Health() error
}

// NewEventBusService 创建事件总线服务
func NewEventBusService(config *EventBusConfig) (*EventBusService, error) {
	if config == nil {
		config = DefaultEventBusConfig()
	}

	// 创建Kafka事件总线
	eventBus, err := NewKafkaEventBus(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka event bus: %w", err)
	}

	// 创建序列化器
	serializerFactory := NewEventSerializerFactory()
	serializer := serializerFactory.CreateJSONSerializer()

	// 创建验证器
	validator := NewEventValidator()

	return &EventBusService{
		eventBus:   eventBus,
		serializer: serializer,
		validator:  validator,
		config:     config,
	}, nil
}

// GetEventBus 获取事件总线实例
func (s *EventBusService) GetEventBus() EventBus {
	return s.eventBus
}

// GetSerializer 获取序列化器
func (s *EventBusService) GetSerializer() EventSerializer {
	return s.serializer
}

// GetValidator 获取验证器
func (s *EventBusService) GetValidator() *EventValidator {
	return s.validator
}

// Start 启动事件总线服务
func (s *EventBusService) Start(ctx context.Context) error {
	log.Println("Starting EventBus service...")
	return s.eventBus.Start(ctx)
}

// Stop 停止事件总线服务
func (s *EventBusService) Stop() error {
	log.Println("Stopping EventBus service...")
	return s.eventBus.Stop()
}

// Health 健康检查
func (s *EventBusService) Health() error {
	return s.eventBus.Health()
}

// ConfigFromEnv 从环境变量创建配置
func ConfigFromEnv() *EventBusConfig {
	config := &EventBusConfig{
		KafkaBootstrapServers: getEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaTopicPrefix:      getEnvOrDefault("KAFKA_TOPIC_PREFIX", "cube_castle"),
		KafkaConsumerGroup:    getEnvOrDefault("KAFKA_CONSUMER_GROUP", "cube_castle_consumers"),
		BatchSize:             getEnvAsIntOrDefault("KAFKA_BATCH_SIZE", 100),
		BatchTimeout:          getEnvAsDurationOrDefault("KAFKA_BATCH_TIMEOUT", time.Millisecond*100),
		MaxRetries:            getEnvAsIntOrDefault("KAFKA_MAX_RETRIES", 3),
		RetryBackoff:          getEnvAsDurationOrDefault("KAFKA_RETRY_BACKOFF", time.Second*2),
		EnableMetrics:         getEnvAsBoolOrDefault("KAFKA_ENABLE_METRICS", true),
		MetricsPrefix:         getEnvOrDefault("KAFKA_METRICS_PREFIX", "cube_castle_eventbus"),
		EnableTLS:             getEnvAsBoolOrDefault("KAFKA_ENABLE_TLS", false),
	}

	// TLS配置
	if config.EnableTLS {
		config.TLSConfig = &TLSConfig{
			CertFile:   getEnvOrDefault("KAFKA_TLS_CERT_FILE", ""),
			KeyFile:    getEnvOrDefault("KAFKA_TLS_KEY_FILE", ""),
			CAFile:     getEnvOrDefault("KAFKA_TLS_CA_FILE", ""),
			SkipVerify: getEnvAsBoolOrDefault("KAFKA_TLS_SKIP_VERIFY", false),
		}
	}

	return config
}

// 环境变量辅助函数

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// EventBusManager 事件总线管理器
type EventBusManager struct {
	services map[string]EventBusServiceInterface
}

// NewEventBusManager 创建事件总线管理器
func NewEventBusManager() *EventBusManager {
	return &EventBusManager{
		services: make(map[string]EventBusServiceInterface),
	}
}

// RegisterService 注册事件总线服务
func (m *EventBusManager) RegisterService(name string, service EventBusServiceInterface) {
	m.services[name] = service
}

// GetService 获取事件总线服务
func (m *EventBusManager) GetService(name string) (EventBusServiceInterface, bool) {
	service, exists := m.services[name]
	return service, exists
}

// StartAll 启动所有服务
func (m *EventBusManager) StartAll(ctx context.Context) error {
	for name, service := range m.services {
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("failed to start service %s: %w", name, err)
		}
		log.Printf("EventBus service %s started successfully", name)
	}
	return nil
}

// StopAll 停止所有服务
func (m *EventBusManager) StopAll() error {
	for name, service := range m.services {
		if err := service.Stop(); err != nil {
			log.Printf("Error stopping service %s: %v", name, err)
		} else {
			log.Printf("EventBus service %s stopped successfully", name)
		}
	}
	return nil
}

// HealthCheckAll 健康检查所有服务
func (m *EventBusManager) HealthCheckAll() map[string]error {
	results := make(map[string]error)
	for name, service := range m.services {
		results[name] = service.Health()
	}
	return results
}