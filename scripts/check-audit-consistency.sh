#!/usr/bin/env bash
set -euo pipefail

# 用法： scripts/check-audit-consistency.sh 1000001 [GRAPHQL_URL]
# 默认 GraphQL_URL: http://localhost:8090/graphql

CODE=${1:-}
GRAPHQL_URL=${2:-http://localhost:8090/graphql}

if [[ -z "$CODE" ]]; then
  echo "Usage: $0 <ORG_CODE> [GRAPHQL_URL]" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required (sudo apt-get install -y jq)." >&2
  exit 1
fi

echo "[1/3] Fetching organizationVersions for code=$CODE ..."
VERSIONS_JSON=$(curl -s -X POST "$GRAPHQL_URL" \
  -H 'Content-Type: application/json' \
  -d "{\"query\":\"query($$code: String!){ organizationVersions(code: $$code){ recordId effectiveDate endDate isCurrent name } }\",\"variables\":{\"code\":\"$CODE\"}}")

ERR=$(echo "$VERSIONS_JSON" | jq -r '.errors[0].message? // empty')
if [[ -n "$ERR" ]]; then
  echo "GraphQL error: $ERR" >&2
  exit 2
fi

COUNT=$(echo "$VERSIONS_JSON" | jq '.data.organizationVersions | length')
echo "Found $COUNT versions"

if [[ "$COUNT" -eq 0 ]]; then
  echo "No versions found for code=$CODE" >&2
  exit 0
fi

FAIL=0
TOTAL=0

echo "[2/3] Checking auditHistory(recordId) for each version ..."
echo "$VERSIONS_JSON" | jq -r '.data.organizationVersions[] | [.recordId, .effectiveDate, .isCurrent, .name] | @tsv' | while IFS=$'\t' read -r RID EFF CUR NAME; do
  TOTAL=$((TOTAL+1))
  echo "- Version $TOTAL: recordId=$RID effectiveDate=$EFF isCurrent=$CUR name=$NAME"

  QPAYLOAD=$(jq -n --arg rid "$RID" '{query:"query($recordId: String!, $limit: Int){ auditHistory(recordId:$recordId, limit:$limit){ auditId recordId operationType timestamp } }", variables:{recordId:$rid, limit:50}}')
  RESP=$(curl -s -X POST "$GRAPHQL_URL" -H 'Content-Type: application/json' -d "$QPAYLOAD")
  GE=$(echo "$RESP" | jq -r '.errors[0].message? // empty')
  if [[ -n "$GE" ]]; then
    echo "  ! GraphQL error: $GE" >&2
    FAIL=$((FAIL+1))
    continue
  fi

  MISMATCH=$(echo "$RESP" | jq --arg rid "$RID" '[.data.auditHistory[] | select(.recordId != $rid)] | length')
  LEN=$(echo "$RESP" | jq '.data.auditHistory | length')
  if [[ "$MISMATCH" -gt 0 ]]; then
    echo "  ! Mismatch: $MISMATCH of $LEN entries do not match recordId=$RID"
    FAIL=$((FAIL+1))
  else
    echo "  ✓ OK: $LEN entries all match recordId=$RID"
  fi
done

echo "[3/3] Summary: total versions=$TOTAL, failures=$FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 3
fi
exit 0

