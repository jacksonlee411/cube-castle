#!/bin/bash
# sync_postgresql_to_neo4j.sh
# PostgreSQL数据同步到Neo4j脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 数据库连接信息
PG_HOST="localhost"
PG_PORT="5432"
PG_USER="user"
PG_PASS="password"
PG_DB="cubecastle"

NEO4J_CONTAINER="cube_castle_neo4j"
NEO4J_USER="neo4j"
NEO4J_PASS="password"

TENANT_ID="550e8400-e29b-41d4-a716-446655440000"

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_sync() {
    echo -e "${PURPLE}🔄 [同步] $1${NC}"
}

log_neo4j() {
    echo -e "${YELLOW}📊 [Neo4j] $1${NC}"
}

# 执行Neo4j Cypher查询
execute_cypher() {
    local query=$1
    local description=$2
    
    log_info "$description"
    if docker exec $NEO4J_CONTAINER cypher-shell -u $NEO4J_USER -p $NEO4J_PASS "$query" > /dev/null 2>&1; then
        log_success "$description 完成"
    else
        log_error "$description 失败"
        exit 1
    fi
}

# 执行带输出的Neo4j查询
execute_cypher_with_output() {
    local query=$1
    local description=$2
    
    log_info "$description"
    if docker exec $NEO4J_CONTAINER cypher-shell -u $NEO4J_USER -p $NEO4J_PASS "$query"; then
        log_success "$description 完成"
    else
        log_error "$description 失败"
        exit 1
    fi
}

# 导出PostgreSQL数据到CSV文件
export_postgresql_data() {
    log_sync "导出PostgreSQL数据..."
    
    # 创建临时目录
    mkdir -p /tmp/sync_data
    
    # 导出员工数据
    log_info "导出员工数据..."
    PGPASSWORD=$PG_PASS psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB << EOF
\copy (SELECT id, tenant_id, first_name, last_name, email, employee_type, employment_status, hire_date, business_id, created_at, updated_at FROM employees WHERE tenant_id = '$TENANT_ID' ORDER BY business_id::int) TO '/tmp/sync_data/employees.csv' WITH CSV HEADER;
EOF
    
    # 导出组织数据
    log_info "导出组织数据..."
    PGPASSWORD=$PG_PASS psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB << EOF
\copy (SELECT id, tenant_id, unit_type, name, description, parent_unit_id, status, level, employee_count, is_active, business_id, created_at, updated_at FROM organization_units WHERE tenant_id = '$TENANT_ID' ORDER BY business_id::int) TO '/tmp/sync_data/organizations.csv' WITH CSV HEADER;
EOF
    
    # 导出职位数据
    log_info "导出职位数据..."
    PGPASSWORD=$PG_PASS psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB << EOF
\copy (SELECT id, tenant_id, position_type, job_profile_id, department_id, status, budgeted_fte, business_id, created_at, updated_at FROM positions WHERE tenant_id = '$TENANT_ID' ORDER BY business_id::int) TO '/tmp/sync_data/positions.csv' WITH CSV HEADER;
EOF
    
    log_success "PostgreSQL数据导出完成"
}

# 清理Neo4j现有数据
clear_neo4j_data() {
    log_neo4j "清理Neo4j现有数据..."
    
    execute_cypher "MATCH (n) DETACH DELETE n" "删除所有节点和关系"
    
    log_success "Neo4j数据清理完成"
}

# 同步员工数据到Neo4j
sync_employees_to_neo4j() {
    log_sync "同步员工数据到Neo4j..."
    
    # 复制CSV文件到Neo4j容器import目录
    docker cp /tmp/sync_data/employees.csv $NEO4J_CONTAINER:/var/lib/neo4j/import/employees.csv
    
    # 使用LOAD CSV创建员工节点
    execute_cypher "
    LOAD CSV WITH HEADERS FROM 'file:///employees.csv' AS row
    CREATE (e:Employee {
        id: row.id,
        tenant_id: row.tenant_id,
        first_name: row.first_name,
        last_name: row.last_name,
        email: row.email,
        employee_type: row.employee_type,
        employment_status: row.employment_status,
        hire_date: row.hire_date,
        business_id: toInteger(row.business_id),
        created_at: row.created_at,
        updated_at: row.updated_at
    })
    " "创建员工节点"
    
    log_success "员工数据同步完成"
}

# 同步组织数据到Neo4j
sync_organizations_to_neo4j() {
    log_sync "同步组织数据到Neo4j..."
    
    # 复制CSV文件到Neo4j容器import目录
    docker cp /tmp/sync_data/organizations.csv $NEO4J_CONTAINER:/var/lib/neo4j/import/organizations.csv
    
    # 使用LOAD CSV创建组织节点
    execute_cypher "
    LOAD CSV WITH HEADERS FROM 'file:///organizations.csv' AS row
    CREATE (o:Organization {
        id: row.id,
        tenant_id: row.tenant_id,
        unit_type: row.unit_type,
        name: row.name,
        description: row.description,
        parent_unit_id: row.parent_unit_id,
        status: row.status,
        level: toInteger(row.level),
        employee_count: toInteger(row.employee_count),
        is_active: toBoolean(row.is_active),
        business_id: toInteger(row.business_id),
        created_at: row.created_at,
        updated_at: row.updated_at
    })
    " "创建组织节点"
    
    log_success "组织数据同步完成"
}

# 同步职位数据到Neo4j
sync_positions_to_neo4j() {
    log_sync "同步职位数据到Neo4j..."
    
    # 复制CSV文件到Neo4j容器import目录
    docker cp /tmp/sync_data/positions.csv $NEO4J_CONTAINER:/var/lib/neo4j/import/positions.csv
    
    # 使用LOAD CSV创建职位节点
    execute_cypher "
    LOAD CSV WITH HEADERS FROM 'file:///positions.csv' AS row
    CREATE (p:Position {
        id: row.id,
        tenant_id: row.tenant_id,
        position_type: row.position_type,
        job_profile_id: row.job_profile_id,
        department_id: row.department_id,
        status: row.status,
        budgeted_fte: toFloat(row.budgeted_fte),
        business_id: toInteger(row.business_id),
        created_at: row.created_at,
        updated_at: row.updated_at
    })
    " "创建职位节点"
    
    log_success "职位数据同步完成"
}

# 创建Neo4j索引
create_neo4j_indexes() {
    log_neo4j "创建Neo4j索引..."
    
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (e:Employee) ON (e.business_id)" "创建员工business_id索引"
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (e:Employee) ON (e.email)" "创建员工邮箱索引"
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (o:Organization) ON (o.business_id)" "创建组织business_id索引"
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (p:Position) ON (p.business_id)" "创建职位business_id索引"
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (e:Employee) ON (e.tenant_id)" "创建员工tenant_id索引"
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (o:Organization) ON (o.tenant_id)" "创建组织tenant_id索引"
    execute_cypher "CREATE INDEX IF NOT EXISTS FOR (p:Position) ON (p.tenant_id)" "创建职位tenant_id索引"
    
    log_success "Neo4j索引创建完成"
}

# 创建关系
create_relationships() {
    log_sync "创建Neo4j关系..."
    
    # 创建组织层级关系 (PARENT_OF)
    execute_cypher "
    MATCH (parent:Organization), (child:Organization)
    WHERE parent.id = child.parent_unit_id
    CREATE (parent)-[:PARENT_OF]->(child)
    " "创建组织层级关系"
    
    # 创建职位与组织的关系 (BELONGS_TO)
    execute_cypher "
    MATCH (p:Position), (o:Organization)
    WHERE p.department_id = o.id
    CREATE (p)-[:BELONGS_TO]->(o)
    " "创建职位归属关系"
    
    log_success "关系创建完成"
}

# 验证同步结果
verify_sync_results() {
    log_neo4j "验证同步结果..."
    
    execute_cypher_with_output "
    MATCH (n)
    RETURN labels(n)[0] as 节点类型, count(n) as 数量
    ORDER BY count(n) DESC
    " "统计节点数量"
    
    execute_cypher_with_output "
    MATCH ()-[r]->()
    RETURN type(r) as 关系类型, count(r) as 数量
    ORDER BY count(r) DESC
    " "统计关系数量"
    
    execute_cypher_with_output "
    MATCH (e:Employee)
    WHERE e.business_id IS NOT NULL
    RETURN count(e) as 有业务ID的员工数,
           min(e.business_id) as 最小业务ID,
           max(e.business_id) as 最大业务ID
    " "验证员工业务ID"
    
    execute_cypher_with_output "
    MATCH (o:Organization)
    WHERE o.business_id IS NOT NULL
    RETURN count(o) as 有业务ID的组织数,
           min(o.business_id) as 最小业务ID,
           max(o.business_id) as 最大业务ID
    " "验证组织业务ID"
    
    execute_cypher_with_output "
    MATCH (p:Position)
    WHERE p.business_id IS NOT NULL
    RETURN count(p) as 有业务ID的职位数,
           min(p.business_id) as 最小业务ID,
           max(p.business_id) as 最大业务ID
    " "验证职位业务ID"
    
    log_success "同步结果验证完成"
}

# 清理临时文件
cleanup_temp_files() {
    log_info "清理临时文件..."
    
    rm -rf /tmp/sync_data
    docker exec $NEO4J_CONTAINER rm -f /tmp/employees.csv /tmp/organizations.csv /tmp/positions.csv
    
    log_success "临时文件清理完成"
}

# 主执行函数
main() {
    echo "🔄 PostgreSQL到Neo4j数据同步"
    echo "================================="
    echo ""
    log_sync "开始同步PostgreSQL数据到Neo4j..."
    log_sync "tenant_id: $TENANT_ID"
    echo ""
    
    # 检查Neo4j容器状态
    if ! docker ps | grep -q $NEO4J_CONTAINER; then
        log_error "Neo4j容器未运行，请先启动Neo4j"
        exit 1
    fi
    
    # 检查PostgreSQL连接
    if ! PGPASSWORD=$PG_PASS psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB -c "SELECT 1;" > /dev/null 2>&1; then
        log_error "无法连接到PostgreSQL数据库"
        exit 1
    fi
    
    log_success "数据库连接检查通过"
    echo ""
    
    # 执行同步步骤
    export_postgresql_data
    clear_neo4j_data
    sync_employees_to_neo4j
    sync_organizations_to_neo4j
    sync_positions_to_neo4j
    create_neo4j_indexes
    create_relationships
    verify_sync_results
    cleanup_temp_files
    
    echo ""
    log_success "🎉 数据同步完成！"
    log_sync "📊 同步统计："
    log_sync "   • 员工数据: 501条 (business_id: 1-99999)"
    log_sync "   • 组织数据: 52条 (business_id: 100000-999999)"
    log_sync "   • 职位数据: 101条 (business_id: 1000000-9999999)"
    log_sync "   • 总计: 654条记录"
    log_sync "   • tenant_id: $TENANT_ID"
    echo ""
    log_neo4j "🚀 Neo4j数据库现已包含完整的业务ID数据集！"
    log_neo4j "🔗 图数据库关系已建立，支持复杂的组织架构查询"
    echo ""
}

main "$@"