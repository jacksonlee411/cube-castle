#!/bin/bash

# 业务ID系统数据库迁移执行脚本
# 文件: execute_business_id_migration.sh
# 日期: 2025-08-04
# 描述: 执行PostgreSQL和Neo4j的业务ID迁移

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 配置变量
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-cubecastle}"
POSTGRES_USER="${POSTGRES_USER:-user}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"

NEO4J_HOST="${NEO4J_HOST:-localhost}"
NEO4J_PORT="${NEO4J_PORT:-7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-password}"

MIGRATION_DIR="$(dirname "$0")"
LOG_DIR="${MIGRATION_DIR}/logs"
BACKUP_DIR="${MIGRATION_DIR}/backups"

# 创建必要的目录
mkdir -p "$LOG_DIR" "$BACKUP_DIR"

# 时间戳
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')

# =============================================
# 预检查函数
# =============================================

check_prerequisites() {
    log_info "开始预检查..."
    
    # 检查PostgreSQL连接
    if ! psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT version();" > /dev/null 2>&1; then
        log_error "无法连接到PostgreSQL数据库"
        exit 1
    fi
    log_success "PostgreSQL连接正常"
    
    # 检查Neo4j连接
    if command -v cypher-shell >/dev/null 2>&1; then
        if ! cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" "RETURN 'connection test' as test" > /dev/null 2>&1; then
            log_error "无法连接到Neo4j数据库"
            exit 1
        fi
        log_success "Neo4j连接正常"
    else
        log_warning "cypher-shell未安装，将跳过Neo4j迁移"
    fi
    
    # 检查磁盘空间
    AVAILABLE_SPACE=$(df "$BACKUP_DIR" | tail -1 | awk '{print $4}')
    if [ "$AVAILABLE_SPACE" -lt 1048576 ]; then  # 1GB in KB
        log_warning "磁盘可用空间不足1GB，请确保有足够空间进行备份"
    fi
    
    log_success "预检查完成"
}

# =============================================
# 备份函数
# =============================================

backup_postgresql() {
    log_info "开始PostgreSQL数据备份..."
    
    local backup_file="${BACKUP_DIR}/postgresql_backup_${TIMESTAMP}.sql"
    
    # 备份完整数据库
    if pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        --verbose --no-owner --no-privileges > "$backup_file" 2>"${LOG_DIR}/postgresql_backup_${TIMESTAMP}.log"; then
        log_success "PostgreSQL备份完成: $backup_file"
        
        # 压缩备份文件
        gzip "$backup_file"
        log_success "备份文件已压缩: ${backup_file}.gz"
    else
        log_error "PostgreSQL备份失败"
        return 1
    fi
}

backup_neo4j() {
    log_info "开始Neo4j数据备份..."
    
    if ! command -v cypher-shell >/dev/null 2>&1; then
        log_warning "cypher-shell未安装，跳过Neo4j备份"
        return 0
    fi
    
    local backup_file="${BACKUP_DIR}/neo4j_backup_${TIMESTAMP}.json"
    
    # 使用APOC导出数据
    if cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
        "CALL apoc.export.json.all('file:///${backup_file}', {})" > "${LOG_DIR}/neo4j_backup_${TIMESTAMP}.log" 2>&1; then
        log_success "Neo4j备份完成: $backup_file"
    else
        log_warning "Neo4j备份失败，但继续执行迁移"
    fi
}

# =============================================
# 迁移执行函数
# =============================================

execute_postgresql_migration() {
    log_info "开始执行PostgreSQL迁移..."
    
    local migration_file="${MIGRATION_DIR}/../go-app/deployments/migrations/004_business_id_migration.sql"
    local log_file="${LOG_DIR}/postgresql_migration_${TIMESTAMP}.log"
    
    if [ ! -f "$migration_file" ]; then
        log_error "迁移文件不存在: $migration_file"
        return 1
    fi
    
    # 设置PostgreSQL密码环境变量
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # 执行迁移
    if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -v ON_ERROR_STOP=1 -f "$migration_file" > "$log_file" 2>&1; then
        log_success "PostgreSQL迁移执行完成"
        
        # 显示迁移结果
        log_info "验证迁移结果..."
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
            -c "SELECT * FROM business_id_validation_report;" | tee -a "$log_file"
        
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
            -c "SELECT * FROM business_id_format_validation WHERE count > 0;" | tee -a "$log_file"
        
        log_success "PostgreSQL迁移验证完成"
    else
        log_error "PostgreSQL迁移执行失败，请查看日志: $log_file"
        return 1
    fi
}

execute_neo4j_migration() {
    log_info "开始执行Neo4j迁移..."
    
    if ! command -v cypher-shell >/dev/null 2>&1; then
        log_warning "cypher-shell未安装，跳过Neo4j迁移"
        return 0
    fi
    
    local migration_file="${MIGRATION_DIR}/neo4j_business_id_migration.cypher"
    local log_file="${LOG_DIR}/neo4j_migration_${TIMESTAMP}.log"
    
    if [ ! -f "$migration_file" ]; then
        log_error "Neo4j迁移文件不存在: $migration_file"
        return 1
    fi
    
    # 执行Neo4j迁移
    if cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
        --file "$migration_file" > "$log_file" 2>&1; then
        log_success "Neo4j迁移执行完成"
        
        # 验证迁移结果
        log_info "验证Neo4j迁移结果..."
        cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "MATCH (e:Employee) RETURN count(e) as total_employees, count(e.business_id) as employees_with_business_id; 
             MATCH (o:Organization) RETURN count(o) as total_organizations, count(o.business_id) as organizations_with_business_id;" \
            | tee -a "$log_file"
        
        log_success "Neo4j迁移验证完成"
    else
        log_error "Neo4j迁移执行失败，请查看日志: $log_file"
        return 1
    fi
}

# =============================================
# 数据一致性验证函数
# =============================================

verify_data_consistency() {
    log_info "开始数据一致性验证..."
    
    local verification_log="${LOG_DIR}/data_consistency_${TIMESTAMP}.log"
    
    # PostgreSQL数据统计
    log_info "获取PostgreSQL数据统计..."
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "SELECT 'PostgreSQL' as source, 'employees' as table_name, COUNT(*) as total_count, COUNT(business_id) as business_id_count FROM corehr.employees 
            UNION ALL
            SELECT 'PostgreSQL', 'organizations', COUNT(*), COUNT(business_id) FROM corehr.organizations;" \
        >> "$verification_log"
    
    # Neo4j数据统计 (如果可用)
    if command -v cypher-shell >/dev/null 2>&1; then
        log_info "获取Neo4j数据统计..."
        cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "MATCH (e:Employee) 
             WITH count(e) as emp_total, count(e.business_id) as emp_with_bid
             MATCH (o:Organization)
             WITH emp_total, emp_with_bid, count(o) as org_total, count(o.business_id) as org_with_bid
             RETURN 'Neo4j' as source, 'employees' as table_name, emp_total as total_count, emp_with_bid as business_id_count
             UNION ALL
             RETURN 'Neo4j' as source, 'organizations' as table_name, org_total as total_count, org_with_bid as business_id_count" \
            >> "$verification_log"
    fi
    
    log_success "数据一致性验证完成，结果保存在: $verification_log"
    cat "$verification_log"
}

# =============================================
# 性能测试函数
# =============================================

run_performance_tests() {
    log_info "开始性能基准测试..."
    
    local perf_log="${LOG_DIR}/performance_test_${TIMESTAMP}.log"
    
    # PostgreSQL性能测试
    log_info "执行PostgreSQL性能测试..."
    echo "=== PostgreSQL Performance Test ===" >> "$perf_log"
    
    # 业务ID查询性能
    echo "Testing business_id query performance:" >> "$perf_log"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "EXPLAIN ANALYZE SELECT * FROM corehr.employees WHERE business_id = '1';" >> "$perf_log" 2>&1
    
    # UUID查询性能对比
    echo "Testing UUID query performance for comparison:" >> "$perf_log"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "EXPLAIN ANALYZE SELECT * FROM corehr.employees WHERE id = (SELECT id FROM corehr.employees LIMIT 1);" >> "$perf_log" 2>&1
    
    # Neo4j性能测试 (如果可用)
    if command -v cypher-shell >/dev/null 2>&1; then
        log_info "执行Neo4j性能测试..."
        echo "=== Neo4j Performance Test ===" >> "$perf_log"
        
        cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "PROFILE MATCH (e:Employee {business_id: '1'}) RETURN e.first_name, e.last_name;" >> "$perf_log" 2>&1
    fi
    
    log_success "性能测试完成，结果保存在: $perf_log"
}

# =============================================
# 回滚函数
# =============================================

rollback_migration() {
    log_warning "开始回滚迁移..."
    
    local rollback_log="${LOG_DIR}/rollback_${TIMESTAMP}.log"
    
    # PostgreSQL回滚
    log_info "回滚PostgreSQL更改..."
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "DROP VIEW IF EXISTS business_id_validation_report CASCADE;
            DROP VIEW IF EXISTS business_id_format_validation CASCADE;
            DROP TRIGGER IF EXISTS trigger_employees_business_id ON corehr.employees;
            DROP TRIGGER IF EXISTS trigger_organizations_business_id ON corehr.organizations;
            DROP TRIGGER IF EXISTS trigger_public_employees_business_id ON public.employees;
            DROP TRIGGER IF EXISTS trigger_organization_units_business_id ON public.organization_units;
            DROP FUNCTION IF EXISTS auto_generate_business_id() CASCADE;
            DROP FUNCTION IF EXISTS generate_business_id(TEXT) CASCADE;
            DROP FUNCTION IF EXISTS validate_business_id(TEXT, TEXT) CASCADE;
            ALTER TABLE corehr.employees DROP COLUMN IF EXISTS business_id;
            ALTER TABLE corehr.organizations DROP COLUMN IF EXISTS business_id;
            DROP SEQUENCE IF EXISTS employee_business_id_seq;
            DROP SEQUENCE IF EXISTS org_business_id_seq; 
            DROP SEQUENCE IF EXISTS position_business_id_seq;" \
        >> "$rollback_log" 2>&1
    
    # Neo4j回滚 (如果可用)
    if command -v cypher-shell >/dev/null 2>&1; then
        log_info "回滚Neo4j更改..."
        cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "MATCH (e:Employee) REMOVE e.business_id;
             MATCH (o:Organization) REMOVE o.business_id;
             DROP INDEX employee_business_id_index IF EXISTS;
             DROP INDEX organization_business_id_index IF EXISTS;
             DROP CONSTRAINT employee_business_id_unique IF EXISTS;
             DROP CONSTRAINT organization_business_id_unique IF EXISTS;" \
            >> "$rollback_log" 2>&1
    fi
    
    log_success "回滚完成，日志保存在: $rollback_log"
}

# =============================================
# 主函数  
# =============================================

main() {
    local command="${1:-migrate}"
    
    case "$command" in
        "migrate")
            log_info "开始业务ID系统迁移..."
            check_prerequisites
            backup_postgresql
            backup_neo4j
            execute_postgresql_migration
            execute_neo4j_migration
            verify_data_consistency
            run_performance_tests
            log_success "业务ID系统迁移完成！"
            ;;
        "rollback")
            log_warning "开始回滚业务ID系统迁移..."
            rollback_migration
            log_success "回滚完成！"
            ;;
        "verify")
            log_info "开始数据验证..."
            verify_data_consistency
            ;;
        "performance")
            log_info "开始性能测试..."
            run_performance_tests
            ;;
        "backup")
            log_info "开始数据备份..."
            backup_postgresql
            backup_neo4j
            ;;
        *)
            echo "用法: $0 {migrate|rollback|verify|performance|backup}"
            echo ""
            echo "命令说明:"
            echo "  migrate     - 执行完整的业务ID迁移"
            echo "  rollback    - 回滚业务ID迁移"
            echo "  verify      - 验证数据一致性"
            echo "  performance - 运行性能基准测试"
            echo "  backup      - 仅执行数据备份"
            echo ""
            echo "环境变量:"
            echo "  POSTGRES_HOST     - PostgreSQL主机 (默认: localhost)"
            echo "  POSTGRES_PORT     - PostgreSQL端口 (默认: 5432)"
            echo "  POSTGRES_DB       - PostgreSQL数据库 (默认: cubecastle)"
            echo "  POSTGRES_USER     - PostgreSQL用户 (默认: user)"
            echo "  POSTGRES_PASSWORD - PostgreSQL密码 (默认: password)"
            echo "  NEO4J_HOST        - Neo4j主机 (默认: localhost)"
            echo "  NEO4J_PORT        - Neo4j端口 (默认: 7687)"
            echo "  NEO4J_USER        - Neo4j用户 (默认: neo4j)"
            echo "  NEO4J_PASSWORD    - Neo4j密码 (默认: password)"
            exit 1
            ;;
    esac
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi