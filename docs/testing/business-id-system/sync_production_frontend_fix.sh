#!/bin/bash
# sync_production_frontend_fix.sh
# 修正tenant_id的业务ID数据同步脚本，匹配前端使用的tenant_id

set -e  

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'  
NC='\033[0m'

# 数据库连接信息
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="user"
DB_PASS="password"
DB_NAME="cubecastle"

# 正确的tenant_id，匹配前端使用的ID
FRONTEND_TENANT_ID="550e8400-e29b-41d4-a716-446655440000"

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_production() {
    echo -e "${PURPLE}🚀 [生产环境] $1${NC}"
}

log_fix() {
    echo -e "${YELLOW}🔧 [修复] $1${NC}"
}

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

main() {
    echo "🔧 修复tenant_id的业务ID数据同步（前端匹配版）"
    echo "================================================="
    echo ""
    log_fix "问题描述：前端使用tenant_id ${FRONTEND_TENANT_ID}"
    log_fix "之前同步的数据使用了错误的tenant_id 00000000-0000-0000-0000-000000000000"
    log_fix "现在重新同步数据以匹配前端配置"
    echo ""
    
    log_info "检查数据库连接..."
    if PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
        log_success "数据库连接正常"
    else
        log_error "数据库连接失败"
        exit 1
    fi
    
    # 清除现有数据
    log_production "清除现有数据..."
    execute_sql "SET session_replication_role = replica;" "禁用外键约束"
    execute_sql "DELETE FROM employees;" "清除员工数据"
    execute_sql "DELETE FROM organization_units;" "清除组织数据"
    execute_sql "DELETE FROM positions;" "清除职位数据"
    execute_sql "SET session_replication_role = DEFAULT;" "恢复外键约束"
    
    # 插入员工数据（使用前端tenant_id）
    log_production "插入员工数据（使用正确的tenant_id）..."
    execute_sql "
    INSERT INTO employees (id, tenant_id, employee_type, first_name, last_name,
                          email, hire_date, employment_status, business_id, 
                          created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        '$FRONTEND_TENANT_ID',
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
        CURRENT_DATE - (s % 1000)::int,
        'ACTIVE',
        (s + 1)::varchar,
        NOW(),
        NOW()
    FROM generate_series(0, 499) s;
    " "生成500个员工记录"
    
    # 插入边界测试员工
    execute_sql "
    INSERT INTO employees (id, tenant_id, employee_type, first_name, last_name,
                          email, hire_date, employment_status, business_id, 
                          created_at, updated_at)
    VALUES 
        (gen_random_uuid(), '$FRONTEND_TENANT_ID', 'FULL_TIME',
         '边界', '测试最大', 'boundary_max_emp@test.com', CURRENT_DATE, 'ACTIVE', '99999',
         NOW(), NOW());
    " "插入边界测试员工"
    
    # 插入组织数据
    log_production "插入组织数据（使用正确的tenant_id）..."
    execute_sql "
    INSERT INTO organization_units (id, tenant_id, unit_type, name, description, parent_unit_id, 
                                   status, level, employee_count, is_active, business_id, 
                                   created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        '$FRONTEND_TENANT_ID',
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
    " "生成50个组织记录"
    
    # 插入边界测试组织
    execute_sql "
    INSERT INTO organization_units (id, tenant_id, unit_type, name, status, level, 
                                   employee_count, is_active, business_id, created_at, updated_at)
    VALUES 
        (gen_random_uuid(), '$FRONTEND_TENANT_ID', 'DEPARTMENT',
         '边界组织999999', 'ACTIVE', 1, 0, true, '999999', NOW(), NOW());
    " "插入边界测试组织"
    
    # 先创建一个department用于positions
    execute_sql "
    INSERT INTO organization_units (id, tenant_id, unit_type, name, status, level, 
                                   employee_count, is_active, business_id, created_at, updated_at)
    VALUES 
        ('11111111-1111-1111-1111-111111111111', '$FRONTEND_TENANT_ID', 'DEPARTMENT',
         '默认部门', 'ACTIVE', 1, 0, true, '100050', NOW(), NOW())
    ON CONFLICT (business_id) DO NOTHING;
    " "创建默认部门"
    
    # 插入职位数据（使用正确的字段）
    log_production "插入职位数据（使用正确的tenant_id）..."
    execute_sql "
    INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id,
                          status, budgeted_fte, business_id, created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        '$FRONTEND_TENANT_ID',
        'REGULAR',
        gen_random_uuid(),
        '11111111-1111-1111-1111-111111111111',
        'ACTIVE',
        1.0,
        (1000000 + s)::varchar,
        NOW(),
        NOW()
    FROM generate_series(0, 99) s;
    " "生成100个职位记录"
    
    # 插入边界测试职位
    execute_sql "
    INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id,
                          status, budgeted_fte, business_id, created_at, updated_at)
    VALUES 
        (gen_random_uuid(), '$FRONTEND_TENANT_ID', 'REGULAR',
         gen_random_uuid(), '11111111-1111-1111-1111-111111111111', 'ACTIVE', 1.0, '9999999', 
         NOW(), NOW());
    " "插入边界测试职位"
    
    # 更新序列
    log_production "更新序列..."
    execute_sql "SELECT setval('employee_business_id_seq', 501);" "更新员工序列"
    execute_sql "SELECT setval('org_business_id_seq', 52);" "更新组织序列"
    execute_sql "SELECT setval('position_business_id_seq', 101);" "更新职位序列"
    
    # 验证结果
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
    " "数据统计验证"
    
    # 验证tenant_id匹配
    log_fix "验证tenant_id匹配..."
    execute_sql_with_output "
    SELECT 
        'tenant_id匹配检查' as 检查项目,
        CASE 
            WHEN (SELECT COUNT(*) FROM employees WHERE tenant_id = '$FRONTEND_TENANT_ID') > 0
            THEN '✅ 员工表匹配前端tenant_id'
            ELSE '❌ 员工表tenant_id不匹配'
        END as 员工表结果,
        CASE 
            WHEN (SELECT COUNT(*) FROM organization_units WHERE tenant_id = '$FRONTEND_TENANT_ID') > 0
            THEN '✅ 组织表匹配前端tenant_id'
            ELSE '❌ 组织表tenant_id不匹配'
        END as 组织表结果,
        CASE 
            WHEN (SELECT COUNT(*) FROM positions WHERE tenant_id = '$FRONTEND_TENANT_ID') > 0
            THEN '✅ 职位表匹配前端tenant_id'
            ELSE '❌ 职位表tenant_id不匹配'
        END as 职位表结果;
    " "tenant_id匹配验证"
    
    echo ""
    log_success "🎉 tenant_id修复完成！"
    log_fix "📊 修复统计："
    log_fix "   • 员工数据: 501条 (ID范围: 1-99999)"
    log_fix "   • 组织数据: 52条 (ID范围: 100000-999999)"
    log_fix "   • 职位数据: 101条 (ID范围: 1000000-9999999)"
    log_fix "   • 总计: 654条记录"
    log_fix "   • tenant_id: $FRONTEND_TENANT_ID (匹配前端)"
    echo ""
    log_production "🚀 现在前端应该能正确显示数据了！"
    log_production "🧪 访问 http://localhost:3000/ 验证数据显示"
    echo ""
}

main "$@"