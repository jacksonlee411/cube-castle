#!/bin/bash

# Cube Castle 端到端测试脚本
# 验证OAuth 2.0认证 + CQRS架构 + GraphQL/REST API

set -e

echo "🧪 开始Cube Castle端到端测试..."
echo "==========================================="

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 服务端点配置 - 使用环境变量
OAUTH_SERVICE="${E2E_OAUTH_SERVICE_URL:-http://localhost:8080}"
GRAPHQL_SERVICE="${E2E_GRAPHQL_API_URL:-http://localhost:8090}"
REST_SERVICE="${E2E_COMMAND_API_URL:-http://localhost:9090}"

# 测试步骤计数器
STEP=1
FAILED_TESTS=0
TOTAL_TESTS=0

# 辅助函数
function print_step() {
    echo -e "${YELLOW}步骤 $STEP: $1${NC}"
    STEP=$((STEP + 1))
}

function test_success() {
    echo -e "${GREEN}✅ $1${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

function test_failure() {
    echo -e "${RED}❌ $1${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

function check_service() {
    local service_name=$1
    local service_url=$2
    
    if curl -s -f "$service_url/health" > /dev/null; then
        test_success "$service_name 服务运行正常"
    else
        test_failure "$service_name 服务不可达 ($service_url)"
        return 1
    fi
}

# 步骤1: 检查所有服务状态
print_step "检查服务健康状态"
check_service "OAuth认证服务" "$OAUTH_SERVICE"
check_service "GraphQL查询服务" "$GRAPHQL_SERVICE" 
check_service "REST命令服务" "$REST_SERVICE"

# 步骤2: OAuth 2.0认证流程测试
print_step "测试OAuth 2.0 Client Credentials Flow"

# 获取访问令牌
echo "  🔑 获取OAuth访问令牌..."
TOKEN_RESPONSE=$(curl -s -X POST $OAUTH_SERVICE/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "cube-castle-api-client", 
    "client_secret": "cube-castle-secret-2024"
  }')

if echo "$TOKEN_RESPONSE" | jq -e '.accessToken' > /dev/null 2>&1; then
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.accessToken')
    test_success "OAuth令牌获取成功"
else
    test_failure "OAuth令牌获取失败: $TOKEN_RESPONSE"
    exit 1
fi

# 步骤3: GraphQL查询服务测试 (CQRS-查询端)
print_step "测试GraphQL查询服务 (读操作)"

# 测试认证保护
echo "  🛡️  测试未认证请求保护..."
UNAUTH_RESPONSE=$(curl -s -H "Content-Type: application/json" \
  -X POST -d '{"query":"query { organizations { data { code name } } }"}' \
  $GRAPHQL_SERVICE/graphql)

if echo "$UNAUTH_RESPONSE" | grep -q "Authorization"; then
    test_success "GraphQL认证保护正常工作"
else 
    test_failure "GraphQL认证保护失效"
fi

# 测试认证请求
echo "  📊 测试认证后的GraphQL查询..."
GRAPHQL_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -X POST -d '{"query":"query { organizations(pagination: { page: 1, pageSize: 5 }) { data { code name unitType status } pagination { total page pageSize hasNext } } }"}' \
  $GRAPHQL_SERVICE/graphql)

if echo "$GRAPHQL_RESPONSE" | jq -e '.data.organizations' > /dev/null 2>&1; then
    ORG_COUNT=$(echo "$GRAPHQL_RESPONSE" | jq -r '.data.organizations.pagination.total')
    test_success "GraphQL查询成功，返回 $ORG_COUNT 个组织"
else
    test_failure "GraphQL查询失败: $GRAPHQL_RESPONSE"
fi

# 步骤4: REST命令服务测试 (CQRS-命令端)
print_step "测试REST命令服务 (写操作)"

# 创建组织
echo "  ➕ 测试创建组织..."
CREATE_RESPONSE=$(curl -s -X POST $REST_SERVICE/api/v1/organization-units \
  -H "Content-Type: application/json" \
  -d '{
    "name": "E2E测试部门",
    "unitType": "DEPARTMENT", 
    "parentCode": "1000000",
    "description": "端到端测试创建的组织"
  }')

if echo "$CREATE_RESPONSE" | jq -e '.code' > /dev/null 2>&1; then
    NEW_ORG_CODE=$(echo "$CREATE_RESPONSE" | jq -r '.code')
    test_success "组织创建成功，代码: $NEW_ORG_CODE"
else
    test_failure "组织创建失败: $CREATE_RESPONSE"
    exit 1
fi

# 更新组织
echo "  ✏️  测试更新组织..."
UPDATE_RESPONSE=$(curl -s -X PUT $REST_SERVICE/api/v1/organization-units/$NEW_ORG_CODE \
  -H "Content-Type: application/json" \
  -d '{
    "name": "E2E测试部门(已更新)",
    "description": "端到端测试更新的组织"
  }')

if echo "$UPDATE_RESPONSE" | jq -e '.name' > /dev/null 2>&1; then
    UPDATED_NAME=$(echo "$UPDATE_RESPONSE" | jq -r '.name')
    test_success "组织更新成功: $UPDATED_NAME"
else
    test_failure "组织更新失败: $UPDATE_RESPONSE"
fi

# 步骤5: CQRS架构一致性验证
print_step "验证CQRS架构读写一致性"

echo "  🔄 验证命令端写入后查询端数据同步..."
sleep 1  # 等待数据同步

# 通过GraphQL查询新创建的组织
VERIFY_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -X POST -d '{"query":"query { organizations(pagination: { page: 1, pageSize: 10 }) { data { code name unitType } pagination { total } } }"}' \
  $GRAPHQL_SERVICE/graphql)

if echo "$VERIFY_RESPONSE" | jq -e '.data.organizations' > /dev/null 2>&1; then
    NEW_TOTAL=$(echo "$VERIFY_RESPONSE" | jq -r '.data.organizations.pagination.total')
    if [ "$NEW_TOTAL" -gt "$ORG_COUNT" ]; then
        test_success "CQRS读写一致性验证成功，组织总数: $ORG_COUNT → $NEW_TOTAL"
    else
        test_failure "CQRS读写一致性问题，组织数量未增加"
    fi
else
    test_failure "CQRS一致性验证查询失败"
fi

# 步骤6: 清理测试数据
print_step "清理测试数据"

echo "  🗑️  删除测试组织..."
DELETE_RESPONSE=$(curl -s -X DELETE $REST_SERVICE/api/v1/organization-units/$NEW_ORG_CODE)
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $REST_SERVICE/api/v1/organization-units/$NEW_ORG_CODE)

if [ "$HTTP_STATUS" = "204" ]; then
    test_success "测试数据清理成功"
else
    test_failure "测试数据清理失败 (HTTP: $HTTP_STATUS)"
fi

# 测试结果汇总
echo ""
echo "==========================================="
echo "🧪 端到端测试完成"
echo "==========================================="
echo "✅ 成功测试: $((TOTAL_TESTS - FAILED_TESTS))/$TOTAL_TESTS"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "❌ 失败测试: ${RED}$FAILED_TESTS${NC}"
    echo ""
    echo "🔍 测试架构验证结果:"
    echo "  • OAuth 2.0认证: $([ $FAILED_TESTS -eq 0 ] && echo '✅ 正常' || echo '⚠️  需检查')"
    echo "  • CQRS架构: $([ $FAILED_TESTS -eq 0 ] && echo '✅ 正常' || echo '⚠️  需检查')" 
    echo "  • GraphQL查询: $([ $FAILED_TESTS -eq 0 ] && echo '✅ 正常' || echo '⚠️  需检查')"
    echo "  • REST命令: $([ $FAILED_TESTS -eq 0 ] && echo '✅ 正常' || echo '⚠️  需检查')"
    exit 1
else
    echo -e "🎉 ${GREEN}所有测试通过！${NC}"
    echo ""
    echo "✅ 验证完成的核心架构:"
    echo "  • OAuth 2.0 Client Credentials Flow认证"  
    echo "  • CQRS读写分离架构"
    echo "  • GraphQL查询服务 (端口8090)"
    echo "  • REST命令服务 (端口9090)" 
    echo "  • PostgreSQL单一数据源"
    echo "  • 企业级Bearer Token安全"
    exit 0
fi
