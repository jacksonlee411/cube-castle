package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"cube-castle/cmd/hrms-server/query/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// 高级统计查询 - 利用PostgreSQL聚合优化
func (r *PostgreSQLRepository) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*model.OrganizationStats, error) {
	start := time.Now()

	// 使用单个复杂查询获取所有统计信息
	query := `
        WITH status_stats AS (
            SELECT 
                COUNT(*) FILTER (WHERE status <> 'DELETED' AND is_current = true)::int as total_count,
                COUNT(*) FILTER (WHERE status = 'ACTIVE' AND is_current = true)::int as active_count,
                COUNT(*) FILTER (WHERE status = 'INACTIVE' AND is_current = true)::int as inactive_count,
                COUNT(*) FILTER (WHERE status = 'PLANNED' AND is_current = true)::int as planned_count,
                COUNT(*) FILTER (WHERE status = 'DELETED')::int as deleted_count
            FROM organization_units WHERE tenant_id = $1
        ),
        type_stats AS (
            SELECT unit_type, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
            GROUP BY unit_type
        ),
        status_detail_stats AS (
            SELECT status, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
            GROUP BY status
        ),
        level_stats AS (
            SELECT level, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
            GROUP BY level
        ),
        temporal_stats AS (
            SELECT 
                COUNT(*) as total_versions,
                COUNT(DISTINCT code) as unique_orgs,
                COALESCE(MIN(effective_date), DATE '1970-01-01') as oldest_date,
                COALESCE(MAX(effective_date), DATE '1970-01-01') as newest_date
            FROM organization_units WHERE tenant_id = $1 AND status <> 'DELETED'
        )
		SELECT 
			s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
			ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date,
			COALESCE(json_agg(DISTINCT jsonb_build_object('unitType', t.unit_type, 'count', t.count)) FILTER (WHERE t.unit_type IS NOT NULL), '[]'),
			COALESCE(json_agg(DISTINCT jsonb_build_object('status', sd.status, 'count', sd.count)) FILTER (WHERE sd.status IS NOT NULL), '[]'),
			COALESCE(json_agg(DISTINCT jsonb_build_object('level', l.level, 'count', l.count)) FILTER (WHERE l.level IS NOT NULL), '[]')
		FROM status_stats s
		CROSS JOIN temporal_stats ts
		LEFT JOIN type_stats t ON true
		LEFT JOIN status_detail_stats sd ON true
		LEFT JOIN level_stats l ON true
		GROUP BY s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
		         ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date`

	row := r.db.QueryRowContext(ctx, query, tenantID.String())

	var stats model.OrganizationStats
	var totalVersions, uniqueOrgs int
	var oldestDate, newestDate time.Time
	var typeStatsJSON, statusStatsJSON, levelStatsJSON string

	err := row.Scan(
		&stats.TotalCountField, &stats.ActiveCountField, &stats.InactiveCountField,
		&stats.PlannedCountField, &stats.DeletedCountField,
		&totalVersions, &uniqueOrgs, &oldestDate, &newestDate,
		&typeStatsJSON, &statusStatsJSON, &levelStatsJSON,
	)
	if err != nil {
		r.logger.Errorf("统计查询失败: %v", err)
		return nil, err
	}

	// 解析JSON统计数据
	var typeStats []model.TypeCount
	if typeStatsJSON != "" {
		if err := json.Unmarshal([]byte(typeStatsJSON), &typeStats); err != nil {
			r.logger.Warnf("解析typeStats失败: %v", err)
		}
	}
	stats.ByTypeField = typeStats

	var statusStats []model.StatusCount
	if statusStatsJSON != "" {
		if err := json.Unmarshal([]byte(statusStatsJSON), &statusStats); err != nil {
			r.logger.Warnf("解析statusStats失败: %v", err)
		}
	}
	stats.ByStatusField = statusStats

	var levelStats []model.LevelCount
	if levelStatsJSON != "" {
		if err := json.Unmarshal([]byte(levelStatsJSON), &levelStats); err != nil {
			r.logger.Warnf("解析levelStats失败: %v", err)
		}
	}
	stats.ByLevelField = levelStats

	// 时态统计
	avgPerOrg := 0.0
	if uniqueOrgs > 0 {
		avgPerOrg = float64(totalVersions) / float64(uniqueOrgs)
	}

	stats.TemporalStatsField = model.TemporalStats{
		TotalVersionsField:         totalVersions,
		AverageVersionsPerOrgField: avgPerOrg,
		OldestEffectiveDateField:   oldestDate.Format("2006-01-02"),
		NewestEffectiveDateField:   newestDate.Format("2006-01-02"),
	}

	duration := time.Since(start)
	r.logger.Infof("统计查询完成，耗时: %v", duration)

	return &stats, nil
}

// 高级层级结构查询 - 严格遵循API规范v4.2.1
func (r *PostgreSQLRepository) GetOrganizationHierarchy(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationHierarchyData, error) {
	start := time.Now()

	// 使用PostgreSQL递归CTE查询完整层级信息
	query := `
        WITH RECURSIVE hierarchy_info AS (
            -- 获取目标组织
            SELECT
                code,
                name,
                level,
                parent_code,
                1 AS hierarchy_depth
            FROM organization_units
            WHERE tenant_id = $1
              AND code = $2
              AND is_current = true
              AND status <> 'DELETED'

            UNION ALL

            -- 递归获取父级信息
            SELECT
                o.code,
                o.name,
                o.level,
                o.parent_code,
                h.hierarchy_depth + 1
            FROM organization_units o
            INNER JOIN hierarchy_info h ON o.code = h.parent_code
            WHERE o.tenant_id = $1
              AND o.is_current = true
              AND o.status <> 'DELETED'
        ),
        aggregated_paths AS (
            SELECT
                '/' || string_agg(code, '/' ORDER BY hierarchy_depth DESC) AS full_code_path,
                '/' || string_agg(name, '/' ORDER BY hierarchy_depth DESC) AS full_name_path,
                COALESCE(
                    array_agg(code ORDER BY hierarchy_depth DESC) FILTER (WHERE hierarchy_depth > 1),
                    ARRAY[]::text[]
                ) AS parent_chain
            FROM hierarchy_info
        ),
        target_info AS (
            SELECT *
            FROM hierarchy_info
            WHERE code = $2
            LIMIT 1
        ),
        children_count AS (
            SELECT COUNT(*) AS count
            FROM organization_units
            WHERE tenant_id = $1
              AND parent_code = $2
              AND is_current = true
              AND status <> 'DELETED'
        )
        SELECT
            t.code,
            t.name,
            t.level,
            t.hierarchy_depth,
            ap.full_code_path,
            ap.full_name_path,
            ap.parent_chain,
            c.count AS children_count,
            (t.parent_code IS NULL) AS is_root,
            (c.count = 0) AS is_leaf
        FROM target_info t
        CROSS JOIN aggregated_paths ap
        CROSS JOIN children_count c
        LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var hierarchy model.OrganizationHierarchyData
	var parentChain []string

	err := row.Scan(
		&hierarchy.CodeField,
		&hierarchy.NameField,
		&hierarchy.LevelField,
		&hierarchy.HierarchyDepthField,
		&hierarchy.CodePathField,
		&hierarchy.NamePathField,
		pq.Array(&parentChain),
		&hierarchy.ChildrenCountField,
		&hierarchy.IsRootField,
		&hierarchy.IsLeafField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Errorf("层级结构查询失败: %v", err)
		return nil, err
	}

	hierarchy.ParentChainField = parentChain

	duration := time.Since(start)
	r.logger.Infof("层级结构查询完成，耗时: %v", duration)

	return &hierarchy, nil
}

// 组织子树查询 - 严格遵循API规范v4.2.1
func (r *PostgreSQLRepository) GetOrganizationSubtree(ctx context.Context, tenantID uuid.UUID, code string, maxDepth int) (*model.OrganizationHierarchyData, error) {
	start := time.Now()

	// 使用PostgreSQL递归CTE查询子树结构，限制深度
	query := `
        WITH RECURSIVE subtree AS (
            -- 根节点
            SELECT 
                code, name, level, 
                COALESCE(hierarchy_depth, level) as hierarchy_depth,
                COALESCE(code_path, '/' || code) as code_path,
                COALESCE(name_path, '/' || name) as name_path,
                parent_code,
                0 as depth_from_root
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
            
            UNION ALL
            
            -- 递归查询子节点
            SELECT 
                o.code, o.name, o.level,
                o.hierarchy_depth, o.code_path, o.name_path, o.parent_code,
                s.depth_from_root + 1
            FROM organization_units o
            INNER JOIN subtree s ON o.parent_code = s.code
            WHERE o.tenant_id = $1 AND o.is_current = true AND o.status <> 'DELETED'
              AND s.depth_from_root < $3
        )
		SELECT code, name, level, hierarchy_depth, code_path, name_path, parent_code
		FROM subtree 
		ORDER BY level, code`

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code, maxDepth)
	if err != nil {
		r.logger.Errorf("子树查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	// 构建树形结构
	nodeMap := make(map[string]*model.OrganizationSubtreeData)
	var root *model.OrganizationSubtreeData

	for rows.Next() {
		node := &model.OrganizationSubtreeData{}
		var parentCode *string

		err := rows.Scan(
			&node.CodeField, &node.NameField, &node.LevelField, &node.HierarchyDepthField,
			&node.CodePathField, &node.NamePathField, &parentCode,
		)
		if err != nil {
			r.logger.Errorf("扫描子树数据失败: %v", err)
			return nil, err
		}

		node.ChildrenField = []model.OrganizationSubtreeData{}
		node.IsRootField = node.CodeField == code
		node.ParentChainField = buildParentChain(node.CodePathField)
		nodeMap[node.CodeField] = node

		if node.CodeField == code {
			root = node
		}
	}

	// 构建父子关系
	for _, node := range nodeMap {
		if root != nil && node.CodeField != code {
			for _, parent := range nodeMap {
				if node.CodeField == parent.CodeField {
					continue
				}
				if node.CodePathField == nil || parent.CodePathField == nil {
					continue
				}
				nodePath := *node.CodePathField
				parentPath := *parent.CodePathField
				if strings.HasPrefix(nodePath, parentPath+"/") {
					parentDepth := strings.Count(parentPath, "/")
					nodeDepth := strings.Count(nodePath, "/")
					if nodeDepth == parentDepth+1 {
						parent.ChildrenField = append(parent.ChildrenField, *node)
						break
					}
				}
			}
		}
	}

	duration := time.Since(start)
	r.logger.Infof("子树查询完成，返回 %d 节点，耗时: %v", len(nodeMap), duration)

	converted := convertSubtreeToHierarchy(root)
	return converted, nil
}

func buildParentChain(codePath *string) []string {
	if codePath == nil {
		return []string{}
	}
	trimmed := strings.Trim(*codePath, "/")
	if trimmed == "" {
		return []string{}
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) <= 1 {
		return []string{}
	}
	return parts[:len(parts)-1]
}

func convertSubtreeToHierarchy(node *model.OrganizationSubtreeData) *model.OrganizationHierarchyData {
	if node == nil {
		return nil
	}

	children := make([]model.OrganizationHierarchyData, 0, len(node.ChildrenField))
	for i := range node.ChildrenField {
		child := convertSubtreeToHierarchy(&node.ChildrenField[i])
		if child != nil {
			children = append(children, *child)
		}
	}

	isLeaf := len(children) == 0
	hybrid := &model.OrganizationHierarchyData{
		CodeField:           node.CodeField,
		NameField:           node.NameField,
		LevelField:          node.LevelField,
		HierarchyDepthField: node.HierarchyDepthField,
		CodePathField:       node.CodePathField,
		NamePathField:       node.NamePathField,
		ParentChainField:    node.ParentChainField,
		ChildrenCountField:  len(children),
		IsRootField:         node.IsRootField,
		IsLeafField:         isLeaf,
		ChildrenField:       children,
	}

	return hybrid
}
