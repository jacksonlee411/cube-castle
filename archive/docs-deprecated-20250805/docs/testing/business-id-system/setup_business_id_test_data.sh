#!/bin/bash
# setup_business_id_test_data.sh
# 业务ID系统测试数据创建脚本

set -e  # 遇到错误立即退出
set -o pipefail  # 管道中任何命令失败都会导致整个管道失败

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 执行SQL文件
execute_sql_file() {
    local sql_file=$1
    local description=$2
    
    log_info "$description"
    if PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$sql_file"; then
        log_success "$description 完成"
    else
        log_error "$description 失败"
        exit 1
    fi
}

# 执行SQL命令
execute_sql() {
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

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_DIR="$SCRIPT_DIR/sql"

echo "🚀 开始创建业务ID系统测试数据..."
echo "📍 脚本目录: $SCRIPT_DIR"
echo "📍 SQL目录: $SQL_DIR"

# 1. 检查数据库连接
check_db_connection

# 2. 创建SQL目录（如果不存在）
mkdir -p "$SQL_DIR"

# 3. 添加业务ID字段
if [ -f "$SQL_DIR/add_business_id_fields.sql" ]; then
    execute_sql_file "$SQL_DIR/add_business_id_fields.sql" "添加业务ID字段"
else
    log_warning "add_business_id_fields.sql 文件不存在，跳过"
fi

# 4. 创建序列
if [ -f "$SQL_DIR/create_sequences.sql" ]; then
    execute_sql_file "$SQL_DIR/create_sequences.sql" "创建业务ID序列"
else
    log_warning "create_sequences.sql 文件不存在，跳过"
fi

# 5. 插入基础测试数据
log_info "插入基础测试数据..."

# 创建组织数据
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
FROM generate_series(0, 49) s
ON CONFLICT (business_id) DO NOTHING;
" "创建50个组织单元"

# 创建职位数据
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
FROM generate_series(0, 199) s
ON CONFLICT (business_id) DO NOTHING;
" "创建200个职位"

# 创建员工数据
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
    (s + 1)::varchar,
    NOW(),
    NOW()
FROM generate_series(0, 999) s
ON CONFLICT (business_id) DO NOTHING;
" "创建1000个员工"

# 6. 创建边界条件测试数据
log_info "创建边界条件测试数据..."

# 员工ID边界值
execute_sql "
INSERT INTO employees (id, tenant_id, employee_number, employee_type, first_name, last_name,
                      email, status, hire_date, employment_status, business_id, 
                      created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'EMP_BOUNDARY_MIN', 'FULL_TIME',
     '边界', '测试最小', 'boundary_min_emp@test.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '99998',
     NOW(), NOW()),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'EMP_BOUNDARY_MAX', 'FULL_TIME', 
     '边界', '测试最大', 'boundary_max_emp@test.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '99999',
     NOW(), NOW())
ON CONFLICT (business_id) DO NOTHING;
" "创建员工边界测试数据"

# 组织ID边界值  
execute_sql "
INSERT INTO organization_units (id, tenant_id, unit_type, name, status, level, 
                               employee_count, is_active, business_id, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'DEPARTMENT', 
     '边界组织999998', 'ACTIVE', 1, 0, true, '999998', NOW(), NOW()),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'DEPARTMENT',
     '边界组织999999', 'ACTIVE', 1, 0, true, '999999', NOW(), NOW())
ON CONFLICT (business_id) DO NOTHING;
" "创建组织边界测试数据"

# 职位ID边界值
execute_sql "
INSERT INTO positions (id, tenant_id, position_type, title, code, job_profile_id,
                      status, budgeted_fte, business_id, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'TECHNICAL',
     '边界职位9999998', 'POS_BOUNDARY_1', gen_random_uuid(), 'ACTIVE', 1.0, '9999998',
     NOW(), NOW()),
     (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'TECHNICAL',
     '边界职位9999999', 'POS_BOUNDARY_2', gen_random_uuid(), 'ACTIVE', 1.0, '9999999', 
     NOW(), NOW())
ON CONFLICT (business_id) DO NOTHING;
" "创建职位边界测试数据"

# 7. 验证数据创建结果
log_info "验证数据创建结果..."

execute_sql "
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
" "数据统计验证"

# 8. 检查业务ID唯一性
execute_sql "
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
" "业务ID唯一性验证"

log_success "✅ 测试数据创建完成！"
log_info "📊 数据总览："
log_info "   • 员工数据: 1002条 (包含边界测试数据)"
log_info "   • 组织数据: 52条 (包含边界测试数据)"  
log_info "   • 职位数据: 202条 (包含边界测试数据)"
log_info ""
log_info "🧪 现在可以运行真实数据库测试了："
log_info "   cd /home/shangmeilin/cube-castle/go-app"
log_info "   export TEST_WITH_REAL_DB=true"
log_info "   go test -v ./internal/common -run TestBusinessIDService.*WithRealDB"
log_info ""
log_success "真实数据库测试环境已准备就绪！"