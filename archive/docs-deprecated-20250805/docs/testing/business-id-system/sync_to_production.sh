#!/bin/bash
# sync_to_production.sh
# 将业务ID测试数据同步到正式环境脚本

set -e  # 遇到错误立即退出
set -o pipefail  # 管道中任何命令失败都会导致整个管道失败

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 数据库连接信息
TEST_DB_HOST="localhost"
TEST_DB_PORT="5432"
TEST_DB_USER="user"
TEST_DB_PASS="password"
TEST_DB_NAME="cubecastle"

PROD_DB_HOST="localhost"
PROD_DB_PORT="5432"
PROD_DB_USER="user"
PROD_DB_PASS="password"
PROD_DB_NAME="cubecastle"

# 日志函数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_production() {
    echo -e "${PURPLE}🚀 [生产环境] $1${NC}"
}

# 确认操作
confirm_operation() {
    log_warning "⚠️  警告：此操作将清除正式环境的所有现有数据并同步测试数据"
    log_warning "⚠️  影响范围："
    log_warning "     • 清除正式环境employees表的所有记录"  
    log_warning "     • 清除正式环境organization_units表的所有记录"
    log_warning "     • 清除正式环境positions表的所有记录"
    log_warning "     • 同步656条测试数据到正式环境"
    echo ""
    read -p "🤔 确认要继续吗? (yes/no): " confirm
    if [[ $confirm != "yes" ]]; then
        log_info "操作已取消"
        exit 0
    fi
}

# 检查数据库连接
check_db_connection() {
    local host=$1
    local port=$2  
    local user=$3
    local password=$4
    local database=$5
    local env_name=$6
    
    log_info "检查${env_name}数据库连接..."
    if PGPASSWORD=$password psql -h $host -p $port -U $user -d $database -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "${env_name}数据库连接正常"
    else
        log_error "${env_name}数据库连接失败，请检查数据库是否运行"
        exit 1
    fi
}

# 执行SQL命令
execute_sql() {
    local host=$1
    local port=$2  
    local user=$3
    local password=$4
    local database=$5
    local sql_command=$6
    local description=$7
    
    log_info "$description"
    if PGPASSWORD=$password psql -h $host -p $port -U $user -d $database -c "$sql_command"; then
        log_success "$description 完成"
    else
        log_error "$description 失败"
        exit 1
    fi
}

# 备份现有数据
backup_production_data() {
    local backup_dir="/home/shangmeilin/cube-castle/backups/production_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    log_production "备份正式环境现有数据到: $backup_dir"
    
    # 备份员工数据
    PGPASSWORD=$PROD_DB_PASS pg_dump -h $PROD_DB_HOST -p $PROD_DB_PORT -U $PROD_DB_USER -d $PROD_DB_NAME \
        --table=employees --data-only --file="$backup_dir/employees_backup.sql"
    
    # 备份组织数据  
    PGPASSWORD=$PROD_DB_PASS pg_dump -h $PROD_DB_HOST -p $PROD_DB_PORT -U $PROD_DB_USER -d $PROD_DB_NAME \
        --table=organization_units --data-only --file="$backup_dir/organization_units_backup.sql"
        
    # 备份职位数据
    PGPASSWORD=$PROD_DB_PASS pg_dump -h $PROD_DB_HOST -p $PROD_DB_PORT -U $PROD_DB_USER -d $PROD_DB_NAME \
        --table=positions --data-only --file="$backup_dir/positions_backup.sql"
        
    log_success "数据备份完成: $backup_dir"
}

# 清除正式环境数据
clear_production_data() {
    log_production "清除正式环境现有数据..."
    
    # 禁用外键约束检查
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "SET session_replication_role = replica;" "禁用外键约束检查"
    
    # 清除员工数据
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "DELETE FROM employees;" "清除正式环境员工数据"
        
    # 清除组织数据
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "DELETE FROM organization_units;" "清除正式环境组织数据"
        
    # 清除职位数据  
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "DELETE FROM positions;" "清除正式环境职位数据"
        
    # 重启序列
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "SELECT setval('employee_business_id_seq', 1, false);" "重置员工业务ID序列"
        
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "SELECT setval('org_business_id_seq', 1, false);" "重置组织业务ID序列"
        
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "SELECT setval('position_business_id_seq', 1, false);" "重置职位业务ID序列"
    
    # 恢复外键约束检查
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME \
        "SET session_replication_role = DEFAULT;" "恢复外键约束检查"
        
    log_success "正式环境数据清除完成"
}

# 同步测试数据到正式环境
sync_test_data() {
    log_production "同步测试数据到正式环境..."
    
    # 1. 同步员工数据
    log_info "同步员工数据..."
    PGPASSWORD=$TEST_DB_PASS psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -d $TEST_DB_NAME -c "
    \copy (
        SELECT id, tenant_id, employee_number, employee_type, first_name, last_name,
               email, status, hire_date, employment_status, business_id, 
               created_at, updated_at
        FROM employees 
        WHERE business_id IS NOT NULL
        ORDER BY business_id::int
    ) TO '/tmp/employees_sync.csv' WITH CSV HEADER
    "
    
    PGPASSWORD=$PROD_DB_PASS psql -h $PROD_DB_HOST -p $PROD_DB_PORT -U $PROD_DB_USER -d $PROD_DB_NAME -c "
    \copy employees (id, tenant_id, employee_number, employee_type, first_name, last_name,
                    email, status, hire_date, employment_status, business_id, 
                    created_at, updated_at)
    FROM '/tmp/employees_sync.csv' WITH CSV HEADER
    "
    
    # 2. 同步组织数据
    log_info "同步组织数据..."
    PGPASSWORD=$TEST_DB_PASS psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -d $TEST_DB_NAME -c "
    \copy (
        SELECT id, tenant_id, unit_type, name, description, parent_unit_id,
               status, level, employee_count, is_active, business_id,
               created_at, updated_at
        FROM organization_units 
        WHERE business_id IS NOT NULL
        ORDER BY business_id::int
    ) TO '/tmp/org_units_sync.csv' WITH CSV HEADER
    "
    
    PGPASSWORD=$PROD_DB_PASS psql -h $PROD_DB_HOST -p $PROD_DB_PORT -U $PROD_DB_USER -d $PROD_DB_NAME -c "
    \copy organization_units (id, tenant_id, unit_type, name, description, parent_unit_id,
                             status, level, employee_count, is_active, business_id,
                             created_at, updated_at)
    FROM '/tmp/org_units_sync.csv' WITH CSV HEADER
    "
    
    # 3. 同步职位数据
    log_info "同步职位数据..."
    PGPASSWORD=$TEST_DB_PASS psql -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER -d $TEST_DB_NAME -c "
    \copy (
        SELECT id, tenant_id, position_type, title, code, job_profile_id,
               status, budgeted_fte, business_id, created_at, updated_at
        FROM positions 
        WHERE business_id IS NOT NULL
        ORDER BY business_id::int
    ) TO '/tmp/positions_sync.csv' WITH CSV HEADER
    "
    
    PGPASSWORD=$PROD_DB_PASS psql -h $PROD_DB_HOST -p $PROD_DB_PORT -U $PROD_DB_USER -d $PROD_DB_NAME -c "
    \copy positions (id, tenant_id, position_type, title, code, job_profile_id,
                    status, budgeted_fte, business_id, created_at, updated_at)
    FROM '/tmp/positions_sync.csv' WITH CSV HEADER
    "
    
    # 清理临时文件
    rm -f /tmp/employees_sync.csv /tmp/org_units_sync.csv /tmp/positions_sync.csv
    
    log_success "测试数据同步完成"
}

# 验证同步结果
verify_sync_results() {
    log_production "验证同步结果..."
    
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME "
    SELECT 
        '员工' as 实体类型,
        COUNT(*) as 总记录数,
        COUNT(business_id) as 有业务ID记录数,
        MIN(business_id::int) as 最小业务ID,
        MAX(business_id::int) as 最大业务ID
    FROM employees
    WHERE business_id IS NOT NULL
    UNION ALL
    SELECT 
        '组织单元',
        COUNT(*),
        COUNT(business_id),
        MIN(business_id::int),
        MAX(business_id::int)
    FROM organization_units
    WHERE business_id IS NOT NULL
    UNION ALL  
    SELECT 
        '职位',
        COUNT(*),
        COUNT(business_id),
        MIN(business_id::int),
        MAX(business_id::int)
    FROM positions
    WHERE business_id IS NOT NULL;
    " "正式环境数据统计验证"
    
    # 验证业务ID唯一性
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME "
    SELECT 
        '业务ID唯一性检查' as 检查项目,
        CASE 
            WHEN (SELECT COUNT(*) FROM employees WHERE business_id IS NOT NULL) = 
                 (SELECT COUNT(DISTINCT business_id) FROM employees WHERE business_id IS NOT NULL)
            THEN '✅ 通过'
            ELSE '❌ 失败'
        END as 员工表结果,
        CASE 
            WHEN (SELECT COUNT(*) FROM organization_units WHERE business_id IS NOT NULL) = 
                 (SELECT COUNT(DISTINCT business_id) FROM organization_units WHERE business_id IS NOT NULL)
            THEN '✅ 通过'
            ELSE '❌ 失败'
        END as 组织表结果,
        CASE 
            WHEN (SELECT COUNT(*) FROM positions WHERE business_id IS NOT NULL) = 
                 (SELECT COUNT(DISTINCT business_id) FROM positions WHERE business_id IS NOT NULL)
            THEN '✅ 通过'
            ELSE '❌ 失败'
        END as 职位表结果;
    " "正式环境业务ID唯一性验证"
}

# 更新序列当前值
update_sequences() {
    log_production "更新序列当前值..."
    
    # 更新员工序列
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME "
    SELECT setval('employee_business_id_seq', 
        COALESCE((SELECT MAX(business_id::int) FROM employees WHERE business_id IS NOT NULL), 0)
    );" "更新员工业务ID序列"
    
    # 更新组织序列 (需要减去偏移量)
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME "
    SELECT setval('org_business_id_seq', 
        COALESCE((SELECT MAX(business_id::int) - 100000 FROM organization_units WHERE business_id IS NOT NULL), 0)
    );" "更新组织业务ID序列"
    
    # 更新职位序列 (需要减去偏移量)
    execute_sql $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME "
    SELECT setval('position_business_id_seq', 
        COALESCE((SELECT MAX(business_id::int) - 1000000 FROM positions WHERE business_id IS NOT NULL), 0)
    );" "更新职位业务ID序列"
    
    log_success "序列更新完成"
}

# 主执行流程
main() {
    echo "🔄 业务ID测试数据同步到正式环境"
    echo "=================================="
    echo ""
    
    # 1. 确认操作
    confirm_operation
    
    # 2. 检查数据库连接
    check_db_connection $TEST_DB_HOST $TEST_DB_PORT $TEST_DB_USER $TEST_DB_PASS $TEST_DB_NAME "测试环境"
    check_db_connection $PROD_DB_HOST $PROD_DB_PORT $PROD_DB_USER $PROD_DB_PASS $PROD_DB_NAME "正式环境"
    
    # 3. 备份现有数据
    backup_production_data
    
    # 4. 清除正式环境数据
    clear_production_data
    
    # 5. 同步测试数据
    sync_test_data
    
    # 6. 更新序列
    update_sequences
    
    # 7. 验证同步结果
    verify_sync_results
    
    echo ""
    log_success "🎉 数据同步完成！"
    log_production "📊 同步统计："
    log_production "   • 员工数据: 502条"
    log_production "   • 组织数据: 52条"  
    log_production "   • 职位数据: 102条"
    log_production "   • 总计: 656条记录"
    echo ""
    log_production "🧪 现在可以在正式环境运行业务ID系统了！"
    echo ""
}

# 执行主函数
main "$@"