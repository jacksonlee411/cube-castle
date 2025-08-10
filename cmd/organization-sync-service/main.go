package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ===== Debezium日期字段处理 =====

// DebeziumDate 处理Debezium序列化的date字段，可能是数字或字符串
type DebeziumDate struct {
	value string
}

// UnmarshalJSON 处理Debezium的日期序列化格式
func (d *DebeziumDate) UnmarshalJSON(data []byte) error {
	// 处理null值
	if string(data) == "null" {
		d.value = ""
		return nil
	}
	
	// 尝试解析为数字（Debezium days since epoch）
	if len(data) > 0 && data[0] != '"' {
		var days int64
		if err := json.Unmarshal(data, &days); err == nil {
			// 转换为YYYY-MM-DD格式
			epochDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			targetDate := epochDate.AddDate(0, 0, int(days))
			d.value = targetDate.Format("2006-01-02")
			return nil
		}
	}
	
	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		d.value = str
		return nil
	}
	
	return fmt.Errorf("cannot unmarshal date field: %s", string(data))
}

// String 返回日期字符串
func (d *DebeziumDate) String() string {
	return d.value
}

// parseDebeziumDate 解析Debezium日期字段
func parseDebeziumDate(rawData json.RawMessage) (string, error) {
	if string(rawData) == "null" {
		return "", nil
	}
	
	// 尝试解析为数字（Debezium days since epoch）
	if len(rawData) > 0 && rawData[0] != '"' {
		var days int64
		if err := json.Unmarshal(rawData, &days); err == nil {
			// 转换为YYYY-MM-DD格式
			epochDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			targetDate := epochDate.AddDate(0, 0, int(days))
			return targetDate.Format("2006-01-02"), nil
		}
	}
	
	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(rawData, &str); err == nil {
		return str, nil
	}
	
	return "", fmt.Errorf("无法解析日期字段: %s", string(rawData))
}

// 项目默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 领域事件模型 =====

type OrganizationCreatedEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	AggregateID string    `json:"aggregate_id"` // 组织代码
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	UnitType    string    `json:"unit_type"`
	ParentCode  *string   `json:"parent_code,omitempty"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrganizationUpdatedEvent struct {
	EventID     uuid.UUID              `json:"event_id"`
	AggregateID string                 `json:"aggregate_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Changes     map[string]interface{} `json:"changes"`
	UpdatedBy   uuid.UUID              `json:"updated_by"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type OrganizationDeletedEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	AggregateID string    `json:"aggregate_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	DeletedBy   uuid.UUID `json:"deleted_by"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// ===== CDC事件模型 =====

type CDCOrganizationEvent struct {
	Before *CDCOrganizationData `json:"before"`
	After  *CDCOrganizationData `json:"after"`
	Source CDCSource            `json:"source"`
	Op     string               `json:"op"` // c, u, d, r
	TsMs   int64                `json:"ts_ms"`
}

type CDCOrganizationData struct {
	ID           *string         `json:"id"`
	TenantID     *string         `json:"tenant_id"`
	Code         *string         `json:"code"`
	ParentCode   *string         `json:"parent_code"`
	Name         *string         `json:"name"`
	UnitType     *string         `json:"unit_type"`
	Status       *string         `json:"status"`
	Level        *int            `json:"level"`
	Path         *string         `json:"path"`
	SortOrder    *int            `json:"sort_order"`
	Description  *string         `json:"description"`
	CreatedAt    *time.Time      `json:"created_at"`
	UpdatedAt    *time.Time      `json:"updated_at"`
	// 时态管理字段 - 使用json.RawMessage进行调试
	EffectiveDate *json.RawMessage `json:"effective_date"`
	EndDate       *json.RawMessage `json:"end_date"`
	IsTemporal    *bool           `json:"is_temporal"`
	Version       *int            `json:"version"`
	ChangeReason  *string         `json:"change_reason"`
	IsCurrent     *bool           `json:"is_current"`
}

type CDCSource struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	DB        string `json:"db"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	TxID      int64  `json:"txId"`
	LSN       int64  `json:"lsn"`
}

// ===== Neo4j同步服务 =====

type Neo4jSyncService struct {
	driver  neo4j.DriverWithContext
	logger  *log.Logger
	session neo4j.SessionWithContext
}

func NewNeo4jSyncService(uri, username, password string, logger *log.Logger) (*Neo4jSyncService, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("创建Neo4j驱动失败: %w", err)
	}

	// 验证连接
	ctx := context.Background()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, fmt.Errorf("Neo4j连接验证失败: %w", err)
	}

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})

	return &Neo4jSyncService{
		driver:  driver,
		logger:  logger,
		session: session,
	}, nil
}

func (s *Neo4jSyncService) Close() error {
	ctx := context.Background()
	if s.session != nil {
		s.session.Close(ctx)
	}
	if s.driver != nil {
		return s.driver.Close(ctx)
	}
	return nil
}

// ===== 领域事件处理 =====

func (s *Neo4jSyncService) HandleOrganizationCreated(ctx context.Context, event OrganizationCreatedEvent) error {
	s.logger.Printf("处理组织创建事件: %s - %s", event.AggregateID, event.Name)

	query := `
		MERGE (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET org.name = $name,
			org.unit_type = $unit_type,
			org.status = 'ACTIVE',
			org.level = CASE WHEN $parent_code IS NULL THEN 1 ELSE 2 END,
			org.path = CASE WHEN $parent_code IS NULL THEN '/' + $code ELSE '/' + $parent_code + '/' + $code END,
			org.sort_order = 0,
			org.description = COALESCE($description, ''),
			org.created_at = datetime($created_at),
			org.updated_at = datetime($created_at)
		WITH org
		OPTIONAL MATCH (parent:OrganizationUnit {code: $parent_code, tenant_id: $tenant_id})
		WHERE $parent_code IS NOT NULL
		FOREACH (p IN CASE WHEN parent IS NOT NULL THEN [parent] ELSE [] END |
			MERGE (p)-[:HAS_CHILD]->(org)
		)
		RETURN org.code as code`

	description := ""
	parentCode := ""
	if event.ParentCode != nil {
		parentCode = *event.ParentCode
	}

	_, err := s.session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"code":        event.AggregateID,
			"tenant_id":   event.TenantID.String(),
			"name":        event.Name,
			"unit_type":   event.UnitType,
			"parent_code": parentCode,
			"description": description,
			"created_at":  event.CreatedAt.Format(time.RFC3339),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			code, _ := result.Record().Get("code")
			return code, nil
		}
		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("Neo4j组织创建失败: %w", err)
	}

	s.logger.Printf("✅ Neo4j组织创建成功: %s", event.AggregateID)
	return nil
}

func (s *Neo4jSyncService) HandleOrganizationUpdated(ctx context.Context, event OrganizationUpdatedEvent) error {
	s.logger.Printf("处理组织更新事件: %s", event.AggregateID)

	// 构建动态更新查询
	setParts := []string{}
	params := map[string]interface{}{
		"code":       event.AggregateID,
		"tenant_id":  event.TenantID.String(),
		"updated_at": event.UpdatedAt.Format(time.RFC3339),
	}

	for field, value := range event.Changes {
		switch field {
		case "name":
			setParts = append(setParts, "org.name = $name")
			params["name"] = value
		case "status":
			setParts = append(setParts, "org.status = $status")
			params["status"] = value
		case "description":
			setParts = append(setParts, "org.description = $description")
			params["description"] = value
		case "sort_order":
			setParts = append(setParts, "org.sort_order = $sort_order")
			params["sort_order"] = value
		}
	}

	if len(setParts) == 0 {
		s.logger.Printf("⚠️ 没有需要更新的字段: %s", event.AggregateID)
		return nil
	}

	query := fmt.Sprintf(`
		MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET %s, org.updated_at = datetime($updated_at)
		RETURN org.code as code`, strings.Join(setParts, ", "))

	_, err := s.session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			code, _ := result.Record().Get("code")
			return code, nil
		}
		return nil, fmt.Errorf("组织不存在: %s", event.AggregateID)
	})

	if err != nil {
		return fmt.Errorf("Neo4j组织更新失败: %w", err)
	}

	s.logger.Printf("✅ Neo4j组织更新成功: %s", event.AggregateID)
	return nil
}

func (s *Neo4jSyncService) HandleOrganizationDeleted(ctx context.Context, event OrganizationDeletedEvent) error {
	s.logger.Printf("处理组织删除事件: %s", event.AggregateID)

	query := `
		MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET org.status = 'INACTIVE',
			org.updated_at = datetime($deleted_at)
		RETURN org.code as code`

	_, err := s.session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"code":       event.AggregateID,
			"tenant_id":  event.TenantID.String(),
			"deleted_at": event.DeletedAt.Format(time.RFC3339),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			code, _ := result.Record().Get("code")
			return code, nil
		}
		return nil, fmt.Errorf("组织不存在: %s", event.AggregateID)
	})

	if err != nil {
		return fmt.Errorf("Neo4j组织删除失败: %w", err)
	}

	s.logger.Printf("✅ Neo4j组织删除成功: %s", event.AggregateID)
	return nil
}

// ===== CDC事件处理 =====

func (s *Neo4jSyncService) HandleCDCEvent(ctx context.Context, event CDCOrganizationEvent) error {
	switch event.Op {
	case "c": // CREATE
		if event.After == nil {
			return fmt.Errorf("CDC CREATE事件缺少after数据")
		}
		return s.handleCDCCreate(ctx, event.After, event.TsMs)
	case "u": // UPDATE
		if event.After == nil {
			return fmt.Errorf("CDC UPDATE事件缺少after数据")
		}
		return s.handleCDCUpdate(ctx, event.After, event.TsMs)
	case "d": // DELETE
		if event.Before == nil {
			return fmt.Errorf("CDC DELETE事件缺少before数据")
		}
		return s.handleCDCDelete(ctx, event.Before, event.TsMs)
	case "r": // READ (snapshot)
		if event.After == nil {
			return fmt.Errorf("CDC READ事件缺少after数据")
		}
		return s.handleCDCCreate(ctx, event.After, event.TsMs)
	default:
		s.logger.Printf("⚠️ 未知的CDC操作类型: %s", event.Op)
		return nil
	}
}

func (s *Neo4jSyncService) handleCDCCreate(ctx context.Context, data *CDCOrganizationData, tsMs int64) error {
	if data.Code == nil || data.TenantID == nil || data.Name == nil {
		return fmt.Errorf("CDC CREATE事件缺少必要字段")
	}

	s.logger.Printf("处理CDC创建事件: %s - %s", *data.Code, *data.Name)

	query := `
		MERGE (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET org.name = $name,
			org.unit_type = $unit_type,
			org.status = COALESCE($status, 'ACTIVE'),
			org.level = COALESCE($level, 1),
			org.path = COALESCE($path, '/' + $code),
			org.sort_order = COALESCE($sort_order, 0),
			org.description = COALESCE($description, ''),
			org.created_at = datetime($created_at),
			org.updated_at = datetime($updated_at),
			org.effective_date = $effective_date,
			org.end_date = $end_date,
			org.is_temporal = COALESCE($is_temporal, false),
			org.version = COALESCE($version, 1),
			org.change_reason = COALESCE($change_reason, ''),
			org.is_current = COALESCE($is_current, true)
		WITH org
		OPTIONAL MATCH (parent:OrganizationUnit {code: $parent_code, tenant_id: $tenant_id})
		WHERE $parent_code IS NOT NULL AND $parent_code <> ''
		FOREACH (p IN CASE WHEN parent IS NOT NULL THEN [parent] ELSE [] END |
			MERGE (p)-[:HAS_CHILD]->(org)
		)
		RETURN org.code as code`

	params := map[string]interface{}{
		"code":      *data.Code,
		"tenant_id": *data.TenantID,
		"name":      *data.Name,
	}

	// 安全处理可选字段
	if data.UnitType != nil {
		params["unit_type"] = *data.UnitType
	} else {
		params["unit_type"] = "DEPARTMENT"
	}

	if data.Status != nil {
		params["status"] = *data.Status
	} else {
		params["status"] = "ACTIVE"
	}

	if data.Level != nil {
		params["level"] = *data.Level
	} else {
		params["level"] = 1
	}

	if data.Path != nil {
		params["path"] = *data.Path
	} else {
		params["path"] = "/" + *data.Code
	}

	if data.SortOrder != nil {
		params["sort_order"] = *data.SortOrder
	} else {
		params["sort_order"] = 0
	}

	if data.Description != nil {
		params["description"] = *data.Description
	} else {
		params["description"] = ""
	}

	if data.CreatedAt != nil {
		params["created_at"] = data.CreatedAt.Format(time.RFC3339)
	} else {
		params["created_at"] = time.Now().Format(time.RFC3339)
	}

	if data.UpdatedAt != nil {
		params["updated_at"] = data.UpdatedAt.Format(time.RFC3339)
	} else {
		params["updated_at"] = time.Now().Format(time.RFC3339)
	}

	if data.ParentCode != nil && *data.ParentCode != "" {
		params["parent_code"] = *data.ParentCode
	} else {
		params["parent_code"] = nil
	}

	// 时态管理字段映射 - 处理json.RawMessage
	if data.EffectiveDate != nil {
		dateStr, err := parseDebeziumDate(*data.EffectiveDate)
		if err != nil {
			log.Printf("解析effective_date失败: %v, 原始数据: %s", err, string(*data.EffectiveDate))
			params["effective_date"] = nil
		} else {
			params["effective_date"] = dateStr
		}
	} else {
		params["effective_date"] = nil
	}

	if data.EndDate != nil {
		dateStr, err := parseDebeziumDate(*data.EndDate)
		if err != nil {
			log.Printf("解析end_date失败: %v, 原始数据: %s", err, string(*data.EndDate))
			params["end_date"] = nil
		} else {
			params["end_date"] = dateStr
		}
	} else {
		params["end_date"] = nil
	}

	if data.IsTemporal != nil {
		params["is_temporal"] = *data.IsTemporal
	} else {
		params["is_temporal"] = false
	}

	if data.Version != nil {
		params["version"] = *data.Version
	} else {
		params["version"] = 1
	}

	if data.ChangeReason != nil {
		params["change_reason"] = *data.ChangeReason
	} else {
		params["change_reason"] = ""
	}

	if data.IsCurrent != nil {
		params["is_current"] = *data.IsCurrent
	} else {
		params["is_current"] = true
	}

	_, err := s.session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			code, _ := result.Record().Get("code")
			return code, nil
		}
		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("Neo4j CDC创建失败: %w", err)
	}

	s.logger.Printf("✅ Neo4j CDC创建成功: %s", *data.Code)
	return nil
}

func (s *Neo4jSyncService) handleCDCUpdate(ctx context.Context, data *CDCOrganizationData, tsMs int64) error {
	if data.Code == nil || data.TenantID == nil {
		return fmt.Errorf("CDC UPDATE事件缺少必要字段")
	}

	s.logger.Printf("处理CDC更新事件: %s", *data.Code)

	query := `
		MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET org.name = COALESCE($name, org.name),
			org.unit_type = COALESCE($unit_type, org.unit_type),
			org.status = COALESCE($status, org.status),
			org.level = COALESCE($level, org.level),
			org.path = COALESCE($path, org.path),
			org.sort_order = COALESCE($sort_order, org.sort_order),
			org.description = COALESCE($description, org.description),
			org.updated_at = datetime($updated_at),
			org.effective_date = CASE WHEN $effective_date IS NULL THEN org.effective_date ELSE datetime($effective_date) END,
			org.end_date = CASE WHEN $end_date IS NULL THEN org.end_date ELSE datetime($end_date) END,
			org.is_temporal = COALESCE($is_temporal, org.is_temporal),
			org.version = COALESCE($version, org.version),
			org.change_reason = COALESCE($change_reason, org.change_reason),
			org.is_current = COALESCE($is_current, org.is_current)
		RETURN org.code as code`

	params := map[string]interface{}{
		"code":      *data.Code,
		"tenant_id": *data.TenantID,
	}

	if data.Name != nil {
		params["name"] = *data.Name
	}
	if data.UnitType != nil {
		params["unit_type"] = *data.UnitType
	}
	if data.Status != nil {
		params["status"] = *data.Status
	}
	if data.Level != nil {
		params["level"] = *data.Level
	}
	if data.Path != nil {
		params["path"] = *data.Path
	}
	if data.SortOrder != nil {
		params["sort_order"] = *data.SortOrder
	}
	if data.Description != nil {
		params["description"] = *data.Description
	}
	if data.UpdatedAt != nil {
		params["updated_at"] = data.UpdatedAt.Format(time.RFC3339)
	} else {
		params["updated_at"] = time.Now().Format(time.RFC3339)
	}

	// 时态管理字段映射 (更新版本) - 处理json.RawMessage
	if data.EffectiveDate != nil {
		dateStr, err := parseDebeziumDate(*data.EffectiveDate)
		if err != nil {
			log.Printf("解析effective_date失败: %v, 原始数据: %s", err, string(*data.EffectiveDate))
			params["effective_date"] = nil
		} else {
			params["effective_date"] = dateStr
		}
	} else {
		params["effective_date"] = nil
	}

	if data.EndDate != nil {
		dateStr, err := parseDebeziumDate(*data.EndDate)
		if err != nil {
			log.Printf("解析end_date失败: %v, 原始数据: %s", err, string(*data.EndDate))
			params["end_date"] = nil
		} else {
			params["end_date"] = dateStr
		}
	} else {
		params["end_date"] = nil
	}

	if data.IsTemporal != nil {
		params["is_temporal"] = *data.IsTemporal
	} else {
		params["is_temporal"] = nil
	}

	if data.Version != nil {
		params["version"] = *data.Version
	} else {
		params["version"] = nil
	}

	if data.ChangeReason != nil {
		params["change_reason"] = *data.ChangeReason
	} else {
		params["change_reason"] = nil
	}

	if data.IsCurrent != nil {
		params["is_current"] = *data.IsCurrent
	} else {
		params["is_current"] = nil
	}

	_, err := s.session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			code, _ := result.Record().Get("code")
			return code, nil
		}
		return nil, fmt.Errorf("组织不存在: %s", *data.Code)
	})

	if err != nil {
		return fmt.Errorf("Neo4j CDC更新失败: %w", err)
	}

	s.logger.Printf("✅ Neo4j CDC更新成功: %s", *data.Code)
	return nil
}

func (s *Neo4jSyncService) handleCDCDelete(ctx context.Context, data *CDCOrganizationData, tsMs int64) error {
	if data.Code == nil || data.TenantID == nil {
		return fmt.Errorf("CDC DELETE事件缺少必要字段")
	}

	s.logger.Printf("处理CDC删除事件: %s", *data.Code)

	query := `
		MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET org.status = 'INACTIVE',
			org.updated_at = datetime($deleted_at)
		RETURN org.code as code`

	_, err := s.session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"code":       *data.Code,
			"tenant_id":  *data.TenantID,
			"deleted_at": time.Now().Format(time.RFC3339),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			code, _ := result.Record().Get("code")
			return code, nil
		}
		return nil, fmt.Errorf("组织不存在: %s", *data.Code)
	})

	if err != nil {
		return fmt.Errorf("Neo4j CDC删除失败: %w", err)
	}

	s.logger.Printf("✅ Neo4j CDC删除成功: %s", *data.Code)
	return nil
}

// ===== Kafka消费者 =====

type KafkaEventConsumer struct {
	consumer *kafka.Consumer
	syncSvc  *Neo4jSyncService
	logger   *log.Logger
}

func NewKafkaEventConsumer(brokers []string, groupID string, syncSvc *Neo4jSyncService, logger *log.Logger) (*KafkaEventConsumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"group.id":          groupID,
		"auto.offset.reset": "latest",  // 从最新位置开始
		"enable.auto.commit": true,
		"auto.commit.interval.ms": 1000,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	return &KafkaEventConsumer{
		consumer: consumer,
		syncSvc:  syncSvc,
		logger:   logger,
	}, nil
}

func (c *KafkaEventConsumer) Subscribe(topics []string) error {
	return c.consumer.SubscribeTopics(topics, nil)
}

func (c *KafkaEventConsumer) StartConsuming(ctx context.Context) error {
	c.logger.Println("🚀 开始消费Kafka事件...")

	for {
		select {
		case <-ctx.Done():
			c.logger.Println("收到停止信号，停止消费...")
			return nil
		default:
			msg, err := c.consumer.ReadMessage(1000)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				c.logger.Printf("消费消息失败: %v", err)
				continue
			}

			c.logger.Printf("收到消息: topic=%s, partition=%d, offset=%d",
				*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)

			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Printf("处理消息失败: %v", err)
			}
		}
	}
}

func (c *KafkaEventConsumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	topic := *msg.TopicPartition.Topic

	switch topic {
	case "organization.events":
		return c.processDomainEvent(ctx, msg)
	case "organization_db.public.organization_units":
		return c.processCDCEvent(ctx, msg)
	default:
		c.logger.Printf("⚠️ 未知主题: %s", topic)
		return nil
	}
}

func (c *KafkaEventConsumer) processDomainEvent(ctx context.Context, msg *kafka.Message) error {
	// 从消息头获取事件类型
	eventType := ""
	for _, header := range msg.Headers {
		if header.Key == "event-type" {
			eventType = string(header.Value)
			break
		}
	}

	c.logger.Printf("处理领域事件: %s", eventType)

	switch eventType {
	case "OrganizationCreated":
		var event OrganizationCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("反序列化OrganizationCreated事件失败: %w", err)
		}
		return c.syncSvc.HandleOrganizationCreated(ctx, event)

	case "OrganizationUpdated":
		var event OrganizationUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("反序列化OrganizationUpdated事件失败: %w", err)
		}
		return c.syncSvc.HandleOrganizationUpdated(ctx, event)

	case "OrganizationDeleted":
		var event OrganizationDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("反序列化OrganizationDeleted事件失败: %w", err)
		}
		return c.syncSvc.HandleOrganizationDeleted(ctx, event)

	default:
		c.logger.Printf("⚠️ 未知领域事件类型: %s", eventType)
		return nil
	}
}

func (c *KafkaEventConsumer) processCDCEvent(ctx context.Context, msg *kafka.Message) error {
	c.logger.Printf("处理CDC事件")

	// 解析Debezium消息格式
	var debeziumMsg struct {
		Payload CDCOrganizationEvent `json:"payload"`
	}
	if err := json.Unmarshal(msg.Value, &debeziumMsg); err != nil {
		return fmt.Errorf("反序列化Debezium消息失败: %w", err)
	}

	c.logger.Printf("CDC操作类型: %s", debeziumMsg.Payload.Op)
	return c.syncSvc.HandleCDCEvent(ctx, debeziumMsg.Payload)
}

func (c *KafkaEventConsumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// ===== 主程序 =====

func main() {
	logger := log.New(os.Stdout, "[NEO4J-SYNC] ", log.LstdFlags)

	// 创建Neo4j同步服务
	syncSvc, err := NewNeo4jSyncService("bolt://localhost:7687", "neo4j", "password", logger)
	if err != nil {
		log.Fatalf("创建Neo4j同步服务失败: %v", err)
	}
	defer syncSvc.Close()

	// 创建Kafka消费者
	consumer, err := NewKafkaEventConsumer(
		[]string{"localhost:9092"},
		"neo4j-sync-full-replay",  // 完全重放所有消息
		syncSvc,
		logger,
	)
	if err != nil {
		log.Fatalf("创建Kafka消费者失败: %v", err)
	}
	defer consumer.Close()

	// 订阅主题
	topics := []string{
		"organization.events",
		"organization_db.public.organization_units",
	}

	if err := consumer.Subscribe(topics); err != nil {
		log.Fatalf("订阅Kafka主题失败: %v", err)
	}

	logger.Printf("🚀 Neo4j同步服务启动成功")
	logger.Printf("监听主题: %v", topics)

	// 启动健康检查服务器
	go startHealthServer(logger)

	// 创建上下文处理优雅关闭
	ctx, cancel := context.WithCancel(context.Background())

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("正在关闭Neo4j同步服务...")
		cancel()
	}()

	// 开始消费
	if err := consumer.StartConsuming(ctx); err != nil {
		log.Fatalf("消费失败: %v", err)
	}

	logger.Println("Neo4j同步服务已关闭")
}

// 健康检查服务器
func startHealthServer(logger *log.Logger) {
	mux := http.NewServeMux()
	
	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "organization-sync-service",
			"status": "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"features": []string{
				"CDC数据捕获",
				"Neo4j实时同步", 
				"Kafka消息消费",
				"Debezium集成",
			},
		}
		json.NewEncoder(w).Encode(response)
	})
	
	// 指标端点
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Sync service metrics\nsync_service_status 1\n"))
	})
	
	server := &http.Server{
		Addr:    ":8084",
		Handler: mux,
	}
	
	logger.Printf("🔍 健康检查服务器启动 - 端口 8084")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Printf("❌ 健康检查服务器错误: %v", err)
	}
}