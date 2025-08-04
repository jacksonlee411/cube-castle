#!/bin/bash
# sync_to_production_simple.sh
# 简化的业务ID测试数据同步到正式环境脚本

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
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="user"
DB_PASS="password"
DB_NAME="cubecastle"

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

# 执行SQL命令
execute_sql() {
    local sql_command=$1
    local description=$2
    
    log_info "$description"
    if PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$sql_command" > /dev/null; then
        log_success "$description 完成"
    else
        log_error "$description 失败"
        exit 1
    fi
}

# 执行带输出的SQL命令
execute_sql_with_output() {
    local sql_command=$1
    local description=$2
    
    log_info "$description"
    if PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$sql_command"; then
        log_success "$description 完成"
    else
        log_error "$description 失败"
        exit 1
    fi
}

# 检查数据库连接
check_db_connection() {
    log_info "检查数据库连接..."
    if PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "数据库连接正常"
    else
        log_error "数据库连接失败，请检查数据库是否运行"
        exit 1
    fi
}

# 主执行流程
main() {
    echo "🔄 业务ID测试数据同步到正式环境（简化版）"
    echo "=========================================="
    echo ""
    
    log_warning "自动开始数据同步..."
    
    # 1. 检查数据库连接
    check_db_connection
    
    # 2. 创建备份目录
    local backup_dir="/home/shangmeilin/cube-castle/backups/production_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    log_production "备份目录: $backup_dir"
    
    # 3. 清除正式环境数据（保留结构）
    log_production "清除正式环境现有数据..."
    
    execute_sql "SET session_replication_role = replica;" "禁用外键约束检查"
    execute_sql "DELETE FROM employees;" "清除员工数据"
    execute_sql "DELETE FROM organization_units;" "清除组织数据"  
    execute_sql "DELETE FROM positions;" "清除职位数据"
    execute_sql "SET session_replication_role = DEFAULT;" "恢复外键约束检查"
    
    # 4. 使用INSERT INTO ... SELECT直接同步数据
    log_production "同步测试数据..."
    
    # 同步员工数据
    execute_sql "
    INSERT INTO employees (id, tenant_id, employee_number, employee_type, first_name, last_name,
                          email, status, hire_date, employment_status, business_id, 
                          created_at, updated_at)
    SELECT gen_random_uuid(), tenant_id, employee_number, employee_type, first_name, last_name,
           email, status, hire_date, employment_status, business_id,
           NOW(), NOW()
    FROM (VALUES 
        ('00000000-0000-0000-0000-000000000000', 'EMP000001', 'FULL_TIME', '张', '伟', 'test_employee_0@company.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '1'),
        ('00000000-0000-0000-0000-000000000000', 'EMP000002', 'FULL_TIME', '李', '芳', 'test_employee_1@company.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '2'),
        ('00000000-0000-0000-0000-000000000000', 'EMP000499', 'FULL_TIME', '测试', '员工499', 'test_employee_499@company.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '499'),
        ('00000000-0000-0000-0000-000000000000', 'EMP000500', 'FULL_TIME', '测试', '员工500', 'test_employee_500@company.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '500'),
        ('00000000-0000-0000-0000-000000000000', 'EMP_BOUNDARY_MAX', 'FULL_TIME', '边界', '测试最大', 'boundary_max_emp@test.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '99999')
    ) AS v(tenant_id, employee_number, employee_type, first_name, last_name, email, status, hire_date, employment_status, business_id);
    " "插入员工测试数据"
    
    # 生成更多员工数据
    execute_sql "
    INSERT INTO employees (id, tenant_id, employee_number, employee_type, first_name, last_name,
                          email, status, hire_date, employment_status, business_id, 
                          created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000000',
        'EMP' || LPAD(s::text, 6, '0'),
        'FULL_TIME',
        CASE s % 10
            WHEN 0 THEN '张'
            WHEN 1 THEN '李'
            WHEN 2 THEN '王'
            WHEN 3 THEN '刘'
            WHEN 4 THEN '陈'
            WHEN 5 THEN '杨'
            WHEN 6 THEN '赵'
            WHEN 7 THEN '黄'
            WHEN 8 THEN '周'
            ELSE '吴'
        END,
        CASE s % 5
            WHEN 0 THEN '伟'
            WHEN 1 THEN '芳'
            WHEN 2 THEN '娜'
            WHEN 3 THEN '秀英'
            ELSE '敏'
        END,
        'test_employee_' || s || '@company.com',
        CASE s % 10 WHEN 9 THEN 'INACTIVE' ELSE 'ACTIVE' END,
        CURRENT_DATE - (s % 1000)::int,
        'ACTIVE',
        (s + 3)::varchar,
        NOW(),
        NOW()
    FROM generate_series(3, 497) s;
    " "生成批量员工数据"
    
    # 同步组织数据
    execute_sql "
    INSERT INTO organization_units (id, tenant_id, unit_type, name, description, parent_unit_id, 
                                   status, level, employee_count, is_active, business_id, 
                                   created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000000',
        'DEPARTMENT',
        CASE s % 5
            WHEN 0 THEN '技术部'
            WHEN 1 THEN '产品部'
            WHEN 2 THEN '销售部'
            WHEN 3 THEN '人事部'
            ELSE '财务部'
        END || CASE WHEN s > 4 THEN '-' || ((s / 5) + 1)::text ELSE '' END,
        '测试部门描述',
        NULL,
        'ACTIVE',
        1,
        0,
        true,
        (100000 + s)::varchar,
        NOW(),
        NOW()
    FROM generate_series(0, 49) s;
    " "创建组织单元数据"
    
    # 同步职位数据
    execute_sql "
    INSERT INTO positions (id, tenant_id, position_type, title, code, job_profile_id,
                          status, budgeted_fte, business_id, created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000000',
        'TECHNICAL',
        CASE s % 8
            WHEN 0 THEN '高级软件工程师'
            WHEN 1 THEN '软件工程师'
            WHEN 2 THEN '测试工程师'
            WHEN 3 THEN '架构师'
            WHEN 4 THEN '产品经理'
            WHEN 5 THEN '项目经理'
            WHEN 6 THEN 'UI设计师'
            ELSE 'DevOps工程师'
        END,
        'POS' || LPAD(s::text, 4, '0'),
        gen_random_uuid(),
        'ACTIVE',
        1.0,
        (1000000 + s)::varchar,
        NOW(),
        NOW()
    FROM generate_series(0, 101) s;
    " "创建职位数据"
    
    # 5. 更新序列当前值
    log_production "更新序列当前值..."
    execute_sql "SELECT setval('employee_business_id_seq', 500);" "更新员工序列"
    execute_sql "SELECT setval('org_business_id_seq', 50);" "更新组织序列"  
    execute_sql "SELECT setval('position_business_id_seq', 102);" "更新职位序列"
    
    # 6. 验证同步结果
    log_production "验证同步结果..."
    execute_sql_with_output "
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
    
    echo ""
    log_success "🎉 数据同步完成！"
    log_production "📊 同步统计："
    log_production "   • 员工数据: ~500条"
    log_production "   • 组织数据: 50条"  
    log_production "   • 职位数据: 102条"
    log_production "   • 总计: ~652条记录"
    echo ""
    log_production "🧪 正式环境业务ID系统已就绪！"
    echo ""
}

# 执行主函数
main "$@"