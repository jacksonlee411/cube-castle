#!/bin/bash

# 正确的版本删除API使用示例
# 解决"九月二日记录删除后，九月一号记录结束日期没有自动更新"的问题

set -e

echo "🧪 正确的版本删除API测试"
echo "演示如何正确删除特定版本并保持时态时间轴连续性"

# 配置
BASE_URL="http://localhost:9090"
API_BASE="$BASE_URL/api/v1/organization-units"
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"
ORG_CODE="TEST001"

# JWT Token (开发模式下的测试token)
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJpYXQiOjE3MzM0NjM2MDAsImV4cCI6MTc2NDk5OTYwMCwidGVuYW50SWQiOiI1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAiLCJyb2xlcyI6WyJhZG1pbiJdLCJwZXJtaXNzaW9ucyI6WyJvcmc6Y3JlYXRlIiwib3JnOnVwZGF0ZSIsIm9yZzpkZWxldGUiLCJvcmc6cXVlcnkiXX0.rIvXhSfQT2_m9p-KlQGJdz5x6h8h5f3nV7Kgr2sL9iE"

# HTTP请求公共header
HEADERS=(
    -H "Content-Type: application/json"
    -H "Authorization: Bearer $JWT_TOKEN"
    -H "X-Tenant-ID: $TENANT_ID"
)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# 检查服务器状态
check_server() {
    log_info "检查服务器状态..."
    if curl -s "$BASE_URL/health" > /dev/null; then
        log_success "服务器运行正常"
    else
        log_error "服务器未运行，请启动服务"
        exit 1
    fi
}

# 查看数据库当前状态
query_timeline() {
    local title="$1"
    log_info "📊 查询时间轴状态: $title"
    
    PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "
    SELECT 
        '  📅 ' || effective_date || ' → ' || 
        COALESCE(end_date::text, '∞') || ' | ' || 
        name || ' (' || status || ')' ||
        CASE WHEN is_current THEN ' ⭐' ELSE '' END as timeline
    FROM organization_units 
    WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' 
      AND status != 'DELETED'
    ORDER BY effective_date;
    " 2>/dev/null | grep -v '^$' || log_warning "未找到时间轴数据"
    echo ""
}

# 步骤1：创建多个版本
create_versions() {
    log_info "步骤1: 通过API创建多个版本"
    
    # 版本2: 2025-09-02
    log_info "创建版本2: 2025-09-02"
    local response2=$(curl -s -X POST "$API_BASE/$ORG_CODE/versions" "${HEADERS[@]}" -d '{
        "name": "测试部门 v2.0",
        "unitType": "DEPARTMENT",
        "effectiveDate": "2025-09-02",
        "operationReason": "第二版本"
    }')
    
    if echo "$response2" | grep -q '"success":true'; then
        log_success "版本2创建成功"
    else
        log_error "版本2创建失败: $response2"
    fi
    
    # 版本3: 2025-09-03
    log_info "创建版本3: 2025-09-03"
    local response3=$(curl -s -X POST "$API_BASE/$ORG_CODE/versions" "${HEADERS[@]}" -d '{
        "name": "测试部门 v3.0", 
        "unitType": "DEPARTMENT",
        "effectiveDate": "2025-09-03",
        "operationReason": "第三版本"
    }')
    
    if echo "$response3" | grep -q '"success":true'; then
        log_success "版本3创建成功"
    else
        log_error "版本3创建失败: $response3"
    fi
    
    # 版本4: 2025-09-04
    log_info "创建版本4: 2025-09-04"
    local response4=$(curl -s -X POST "$API_BASE/$ORG_CODE/versions" "${HEADERS[@]}" -d '{
        "name": "测试部门 v4.0",
        "unitType": "DEPARTMENT", 
        "effectiveDate": "2025-09-04",
        "operationReason": "第四版本"
    }')
    
    if echo "$response4" | grep -q '"success":true'; then
        log_success "版本4创建成功"
    else
        log_error "版本4创建失败: $response4"
    fi
    
    log_success "多版本创建完成"
}

# 步骤2：演示错误的删除方式 (用户之前使用的)
demo_wrong_deletion() {
    log_warning "❌ 注意：物理删除端点 DELETE /{code} 已被移除"
    log_warning "现在只能通过版本删除端点 DELETE /versions/{recordId} 删除特定版本"
    log_warning "这确保了所有删除操作都会触发时态时间轴重计算"
    echo ""
}

# 步骤3：演示正确的删除方式
demo_correct_deletion() {
    log_info "步骤2: 正确删除特定版本 (2025-09-02)"
    
    # 获取2025-09-02版本的record_id
    local record_id=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "
    SELECT record_id 
    FROM organization_units 
    WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' 
      AND effective_date = '2025-09-02'
      AND status != 'DELETED'
    LIMIT 1;
    " 2>/dev/null | xargs)
    
    if [ -z "$record_id" ]; then
        log_error "未找到2025-09-02版本的record_id"
        return 1
    fi
    
    log_info "找到版本record_id: $record_id"
    
    # ✅ 正确的删除方式：使用版本删除端点
    log_info "🚀 使用正确的API端点: DELETE /versions/{recordId}"
    
    local delete_response=$(curl -s -X DELETE "$API_BASE/versions/$record_id" "${HEADERS[@]}")
    
    if echo "$delete_response" | grep -q '"success":true'; then
        log_success "✅ 版本删除成功！时态时间轴已自动重新计算"
    else
        log_error "版本删除失败: $delete_response"
        return 1
    fi
}

# 步骤4：验证结果
verify_timeline_consistency() {
    log_info "步骤3: 验证时态时间轴连续性"
    
    # 检查时间断档
    local gaps=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "
    WITH timeline AS (
        SELECT 
            effective_date,
            end_date,
            LEAD(effective_date) OVER (ORDER BY effective_date) as next_start
        FROM organization_units 
        WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' 
          AND status != 'DELETED'
        ORDER BY effective_date
    )
    SELECT COUNT(*) 
    FROM timeline 
    WHERE end_date IS NOT NULL 
      AND next_start IS NOT NULL 
      AND end_date + INTERVAL '1 day' != next_start;
    " 2>/dev/null | xargs)
    
    if [ "$gaps" = "0" ]; then
        log_success "✅ 无时间断档 - 时间轴连续性保持完好"
    else
        log_error "❌ 发现 $gaps 个时间断档"
    fi
    
    # 检查尾部开放
    local tail_open=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "
    SELECT COUNT(*)
    FROM organization_units 
    WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' 
      AND status != 'DELETED'
      AND effective_date = (
          SELECT MAX(effective_date) 
          FROM organization_units 
          WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' 
            AND status != 'DELETED'
      )
      AND end_date IS NULL;
    " 2>/dev/null | xargs)
    
    if [ "$tail_open" = "1" ]; then
        log_success "✅ 尾部开放正确 - 最后版本end_date为NULL"
    else
        log_error "❌ 尾部开放检查失败"
    fi
    
    # 检查当前版本唯一性
    local current_count=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "
    SELECT COUNT(*) 
    FROM organization_units 
    WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' 
      AND status != 'DELETED'
      AND is_current = true;
    " 2>/dev/null | xargs)
    
    if [ "$current_count" = "1" ]; then
        log_success "✅ 当前版本唯一性正确"
    else
        log_error "❌ 当前版本数量异常: $current_count (应该为1)"
    fi
}

# API使用指导
provide_api_guidance() {
    log_info "📖 正确的API使用指导"
    echo ""
    echo "🔍 删除端点说明："
    echo "  ❌ 已移除：DELETE /api/v1/organization-units/{code}"
    echo "     - 物理删除端点已从API中移除"
    echo "     - 避免误用导致的时态时间轴不一致问题"
    echo ""
    echo "  ✅ 唯一删除方式：DELETE /api/v1/organization-units/versions/{recordId}"
    echo "     - 删除特定版本 (软删除)"
    echo "     - 自动触发时态时间轴重计算"
    echo "     - 适用于：删除错误版本、修正历史记录"
    echo ""
    echo "🎯 用户问题解决方案："
    echo "  1. 获取要删除版本的record_id"
    echo "  2. 使用 DELETE /versions/{recordId} 端点"
    echo "  3. 系统自动重新计算时间轴"
    echo "  4. 验证连续性（无断档、尾部开放、唯一当前版本）"
    echo ""
}

# 主流程
main() {
    echo "=================================================="
    echo "🚀 正确的版本删除API使用演示"
    echo "解决：九月二日记录删除后，九月一号记录结束日期没有自动更新"
    echo "=================================================="
    
    check_server
    
    query_timeline "初始状态"
    
    create_versions
    query_timeline "创建多版本后"
    
    demo_wrong_deletion
    
    demo_correct_deletion
    query_timeline "正确删除版本后"
    
    verify_timeline_consistency
    
    provide_api_guidance
    
    echo "=================================================="
    log_success "🎯 时态时间轴连续性验证完成！"
    echo "=================================================="
}

# 执行主流程
main "$@"
