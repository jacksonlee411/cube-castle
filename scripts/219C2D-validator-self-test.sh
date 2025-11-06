#!/usr/bin/env bash
#
# 219C2D Validator 自测脚本
# 目标：
#   1. 覆盖 Job Catalog JC-* 规则（JC-TEMPORAL / JC-SEQUENCE）正反场景
#   2. 覆盖 Position / Assignment 关键命令（Fill / Close）正反场景，补齐 POS-HEADCOUNT、ASSIGN-STATE 等验证证据
#   3. 生成 REST 与 GraphQL 双通道记录，并输出报告至 tests/e2e/organization-validator/report-Day24.json

set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

BASE_URL_COMMAND="${BASE_URL_COMMAND:-http://localhost:9090}"
BASE_URL_QUERY="${BASE_URL_QUERY:-http://localhost:8090}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
TOKEN_FILE=".cache/dev.jwt"
VALIDATION_LOG="$ROOT_DIR/logs/219C2/validation.log"
REPORT_DIR="$ROOT_DIR/tests/e2e/organization-validator"
REPORT_TMP="$(mktemp)"
TIMESTAMP="$(date '+%Y-%m-%dT%H:%M:%S%z')"

mkdir -p "$(dirname "$VALIDATION_LOG")" "$REPORT_DIR"

print_log() {
  printf '%s %s\n' "[$(date '+%Y-%m-%dT%H:%M:%S%z')]" "$1" | tee -a "$VALIDATION_LOG" >/dev/null
}

section_log() {
  printf '\n========== %s ==========\n' "$1" | tee -a "$VALIDATION_LOG" >/dev/null
}

ensure_token() {
  if [[ ! -f "$TOKEN_FILE" ]]; then
    print_log "令牌不存在，执行 make jwt-dev-mint ..."
    make jwt-dev-mint >/dev/null
  fi
  TOKEN="$(cat "$TOKEN_FILE")"
}

rest_request() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local curl_args=(-sS -w '\n%{http_code}' -X "$method" "$BASE_URL_COMMAND$path"
    -H "Content-Type: application/json"
    -H "Authorization: Bearer $TOKEN"
    -H "X-Tenant-ID: $TENANT_ID")
  if [[ -n "$data" ]]; then
    curl_args+=(-d "$data")
  fi
  curl "${curl_args[@]}"
}

graphql_request() {
  local query="$1"
  local variables_json="{}"
  if [[ $# -ge 2 && -n "${2:-}" ]]; then
    variables_json="$2"
  fi

  local body
  if ! body=$(jq -n --arg q "$query" --arg vars "$variables_json" '
    (try ($vars | fromjson) catch {}) as $parsed
    | {query: $q, variables: $parsed}
  '); then
    print_log "⚠️  GraphQL variables JSON 解析失败，使用空变量。"
    body=$(jq -n --arg q "$query" '{query: $q, variables: {}}')
  fi

  curl -sS -w '\n%{http_code}' -X POST "$BASE_URL_QUERY/graphql" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -d "$body"
}

append_report() {
  local json="$1"
  echo "$json" >>"$REPORT_TMP"
}

finalize_report() {
  jq -s '.' "$REPORT_TMP" >"$REPORT_DIR/report-Day24.json"
  rm -f "$REPORT_TMP"
}

extract_field() {
  local body="$1"
  local filter="$2"
  echo "$body" | jq -r "$filter // empty" 2>/dev/null || true
}

ensure_token

section_log "219C2D Validator 自测开始 ($TIMESTAMP)"

# ---------- 准备测试数据 ----------

print_log "准备基础数据（组织 + Job Catalog + 职位）"

random_suffix="$(date +%s%N | tail -c 6)"
FILL_EFFECTIVE_DATE="${FILL_EFFECTIVE_DATE:-$(date '+%Y-%m-%d')}"

# 1. 创建组织
ORG_CODE="1$(printf '%06d' $((RANDOM % 900000 + 100000)))"
org_payload=$(jq -n \
  --arg code "$ORG_CODE" \
  '{code:$code,name:"219C2D 测试组织",unitType:"DEPARTMENT",operationReason:"219C2D validator self-test"}')
org_resp=$(rest_request POST "/api/v1/organization-units" "$org_payload")
org_status=$(echo "$org_resp" | tail -n1)
org_body=$(echo "$org_resp" | head -n -1)
if [[ "$org_status" != "201" ]]; then
  print_log "❌ 创建组织失败: $org_body"
  exit 1
fi
print_log "✅ 组织创建成功 (code=$ORG_CODE)"

# 2. 创建 Job Catalog 层级
JFG_CODE="JFG-219C2D-$random_suffix"
JF_CODE="JF-219C2D-$random_suffix"
JR_CODE="JR-219C2D-$random_suffix"
JL_CODE="JL-219C2D-$random_suffix"
EFFECTIVE_BASE="2025-11-01"

create_job_family_group() {
  local payload=$(jq -n \
    --arg code "$JFG_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,name:"219C2D 自测职类",status:"ACTIVE",effectiveDate:$effective}')
  local resp=$(rest_request POST "/api/v1/job-family-groups" "$payload")
  local status=$(echo "$resp" | tail -n1)
  local body=$(echo "$resp" | head -n -1)
  if [[ "$status" != "201" ]]; then
    print_log "❌ 创建职类失败: $body"
    exit 1
  fi
  JFG_RECORD_ID=$(extract_field "$body" '.data.RecordID // .data.recordId')
  print_log "✅ 职类创建成功 (code=$JFG_CODE, recordId=$JFG_RECORD_ID)"
}

create_job_family() {
  local payload=$(jq -n \
    --arg code "$JF_CODE" \
    --arg group "$JFG_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,jobFamilyGroupCode:$group,name:"219C2D 自测职种",status:"ACTIVE",effectiveDate:$effective}')
  local resp=$(rest_request POST "/api/v1/job-families" "$payload")
  local status=$(echo "$resp" | tail -n1)
  local body=$(echo "$resp" | head -n -1)
  if [[ "$status" != "201" ]]; then
    print_log "❌ 创建职种失败: $body"
    exit 1
  fi
  JF_RECORD_ID=$(extract_field "$body" '.data.RecordID // .data.recordId')
  print_log "✅ 职种创建成功 (code=$JF_CODE, recordId=$JF_RECORD_ID)"
}

create_job_role() {
  local payload=$(jq -n \
    --arg code "$JR_CODE" \
    --arg family "$JF_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,jobFamilyCode:$family,name:"219C2D 自测职务",status:"ACTIVE",effectiveDate:$effective}')
  local resp=$(rest_request POST "/api/v1/job-roles" "$payload")
  local status=$(echo "$resp" | tail -n1)
  local body=$(echo "$resp" | head -n -1)
  if [[ "$status" != "201" ]]; then
    print_log "❌ 创建职务失败: $body"
    exit 1
  fi
  JR_RECORD_ID=$(extract_field "$body" '.data.RecordID // .data.recordId')
  print_log "✅ 职务创建成功 (code=$JR_CODE, recordId=$JR_RECORD_ID)"
}

create_job_level() {
  local payload=$(jq -n \
    --arg code "$JL_CODE" \
    --arg role "$JR_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,jobRoleCode:$role,levelRank:"1",name:"219C2D 自测职级",status:"ACTIVE",effectiveDate:$effective}')
  local resp=$(rest_request POST "/api/v1/job-levels" "$payload")
  local status=$(echo "$resp" | tail -n1)
  local body=$(echo "$resp" | head -n -1)
  if [[ "$status" != "201" ]]; then
    print_log "❌ 创建职级失败: $body"
    exit 1
  fi
  JL_RECORD_ID=$(extract_field "$body" '.data.RecordID // .data.recordId')
  print_log "✅ 职级创建成功 (code=$JL_CODE, recordId=$JL_RECORD_ID)"
}

create_job_family_group
create_job_family
create_job_role
create_job_level

# 3. 创建职位
create_position() {
  local payload=$(jq -n \
    --arg title "219C2D Validator Position" \
    --arg group "$JFG_CODE" \
    --arg family "$JF_CODE" \
    --arg role "$JR_CODE" \
    --arg level "$JL_CODE" \
    --arg org "$ORG_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{title:$title,jobFamilyGroupCode:$group,jobFamilyCode:$family,jobRoleCode:$role,jobLevelCode:$level,
      organizationCode:$org,positionType:"REGULAR",employmentType:"FULL_TIME",
      headcountCapacity:1,operationReason:"219C2D validator self-test",effectiveDate:$effective}')
  local resp=$(rest_request POST "/api/v1/positions" "$payload")
  local status=$(echo "$resp" | tail -n1)
  local body=$(echo "$resp" | head -n -1)
  if [[ "$status" != "201" ]]; then
    print_log "❌ 创建职位失败: $body"
    exit 1
  fi
  POSITION_CODE=$(extract_field "$body" '.data.code // .data.Code')
  POSITION_RECORD_ID=$(extract_field "$body" '.data.recordId // .data.RecordID')
  print_log "✅ 职位创建成功 (code=$POSITION_CODE, recordId=$POSITION_RECORD_ID)"
  POSITION_CREATE_PAYLOAD="$payload"
}

create_position

# ---------- 场景执行 ----------

log_rest_scenario() {
  local command="$1"
  local scenario="$2"
  local http_status="$3"
  local body="$4"
  local expected_status="${5:-}"
  local expected_code="${6:-}"
  local expected_rule="${7:-}"
  local expected_severity="${8:-}"

  local error_code request_id rule_id severity pass="passed"
  error_code=$(extract_field "$body" '.error.code // .error.Code // empty')
  request_id=$(extract_field "$body" '.requestId // empty')
  rule_id=$(extract_field "$body" '.error.details.ruleId // .error.details.validationErrors[0].context.ruleId // empty')
  severity=$(extract_field "$body" '.error.details.validationErrors[0].severity // empty')

  if [[ -n "$expected_status" && "$http_status" != "$expected_status" ]]; then
    pass="failed"
    print_log "⚠️  [$command/$scenario] 期望 HTTP $expected_status 实际 $http_status"
  fi
  if [[ -n "$expected_code" && "$error_code" != "$expected_code" ]]; then
    pass="failed"
    print_log "⚠️  [$command/$scenario] 期望错误码 $expected_code 实际 ${error_code:-<空>}"
  fi
  if [[ -n "$expected_rule" && "$rule_id" != "$expected_rule" ]]; then
    pass="failed"
    print_log "⚠️  [$command/$scenario] 期望 ruleId $expected_rule 实际 ${rule_id:-<空>}"
  fi
  if [[ -n "$expected_severity" && "${severity^^}" != "${expected_severity^^}" ]]; then
    pass="failed"
    print_log "⚠️  [$command/$scenario] 期望 severity $expected_severity 实际 ${severity:-<空>}"
  fi

  append_report "$(jq -n \
    --arg command "$command" \
    --arg scenario "$scenario" \
    --arg channel "REST" \
    --arg status "$http_status" \
    --arg err "$error_code" \
    --arg rid "$rule_id" \
    --arg req "$request_id" \
    --arg sev "$severity" \
    --arg pass "$pass" \
    '{command:$command,scenario:$scenario,channel:$channel,
      result:{
        status: $pass,
        httpStatus: ($status|tonumber?),
        errorCode: ($err|select(.!="")),
        ruleId: ($rid|select(.!="")),
        severity: ($sev|select(.!="")),
        requestId: ($req|select(.!=""))
      }}')"
}

log_graphql_scenario() {
  local command="$1"
  local scenario="$2"
  local description="$3"
  local http_status="$4"
  append_report "$(jq -n \
    --arg command "$command" \
    --arg scenario "$scenario" \
    --arg channel "GraphQL" \
    --arg desc "$description" \
    --arg status "$http_status" \
    '{command:$command,scenario:$scenario,channel:$channel,
      result:{
        status: (if (($status|tonumber?) // 500) < 400 then "passed" else "failed" end),
        httpStatus: ($status|tonumber?),
        note: $desc
      }}')"
}

# --- Command A: Job Catalog Version ---

catalog_parent_id="$JFG_RECORD_ID"

# A1: 成功创建新版本
section_log "A1. Job Catalog Version - 成功"
payload=$(jq -n \
  --arg name "219C2D 自测职类版本" \
  --arg status "ACTIVE" \
  --arg effective "2025-12-01" \
  --arg parent "$catalog_parent_id" \
  '{name:$name,status:$status,effectiveDate:$effective,parentRecordId:$parent}')
resp=$(rest_request POST "/api/v1/job-family-groups/$JFG_CODE/versions" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[A1] Payload: $payload" >>"$VALIDATION_LOG"
echo "[A1] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
if [[ "$status_code" != "201" ]]; then
  print_log "❌ A1 失败"
  exit 1
fi
catalog_latest_record=$(extract_field "$body" '.data.RecordID // .data.recordId')
log_rest_scenario "jobCatalog.createVersion" "success" "$status_code" "$body" "201"

graphql_query='query ($code: JobFamilyGroupCode!) {
  jobFamilies(groupCode: $code) {
    code
    status
    effectiveDate
    endDate
  }
}'
graphql_resp=$(graphql_request "$graphql_query" "$(jq -n --arg code "$JFG_CODE" '{code:$code}')" )
graphql_status=$(echo "$graphql_resp" | tail -n1)
graphql_body=$(echo "$graphql_resp" | head -n -1)
echo "[A1] GraphQL jobFamilies response:" >>"$VALIDATION_LOG"
echo "$graphql_body" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "jobCatalog.createVersion" "success.query" "jobFamilies timeline snapshot" "$graphql_status"

# A2: 时态冲突
section_log "A2. Job Catalog Version - 时态冲突 (JC-TEMPORAL)"
payload=$(jq -n \
  --arg name "219C2D 冲突版本" \
  --arg status "ACTIVE" \
  --arg effective "2025-12-01" \
  --arg parent "$catalog_latest_record" \
  '{name:$name,status:$status,effectiveDate:$effective,parentRecordId:$parent}')
resp=$(rest_request POST "/api/v1/job-family-groups/$JFG_CODE/versions" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[A2] Payload: $payload" >>"$VALIDATION_LOG"
echo "[A2] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "jobCatalog.createVersion" "temporal_conflict" "$status_code" "$body" "400" "JOB_CATALOG_TEMPORAL_CONFLICT" "JC-TEMPORAL" "HIGH"

graphql_resp=$(graphql_request "$graphql_query" "$(jq -n --arg code "$JFG_CODE" '{code:$code}')" )
graphql_status=$(echo "$graphql_resp" | tail -n1)
graphql_body=$(echo "$graphql_resp" | head -n -1)
echo "[A2] GraphQL jobFamilies response:" >>"$VALIDATION_LOG"
echo "$graphql_body" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "jobCatalog.createVersion" "temporal_conflict.query" "jobFamilies timeline unchanged" "$graphql_status"

# A3: Job Family Version - 成功
section_log "A3. Job Family Version - 成功"
payload=$(jq -n \
  --arg name "219C2D 自测职种版本" \
  --arg status "ACTIVE" \
  --arg effective "2025-12-01" \
  --arg parent "$JF_RECORD_ID" \
  '{name:$name,status:$status,effectiveDate:$effective,parentRecordId:$parent}')
resp=$(rest_request POST "/api/v1/job-families/$JF_CODE/versions" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[A3] Payload: $payload" >>"$VALIDATION_LOG"
echo "[A3] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
if [[ "$status_code" != "201" ]]; then
  print_log "❌ A3 失败"
  exit 1
fi
JOB_FAMILY_LATEST_VERSION=$(extract_field "$body" '.data.recordId // .data.RecordID')
log_rest_scenario "jobFamily.createVersion" "success" "$status_code" "$body" "201"

job_family_query='query ($group: JobFamilyGroupCode!, $code: JobFamilyCode!) {
  jobFamily(groupCode: $group, code: $code) {
    code
    status
    versions {
      recordId
      effectiveDate
      parentRecordId
    }
  }
}'
graphql_resp=$(graphql_request "$job_family_query" "$(jq -n --arg group "$JFG_CODE" --arg code "$JF_CODE" '{group:$group,code:$code}')" )
graphql_status=$(echo "$graphql_resp" | tail -n1)
graphql_body=$(echo "$graphql_resp" | head -n -1)
echo "[A3] GraphQL jobFamily response:" >>"$VALIDATION_LOG"
echo "$graphql_body" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "jobFamily.createVersion" "success.query" "jobFamily versions snapshot" "$graphql_status"

# A4: Job Family Version - 序列不连续 (JC-SEQUENCE)
section_log "A4. Job Family Version - 序列不连续 (JC-SEQUENCE)"
payload=$(jq -n \
  --arg name "219C2D 序列测试" \
  --arg status "ACTIVE" \
  --arg effective "2026-01-01" \
  --arg parent "$JF_RECORD_ID" \
  '{name:$name,status:$status,effectiveDate:$effective,parentRecordId:$parent}')
resp=$(rest_request POST "/api/v1/job-families/$JF_CODE/versions" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[A4] Payload: $payload" >>"$VALIDATION_LOG"
echo "[A4] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "jobFamily.createVersion" "sequence_mismatch" "$status_code" "$body" "400" "JOB_CATALOG_SEQUENCE_MISMATCH" "JC-SEQUENCE" "HIGH"

graphql_resp=$(graphql_request "$job_family_query" "$(jq -n --arg group "$JFG_CODE" --arg code "$JF_CODE" '{group:$group,code:$code}')" )
graphql_status=$(echo "$graphql_resp" | tail -n1)
graphql_body=$(echo "$graphql_resp" | head -n -1)
echo "[A4] GraphQL jobFamily response:" >>"$VALIDATION_LOG"
echo "$graphql_body" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "jobFamily.createVersion" "sequence_mismatch.query" "jobFamily timeline snapshot" "$graphql_status"

# --- Command B: Position Fill ---

fill_payload() {
  local employee="$1"
  local fte="$2"
  jq -n \
    --arg emp "$employee" \
    --argjson fte "$fte" \
    --arg date "$FILL_EFFECTIVE_DATE" \
    '{employeeId:$emp,employeeName:"Validator Agent",assignmentType:"PRIMARY",
      effectiveDate:$date,operationReason:"219C2D validator self-test",fte:$fte}'
}

section_log "B1. FillPosition - 成功"
payload=$(fill_payload "00000000-0000-0000-0000-00000000a001" "1.0")
resp=$(rest_request POST "/api/v1/positions/$POSITION_CODE/fill" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[B1] Payload: $payload" >>"$VALIDATION_LOG"
echo "[B1] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
if [[ "$status_code" != "200" ]]; then
  print_log "❌ B1 失败"
  exit 1
fi
ASSIGNMENT_ID=$(extract_field "$body" '.data.currentAssignment.assignmentId // .data.CurrentAssignment.AssignmentID')
if [[ -z "$ASSIGNMENT_ID" ]]; then
  ASSIGNMENT_ID=$(extract_field "$body" '.data.assignmentHistory[0].assignmentId // .data.AssignmentHistory[0].AssignmentID')
fi
if [[ -z "$ASSIGNMENT_ID" ]]; then
  print_log "❌ 未能解析当前任职 ID，无法继续"
  exit 1
fi
log_rest_scenario "position.fill" "success" "$status_code" "$body" "200"

graphql_positions_query='query ($code: PositionCode!) {
  position(code: $code) {
    code
    status
    currentAssignment { assignmentId assignmentStatus fte }
  }
}'
resp_graph=$(graphql_request "$graphql_positions_query" "$(jq -n --arg code "$POSITION_CODE" '{code:$code}')" )
status_graph=$(echo "$resp_graph" | tail -n1)
body_graph=$(echo "$resp_graph" | head -n -1)
echo "[B1] GraphQL position response:" >>"$VALIDATION_LOG"
echo "$body_graph" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "position.fill" "success.query" "position snapshot after fill" "$status_graph"

# B2: Headcount exceeded
section_log "B2. FillPosition - POS-HEADCOUNT"
payload=$(fill_payload "00000000-0000-0000-0000-00000000a002" "1.0")
resp=$(rest_request POST "/api/v1/positions/$POSITION_CODE/fill" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[B2] Payload: $payload" >>"$VALIDATION_LOG"
echo "[B2] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "position.fill" "headcount_exceeded" "$status_code" "$body" "400" "POS_HEADCOUNT_EXCEEDED" "POS-HEADCOUNT" "HIGH"

resp_graph=$(graphql_request "$graphql_positions_query" "$(jq -n --arg code "$POSITION_CODE" '{code:$code}')" )
status_graph=$(echo "$resp_graph" | tail -n1)
body_graph=$(echo "$resp_graph" | head -n -1)
echo "[B2] GraphQL position response:" >>"$VALIDATION_LOG"
echo "$body_graph" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "position.fill" "headcount_exceeded.query" "position snapshot" "$status_graph"

# Command C scenario 1: close assignment success
section_log "C1. CloseAssignment - 成功"
close_payload=$(jq -n \
  --arg endDate "2025-12-31" \
  '{endDate:$endDate,operationReason:"219C2D validator self-test"}')
resp=$(rest_request POST "/api/v1/positions/$POSITION_CODE/assignments/$ASSIGNMENT_ID/close" "$close_payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[C1] Payload: $close_payload" >>"$VALIDATION_LOG"
echo "[C1] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "assignment.close" "success" "$status_code" "$body" "200"

resp_graph=$(graphql_request "$graphql_positions_query" "$(jq -n --arg code "$POSITION_CODE" '{code:$code}')" )
status_graph=$(echo "$resp_graph" | tail -n1)
body_graph=$(echo "$resp_graph" | head -n -1)
echo "[C1] GraphQL position response:" >>"$VALIDATION_LOG"
echo "$body_graph" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "assignment.close" "success.query" "position snapshot after close" "$status_graph"

# B3: FillPosition with inactive position (ASSIGN-STATE)
section_log "B3. FillPosition - ASSIGN-STATE"
print_log "将职位状态置为 INACTIVE 以触发 ASSIGN-STATE"
update_payload=$(jq -n \
  --arg title "219C2D Validator Position" \
  --arg group "$JFG_CODE" \
  --arg family "$JF_CODE" \
  --arg role "$JR_CODE" \
  --arg level "$JL_CODE" \
  --arg org "$ORG_CODE" \
  --arg status "INACTIVE" \
  --arg effective "$EFFECTIVE_BASE" \
  '{title:$title,jobFamilyGroupCode:$group,jobFamilyCode:$family,jobRoleCode:$role,jobLevelCode:$level,
    organizationCode:$org,positionType:"REGULAR",employmentType:"FULL_TIME",
    headcountCapacity:1,status:$status,operationReason:"219C2D validator self-test - deactivate",effectiveDate:$effective}')
resp=$(rest_request PUT "/api/v1/positions/$POSITION_CODE" "$update_payload")
update_status=$(echo "$resp" | tail -n1)
if [[ "$update_status" != "200" ]]; then
  print_log "⚠️ 职位状态更新失败，后续 ASSIGN-STATE 可能无法触发"
fi

payload=$(fill_payload "00000000-0000-0000-0000-00000000a003" "0.5")
resp=$(rest_request POST "/api/v1/positions/$POSITION_CODE/fill" "$payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[B3] Payload: $payload" >>"$VALIDATION_LOG"
echo "[B3] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "position.fill" "assign_state_inactive_position" "$status_code" "$body" "400" "ASSIGN_INVALID_STATE" "ASSIGN-STATE" "CRITICAL"

resp_graph=$(graphql_request "$graphql_positions_query" "$(jq -n --arg code "$POSITION_CODE" '{code:$code}')" )
status_graph=$(echo "$resp_graph" | tail -n1)
body_graph=$(echo "$resp_graph" | head -n -1)
echo "[B3] GraphQL position response:" >>"$VALIDATION_LOG"
echo "$body_graph" | jq '.' >>"$VALIDATION_LOG"
log_graphql_scenario "position.fill" "assign_state.query" "position snapshot (inactive)" "$status_graph"

# C2: Close assignment again (ASSIGN-STATE)
section_log "C2. CloseAssignment - 已结束重复关闭 (ASSIGN-STATE)"
resp=$(rest_request POST "/api/v1/positions/$POSITION_CODE/assignments/$ASSIGNMENT_ID/close" "$close_payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[C2] Payload: $close_payload" >>"$VALIDATION_LOG"
echo "[C2] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "assignment.close" "already_closed" "$status_code" "$body" "400" "ASSIGN_INVALID_STATE" "ASSIGN-STATE" "CRITICAL"

# C3: 关闭不存在的任职
section_log "C3. CloseAssignment - 任职不存在"
fake_assignment="00000000-0000-0000-0000-0000deadbeef"
resp=$(rest_request POST "/api/v1/positions/$POSITION_CODE/assignments/$fake_assignment/close" "$close_payload")
status_code=$(echo "$resp" | tail -n1)
body=$(echo "$resp" | head -n -1)
echo -e "\n[C3] Payload: $close_payload" >>"$VALIDATION_LOG"
echo "[C3] HTTP: $status_code" >>"$VALIDATION_LOG"
echo "$body" | jq '.' >>"$VALIDATION_LOG"
log_rest_scenario "assignment.close" "assignment_not_found" "$status_code" "$body" "404" "ASSIGNMENT_NOT_FOUND"

section_log "219C2D Validator 自测完成"

finalize_report
print_log "报告已写入 $REPORT_DIR/report-Day24.json"
print_log "日志追加至 $VALIDATION_LOG"
