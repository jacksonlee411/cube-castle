// =============================================
// Neo4j时态图结构设计
// 负责层级计算、路径生成和复杂查询
// 支持17级层级和无限历史保留
// =============================================

// 1. 创建时态组织节点约束
CREATE CONSTRAINT temporal_org_unique 
FOR (n:TemporalOrganization) 
REQUIRE (n.tenant_id, n.code, n.effective_date) IS UNIQUE;

CREATE CONSTRAINT temporal_org_code_date
FOR (n:TemporalOrganization)
REQUIRE (n.code, n.effective_date) IS UNIQUE;

// 2. 创建时态组织节点索引
CREATE INDEX temporal_org_current FOR (n:TemporalOrganization) ON (n.tenant_id, n.is_current) WHERE n.is_current = true;
CREATE INDEX temporal_org_date FOR (n:TemporalOrganization) ON (n.tenant_id, n.effective_date);
CREATE INDEX temporal_org_parent FOR (n:TemporalOrganization) ON (n.tenant_id, n.parent_code, n.effective_date);
CREATE INDEX temporal_org_hierarchy FOR (n:TemporalOrganization) ON (n.tenant_id, n.level, n.effective_date);

// 3. 时态组织节点结构
// 每个节点代表组织在特定日期的状态
(:TemporalOrganization {
  // 基础标识
  tenant_id: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
  code: "1000056",
  parent_code: "1000001",
  
  // 业务属性
  name: "技术研发部",
  unit_type: "DEPARTMENT",
  status: "ACTIVE",
  
  // 时态属性
  effective_date: date("2025-08-11"),
  end_date: null,
  is_current: true,
  change_reason: "组织架构调整",
  
  // Neo4j专属计算字段
  level: 2,
  path: "/1000001/1000056",
  hierarchy_codes: ["1000001", "1000056"],
  full_path_names: "高谷集团/技术研发部",
  
  // 层级统计（缓存优化）
  direct_children_count: 5,
  all_descendants_count: 23,
  max_descendant_level: 5,
  
  // 计算缓存
  path_last_calculated: datetime("2025-08-11T10:30:00Z"),
  hierarchy_hash: "abc123def456", // 用于快速检测层级变化
  
  // 审计字段
  created_at: datetime("2025-08-11T10:00:00Z"),
  updated_at: datetime("2025-08-11T10:30:00Z"),
  synced_from_pg: datetime("2025-08-11T10:30:15Z")
})

// 4. 时态层级关系设计
// 直接父子关系（同时间点）
(:TemporalOrganization)-[:PARENT_OF {
  effective_from: date("2025-08-11"),
  effective_to: null,
  relationship_level: 1,
  created_at: datetime("2025-08-11T10:30:00Z")
}]->(:TemporalOrganization)

// 祖先后代关系（缓存优化，支持快速查询）
(:TemporalOrganization)-[:ANCESTOR_OF {
  effective_from: date("2025-08-11"),
  effective_to: null,
  relationship_level: 3, // 相隔层级数
  path_through: ["1000001", "1000023"], // 中间路径
  created_at: datetime("2025-08-11T10:30:00Z")
}]->(:TemporalOrganization)

// 时态演进关系
(:TemporalOrganization)-[:EVOLVES_TO {
  change_type: "UPDATE", // CREATE, UPDATE, DELETE, MOVE
  change_fields: ["name", "status"],
  change_reason: "部门重组",
  created_at: datetime("2025-08-11T10:30:00Z")
}]->(:TemporalOrganization)

// 5. 时态层级计算函数
// 重新计算单个组织的层级路径
CALL apoc.custom.asProcedure(
  'temporal.calculateHierarchy',
  'WITH $tenant_id as tenant_id, $code as code, $as_of_date as as_of_date
   MATCH (org:TemporalOrganization {tenant_id: tenant_id, code: code})
   WHERE org.effective_date <= as_of_date 
     AND (org.end_date IS NULL OR org.end_date >= as_of_date)
   
   // 递归查找所有祖先
   OPTIONAL MATCH path = (org)<-[:PARENT_OF*1..17]-(ancestors:TemporalOrganization)
   WHERE ALL(r IN relationships(path) WHERE 
     r.effective_from <= as_of_date AND 
     (r.effective_to IS NULL OR r.effective_to >= as_of_date)
   )
   
   WITH org, 
        CASE WHEN ancestors IS NULL THEN [] ELSE collect(DISTINCT ancestors.code) END as ancestor_codes,
        CASE WHEN ancestors IS NULL THEN [] ELSE collect(DISTINCT ancestors.name) END as ancestor_names,
        CASE WHEN ancestors IS NULL THEN 1 ELSE length(path) + 1 END as calculated_level
   
   // 更新组织的层级信息
   SET org.level = calculated_level,
       org.hierarchy_codes = ancestor_codes + [org.code],
       org.path = "/" + reduce(path = "", code IN ancestor_codes + [org.code] | path + "/" + code),
       org.full_path_names = reduce(path = "", name IN ancestor_names + [org.name] | 
         CASE WHEN path = "" THEN name ELSE path + "/" + name END),
       org.path_last_calculated = datetime()
   
   RETURN org.code as updated_code, org.level as new_level, org.path as new_path',
  'READ',
  [['tenant_id','STRING'], ['code','STRING'], ['as_of_date','DATE']]
);

// 6. 批量重建层级关系
CALL apoc.custom.asProcedure(
  'temporal.rebuildHierarchy',
  'WITH $tenant_id as tenant_id, $as_of_date as as_of_date
   
   // 删除现有的祖先后代关系
   MATCH ()-[r:ANCESTOR_OF]->()
   WHERE r.effective_from <= as_of_date AND (r.effective_to IS NULL OR r.effective_to >= as_of_date)
   DELETE r
   
   // 重新创建祖先后代关系
   MATCH (descendant:TemporalOrganization {tenant_id: tenant_id})
   WHERE descendant.effective_date <= as_of_date 
     AND (descendant.end_date IS NULL OR descendant.end_date >= as_of_date)
   
   MATCH (ancestor:TemporalOrganization {tenant_id: tenant_id})
   WHERE ancestor.effective_date <= as_of_date 
     AND (ancestor.end_date IS NULL OR ancestor.end_date >= as_of_date)
     AND ancestor.code IN descendant.hierarchy_codes
     AND ancestor.code <> descendant.code
   
   CREATE (ancestor)-[:ANCESTOR_OF {
     effective_from: as_of_date,
     effective_to: null,
     relationship_level: size(filter(code IN descendant.hierarchy_codes WHERE code = ancestor.code)) - 
                        size(filter(code IN ancestor.hierarchy_codes WHERE code = ancestor.code)),
     created_at: datetime()
   }]->(descendant)
   
   RETURN count(*) as relationships_created',
  'WRITE',
  [['tenant_id','STRING'], ['as_of_date','DATE']]
);

// 7. 特定日期查询优化
CALL apoc.custom.asProcedure(
  'temporal.getOrganizationByDate',
  'WITH $tenant_id as tenant_id, $code as code, $as_of_date as as_of_date
   MATCH (org:TemporalOrganization {tenant_id: tenant_id, code: code})
   WHERE org.effective_date <= as_of_date 
     AND (org.end_date IS NULL OR org.end_date >= as_of_date)
   
   RETURN org.code as code,
          org.name as name,
          org.parent_code as parent_code,
          org.unit_type as unit_type,
          org.status as status,
          org.level as level,
          org.path as path,
          org.hierarchy_codes as hierarchy_codes,
          org.full_path_names as full_path_names,
          org.effective_date as effective_date,
          org.end_date as end_date,
          org.is_current as is_current
   
   ORDER BY org.effective_date DESC
   LIMIT 1',
  'READ',
  [['tenant_id','STRING'], ['code','STRING'], ['as_of_date','DATE']]
);

// 8. 层级树查询（支持17级深度）
CALL apoc.custom.asProcedure(
  'temporal.getHierarchyTree',
  'WITH $tenant_id as tenant_id, $root_code as root_code, $as_of_date as as_of_date, 
        coalesce($max_depth, 17) as max_depth
   
   MATCH (root:TemporalOrganization {tenant_id: tenant_id, code: root_code})
   WHERE root.effective_date <= as_of_date 
     AND (root.end_date IS NULL OR root.end_date >= as_of_date)
   
   MATCH path = (root)-[:PARENT_OF*0..max_depth]->(descendants:TemporalOrganization)
   WHERE ALL(r IN relationships(path) WHERE 
     r.effective_from <= as_of_date AND 
     (r.effective_to IS NULL OR r.effective_to >= as_of_date)
   )
   AND descendants.effective_date <= as_of_date 
   AND (descendants.end_date IS NULL OR descendants.end_date >= as_of_date)
   
   RETURN descendants.code as code,
          descendants.name as name,
          descendants.parent_code as parent_code,
          descendants.level as level,
          descendants.path as path,
          descendants.unit_type as unit_type,
          descendants.status as status,
          length(path) as depth_from_root
   
   ORDER BY descendants.level, descendants.path',
  'READ',
  [['tenant_id','STRING'], ['root_code','STRING'], ['as_of_date','DATE'], ['max_depth','LONG']]
);

// 9. 历史轨迹查询（无限历史保留）
CALL apoc.custom.asProcedure(
  'temporal.getOrganizationHistory',
  'WITH $tenant_id as tenant_id, $code as code
   MATCH (org:TemporalOrganization {tenant_id: tenant_id, code: code})
   
   RETURN org.code as code,
          org.name as name,
          org.parent_code as parent_code,
          org.unit_type as unit_type,
          org.status as status,
          org.level as level,
          org.path as path,
          org.effective_date as effective_date,
          org.end_date as end_date,
          org.is_current as is_current,
          org.change_reason as change_reason
   
   ORDER BY org.effective_date DESC',
  'READ',
  [['tenant_id','STRING'], ['code','STRING']]
);

// 10. 性能监控查询
CALL apoc.custom.asProcedure(
  'temporal.getPerformanceStats',
  'MATCH (org:TemporalOrganization)
   
   WITH count(org) as total_nodes,
        count(CASE WHEN org.is_current = true THEN 1 END) as current_nodes,
        count(CASE WHEN org.path_last_calculated < datetime() - duration("PT1H") THEN 1 END) as stale_paths,
        max(org.level) as max_level,
        avg(org.level) as avg_level
   
   MATCH ()-[r:PARENT_OF]->()
   WITH total_nodes, current_nodes, stale_paths, max_level, avg_level, count(r) as total_relationships
   
   MATCH ()-[r2:ANCESTOR_OF]->()
   WITH total_nodes, current_nodes, stale_paths, max_level, avg_level, total_relationships, count(r2) as ancestor_relationships
   
   RETURN {
     total_temporal_nodes: total_nodes,
     current_active_nodes: current_nodes,
     stale_path_nodes: stale_paths,
     max_hierarchy_level: max_level,
     avg_hierarchy_level: avg_level,
     direct_relationships: total_relationships,
     ancestor_relationships: ancestor_relationships,
     calculated_at: datetime()
   } as stats',
  'READ',
  []
);

// 11. 数据同步验证
CALL apoc.custom.asProcedure(
  'temporal.validateSyncIntegrity',
  'WITH $tenant_id as tenant_id
   
   // 检查是否有孤儿节点（parent_code指向不存在的节点）
   MATCH (org:TemporalOrganization {tenant_id: tenant_id})
   WHERE org.parent_code IS NOT NULL 
     AND org.is_current = true
   
   OPTIONAL MATCH (parent:TemporalOrganization {tenant_id: tenant_id, code: org.parent_code})
   WHERE parent.is_current = true
   
   WITH count(org) as total_with_parent,
        count(parent) as valid_parents,
        collect(CASE WHEN parent IS NULL THEN org.code END) as orphan_codes
   
   // 检查层级深度是否超限
   MATCH (deep_org:TemporalOrganization {tenant_id: tenant_id})
   WHERE deep_org.level > 17 AND deep_org.is_current = true
   
   WITH total_with_parent, valid_parents, orphan_codes,
        collect(deep_org.code) as deep_codes,
        count(deep_org) as deep_count
   
   RETURN {
     organizations_with_parent: total_with_parent,
     valid_parent_references: valid_parents,
     orphan_organizations: orphan_codes,
     organizations_too_deep: deep_codes,
     max_depth_violations: deep_count,
     integrity_score: CASE 
       WHEN total_with_parent = 0 THEN 1.0 
       ELSE toFloat(valid_parents) / total_with_parent 
     END,
     validated_at: datetime()
   } as integrity_report',
  'READ',
  [['tenant_id','STRING']]
);

// 使用示例和验证查询
/*
// 创建测试数据
CREATE (root:TemporalOrganization {
  tenant_id: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
  code: "1000001",
  parent_code: null,
  name: "高谷集团",
  unit_type: "COMPANY",
  status: "ACTIVE",
  effective_date: date("2025-01-01"),
  end_date: null,
  is_current: true,
  level: 1,
  path: "/1000001",
  hierarchy_codes: ["1000001"],
  full_path_names: "高谷集团"
});

CREATE (dept:TemporalOrganization {
  tenant_id: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
  code: "1000056",
  parent_code: "1000001",
  name: "技术研发部",
  unit_type: "DEPARTMENT", 
  status: "ACTIVE",
  effective_date: date("2025-08-11"),
  end_date: null,
  is_current: true,
  level: 2,
  path: "/1000001/1000056",
  hierarchy_codes: ["1000001", "1000056"],
  full_path_names: "高谷集团/技术研发部"
});

// 建立关系
MATCH (root {code: "1000001"}), (dept {code: "1000056"})
CREATE (root)-[:PARENT_OF {
  effective_from: date("2025-08-11"),
  effective_to: null,
  relationship_level: 1
}]->(dept);

// 测试查询
CALL temporal.getOrganizationByDate("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", "1000056", date("2025-08-11"));
CALL temporal.getHierarchyTree("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", "1000001", date("2025-08-11"), 5);
CALL temporal.getPerformanceStats();
*/