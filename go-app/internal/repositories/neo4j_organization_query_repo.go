package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
)

// Neo4jOrganizationQueryRepository Neo4j组织查询仓储实现
type Neo4jOrganizationQueryRepository struct {
	driver neo4j.DriverWithContext
	logger Logger
}

// NewNeo4jOrganizationQueryRepository 创建Neo4j组织查询仓储
func NewNeo4jOrganizationQueryRepository(driver neo4j.DriverWithContext, logger Logger) *Neo4jOrganizationQueryRepository {
	return &Neo4jOrganizationQueryRepository{
		driver: driver,
		logger: logger,
	}
}

// GetOrganization 获取单个组织
func (r *Neo4jOrganizationQueryRepository) GetOrganization(ctx context.Context, query queries.GetOrganizationQuery) (*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
		OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
		RETURN o, 
			   p.id as parent_id,
			   COUNT(c) as child_count`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":        query.ID.String(),
			"tenant_id": query.TenantID.String(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			return r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
		}

		return nil, nil
	})

	if err != nil {
		r.logger.Error("Failed to get organization", "error", err, "org_id", query.ID)
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("organization not found")
	}

	org := result.(*Organization)
	r.logger.Info("Organization retrieved successfully", "org_id", org.ID, "name", org.Name)
	return org, nil
}

// ListOrganizations 获取组织列表
func (r *Neo4jOrganizationQueryRepository) ListOrganizations(ctx context.Context, query queries.ListOrganizationsQuery) ([]*Organization, *PaginationInfo, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	// 构建过滤条件
	whereClause, params := r.buildWhereClause(query)
	params["tenant_id"] = query.TenantID.String()

	// 计算总数
	countCypher := fmt.Sprintf(`
		MATCH (o:Organization {tenant_id: $tenant_id})
		%s
		RETURN COUNT(o) as total`, whereClause)

	totalCount, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, countCypher, params)
		if err != nil {
			return 0, err
		}
		if result.Next(ctx) {
			count, _ := result.Record().Get("total")
			return count.(int64), nil
		}
		return 0, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to count organizations: %w", err)
	}

	// 获取分页数据
	offset := (query.Page - 1) * query.PageSize
	params["offset"] = offset
	params["limit"] = query.PageSize

	listCypher := fmt.Sprintf(`
		MATCH (o:Organization {tenant_id: $tenant_id})
		%s
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
		OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
		RETURN o, 
			   p.id as parent_id,
			   COUNT(c) as child_count
		ORDER BY o.level ASC, o.created_at DESC
		SKIP $offset LIMIT $limit`, whereClause)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, listCypher, params)
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to list organizations", "error", err)
		return nil, nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	organizations := result.([]*Organization)

	// 构建分页信息
	total := totalCount.(int64)
	totalPages := (total + int64(query.PageSize) - 1) / int64(query.PageSize)
	paginationInfo := &PaginationInfo{
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      int(total),
		TotalPages: int(totalPages),
		HasNext:    query.Page < int(totalPages),
		HasPrev:    query.Page > 1,
	}

	r.logger.Info("Organizations listed successfully", "count", len(organizations), "total", total)
	return organizations, paginationInfo, nil
}

// GetOrganizationTree 获取组织树
func (r *Neo4jOrganizationQueryRepository) GetOrganizationTree(ctx context.Context, query queries.GetOrganizationTreeQuery) ([]*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	var whereClause strings.Builder
	params := map[string]any{
		"tenant_id": query.TenantID.String(),
	}

	// 构建查询条件
	whereClause.WriteString("WHERE o.tenant_id = $tenant_id")

	if query.RootUnitID != nil {
		whereClause.WriteString(" AND root.id = $root_id")
		params["root_id"] = query.RootUnitID.String()
	}

	if query.MaxDepth > 0 {
		whereClause.WriteString(" AND LENGTH(path) <= $max_depth")
		params["max_depth"] = query.MaxDepth
	}

	if !query.IncludeInactive {
		whereClause.WriteString(" AND o.is_active = true")
	}

	cypher := fmt.Sprintf(`
		MATCH path = (root:Organization)-[:PARENT_OF*0..%d]->(o:Organization)
		%s
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
		OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
		RETURN o, 
			   p.id as parent_id,
			   COUNT(c) as child_count,
			   LENGTH(path) as depth
		ORDER BY LENGTH(path), o.level, o.name`,
		r.getMaxDepthLimit(query.MaxDepth),
		whereClause.String())

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to get organization tree", "error", err)
		return nil, fmt.Errorf("failed to get organization tree: %w", err)
	}

	organizations := result.([]*Organization)
	r.logger.Info("Organization tree retrieved successfully", "count", len(organizations))
	return organizations, nil
}

// GetOrganizationStats 获取组织统计
func (r *Neo4jOrganizationQueryRepository) GetOrganizationStats(ctx context.Context, query queries.GetOrganizationStatsQuery) (*OrganizationStats, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (o:Organization {tenant_id: $tenant_id})
		OPTIONAL MATCH (o)-[:PARENT_OF*]->(child:Organization)
		RETURN 
			COUNT(DISTINCT o) as total_organizations,
			COUNT(DISTINCT CASE WHEN o.is_active = true THEN o END) as active_organizations,
			COUNT(DISTINCT CASE WHEN o.unit_type = 'COMPANY' THEN o END) as companies,
			COUNT(DISTINCT CASE WHEN o.unit_type = 'DEPARTMENT' THEN o END) as departments,
			COUNT(DISTINCT CASE WHEN o.unit_type = 'TEAM' THEN o END) as teams,
			MAX(o.level) as max_depth,
			SUM(o.employee_count) as total_employees`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"tenant_id": query.TenantID.String(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			totalOrgs, _ := record.Get("total_organizations")
			activeOrgs, _ := record.Get("active_organizations")
			companies, _ := record.Get("companies")
			departments, _ := record.Get("departments")
			teams, _ := record.Get("teams")
			maxDepth, _ := record.Get("max_depth")
			totalEmployees, _ := record.Get("total_employees")

			return &OrganizationStats{
				TotalOrganizations:  int(totalOrgs.(int64)),
				ActiveOrganizations: int(activeOrgs.(int64)),
				Companies:           int(companies.(int64)),
				Departments:         int(departments.(int64)),
				Teams:               int(teams.(int64)),
				MaxDepth:            int(maxDepth.(int64)),
				TotalEmployees:      int(totalEmployees.(int64)),
			}, nil
		}

		return nil, fmt.Errorf("no stats found")
	})

	if err != nil {
		r.logger.Error("Failed to get organization stats", "error", err)
		return nil, fmt.Errorf("failed to get organization stats: %w", err)
	}

	stats := result.(*OrganizationStats)
	r.logger.Info("Organization stats retrieved successfully", "total", stats.TotalOrganizations)
	return stats, nil
}

// SearchOrganizations 搜索组织
func (r *Neo4jOrganizationQueryRepository) SearchOrganizations(ctx context.Context, query queries.SearchOrganizationsQuery) ([]*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	// 构建搜索查询
	whereClause := "WHERE o.tenant_id = $tenant_id"
	params := map[string]any{
		"tenant_id": query.TenantID.String(),
	}

	if query.Query != "" {
		whereClause += " AND (o.name CONTAINS $keyword OR o.description CONTAINS $keyword)"
		params["keyword"] = query.Query
	}

	if len(query.UnitTypes) > 0 {
		whereClause += " AND o.unit_type IN $unit_types"
		params["unit_types"] = query.UnitTypes
	}

	if len(query.Status) > 0 {
		whereClause += " AND o.status IN $status"
		params["status"] = query.Status
	}

	if query.Limit > 0 {
		params["limit"] = query.Limit
	} else {
		params["limit"] = 50 // 默认限制
	}

	cypher := fmt.Sprintf(`
		MATCH (o:Organization)
		%s
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
		OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
		RETURN o, 
			   p.id as parent_id,
			   COUNT(c) as child_count
		ORDER BY o.level, o.name
		LIMIT $limit`, whereClause)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to search organizations", "error", err, "query", query.Query)
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}

	organizations := result.([]*Organization)
	r.logger.Info("Organization search completed", "count", len(organizations), "query", query.Query)
	return organizations, nil
}

// GetOrganizationHierarchy 获取组织层级
func (r *Neo4jOrganizationQueryRepository) GetOrganizationHierarchy(ctx context.Context, targetID uuid.UUID, direction string, maxDepth int, tenantID uuid.UUID) ([]*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	var cypher string
	params := map[string]any{
		"target_id":  targetID.String(),
		"tenant_id":  tenantID.String(),
		"max_depth":  maxDepth,
	}

	switch direction {
	case "up", "ancestors":
		cypher = `
			MATCH path = (target:Organization {id: $target_id, tenant_id: $tenant_id})<-[:PARENT_OF*1..%d]-(o:Organization)
			OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
			OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
			RETURN o, 
				   p.id as parent_id,
				   COUNT(c) as child_count,
				   LENGTH(path) as depth
			ORDER BY LENGTH(path) DESC`
	case "down", "descendants":
		cypher = `
			MATCH path = (target:Organization {id: $target_id, tenant_id: $tenant_id})-[:PARENT_OF*1..%d]->(o:Organization)
			OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
			OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
			RETURN o, 
				   p.id as parent_id,
				   COUNT(c) as child_count,
				   LENGTH(path) as depth
			ORDER BY LENGTH(path), o.level`
	default:
		return nil, fmt.Errorf("invalid direction: %s", direction)
	}

	cypher = fmt.Sprintf(cypher, maxDepth)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to get organization hierarchy", "error", err, "target_id", targetID, "direction", direction)
		return nil, fmt.Errorf("failed to get organization hierarchy: %w", err)
	}

	organizations := result.([]*Organization)
	r.logger.Info("Organization hierarchy retrieved", "count", len(organizations), "target_id", targetID, "direction", direction)
	return organizations, nil
}

// GetOrganizationPath 获取组织路径
func (r *Neo4jOrganizationQueryRepository) GetOrganizationPath(ctx context.Context, fromID, toID uuid.UUID, tenantID uuid.UUID) ([]*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	cypher := `
		MATCH path = shortestPath((from:Organization {id: $from_id, tenant_id: $tenant_id})-[:PARENT_OF*]-(to:Organization {id: $to_id, tenant_id: $tenant_id}))
		UNWIND nodes(path) as o
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(o)
		OPTIONAL MATCH (o)-[:PARENT_OF]->(c:Organization)
		RETURN o, 
			   p.id as parent_id,
			   COUNT(c) as child_count
		ORDER BY o.level`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"from_id":   fromID.String(),
			"to_id":     toID.String(),
			"tenant_id": tenantID.String(),
		})
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to get organization path", "error", err, "from_id", fromID, "to_id", toID)
		return nil, fmt.Errorf("failed to get organization path: %w", err)
	}

	organizations := result.([]*Organization)
	r.logger.Info("Organization path retrieved", "count", len(organizations), "from_id", fromID, "to_id", toID)
	return organizations, nil
}

// GetSiblingOrganizations 获取兄弟组织
func (r *Neo4jOrganizationQueryRepository) GetSiblingOrganizations(ctx context.Context, unitID uuid.UUID, includeSelf bool, tenantID uuid.UUID) ([]*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	var whereClause string
	if includeSelf {
		whereClause = ""
	} else {
		whereClause = " AND sibling.id <> $unit_id"
	}

	cypher := fmt.Sprintf(`
		MATCH (unit:Organization {id: $unit_id, tenant_id: $tenant_id})
		MATCH (parent:Organization)-[:PARENT_OF]->(unit)
		MATCH (parent)-[:PARENT_OF]->(sibling:Organization)%s
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(sibling)
		OPTIONAL MATCH (sibling)-[:PARENT_OF]->(c:Organization)
		RETURN sibling as o, 
			   p.id as parent_id,
			   COUNT(c) as child_count
		ORDER BY sibling.name`, whereClause)

	params := map[string]any{
		"unit_id":   unitID.String(),
		"tenant_id": tenantID.String(),
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to get sibling organizations", "error", err, "unit_id", unitID)
		return nil, fmt.Errorf("failed to get sibling organizations: %w", err)
	}

	organizations := result.([]*Organization)
	r.logger.Info("Sibling organizations retrieved", "count", len(organizations), "unit_id", unitID)
	return organizations, nil
}

// GetChildOrganizations 获取子组织
func (r *Neo4jOrganizationQueryRepository) GetChildOrganizations(ctx context.Context, parentID uuid.UUID, tenantID uuid.UUID) ([]*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})-[:PARENT_OF]->(child:Organization)
		OPTIONAL MATCH (p:Organization)-[:PARENT_OF]->(child)
		OPTIONAL MATCH (child)-[:PARENT_OF]->(c:Organization)
		RETURN child as o, 
			   p.id as parent_id,
			   COUNT(c) as child_count
		ORDER BY child.level, child.name`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"parent_id": parentID.String(),
			"tenant_id": tenantID.String(),
		})
		if err != nil {
			return nil, err
		}

		var organizations []*Organization
		for result.Next(ctx) {
			record := result.Record()
			orgNode, _ := record.Get("o")
			parentID, _ := record.Get("parent_id")
			childCount, _ := record.Get("child_count")

			org, err := r.nodeToOrganization(orgNode.(neo4j.Node), parentID, childCount.(int64))
			if err != nil {
				r.logger.Warn("Failed to convert node to organization", "error", err)
				continue
			}
			organizations = append(organizations, org)
		}

		return organizations, nil
	})

	if err != nil {
		r.logger.Error("Failed to get child organizations", "error", err, "parent_id", parentID)
		return nil, fmt.Errorf("failed to get child organizations: %w", err)
	}

	organizations := result.([]*Organization)
	r.logger.Info("Child organizations retrieved", "count", len(organizations), "parent_id", parentID)
	return organizations, nil
}

// OrganizationExists 检查组织是否存在
func (r *Neo4jOrganizationQueryRepository) OrganizationExists(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (bool, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)

	cypher := `
		MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
		RETURN COUNT(o) > 0 as exists`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{
			"id":        id.String(),
			"tenant_id": tenantID.String(),
		})
		if err != nil {
			return false, err
		}

		if result.Next(ctx) {
			exists, _ := result.Record().Get("exists")
			return exists.(bool), nil
		}

		return false, nil
	})

	if err != nil {
		r.logger.Error("Failed to check organization existence", "error", err, "org_id", id)
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}

	exists := result.(bool)
	r.logger.Info("Organization existence checked", "org_id", id, "exists", exists)
	return exists, nil
}

// 辅助方法

// nodeToOrganization 将Neo4j节点转换为Organization对象
func (r *Neo4jOrganizationQueryRepository) nodeToOrganization(node neo4j.Node, parentID any, childCount int64) (*Organization, error) {
	props := node.Props

	// 解析ID
	idStr, ok := props["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid organization id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse organization id: %w", err)
	}

	// 解析TenantID
	tenantIDStr, ok := props["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid tenant id")
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant id: %w", err)
	}

	// 解析ParentUnitID
	var parentUnitID *uuid.UUID
	if parentID != nil && parentID != "" {
		if parentIDStr, ok := parentID.(string); ok {
			if pid, err := uuid.Parse(parentIDStr); err == nil {
				parentUnitID = &pid
			}
		}
	}

	// 解析Profile
	var profile map[string]interface{}
	if profileStr, ok := props["profile"].(string); ok {
		if err := json.Unmarshal([]byte(profileStr), &profile); err != nil {
			r.logger.Warn("Failed to unmarshal profile", "error", err)
			profile = make(map[string]interface{})
		}
	}

	// 解析时间字段
	createdAt, _ := time.Parse(time.RFC3339, props["created_at"].(string))
	updatedAt, _ := time.Parse(time.RFC3339, props["updated_at"].(string))

	// 获取Description
	var description *string
	if desc, ok := props["description"].(string); ok && desc != "" {
		description = &desc
	}

	org := &Organization{
		ID:            id,
		TenantID:      tenantID,
		UnitType:      props["unit_type"].(string),
		Name:          props["name"].(string),
		Description:   description,
		ParentUnitID:  parentUnitID,
		Status:        props["status"].(string),
		Profile:       profile,
		Level:         int(props["level"].(int64)),
		EmployeeCount: int(childCount), // 使用子组织数量作为员工数量
		IsActive:      props["is_active"].(bool),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return org, nil
}

// buildWhereClause 构建WHERE子句
func (r *Neo4jOrganizationQueryRepository) buildWhereClause(query queries.ListOrganizationsQuery) (string, map[string]any) {
	var conditions []string
	params := make(map[string]any)

	if query.UnitType != nil && *query.UnitType != "" {
		conditions = append(conditions, "o.unit_type = $unit_type")
		params["unit_type"] = *query.UnitType
	}

	if query.Status != nil && *query.Status != "" {
		conditions = append(conditions, "o.status = $status")
		params["status"] = *query.Status
	}

	if query.ParentUnitID != nil {
		conditions = append(conditions, "EXISTS { (parent:Organization {id: $parent_unit_id})-[:PARENT_OF]->(o) }")
		params["parent_unit_id"] = query.ParentUnitID.String()
	}

	if query.Search != nil && *query.Search != "" {
		conditions = append(conditions, "(o.name CONTAINS $search OR o.description CONTAINS $search)")
		params["search"] = *query.Search
	}

	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND "), params
	}

	return "", params
}

// getMaxDepthLimit 获取最大深度限制
func (r *Neo4jOrganizationQueryRepository) getMaxDepthLimit(maxDepth int) int {
	if maxDepth <= 0 {
		return 10 // 默认最大深度
	}
	if maxDepth > 20 {
		return 20 // 限制最大深度
	}
	return maxDepth
}