#!/bin/bash

# 219C2Y Job Level 验证测试脚本
# 用于验证 Job Level API 必填字段验证修复

set -e

# 配置
ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

BASE_URL_COMMAND="${BASE_URL_COMMAND:-http://localhost:9090}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
TOKEN_FILE=".cache/dev.jwt"

# 日志输出
log_info() {
    echo "[$(date '+%Y-%m-%dT%H:%M:%S%z')] ℹ️  $1"
}

log_pass() {
    echo "[$(date '+%Y-%m-%dT%H:%M:%S%z')] ✅ $1"
}

log_fail() {
    echo "[$(date '+%Y-%m-%dT%H:%M:%S%z')] ❌ $1"
}

# 获取token
if [ ! -f "$TOKEN_FILE" ]; then
    log_info "Token file not found, generating..."
    make jwt-dev-mint > /dev/null 2>&1 || {
        log_fail "Failed to generate token"
        exit 1
    }
fi

TOKEN=$(cat "$TOKEN_FILE")
log_pass "Token loaded"

# 创建辅助Job Catalog数据
log_info "========== Creating Job Catalog Prerequisites =========="

# Job Family Group
JFG_CODE="JFG-219C2Y-$(date +%s%N | tail -c 6)"
log_info "Creating Job Family Group: $JFG_CODE"
JFG_RESP=$(curl -s -X POST "$BASE_URL_COMMAND/api/v1/job-family-groups" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"$JFG_CODE\",
    \"name\": \"Test Family Group for 219C2Y\",
    \"status\": \"ACTIVE\",
    \"effectiveDate\": \"2025-11-01\"
  }")

JFG=$(echo "$JFG_RESP" | jq -r '.data.Code // .data.code // empty' 2>/dev/null)
if [ -z "$JFG" ]; then
    log_fail "Failed to create Job Family Group"
    echo "$JFG_RESP" | jq '.' >> logs/219C2/validation.log
    exit 1
fi
log_pass "Created Job Family Group: $JFG"

# Job Family
JF_CODE="JF-219C2Y-$(date +%s%N | tail -c 6)"
log_info "Creating Job Family: $JF_CODE"
JF_RESP=$(curl -s -X POST "$BASE_URL_COMMAND/api/v1/job-families" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"$JF_CODE\",
    \"jobFamilyGroupCode\": \"$JFG\",
    \"name\": \"Test Family for 219C2Y\",
    \"status\": \"ACTIVE\",
    \"effectiveDate\": \"2025-11-01\"
  }")

JF=$(echo "$JF_RESP" | jq -r '.data.Code // .data.code // empty' 2>/dev/null)
if [ -z "$JF" ]; then
    log_fail "Failed to create Job Family"
    echo "$JF_RESP" | jq '.' >> logs/219C2/validation.log
    exit 1
fi
log_pass "Created Job Family: $JF"

# Job Role
JR_CODE="JR-219C2Y-$(date +%s%N | tail -c 6)"
log_info "Creating Job Role: $JR_CODE"
JR_RESP=$(curl -s -X POST "$BASE_URL_COMMAND/api/v1/job-roles" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"$JR_CODE\",
    \"jobFamilyCode\": \"$JF\",
    \"name\": \"Test Role for 219C2Y\",
    \"status\": \"ACTIVE\",
    \"effectiveDate\": \"2025-11-01\"
  }")

JR=$(echo "$JR_RESP" | jq -r '.data.Code // .data.code // empty' 2>/dev/null)
if [ -z "$JR" ]; then
    log_fail "Failed to create Job Role"
    echo "$JR_RESP" | jq '.' >> logs/219C2/validation.log
    exit 1
fi
log_pass "Created Job Role: $JR"

# ========== 验证测试 ==========
log_info ""
log_info "========== Job Level Validation Tests =========="

# Test 1: 缺少 name 字段 - 应该返回 400
log_info ""
log_info "Test 1: Missing 'name' field (expecting HTTP 400)"
TEST1_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL_COMMAND/api/v1/job-levels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"L-MISSING-NAME\",
    \"jobRoleCode\": \"$JR\",
    \"levelRank\": \"1\",
    \"status\": \"ACTIVE\",
    \"effectiveDate\": \"2025-11-01\"
  }")

TEST1_HTTP=$(echo "$TEST1_RESP" | grep "HTTP_STATUS" | cut -d: -f2)
TEST1_BODY=$(echo "$TEST1_RESP" | sed '$d')

echo "Payload: {\"code\":\"L-MISSING-NAME\",\"jobRoleCode\":\"$JR\",\"levelRank\":\"1\",\"status\":\"ACTIVE\",\"effectiveDate\":\"2025-11-01\"}" >> logs/219C2/validation.log
echo "HTTP Status: $TEST1_HTTP" >> logs/219C2/validation.log
echo "Response:" >> logs/219C2/validation.log
echo "$TEST1_BODY" | jq '.' >> logs/219C2/validation.log

if [ "$TEST1_HTTP" = "400" ]; then
    log_pass "Test 1 PASSED: Got HTTP 400 for missing name"
    ERROR_CODE=$(echo "$TEST1_BODY" | jq -r '.error.code // empty')
    echo "Error Code: $ERROR_CODE" >> logs/219C2/validation.log
else
    log_fail "Test 1 FAILED: Expected HTTP 400 but got $TEST1_HTTP"
    echo "$TEST1_BODY" | jq '.' >> logs/219C2/validation.log
fi

# Test 2: 缺少 status 字段 - 应该返回 400
log_info ""
log_info "Test 2: Missing 'status' field (expecting HTTP 400)"
TEST2_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL_COMMAND/api/v1/job-levels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"L-MISSING-STATUS\",
    \"jobRoleCode\": \"$JR\",
    \"name\": \"Test Level\",
    \"levelRank\": \"1\",
    \"effectiveDate\": \"2025-11-01\"
  }")

TEST2_HTTP=$(echo "$TEST2_RESP" | grep "HTTP_STATUS" | cut -d: -f2)
TEST2_BODY=$(echo "$TEST2_RESP" | sed '$d')

echo "" >> logs/219C2/validation.log
echo "Payload: {\"code\":\"L-MISSING-STATUS\",\"jobRoleCode\":\"$JR\",\"name\":\"Test Level\",\"levelRank\":\"1\",\"effectiveDate\":\"2025-11-01\"}" >> logs/219C2/validation.log
echo "HTTP Status: $TEST2_HTTP" >> logs/219C2/validation.log
echo "Response:" >> logs/219C2/validation.log
echo "$TEST2_BODY" | jq '.' >> logs/219C2/validation.log

if [ "$TEST2_HTTP" = "400" ]; then
    log_pass "Test 2 PASSED: Got HTTP 400 for missing status"
else
    log_fail "Test 2 FAILED: Expected HTTP 400 but got $TEST2_HTTP"
fi

# Test 3: 缺少 levelRank 字段 - 应该返回 400
log_info ""
log_info "Test 3: Missing 'levelRank' field (expecting HTTP 400)"
TEST3_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL_COMMAND/api/v1/job-levels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"L-MISSING-RANK\",
    \"jobRoleCode\": \"$JR\",
    \"name\": \"Test Level\",
    \"status\": \"ACTIVE\",
    \"effectiveDate\": \"2025-11-01\"
  }")

TEST3_HTTP=$(echo "$TEST3_RESP" | grep "HTTP_STATUS" | cut -d: -f2)
TEST3_BODY=$(echo "$TEST3_RESP" | sed '$d')

echo "" >> logs/219C2/validation.log
echo "Payload: {\"code\":\"L-MISSING-RANK\",\"jobRoleCode\":\"$JR\",\"name\":\"Test Level\",\"status\":\"ACTIVE\",\"effectiveDate\":\"2025-11-01\"}" >> logs/219C2/validation.log
echo "HTTP Status: $TEST3_HTTP" >> logs/219C2/validation.log
echo "Response:" >> logs/219C2/validation.log
echo "$TEST3_BODY" | jq '.' >> logs/219C2/validation.log

if [ "$TEST3_HTTP" = "400" ]; then
    log_pass "Test 3 PASSED: Got HTTP 400 for missing levelRank"
else
    log_fail "Test 3 FAILED: Expected HTTP 400 but got $TEST3_HTTP"
fi

# Test 4: 提供完整字段 - 应该返回 201
log_info ""
log_info "Test 4: Complete request with all required fields (expecting HTTP 201)"
JL_CODE="JL-219C2Y-$(date +%s%N | tail -c 6)"
TEST4_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL_COMMAND/api/v1/job-levels" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"$JL_CODE\",
    \"jobRoleCode\": \"$JR\",
    \"name\": \"Test Level Complete\",
    \"levelRank\": \"1\",
    \"status\": \"ACTIVE\",
    \"effectiveDate\": \"2025-11-01\"
  }")

TEST4_HTTP=$(echo "$TEST4_RESP" | grep "HTTP_STATUS" | cut -d: -f2)
TEST4_BODY=$(echo "$TEST4_RESP" | sed '$d')

echo "" >> logs/219C2/validation.log
echo "Payload: {\"code\":\"$JL_CODE\",\"jobRoleCode\":\"$JR\",\"name\":\"Test Level Complete\",\"levelRank\":\"1\",\"status\":\"ACTIVE\",\"effectiveDate\":\"2025-11-01\"}" >> logs/219C2/validation.log
echo "HTTP Status: $TEST4_HTTP" >> logs/219C2/validation.log
echo "Response:" >> logs/219C2/validation.log
echo "$TEST4_BODY" | jq '.' >> logs/219C2/validation.log

if [ "$TEST4_HTTP" = "201" ]; then
    log_pass "Test 4 PASSED: Got HTTP 201 and successfully created Job Level"
    JL_CODE_RESULT=$(echo "$TEST4_BODY" | jq -r '.data.Code // .data.code // empty')
    echo "Created Job Level: $JL_CODE_RESULT" >> logs/219C2/validation.log
else
    log_fail "Test 4 FAILED: Expected HTTP 201 but got $TEST4_HTTP"
fi

# 总结
log_info ""
log_info "========== Test Summary =========="
PASS_COUNT=0
[ "$TEST1_HTTP" = "400" ] && ((PASS_COUNT++)) || true
[ "$TEST2_HTTP" = "400" ] && ((PASS_COUNT++)) || true
[ "$TEST3_HTTP" = "400" ] && ((PASS_COUNT++)) || true
[ "$TEST4_HTTP" = "201" ] && ((PASS_COUNT++)) || true

if [ $PASS_COUNT -eq 4 ]; then
    log_pass "All 4 tests PASSED!"
    echo "" >> logs/219C2/validation.log
    echo "========== Test Summary ==========" >> logs/219C2/validation.log
    echo "✅ All 4 tests PASSED!" >> logs/219C2/validation.log
    echo "- Test 1 (missing name): HTTP 400 ✅" >> logs/219C2/validation.log
    echo "- Test 2 (missing status): HTTP 400 ✅" >> logs/219C2/validation.log
    echo "- Test 3 (missing levelRank): HTTP 400 ✅" >> logs/219C2/validation.log
    echo "- Test 4 (complete request): HTTP 201 ✅" >> logs/219C2/validation.log
    exit 0
else
    log_fail "$PASS_COUNT/4 tests PASSED"
    echo "" >> logs/219C2/validation.log
    echo "========== Test Summary ==========" >> logs/219C2/validation.log
    echo "Tests Passed: $PASS_COUNT/4" >> logs/219C2/validation.log
    exit 1
fi

