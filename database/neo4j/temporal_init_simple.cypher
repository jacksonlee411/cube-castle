// =============================================
// Neo4j时态图结构设计 - 简化版
// 负责层级计算、路径生成和复杂查询
// =============================================

// 1. 创建时态组织节点约束
CREATE CONSTRAINT temporal_org_unique IF NOT EXISTS
FOR (n:TemporalOrganization) 
REQUIRE (n.tenant_id, n.code, n.effective_date) IS UNIQUE;

// 2. 创建时态组织节点索引
CREATE INDEX temporal_org_current IF NOT EXISTS FOR (n:TemporalOrganization) ON (n.tenant_id, n.is_current);
CREATE INDEX temporal_org_date IF NOT EXISTS FOR (n:TemporalOrganization) ON (n.tenant_id, n.effective_date);
CREATE INDEX temporal_org_parent IF NOT EXISTS FOR (n:TemporalOrganization) ON (n.tenant_id, n.parent_code, n.effective_date);
CREATE INDEX temporal_org_hierarchy IF NOT EXISTS FOR (n:TemporalOrganization) ON (n.tenant_id, n.level, n.effective_date);

// 3. 清理现有的时态组织数据（如果存在）
MATCH (n:TemporalOrganization) DELETE n;

// 4. 验证索引和约束创建
SHOW CONSTRAINTS YIELD name, type WHERE name CONTAINS 'temporal';
SHOW INDEXES YIELD name, type WHERE name CONTAINS 'temporal';