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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// 项目默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 命令模型 =====

// 组织命令基础接口
type OrganizationCommand interface {
	GetCommandID() uuid.UUID
	GetTenantID() uuid.UUID
	GetCommandType() string
	Validate() error
}

// 创建组织命令
type CreateOrganizationCommand struct {
	CommandID    uuid.UUID `json:"command_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	RequestedCode *string   `json:"requested_code,omitempty"` // 用户请求的编码
	Name         string    `json:"name" validate:"required,min=1,max=100"`
	ParentCode   *string   `json:"parent_code,omitempty"`
	UnitType     string    `json:"unit_type" validate:"required,oneof=COMPANY DEPARTMENT TEAM"`
	Description  *string   `json:"description,omitempty"`
	SortOrder    *int      `json:"sort_order,omitempty"`
	RequestedBy  uuid.UUID `json:"requested_by" validate:"required"`
}

func (c CreateOrganizationCommand) GetCommandID() uuid.UUID { return c.CommandID }
func (c CreateOrganizationCommand) GetTenantID() uuid.UUID  { return c.TenantID }
func (c CreateOrganizationCommand) GetCommandType() string  { return "CreateOrganization" }
func (c CreateOrganizationCommand) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}

// 更新组织命令
type UpdateOrganizationCommand struct {
	CommandID   uuid.UUID `json:"command_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Code        string    `json:"code" validate:"required"`
	Name        *string   `json:"name,omitempty"`
	Status      *string   `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Description *string   `json:"description,omitempty"`
	SortOrder   *int      `json:"sort_order,omitempty"`
	RequestedBy uuid.UUID `json:"requested_by" validate:"required"`
}

func (c UpdateOrganizationCommand) GetCommandID() uuid.UUID { return c.CommandID }
func (c UpdateOrganizationCommand) GetTenantID() uuid.UUID  { return c.TenantID }
func (c UpdateOrganizationCommand) GetCommandType() string  { return "UpdateOrganization" }
func (c UpdateOrganizationCommand) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}

// 删除组织命令
type DeleteOrganizationCommand struct {
	CommandID   uuid.UUID `json:"command_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Code        string    `json:"code" validate:"required"`
	RequestedBy uuid.UUID `json:"requested_by" validate:"required"`
}

func (c DeleteOrganizationCommand) GetCommandID() uuid.UUID { return c.CommandID }
func (c DeleteOrganizationCommand) GetTenantID() uuid.UUID  { return c.TenantID }
func (c DeleteOrganizationCommand) GetCommandType() string  { return "DeleteOrganization" }
func (c DeleteOrganizationCommand) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}

// ===== 命令结果模型 =====

type CreateOrganizationResult struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	UnitType  string    `json:"unit_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateOrganizationResult struct {
	Code      string    `json:"code"`
	UpdatedAt time.Time `json:"updated_at"`
	Changes   map[string]interface{} `json:"changes"`
}

type DeleteOrganizationResult struct {
	Code      string    `json:"code"`
	DeletedAt time.Time `json:"deleted_at"`
}

// ===== 事件模型 =====

// 组织事件基础接口
type OrganizationEvent interface {
	GetEventID() uuid.UUID
	GetAggregateID() string
	GetTenantID() uuid.UUID
	GetEventType() string
	GetEventTime() time.Time
}

// 组织创建事件
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

func (e OrganizationCreatedEvent) GetEventID() uuid.UUID     { return e.EventID }
func (e OrganizationCreatedEvent) GetAggregateID() string    { return e.AggregateID }
func (e OrganizationCreatedEvent) GetTenantID() uuid.UUID    { return e.TenantID }
func (e OrganizationCreatedEvent) GetEventType() string      { return "OrganizationCreated" }
func (e OrganizationCreatedEvent) GetEventTime() time.Time   { return e.CreatedAt }

// 组织更新事件
type OrganizationUpdatedEvent struct {
	EventID     uuid.UUID              `json:"event_id"`
	AggregateID string                 `json:"aggregate_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Changes     map[string]interface{} `json:"changes"`
	UpdatedBy   uuid.UUID              `json:"updated_by"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func (e OrganizationUpdatedEvent) GetEventID() uuid.UUID     { return e.EventID }
func (e OrganizationUpdatedEvent) GetAggregateID() string    { return e.AggregateID }
func (e OrganizationUpdatedEvent) GetTenantID() uuid.UUID    { return e.TenantID }
func (e OrganizationUpdatedEvent) GetEventType() string      { return "OrganizationUpdated" }
func (e OrganizationUpdatedEvent) GetEventTime() time.Time   { return e.UpdatedAt }

// 组织删除事件
type OrganizationDeletedEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	AggregateID string    `json:"aggregate_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	DeletedBy   uuid.UUID `json:"deleted_by"`
	DeletedAt   time.Time `json:"deleted_at"`
}

func (e OrganizationDeletedEvent) GetEventID() uuid.UUID     { return e.EventID }
func (e OrganizationDeletedEvent) GetAggregateID() string    { return e.AggregateID }
func (e OrganizationDeletedEvent) GetTenantID() uuid.UUID    { return e.TenantID }
func (e OrganizationDeletedEvent) GetEventType() string      { return "OrganizationDeleted" }
func (e OrganizationDeletedEvent) GetEventTime() time.Time   { return e.DeletedAt }

// ===== Kafka事件总线 =====

type KafkaEventBus struct {
	producer *kafka.Producer
	logger   *log.Logger
}

func NewKafkaEventBus(brokers []string, logger *log.Logger) (*KafkaEventBus, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"client.id":         "organization-command-service",
		"acks":             "all",
		"retries":          3,
		"batch.size":       16384,
		"linger.ms":        10,
	})

	if err != nil {
		return nil, fmt.Errorf("创建Kafka生产者失败: %w", err)
	}

	return &KafkaEventBus{
		producer: producer,
		logger:   logger,
	}, nil
}

func (bus *KafkaEventBus) Publish(ctx context.Context, topic string, event OrganizationEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("事件序列化失败: %w", err)
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.GetAggregateID()),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.GetEventType())},
			{Key: "tenant-id", Value: []byte(event.GetTenantID().String())},
			{Key: "event-id", Value: []byte(event.GetEventID().String())},
		},
	}

	deliveryChan := make(chan kafka.Event, 1)
	err = bus.producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("事件发布失败: %w", err)
	}

	// 等待发布确认（带超时）
	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("事件发布确认失败: %w", m.TopicPartition.Error)
		}
		bus.logger.Printf("事件发布成功: topic=%s, partition=%d, offset=%d, event_id=%s",
			topic, m.TopicPartition.Partition, m.TopicPartition.Offset, event.GetEventID())
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("事件发布超时")
	}
}

func (bus *KafkaEventBus) Close() {
	if bus.producer != nil {
		bus.producer.Close()
	}
}

// ===== PostgreSQL仓储层 =====

type PostgresOrganizationRepository struct {
	pool   *pgxpool.Pool
	logger *log.Logger
}

func NewPostgresOrganizationRepository(pool *pgxpool.Pool, logger *log.Logger) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *PostgresOrganizationRepository) CreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error) {
	// 确定使用的组织代码
	var code string
	var err error
	
	if cmd.RequestedCode != nil && *cmd.RequestedCode != "" {
		// 使用用户提供的编码，但需要验证唯一性
		code = *cmd.RequestedCode
		exists, err := r.codeExists(ctx, code, cmd.TenantID)
		if err != nil {
			return nil, fmt.Errorf("检查编码唯一性失败: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("组织编码 '%s' 已存在", code)
		}
	} else {
		// 自动生成编码
		code, err = r.generateOrganizationCode(ctx, cmd.TenantID)
		if err != nil {
			return nil, fmt.Errorf("生成组织代码失败: %w", err)
		}
	}

	// 计算层级和路径
	level := 1
	path := fmt.Sprintf("/%s", code)
	if cmd.ParentCode != nil {
		parentInfo, err := r.getParentInfo(ctx, *cmd.ParentCode, cmd.TenantID)
		if err != nil {
			return nil, fmt.Errorf("获取父组织信息失败: %w", err)
		}
		level = parentInfo.Level + 1
		path = fmt.Sprintf("%s/%s", parentInfo.Path, code)
	}

	sortOrder := 0
	if cmd.SortOrder != nil {
		sortOrder = *cmd.SortOrder
	}

	// 执行插入操作
	query := `
		INSERT INTO organization_units (
			code, parent_code, tenant_id, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, 'ACTIVE', $6, $7, $8, $9, $10, $10
		)
		RETURNING created_at`

	var createdAt time.Time
	err = r.pool.QueryRow(ctx, query,
		code, cmd.ParentCode, cmd.TenantID, cmd.Name, cmd.UnitType,
		level, path, sortOrder, cmd.Description, time.Now(),
	).Scan(&createdAt)

	if err != nil {
		return nil, fmt.Errorf("插入组织记录失败: %w", err)
	}

	r.logger.Printf("组织创建成功: code=%s, name=%s (用户提供编码: %v)", 
		code, cmd.Name, cmd.RequestedCode != nil)

	return &CreateOrganizationResult{
		Code:      code,
		Name:      cmd.Name,
		UnitType:  cmd.UnitType,
		Status:    "ACTIVE",
		CreatedAt: createdAt,
	}, nil
}

func (r *PostgresOrganizationRepository) UpdateOrganization(ctx context.Context, cmd UpdateOrganizationCommand) (*UpdateOrganizationResult, error) {
	// 构建动态更新查询
	setParts := []string{}
	args := []interface{}{}
	changes := make(map[string]interface{})

	// 收集需要更新的字段
	if cmd.Name != nil {
		setParts = append(setParts, "name = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *cmd.Name)
		changes["name"] = *cmd.Name
	}

	if cmd.Status != nil {
		setParts = append(setParts, "status = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *cmd.Status)
		changes["status"] = *cmd.Status
	}

	if cmd.Description != nil {
		setParts = append(setParts, "description = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *cmd.Description)
		changes["description"] = *cmd.Description
	}

	if cmd.SortOrder != nil {
		setParts = append(setParts, "sort_order = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *cmd.SortOrder)
		changes["sort_order"] = *cmd.SortOrder
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("没有提供更新字段")
	}

	// 添加updated_at字段
	now := time.Now()
	setParts = append(setParts, "updated_at = $"+fmt.Sprintf("%d", len(args)+1))
	args = append(args, now)

	// 添加WHERE条件参数
	args = append(args, cmd.Code)
	whereCode := "$" + fmt.Sprintf("%d", len(args))
	args = append(args, cmd.TenantID)
	whereTenant := "$" + fmt.Sprintf("%d", len(args))

	query := fmt.Sprintf(`
		UPDATE organization_units 
		SET %s
		WHERE code = %s AND tenant_id = %s`,
		strings.Join(setParts, ", "), whereCode, whereTenant)

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("更新组织记录失败: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("组织不存在或无权限: %s", cmd.Code)
	}

	r.logger.Printf("组织更新成功: code=%s, changes=%v", cmd.Code, changes)

	return &UpdateOrganizationResult{
		Code:      cmd.Code,
		UpdatedAt: now,
		Changes:   changes,
	}, nil
}

func (r *PostgresOrganizationRepository) DeleteOrganization(ctx context.Context, cmd DeleteOrganizationCommand) (*DeleteOrganizationResult, error) {
	// 检查是否有子组织
	var childCount int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM organization_units WHERE parent_code = $1 AND tenant_id = $2",
		cmd.Code, cmd.TenantID,
	).Scan(&childCount)

	if err != nil {
		return nil, fmt.Errorf("检查子组织失败: %w", err)
	}

	if childCount > 0 {
		return nil, fmt.Errorf("无法删除组织，存在 %d 个子组织", childCount)
	}

	// 执行软删除
	now := time.Now()
	result, err := r.pool.Exec(ctx,
		"UPDATE organization_units SET status = 'INACTIVE', updated_at = $1 WHERE code = $2 AND tenant_id = $3",
		now, cmd.Code, cmd.TenantID,
	)

	if err != nil {
		return nil, fmt.Errorf("删除组织记录失败: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("组织不存在或无权限: %s", cmd.Code)
	}

	r.logger.Printf("组织删除成功: code=%s", cmd.Code)

	return &DeleteOrganizationResult{
		Code:      cmd.Code,
		DeletedAt: now,
	}, nil
}

// 辅助方法
func (r *PostgresOrganizationRepository) codeExists(ctx context.Context, code string, tenantID uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM organization_units WHERE code = $1 AND tenant_id = $2",
		code, tenantID,
	).Scan(&count)
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

func (r *PostgresOrganizationRepository) generateOrganizationCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var maxCode int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(code AS INTEGER)), 1000000) 
		 FROM organization_units 
		 WHERE tenant_id = $1 AND code ~ '^[0-9]+$'`,
		tenantID,
	).Scan(&maxCode)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", maxCode+1), nil
}

type ParentInfo struct {
	Level int
	Path  string
}

func (r *PostgresOrganizationRepository) getParentInfo(ctx context.Context, parentCode string, tenantID uuid.UUID) (*ParentInfo, error) {
	var info ParentInfo
	err := r.pool.QueryRow(ctx,
		"SELECT level, path FROM organization_units WHERE code = $1 AND tenant_id = $2",
		parentCode, tenantID,
	).Scan(&info.Level, &info.Path)

	if err != nil {
		return nil, err
	}

	return &info, nil
}

// ===== 命令处理器 =====

type OrganizationCommandHandler struct {
	repo        *PostgresOrganizationRepository
	eventBus    *KafkaEventBus
	logger      *log.Logger
	validator   *validator.Validate
}

func NewOrganizationCommandHandler(
	repo *PostgresOrganizationRepository,
	eventBus *KafkaEventBus,
	logger *log.Logger,
) *OrganizationCommandHandler {
	return &OrganizationCommandHandler{
		repo:      repo,
		eventBus:  eventBus,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *OrganizationCommandHandler) HandleCreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error) {
	h.logger.Printf("处理创建组织命令 - 租户: %s, 名称: %s, 命令ID: %s",
		cmd.TenantID, cmd.Name, cmd.CommandID)

	// 1. 命令验证
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("命令验证失败: %w", err)
	}

	// 2. 业务规则验证
	if err := h.validateCreateBusinessRules(ctx, cmd); err != nil {
		return nil, fmt.Errorf("业务规则验证失败: %w", err)
	}

	// 3. 执行命令
	result, err := h.repo.CreateOrganization(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	// 4. 发布事件
	event := OrganizationCreatedEvent{
		EventID:     uuid.New(),
		AggregateID: result.Code,
		TenantID:    cmd.TenantID,
		Name:        cmd.Name,
		UnitType:    cmd.UnitType,
		ParentCode:  cmd.ParentCode,
		CreatedBy:   cmd.RequestedBy,
		CreatedAt:   result.CreatedAt,
	}

	if err := h.eventBus.Publish(ctx, "organization.events", event); err != nil {
		h.logger.Printf("事件发布失败 (非致命): %v", err)
		// 注意：事件发布失败不应该回滚业务操作
	}

	h.logger.Printf("组织创建成功: code=%s", result.Code)
	return result, nil
}

func (h *OrganizationCommandHandler) HandleUpdateOrganization(ctx context.Context, cmd UpdateOrganizationCommand) (*UpdateOrganizationResult, error) {
	h.logger.Printf("处理更新组织命令 - 租户: %s, 代码: %s, 命令ID: %s",
		cmd.TenantID, cmd.Code, cmd.CommandID)

	// 1. 命令验证
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("命令验证失败: %w", err)
	}

	// 2. 执行命令
	result, err := h.repo.UpdateOrganization(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("更新组织失败: %w", err)
	}

	// 3. 发布事件
	event := OrganizationUpdatedEvent{
		EventID:     uuid.New(),
		AggregateID: cmd.Code,
		TenantID:    cmd.TenantID,
		Changes:     result.Changes,
		UpdatedBy:   cmd.RequestedBy,
		UpdatedAt:   result.UpdatedAt,
	}

	if err := h.eventBus.Publish(ctx, "organization.events", event); err != nil {
		h.logger.Printf("事件发布失败 (非致命): %v", err)
	}

	h.logger.Printf("组织更新成功: code=%s", result.Code)
	return result, nil
}

func (h *OrganizationCommandHandler) HandleDeleteOrganization(ctx context.Context, cmd DeleteOrganizationCommand) (*DeleteOrganizationResult, error) {
	h.logger.Printf("处理删除组织命令 - 租户: %s, 代码: %s, 命令ID: %s",
		cmd.TenantID, cmd.Code, cmd.CommandID)

	// 1. 命令验证
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("命令验证失败: %w", err)
	}

	// 2. 执行命令
	result, err := h.repo.DeleteOrganization(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("删除组织失败: %w", err)
	}

	// 3. 发布事件
	event := OrganizationDeletedEvent{
		EventID:     uuid.New(),
		AggregateID: cmd.Code,
		TenantID:    cmd.TenantID,
		DeletedBy:   cmd.RequestedBy,
		DeletedAt:   result.DeletedAt,
	}

	if err := h.eventBus.Publish(ctx, "organization.events", event); err != nil {
		h.logger.Printf("事件发布失败 (非致命): %v", err)
	}

	h.logger.Printf("组织删除成功: code=%s", result.Code)
	return result, nil
}

func (h *OrganizationCommandHandler) validateCreateBusinessRules(ctx context.Context, cmd CreateOrganizationCommand) error {
	// 这里可以添加更多业务规则验证
	// 例如：父组织是否存在、名称是否重复等
	return nil
}

// ===== HTTP API处理器 =====

type CommandAPIHandler struct {
	commandHandler *OrganizationCommandHandler
	logger         *log.Logger
}

func NewCommandAPIHandler(commandHandler *OrganizationCommandHandler, logger *log.Logger) *CommandAPIHandler {
	return &CommandAPIHandler{
		commandHandler: commandHandler,
		logger:         logger,
	}
}

func (h *CommandAPIHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// 解析租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		tenantIDStr = DefaultTenantIDString
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req struct {
		Code        *string `json:"code,omitempty"`        // 添加用户输入的编码字段
		Name        string  `json:"name"`
		ParentCode  *string `json:"parent_code,omitempty"`
		UnitType    string  `json:"unit_type"`
		Description *string `json:"description,omitempty"`
		SortOrder   *int    `json:"sort_order,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 构建命令
	cmd := CreateOrganizationCommand{
		CommandID:    uuid.New(),
		TenantID:     tenantID,
		RequestedCode: req.Code,               // 传递用户提供的编码
		Name:         req.Name,
		ParentCode:   req.ParentCode,
		UnitType:     req.UnitType,
		Description:  req.Description,
		SortOrder:    req.SortOrder,
		RequestedBy:  uuid.New(), // 实际应用中应从认证信息获取
	}

	// 执行命令
	result, err := h.commandHandler.HandleCreateOrganization(r.Context(), cmd)
	if err != nil {
		h.logger.Printf("创建组织API失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *CommandAPIHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	// 解析租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		tenantIDStr = DefaultTenantIDString
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// 获取组织代码
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "Organization code is required", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var req struct {
		Name        *string `json:"name,omitempty"`
		Status      *string `json:"status,omitempty"`
		Description *string `json:"description,omitempty"`
		SortOrder   *int    `json:"sort_order,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 构建命令
	cmd := UpdateOrganizationCommand{
		CommandID:   uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		Name:        req.Name,
		Status:      req.Status,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		RequestedBy: uuid.New(),
	}

	// 执行命令
	result, err := h.commandHandler.HandleUpdateOrganization(r.Context(), cmd)
	if err != nil {
		h.logger.Printf("更新组织API失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *CommandAPIHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// 解析租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		tenantIDStr = DefaultTenantIDString
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// 获取组织代码
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "Organization code is required", http.StatusBadRequest)
		return
	}

	// 构建命令
	cmd := DeleteOrganizationCommand{
		CommandID:   uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		RequestedBy: uuid.New(),
	}

	// 执行命令
	result, err := h.commandHandler.HandleDeleteOrganization(r.Context(), cmd)
	if err != nil {
		h.logger.Printf("删除组织API失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ===== 主程序 =====

func main() {
	logger := log.New(os.Stdout, "[ORG-COMMAND] ", log.LstdFlags)

	// 数据库连接
	dbConfig, err := pgxpool.ParseConfig("postgresql://user:password@localhost:5432/cubecastle")
	if err != nil {
		log.Fatalf("解析数据库配置失败: %v", err)
	}
	dbConfig.MaxConns = 10

	dbPool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("创建数据库连接池失败: %v", err)
	}
	defer dbPool.Close()

	// 测试数据库连接
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// Kafka事件总线
	eventBus, err := NewKafkaEventBus([]string{"localhost:9092"}, logger)
	if err != nil {
		log.Fatalf("创建Kafka事件总线失败: %v", err)
	}
	defer eventBus.Close()

	// 创建依赖组件
	repo := NewPostgresOrganizationRepository(dbPool, logger)
	commandHandler := NewOrganizationCommandHandler(repo, eventBus, logger)
	apiHandler := NewCommandAPIHandler(commandHandler, logger)

	// 创建HTTP路由器
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 命令端API路由
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		r.Post("/", apiHandler.CreateOrganization)
		r.Put("/{code}", apiHandler.UpdateOrganization)
		r.Delete("/{code}", apiHandler.DeleteOrganization)
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    ":9090",
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("正在关闭命令端服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("命令端服务器关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 CQRS组织命令端服务器启动在端口 :9090")
	logger.Printf("严格按照CQRS统一实施指南标准实现")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("命令端服务器启动失败: %v", err)
	}

	logger.Println("命令端服务器已关闭")
}