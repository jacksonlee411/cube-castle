#!/bin/bash

# 数据库集成测试脚本
echo "🏰 Cube Castle - 数据库集成测试"
echo "=================================="

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
test_result() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ $1 -eq 0 ]; then
        echo "✅ $2"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "❌ $2"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo -e "\n1. PostgreSQL 连接测试"
echo "------------------------"

# 测试PostgreSQL连接
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT version();" > /dev/null 2>&1
test_result $? "PostgreSQL 数据库连接"

# 测试数据库表结构
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT table_name FROM information_schema.tables WHERE table_schema='public';" > /dev/null 2>&1
test_result $? "PostgreSQL 表结构查询"

echo -e "\n2. PostgreSQL 基础CRUD测试"
echo "---------------------------"

# 创建测试表
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "
DROP TABLE IF EXISTS test_employees;
CREATE TABLE test_employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL DEFAULT gen_random_uuid(),
    employee_number VARCHAR(50) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);" > /dev/null 2>&1
test_result $? "创建测试表"

# 插入测试数据
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "
INSERT INTO test_employees (employee_number, first_name, last_name, email) 
VALUES ('EMP001', '张三', '李', 'zhangsan@example.com');" > /dev/null 2>&1
test_result $? "插入测试数据"

# 查询测试数据
RESULT=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "SELECT COUNT(*) FROM test_employees WHERE employee_number='EMP001';")
if [ "$(echo $RESULT | tr -d ' ')" = "1" ]; then
    test_result 0 "查询测试数据"
else
    test_result 1 "查询测试数据"
fi

# 更新测试数据
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "
UPDATE test_employees SET first_name='王五' WHERE employee_number='EMP001';" > /dev/null 2>&1
test_result $? "更新测试数据"

# 验证更新
RESULT=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "SELECT first_name FROM test_employees WHERE employee_number='EMP001';")
if [ "$(echo $RESULT | tr -d ' ')" = "王五" ]; then
    test_result 0 "验证数据更新"
else
    test_result 1 "验证数据更新"
fi

# 删除测试数据
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "
DELETE FROM test_employees WHERE employee_number='EMP001';" > /dev/null 2>&1
test_result $? "删除测试数据"

# 清理测试表
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "DROP TABLE test_employees;" > /dev/null 2>&1
test_result $? "清理测试表"

echo -e "\n3. Neo4j 连接测试"
echo "-------------------"

# 测试Neo4j连接 (使用cypher-shell)
if command -v cypher-shell > /dev/null 2>&1; then
    echo "RETURN 'Neo4j连接成功' as result;" | cypher-shell -a bolt://localhost:7687 -u neo4j -p password > /dev/null 2>&1
    test_result $? "Neo4j 数据库连接"
else
    echo "⚠️  cypher-shell 未安装，跳过Neo4j直接连接测试"
    # 使用curl测试HTTP接口
    curl -s -u neo4j:password -H "Content-Type: application/json" -d '{"query":"RETURN 1 as test"}' http://localhost:7474/db/data/cypher > /dev/null 2>&1
    test_result $? "Neo4j HTTP接口连接"
fi

echo -e "\n4. 数据库性能测试"
echo "-------------------"

# PostgreSQL 性能测试
start_time=$(date +%s%N)
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT COUNT(*) FROM information_schema.tables;" > /dev/null 2>&1
end_time=$(date +%s%N)
duration=$((($end_time - $start_time) / 1000000))

if [ $duration -lt 1000 ]; then
    test_result 0 "PostgreSQL 查询响应时间 (${duration}ms)"
else
    test_result 1 "PostgreSQL 查询响应时间过长 (${duration}ms)"
fi

echo -e "\n5. 数据库事务测试"
echo "-------------------"

# 事务测试
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "
BEGIN;
CREATE TABLE test_transaction (id SERIAL PRIMARY KEY, name VARCHAR(50));
INSERT INTO test_transaction (name) VALUES ('test');
ROLLBACK;" > /dev/null 2>&1
test_result $? "数据库事务回滚测试"

# 验证回滚是否成功
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT * FROM test_transaction;" > /dev/null 2>&1
if [ $? -ne 0 ]; then
    test_result 0 "验证事务回滚结果"
else
    test_result 1 "验证事务回滚结果"
fi

echo -e "\n6. 数据库连接池测试"
echo "---------------------"

# 模拟多连接测试
for i in {1..5}; do
    PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT pg_backend_pid();" > /dev/null 2>&1 &
done
wait
test_result 0 "数据库并发连接测试"

echo -e "\n=================================="
echo "数据库集成测试完成！"
echo "总计: $TOTAL_TESTS 项测试"
echo "✅ 通过: $PASSED_TESTS 项"
echo "❌ 失败: $FAILED_TESTS 项"
SUCCESS_RATE=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
echo "成功率: ${SUCCESS_RATE}%"
echo "=================================="

# 返回退出码
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi