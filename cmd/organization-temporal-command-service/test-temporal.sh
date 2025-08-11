#!/bin/bash

# 快速测试脚本 - 时态管理API测试

echo "🧪 时态管理API测试脚本"
echo "============================"

# 服务健康检查
echo "1️⃣ 服务健康检查"
curl -s http://localhost:9093/health | jq '.status'

# 直接SQL测试时态查询
echo -e "\n2️⃣ 直接SQL测试时态查询"
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
SELECT tenant_id, code, 
       COALESCE(parent_code, '') as parent_code,
       name, unit_type, status, level, path, sort_order,
       COALESCE(description, '') as description,
       created_at, updated_at,
       COALESCE(effective_date, CURRENT_DATE) as effective_date,
       end_date,
       COALESCE(change_reason, '') as change_reason,
       COALESCE(is_current, false) as is_current
FROM organization_units 
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9' 
  AND code = '1000999'
  AND effective_date <= '2025-08-03'::date 
  AND (end_date IS NULL OR end_date > '2025-08-03'::date)
ORDER BY effective_date DESC;"

# API测试不同的时态查询
echo -e "\n3️⃣ API时态查询测试"

echo "测试 as_of_date=2025-08-03:"
curl -s "http://localhost:9093/api/v1/organization-units/1000999/temporal?as_of_date=2025-08-03"

echo -e "\n测试 as_of_date=2025-08-07:"
curl -s "http://localhost:9093/api/v1/organization-units/1000999/temporal?as_of_date=2025-08-07"

echo -e "\n测试 as_of_date=2025-08-11:"
curl -s "http://localhost:9093/api/v1/organization-units/1000999/temporal?as_of_date=2025-08-11"

echo -e "\n测试 include_history=true:"
curl -s "http://localhost:9093/api/v1/organization-units/1000999/temporal?include_history=true"

echo -e "\n测试完成 ✅"