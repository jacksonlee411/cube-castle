package common

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

// BusinessIDService 业务ID管理服务
type BusinessIDService struct {
	db *sql.DB
}

// NewBusinessIDService 创建业务ID服务实例
func NewBusinessIDService(db *sql.DB) *BusinessIDService {
	return &BusinessIDService{db: db}
}

// EntityType 实体类型枚举
type EntityType string

const (
	EntityTypeEmployee     EntityType = "employee"
	EntityTypeOrganization EntityType = "organization"
	EntityTypePosition     EntityType = "position"
)

// BusinessIDRange 业务ID范围定义
type BusinessIDRange struct {
	Min    int64
	Max    int64
	Length int
	Prefix string
}

// GetBusinessIDRange 获取实体类型的业务ID范围
func GetBusinessIDRange(entityType EntityType) BusinessIDRange {
	switch entityType {
	case EntityTypeEmployee:
		return BusinessIDRange{Min: 1, Max: 99999, Length: 5, Prefix: ""}
	case EntityTypeOrganization:
		return BusinessIDRange{Min: 100000, Max: 999999, Length: 6, Prefix: ""}
	case EntityTypePosition:
		return BusinessIDRange{Min: 1000000, Max: 9999999, Length: 7, Prefix: ""}
	default:
		return BusinessIDRange{}
	}
}

// ValidateBusinessID 验证业务ID格式
func ValidateBusinessID(entityType EntityType, businessID string) error {
	if businessID == "" {
		return fmt.Errorf("business ID cannot be empty")
	}

	var pattern string
	var rangeDef BusinessIDRange

	switch entityType {
	case EntityTypeEmployee:
		pattern = `^[1-9][0-9]{0,4}$`
		rangeDef = GetBusinessIDRange(EntityTypeEmployee)
	case EntityTypeOrganization:
		pattern = `^[1-9][0-9]{5}$`
		rangeDef = GetBusinessIDRange(EntityTypeOrganization)
	case EntityTypePosition:
		pattern = `^[1-9][0-9]{6}$`
		rangeDef = GetBusinessIDRange(EntityTypePosition)
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	// 检查格式
	matched, err := regexp.MatchString(pattern, businessID)
	if err != nil {
		return fmt.Errorf("regex pattern error: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid business ID format for %s: %s", entityType, businessID)
	}

	// 检查范围
	id, err := strconv.ParseInt(businessID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid business ID number: %s", businessID)
	}

	if id < rangeDef.Min || id > rangeDef.Max {
		return fmt.Errorf("business ID %s out of range [%d-%d] for %s", 
			businessID, rangeDef.Min, rangeDef.Max, entityType)
	}

	return nil
}

// GenerateBusinessID 生成新的业务ID
func (s *BusinessIDService) GenerateBusinessID(ctx context.Context, entityType EntityType) (string, error) {
	var sequenceName string
	var offset int64

	switch entityType {
	case EntityTypeEmployee:
		sequenceName = "employee_business_id_seq"
		offset = 0
	case EntityTypeOrganization:
		sequenceName = "org_business_id_seq"
		offset = 100000
	case EntityTypePosition:
		sequenceName = "position_business_id_seq"
		offset = 1000000
	default:
		return "", fmt.Errorf("unknown entity type: %s", entityType)
	}

	var nextVal int64
	query := fmt.Sprintf("SELECT nextval('%s')", sequenceName)
	err := s.db.QueryRowContext(ctx, query).Scan(&nextVal)
	if err != nil {
		return "", fmt.Errorf("failed to generate business ID for %s: %w", entityType, err)
	}

	businessID := strconv.FormatInt(nextVal+offset, 10)

	// 验证生成的ID是否在有效范围内
	if err := ValidateBusinessID(entityType, businessID); err != nil {
		return "", fmt.Errorf("generated invalid business ID: %w", err)
	}

	return businessID, nil
}

// IsUUID 检查字符串是否为UUID格式
func IsUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

// BusinessIDGenerator 业务ID生成器接口
type BusinessIDGenerator interface {
	GenerateBusinessID(ctx context.Context, entityType EntityType) (string, error)
}

// BusinessIDValidator 业务ID验证器接口
type BusinessIDValidator interface {
	ValidateBusinessID(entityType EntityType, businessID string) error
}

// BusinessIDManagerConfig 业务ID管理器配置
type BusinessIDManagerConfig struct {
	EnableAutoGeneration bool
	EnableValidation     bool
	MaxRetries          int
}

// DefaultBusinessIDManagerConfig 默认配置
func DefaultBusinessIDManagerConfig() BusinessIDManagerConfig {
	return BusinessIDManagerConfig{
		EnableAutoGeneration: true,
		EnableValidation:     true,
		MaxRetries:          3,
	}
}

// BusinessIDManager 业务ID管理器
type BusinessIDManager struct {
	service   *BusinessIDService
	config    BusinessIDManagerConfig
}

// NewBusinessIDManager 创建业务ID管理器
func NewBusinessIDManager(service *BusinessIDService, config BusinessIDManagerConfig) *BusinessIDManager {
	return &BusinessIDManager{
		service: service,
		config:  config,
	}
}

// GenerateUniqueBusinessID 生成唯一的业务ID（带重试机制）
func (m *BusinessIDManager) GenerateUniqueBusinessID(ctx context.Context, entityType EntityType) (string, error) {
	if !m.config.EnableAutoGeneration {
		return "", fmt.Errorf("auto generation is disabled")
	}

	var lastErr error
	for i := 0; i < m.config.MaxRetries; i++ {
		businessID, err := m.service.GenerateBusinessID(ctx, entityType)
		if err != nil {
			lastErr = err
			continue
		}

		if m.config.EnableValidation {
			if err := ValidateBusinessID(entityType, businessID); err != nil {
				lastErr = err
				continue
			}
		}

		return businessID, nil
	}

	return "", fmt.Errorf("failed to generate unique business ID after %d retries: %w", 
		m.config.MaxRetries, lastErr)
}

// BusinessIDLookupResult 业务ID查询结果
type BusinessIDLookupResult struct {
	BusinessID string    `json:"business_id"`
	UUID      uuid.UUID `json:"uuid,omitempty"`
	Found     bool      `json:"found"`
}

// LookupByBusinessID 通过业务ID查找UUID
func (s *BusinessIDService) LookupByBusinessID(ctx context.Context, entityType EntityType, businessID string) (*BusinessIDLookupResult, error) {
	if err := ValidateBusinessID(entityType, businessID); err != nil {
		return nil, err
	}

	var tableName string
	switch entityType {
	case EntityTypeEmployee:
		tableName = "employees"
	case EntityTypeOrganization:
		tableName = "organization_units"
	case EntityTypePosition:
		tableName = "positions"
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}

	var uuidStr string
	query := fmt.Sprintf("SELECT id FROM %s WHERE business_id = $1", tableName)
	err := s.db.QueryRowContext(ctx, query, businessID).Scan(&uuidStr)
	
	if err == sql.ErrNoRows {
		return &BusinessIDLookupResult{
			BusinessID: businessID,
			Found:     false,
		}, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("lookup failed for business ID %s: %w", businessID, err)
	}

	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	return &BusinessIDLookupResult{
		BusinessID: businessID,
		UUID:      parsedUUID,
		Found:     true,
	}, nil
}

// LookupByUUID 通过UUID查找业务ID
func (s *BusinessIDService) LookupByUUID(ctx context.Context, entityType EntityType, id uuid.UUID) (*BusinessIDLookupResult, error) {
	var tableName string
	switch entityType {
	case EntityTypeEmployee:
		tableName = "employees"  
	case EntityTypeOrganization:
		tableName = "organization_units"
	case EntityTypePosition:
		tableName = "positions"
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}

	var businessID string
	query := fmt.Sprintf("SELECT business_id FROM %s WHERE id = $1", tableName)
	err := s.db.QueryRowContext(ctx, query, id).Scan(&businessID)
	
	if err == sql.ErrNoRows {
		return &BusinessIDLookupResult{
			UUID:  id,
			Found: false,
		}, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("lookup failed for UUID %s: %w", id, err)
	}

	return &BusinessIDLookupResult{
		BusinessID: businessID,
		UUID:      id,
		Found:     true,
	}, nil
}