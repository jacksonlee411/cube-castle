#!/bin/bash

# Neo4j层级关系修复脚本
# 解决HAS_CHILD vs PARENT_OF关系不一致问题

set -e

echo "🔧 开始修复Neo4j层级关系问题..."

# 1. 检查当前关系状态
echo "📊 检查当前Neo4j关系状态..."
docker exec cube_castle_neo4j cypher-shell -u neo4j -p password << 'EOF'
MATCH ()-[r]-() 
RETURN DISTINCT type(r) AS relationship_types, count(r) AS count 
ORDER BY relationship_types;
EOF

# 2. 为所有现有的HAS_CHILD关系创建对应的PARENT_OF关系
echo "🔄 创建PARENT_OF关系..."
docker exec cube_castle_neo4j cypher-shell -u neo4j -p password << 'EOF'
// 为每个HAS_CHILD关系创建对应的PARENT_OF关系
MATCH (parent)-[hc:HAS_CHILD]->(child)
MERGE (child)-[:PARENT_OF]->(parent)
RETURN count(*) AS created_parent_relations;
EOF

# 3. 验证关系创建结果
echo "✅ 验证关系创建结果..."
docker exec cube_castle_neo4j cypher-shell -u neo4j -p password << 'EOF'
MATCH ()-[r]-() 
RETURN DISTINCT type(r) AS relationship_types, count(r) AS count 
ORDER BY relationship_types;
EOF

# 4. 测试层级查询是否正常工作
echo "🧪 测试层级查询..."
docker exec cube_castle_neo4j cypher-shell -u neo4j -p password << 'EOF'
// 测试使用PARENT_OF关系的层级查询
MATCH (org:OrganizationUnit {code: '1000056'})
OPTIONAL MATCH path = (org)-[:PARENT_OF*1..5]->(ancestors)
RETURN 
  org.code as org_code,
  org.name as org_name,
  length(path) + 1 as calculated_level,
  [node in nodes(path) | node.code] as hierarchy_path
LIMIT 5;
EOF

echo "✅ Neo4j层级关系修复完成！"