package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"cube-castle-deployment-test/internal/cache"
)

// Neo4j查询服务 - L3数据源实现
type Neo4jQueryService struct {
	driver neo4j.DriverWithContext
	logger *log.Logger
}

func NewNeo4jQueryService(driver neo4j.DriverWithContext, logger *log.Logger) *Neo4jQueryService {
	return &Neo4jQueryService{
		driver: driver,
		logger: logger,
	}
}

// 实现L3QueryInterface接口

// 获取组织列表
func (service *Neo4jQueryService) GetOrganizations(ctx context.Context, tenantID uuid.UUID, params cache.QueryParams) ([]cache.Organization, error) {
	session := service.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// 构建查询条件
	searchCondition := ""
	queryParams := map[string]interface{}{
		"tenant_id": tenantID.String(),
		"first":     int64(params.First),
		"offset":    int64(params.Offset),
	}

	if params.SearchText != "" {
		searchCondition = "AND (o.name CONTAINS $searchText OR o.code CONTAINS $searchText)"
		queryParams["searchText"] = params.SearchText
	}

	query := fmt.Sprintf(`
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		WHERE o.status <> 'DELETED' %s
		RETURN o.code as code, o.name as name, o.unit_type as unit_type,
			   o.status as status, o.level as level, o.path as path,
			   o.sort_order as sort_order, o.description as description,
			   o.parent_code as parent_code,
			   toString(o.created_at) as created_at, toString(o.updated_at) as updated_at
		ORDER BY o.sort_order ASC, o.code ASC
		SKIP $offset LIMIT $first
	`, searchCondition)

	result, err := session.Run(ctx, query, queryParams)
	if err != nil {
		return nil, fmt.Errorf("Neo4j查询失败: %w", err)
	}

	var organizations []cache.Organization
	for result.Next(ctx) {
		record := result.Record()

		org := cache.Organization{
			Code:        getStringValue(record, "code"),
			TenantID:    tenantID.String(),
			Name:        getStringValue(record, "name"),
			UnitType:    getStringValue(record, "unit_type"),
			Status:      getStringValue(record, "status"),
			Level:       getIntValue(record, "level"),
			Path:        getStringValue(record, "path"),
			SortOrder:   getIntValue(record, "sort_order"),
			Description: getStringValue(record, "description"),
			ParentCode:  getStringValue(record, "parent_code"),
		}

		// 解析时间字段
		if createdAt := getStringValue(record, "created_at"); createdAt != "" {
			if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
				org.CreatedAt = t
			}
		}

		if updatedAt := getStringValue(record, "updated_at"); updatedAt != "" {
			if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
				org.UpdatedAt = t
			}
		}

		organizations = append(organizations, org)
	}

	service.logger.Printf("[L3] Neo4j组织列表查询完成: 返回%d条记录", len(organizations))
	return organizations, result.Err()
}

// 获取单个组织
func (service *Neo4jQueryService) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*cache.Organization, error) {
	session := service.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id, code: $code})
		WHERE o.status <> 'DELETED'
		RETURN o.code as code, o.name as name, o.unit_type as unit_type,
			   o.status as status, o.level as level, o.path as path,
			   o.sort_order as sort_order, o.description as description,
			   o.parent_code as parent_code,
			   toString(o.created_at) as created_at, toString(o.updated_at) as updated_at
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"tenant_id": tenantID.String(),
		"code":      code,
	})
	if err != nil {
		return nil, fmt.Errorf("Neo4j单个组织查询失败: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()

		org := &cache.Organization{
			Code:        getStringValue(record, "code"),
			TenantID:    tenantID.String(),
			Name:        getStringValue(record, "name"),
			UnitType:    getStringValue(record, "unit_type"),
			Status:      getStringValue(record, "status"),
			Level:       getIntValue(record, "level"),
			Path:        getStringValue(record, "path"),
			SortOrder:   getIntValue(record, "sort_order"),
			Description: getStringValue(record, "description"),
			ParentCode:  getStringValue(record, "parent_code"),
		}

		// 解析时间字段
		if createdAt := getStringValue(record, "created_at"); createdAt != "" {
			if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
				org.CreatedAt = t
			}
		}

		if updatedAt := getStringValue(record, "updated_at"); updatedAt != "" {
			if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
				org.UpdatedAt = t
			}
		}

		service.logger.Printf("[L3] Neo4j单个组织查询完成: %s", code)
		return org, nil
	}

	return nil, nil // 未找到
}

// 获取组织统计信息
func (service *Neo4jQueryService) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*cache.OrganizationStats, error) {
	session := service.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// 1. 获取总数
	totalQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		WHERE o.status <> 'DELETED'
		RETURN count(o) as total
	`

	totalResult, err := session.Run(ctx, totalQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	var total int
	if totalResult.Next(ctx) {
		record := totalResult.Record()
		total = getIntValue(record, "total")
	}

	// 2. 按类型统计
	typeQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		WHERE o.status <> 'DELETED'
		RETURN o.unit_type as unit_type, count(o) as count
		ORDER BY unit_type
	`

	typeResult, err := session.Run(ctx, typeQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("按类型统计失败: %w", err)
	}

	var byType []cache.TypeCount
	for typeResult.Next(ctx) {
		record := typeResult.Record()
		unitType := getStringValue(record, "unit_type")
		count := getIntValue(record, "count")
		byType = append(byType, cache.TypeCount{
			UnitType: unitType,
			Count:    count,
		})
	}

	// 3. 按状态统计
	statusQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		WHERE o.status <> 'DELETED'
		RETURN o.status as status, count(o) as count
		ORDER BY status
	`

	statusResult, err := session.Run(ctx, statusQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("按状态统计失败: %w", err)
	}

	var byStatus []cache.StatusCount
	for statusResult.Next(ctx) {
		record := statusResult.Record()
		status := getStringValue(record, "status")
		count := getIntValue(record, "count")
		byStatus = append(byStatus, cache.StatusCount{
			Status: status,
			Count:  count,
		})
	}

	// 4. 按级别统计
	levelQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		WHERE o.status <> 'DELETED'
		RETURN toString(o.level) as level, count(o) as count
		ORDER BY level
	`

	levelResult, err := session.Run(ctx, levelQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("按级别统计失败: %w", err)
	}

	var byLevel []cache.LevelCount
	for levelResult.Next(ctx) {
		record := levelResult.Record()
		level := getStringValue(record, "level")
		count := getIntValue(record, "count")
		byLevel = append(byLevel, cache.LevelCount{
			Level: fmt.Sprintf("级别%s", level),
			Count: count,
		})
	}

	// 构建统计结果
	stats := &cache.OrganizationStats{
		TotalCount: total,
		ByType:     byType,
		ByStatus:   byStatus,
		ByLevel:    byLevel,
	}

	service.logger.Printf("[L3] Neo4j统计查询完成: 总数=%d, 类型=%d, 状态=%d, 级别=%d",
		total, len(byType), len(byStatus), len(byLevel))

	return stats, nil
}

// 辅助函数：安全获取字符串值
func getStringValue(record *neo4j.Record, key string) string {
	if value, ok := record.Get(key); ok && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：安全获取整数值
func getIntValue(record *neo4j.Record, key string) int {
	if value, ok := record.Get(key); ok && value != nil {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}
