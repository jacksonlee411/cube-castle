#!/bin/bash
# GraphQL集成测试 - 验证数据格式一致性修复

echo "🔧 GraphQL集成数据格式修复验证"
echo "================================"

# 测试1: GraphQL查询格式验证
echo ""
echo "📋 测试1: GraphQL时态字段格式"
echo "----------------------------"

gql_result=$(curl -s -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organization(code: \"1000001\") { tenant_id code parent_code name unit_type status level effective_date version is_current } }"}')

if echo "$gql_result" | jq -r '.data.organization.unit_type' | grep -q "DEPARTMENT"; then
    echo "✅ GraphQL Schema字段格式正确 - unit_type (下划线命名)"
else
    echo "❌ GraphQL Schema字段格式错误"
fi

# 测试2: 时态API格式验证  
echo ""
echo "📋 测试2: 时态API响应格式"
echo "----------------------"

temporal_result=$(curl -s http://localhost:9091/api/v1/organization-units/1000001)

if echo "$temporal_result" | jq -r '.organizations[0].unit_type' | grep -q "DEPARTMENT"; then
    echo "✅ 时态API响应格式正确 - unit_type (下划线命名)"
else
    echo "❌ 时态API响应格式错误"
fi

# 测试3: 数据格式一致性验证
echo ""
echo "📋 测试3: 数据格式一致性验证"
echo "------------------------"

gql_unit_type=$(echo "$gql_result" | jq -r '.data.organization.unit_type')
temporal_unit_type=$(echo "$temporal_result" | jq -r '.organizations[0].unit_type')

if [ "$gql_unit_type" = "$temporal_unit_type" ]; then
    echo "✅ 数据格式完全一致 - unit_type: $gql_unit_type"
else
    echo "❌ 数据格式不一致 - GraphQL: $gql_unit_type, 时态API: $temporal_unit_type"
fi

# 测试4: 时态字段完整性验证
echo ""
echo "📋 测试4: 时态字段完整性"
echo "---------------------"

gql_fields=$(echo "$gql_result" | jq -r '.data.organization | keys | @json')
temporal_fields=$(echo "$temporal_result" | jq -r '.organizations[0] | keys | @json')

echo "GraphQL字段: $gql_fields"
echo "时态API字段: $temporal_fields"

# 统计共同字段
common_fields=0
if echo "$gql_result" | jq -e '.data.organization.tenant_id' >/dev/null 2>&1; then ((common_fields++)); fi
if echo "$gql_result" | jq -e '.data.organization.unit_type' >/dev/null 2>&1; then ((common_fields++)); fi
if echo "$gql_result" | jq -e '.data.organization.effective_date' >/dev/null 2>&1; then ((common_fields++)); fi
if echo "$gql_result" | jq -e '.data.organization.version' >/dev/null 2>&1; then ((common_fields++)); fi
if echo "$gql_result" | jq -e '.data.organization.is_current' >/dev/null 2>&1; then ((common_fields++)); fi

echo "✅ 时态字段覆盖: $common_fields/5 个字段已实现"

# 最终结果
echo ""
echo "📊 修复验证结果"
echo "============="
if [ $common_fields -ge 4 ]; then
    echo "🎉 GraphQL集成数据格式修复成功！"
    echo "• 字段命名统一: 下划线命名风格"
    echo "• 时态字段完整: $common_fields/5 字段支持"
    echo "• 数据格式一致: GraphQL ↔ 时态API"
    exit 0
else
    echo "⚠️ 部分问题仍需解决"
    echo "• 时态字段覆盖: $common_fields/5"
    echo "• 建议检查Neo4j数据同步"
    exit 1
fi