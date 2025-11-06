#!/usr/bin/env bash
#
# 219C3 REST 命令自测脚本
# 目标：
#   1. 通过 REST 覆盖组织/职位/任职命令的关键正反场景（Create / Fill / Close 等）。
#   2. 补充 Job Level 版本命令的正反场景，用于验证 validator 规则与审计写入。
#   3. 将执行结果写入 logs/219C3/validation.log 与 report.json，供 219C3 验收引用。

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq 未安装，无法执行脚本。" >&2
  exit 1
fi

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

BASE_URL_COMMAND="${BASE_URL_COMMAND:-http://localhost:9090}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
TOKEN_FILE=".cache/dev.jwt"
LOG_DIR="$ROOT_DIR/logs/219C3"
VALIDATION_LOG="$LOG_DIR/validation.log"
REPORT_TMP="$(mktemp)"
TIMESTAMP="$(date '+%Y-%m-%dT%H:%M:%S%z')"
SUMMARY_FILE="$LOG_DIR/report.json"

mkdir -p "$LOG_DIR"

print_log() {
  local msg="$1"
  printf '%s %s\n' "[$(date '+%Y-%m-%dT%H:%M:%S%z')]" "$msg" | tee -a "$VALIDATION_LOG" >/dev/null
}

section_log() {
  local msg="$1"
  printf '\n========== %s ==========\n' "$msg" | tee -a "$VALIDATION_LOG" >/dev/null
}

append_report() {
  echo "$1" >>"$REPORT_TMP"
}

finalize_report() {
  jq -s '.' "$REPORT_TMP" >"$SUMMARY_FILE"
  rm -f "$REPORT_TMP"
  print_log "报告已写入 $SUMMARY_FILE"
}

ensure_token() {
  if [[ ! -f "$TOKEN_FILE" ]]; then
    print_log "检测到缺少令牌，执行 make jwt-dev-mint ..."
    make jwt-dev-mint >/dev/null
  fi
  TOKEN="$(cat "$TOKEN_FILE")"
}

REST_BODY=""
REST_STATUS=""
REST_REQUEST_ID=""
REST_CORRELATION_ID=""

rest_request() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local tmp_headers
  tmp_headers="$(mktemp)"

  local curl_args=(-sS -w '\n%{http_code}' -X "$method" "$BASE_URL_COMMAND$path"
    -H "Content-Type: application/json"
    -H "Authorization: Bearer $TOKEN"
    -H "X-Tenant-ID: $TENANT_ID"
    -D "$tmp_headers")
  if [[ -n "$data" ]]; then
    curl_args+=(-d "$data")
  fi

  local response
  response="$(curl "${curl_args[@]}")" || true

  REST_STATUS="$(echo "$response" | tail -n1)"
  REST_BODY="$(echo "$response" | head -n -1)"
  REST_REQUEST_ID="$(sed -n 's/^[Xx]-Request-Id:[[:space:]]*//p' "$tmp_headers" | tail -n1 | tr -d '\r')"
  REST_CORRELATION_ID="$(sed -n 's/^[Xx]-Correlation-Id:[[:space:]]*//p' "$tmp_headers" | tail -n1 | tr -d '\r')"
  rm -f "$tmp_headers"

  if [[ -z "$REST_REQUEST_ID" ]]; then
    REST_REQUEST_ID="$(echo "$REST_BODY" | jq -r '.requestId // empty' 2>/dev/null || true)"
  fi
  if [[ -z "$REST_CORRELATION_ID" ]]; then
    REST_CORRELATION_ID="$(echo "$REST_BODY" | jq -r '.correlationId // empty' 2>/dev/null || true)"
  fi
}

extract_field() {
  echo "$1" | jq -r "$2 // empty" 2>/dev/null || true
}

log_rest_scenario() {
  local command="$1"
  local scenario="$2"
  local expected_status="${3:-}"
  local expected_code="${4:-}"
  local expected_rule="${5:-}"
  local expected_severity="${6:-}"

  local status="$REST_STATUS"
  local body="$REST_BODY"
  local request_id="$REST_REQUEST_ID"
  local correlation_id="$REST_CORRELATION_ID"
  local error_code rule_id severity outcome="passed"

  error_code="$(extract_field "$body" '.error.code // .error.Code // empty')"
  rule_id="$(extract_field "$body" '.error.details.ruleId // .error.details.validationErrors[0].context.ruleId // empty')"
  severity="$(extract_field "$body" '.error.details.validationErrors[0].severity // empty')"

  if [[ -n "$expected_status" && "$status" != "$expected_status" ]]; then
    outcome="failed"
    print_log "⚠️  [$command/$scenario] 期望 HTTP $expected_status，实际 $status"
  fi
  if [[ -n "$expected_code" && "$error_code" != "$expected_code" ]]; then
    outcome="failed"
    print_log "⚠️  [$command/$scenario] 期望错误码 $expected_code，实际 ${error_code:-<空>}"
  fi
  if [[ -n "$expected_rule" && "$rule_id" != "$expected_rule" ]]; then
    outcome="failed"
    print_log "⚠️  [$command/$scenario] 期望 ruleId $expected_rule，实际 ${rule_id:-<空>}"
  fi
  if [[ -n "$expected_severity" && "${severity^^}" != "${expected_severity^^}" ]]; then
    outcome="failed"
    print_log "⚠️  [$command/$scenario] 期望 severity $expected_severity，实际 ${severity:-<空>}"
  fi

  printf '\n[%s/%s] HTTP %s\n' "$command" "$scenario" "$status" >>"$VALIDATION_LOG"
  echo "$body" | jq '.' >>"$VALIDATION_LOG" || echo "$body" >>"$VALIDATION_LOG"

  append_report "$(jq -n \
    --arg command "$command" \
    --arg scenario "$scenario" \
    --arg status "$status" \
    --arg outcome "$outcome" \
    --arg err "$error_code" \
    --arg rid "$rule_id" \
    --arg sev "$severity" \
    --arg req "$request_id" \
    --arg corr "$correlation_id" \
    '{command:$command,scenario:$scenario,
      result:{
        status:$outcome,
        httpStatus: ($status|tonumber?),
        errorCode: ($err|select(.!="")),
        ruleId: ($rid|select(.!="")),
        severity: ($sev|select(.!="")),
        requestId: ($req|select(.!="")),
        correlationId: ($corr|select(.!=""))
      }}')"
}

verify_audit() {
  local request_id="$1"
  if [[ -z "$request_id" ]]; then
    print_log "⚠️  未提供 requestId，跳过审计验证。"
    return
  fi
  if [[ -z "${DATABASE_URL:-}" ]]; then
    print_log "ℹ️  未设置 DATABASE_URL，跳过审计查询（requestId=$request_id）。"
    return
  fi
  if ! command -v psql >/dev/null 2>&1; then
    print_log "ℹ️  未检测到 psql，跳过审计查询（requestId=$request_id）。"
    return
  fi

  print_log "查询审计记录（requestId=$request_id）"
  local sql="SELECT tenant_id,
       resource_type,
       request_id,
       business_context->>'ruleId' AS rule_id,
       business_context->>'severity' AS severity,
       business_context->>'correlationId' AS correlation
  FROM audit_logs
 WHERE request_id = '$request_id'
 ORDER BY \"timestamp\" DESC
 LIMIT 1;"
  psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -P footer=off -P "format=aligned" \
    -c "$sql" >>"$VALIDATION_LOG" 2>&1 || {
      print_log "⚠️  审计查询失败（requestId=$request_id）"
    }
}

ensure_token

section_log "219C3 REST 命令自测开始 ($TIMESTAMP)"

print_log "准备基础数据（组织 + Job Catalog + 职位）"

random_suffix="$(date +%s%N | tail -c 6)"
EFFECTIVE_BASE="${EFFECTIVE_BASE:-$(date '+%Y-%m-%d')}"

# 1. 创建组织
ORG_CODE="REST${random_suffix}"
org_payload=$(jq -n \
  --arg code "$ORG_CODE" \
  --arg name "219C3 测试组织" \
  '{code:$code,name:$name,unitType:"DEPARTMENT",operationReason:"219C3 rest self-test"}')
rest_request POST "/api/v1/organization-units" "$org_payload"
log_rest_scenario "organization.create" "success" "201"
verify_audit "$REST_REQUEST_ID"

# 2. 创建 Job Catalog 层级
JFG_CODE="JFG-219C3-$random_suffix"
JF_CODE="JF-219C3-$random_suffix"
JR_CODE="JR-219C3-$random_suffix"
JL_CODE="JL-219C3-$random_suffix"

create_job_family_group() {
  local payload=$(jq -n \
    --arg code "$JFG_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,name:"219C3 自测职类",status:"ACTIVE",effectiveDate:$effective}')
  rest_request POST "/api/v1/job-family-groups" "$payload"
  log_rest_scenario "jobFamilyGroup.create" "success" "201"
  JFG_RECORD_ID="$(extract_field "$REST_BODY" '.data.recordId // .data.RecordID')"
}

create_job_family() {
  local payload=$(jq -n \
    --arg code "$JF_CODE" \
    --arg group "$JFG_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,jobFamilyGroupCode:$group,name:"219C3 自测职种",status:"ACTIVE",effectiveDate:$effective}')
  rest_request POST "/api/v1/job-families" "$payload"
  log_rest_scenario "jobFamily.create" "success" "201"
  JF_RECORD_ID="$(extract_field "$REST_BODY" '.data.recordId // .data.RecordID')"
}

create_job_role() {
  local payload=$(jq -n \
    --arg code "$JR_CODE" \
    --arg family "$JF_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,jobFamilyCode:$family,name:"219C3 自测职务",status:"ACTIVE",effectiveDate:$effective}')
  rest_request POST "/api/v1/job-roles" "$payload"
  log_rest_scenario "jobRole.create" "success" "201"
  JR_RECORD_ID="$(extract_field "$REST_BODY" '.data.recordId // .data.RecordID')"
}

create_job_level() {
  local payload=$(jq -n \
    --arg code "$JL_CODE" \
    --arg role "$JR_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{code:$code,jobRoleCode:$role,levelRank:"1",name:"219C3 自测职级",status:"ACTIVE",effectiveDate:$effective}')
  rest_request POST "/api/v1/job-levels" "$payload"
  log_rest_scenario "jobLevel.create" "success" "201"
  JL_RECORD_ID="$(extract_field "$REST_BODY" '.data.recordId // .data.RecordID')"
}

create_job_family_group
create_job_family
create_job_role
create_job_level

# 3. 创建职位
create_position() {
  local payload=$(jq -n \
    --arg title "219C3 Validator Position" \
    --arg group "$JFG_CODE" \
    --arg family "$JF_CODE" \
    --arg role "$JR_CODE" \
    --arg level "$JL_CODE" \
    --arg org "$ORG_CODE" \
    --arg effective "$EFFECTIVE_BASE" \
    '{title:$title,jobFamilyGroupCode:$group,jobFamilyCode:$family,jobRoleCode:$role,jobLevelCode:$level,
      organizationCode:$org,positionType:"REGULAR",employmentType:"FULL_TIME",
      headcountCapacity:1,operationReason:"219C3 rest self-test",effectiveDate:$effective}')
  rest_request POST "/api/v1/positions" "$payload"
  log_rest_scenario "position.create" "success" "201"
  POSITION_CODE="$(extract_field "$REST_BODY" '.data.code // .data.Code')"
  POSITION_RECORD_ID="$(extract_field "$REST_BODY" '.data.recordId // .data.RecordID')"
}

create_position

# 4. 任职相关场景
fill_payload() {
  local worker_id="$1"
  local fte="$2"
  jq -n \
    --arg worker "$worker_id" \
    --arg fte "$fte" \
    --arg effective "$EFFECTIVE_BASE" \
    '{workerId:$worker,fte:($fte|tonumber),effectiveDate:$effective,operationReason:"219C3 rest self-test"}'
}

# A. 成功填充任职
payload=$(fill_payload "00000000-0000-0000-0000-00000000a001" "1.0")
rest_request POST "/api/v1/positions/$POSITION_CODE/fill" "$payload"
log_rest_scenario "position.fill" "success" "200"
ASSIGNMENT_ID="$(extract_field "$REST_BODY" '.data.assignmentId // .data.assignmentID')"
verify_audit "$REST_REQUEST_ID"

# B. headcount 超限
payload=$(fill_payload "00000000-0000-0000-0000-00000000a002" "1.0")
rest_request POST "/api/v1/positions/$POSITION_CODE/fill" "$payload"
log_rest_scenario "position.fill" "headcount_exceeded" "400" "POS_HEADCOUNT_EXCEEDED" "POS-HEADCOUNT" "HIGH"
verify_audit "$REST_REQUEST_ID"

# C. 关闭任职成功
close_payload=$(jq -n \
  --arg reason "219C3 rest self-test" \
  '{operationReason:$reason}')
rest_request POST "/api/v1/positions/$POSITION_CODE/assignments/$ASSIGNMENT_ID/close" "$close_payload"
log_rest_scenario "assignment.close" "success" "200"
verify_audit "$REST_REQUEST_ID"

# D. 再次关闭，触发状态校验
rest_request POST "/api/v1/positions/$POSITION_CODE/assignments/$ASSIGNMENT_ID/close" "$close_payload"
log_rest_scenario "assignment.close" "already_closed" "400" "ASSIGN_INVALID_STATE" "ASSIGN-STATE" "CRITICAL"
verify_audit "$REST_REQUEST_ID"

# 5. Job Level 版本场景
# 使用已有 recordId 创建新版本（成功）
payload=$(jq -n \
  --arg name "219C3 职级版本成功" \
  --arg status "ACTIVE" \
  --arg effective "$(date -d "$EFFECTIVE_BASE +7 day" '+%Y-%m-%d')" \
  --arg parent "$JL_RECORD_ID" \
  '{name:$name,status:$status,effectiveDate:$effective,parentRecordId:$parent}')
rest_request POST "/api/v1/job-levels/$JL_CODE/versions" "$payload"
log_rest_scenario "jobLevel.version" "success" "201"
verify_audit "$REST_REQUEST_ID"

# 冲突版本：使用相同生效日
payload=$(jq -n \
  --arg name "219C3 职级版本冲突" \
  --arg status "ACTIVE" \
  --arg effective "$(date -d "$EFFECTIVE_BASE +7 day" '+%Y-%m-%d')" \
  --arg parent "$JL_RECORD_ID" \
  '{name:$name,status:$status,effectiveDate:$effective,parentRecordId:$parent}')
rest_request POST "/api/v1/job-levels/$JL_CODE/versions" "$payload"
log_rest_scenario "jobLevel.version" "temporal_conflict" "400" "JOB_CATALOG_TEMPORAL_CONFLICT" "JC-TEMPORAL" "HIGH"
verify_audit "$REST_REQUEST_ID"

section_log "219C3 REST 命令自测结束"
finalize_report

print_log "如需进一步核对审计，可参考日志中的 requestId 调用 verify_audit。"
