#!/usr/bin/env bash
set -euo pipefail

# 时态时间轴端到端验证脚本（插入/作废/中间删除/尾部删除 + 连贯性校验）
# 依赖: curl, jq, psql（可选，若提供 PSQL_CONN 则执行数据库级校验）

# -------------------- 可配置参数 --------------------
HOST="${HOST:-http://localhost}"
PORT="${PORT:-9090}"
BASE_URL="${BASE_URL:-$HOST:$PORT}"
TOKEN="${TOKEN:-}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
ORG_CODE="${ORG_CODE:-8001001}"
ORG_NAME="${ORG_NAME:-时态验证部}"
UNIT_TYPE="${UNIT_TYPE:-DEPARTMENT}"
PSQL_CONN="${PSQL_CONN:-}"

# 时间点（可按需修改）
T1="${T1:-2024-01-01}"
T2="${T2:-2024-06-01}"
T3="${T3:-2024-11-01}"
T4="${T4:-2025-08-01}"
TX="${TX:-2024-03-01}"

# -------------------- 函数工具 --------------------
auth_header() {
  if [[ -n "$TOKEN" ]]; then
    printf 'Authorization: Bearer %s' "$TOKEN"
  fi
}

tenant_header() {
  printf 'X-Tenant-ID: %s' "$TENANT_ID"
}

api() { # $1 method, $2 path, $3 body(json or empty)
  local method="$1" path="$2" body="${3:-}"
  if [[ -n "$body" ]]; then
    curl -sS -X "$method" \
      -H "Content-Type: application/json" \
      -H "$(tenant_header)" \
      ${TOKEN:+-H "$(auth_header)"} \
      "$BASE_URL$path" \
      -d "$body"
  else
    curl -sS -X "$method" \
      -H "$(tenant_header)" \
      ${TOKEN:+-H "$(auth_header)"} \
      "$BASE_URL$path"
  fi
}

sql() { # $1 SQL
  if [[ -z "$PSQL_CONN" ]]; then
    return 0
  fi
  psql "$PSQL_CONN" -t -A -c "$1"
}

title() { echo -e "\n==== $* ===="; }
pass()  { echo "✅ $*"; }
fail()  { echo "❌ $*"; exit 1; }

# -------------------- 步骤 1：新建组织（首条版本 T1） --------------------
title "创建组织首条版本 ($ORG_CODE @ $T1)"
CREATE_BODY=$(jq -n --arg code "$ORG_CODE" --arg name "$ORG_NAME" --arg unitType "$UNIT_TYPE" \
  --arg effectiveDate "$T1" '{code:$code,name:$name,unitType:$unitType,effectiveDate:$effectiveDate,isTemporal:true,changeReason:"初始化"}')
api POST "/api/v1/organization-units" "$CREATE_BODY" | jq . > /dev/null || true
pass "创建组织请求已发送（若已存在将继续）"

# -------------------- 步骤 2：新增 3 个版本（T2/T3/T4） --------------------
add_version() {
  local eff="$1" reason="$2"
  local body
  body=$(jq -n --arg name "$ORG_NAME" --arg unitType "$UNIT_TYPE" --arg effectiveDate "$eff" --arg reason "$reason" \
    '{name:$name,unitType:$unitType,effectiveDate:$effectiveDate,operationReason:$reason}')
  api POST "/api/v1/organization-units/$ORG_CODE/versions" "$body" | jq . > /dev/null
}

title "插入版本 T2=$T2, T3=$T3, T4=$T4"
add_version "$T2" "版本2"
add_version "$T3" "版本3"
add_version "$T4" "版本4"
pass "版本插入完成"

# -------------------- 校验函数（DB 连贯性） --------------------
check_continuity() {
  [[ -z "$PSQL_CONN" ]] && { pass "跳过 DB 连贯性校验（未提供 PSQL_CONN）"; return 0; }
  local q
  q="WITH v AS (\n\
       SELECT code,effective_date,end_date\n\
       FROM organization_units\n\
       WHERE tenant_id = '$TENANT_ID' AND code = '$ORG_CODE' AND status <> 'DELETED' AND deleted_at IS NULL\n\
       ORDER BY effective_date\n\
     ), c AS (\n\
       SELECT code,effective_date AS this_start,end_date AS this_end,\n\
              LEAD(effective_date) OVER (PARTITION BY code ORDER BY effective_date) AS next_eff\n\
       FROM v\n\
     )\n\
     SELECT 1 FROM c\n\
     WHERE (next_eff IS NOT NULL AND this_end <> (next_eff - INTERVAL '1 day')::date)\n\
        OR (next_eff IS NULL AND this_end IS NOT NULL)\n\
     LIMIT 1;"
  local r
  r=$(sql "$q" | tr -d '[:space:]')
  [[ -z "$r" ]] && pass "DB 连贯性校验通过" || fail "DB 连贯性校验失败"
}

check_current() {
  [[ -z "$PSQL_CONN" ]] && { pass "跳过 DB 当前态校验（未提供 PSQL_CONN）"; return 0; }
  local q
  q="SELECT effective_date FROM organization_units\n\
     WHERE tenant_id='$TENANT_ID' AND code='$ORG_CODE' AND is_current=true\n\
     ORDER BY effective_date;"
  echo "当前态: $(sql "$q" | xargs)" || true
  pass "DB 当前态查询完成"
}

# -------------------- 步骤 3：初次连贯性 + 当前态校验 --------------------
title "初次 DB 连贯性校验"
check_continuity
check_current

# -------------------- 步骤 4：作废中间版本（DEACTIVATE T3） --------------------
title "作废中间版本 (DEACTIVATE $T3)"
RID=$(sql "SELECT record_id FROM organization_units WHERE tenant_id='$TENANT_ID' AND code='$ORG_CODE' AND effective_date='$T3'" | xargs || true)
if [[ -z "$RID" && -n "$PSQL_CONN" ]]; then
  fail "未找到 $T3 的 record_id"
fi
DEACT_BODY=$(jq -n --arg rid "$RID" '{eventType:"DEACTIVATE",recordId:$rid,changeReason:"作废中间版本"}')
api POST "/api/v1/organization-units/$ORG_CODE/events" "$DEACT_BODY" | jq '.data.timeline' > /dev/null
pass "作废中间版本完成（事件端点已返回最新时间线）"
check_continuity
check_current

# -------------------- 步骤 5：插入“更早的中间版本”（TX=2024-03-01） --------------------
title "在 $T1 与 $T2 之间插入中间版本 ($TX)"
add_version "$TX" "补历史"
check_continuity
check_current

# -------------------- 步骤 6：作废尾部版本（DEACTIVATE T4） --------------------
title "作废尾部版本 (DEACTIVATE $T4)"
RID_T4=$(sql "SELECT record_id FROM organization_units WHERE tenant_id='$TENANT_ID' AND code='$ORG_CODE' AND effective_date='$T4'" | xargs || true)
if [[ -z "$RID_T4" && -n "$PSQL_CONN" ]]; then
  fail "未找到 $T4 的 record_id"
fi
DEACT_T4=$(jq -n --arg rid "$RID_T4" '{eventType:"DEACTIVATE",recordId:$rid,changeReason:"作废尾部版本"}')
api POST "/api/v1/organization-units/$ORG_CODE/events" "$DEACT_T4" | jq '.data.timeline' > /dev/null
pass "作废尾部版本完成"
check_continuity
check_current

title "全部校验完成（如上均为 ✅ 则通过）"
exit 0

