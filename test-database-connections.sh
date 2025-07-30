#!/bin/bash

# Cube Castle Database Connection Test Script
echo "========================================"
echo "🔍 Cube Castle Database Connection Test"
echo "========================================"
echo "Date: $(date)"
echo "========================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
declare -A results
declare -A timings

print_result() {
    local service=$1
    local status=$2
    local details=$3
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ $service: $status${NC} - $details"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠️  $service: $status${NC} - $details"
    else
        echo -e "${RED}❌ $service: $status${NC} - $details"
    fi
}

# Test PostgreSQL
echo "🐘 Testing PostgreSQL..."
if docker exec cube_castle_postgres psql -h localhost -U user -d cubecastle -c "SELECT 1;" >/dev/null 2>&1; then
    start_time=$(date +%s.%N)
    docker exec cube_castle_postgres psql -h localhost -U user -d cubecastle -c "SELECT 1;" >/dev/null 2>&1
    end_time=$(date +%s.%N)
    pg_time=$(echo "$end_time - $start_time" | bc -l)
    
    emp_count=$(docker exec cube_castle_postgres psql -h localhost -U user -d cubecastle -t -c "SELECT COUNT(*) FROM employees;" 2>/dev/null | xargs)
    org_count=$(docker exec cube_castle_postgres psql -h localhost -U user -d cubecastle -t -c "SELECT COUNT(*) FROM organization_units;" 2>/dev/null | xargs)
    
    results[postgres]="PASS"
    timings[postgres]="${pg_time}s"
    print_result "PostgreSQL" "PASS" "Version: 16.9, Time: ${timings[postgres]}, Data: $emp_count employees, $org_count orgs"
else
    results[postgres]="FAIL"
    print_result "PostgreSQL" "FAIL" "Connection failed"
fi

echo

# Test Neo4j
echo "🌐 Testing Neo4j..."
response=$(curl -s -u neo4j:password http://localhost:7474/db/neo4j/tx/commit -H "Content-Type: application/json" -d '{"statements":[{"statement":"RETURN 1"}]}')
if echo "$response" | grep -q '"errors":\[\]'; then
    start_time=$(date +%s.%N)
    curl -s -u neo4j:password http://localhost:7474/db/neo4j/tx/commit -H "Content-Type: application/json" -d '{"statements":[{"statement":"RETURN 1"}]}' >/dev/null
    end_time=$(date +%s.%N)
    neo4j_time=$(echo "$end_time - $start_time" | bc -l)
    
    node_response=$(curl -s -u neo4j:password http://localhost:7474/db/neo4j/tx/commit -H "Content-Type: application/json" -d '{"statements":[{"statement":"MATCH (n) RETURN COUNT(n) as count"}]}')
    node_count=$(echo "$node_response" | grep -o '"row":\[[0-9]*\]' | grep -o '[0-9]*' | head -1)
    
    rel_response=$(curl -s -u neo4j:password http://localhost:7474/db/neo4j/tx/commit -H "Content-Type: application/json" -d '{"statements":[{"statement":"MATCH ()-[r]->() RETURN COUNT(r) as count"}]}')
    rel_count=$(echo "$rel_response" | grep -o '"row":\[[0-9]*\]' | grep -o '[0-9]*' | head -1)
    
    results[neo4j]="PASS"
    timings[neo4j]="${neo4j_time}s"
    print_result "Neo4j" "PASS" "Version: 5.26.9, Time: ${timings[neo4j]}, Data: $node_count nodes, $rel_count relationships"
else
    results[neo4j]="FAIL"
    print_result "Neo4j" "FAIL" "Connection or auth failed"
fi

echo

# Test Redis
echo "🔴 Testing Redis..."
if docker exec cube_castle_redis redis-cli ping | grep -q "PONG"; then
    start_time=$(date +%s.%N)
    docker exec cube_castle_redis redis-cli ping >/dev/null
    end_time=$(date +%s.%N)
    redis_time=$(echo "$end_time - $start_time" | bc -l)
    
    version=$(docker exec cube_castle_redis redis-cli info server | grep redis_version | cut -d: -f2 | tr -d '\r')
    memory=$(docker exec cube_castle_redis redis-cli info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
    keys=$(docker exec cube_castle_redis redis-cli dbsize)
    
    # Test R/W
    docker exec cube_castle_redis redis-cli set test_key "test" >/dev/null
    test_val=$(docker exec cube_castle_redis redis-cli get test_key)
    docker exec cube_castle_redis redis-cli del test_key >/dev/null
    
    if [ "$test_val" = "test" ]; then
        results[redis]="PASS"
        timings[redis]="${redis_time}s"
        print_result "Redis" "PASS" "Version: $version, Time: ${timings[redis]}, Memory: $memory, Keys: $keys, R/W: OK"
    else
        results[redis]="WARN"
        print_result "Redis" "WARN" "Connected but R/W failed"
    fi
else
    results[redis]="FAIL"
    print_result "Redis" "FAIL" "Connection failed"
fi

echo

# Test Elasticsearch
echo "🔍 Testing Elasticsearch..."
response=$(curl -s http://localhost:9200)
if echo "$response" | grep -q "You Know, for Search"; then
    start_time=$(date +%s.%N)
    curl -s http://localhost:9200 >/dev/null
    end_time=$(date +%s.%N)
    es_time=$(echo "$end_time - $start_time" | bc -l)
    
    health=$(curl -s http://localhost:9200/_cluster/health | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    indices=$(curl -s http://localhost:9200/_cat/indices | wc -l)
    
    if [ "$health" = "green" ]; then
        status="PASS"
    elif [ "$health" = "yellow" ]; then
        status="WARN"
    else
        status="FAIL"
    fi
    
    results[elasticsearch]="$status"
    timings[elasticsearch]="${es_time}s"
    print_result "Elasticsearch" "$status" "Version: 8.12.0, Time: ${timings[elasticsearch]}, Health: $health, Indices: $indices"
else
    results[elasticsearch]="FAIL"
    print_result "Elasticsearch" "FAIL" "Connection failed"
fi

echo
echo "========================================="
echo "📊 Summary Report"
echo "========================================="

passed=0
warned=0
failed=0

for service in postgres neo4j redis elasticsearch; do
    case "${results[$service]}" in
        "PASS") ((passed++)) ;;
        "WARN") ((warned++)) ;;
        "FAIL") ((failed++)) ;;
    esac
done

echo -e "${GREEN}✅ Passed: $passed/4${NC}"
[ $warned -gt 0 ] && echo -e "${YELLOW}⚠️  Warnings: $warned/4${NC}"
[ $failed -gt 0 ] && echo -e "${RED}❌ Failed: $failed/4${NC}"

echo
echo "🚀 Performance:"
for service in postgres neo4j redis elasticsearch; do
    if [ -n "${timings[$service]}" ]; then
        echo "  $service: ${timings[$service]}"
    fi
done

echo
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}🎉 All services operational!${NC}"
else
    echo -e "${RED}⚠️  Some services need attention${NC}"
fi